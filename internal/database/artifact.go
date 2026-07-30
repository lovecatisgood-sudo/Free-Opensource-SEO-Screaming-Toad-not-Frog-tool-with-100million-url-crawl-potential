package database

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/seo-auditor/seo-auditor/internal/contracts"
)

type ArtifactRecord struct {
	ID           contracts.ID `json:"artifact_id"`
	CrawlID      contracts.ID `json:"crawl_id"`
	Format       string       `json:"format"`
	RelativePath string       `json:"relative_path"`
	Checksum     string       `json:"sha256"`
	SizeBytes    int64        `json:"size_bytes"`
	CreatedAt    string       `json:"created_at"`
	ExpiresAt    string       `json:"expires_at,omitempty"`
}
type PageArtifactRecord struct {
	ArtifactRecord
	CrawlURLID    int64  `json:"crawl_url_id"`
	Kind          string `json:"kind"`
	MIMEType      string `json:"mime_type"`
	Viewport      string `json:"viewport,omitempty"`
	EngineVersion string `json:"engine_version,omitempty"`
}

func (f *Frontier) RecordPageArtifact(ctx context.Context, record PageArtifactRecord) error {
	return f.writer.Submit(ctx, func(ctx context.Context, tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx, `INSERT INTO artifact(id,crawl_id,format,relative_path,checksum,size_bytes,created_at,expires_at) VALUES(?,?,?,?,?,?,?,?)`, record.ID, record.CrawlID, record.Format, record.RelativePath, record.Checksum, record.SizeBytes, record.CreatedAt, nullableString(record.ExpiresAt)); err != nil {
			return err
		}
		_, err := tx.ExecContext(ctx, `INSERT INTO page_artifact(artifact_id,crawl_url_id,kind,mime_type,viewport,engine_version) VALUES(?,?,?,?,?,?)`, record.ID, record.CrawlURLID, record.Kind, record.MIMEType, record.Viewport, record.EngineVersion)
		return err
	})
}

func (f *Frontier) PageArtifactBytes(ctx context.Context, crawlID contracts.ID) (int64, error) {
	var total int64
	err := f.db.QueryRowContext(ctx, `SELECT COALESCE(sum(a.size_bytes),0) FROM artifact a JOIN page_artifact pa ON pa.artifact_id=a.id WHERE a.crawl_id=?`, crawlID).Scan(&total)
	return total, err
}

func (f *Frontier) RecordArtifact(ctx context.Context, record ArtifactRecord) error {
	return f.writer.Submit(ctx, func(ctx context.Context, tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, `INSERT INTO artifact(id,crawl_id,format,relative_path,checksum,size_bytes,created_at,expires_at) VALUES (?,?,?,?,?,?,?,?)`, record.ID, record.CrawlID, record.Format, record.RelativePath, record.Checksum, record.SizeBytes, record.CreatedAt, nullableString(record.ExpiresAt))
		return err
	})
}

func (f *Frontier) GetArtifact(ctx context.Context, id contracts.ID) (ArtifactRecord, error) {
	var result ArtifactRecord
	var expiry sql.NullString
	err := f.db.QueryRowContext(ctx, `SELECT id,crawl_id,format,relative_path,checksum,size_bytes,created_at,expires_at FROM artifact WHERE id=?`, id).Scan(&result.ID, &result.CrawlID, &result.Format, &result.RelativePath, &result.Checksum, &result.SizeBytes, &result.CreatedAt, &expiry)
	if expiry.Valid {
		result.ExpiresAt = expiry.String
	}
	return result, err
}

func (f *Frontier) DeleteExpiredArtifacts(ctx context.Context, before time.Time) ([]string, error) {
	var paths []string
	rows, err := f.db.QueryContext(ctx, `SELECT relative_path FROM artifact WHERE expires_at IS NOT NULL AND expires_at<=?`, before.UTC().Format(time.RFC3339Nano))
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var value string
		if err := rows.Scan(&value); err != nil {
			rows.Close()
			return nil, err
		}
		paths = append(paths, value)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	err = f.writer.Submit(ctx, func(ctx context.Context, tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, `DELETE FROM artifact WHERE expires_at IS NOT NULL AND expires_at<=?`, before.UTC().Format(time.RFC3339Nano))
		return err
	})
	return paths, err
}

// ArtifactRelativePaths returns the filenames owned by the artifact database.
// Startup reconciliation uses this to remove only managed artifact_* files
// that no longer have a corresponding database record.
func (f *Frontier) ArtifactRelativePaths(ctx context.Context) (map[string]struct{}, error) {
	rows, err := f.db.QueryContext(ctx, `SELECT relative_path FROM artifact`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	paths := make(map[string]struct{})
	for rows.Next() {
		var path string
		if err := rows.Scan(&path); err != nil {
			return nil, err
		}
		if path == "" {
			return nil, errors.New("artifact has an empty managed path")
		}
		paths[path] = struct{}{}
	}
	return paths, rows.Err()
}

func nullableString(value string) any {
	if value == "" {
		return nil
	}
	return value
}
