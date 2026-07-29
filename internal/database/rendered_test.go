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

func TestRenderedAnalysisPreservesRawEvidenceAndDifferences(t *testing.T) {
	t.Parallel()
	frontier, projectID, crawlID := testFrontier(t)
	ctx := context.Background()
	target, _ := fetchpolicy.NormalizeURL("https://example.com/")
	if _, err := frontier.Enqueue(ctx, Discovery{CrawlID: crawlID, ProjectID: projectID, URL: target, DiscoveryKind: "seed", MaximumURLs: 10}); err != nil {
		t.Fatal(err)
	}
	leases, err := frontier.Lease(ctx, crawlID, "renderer-test", 1, time.Minute)
	if err != nil || len(leases) != 1 {
		t.Fatalf("leases=%+v err=%v", leases, err)
	}
	now := time.Now().UTC()
	if err := frontier.CompleteFetch(ctx, crawlID, FetchCompletion{Lease: leases[0], StatusCode: 200, ContentType: "text/html", StartedAt: now, FinishedAt: now}); err != nil {
		t.Fatal(err)
	}
	raw, err := extractor.Extract(target.RequestKey, http.Header{}, []byte(`<html><head><title>Raw</title></head><body>raw</body></html>`))
	if err != nil {
		t.Fatal(err)
	}
	if err := frontier.SaveAnalysis(ctx, crawlID, projectID, leases[0], raw, nil); err != nil {
		t.Fatal(err)
	}
	rendered, err := extractor.Extract(target.RequestKey, http.Header{}, []byte(`<html><head><title>Rendered</title></head><body><h1>Client</h1><a href="/client">client</a></body></html>`))
	if err != nil {
		t.Fatal(err)
	}
	issues := []rules.Issue{{RuleID: "AUD-02", RuleVersion: 1, Severity: rules.SeverityWarning, Evidence: map[string]any{"field": "title"}}}
	if err := frontier.SaveRenderedAnalysis(ctx, crawlID, projectID, leases[0], raw, rendered, issues, RenderMetadata{Status: "completed", FinalURL: target.RequestKey, RequestCount: 2, TransferredBytes: 512}); err != nil {
		t.Fatal(err)
	}
	pages, err := frontier.ListPages(ctx, crawlID, contracts.PageRequest{Limit: 10})
	if err != nil || len(pages.Items) != 1 {
		t.Fatalf("pages=%+v err=%v", pages, err)
	}
	page := pages.Items[0]
	if page.Title != "Raw" || page.ExtractionMode != "raw" || page.RenderStatus != "completed" || page.RenderedTitle != "Rendered" {
		t.Fatalf("page=%+v", page)
	}
	detail, err := frontier.GetPage(ctx, crawlID, page.ID)
	if err != nil {
		t.Fatal(err)
	}
	if detail.Rendered == nil || detail.Rendered.Title != "Rendered" || len(detail.RenderDifferences) == 0 || len(detail.Rendered.Headings) != 1 {
		t.Fatalf("detail=%+v", detail)
	}
	if len(detail.Outlinks) != 1 || detail.Outlinks[0].ExtractionMode != "rendered" {
		t.Fatalf("outlinks=%+v", detail.Outlinks)
	}
	if len(detail.Issues) != 1 || detail.Issues[0].SubjectType != "rendered_page" {
		t.Fatalf("issues=%+v", detail.Issues)
	}
}
