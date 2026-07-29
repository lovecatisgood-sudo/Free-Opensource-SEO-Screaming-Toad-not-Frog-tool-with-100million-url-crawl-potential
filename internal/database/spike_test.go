package database

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

// TestSynthetic100K documents the foundation capacity spike and remains in the
// normal suite so schema or driver regressions cannot silently invalidate it.
func TestSynthetic100K(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping 100k persistence spike in short mode")
	}
	ctx := context.Background()
	db, err := Open(ctx, filepath.Join(t.TempDir(), "auditor.db"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	writer := NewWriter(db.SQL(), 1)
	t.Cleanup(writer.Close)

	started := time.Now()
	err = writer.Submit(ctx, func(ctx context.Context, tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx, `INSERT INTO project(id, name, created_at, updated_at)
            VALUES ('project_spike', 'spike', '2026-01-01T00:00:00Z', '2026-01-01T00:00:00Z')`); err != nil {
			return err
		}
		stmt, err := tx.PrepareContext(ctx, `INSERT INTO url(project_id, request_key, original_url, scheme, host, path, created_at)
            VALUES ('project_spike', ?, ?, 'https', 'example.com', ?, '2026-01-01T00:00:00Z')`)
		if err != nil {
			return err
		}
		defer stmt.Close()
		for i := 0; i < 100_000; i++ {
			path := fmt.Sprintf("/page/%06d", i)
			url := "https://example.com" + path
			if _, err := stmt.ExecContext(ctx, url, url, path); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("insert spike: %v", err)
	}
	var count int
	if err := db.SQL().QueryRowContext(ctx, "SELECT count(*) FROM url").Scan(&count); err != nil {
		t.Fatalf("count: %v", err)
	}
	if count != 100_000 {
		t.Fatalf("count = %d", count)
	}
	var memory runtime.MemStats
	runtime.ReadMemStats(&memory)
	t.Logf("inserted=%d elapsed=%s heap_alloc=%dMiB", count, time.Since(started).Round(time.Millisecond), memory.Alloc/(1024*1024))
}
