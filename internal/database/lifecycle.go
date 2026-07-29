package database

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/seo-auditor/seo-auditor/internal/contracts"
)

func (f *Frontier) TrashCrawl(ctx context.Context, crawlID contracts.ID) error {
	return f.writer.Submit(ctx, func(ctx context.Context, tx *sql.Tx) error {
		now := time.Now().UTC().Format(time.RFC3339Nano)
		result, err := tx.ExecContext(ctx, `UPDATE crawl SET deleted_at=?,updated_at=? WHERE id=? AND deleted_at IS NULL AND status IN ('cancelled','completed','failed','limit_reached')`, now, now, crawlID)
		if err != nil {
			return err
		}
		changed, _ := result.RowsAffected()
		if changed != 1 {
			return errors.New("only a terminal active crawl can be moved to trash")
		}
		return nil
	})
}
func (f *Frontier) RestoreCrawl(ctx context.Context, crawlID contracts.ID) error {
	return f.writer.Submit(ctx, func(ctx context.Context, tx *sql.Tx) error {
		now := time.Now().UTC().Format(time.RFC3339Nano)
		result, err := tx.ExecContext(ctx, `UPDATE crawl SET deleted_at=NULL,updated_at=? WHERE id=? AND deleted_at IS NOT NULL`, now, crawlID)
		if err != nil {
			return err
		}
		changed, _ := result.RowsAffected()
		if changed != 1 {
			return errors.New("crawl is not in trash")
		}
		return nil
	})
}
