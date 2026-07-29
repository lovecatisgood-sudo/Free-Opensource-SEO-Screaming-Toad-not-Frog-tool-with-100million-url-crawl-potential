package database

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/seo-auditor/seo-auditor/internal/contracts"
	"github.com/seo-auditor/seo-auditor/internal/fetchpolicy"
)

func testFrontier(t *testing.T) (*Frontier, contracts.ID, contracts.ID) {
	t.Helper()
	db, err := Open(context.Background(), filepath.Join(t.TempDir(), "auditor.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	frontier := NewFrontier(db, 32)
	t.Cleanup(frontier.Close)
	projectID := contracts.ID("project_test")
	crawlID := contracts.ID("crawl_test")
	if err := frontier.CreateProject(context.Background(), projectID, "Test"); err != nil {
		t.Fatal(err)
	}
	seed, _ := fetchpolicy.NormalizeURL("https://example.com/")
	configuration := contracts.CrawlConfiguration{SeedURL: seed.RequestKey, AllowedHosts: []string{seed.URL.Hostname()}, UserAgent: "test", RenderingMode: "raw", Limits: contracts.DefaultCrawlLimits()}
	if err := frontier.CreateCrawl(context.Background(), crawlID, projectID, "", seed, configuration); err != nil {
		t.Fatal(err)
	}
	return frontier, projectID, crawlID
}

func TestFrontierDeduplicatesAndEnforcesLimit(t *testing.T) {
	t.Parallel()
	frontier, projectID, crawlID := testFrontier(t)
	ctx := context.Background()
	first, _ := fetchpolicy.NormalizeURL("https://example.com/a")
	inserted, err := frontier.Enqueue(ctx, Discovery{CrawlID: crawlID, ProjectID: projectID, URL: first, DiscoveryKind: "seed", MaximumURLs: 1})
	if err != nil || !inserted {
		t.Fatalf("first enqueue = %v, %v", inserted, err)
	}
	inserted, err = frontier.Enqueue(ctx, Discovery{CrawlID: crawlID, ProjectID: projectID, URL: first, DiscoveryKind: "link", MaximumURLs: 1})
	if err != nil || inserted {
		t.Fatalf("duplicate enqueue = %v, %v", inserted, err)
	}
	second, _ := fetchpolicy.NormalizeURL("https://example.com/b")
	_, err = frontier.Enqueue(ctx, Discovery{CrawlID: crawlID, ProjectID: projectID, URL: second, DiscoveryKind: "link", MaximumURLs: 1})
	if !errors.Is(err, ErrURLLimitReached) {
		t.Fatalf("limit error = %v", err)
	}
}

func TestFrontierLeaseRecoversAndCompletes(t *testing.T) {
	t.Parallel()
	frontier, projectID, crawlID := testFrontier(t)
	ctx := context.Background()
	target, _ := fetchpolicy.NormalizeURL("https://example.com/")
	_, err := frontier.Enqueue(ctx, Discovery{CrawlID: crawlID, ProjectID: projectID, URL: target, DiscoveryKind: "seed", MaximumURLs: 10})
	if err != nil {
		t.Fatal(err)
	}
	leases, err := frontier.Lease(ctx, crawlID, "worker-one", 1, time.Nanosecond)
	if err != nil || len(leases) != 1 {
		t.Fatalf("lease = %+v, %v", leases, err)
	}
	time.Sleep(time.Millisecond)
	recovered, err := frontier.Lease(ctx, crawlID, "worker-two", 1, time.Minute)
	if err != nil || len(recovered) != 1 || recovered[0].Attempt != 2 {
		t.Fatalf("recovered = %+v, %v", recovered, err)
	}
	now := time.Now().UTC()
	if err := frontier.CompleteFetch(ctx, crawlID, FetchCompletion{Lease: recovered[0], StatusCode: 200, ContentType: "text/html", StartedAt: now, FinishedAt: now}); err != nil {
		t.Fatal(err)
	}
	progress, err := frontier.Progress(ctx, crawlID)
	if err != nil {
		t.Fatal(err)
	}
	if progress.Discovered != 1 || progress.Fetched != 1 || progress.Queued != 0 {
		t.Fatalf("unexpected progress: %+v", progress)
	}
}

func TestFrontierControlTransitions(t *testing.T) {
	t.Parallel()
	frontier, _, crawlID := testFrontier(t)
	ctx := context.Background()
	if err := frontier.SetStatus(ctx, crawlID, []contracts.CrawlStatus{contracts.CrawlPending}, contracts.CrawlRunning, ""); err != nil {
		t.Fatal(err)
	}
	if err := frontier.RequestPause(ctx, crawlID); err != nil {
		t.Fatal(err)
	}
	if err := frontier.SetStatus(ctx, crawlID, []contracts.CrawlStatus{contracts.CrawlPausing}, contracts.CrawlPaused, ""); err != nil {
		t.Fatal(err)
	}
	if err := frontier.RequestCancel(ctx, crawlID); err != nil {
		t.Fatal(err)
	}
	if err := frontier.SetStatus(ctx, crawlID, []contracts.CrawlStatus{contracts.CrawlCancelling}, contracts.CrawlCancelled, "user_cancelled"); err != nil {
		t.Fatal(err)
	}
	progress, err := frontier.Progress(ctx, crawlID)
	if err != nil || progress.Status != contracts.CrawlCancelled {
		t.Fatalf("progress = %+v, err = %v", progress, err)
	}
	events, err := frontier.ListEvents(ctx, crawlID, contracts.PageRequest{Limit: 20})
	if err != nil || len(events.Items) != 6 {
		t.Fatalf("events=%+v err=%v", events, err)
	}
	if events.Items[0].Event != "created" || events.Items[len(events.Items)-1].Event != "status_changed" {
		t.Fatalf("unexpected event timeline: %+v", events.Items)
	}
}

func TestInterruptedCrawlRecoversPausedWithPersistedConfiguration(t *testing.T) {
	t.Parallel()
	frontier, projectID, crawlID := testFrontier(t)
	ctx := context.Background()
	target, _ := fetchpolicy.NormalizeURL("https://example.com/recover")
	if _, err := frontier.Enqueue(ctx, Discovery{CrawlID: crawlID, ProjectID: projectID, URL: target, DiscoveryKind: "link", MaximumURLs: 10}); err != nil {
		t.Fatal(err)
	}
	if err := frontier.SetStatus(ctx, crawlID, []contracts.CrawlStatus{contracts.CrawlPending}, contracts.CrawlRunning, ""); err != nil {
		t.Fatal(err)
	}
	if leases, err := frontier.Lease(ctx, crawlID, "dead-worker", 1, time.Hour); err != nil || len(leases) != 1 {
		t.Fatalf("lease=%+v err=%v", leases, err)
	}
	if err := frontier.RecoverInterruptedCrawls(ctx); err != nil {
		t.Fatal(err)
	}
	progress, err := frontier.Progress(ctx, crawlID)
	if err != nil || progress.Status != contracts.CrawlPaused || progress.Queued != 1 || progress.TerminalReason != "recovered_after_restart" {
		t.Fatalf("progress=%+v err=%v", progress, err)
	}
	stored, err := frontier.LoadCrawl(ctx, crawlID)
	if err != nil {
		t.Fatal(err)
	}
	if stored.Configuration.SeedURL != "https://example.com/" || stored.Configuration.Limits.MaximumURLs != contracts.DefaultCrawlLimits().MaximumURLs {
		t.Fatalf("configuration=%+v", stored.Configuration)
	}
}

func TestTerminalCrawlTrashAndRestore(t *testing.T) {
	t.Parallel()
	frontier, _, crawlID := testFrontier(t)
	ctx := context.Background()
	if err := frontier.SetStatus(ctx, crawlID, []contracts.CrawlStatus{contracts.CrawlPending}, contracts.CrawlCompleted, ""); err != nil {
		t.Fatal(err)
	}
	if err := frontier.TrashCrawl(ctx, crawlID); err != nil {
		t.Fatal(err)
	}
	var deleted any
	if err := frontier.db.QueryRowContext(ctx, `SELECT deleted_at FROM crawl WHERE id=?`, crawlID).Scan(&deleted); err != nil {
		t.Fatal(err)
	}
	if deleted == nil {
		t.Fatal("crawl was not trashed")
	}
	if err := frontier.RestoreCrawl(ctx, crawlID); err != nil {
		t.Fatal(err)
	}
	var active int
	if err := frontier.db.QueryRowContext(ctx, `SELECT count(*) FROM crawl WHERE id=? AND deleted_at IS NULL`, crawlID).Scan(&active); err != nil {
		t.Fatal(err)
	}
	if active != 1 {
		t.Fatal("crawl was not restored")
	}
}
