package database

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/seo-auditor/seo-auditor/internal/contracts"
)

type ScheduledAuditRecord struct {
	ID              contracts.ID `json:"schedule_id"`
	ProjectID       contracts.ID `json:"project_id"`
	ProfileID       contracts.ID `json:"profile_id"`
	Name            string       `json:"name"`
	IntervalSeconds int64        `json:"interval_seconds"`
	Enabled         bool         `json:"enabled"`
	NextRunAt       string       `json:"next_run_at"`
	LastRunAt       string       `json:"last_run_at,omitempty"`
	LastCrawlID     contracts.ID `json:"last_crawl_id,omitempty"`
	LastError       string       `json:"last_error,omitempty"`
	CreatedAt       string       `json:"created_at"`
	UpdatedAt       string       `json:"updated_at"`
	rowID           int64
}

func (f *Frontier) CreateSchedule(ctx context.Context, record ScheduledAuditRecord) error {
	if record.ID == "" || record.ProjectID == "" || record.ProfileID == "" || record.Name == "" || record.IntervalSeconds < 900 || record.IntervalSeconds > 2592000 {
		return errors.New("scheduled audit is invalid")
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	if record.NextRunAt == "" {
		record.NextRunAt = time.Now().UTC().Add(time.Duration(record.IntervalSeconds) * time.Second).Format(time.RFC3339Nano)
	}
	return f.writer.Submit(ctx, func(ctx context.Context, tx *sql.Tx) error {
		result, err := tx.ExecContext(ctx, `INSERT INTO scheduled_audit(id,project_id,profile_id,name,interval_seconds,enabled,next_run_at,created_at,updated_at) SELECT ?,?,?,?,?,1,?,?,? FROM crawl_profile WHERE id=? AND project_id=?`, record.ID, record.ProjectID, record.ProfileID, record.Name, record.IntervalSeconds, record.NextRunAt, now, now, record.ProfileID, record.ProjectID)
		if err != nil {
			return err
		}
		if rows, _ := result.RowsAffected(); rows != 1 {
			return errors.New("crawl profile does not belong to scheduled audit project")
		}
		return nil
	})
}
func (f *Frontier) ListSchedules(ctx context.Context, projectID contracts.ID, page contracts.PageRequest) (contracts.Page[ScheduledAuditRecord], error) {
	return listKeyset(ctx, page, func(after int64, limit int) (*sql.Rows, error) {
		return f.db.QueryContext(ctx, `SELECT rowid,id,project_id,profile_id,name,interval_seconds,enabled,next_run_at,last_run_at,COALESCE(last_crawl_id,''),last_error,created_at,updated_at FROM scheduled_audit WHERE project_id=? AND rowid>? ORDER BY rowid LIMIT ?`, projectID, after, limit)
	}, func(rows *sql.Rows) (ScheduledAuditRecord, error) {
		var item ScheduledAuditRecord
		err := rows.Scan(&item.rowID, &item.ID, &item.ProjectID, &item.ProfileID, &item.Name, &item.IntervalSeconds, &item.Enabled, &item.NextRunAt, &item.LastRunAt, &item.LastCrawlID, &item.LastError, &item.CreatedAt, &item.UpdatedAt)
		return item, err
	}, func(item ScheduledAuditRecord) int64 { return item.rowID })
}
func (f *Frontier) DeleteSchedule(ctx context.Context, projectID, id contracts.ID) error {
	return f.writer.Submit(ctx, func(ctx context.Context, tx *sql.Tx) error {
		result, err := tx.ExecContext(ctx, `DELETE FROM scheduled_audit WHERE id=? AND project_id=?`, id, projectID)
		if err != nil {
			return err
		}
		if rows, _ := result.RowsAffected(); rows != 1 {
			return sql.ErrNoRows
		}
		return nil
	})
}
func (f *Frontier) ClaimDueSchedules(ctx context.Context, now time.Time, limit int) ([]ScheduledAuditRecord, error) {
	if limit < 1 || limit > 100 {
		return nil, errors.New("scheduled audit claim limit is invalid")
	}
	var claimed []ScheduledAuditRecord
	err := f.writer.Submit(ctx, func(ctx context.Context, tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx, `SELECT id,project_id,profile_id,name,interval_seconds,next_run_at,created_at,updated_at FROM scheduled_audit s WHERE enabled=1 AND next_run_at<=? AND NOT EXISTS(SELECT 1 FROM crawl c WHERE c.id=s.last_crawl_id AND c.status IN('pending','running','pausing','paused','cancelling')) ORDER BY next_run_at,id LIMIT ?`, now.UTC().Format(time.RFC3339Nano), limit)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var item ScheduledAuditRecord
			if err := rows.Scan(&item.ID, &item.ProjectID, &item.ProfileID, &item.Name, &item.IntervalSeconds, &item.NextRunAt, &item.CreatedAt, &item.UpdatedAt); err != nil {
				return err
			}
			item.Enabled = true
			claimed = append(claimed, item)
		}
		if err := rows.Err(); err != nil {
			return err
		}
		stamp := now.UTC().Format(time.RFC3339Nano)
		for _, item := range claimed {
			next := now.Add(time.Duration(item.IntervalSeconds) * time.Second).UTC().Format(time.RFC3339Nano)
			if _, err := tx.ExecContext(ctx, `UPDATE scheduled_audit SET next_run_at=?,last_run_at=?,last_error='',updated_at=? WHERE id=?`, next, stamp, stamp, item.ID); err != nil {
				return err
			}
		}
		return nil
	})
	return claimed, err
}
func (f *Frontier) RecordScheduleResult(ctx context.Context, id, crawlID contracts.ID, runErr string) error {
	if len(runErr) > 1000 {
		runErr = runErr[:1000]
	}
	return f.writer.Submit(ctx, func(ctx context.Context, tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, `UPDATE scheduled_audit SET last_crawl_id=?,last_error=?,updated_at=? WHERE id=?`, nullableID(crawlID), runErr, time.Now().UTC().Format(time.RFC3339Nano), id)
		return err
	})
}
