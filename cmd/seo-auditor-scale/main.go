// seo-auditor-scale runs repeatable synthetic campaign benchmarks through the
// production frontier, extraction, rules, graph, segment, and verification
// paths. It does not perform network fetches and is never live-fetch evidence.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/seo-auditor/seo-auditor/internal/contracts"
	"github.com/seo-auditor/seo-auditor/internal/crawler"
	"github.com/seo-auditor/seo-auditor/internal/database"
	"github.com/seo-auditor/seo-auditor/internal/fetchpolicy"
)

type generator struct {
	total  int64
	fanout int64
	calls  atomic.Int64
}

func (g *generator) Fetch(_ context.Context, raw string) (fetchpolicy.FetchResult, error) {
	index, err := strconv.ParseInt(strings.TrimPrefix(raw, "https://synthetic.invalid/"), 10, 64)
	if err != nil {
		return fetchpolicy.FetchResult{}, err
	}
	g.calls.Add(1)
	var body strings.Builder
	fmt.Fprintf(&body, `<!doctype html><html lang="en"><head><title>Synthetic page %d technical audit fixture</title><meta name="description" content="Deterministic campaign benchmark page %d with retained SEO evidence."><link rel="canonical" href="%s"></head><body><h1>Synthetic page %d</h1><p>Deterministic unique content token %d for campaign verification and similarity evidence.</p>`, index, index, raw, index, index)
	for child := index*g.fanout + 1; child <= index*g.fanout+g.fanout && child < g.total; child++ {
		fmt.Fprintf(&body, `<a href="/%d">Child %d</a>`, child, child)
	}
	body.WriteString(`</body></html>`)
	now := time.Now().UTC()
	return fetchpolicy.FetchResult{FinalURL: raw, StatusCode: http.StatusOK, Header: http.Header{"Content-Type": {"text/html; charset=utf-8"}}, ContentType: "text/html", Body: []byte(body.String()), CompressedBytes: int64(body.Len()), DecodedBytes: int64(body.Len()), StartedAt: now, FinishedAt: now}, nil
}

func main() {
	if len(os.Args) < 2 {
		fail(errors.New("usage: seo-auditor-scale <run|resume|status|verify> --database PATH"))
	}
	if os.Args[1] == "status" {
		status(os.Args[2:])
		return
	}
	if os.Args[1] == "verify" {
		verify(os.Args[2:])
		return
	}
	if os.Args[1] == "resume" {
		resume(os.Args[2:])
		return
	}
	if os.Args[1] != "run" {
		fail(errors.New("usage: seo-auditor-scale <run|resume|status|verify> --database PATH"))
	}
	flags := flag.NewFlagSet("run", flag.ExitOnError)
	urls := flags.Int64("urls", 1_000_000, "unique synthetic URLs")
	databasePath := flags.String("database", ".data/scale.sqlite3", "benchmark database path")
	segmentSize := flags.Int64("segment-size", 100_000, "durable segment size")
	workers := flags.Int("workers", 64, "worker concurrency")
	diskBytes := flags.Int64("disk-bytes", 250<<30, "maximum database bytes")
	_ = flags.Parse(os.Args[2:])
	if *urls < 1 || *urls > 100_000_000 {
		fail(errors.New("--urls must be between 1 and 100000000"))
	}
	if err := os.MkdirAll(filepath.Dir(*databasePath), 0o700); err != nil {
		fail(err)
	}
	ctx := context.Background()
	db, err := database.Open(ctx, *databasePath)
	if err != nil {
		fail(err)
	}
	defer db.Close()
	frontier := database.NewFrontier(db, 4096)
	defer frontier.Close()
	projectID, crawlID := contracts.ID("project_scale"), contracts.ID("crawl_scale")
	if err := frontier.CreateProject(ctx, projectID, "Synthetic scale benchmark"); err != nil {
		fail(fmt.Errorf("database must be new or empty: %w", err))
	}
	seed, _ := fetchpolicy.NormalizeURL("https://synthetic.invalid/0")
	limits := contracts.DefaultCrawlLimits()
	limits.MaximumURLs, limits.MaximumDepth, limits.MaximumDuration, limits.MaximumDiskBytes = *urls, 1000, 30*24*time.Hour, *diskBytes
	limits.GlobalConcurrency, limits.PerHostConcurrency, limits.MinimumHostDelay = *workers, *workers, 0
	configuration := contracts.CrawlConfiguration{SeedURL: seed.RequestKey, AllowedHosts: []string{seed.URL.Hostname()}, UserAgent: "SEOAuditor synthetic benchmark", RenderingMode: "raw", SegmentSize: *segmentSize, Limits: limits}
	if err := frontier.CreateCrawl(ctx, crawlID, projectID, "", seed, configuration); err != nil {
		fail(err)
	}
	if _, err := frontier.Enqueue(ctx, database.Discovery{CrawlID: crawlID, ProjectID: projectID, URL: seed, DiscoveryKind: "seed", MaximumURLs: *urls}); err != nil {
		fail(err)
	}
	scope, _ := fetchpolicy.CompileScope(fetchpolicy.ScopeConfig{AllowedHosts: []string{seed.URL.Hostname()}})
	fixture := &generator{total: *urls, fanout: 10}
	started := time.Now()
	engine := crawler.Engine{Frontier: frontier, Fetcher: fixture, Scope: scope, LeaseTime: 2 * time.Minute, MaxLinksPerPage: 20}
	if err := engine.Run(ctx, crawler.RunRequest{CrawlID: crawlID, ProjectID: projectID, Limits: limits, WorkerID: "synthetic-scale", SegmentSize: *segmentSize, NearDuplicateDistance: 3}); err != nil {
		fail(err)
	}
	verification, err := frontier.VerifyCampaign(ctx, crawlID)
	if err != nil {
		fail(err)
	}
	result := map[string]any{"transport": "synthetic_in_process_not_live_guarded_fetch", "requested_urls": *urls, "fetch_calls": fixture.calls.Load(), "elapsed_seconds": time.Since(started).Seconds(), "verification": verification}
	_ = json.NewEncoder(os.Stdout).Encode(result)
	if err := verification.Error(); err != nil {
		os.Exit(1)
	}
}

func verify(arguments []string) {
	flags := flag.NewFlagSet("verify", flag.ExitOnError)
	databasePath := flags.String("database", ".data/scale.sqlite3", "benchmark database path")
	_ = flags.Parse(arguments)
	ctx := context.Background()
	db, err := database.OpenReadOnly(ctx, *databasePath)
	if err != nil {
		fail(err)
	}
	defer db.Close()
	frontier := database.NewFrontier(db, 16)
	defer frontier.Close()
	result, err := frontier.VerifyCampaign(ctx, contracts.ID("crawl_scale"))
	if err != nil {
		fail(err)
	}
	_ = json.NewEncoder(os.Stdout).Encode(result)
	if err := result.Error(); err != nil {
		os.Exit(1)
	}
}

func status(arguments []string) {
	flags := flag.NewFlagSet("status", flag.ExitOnError)
	databasePath := flags.String("database", ".data/scale.sqlite3", "benchmark database path")
	brief := flags.Bool("brief", false, "print progress and aggregate segment status")
	_ = flags.Parse(arguments)
	ctx := context.Background()
	db, err := database.OpenReadOnly(ctx, *databasePath)
	if err != nil {
		fail(err)
	}
	defer db.Close()
	frontier := database.NewFrontier(db, 16)
	defer frontier.Close()
	progress, err := frontier.Progress(ctx, contracts.ID("crawl_scale"))
	if err != nil {
		fail(err)
	}
	segments, err := frontier.ListSegments(ctx, contracts.ID("crawl_scale"))
	if err != nil {
		fail(err)
	}
	storage, err := frontier.StorageBytes()
	if err != nil {
		fail(err)
	}
	if *brief {
		completed := 0
		for _, segment := range segments {
			if segment.Status == "completed" {
				completed++
			}
		}
		_ = json.NewEncoder(os.Stdout).Encode(map[string]any{"progress": progress, "completed_segments": completed, "storage_bytes": storage})
		return
	}
	_ = json.NewEncoder(os.Stdout).Encode(map[string]any{"progress": progress, "segments": segments, "storage_bytes": storage})
}

func resume(arguments []string) {
	flags := flag.NewFlagSet("resume", flag.ExitOnError)
	databasePath := flags.String("database", ".data/scale.sqlite3", "benchmark database path")
	workers := flags.Int("workers", 64, "worker concurrency")
	_ = flags.Parse(arguments)
	ctx := context.Background()
	db, err := database.Open(ctx, *databasePath)
	if err != nil {
		fail(err)
	}
	defer db.Close()
	frontier := database.NewFrontier(db, 4096)
	defer frontier.Close()
	if err := frontier.RecoverInterruptedCrawls(ctx); err != nil {
		fail(err)
	}
	stored, err := frontier.LoadCrawl(ctx, contracts.ID("crawl_scale"))
	if err != nil {
		fail(err)
	}
	stored.Configuration.Limits.GlobalConcurrency = *workers
	stored.Configuration.Limits.PerHostConcurrency = *workers
	scope, err := fetchpolicy.CompileScope(fetchpolicy.ScopeConfig{AllowedHosts: stored.Configuration.AllowedHosts})
	if err != nil {
		fail(err)
	}
	fixture := &generator{total: stored.Configuration.Limits.MaximumURLs, fanout: 10}
	started := time.Now()
	engine := crawler.Engine{Frontier: frontier, Fetcher: fixture, Scope: scope, LeaseTime: 2 * time.Minute, MaxLinksPerPage: 20}
	request := crawler.RunRequest{CrawlID: stored.CrawlID, ProjectID: stored.ProjectID, Limits: stored.Configuration.Limits, WorkerID: "synthetic-scale-resume", SegmentSize: stored.Configuration.EffectiveSegmentSize(), NearDuplicateDistance: stored.Configuration.EffectiveNearDuplicateDistance()}
	if err := engine.Run(ctx, request); err != nil {
		fail(err)
	}
	verification, err := frontier.VerifyCampaign(ctx, stored.CrawlID)
	if err != nil {
		fail(err)
	}
	result := map[string]any{"transport": "synthetic_in_process_not_live_guarded_fetch", "resumed_after_interruption": true, "fetch_calls_after_resume": fixture.calls.Load(), "elapsed_seconds_after_resume": time.Since(started).Seconds(), "verification": verification}
	_ = json.NewEncoder(os.Stdout).Encode(result)
	if err := verification.Error(); err != nil {
		os.Exit(1)
	}
}

func fail(err error) {
	_ = json.NewEncoder(os.Stderr).Encode(map[string]string{"error": err.Error()})
	os.Exit(1)
}
