package crawler

import (
	"context"
	"math"
	"net/http"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/seo-auditor/seo-auditor/internal/contracts"
	"github.com/seo-auditor/seo-auditor/internal/database"
	"github.com/seo-auditor/seo-auditor/internal/fetchpolicy"
	"github.com/seo-auditor/seo-auditor/internal/renderer"
)

type mapFetcher struct {
	mu    sync.Mutex
	pages map[string]string
	calls map[string]int
}

type mapRenderer struct{}

func (mapRenderer) Render(_ context.Context, request renderer.Request) (renderer.Result, error) {
	html := `<html><head><title>Rendered child</title></head><body>child</body></html>`
	if request.URL == "https://example.com/" {
		html = `<html><head><title>Rendered root</title></head><body><a href="/client">client</a></body></html>`
	}
	return renderer.Result{Status: "completed", HTML: html, FinalURL: request.URL, RequestCount: 1, TransferredBytes: int64(len(html))}, nil
}

func (f *mapFetcher) Fetch(_ context.Context, raw string) (fetchpolicy.FetchResult, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls[raw]++
	body := f.pages[raw]
	now := time.Now().UTC()
	return fetchpolicy.FetchResult{
		RequestedURL: raw, FinalURL: raw, StatusCode: http.StatusOK,
		Header: http.Header{"Content-Type": []string{"text/html"}}, ContentType: "text/html",
		Body: []byte(body), CompressedBytes: int64(len(body)), DecodedBytes: int64(len(body)),
		StartedAt: now, FinishedAt: now,
	}, nil
}

func TestEngineRenderedModeDiscoversClientLinksAndPreservesRaw(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db, err := database.Open(ctx, filepath.Join(t.TempDir(), "auditor.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	frontier := database.NewFrontier(db, 32)
	t.Cleanup(frontier.Close)
	projectID := contracts.ID("project_rendered_engine")
	crawlID := contracts.ID("crawl_rendered_engine")
	if err := frontier.CreateProject(ctx, projectID, "Rendered Engine"); err != nil {
		t.Fatal(err)
	}
	seed, _ := fetchpolicy.NormalizeURL("https://example.com/")
	limits := contracts.DefaultCrawlLimits()
	limits.MaximumURLs = 10
	limits.GlobalConcurrency = 1
	limits.PerHostConcurrency = 1
	limits.MinimumHostDelay = 0
	configuration := contracts.CrawlConfiguration{SeedURL: seed.RequestKey, AllowedHosts: []string{seed.URL.Hostname()}, UserAgent: "test", RenderingMode: "rendered", Limits: limits}
	if err := frontier.CreateCrawl(ctx, crawlID, projectID, "", seed, configuration); err != nil {
		t.Fatal(err)
	}
	if _, err := frontier.Enqueue(ctx, database.Discovery{CrawlID: crawlID, ProjectID: projectID, URL: seed, DiscoveryKind: "seed", MaximumURLs: limits.MaximumURLs}); err != nil {
		t.Fatal(err)
	}
	scope, _ := fetchpolicy.CompileScope(fetchpolicy.ScopeConfig{AllowedHosts: []string{"example.com"}})
	fetcher := &mapFetcher{pages: map[string]string{
		"https://example.com/":       `<html><head><title>Raw root</title></head><body>raw</body></html>`,
		"https://example.com/client": `<html><head><title>Raw child</title></head><body>raw child</body></html>`,
	}, calls: make(map[string]int)}
	engine := &Engine{Frontier: frontier, Fetcher: fetcher, Scope: scope, Renderer: mapRenderer{}, LeaseTime: time.Minute}
	if err := engine.Run(ctx, RunRequest{CrawlID: crawlID, ProjectID: projectID, Limits: limits, WorkerID: "render-test", RenderingMode: "rendered"}); err != nil {
		t.Fatal(err)
	}
	progress, err := frontier.Progress(ctx, crawlID)
	if err != nil || progress.Discovered != 2 || progress.Analysed != 2 {
		t.Fatalf("progress=%+v err=%v", progress, err)
	}
	pages, err := frontier.ListPages(ctx, crawlID, contracts.PageRequest{Limit: 10})
	if err != nil || len(pages.Items) != 2 {
		t.Fatalf("pages=%+v err=%v", pages, err)
	}
	root := pages.Items[0]
	if root.Title != "Raw root" || root.RenderedTitle != "Rendered root" || root.RenderStatus != "completed" {
		t.Fatalf("root=%+v", root)
	}
}

func TestEngineCrawlsBoundedGraphOnce(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db, err := database.Open(ctx, filepath.Join(t.TempDir(), "auditor.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	frontier := database.NewFrontier(db, 32)
	t.Cleanup(frontier.Close)
	projectID := contracts.ID("project_engine")
	crawlID := contracts.ID("crawl_engine")
	if err := frontier.CreateProject(ctx, projectID, "Engine"); err != nil {
		t.Fatal(err)
	}
	seed, _ := fetchpolicy.NormalizeURL("https://example.com/")
	limits := contracts.DefaultCrawlLimits()
	limits.MaximumURLs = 10
	limits.GlobalConcurrency = 2
	limits.MinimumHostDelay = 0
	configuration := contracts.CrawlConfiguration{SeedURL: seed.RequestKey, AllowedHosts: []string{seed.URL.Hostname()}, UserAgent: "test", RenderingMode: "raw", Limits: limits}
	if err := frontier.CreateCrawl(ctx, crawlID, projectID, "", seed, configuration); err != nil {
		t.Fatal(err)
	}
	if _, err := frontier.Enqueue(ctx, database.Discovery{CrawlID: crawlID, ProjectID: projectID, URL: seed, DiscoveryKind: "seed", MaximumURLs: limits.MaximumURLs}); err != nil {
		t.Fatal(err)
	}
	scope, _ := fetchpolicy.CompileScope(fetchpolicy.ScopeConfig{AllowedHosts: []string{"example.com"}})
	fetcher := &mapFetcher{pages: map[string]string{
		"https://example.com/":  `<a href="/a">a</a><a href="/b">b</a><a href="/a#again">duplicate</a>`,
		"https://example.com/a": `<a href="/b">b</a>`,
		"https://example.com/b": `done`,
	}, calls: make(map[string]int)}
	engine := &Engine{Frontier: frontier, Fetcher: fetcher, Scope: scope, LeaseTime: time.Minute}
	if err := engine.Run(ctx, RunRequest{CrawlID: crawlID, ProjectID: projectID, Limits: limits, WorkerID: "test"}); err != nil {
		t.Fatalf("Run: %v", err)
	}
	progress, err := frontier.Progress(ctx, crawlID)
	if err != nil {
		t.Fatal(err)
	}
	if progress.Status != contracts.CrawlCompleted || progress.Discovered != 3 || progress.Fetched != 3 || progress.Analysed != 3 || progress.Queued != 0 {
		t.Fatalf("unexpected progress: %+v", progress)
	}
	for target, calls := range fetcher.calls {
		if calls != 1 {
			t.Errorf("%s fetched %d times", target, calls)
		}
	}
}

func TestProjectedStorageBytesRoundsUpAndSaturates(t *testing.T) {
	t.Parallel()
	if got := projectedStorageBytes(1_001, 1_000, 100_000); got != 100_100 {
		t.Fatalf("projection=%d", got)
	}
	if got := projectedStorageBytes(math.MaxInt64, 1, 100_000_000); got != math.MaxInt64 {
		t.Fatalf("saturated projection=%d", got)
	}
}

func TestStorageProjectionWaitsForCompleteSegment(t *testing.T) {
	t.Parallel()
	if shouldProjectStorage(1_135, 100_000) {
		t.Fatal("startup overhead sample must not trigger projection")
	}
	if !shouldProjectStorage(100_000, 100_000) {
		t.Fatal("complete segment must trigger projection")
	}
}
