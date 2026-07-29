package database

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestFencedCoordinatorAndHostOwnership(t *testing.T) {
	t.Parallel()
	frontier, _, crawlID := testFrontier(t)
	ctx := context.Background()
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	first, err := frontier.AcquireCoordinator(ctx, crawlID, "coordinator-a", now, time.Minute)
	if err != nil || first.Epoch != 1 {
		t.Fatalf("first=%+v err=%v", first, err)
	}
	if _, err := frontier.AcquireCoordinator(ctx, crawlID, "coordinator-b", now.Add(30*time.Second), time.Minute); !errors.Is(err, errLeaseHeld) {
		t.Fatalf("competing coordinator err=%v", err)
	}
	second, err := frontier.AcquireCoordinator(ctx, crawlID, "coordinator-b", now.Add(2*time.Minute), time.Minute)
	if err != nil || second.Epoch != 2 {
		t.Fatalf("failover=%+v err=%v", second, err)
	}
	if _, err := frontier.AcquireHostOwner(ctx, crawlID, "example.com", "worker-a", now, time.Minute); err != nil {
		t.Fatal(err)
	}
	if _, err := frontier.AcquireHostOwner(ctx, crawlID, "example.com", "worker-b", now, time.Minute); !errors.Is(err, errLeaseHeld) {
		t.Fatalf("competing host owner err=%v", err)
	}
}

func TestImmutableResultSegmentCommitIsIdempotent(t *testing.T) {
	t.Parallel()
	frontier, _, crawlID := testFrontier(t)
	segment := ResultSegment{CrawlID: crawlID, Sequence: 0, ObjectKey: "segments/000000.parquet", Checksum: strings.Repeat("a", 64), RowCount: 100_000, SchemaVersion: 1, RuleVersion: 1}
	inserted, err := frontier.CommitResultSegment(context.Background(), segment)
	if err != nil || !inserted {
		t.Fatalf("first inserted=%v err=%v", inserted, err)
	}
	inserted, err = frontier.CommitResultSegment(context.Background(), segment)
	if err != nil || inserted {
		t.Fatalf("replay inserted=%v err=%v", inserted, err)
	}
	segment.Checksum = strings.Repeat("b", 64)
	if _, err := frontier.CommitResultSegment(context.Background(), segment); !errors.Is(err, errSegmentClash) {
		t.Fatalf("conflict err=%v", err)
	}
}
