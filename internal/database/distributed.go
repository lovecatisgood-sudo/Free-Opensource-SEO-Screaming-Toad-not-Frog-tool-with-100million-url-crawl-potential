package database

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"time"

	"github.com/seo-auditor/seo-auditor/internal/contracts"
)

var (
	errLeaseHeld     = errors.New("lease is held by another owner")
	errSegmentClash  = errors.New("immutable segment conflicts with committed content")
	checksumPattern  = regexp.MustCompile(`^[0-9a-f]{64}$`)
	objectKeyPattern = regexp.MustCompile(`^segments/[A-Za-z0-9._/-]{1,500}$`)
)

type FencedLease struct {
	Owner     string `json:"owner"`
	Epoch     int64  `json:"epoch"`
	ExpiresAt string `json:"expires_at"`
}

type ResultSegment struct {
	CrawlID       contracts.ID `json:"crawl_id"`
	Sequence      int64        `json:"sequence"`
	ObjectKey     string       `json:"object_key"`
	Checksum      string       `json:"checksum"`
	RowCount      int64        `json:"row_count"`
	SchemaVersion int          `json:"schema_version"`
	RuleVersion   int          `json:"rule_version"`
	CreatedAt     string       `json:"created_at"`
}

func (f *Frontier) AcquireCoordinator(ctx context.Context, crawlID contracts.ID, owner string, now time.Time, duration time.Duration) (FencedLease, error) {
	return f.acquireFencedLease(ctx, "coordinator_lease", crawlID, "", owner, now, duration)
}

func (f *Frontier) AcquireHostOwner(ctx context.Context, crawlID contracts.ID, host, owner string, now time.Time, duration time.Duration) (FencedLease, error) {
	if host == "" {
		return FencedLease{}, errors.New("host is required")
	}
	return f.acquireFencedLease(ctx, "host_owner_lease", crawlID, host, owner, now, duration)
}

func (f *Frontier) acquireFencedLease(ctx context.Context, table string, crawlID contracts.ID, host, owner string, now time.Time, duration time.Duration) (FencedLease, error) {
	if crawlID == "" || owner == "" || duration <= 0 || duration > time.Hour {
		return FencedLease{}, errors.New("valid crawl, owner and bounded duration are required")
	}
	result := FencedLease{}
	err := f.writer.Submit(ctx, func(ctx context.Context, tx *sql.Tx) error {
		nowText := now.UTC().Format(time.RFC3339Nano)
		expires := now.UTC().Add(duration).Format(time.RFC3339Nano)
		var currentOwner, currentExpiry string
		var epoch int64
		var err error
		if table == "coordinator_lease" {
			err = tx.QueryRowContext(ctx, `SELECT owner,epoch,expires_at FROM coordinator_lease WHERE crawl_id=?`, crawlID).Scan(&currentOwner, &epoch, &currentExpiry)
		} else {
			err = tx.QueryRowContext(ctx, `SELECT owner,epoch,expires_at FROM host_owner_lease WHERE crawl_id=? AND host=?`, crawlID, host).Scan(&currentOwner, &epoch, &currentExpiry)
		}
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return err
		}
		if err == nil && currentOwner != owner && currentExpiry > nowText {
			return errLeaseHeld
		}
		if err == nil && currentOwner == owner && currentExpiry > nowText {
			result = FencedLease{Owner: owner, Epoch: epoch, ExpiresAt: expires}
			if table == "coordinator_lease" {
				_, err = tx.ExecContext(ctx, `UPDATE coordinator_lease SET expires_at=?,updated_at=? WHERE crawl_id=? AND owner=? AND epoch=?`, expires, nowText, crawlID, owner, epoch)
			} else {
				_, err = tx.ExecContext(ctx, `UPDATE host_owner_lease SET expires_at=?,updated_at=? WHERE crawl_id=? AND host=? AND owner=? AND epoch=?`, expires, nowText, crawlID, host, owner, epoch)
			}
			return err
		}
		epoch++
		result = FencedLease{Owner: owner, Epoch: epoch, ExpiresAt: expires}
		if table == "coordinator_lease" {
			_, err = tx.ExecContext(ctx, `INSERT INTO coordinator_lease(crawl_id,owner,epoch,expires_at,updated_at) VALUES (?,?,?,?,?) ON CONFLICT(crawl_id) DO UPDATE SET owner=excluded.owner,epoch=excluded.epoch,expires_at=excluded.expires_at,updated_at=excluded.updated_at`, crawlID, owner, epoch, expires, nowText)
		} else {
			_, err = tx.ExecContext(ctx, `INSERT INTO host_owner_lease(crawl_id,host,owner,epoch,expires_at,updated_at) VALUES (?,?,?,?,?,?) ON CONFLICT(crawl_id,host) DO UPDATE SET owner=excluded.owner,epoch=excluded.epoch,expires_at=excluded.expires_at,updated_at=excluded.updated_at`, crawlID, host, owner, epoch, expires, nowText)
		}
		return err
	})
	return result, err
}

func (f *Frontier) CommitResultSegment(ctx context.Context, segment ResultSegment) (bool, error) {
	if segment.CrawlID == "" || segment.Sequence < 0 || !objectKeyPattern.MatchString(segment.ObjectKey) || !checksumPattern.MatchString(segment.Checksum) || segment.RowCount < 1 || segment.SchemaVersion < 1 || segment.RuleVersion < 1 {
		return false, errors.New("result segment metadata is invalid")
	}
	inserted := false
	err := f.writer.Submit(ctx, func(ctx context.Context, tx *sql.Tx) error {
		var checksum, objectKey string
		err := tx.QueryRowContext(ctx, `SELECT checksum,object_key FROM immutable_result_segment WHERE crawl_id=? AND sequence=?`, segment.CrawlID, segment.Sequence).Scan(&checksum, &objectKey)
		if err == nil {
			if checksum != segment.Checksum || objectKey != segment.ObjectKey {
				return errSegmentClash
			}
			return nil
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return err
		}
		created := segment.CreatedAt
		if created == "" {
			created = time.Now().UTC().Format(time.RFC3339Nano)
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO immutable_result_segment(crawl_id,sequence,object_key,checksum,row_count,schema_version,rule_version,created_at) VALUES (?,?,?,?,?,?,?,?)`, segment.CrawlID, segment.Sequence, segment.ObjectKey, segment.Checksum, segment.RowCount, segment.SchemaVersion, segment.RuleVersion, created)
		inserted = err == nil
		return err
	})
	return inserted, err
}
