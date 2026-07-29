package database

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/seo-auditor/seo-auditor/internal/contracts"
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
		if _, err := tx.ExecContext(ctx, `INSERT INTO crawl(id,project_id,seed_url,config_json,status,discovered_count,fetched_count,analysed_count,created_at,updated_at,finished_at) VALUES ('crawl_spike','project_spike','https://example.com/','{}','completed',100000,100000,100000,'2026-01-01T00:00:00Z','2026-01-01T00:10:00Z','2026-01-01T00:10:00Z')`); err != nil {
			return err
		}
		stmt, err := tx.PrepareContext(ctx, `INSERT INTO url(project_id, request_key, original_url, scheme, host, path, created_at)
            VALUES ('project_spike', ?, ?, 'https', 'example.com', ?, '2026-01-01T00:00:00Z')`)
		if err != nil {
			return err
		}
		defer stmt.Close()
		crawlURL, err := tx.PrepareContext(ctx, `INSERT INTO crawl_url(crawl_id,url_id,state,depth,discovery_kind,attempt_count,created_at,updated_at) VALUES ('crawl_spike',?,'analysed',1,'fixture',1,'2026-01-01T00:00:00Z','2026-01-01T00:00:01Z')`)
		if err != nil {
			return err
		}
		defer crawlURL.Close()
		fetch, err := tx.PrepareContext(ctx, `INSERT INTO fetch_attempt(crawl_url_id,attempt,started_at,finished_at,status_code,content_type,decoded_bytes) VALUES (?,1,'2026-01-01T00:00:00Z','2026-01-01T00:00:01Z',200,'text/html',1000)`)
		if err != nil {
			return err
		}
		defer fetch.Close()
		page, err := tx.PrepareContext(ctx, `INSERT INTO page(crawl_url_id,extraction_mode,title,meta_description,robots_directives,language,text_length,content_hash,extracted_at,social_json) VALUES (?,'raw',?,'Description','','en',100,?,'2026-01-01T00:00:01Z','{}')`)
		if err != nil {
			return err
		}
		defer page.Close()
		for i := 0; i < 100_000; i++ {
			path := fmt.Sprintf("/page/%06d", i)
			url := "https://example.com" + path
			result, err := stmt.ExecContext(ctx, url, url, path)
			if err != nil {
				return err
			}
			urlID, err := result.LastInsertId()
			if err != nil {
				return err
			}
			result, err = crawlURL.ExecContext(ctx, urlID)
			if err != nil {
				return err
			}
			crawlURLID, err := result.LastInsertId()
			if err != nil {
				return err
			}
			if _, err := fetch.ExecContext(ctx, crawlURLID); err != nil {
				return err
			}
			if _, err := page.ExecContext(ctx, crawlURLID, "Page "+fmt.Sprint(i), fmt.Sprintf("hash-%06d", i)); err != nil {
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
	frontier := NewFrontier(db, 8)
	defer frontier.Close()
	queryStarted := time.Now()
	resultPage, err := frontier.ListPages(ctx, "crawl_spike", contracts.PageRequest{Limit: 1000})
	if err != nil {
		t.Fatal(err)
	}
	if len(resultPage.Items) != 1000 || resultPage.NextCursor == "" {
		t.Fatalf("page=%+v", resultPage)
	}
	var memory runtime.MemStats
	runtime.ReadMemStats(&memory)
	t.Logf("inserted=%d elapsed=%s first_1000_query=%s heap_alloc=%dMiB", count, time.Since(started).Round(time.Millisecond), time.Since(queryStarted).Round(time.Millisecond), memory.Alloc/(1024*1024))
}
