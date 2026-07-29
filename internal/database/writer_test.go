package database

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"sync"
	"testing"
)

func TestWriterSerializesTransactionsAndRollsBack(t *testing.T) {
	t.Parallel()

	db, err := Open(context.Background(), filepath.Join(t.TempDir(), "auditor.db"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	writer := NewWriter(db.SQL(), 8)
	t.Cleanup(writer.Close)

	const count = 20
	var wg sync.WaitGroup
	errs := make(chan error, count)
	for i := 0; i < count; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			errs <- writer.Submit(context.Background(), func(ctx context.Context, tx *sql.Tx) error {
				_, err := tx.ExecContext(ctx, `INSERT INTO project(id, name, created_at, updated_at)
                    VALUES (lower(hex(randomblob(16))), 'test', '2026-01-01T00:00:00Z', '2026-01-01T00:00:00Z')`)
				return err
			})
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatalf("Submit: %v", err)
		}
	}

	wantRollback := errors.New("rollback")
	if err := writer.Submit(context.Background(), func(context.Context, *sql.Tx) error { return wantRollback }); !errors.Is(err, wantRollback) {
		t.Fatalf("rollback error = %v", err)
	}
	var got int
	if err := db.SQL().QueryRow("SELECT count(*) FROM project").Scan(&got); err != nil {
		t.Fatalf("count: %v", err)
	}
	if got != count {
		t.Fatalf("project count = %d, want %d", got, count)
	}
}

func TestWriterRejectsAfterClose(t *testing.T) {
	t.Parallel()

	db, err := Open(context.Background(), filepath.Join(t.TempDir(), "auditor.db"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	writer := NewWriter(db.SQL(), 1)
	writer.Close()
	if err := writer.Submit(context.Background(), func(context.Context, *sql.Tx) error { return nil }); !errors.Is(err, ErrWriterClosed) {
		t.Fatalf("Submit error = %v", err)
	}
}
