package database

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/seo-auditor/seo-auditor/internal/contracts"
	"github.com/seo-auditor/seo-auditor/internal/extractor"
	"github.com/seo-auditor/seo-auditor/internal/fetchpolicy"
	"github.com/seo-auditor/seo-auditor/internal/rules"
)

func TestAnalysisAndDiscoveriesRollbackTogether(t *testing.T) {
	t.Parallel()
	frontier, projectID, crawlID := testFrontier(t)
	ctx := context.Background()
	seed, _ := fetchpolicy.NormalizeURL("https://example.com/")
	if _, err := frontier.Enqueue(ctx, Discovery{CrawlID: crawlID, ProjectID: projectID, URL: seed, DiscoveryKind: "seed", MaximumURLs: 100}); err != nil {
		t.Fatal(err)
	}
	leases, err := frontier.Lease(ctx, crawlID, "worker", 1, time.Minute)
	if err != nil || len(leases) != 1 {
		t.Fatalf("leases=%v err=%v", leases, err)
	}
	now := time.Now().UTC()
	page := extractor.Page{URL: seed.RequestKey, Title: "Atomic page", VisibleText: "atomic page content", ContentHash: "content", SimilarityHash: "0000000000000000", Social: map[string]string{}}
	child, _ := fetchpolicy.NormalizeURL("https://example.com/child")
	commits := []AnalysisCommit{{Fetch: FetchCompletion{Lease: leases[0], StatusCode: http.StatusOK, ContentType: "text/html", StartedAt: now, FinishedAt: now}, ProjectID: projectID, Page: page}}
	discoveries := []Discovery{
		{CrawlID: crawlID, ProjectID: projectID, URL: child, Depth: 1, DiscoveryKind: "link", MaximumURLs: 100},
		{CrawlID: crawlID, ProjectID: contracts.ID("different_project"), URL: child, Depth: 1, DiscoveryKind: "link", MaximumURLs: 100},
	}
	if _, err := frontier.CommitAnalysesAndDiscoveries(ctx, crawlID, commits, discoveries); err == nil {
		t.Fatal("inconsistent discovery batch succeeded")
	}
	var state string
	if err := frontier.db.QueryRowContext(ctx, `SELECT state FROM crawl_url WHERE id=?`, leases[0].CrawlURLID).Scan(&state); err != nil {
		t.Fatal(err)
	}
	if state != "leased" {
		t.Fatalf("state=%s; fetch/analysis was not rolled back", state)
	}
	progress, err := frontier.Progress(ctx, crawlID)
	if err != nil {
		t.Fatal(err)
	}
	if progress.Fetched != 0 || progress.Analysed != 0 || progress.Discovered != 1 {
		t.Fatalf("progress=%+v", progress)
	}
}

func TestAnalysisPersistsFindingClassificationAndEvidenceSource(t *testing.T) {
	t.Parallel()
	frontier, projectID, crawlID := testFrontier(t)
	ctx := context.Background()
	seed, _ := fetchpolicy.NormalizeURL("https://example.com/provenance")
	if _, err := frontier.Enqueue(ctx, Discovery{CrawlID: crawlID, ProjectID: projectID, URL: seed, DiscoveryKind: "seed", MaximumURLs: 10}); err != nil {
		t.Fatal(err)
	}
	leases, err := frontier.Lease(ctx, crawlID, "worker", 1, time.Minute)
	if err != nil || len(leases) != 1 {
		t.Fatalf("leases=%v err=%v", leases, err)
	}
	now := time.Now().UTC()
	issue := rules.Issue{RuleID: "AUD-02", RuleVersion: 1, Severity: rules.SeverityError, Message: "Title is missing or empty", Evidence: map[string]any{"field": "title", "length": 0}, Classification: rules.ClassificationDeterministic}
	page := extractor.Page{URL: seed.RequestKey, VisibleText: "provenance", ContentHash: "content", Social: map[string]string{}}
	commit := AnalysisCommit{Fetch: FetchCompletion{Lease: leases[0], StatusCode: http.StatusOK, ContentType: "text/html", StartedAt: now, FinishedAt: now}, ProjectID: projectID, Page: page, Issues: []rules.Issue{issue}}
	if _, err := frontier.CommitAnalysesAndDiscoveries(ctx, crawlID, []AnalysisCommit{commit}, nil); err != nil {
		t.Fatal(err)
	}
	items, err := frontier.ListIssues(ctx, crawlID, contracts.PageRequest{Limit: 10})
	if err != nil || len(items.Items) != 1 {
		t.Fatalf("issues=%+v err=%v", items, err)
	}
	if items.Items[0].Classification != "deterministic" || items.Items[0].EvidenceSource != "raw" {
		t.Fatalf("issue provenance=%+v", items.Items[0])
	}
}
