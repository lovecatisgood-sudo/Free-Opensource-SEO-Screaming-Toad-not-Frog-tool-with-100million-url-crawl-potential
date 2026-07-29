package crawler

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/seo-auditor/seo-auditor/internal/contracts"
	"github.com/seo-auditor/seo-auditor/internal/database"
	"github.com/seo-auditor/seo-auditor/internal/fetchpolicy"
)

type syntheticTreeFetcher struct{ calls atomic.Int64 }

func (f *syntheticTreeFetcher) Fetch(_ context.Context, raw string) (fetchpolicy.FetchResult, error) {
	f.calls.Add(1)
	path := strings.TrimPrefix(raw, "https://example.com/")
	index, _ := strconv.Atoi(path)
	var body strings.Builder
	for child := index*10 + 1; child <= index*10+10 && child < 10_000; child++ {
		fmt.Fprintf(&body, `<a href="/%d">%d</a>`, child, child)
	}
	now := time.Now().UTC()
	return fetchpolicy.FetchResult{FinalURL: raw, StatusCode: http.StatusOK,
		Header: http.Header{"Content-Type": {"text/html"}}, ContentType: "text/html", Body: []byte(body.String()),
		CompressedBytes: int64(body.Len()), DecodedBytes: int64(body.Len()), StartedAt: now, FinishedAt: now}, nil
}

func TestEngineSynthetic10K(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping 10k durable crawl in short mode")
	}
	ctx := context.Background()
	db, err := database.Open(ctx, filepath.Join(t.TempDir(), "auditor.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	frontier := database.NewFrontier(db, 1024)
	t.Cleanup(frontier.Close)
	projectID, crawlID := contracts.ID("project_10k"), contracts.ID("crawl_10k")
	if err := frontier.CreateProject(ctx, projectID, "10k"); err != nil {
		t.Fatal(err)
	}
	seed, _ := fetchpolicy.NormalizeURL("https://example.com/0")
	limits := contracts.DefaultCrawlLimits()
	limits.MaximumURLs, limits.MaximumDepth = 10_000, 10
	limits.GlobalConcurrency, limits.PerHostConcurrency, limits.MinimumHostDelay = 32, 8, 0
	configuration := contracts.CrawlConfiguration{SeedURL: seed.RequestKey, AllowedHosts: []string{seed.URL.Hostname()}, UserAgent: "test", RenderingMode: "raw", Limits: limits}
	if err := frontier.CreateCrawl(ctx, crawlID, projectID, "", seed, configuration); err != nil {
		t.Fatal(err)
	}
	if _, err := frontier.Enqueue(ctx, database.Discovery{CrawlID: crawlID, ProjectID: projectID, URL: seed, DiscoveryKind: "seed", MaximumURLs: limits.MaximumURLs}); err != nil {
		t.Fatal(err)
	}
	scope, _ := fetchpolicy.CompileScope(fetchpolicy.ScopeConfig{AllowedHosts: []string{"example.com"}})
	fetcher := &syntheticTreeFetcher{}
	started := time.Now()
	engine := &Engine{Frontier: frontier, Fetcher: fetcher, Scope: scope, LeaseTime: time.Minute, MaxLinksPerPage: 20}
	if err := engine.Run(ctx, RunRequest{CrawlID: crawlID, ProjectID: projectID, Limits: limits, WorkerID: "scale"}); err != nil {
		t.Fatal(err)
	}
	progress, err := frontier.Progress(ctx, crawlID)
	if err != nil {
		t.Fatal(err)
	}
	if progress.Status != contracts.CrawlCompleted || progress.Discovered != 10_000 || progress.Fetched != 10_000 || fetcher.calls.Load() != 10_000 {
		t.Fatalf("unexpected 10k result: progress=%+v calls=%d", progress, fetcher.calls.Load())
	}
	t.Logf("durable_urls=%d elapsed=%s", progress.Fetched, time.Since(started).Round(time.Millisecond))
}
