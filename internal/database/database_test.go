package database

import (
	"context"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSQLiteFileURLUsesAbsoluteURI(t *testing.T) {
	abs, err := filepath.Abs(filepath.Join(t.TempDir(), "auditor.db"))
	if err != nil {
		t.Fatal(err)
	}
	dsn := sqliteFileURL(abs, url.Values{"mode": []string{"ro"}})
	if !strings.HasPrefix(dsn, "file:///") {
		t.Fatalf("expected absolute SQLite file URI, got %q", dsn)
	}
	parsed, err := url.Parse(dsn)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.Host != "" || parsed.Query().Get("mode") != "ro" {
		t.Fatalf("unexpected SQLite file URI: %#v", parsed)
	}
}

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
	if count != 13 {
		t.Fatalf("migration count = %d, want 13", count)
	}
	var classification, source string
	if err := db.SQL().QueryRow("SELECT (SELECT dflt_value FROM pragma_table_info('issue') WHERE name='classification'), (SELECT dflt_value FROM pragma_table_info('issue') WHERE name='evidence_source')").Scan(&classification, &source); err != nil {
		t.Fatalf("issue provenance migration: %v", err)
	}
	if classification != "'review'" || source != "'raw'" {
		t.Fatalf("issue provenance defaults classification=%q source=%q", classification, source)
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

func TestOpenReadOnlyDoesNotMigrateOrWrite(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "readonly.db")
	db, err := Open(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	readonly, err := OpenReadOnly(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	defer readonly.Close()
	if _, err := readonly.SQL().ExecContext(ctx, `INSERT INTO project(id,name,created_at,updated_at) VALUES ('x','x','x','x')`); err == nil {
		t.Fatal("read-only database accepted a write")
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
