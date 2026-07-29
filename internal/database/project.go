package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/seo-auditor/seo-auditor/internal/contracts"
)

type ProjectRecord struct {
	ID        contracts.ID `json:"project_id"`
	Name      string       `json:"name"`
	CreatedAt string       `json:"created_at"`
	UpdatedAt string       `json:"updated_at"`
	Archived  bool         `json:"archived"`
	rowID     int64
}
type ProfileRecord struct {
	ID            contracts.ID                 `json:"profile_id"`
	ProjectID     contracts.ID                 `json:"project_id"`
	Version       int                          `json:"version"`
	Name          string                       `json:"name"`
	Configuration contracts.CrawlConfiguration `json:"configuration"`
	CreatedAt     string                       `json:"created_at"`
	rowID         int64
}

func (f *Frontier) ProjectExists(ctx context.Context, id contracts.ID) error {
	var count int
	if err := f.db.QueryRowContext(ctx, `SELECT count(*) FROM project WHERE id=? AND deleted_at IS NULL`, id).Scan(&count); err != nil {
		return err
	}
	if count != 1 {
		return sql.ErrNoRows
	}
	return nil
}
func (f *Frontier) GetProject(ctx context.Context, id contracts.ID) (ProjectRecord, error) {
	var item ProjectRecord
	err := f.db.QueryRowContext(ctx, `SELECT rowid,id,name,created_at,updated_at,archived_at IS NOT NULL FROM project WHERE id=? AND deleted_at IS NULL`, id).Scan(&item.rowID, &item.ID, &item.Name, &item.CreatedAt, &item.UpdatedAt, &item.Archived)
	return item, err
}
func (f *Frontier) ListProjects(ctx context.Context, page contracts.PageRequest) (contracts.Page[ProjectRecord], error) {
	return listKeyset(ctx, page, func(after int64, limit int) (*sql.Rows, error) {
		return f.db.QueryContext(ctx, `SELECT rowid,id,name,created_at,updated_at,archived_at IS NOT NULL FROM project WHERE deleted_at IS NULL AND rowid>? ORDER BY rowid LIMIT ?`, after, limit)
	}, func(rows *sql.Rows) (ProjectRecord, error) {
		var item ProjectRecord
		err := rows.Scan(&item.rowID, &item.ID, &item.Name, &item.CreatedAt, &item.UpdatedAt, &item.Archived)
		return item, err
	}, func(item ProjectRecord) int64 { return item.rowID })
}
func (f *Frontier) RenameProject(ctx context.Context, id contracts.ID, name string) error {
	if name == "" || len(name) > 200 {
		return errors.New("project name must contain 1 to 200 characters")
	}
	return f.writer.Submit(ctx, func(ctx context.Context, tx *sql.Tx) error {
		result, err := tx.ExecContext(ctx, `UPDATE project SET name=?,updated_at=? WHERE id=? AND deleted_at IS NULL`, name, time.Now().UTC().Format(time.RFC3339Nano), id)
		if err != nil {
			return err
		}
		changed, _ := result.RowsAffected()
		if changed != 1 {
			return sql.ErrNoRows
		}
		return nil
	})
}
func (f *Frontier) SetProjectArchived(ctx context.Context, id contracts.ID, archived bool) error {
	return f.writer.Submit(ctx, func(ctx context.Context, tx *sql.Tx) error {
		var value any
		if archived {
			value = time.Now().UTC().Format(time.RFC3339Nano)
		}
		result, err := tx.ExecContext(ctx, `UPDATE project SET archived_at=?,updated_at=? WHERE id=? AND deleted_at IS NULL`, value, time.Now().UTC().Format(time.RFC3339Nano), id)
		if err != nil {
			return err
		}
		changed, _ := result.RowsAffected()
		if changed != 1 {
			return sql.ErrNoRows
		}
		return nil
	})
}
func (f *Frontier) TrashProject(ctx context.Context, id contracts.ID) error {
	return f.writer.Submit(ctx, func(ctx context.Context, tx *sql.Tx) error {
		var active int
		if err := tx.QueryRowContext(ctx, `SELECT count(*) FROM crawl WHERE project_id=? AND status IN ('pending','running','pausing','cancelling')`, id).Scan(&active); err != nil {
			return err
		}
		if active > 0 {
			return errors.New("project has an active crawl")
		}
		now := time.Now().UTC().Format(time.RFC3339Nano)
		result, err := tx.ExecContext(ctx, `UPDATE project SET deleted_at=?,updated_at=? WHERE id=? AND deleted_at IS NULL`, now, now, id)
		if err != nil {
			return err
		}
		changed, _ := result.RowsAffected()
		if changed != 1 {
			return sql.ErrNoRows
		}
		return nil
	})
}
func (f *Frontier) RestoreProject(ctx context.Context, id contracts.ID) error {
	return f.writer.Submit(ctx, func(ctx context.Context, tx *sql.Tx) error {
		now := time.Now().UTC().Format(time.RFC3339Nano)
		result, err := tx.ExecContext(ctx, `UPDATE project SET deleted_at=NULL,updated_at=? WHERE id=? AND deleted_at IS NOT NULL`, now, id)
		if err != nil {
			return err
		}
		changed, _ := result.RowsAffected()
		if changed != 1 {
			return sql.ErrNoRows
		}
		return nil
	})
}

func (f *Frontier) CreateProfile(ctx context.Context, id, projectID contracts.ID, name string, configuration contracts.CrawlConfiguration) (ProfileRecord, error) {
	var result ProfileRecord
	if name == "" || len(name) > 200 {
		return result, errors.New("profile name must contain 1 to 200 characters")
	}
	raw, err := json.Marshal(configuration)
	if err != nil {
		return result, err
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	err = f.writer.Submit(ctx, func(ctx context.Context, tx *sql.Tx) error {
		var version int
		if err := tx.QueryRowContext(ctx, `SELECT COALESCE(max(version),0)+1 FROM crawl_profile WHERE project_id=?`, projectID).Scan(&version); err != nil {
			return err
		}
		_, err := tx.ExecContext(ctx, `INSERT INTO crawl_profile(id,project_id,version,name,config_json,created_at) VALUES (?,?,?,?,?,?)`, id, projectID, version, name, string(raw), now)
		if err == nil {
			result = ProfileRecord{ID: id, ProjectID: projectID, Version: version, Name: name, Configuration: configuration, CreatedAt: now}
		}
		return err
	})
	return result, err
}
func (f *Frontier) GetProfile(ctx context.Context, projectID, id contracts.ID) (ProfileRecord, error) {
	var result ProfileRecord
	var raw string
	err := f.db.QueryRowContext(ctx, `SELECT id,project_id,version,name,config_json,created_at FROM crawl_profile WHERE id=? AND project_id=?`, id, projectID).Scan(&result.ID, &result.ProjectID, &result.Version, &result.Name, &raw, &result.CreatedAt)
	if err != nil {
		return result, err
	}
	err = json.Unmarshal([]byte(raw), &result.Configuration)
	return result, err
}
func (f *Frontier) ListProfiles(ctx context.Context, projectID contracts.ID, page contracts.PageRequest) (contracts.Page[ProfileRecord], error) {
	return listKeyset(ctx, page, func(after int64, limit int) (*sql.Rows, error) {
		return f.db.QueryContext(ctx, `SELECT rowid,id,project_id,version,name,config_json,created_at FROM crawl_profile WHERE project_id=? AND rowid>? ORDER BY rowid LIMIT ?`, projectID, after, limit)
	}, func(rows *sql.Rows) (ProfileRecord, error) {
		var raw string
		var item ProfileRecord
		err := rows.Scan(&item.rowID, &item.ID, &item.ProjectID, &item.Version, &item.Name, &raw, &item.CreatedAt)
		if err == nil {
			err = json.Unmarshal([]byte(raw), &item.Configuration)
		}
		return item, err
	}, func(item ProfileRecord) int64 { return item.rowID })
}
