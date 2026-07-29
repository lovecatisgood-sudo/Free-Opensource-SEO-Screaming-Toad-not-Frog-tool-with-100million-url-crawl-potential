package campaign

import (
	"context"
	"errors"
	"io"
	"time"

	"github.com/seo-auditor/seo-auditor/internal/contracts"
	"github.com/seo-auditor/seo-auditor/internal/database"
	"github.com/seo-auditor/seo-auditor/internal/segmentstore"
)

type Coordinator struct {
	Frontier *database.Frontier
	Segments *segmentstore.Store
}

type SegmentRequest struct {
	CrawlID       contracts.ID
	Sequence      int64
	MaximumBytes  int64
	RowCount      int64
	SchemaVersion int
	RuleVersion   int
	Body          io.Reader
}

type SegmentCommit struct {
	ObjectKey string `json:"object_key"`
	Checksum  string `json:"checksum"`
	SizeBytes int64  `json:"size_bytes"`
	Inserted  bool   `json:"inserted"`
}

func (c *Coordinator) CommitSegment(ctx context.Context, request SegmentRequest) (SegmentCommit, error) {
	if c == nil || c.Frontier == nil || c.Segments == nil {
		return SegmentCommit{}, errors.New("campaign coordinator is incomplete")
	}
	stored, err := c.Segments.Commit(ctx, string(request.CrawlID), request.Sequence, request.MaximumBytes, request.Body)
	if err != nil {
		return SegmentCommit{}, err
	}
	inserted, err := c.Frontier.CommitResultSegment(ctx, database.ResultSegment{CrawlID: request.CrawlID, Sequence: request.Sequence, ObjectKey: stored.ObjectKey, Checksum: stored.Checksum, RowCount: request.RowCount, SchemaVersion: request.SchemaVersion, RuleVersion: request.RuleVersion, CreatedAt: time.Now().UTC().Format(time.RFC3339Nano)})
	if err != nil {
		return SegmentCommit{}, err
	}
	return SegmentCommit{ObjectKey: stored.ObjectKey, Checksum: stored.Checksum, SizeBytes: stored.SizeBytes, Inserted: inserted}, nil
}

func (c *Coordinator) Acquire(ctx context.Context, crawlID contracts.ID, owner string, now time.Time, duration time.Duration) (database.FencedLease, error) {
	if c == nil || c.Frontier == nil {
		return database.FencedLease{}, errors.New("campaign coordinator is incomplete")
	}
	return c.Frontier.AcquireCoordinator(ctx, crawlID, owner, now, duration)
}

func (c *Coordinator) AcquireHost(ctx context.Context, crawlID contracts.ID, host, owner string, now time.Time, duration time.Duration) (database.FencedLease, error) {
	if c == nil || c.Frontier == nil {
		return database.FencedLease{}, errors.New("campaign coordinator is incomplete")
	}
	return c.Frontier.AcquireHostOwner(ctx, crawlID, host, owner, now, duration)
}
