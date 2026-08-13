package sqlitespike

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()
	store, err := Open(filepath.Join(t.TempDir(), "calathea.db"))
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })
	if err := store.InitializeLatest(context.Background()); err != nil {
		t.Fatalf("InitializeLatest() error = %v", err)
	}
	return store
}

func writeRequest(operationID, entityID, recordID, payload string, expected, next int64) WriteRequest {
	return WriteRequest{
		OperationID:      operationID,
		EntityID:         entityID,
		RecordID:         recordID,
		Payload:          payload,
		ExpectedRevision: expected,
		NewRevision:      next,
	}
}

func TestAuthoritativeAtomicityAndIdempotentLostResponseRetry(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)

	request := writeRequest("op-1", "project-1", "record-1", "payload-v1", 0, 1)
	request.InjectBeforeCommit = true
	if _, err := store.Apply(ctx, request); !errors.Is(err, ErrInjectedFailure) {
		t.Fatalf("Apply(injected failure) error = %v, want ErrInjectedFailure", err)
	}
	records, currents, operations, err := store.Counts(ctx)
	if err != nil {
		t.Fatalf("Counts() error = %v", err)
	}
	if records != 0 || currents != 0 || operations != 0 {
		t.Fatalf("rolled-back counts = records:%d currents:%d operations:%d, want all zero", records, currents, operations)
	}

	request.InjectBeforeCommit = false
	result, err := store.Apply(ctx, request)
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}
	if result.RecordID != "record-1" || result.Replayed {
		t.Fatalf("Apply() result = %#v", result)
	}

	retry, err := store.Apply(ctx, request)
	if err != nil {
		t.Fatalf("Apply(retry) error = %v", err)
	}
	if retry.RecordID != result.RecordID || !retry.Replayed {
		t.Fatalf("retry result = %#v, first = %#v", retry, result)
	}
	records, currents, operations, err = store.Counts(ctx)
	if err != nil {
		t.Fatalf("Counts() after retry error = %v", err)
	}
	if records != 1 || currents != 1 || operations != 1 {
		t.Fatalf("retry duplicated durability: records:%d currents:%d operations:%d", records, currents, operations)
	}

	conflict := request
	conflict.Payload = "materially-different"
	if _, err := store.Apply(ctx, conflict); !errors.Is(err, ErrIdempotencyConflict) {
		t.Fatalf("Apply(conflicting retry) error = %v, want ErrIdempotencyConflict", err)
	}
}

func TestExpectedRevisionConflictIsSemanticNotLastWriteWins(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "concurrency.db")
	first, err := Open(path)
	if err != nil {
		t.Fatalf("Open(first) error = %v", err)
	}
	defer first.Close()
	if err := first.InitializeLatest(ctx); err != nil {
		t.Fatalf("InitializeLatest() error = %v", err)
	}
	second, err := openExisting(path)
	if err != nil {
		t.Fatalf("openExisting(second) error = %v", err)
	}
	defer second.Close()

	if _, err := first.Apply(ctx, writeRequest("op-base", "project-1", "record-1", "v1", 0, 1)); err != nil {
		t.Fatalf("initial Apply() error = %v", err)
	}
	lockTx, err := first.db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("BeginTx(lock holder) error = %v", err)
	}
	if _, err := lockTx.ExecContext(ctx, `UPDATE current_records SET revision = revision WHERE entity_id = ?`, "project-1"); err != nil {
		_ = lockTx.Rollback()
		t.Fatalf("acquire write lock error = %v", err)
	}
	contended := writeRequest("op-contended", "project-1", "record-contended", "contended", 1, 2)
	if _, err := second.Apply(ctx, contended); !errors.Is(err, ErrStorageBusy) {
		_ = lockTx.Rollback()
		t.Fatalf("contended Apply() error = %v, want ErrStorageBusy", err)
	}
	if err := lockTx.Rollback(); err != nil {
		t.Fatalf("Rollback(lock holder) error = %v", err)
	}

	staleA := writeRequest("op-a", "project-1", "record-2a", "v2-a", 1, 2)
	staleB := writeRequest("op-b", "project-1", "record-2b", "v2-b", 1, 2)
	if _, err := first.Apply(ctx, staleA); err != nil {
		t.Fatalf("first stale writer error = %v", err)
	}
	if _, err := second.Apply(ctx, staleB); !errors.Is(err, ErrExpectedConflict) {
		t.Fatalf("second stale writer error = %v, want ErrExpectedConflict", err)
	}
	records, _, operations, err := first.Counts(ctx)
	if err != nil {
		t.Fatalf("Counts() error = %v", err)
	}
	if records != 2 || operations != 2 {
		t.Fatalf("stale writer persisted: records:%d operations:%d, want 2/2", records, operations)
	}
}

func TestProjectionRebuildUsesAuthoritativeHistoryOnly(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)
	if _, err := store.Apply(ctx, writeRequest("op-1", "project-1", "record-1", "one", 0, 1)); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Apply(ctx, writeRequest("op-2", "project-2", "record-2", "two", 0, 1)); err != nil {
		t.Fatal(err)
	}
	if err := store.RebuildProjection(ctx); err != nil {
		t.Fatalf("RebuildProjection() error = %v", err)
	}
	want := []ProjectionRow{
		{EntityID: "project-1", RecordID: "record-1", Payload: "one", SourceRevision: 1},
		{EntityID: "project-2", RecordID: "record-2", Payload: "two", SourceRevision: 1},
	}
	got, err := store.Projection(ctx)
	if err != nil {
		t.Fatalf("Projection() error = %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Projection() = %#v, want %#v", got, want)
	}
	if err := store.CorruptProjectionForTest(ctx, "project-1"); err != nil {
		t.Fatalf("CorruptProjectionForTest() error = %v", err)
	}
	if err := store.RebuildProjection(ctx); err != nil {
		t.Fatalf("second RebuildProjection() error = %v", err)
	}
	got, err = store.Projection(ctx)
	if err != nil {
		t.Fatalf("Projection() after rebuild error = %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("rebuilt Projection() = %#v, want %#v", got, want)
	}
	records, _, operations, err := store.Counts(ctx)
	if err != nil {
		t.Fatalf("Counts() error = %v", err)
	}
	if records != 2 || operations != 2 {
		t.Fatalf("projection rebuild changed authoritative counts: records:%d operations:%d", records, operations)
	}
}

func TestForwardMigrationAndUnsupportedNewerSchema(t *testing.T) {
	ctx := context.Background()
	store, err := Open(filepath.Join(t.TempDir(), "migration.db"))
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer store.Close()
	if err := store.InitializeV0(ctx); err != nil {
		t.Fatalf("InitializeV0() error = %v", err)
	}
	if version, err := store.SchemaVersion(ctx); err != nil || version != 0 {
		t.Fatalf("SchemaVersion() = %d, %v, want 0", version, err)
	}
	if err := store.MigrateToLatest(ctx); err != nil {
		t.Fatalf("MigrateToLatest() error = %v", err)
	}
	if err := store.MigrateToLatest(ctx); err != nil {
		t.Fatalf("idempotent MigrateToLatest() error = %v", err)
	}
	if version, err := store.SchemaVersion(ctx); err != nil || version != PersistedSchemaVersion {
		t.Fatalf("SchemaVersion() = %d, %v, want %d", version, err, PersistedSchemaVersion)
	}
	if err := store.SetSchemaVersionForTest(ctx, PersistedSchemaVersion+1); err != nil {
		t.Fatalf("SetSchemaVersionForTest() error = %v", err)
	}
	if err := store.MigrateToLatest(ctx); !errors.Is(err, ErrUnsupportedSchema) {
		t.Fatalf("MigrateToLatest(newer) error = %v, want ErrUnsupportedSchema", err)
	}
}

func TestIntegrityLayersDistinguishSQLiteFromDomainHistory(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)
	if err := store.InsertSemanticMismatchForTest(ctx); err != nil {
		t.Fatalf("InsertSemanticMismatchForTest() error = %v", err)
	}
	if err := validateSQLiteIntegrity(ctx, store.db); err != nil {
		t.Fatalf("structurally valid database failed integrity_check: %v", err)
	}
	if err := validateForeignKeys(ctx, store.db); err != nil {
		t.Fatalf("foreign-key-valid semantic mismatch failed foreign_key_check: %v", err)
	}
	if err := store.Validate(ctx); !errors.Is(err, ErrDomainIntegrity) {
		t.Fatalf("Validate() error = %v, want ErrDomainIntegrity", err)
	}
}

func TestCorruptTruncatedCandidateFailsWithoutChangingCurrentStore(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)
	if _, err := store.Apply(ctx, writeRequest("op-1", "project-1", "record-1", "safe", 0, 1)); err != nil {
		t.Fatal(err)
	}
	if err := store.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	data, err := os.ReadFile(store.Path())
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if len(data) < 2 {
		t.Fatalf("sqlite file unexpectedly small: %d", len(data))
	}
	candidate := filepath.Join(t.TempDir(), "truncated.db")
	if err := os.WriteFile(candidate, data[:len(data)/2], 0o600); err != nil {
		t.Fatalf("WriteFile(truncated) error = %v", err)
	}
	if err := ValidateFile(ctx, candidate); err == nil {
		t.Fatal("ValidateFile(truncated) succeeded")
	}

	reopened, err := openExisting(store.Path())
	if err != nil {
		t.Fatalf("reopen original error = %v", err)
	}
	defer reopened.Close()
	if err := reopened.Validate(ctx); err != nil {
		t.Fatalf("original store changed after corrupt candidate validation: %v", err)
	}
	records, _, _, err := reopened.Counts(ctx)
	if err != nil || records != 1 {
		t.Fatalf("original record count = %d, %v, want 1", records, err)
	}
}

func TestVacuumIntoBackupIsIndependentAndValidated(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)
	if _, err := store.Apply(ctx, writeRequest("op-1", "project-1", "record-1", "v1", 0, 1)); err != nil {
		t.Fatal(err)
	}
	backup := filepath.Join(t.TempDir(), "backup.db")
	if err := store.BackupTo(ctx, backup); err != nil {
		t.Fatalf("BackupTo() error = %v", err)
	}
	if _, err := store.Apply(ctx, writeRequest("op-2", "project-1", "record-2", "v2", 1, 2)); err != nil {
		t.Fatal(err)
	}
	candidate, err := openExisting(backup)
	if err != nil {
		t.Fatalf("openExisting(backup) error = %v", err)
	}
	defer candidate.Close()
	if err := candidate.Validate(ctx); err != nil {
		t.Fatalf("backup Validate() error = %v", err)
	}
	records, _, operations, err := candidate.Counts(ctx)
	if err != nil {
		t.Fatalf("backup Counts() error = %v", err)
	}
	if records != 1 || operations != 1 {
		t.Fatalf("backup mutated with source: records:%d operations:%d, want 1/1", records, operations)
	}
}

func TestConnectionBaselineUsesForeignKeysRollbackJournalAndPrivateFile(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)
	enabled, err := store.ForeignKeysEnabled(ctx)
	if err != nil || !enabled {
		t.Fatalf("ForeignKeysEnabled() = %v, %v, want true", enabled, err)
	}
	mode, err := store.JournalMode(ctx)
	if err != nil {
		t.Fatalf("JournalMode() error = %v", err)
	}
	if mode != "delete" {
		t.Fatalf("JournalMode() = %q, want delete", mode)
	}
	synchronous, err := store.SynchronousLevel(ctx)
	if err != nil {
		t.Fatalf("SynchronousLevel() error = %v", err)
	}
	if synchronous != 2 {
		t.Fatalf("SynchronousLevel() = %d, want FULL(2)", synchronous)
	}
	info, err := os.Stat(store.Path())
	if err != nil {
		t.Fatalf("Stat() error = %v", err)
	}
	if info.Mode().Perm()&0o077 != 0 {
		t.Fatalf("database permissions = %o, want no group/other bits", info.Mode().Perm())
	}
}
