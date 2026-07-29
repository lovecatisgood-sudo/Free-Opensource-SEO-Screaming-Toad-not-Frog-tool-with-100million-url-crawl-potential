package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/seo-auditor/seo-auditor/internal/contracts"
	"github.com/seo-auditor/seo-auditor/internal/fetchpolicy"
)

var ErrURLLimitReached = errors.New("crawl URL limit reached")

type Frontier struct {
	db     *sql.DB
	owner  *DB
	writer *Writer
}

func NewFrontier(db *DB, queueSize int) *Frontier {
	return &Frontier{db: db.SQL(), owner: db, writer: NewWriter(db.SQL(), queueSize)}
}

func (f *Frontier) Close() { f.writer.Close() }

func (f *Frontier) StorageBytes() (int64, error) { return f.owner.StorageBytes() }

func (f *Frontier) CreateProject(ctx context.Context, id contracts.ID, name string) error {
	if id == "" || name == "" {
		return errors.New("project id and name are required")
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	return f.writer.Submit(ctx, func(ctx context.Context, tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, `INSERT INTO project(id, name, created_at, updated_at) VALUES (?, ?, ?, ?)`, id, name, now, now)
		return err
	})
}

func (f *Frontier) CreateCrawl(ctx context.Context, crawlID, projectID, profileID contracts.ID, seed fetchpolicy.NormalizedURL, configuration contracts.CrawlConfiguration) error {
	if err := configuration.Limits.Validate(); err != nil {
		return err
	}
	config, err := json.Marshal(configuration)
	if err != nil {
		return err
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	return f.writer.Submit(ctx, func(ctx context.Context, tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, `INSERT INTO crawl(id, project_id, profile_id, seed_url, config_json, status, created_at, updated_at)
            VALUES (?, ?, ?, ?, ?, 'pending', ?, ?)`, crawlID, projectID, nullableID(profileID), seed.RequestKey, string(config), now, now)
		return err
	})
}

func nullableID(value contracts.ID) any {
	if value == "" {
		return nil
	}
	return value
}

type StoredCrawl struct {
	CrawlID       contracts.ID
	ProjectID     contracts.ID
	Status        contracts.CrawlStatus
	Configuration contracts.CrawlConfiguration
}

func (f *Frontier) LoadCrawl(ctx context.Context, crawlID contracts.ID) (StoredCrawl, error) {
	var result StoredCrawl
	var raw, seed string
	err := f.db.QueryRowContext(ctx, `SELECT id,project_id,status,seed_url,config_json FROM crawl WHERE id=?`, crawlID).Scan(&result.CrawlID, &result.ProjectID, &result.Status, &seed, &raw)
	if err != nil {
		return result, err
	}
	if err := json.Unmarshal([]byte(raw), &result.Configuration); err != nil {
		return result, fmt.Errorf("decode crawl configuration: %w", err)
	}
	// Compatibility with databases created before complete configuration was stored.
	if result.Configuration.SeedURL == "" {
		var limits contracts.CrawlLimits
		if err := json.Unmarshal([]byte(raw), &limits); err != nil {
			return result, fmt.Errorf("decode legacy crawl limits: %w", err)
		}
		normalized, err := fetchpolicy.NormalizeURL(seed)
		if err != nil {
			return result, err
		}
		result.Configuration = contracts.CrawlConfiguration{SeedURL: normalized.RequestKey, AllowedHosts: []string{normalized.URL.Hostname()}, UserAgent: "SEOAuditor/0.1", RenderingMode: "raw", Limits: limits}
	}
	return result, nil
}

// RecoverInterruptedCrawls releases leases owned by a dead process and leaves
// work paused for an explicit operator resume.
func (f *Frontier) RecoverInterruptedCrawls(ctx context.Context) error {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	return f.writer.Submit(ctx, func(ctx context.Context, tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx, `UPDATE crawl_url SET state='queued',lease_owner=NULL,lease_expires_at=NULL,updated_at=? WHERE crawl_id IN (SELECT id FROM crawl WHERE status IN ('running','pausing')) AND state='leased'`, now); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `UPDATE crawl SET status='paused',terminal_reason='recovered_after_restart',updated_at=? WHERE status IN ('running','pausing')`, now); err != nil {
			return err
		}
		_, err := tx.ExecContext(ctx, `UPDATE crawl SET status='cancelled',terminal_reason='recovered_cancel',finished_at=?,updated_at=? WHERE status='cancelling'`, now, now)
		return err
	})
}

type Discovery struct {
	CrawlID        contracts.ID
	ProjectID      contracts.ID
	URL            fetchpolicy.NormalizedURL
	Depth          int
	DiscoveryKind  string
	DiscoveredFrom *int64
	MaximumURLs    int64
}

func (f *Frontier) Enqueue(ctx context.Context, discovery Discovery) (bool, error) {
	if discovery.Depth < 0 || discovery.DiscoveryKind == "" || discovery.MaximumURLs < 1 {
		return false, errors.New("invalid discovery")
	}
	inserted := false
	err := f.writer.Submit(ctx, func(ctx context.Context, tx *sql.Tx) error {
		var current int64
		if err := tx.QueryRowContext(ctx, "SELECT discovered_count FROM crawl WHERE id = ?", discovery.CrawlID).Scan(&current); err != nil {
			return err
		}
		var urlID int64
		now := time.Now().UTC().Format(time.RFC3339Nano)
		_, err := tx.ExecContext(ctx, `INSERT INTO url(project_id, request_key, original_url, scheme, host, port, path, query, created_at)
            VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?) ON CONFLICT(project_id, request_key) DO NOTHING`,
			discovery.ProjectID, discovery.URL.RequestKey, discovery.URL.URL.String(), discovery.URL.URL.Scheme,
			discovery.URL.URL.Hostname(), discovery.URL.URL.Port(), discovery.URL.URL.EscapedPath(), discovery.URL.URL.RawQuery, now)
		if err != nil {
			return err
		}
		if err := tx.QueryRowContext(ctx, "SELECT id FROM url WHERE project_id = ? AND request_key = ?", discovery.ProjectID, discovery.URL.RequestKey).Scan(&urlID); err != nil {
			return err
		}
		var exists int
		if err := tx.QueryRowContext(ctx, "SELECT count(*) FROM crawl_url WHERE crawl_id = ? AND url_id = ?", discovery.CrawlID, urlID).Scan(&exists); err != nil {
			return err
		}
		if exists != 0 {
			return nil
		}
		if current >= discovery.MaximumURLs {
			return ErrURLLimitReached
		}
		result, err := tx.ExecContext(ctx, `INSERT INTO crawl_url(crawl_id, url_id, state, depth, discovered_from_id, discovery_kind, created_at, updated_at)
            VALUES (?, ?, 'queued', ?, ?, ?, ?, ?)`, discovery.CrawlID, urlID, discovery.Depth, discovery.DiscoveredFrom, discovery.DiscoveryKind, now, now)
		if err != nil {
			return err
		}
		rows, err := result.RowsAffected()
		if err != nil {
			return err
		}
		inserted = rows == 1
		if inserted {
			_, err = tx.ExecContext(ctx, "UPDATE crawl SET discovered_count = discovered_count + 1, updated_at = ? WHERE id = ?", now, discovery.CrawlID)
		}
		return err
	})
	return inserted, err
}

type Lease struct {
	CrawlURLID int64
	URLID      int64
	URL        string
	Depth      int
	Attempt    int
}

func (f *Frontier) Lease(ctx context.Context, crawlID contracts.ID, owner string, limit int, duration time.Duration) ([]Lease, error) {
	if owner == "" || limit < 1 || limit > 1000 || duration <= 0 {
		return nil, errors.New("invalid lease request")
	}
	leases := make([]Lease, 0, limit)
	err := f.writer.Submit(ctx, func(ctx context.Context, tx *sql.Tx) error {
		nowTime := time.Now().UTC()
		now := nowTime.Format(time.RFC3339Nano)
		expires := nowTime.Add(duration).Format(time.RFC3339Nano)
		if _, err := tx.ExecContext(ctx, `UPDATE crawl_url SET state = 'queued', lease_owner = NULL, lease_expires_at = NULL, updated_at = ?
            WHERE crawl_id = ? AND state = 'leased' AND lease_expires_at < ?`, now, crawlID, now); err != nil {
			return err
		}
		rows, err := tx.QueryContext(ctx, `SELECT cu.id, cu.url_id, u.request_key, cu.depth, cu.attempt_count + 1
            FROM crawl_url cu JOIN url u ON u.id = cu.url_id
            WHERE cu.crawl_id = ? AND (cu.state = 'queued' OR (cu.state = 'retry_wait' AND cu.next_attempt_at <= ?))
            ORDER BY cu.priority DESC, cu.depth, cu.id LIMIT ?`, crawlID, now, limit)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var lease Lease
			if err := rows.Scan(&lease.CrawlURLID, &lease.URLID, &lease.URL, &lease.Depth, &lease.Attempt); err != nil {
				return err
			}
			leases = append(leases, lease)
		}
		if err := rows.Err(); err != nil {
			return err
		}
		for _, lease := range leases {
			result, err := tx.ExecContext(ctx, `UPDATE crawl_url SET state = 'leased', lease_owner = ?, lease_expires_at = ?, attempt_count = ?, updated_at = ?
                WHERE id = ? AND state IN ('queued','retry_wait')`, owner, expires, lease.Attempt, now, lease.CrawlURLID)
			if err != nil {
				return err
			}
			if changed, _ := result.RowsAffected(); changed != 1 {
				return fmt.Errorf("lease state changed concurrently")
			}
		}
		return nil
	})
	return leases, err
}

type FetchCompletion struct {
	Lease           Lease
	StatusCode      int
	ContentType     string
	ResolvedIP      string
	CompressedBytes int64
	DecodedBytes    int64
	StartedAt       time.Time
	FinishedAt      time.Time
	Redirects       []fetchpolicy.RedirectEvidence
}

func (f *Frontier) CompleteFetch(ctx context.Context, crawlID contracts.ID, completion FetchCompletion) error {
	now := completion.FinishedAt.UTC().Format(time.RFC3339Nano)
	return f.writer.Submit(ctx, func(ctx context.Context, tx *sql.Tx) error {
		result, err := tx.ExecContext(ctx, `UPDATE crawl_url SET state = 'fetched', robots_decision = COALESCE(robots_decision, 'allowed'), lease_owner = NULL, lease_expires_at = NULL, updated_at = ?
            WHERE id = ? AND state = 'leased'`, now, completion.Lease.CrawlURLID)
		if err != nil {
			return err
		}
		if rows, _ := result.RowsAffected(); rows != 1 {
			return errors.New("crawl URL is not actively leased")
		}
		attemptResult, err := tx.ExecContext(ctx, `INSERT INTO fetch_attempt(crawl_url_id, attempt, started_at, finished_at, resolved_ip, status_code, content_type, compressed_bytes, decoded_bytes)
            VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`, completion.Lease.CrawlURLID, completion.Lease.Attempt,
			completion.StartedAt.UTC().Format(time.RFC3339Nano), now, completion.ResolvedIP, completion.StatusCode,
			completion.ContentType, completion.CompressedBytes, completion.DecodedBytes)
		if err != nil {
			return err
		}
		attemptID, err := attemptResult.LastInsertId()
		if err != nil {
			return err
		}
		for _, hop := range completion.Redirects {
			if _, err := tx.ExecContext(ctx, `INSERT INTO redirect_hop(fetch_attempt_id, hop, source_url, status_code, target_url, policy_decision) VALUES (?, ?, ?, ?, ?, 'allowed')`, attemptID, hop.Hop, hop.SourceURL, hop.StatusCode, hop.TargetURL); err != nil {
				return err
			}
		}
		_, err = tx.ExecContext(ctx, "UPDATE crawl SET fetched_count = fetched_count + 1, updated_at = ? WHERE id = ?", now, crawlID)
		return err
	})
}

func (f *Frontier) Fail(ctx context.Context, crawlID contracts.ID, lease Lease, code, detail string, retryAt *time.Time) error {
	nowTime := time.Now().UTC()
	now := nowTime.Format(time.RFC3339Nano)
	state := "failed"
	var next any
	if retryAt != nil {
		state = "retry_wait"
		next = retryAt.UTC().Format(time.RFC3339Nano)
	}
	return f.writer.Submit(ctx, func(ctx context.Context, tx *sql.Tx) error {
		result, err := tx.ExecContext(ctx, `UPDATE crawl_url SET state = ?, lease_owner = NULL, lease_expires_at = NULL, next_attempt_at = ?, updated_at = ?
            WHERE id = ? AND state = 'leased'`, state, next, now, lease.CrawlURLID)
		if err != nil {
			return err
		}
		if rows, _ := result.RowsAffected(); rows != 1 {
			return errors.New("crawl URL is not actively leased")
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO fetch_attempt(crawl_url_id, attempt, started_at, finished_at, error_code, error_detail)
            VALUES (?, ?, ?, ?, ?, ?)`, lease.CrawlURLID, lease.Attempt, now, now, code, detail)
		if err != nil {
			return err
		}
		if state == "failed" {
			_, err = tx.ExecContext(ctx, "UPDATE crawl SET failed_count = failed_count + 1, updated_at = ? WHERE id = ?", now, crawlID)
		}
		return err
	})
}

func (f *Frontier) Skip(ctx context.Context, crawlID contracts.ID, lease Lease, reason string) error {
	if reason == "" {
		return errors.New("skip reason is required")
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	return f.writer.Submit(ctx, func(ctx context.Context, tx *sql.Tx) error {
		robotsDecision := any(nil)
		if reason == "robots_disallowed" {
			robotsDecision = "disallowed"
		}
		result, err := tx.ExecContext(ctx, `UPDATE crawl_url SET state = 'skipped', skip_reason = ?, robots_decision = COALESCE(?, robots_decision), lease_owner = NULL, lease_expires_at = NULL, updated_at = ?
            WHERE id = ? AND state = 'leased'`, reason, robotsDecision, now, lease.CrawlURLID)
		if err != nil {
			return err
		}
		if rows, _ := result.RowsAffected(); rows != 1 {
			return errors.New("crawl URL is not actively leased")
		}
		_, err = tx.ExecContext(ctx, "UPDATE crawl SET updated_at = ? WHERE id = ?", now, crawlID)
		return err
	})
}

func (f *Frontier) SetStatus(ctx context.Context, crawlID contracts.ID, from []contracts.CrawlStatus, to contracts.CrawlStatus, reason string) error {
	if len(from) == 0 {
		return errors.New("source status is required")
	}
	return f.writer.Submit(ctx, func(ctx context.Context, tx *sql.Tx) error {
		var current contracts.CrawlStatus
		if err := tx.QueryRowContext(ctx, "SELECT status FROM crawl WHERE id = ?", crawlID).Scan(&current); err != nil {
			return err
		}
		allowed := false
		for _, status := range from {
			allowed = allowed || current == status
		}
		if !allowed {
			return fmt.Errorf("invalid crawl transition from %s to %s", current, to)
		}
		now := time.Now().UTC().Format(time.RFC3339Nano)
		var started, finished any
		if to == contracts.CrawlRunning && current == contracts.CrawlPending {
			started = now
		}
		if to.Terminal() {
			finished = now
		}
		_, err := tx.ExecContext(ctx, `UPDATE crawl SET status = ?, terminal_reason = ?, started_at = COALESCE(started_at, ?), finished_at = COALESCE(?, finished_at), updated_at = ? WHERE id = ?`,
			to, reason, started, finished, now, crawlID)
		return err
	})
}

func (f *Frontier) Progress(ctx context.Context, crawlID contracts.ID) (contracts.CrawlProgress, error) {
	var progress contracts.CrawlProgress
	var started sql.NullString
	var updated string
	err := f.db.QueryRowContext(ctx, `SELECT id, status, discovered_count,
        (SELECT count(*) FROM crawl_url WHERE crawl_id = crawl.id AND state IN ('queued','leased','retry_wait')),
        fetched_count, analysed_count, failed_count, started_at, updated_at, terminal_reason
        FROM crawl WHERE id = ?`, crawlID).Scan(&progress.CrawlID, &progress.Status, &progress.Discovered, &progress.Queued,
		&progress.Fetched, &progress.Analysed, &progress.Failed, &started, &updated, &progress.TerminalReason)
	if err != nil {
		return progress, err
	}
	progress.UpdatedAt, _ = time.Parse(time.RFC3339Nano, updated)
	if started.Valid {
		value, _ := time.Parse(time.RFC3339Nano, started.String)
		progress.StartedAt = &value
	}
	return progress, nil
}

func (f *Frontier) RequestPause(ctx context.Context, crawlID contracts.ID) error {
	return f.SetStatus(ctx, crawlID, []contracts.CrawlStatus{contracts.CrawlRunning}, contracts.CrawlPausing, "")
}

func (f *Frontier) Resume(ctx context.Context, crawlID contracts.ID) error {
	return f.SetStatus(ctx, crawlID, []contracts.CrawlStatus{contracts.CrawlPaused}, contracts.CrawlRunning, "")
}

func (f *Frontier) RequestCancel(ctx context.Context, crawlID contracts.ID) error {
	return f.SetStatus(ctx, crawlID, []contracts.CrawlStatus{contracts.CrawlPending, contracts.CrawlRunning, contracts.CrawlPaused, contracts.CrawlPausing}, contracts.CrawlCancelling, "")
}
