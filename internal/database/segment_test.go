package database

import (
	"context"
	"testing"
)

func TestCampaignSegmentsCheckpointAndFinalize(t *testing.T) {
	t.Parallel()
	frontier, _, crawlID := testFrontier(t)
	ctx := context.Background()
	if _, err := frontier.db.ExecContext(ctx, `UPDATE crawl SET analysed_count=25000 WHERE id=?`, crawlID); err != nil {
		t.Fatal(err)
	}
	if err := frontier.CheckpointSegments(ctx, crawlID, 10_000, 1_000_000); err != nil {
		t.Fatal(err)
	}
	segments, err := frontier.ListSegments(ctx, crawlID)
	if err != nil {
		t.Fatal(err)
	}
	if len(segments) != 3 || segments[0].Status != "completed" || segments[1].Status != "completed" || segments[2].Status != "active" || segments[0].Checksum == "" {
		t.Fatalf("segments=%+v", segments)
	}
	if err := frontier.FinalizeSegments(ctx, crawlID, 10_000, 1_100_000); err != nil {
		t.Fatal(err)
	}
	segments, err = frontier.ListSegments(ctx, crawlID)
	if err != nil {
		t.Fatal(err)
	}
	if segments[2].Status != "completed" || segments[2].EndAnalysed == nil || *segments[2].EndAnalysed != 25_000 {
		t.Fatalf("final segment=%+v", segments[2])
	}
	var events int
	if err := frontier.db.QueryRowContext(ctx, `SELECT count(*) FROM crawl_event WHERE crawl_id=? AND event_type='segment_completed'`, crawlID).Scan(&events); err != nil {
		t.Fatal(err)
	}
	if events != 2 {
		t.Fatalf("segment events=%d", events)
	}
}
