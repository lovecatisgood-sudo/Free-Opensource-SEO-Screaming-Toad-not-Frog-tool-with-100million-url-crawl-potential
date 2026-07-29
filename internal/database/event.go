package database

import (
	"context"
	"database/sql"
	"errors"

	"github.com/seo-auditor/seo-auditor/internal/contracts"
)

type CrawlEventRecord struct {
	ID          int64  `json:"id"`
	CrawlID     string `json:"crawl_id"`
	Event       string `json:"event"`
	Severity    string `json:"severity"`
	DetailsJSON string `json:"details_json"`
	CreatedAt   string `json:"created_at"`
}

func (f *Frontier) ListEvents(ctx context.Context, crawlID contracts.ID, page contracts.PageRequest) (contracts.Page[CrawlEventRecord], error) {
	if page.Search != "" || page.Severity != "" || page.RuleID != "" || (page.Sort != "" && page.Sort != "id") {
		return contracts.Page[CrawlEventRecord]{}, errors.New("event pagination supports only id order")
	}
	return listKeyset(ctx, page, func(after int64, limit int) (*sql.Rows, error) {
		return f.db.QueryContext(ctx, `SELECT id,crawl_id,event_type,CASE WHEN event_type='recovered' OR (event_type='status_changed' AND (json_extract(payload_json,'$.to')='failed' OR json_extract(payload_json,'$.to')='limit_reached')) THEN 'warning' ELSE 'info' END,payload_json,created_at FROM crawl_event WHERE crawl_id=? AND id>? ORDER BY id LIMIT ?`, crawlID, after, limit)
	}, func(rows *sql.Rows) (CrawlEventRecord, error) {
		var item CrawlEventRecord
		err := rows.Scan(&item.ID, &item.CrawlID, &item.Event, &item.Severity, &item.DetailsJSON, &item.CreatedAt)
		return item, err
	}, func(item CrawlEventRecord) int64 { return item.ID })
}
