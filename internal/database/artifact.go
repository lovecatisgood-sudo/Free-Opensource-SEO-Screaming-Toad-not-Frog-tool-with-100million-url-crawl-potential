package database

import (
	"context"
	"database/sql"
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

func nullableString(value string) any {
	if value == "" {
		return nil
	}
	return value
}
