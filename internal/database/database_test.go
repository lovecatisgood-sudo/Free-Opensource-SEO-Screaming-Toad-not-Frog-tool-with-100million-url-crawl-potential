package database

import (
	"context"
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
	if count != 1 {
		t.Fatalf("migration count = %d, want 1", count)
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
