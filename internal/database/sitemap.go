package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/seo-auditor/seo-auditor/internal/contracts"
)

type SitemapRecord struct {
	URL, Status, DiscoveredFrom string
	Entries                     []string
}

func (f *Frontier) RecordSitemaps(ctx context.Context, crawlID, projectID contracts.ID, records []SitemapRecord) error {
	if len(records) == 0 {
		return nil
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	return f.writer.Submit(ctx, func(ctx context.Context, tx *sql.Tx) error {
		for _, record := range records {
			if record.URL == "" || record.Status == "" {
				return fmt.Errorf("invalid sitemap record")
			}
			_, err := tx.ExecContext(ctx, `INSERT INTO sitemap(crawl_id, url, status, discovered_from) VALUES (?, ?, ?, ?) ON CONFLICT(crawl_id, url) DO UPDATE SET status=excluded.status`, crawlID, record.URL, record.Status, record.DiscoveredFrom)
			if err != nil {
				return err
			}
			var sitemapID int64
			if err := tx.QueryRowContext(ctx, "SELECT id FROM sitemap WHERE crawl_id=? AND url=?", crawlID, record.URL).Scan(&sitemapID); err != nil {
				return err
			}
			for _, raw := range record.Entries {
				urlID, err := ensureURL(ctx, tx, projectID, raw, now)
				if err != nil {
					continue
				}
				if _, err := tx.ExecContext(ctx, "INSERT OR IGNORE INTO sitemap_entry(sitemap_id, url_id) VALUES (?, ?)", sitemapID, urlID); err != nil {
					return err
				}
			}
		}
		return nil
	})
}
