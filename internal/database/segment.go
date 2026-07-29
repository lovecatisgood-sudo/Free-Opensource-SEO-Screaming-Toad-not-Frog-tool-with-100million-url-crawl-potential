package database

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/seo-auditor/seo-auditor/internal/contracts"
)

type CampaignSegment struct {
	ID            int64        `json:"id"`
	CrawlID       contracts.ID `json:"crawl_id"`
	Sequence      int64        `json:"sequence"`
	StartAnalysed int64        `json:"start_analysed"`
	EndAnalysed   *int64       `json:"end_analysed,omitempty"`
	Status        string       `json:"status"`
	StorageBytes  int64        `json:"storage_bytes"`
	Checksum      string       `json:"checksum,omitempty"`
	StartedAt     string       `json:"started_at"`
	CompletedAt   string       `json:"completed_at,omitempty"`
}

func (f *Frontier) CheckpointSegments(ctx context.Context, crawlID contracts.ID, segmentSize, storageBytes int64) error {
	if crawlID == "" || segmentSize < 1 || storageBytes < 0 {
		return errors.New("crawl ID, segment size and storage bytes are required")
	}
	return f.writer.Submit(ctx, func(ctx context.Context, tx *sql.Tx) error {
		var analysed int64
		if err := tx.QueryRowContext(ctx, "SELECT analysed_count FROM crawl WHERE id=?", crawlID).Scan(&analysed); err != nil {
			return err
		}
		now := time.Now().UTC().Format(time.RFC3339Nano)
		if _, err := tx.ExecContext(ctx, `INSERT INTO campaign_segment(crawl_id,sequence,start_analysed,status,started_at) VALUES (?,0,0,'active',?) ON CONFLICT(crawl_id,sequence) DO NOTHING`, crawlID, now); err != nil {
			return err
		}
		for {
			var sequence, start int64
			var status string
			if err := tx.QueryRowContext(ctx, `SELECT sequence,start_analysed,status FROM campaign_segment WHERE crawl_id=? ORDER BY sequence DESC LIMIT 1`, crawlID).Scan(&sequence, &start, &status); err != nil {
				return err
			}
			boundary := (sequence + 1) * segmentSize
			if status != "active" || analysed < boundary {
				return nil
			}
			checksum := segmentChecksum(crawlID, sequence, start, boundary, storageBytes)
			if _, err := tx.ExecContext(ctx, `UPDATE campaign_segment SET end_analysed=?,status='completed',storage_bytes=?,checksum=?,completed_at=? WHERE crawl_id=? AND sequence=? AND status='active'`, boundary, storageBytes, checksum, now, crawlID, sequence); err != nil {
				return err
			}
			payload := fmt.Sprintf(`{"sequence":%d,"start_analysed":%d,"end_analysed":%d,"storage_bytes":%d,"checksum":"%s"}`, sequence, start, boundary, storageBytes, checksum)
			if _, err := tx.ExecContext(ctx, `INSERT INTO crawl_event(crawl_id,event_type,payload_json,created_at) VALUES (?,'segment_completed',?,?)`, crawlID, payload, now); err != nil {
				return err
			}
			if _, err := tx.ExecContext(ctx, `INSERT INTO campaign_segment(crawl_id,sequence,start_analysed,status,started_at) VALUES (?,?,?,'active',?)`, crawlID, sequence+1, boundary, now); err != nil {
				return err
			}
		}
	})
}

func (f *Frontier) FinalizeSegments(ctx context.Context, crawlID contracts.ID, segmentSize, storageBytes int64) error {
	if err := f.CheckpointSegments(ctx, crawlID, segmentSize, storageBytes); err != nil {
		return err
	}
	return f.writer.Submit(ctx, func(ctx context.Context, tx *sql.Tx) error {
		var analysed, sequence, start int64
		if err := tx.QueryRowContext(ctx, `SELECT c.analysed_count,s.sequence,s.start_analysed FROM crawl c JOIN campaign_segment s ON s.crawl_id=c.id WHERE c.id=? AND s.status='active' ORDER BY s.sequence DESC LIMIT 1`, crawlID).Scan(&analysed, &sequence, &start); err != nil {
			return err
		}
		if analysed == start && sequence > 0 {
			_, err := tx.ExecContext(ctx, `DELETE FROM campaign_segment WHERE crawl_id=? AND sequence=? AND status='active'`, crawlID, sequence)
			return err
		}
		now := time.Now().UTC().Format(time.RFC3339Nano)
		checksum := segmentChecksum(crawlID, sequence, start, analysed, storageBytes)
		_, err := tx.ExecContext(ctx, `UPDATE campaign_segment SET end_analysed=?,status='completed',storage_bytes=?,checksum=?,completed_at=? WHERE crawl_id=? AND sequence=? AND status='active'`, analysed, storageBytes, checksum, now, crawlID, sequence)
		return err
	})
}

func (f *Frontier) ListSegments(ctx context.Context, crawlID contracts.ID) ([]CampaignSegment, error) {
	rows, err := f.db.QueryContext(ctx, `SELECT id,crawl_id,sequence,start_analysed,end_analysed,status,storage_bytes,COALESCE(checksum,''),started_at,COALESCE(completed_at,'') FROM campaign_segment WHERE crawl_id=? ORDER BY sequence`, crawlID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []CampaignSegment
	for rows.Next() {
		var item CampaignSegment
		var end sql.NullInt64
		if err := rows.Scan(&item.ID, &item.CrawlID, &item.Sequence, &item.StartAnalysed, &end, &item.Status, &item.StorageBytes, &item.Checksum, &item.StartedAt, &item.CompletedAt); err != nil {
			return nil, err
		}
		if end.Valid {
			item.EndAnalysed = &end.Int64
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func segmentChecksum(crawlID contracts.ID, sequence, start, end, storage int64) string {
	digest := sha256.Sum256([]byte(fmt.Sprintf("%s\x00%d\x00%d\x00%d\x00%d", crawlID, sequence, start, end, storage)))
	return hex.EncodeToString(digest[:])
}
