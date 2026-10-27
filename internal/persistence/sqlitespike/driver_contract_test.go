package sqlitespike

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestPinnedSQLiteRuntimeVersion(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)
	version, err := store.SQLiteVersion(ctx)
	if err != nil {
		t.Fatalf("SQLiteVersion() error = %v", err)
	}
	if runtime.GOOS == "linux" && runtime.GOARCH == "amd64" {
		const expected = "3.53.3"
		if version != expected {
			t.Fatalf("SQLiteVersion() = %q, want pinned Linux/amd64 runtime %q", version, expected)
		}
	}
}

func TestBackupPathIsBoundAndExistingCandidateIsNeverOverwritten(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)
	if _, err := store.Apply(ctx, writeRequest("op-1", "project-1", "record-1", "v1", 0, 1)); err != nil {
		t.Fatalf("Apply() error = %v", err)
	}

	quoted := filepath.Join(t.TempDir(), "backup-'quoted'.db")
	if err := store.BackupTo(ctx, quoted); err != nil {
		t.Fatalf("BackupTo(quoted path) error = %v", err)
	}
	if err := ValidateFile(ctx, quoted); err != nil {
		t.Fatalf("ValidateFile(quoted backup) error = %v", err)
	}

	incomplete := filepath.Join(t.TempDir(), "incomplete.db")
	const marker = "partial-backup-candidate"
	if err := os.WriteFile(incomplete, []byte(marker), 0o600); err != nil {
		t.Fatalf("WriteFile(incomplete) error = %v", err)
	}
	if err := store.BackupTo(ctx, incomplete); err == nil {
		t.Fatal("BackupTo(existing incomplete candidate) succeeded")
	}
	data, err := os.ReadFile(incomplete)
	if err != nil {
		t.Fatalf("ReadFile(incomplete) error = %v", err)
	}
	if string(data) != marker {
		t.Fatalf("existing backup candidate was modified: got %q, want %q", string(data), marker)
	}
}
