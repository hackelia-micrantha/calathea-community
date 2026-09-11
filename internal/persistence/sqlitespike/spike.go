// Package sqlitespike is an isolated executable validation harness for ADR 0006.
// It is not Calathea's production persistence adapter or application persistence API.
package sqlitespike

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	_ "modernc.org/sqlite"
)

const (
	PersistedSchemaVersion = 1
	requestFingerprintV1   = "sqlite-spike.write-request.v1"
)

var (
	ErrIdempotencyConflict = errors.New("idempotency conflict")
	ErrExpectedConflict    = errors.New("expected revision conflict")
	ErrStorageBusy         = errors.New("sqlite storage busy or locked")
	ErrUnsupportedSchema   = errors.New("unsupported persisted schema")
	ErrDomainIntegrity     = errors.New("domain integrity failure")
	ErrInjectedFailure     = errors.New("injected pre-commit failure")
)

type Store struct {
	db   *sql.DB
	path string
}

type WriteRequest struct {
	OperationID        string
	EntityID           string
	RecordID           string
	Payload            string
	ExpectedRevision   int64
	NewRevision        int64
	InjectBeforeCommit bool
}

type WriteResult struct {
	RecordID string
	Replayed bool
}

type ProjectionRow struct {
	EntityID       string
	RecordID       string
	Payload        string
	SourceRevision int64
}

func Open(path string) (*Store, error) {
	if strings.TrimSpace(path) == "" {
		return nil, errors.New("sqlite path must not be empty")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, fmt.Errorf("create sqlite parent: %w", err)
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, fmt.Errorf("create sqlite file: %w", err)
	}
	if err := file.Close(); err != nil {
		return nil, fmt.Errorf("close sqlite bootstrap file: %w", err)
	}
	return openExisting(path)
}

func openExisting(path string) (*Store, error) {
	if _, err := os.Stat(path); err != nil {
		return nil, fmt.Errorf("stat sqlite file: %w", err)
	}
	dsn := sqliteDSN(path)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	db.SetMaxOpenConns(4)
	db.SetMaxIdleConns(4)
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping sqlite: %w", err)
	}
	store := &Store{db: db, path: path}
	enabled, err := store.ForeignKeysEnabled(context.Background())
	if err != nil {
		db.Close()
		return nil, err
	}
	if !enabled {
		db.Close()
		return nil, errors.New("sqlite foreign key enforcement is disabled")
	}
	return store, nil
}

func sqliteDSN(path string) string {
	u := url.URL{Scheme: "file", Path: filepath.ToSlash(path)}
	query := u.Query()
	query.Set("_busy_timeout", "250")
	query.Set("_foreign_keys", "on")
	query.Set("_journal_mode", "delete")
	query.Set("_synchronous", "full")
	query.Set("_dqs", "false")
	u.RawQuery = query.Encode()
	return u.String()
}

func (s *Store) Close() error { return s.db.Close() }
func (s *Store) Path() string { return s.path }

func (s *Store) InitializeV0(ctx context.Context) error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS calathea_meta (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS authoritative_records (
			id TEXT PRIMARY KEY,
			entity_id TEXT NOT NULL,
			revision INTEGER NOT NULL CHECK (revision > 0),
			payload TEXT NOT NULL,
			UNIQUE (entity_id, revision)
		)`,
		`CREATE TABLE IF NOT EXISTS current_records (
			entity_id TEXT PRIMARY KEY,
			record_id TEXT NOT NULL REFERENCES authoritative_records(id),
			revision INTEGER NOT NULL CHECK (revision > 0)
		)`,
		`CREATE TABLE IF NOT EXISTS operations (
			operation_id TEXT PRIMARY KEY,
			fingerprint_version TEXT NOT NULL,
			fingerprint TEXT NOT NULL,
			result_record_id TEXT NOT NULL REFERENCES authoritative_records(id)
		)`,
	}
	return s.withTx(ctx, func(tx *sql.Tx) error {
		for _, statement := range statements {
			if _, err := tx.ExecContext(ctx, statement); err != nil {
				return fmt.Errorf("initialize v0 schema: %w", err)
			}
		}
		_, err := tx.ExecContext(ctx,
			`INSERT INTO calathea_meta(key, value) VALUES ('persisted_schema_version', '0')
			 ON CONFLICT(key) DO NOTHING`)
		if err != nil {
			return fmt.Errorf("initialize persisted schema version: %w", err)
		}
		return nil
	})
}

func (s *Store) InitializeLatest(ctx context.Context) error {
	if err := s.InitializeV0(ctx); err != nil {
		return err
	}
	return s.MigrateToLatest(ctx)
}

func (s *Store) MigrateToLatest(ctx context.Context) error {
	return s.withTx(ctx, func(tx *sql.Tx) error {
		version, err := readSchemaVersion(ctx, tx)
		if err != nil {
			return err
		}
		if version > PersistedSchemaVersion {
			return fmt.Errorf("%w: found %d, max supported %d", ErrUnsupportedSchema, version, PersistedSchemaVersion)
		}
		if version == PersistedSchemaVersion {
			return nil
		}
		if version != 0 {
			return fmt.Errorf("%w: no migration path from %d", ErrUnsupportedSchema, version)
		}
		if _, err := tx.ExecContext(ctx, `CREATE TABLE record_projection (
			entity_id TEXT PRIMARY KEY,
			record_id TEXT NOT NULL,
			payload TEXT NOT NULL,
			source_revision INTEGER NOT NULL CHECK (source_revision > 0)
		)`); err != nil {
			return fmt.Errorf("create v1 projection table: %w", err)
		}
		if _, err := tx.ExecContext(ctx,
			`UPDATE calathea_meta SET value = ? WHERE key = 'persisted_schema_version'`,
			strconv.Itoa(PersistedSchemaVersion)); err != nil {
			return fmt.Errorf("advance persisted schema version: %w", err)
		}
		return nil
	})
}

func readSchemaVersion(ctx context.Context, queryer interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}) (int, error) {
	var value string
	if err := queryer.QueryRowContext(ctx,
		`SELECT value FROM calathea_meta WHERE key = 'persisted_schema_version'`).Scan(&value); err != nil {
		return 0, fmt.Errorf("read persisted schema version: %w", err)
	}
	version, err := strconv.Atoi(value)
	if err != nil || version < 0 {
		return 0, fmt.Errorf("invalid persisted schema version %q", value)
	}
	return version, nil
}

func (s *Store) SchemaVersion(ctx context.Context) (int, error) {
	return readSchemaVersion(ctx, s.db)
}

func (s *Store) SetSchemaVersionForTest(ctx context.Context, version int) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE calathea_meta SET value = ? WHERE key = 'persisted_schema_version'`,
		strconv.Itoa(version))
	return err
}

func (s *Store) Apply(ctx context.Context, request WriteRequest) (WriteResult, error) {
	if err := validateWriteRequest(request); err != nil {
		return WriteResult{}, err
	}
	fingerprint := materialFingerprint(request)
	var result WriteResult
	err := s.withTx(ctx, func(tx *sql.Tx) error {
		var priorFingerprint, priorResult string
		err := tx.QueryRowContext(ctx,
			`SELECT fingerprint, result_record_id FROM operations WHERE operation_id = ?`,
			request.OperationID).Scan(&priorFingerprint, &priorResult)
		switch {
		case err == nil:
			if priorFingerprint != fingerprint {
				return fmt.Errorf("%w: operation %q", ErrIdempotencyConflict, request.OperationID)
			}
			result = WriteResult{RecordID: priorResult, Replayed: true}
			return nil
		case !errors.Is(err, sql.ErrNoRows):
			return fmt.Errorf("look up operation: %w", err)
		}

		currentRevision, exists, err := currentRevision(ctx, tx, request.EntityID)
		if err != nil {
			return err
		}
		if (!exists && request.ExpectedRevision != 0) || (exists && currentRevision != request.ExpectedRevision) {
			return fmt.Errorf("%w: entity %q expected %d current %d", ErrExpectedConflict, request.EntityID, request.ExpectedRevision, currentRevision)
		}

		if _, err := tx.ExecContext(ctx,
			`INSERT INTO authoritative_records(id, entity_id, revision, payload) VALUES (?, ?, ?, ?)`,
			request.RecordID, request.EntityID, request.NewRevision, request.Payload); err != nil {
			return fmt.Errorf("insert authoritative record: %w", err)
		}

		if !exists {
			if request.ExpectedRevision != 0 || request.NewRevision != 1 {
				return fmt.Errorf("%w: new entity %q must advance 0 -> 1", ErrExpectedConflict, request.EntityID)
			}
			if _, err := tx.ExecContext(ctx,
				`INSERT INTO current_records(entity_id, record_id, revision) VALUES (?, ?, ?)`,
				request.EntityID, request.RecordID, request.NewRevision); err != nil {
				return fmt.Errorf("insert current reference: %w", err)
			}
		} else {
			if request.NewRevision != request.ExpectedRevision+1 {
				return fmt.Errorf("new revision %d must follow expected revision %d", request.NewRevision, request.ExpectedRevision)
			}
			update, err := tx.ExecContext(ctx,
				`UPDATE current_records SET record_id = ?, revision = ? WHERE entity_id = ? AND revision = ?`,
				request.RecordID, request.NewRevision, request.EntityID, request.ExpectedRevision)
			if err != nil {
				return fmt.Errorf("advance current reference: %w", err)
			}
			rows, err := update.RowsAffected()
			if err != nil {
				return fmt.Errorf("read current-reference update count: %w", err)
			}
			if rows != 1 {
				return fmt.Errorf("%w: entity %q current reference changed", ErrExpectedConflict, request.EntityID)
			}
		}

		if _, err := tx.ExecContext(ctx,
			`INSERT INTO operations(operation_id, fingerprint_version, fingerprint, result_record_id) VALUES (?, ?, ?, ?)`,
			request.OperationID, requestFingerprintV1, fingerprint, request.RecordID); err != nil {
			return fmt.Errorf("insert operation record: %w", err)
		}
		if request.InjectBeforeCommit {
			return ErrInjectedFailure
		}
		result = WriteResult{RecordID: request.RecordID}
		return nil
	})
	return result, err
}

func validateWriteRequest(request WriteRequest) error {
	fields := []struct {
		name  string
		value string
	}{
		{name: "operation id", value: request.OperationID},
		{name: "entity id", value: request.EntityID},
		{name: "record id", value: request.RecordID},
		{name: "payload", value: request.Payload},
	}
	for _, field := range fields {
		if strings.TrimSpace(field.value) == "" {
			return fmt.Errorf("%s must not be empty", field.name)
		}
	}
	if request.ExpectedRevision < 0 || request.NewRevision <= 0 {
		return errors.New("revisions must be non-negative expected and positive new values")
	}
	return nil
}

func materialFingerprint(request WriteRequest) string {
	var material strings.Builder
	writeCanonicalField(&material, "schema", requestFingerprintV1)
	writeCanonicalField(&material, "entity_id", request.EntityID)
	writeCanonicalField(&material, "record_id", request.RecordID)
	writeCanonicalField(&material, "payload", request.Payload)
	writeCanonicalField(&material, "expected_revision", strconv.FormatInt(request.ExpectedRevision, 10))
	writeCanonicalField(&material, "new_revision", strconv.FormatInt(request.NewRevision, 10))
	digest := sha256.Sum256([]byte(material.String()))
	return hex.EncodeToString(digest[:])
}

func writeCanonicalField(builder *strings.Builder, name, value string) {
	fmt.Fprintf(builder, "%d:%s=%d:%s\n", len(name), name, len(value), value)
}

func currentRevision(ctx context.Context, tx *sql.Tx, entityID string) (int64, bool, error) {
	var revision int64
	err := tx.QueryRowContext(ctx,
		`SELECT revision FROM current_records WHERE entity_id = ?`, entityID).Scan(&revision)
	switch {
	case err == nil:
		return revision, true, nil
	case errors.Is(err, sql.ErrNoRows):
		return 0, false, nil
	default:
		return 0, false, fmt.Errorf("read current revision: %w", err)
	}
}

func (s *Store) RebuildProjection(ctx context.Context) error {
	return s.withTx(ctx, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx, `DELETE FROM record_projection`); err != nil {
			return fmt.Errorf("clear projection: %w", err)
		}
		_, err := tx.ExecContext(ctx, `
			INSERT INTO record_projection(entity_id, record_id, payload, source_revision)
			SELECT c.entity_id, r.id, r.payload, c.revision
			FROM current_records c
			JOIN authoritative_records r ON r.id = c.record_id
			ORDER BY c.entity_id`)
		if err != nil {
			return fmt.Errorf("rebuild projection: %w", err)
		}
		return nil
	})
}

func (s *Store) Projection(ctx context.Context) ([]ProjectionRow, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT entity_id, record_id, payload, source_revision FROM record_projection ORDER BY entity_id`)
	if err != nil {
		return nil, fmt.Errorf("query projection: %w", err)
	}
	defer rows.Close()
	var result []ProjectionRow
	for rows.Next() {
		var row ProjectionRow
		if err := rows.Scan(&row.EntityID, &row.RecordID, &row.Payload, &row.SourceRevision); err != nil {
			return nil, fmt.Errorf("scan projection: %w", err)
		}
		result = append(result, row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate projection: %w", err)
	}
	return result, nil
}

func (s *Store) CorruptProjectionForTest(ctx context.Context, entityID string) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE record_projection SET payload = 'corrupt-projection' WHERE entity_id = ?`, entityID)
	return err
}

func (s *Store) InsertSemanticMismatchForTest(ctx context.Context) error {
	return s.withTx(ctx, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO authoritative_records(id, entity_id, revision, payload) VALUES ('foreign-record', 'entity-b', 1, 'payload-b')`); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO current_records(entity_id, record_id, revision) VALUES ('entity-a', 'foreign-record', 1)`); err != nil {
			return err
		}
		return nil
	})
}

func (s *Store) Counts(ctx context.Context) (records, currents, operations int, err error) {
	for _, item := range []struct {
		query string
		dest  *int
	}{
		{`SELECT count(*) FROM authoritative_records`, &records},
		{`SELECT count(*) FROM current_records`, &currents},
		{`SELECT count(*) FROM operations`, &operations},
	} {
		if scanErr := s.db.QueryRowContext(ctx, item.query).Scan(item.dest); scanErr != nil {
			return 0, 0, 0, scanErr
		}
	}
	return records, currents, operations, nil
}

func (s *Store) ForeignKeysEnabled(ctx context.Context) (bool, error) {
	var enabled int
	if err := s.db.QueryRowContext(ctx, `PRAGMA foreign_keys`).Scan(&enabled); err != nil {
		return false, fmt.Errorf("read foreign_keys pragma: %w", err)
	}
	return enabled == 1, nil
}

func (s *Store) JournalMode(ctx context.Context) (string, error) {
	var mode string
	if err := s.db.QueryRowContext(ctx, `PRAGMA journal_mode`).Scan(&mode); err != nil {
		return "", fmt.Errorf("read journal mode: %w", err)
	}
	return strings.ToLower(mode), nil
}

func (s *Store) SynchronousLevel(ctx context.Context) (int, error) {
	var level int
	if err := s.db.QueryRowContext(ctx, `PRAGMA synchronous`).Scan(&level); err != nil {
		return 0, fmt.Errorf("read synchronous pragma: %w", err)
	}
	return level, nil
}

func (s *Store) Validate(ctx context.Context) error {
	if err := validateSQLiteIntegrity(ctx, s.db); err != nil {
		return err
	}
	if err := validateForeignKeys(ctx, s.db); err != nil {
		return err
	}
	var currentEntity, recordEntity string
	var currentRevision, recordRevision int64
	err := s.db.QueryRowContext(ctx, `
		SELECT c.entity_id, r.entity_id, c.revision, r.revision
		FROM current_records c
		JOIN authoritative_records r ON r.id = c.record_id
		WHERE c.entity_id <> r.entity_id OR c.revision <> r.revision
		ORDER BY c.entity_id
		LIMIT 1`).Scan(&currentEntity, &recordEntity, &currentRevision, &recordRevision)
	switch {
	case err == nil:
		return fmt.Errorf("%w: current entity/revision %q/%d references record entity/revision %q/%d",
			ErrDomainIntegrity, currentEntity, currentRevision, recordEntity, recordRevision)
	case errors.Is(err, sql.ErrNoRows):
		return nil
	default:
		return fmt.Errorf("validate domain current-record invariant: %w", err)
	}
}

func validateSQLiteIntegrity(ctx context.Context, db *sql.DB) error {
	rows, err := db.QueryContext(ctx, `PRAGMA integrity_check`)
	if err != nil {
		return fmt.Errorf("sqlite integrity_check: %w", err)
	}
	defer rows.Close()
	seen := false
	for rows.Next() {
		seen = true
		var message string
		if err := rows.Scan(&message); err != nil {
			return fmt.Errorf("scan sqlite integrity result: %w", err)
		}
		if strings.ToLower(message) != "ok" {
			return fmt.Errorf("sqlite integrity failure: %s", message)
		}
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate sqlite integrity result: %w", err)
	}
	if !seen {
		return errors.New("sqlite integrity_check returned no result")
	}
	return nil
}

func validateForeignKeys(ctx context.Context, db *sql.DB) error {
	rows, err := db.QueryContext(ctx, `PRAGMA foreign_key_check`)
	if err != nil {
		return fmt.Errorf("sqlite foreign_key_check: %w", err)
	}
	defer rows.Close()
	if rows.Next() {
		var table string
		var rowID any
		var parent string
		var fkID int
		if err := rows.Scan(&table, &rowID, &parent, &fkID); err != nil {
			return fmt.Errorf("scan foreign key violation: %w", err)
		}
		return fmt.Errorf("foreign key violation: table=%s row=%v parent=%s fk=%d", table, rowID, parent, fkID)
	}
	return rows.Err()
}

func ValidateFile(ctx context.Context, path string) error {
	store, err := openExisting(path)
	if err != nil {
		return err
	}
	defer store.Close()
	version, err := store.SchemaVersion(ctx)
	if err != nil {
		return err
	}
	if version != PersistedSchemaVersion {
		return fmt.Errorf("%w: found %d, expected %d", ErrUnsupportedSchema, version, PersistedSchemaVersion)
	}
	return store.Validate(ctx)
}

func (s *Store) BackupTo(ctx context.Context, destination string) error {
	if strings.TrimSpace(destination) == "" {
		return errors.New("backup destination must not be empty")
	}
	if _, err := os.Stat(destination); err == nil {
		return fmt.Errorf("backup destination already exists: %s", destination)
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("stat backup destination: %w", err)
	}
	if _, err := s.db.ExecContext(ctx, `VACUUM INTO ?`, destination); err != nil {
		return fmt.Errorf("vacuum into backup: %w", err)
	}
	if err := os.Chmod(destination, 0o600); err != nil {
		return fmt.Errorf("set backup permissions: %w", err)
	}
	if err := ValidateFile(ctx, destination); err != nil {
		return fmt.Errorf("validate staged backup: %w", err)
	}
	return nil
}

func (s *Store) withTx(ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return classifyStorageError(fmt.Errorf("begin sqlite transaction: %w", err))
	}
	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return classifyStorageError(err)
	}
	if err := tx.Commit(); err != nil {
		return classifyStorageError(fmt.Errorf("commit sqlite transaction: %w", err))
	}
	return nil
}

type sqliteErrorCoder interface {
	Code() int
}

func classifyStorageError(err error) error {
	if err == nil || errors.Is(err, ErrStorageBusy) {
		return err
	}
	var coded sqliteErrorCoder
	if errors.As(err, &coded) {
		switch coded.Code() & 0xff {
		case 5, 6: // SQLITE_BUSY, SQLITE_LOCKED (including extended result codes).
			return fmt.Errorf("%w: %w", ErrStorageBusy, err)
		}
	}
	return err
}
