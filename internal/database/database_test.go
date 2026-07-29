package database

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestOpenMigratesAndEnablesWAL(t *testing.T) {
	t.Parallel()

	db, err := Open(context.Background(), filepath.Join(t.TempDir(), "auditor.db"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	mode, err := db.JournalMode(context.Background())
	if err != nil {
		t.Fatalf("JournalMode: %v", err)
	}
	if mode != "wal" {
		t.Fatalf("journal mode = %q, want wal", mode)
	}
	var count int
	if err := db.SQL().QueryRow("SELECT count(*) FROM schema_migration").Scan(&count); err != nil {
		t.Fatalf("migration count: %v", err)
	}
	if count != 7 {
		t.Fatalf("migration count = %d, want 7", count)
	}
}

func TestVerifyAndBackupProduceReadableSnapshot(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	directory := t.TempDir()
	db, err := Open(ctx, filepath.Join(directory, "source.db"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.SQL().ExecContext(ctx, `INSERT INTO project(id,name,created_at,updated_at) VALUES ('project_backup','Backup','2026-01-01T00:00:00Z','2026-01-01T00:00:00Z')`); err != nil {
		t.Fatal(err)
	}
	if err := db.Verify(ctx); err != nil {
		t.Fatal(err)
	}
	destination := filepath.Join(directory, "snapshot.sqlite3")
	if err := db.Backup(ctx, destination); err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(destination)
	if err != nil || info.Size() == 0 {
		t.Fatalf("backup info=%v err=%v", info, err)
	}
	backup, err := Open(ctx, destination)
	if err != nil {
		t.Fatal(err)
	}
	defer backup.Close()
	var count int
	if err := backup.SQL().QueryRowContext(ctx, `SELECT count(*) FROM project WHERE id='project_backup'`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("project count=%d", count)
	}
}

func TestForeignKeysAreEnforced(t *testing.T) {
	t.Parallel()

	db, err := Open(context.Background(), filepath.Join(t.TempDir(), "auditor.db"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	_, err = db.SQL().Exec(`INSERT INTO crawl(id, project_id, seed_url, config_json, status, created_at, updated_at)
        VALUES ('crawl_test', 'missing', 'https://example.com/', '{}', 'pending', '2026-01-01T00:00:00Z', '2026-01-01T00:00:00Z')`)
	if err == nil {
		t.Fatal("expected foreign key violation")
	}
}
