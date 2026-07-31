package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/seo-auditor/seo-auditor/internal/contracts"
)

type IntegrationObservationRecord struct {
	ID             contracts.ID    `json:"observation_id"`
	ProjectID      contracts.ID    `json:"project_id"`
	CrawlID        contracts.ID    `json:"crawl_id,omitempty"`
	Provider       string          `json:"provider"`
	EvidenceSource string          `json:"evidence_source"`
	ProfileVersion string          `json:"profile_version"`
	Scope          string          `json:"scope"`
	Freshness      string          `json:"freshness,omitempty"`
	Result         json.RawMessage `json:"result"`
	ObservedAt     string          `json:"observed_at"`
	CreatedAt      string          `json:"created_at"`
	rowID          int64
}

func (f *Frontier) RecordIntegrationObservation(ctx context.Context, record IntegrationObservationRecord) error {
	if len(record.Result) == 0 || len(record.Result) > 8<<20 || !json.Valid(record.Result) {
		return errors.New("integration result is invalid or exceeds 8 MiB")
	}
	if record.ID == "" || record.ProjectID == "" || record.Provider == "" || record.EvidenceSource == "" || record.ProfileVersion == "" || record.Scope == "" {
		return errors.New("integration observation metadata is incomplete")
	}
	if record.CreatedAt == "" {
		record.CreatedAt = time.Now().UTC().Format(time.RFC3339Nano)
	}
	return f.writer.Submit(ctx, func(ctx context.Context, tx *sql.Tx) error {
		if record.CrawlID != "" {
			var count int
			if err := tx.QueryRowContext(ctx, `SELECT count(*) FROM crawl WHERE id=? AND project_id=?`, record.CrawlID, record.ProjectID).Scan(&count); err != nil {
				return err
			}
			if count != 1 {
				return errors.New("crawl does not belong to integration project")
			}
		}
		_, err := tx.ExecContext(ctx, `INSERT INTO integration_observation(id,project_id,crawl_id,provider,evidence_source,profile_version,scope,freshness,result_json,observed_at,created_at) VALUES(?,?,?,?,?,?,?,?,?,?,?)`, record.ID, record.ProjectID, nullableID(record.CrawlID), record.Provider, record.EvidenceSource, record.ProfileVersion, record.Scope, record.Freshness, string(record.Result), record.ObservedAt, record.CreatedAt)
		return err
	})
}
func (f *Frontier) ListIntegrationObservations(ctx context.Context, projectID contracts.ID, provider string, page contracts.PageRequest) (contracts.Page[IntegrationObservationRecord], error) {
	return listKeyset(ctx, page, func(after int64, limit int) (*sql.Rows, error) {
		if provider == "" {
			return f.db.QueryContext(ctx, `SELECT rowid,id,project_id,COALESCE(crawl_id,''),provider,evidence_source,profile_version,scope,freshness,result_json,observed_at,created_at FROM integration_observation WHERE project_id=? AND rowid>? ORDER BY rowid LIMIT ?`, projectID, after, limit)
		}
		return f.db.QueryContext(ctx, `SELECT rowid,id,project_id,COALESCE(crawl_id,''),provider,evidence_source,profile_version,scope,freshness,result_json,observed_at,created_at FROM integration_observation WHERE project_id=? AND provider=? AND rowid>? ORDER BY rowid LIMIT ?`, projectID, provider, after, limit)
	}, func(rows *sql.Rows) (IntegrationObservationRecord, error) {
		var item IntegrationObservationRecord
		var result string
		err := rows.Scan(&item.rowID, &item.ID, &item.ProjectID, &item.CrawlID, &item.Provider, &item.EvidenceSource, &item.ProfileVersion, &item.Scope, &item.Freshness, &result, &item.ObservedAt, &item.CreatedAt)
		item.Result = json.RawMessage(result)
		return item, err
	}, func(item IntegrationObservationRecord) int64 { return item.rowID })
}
