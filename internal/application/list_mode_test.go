package application

import (
	"context"
	"testing"

	"github.com/seo-auditor/seo-auditor/internal/contracts"
)

func TestPrepareListModeDeduplicatesInOrderAndExpandsHosts(t *testing.T) {
	t.Parallel()
	service, err := Open(context.Background(), t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer service.Close()
	limits := contracts.DefaultCrawlLimits()
	prepared, err := service.prepare(context.Background(), CrawlRequest{ProjectName: "List", SeedURLs: []string{"https://example.com/a", "https://example.org/b", "https://example.com/a#duplicate"}, Limits: limits})
	if err != nil {
		t.Fatal(err)
	}
	if len(prepared.seeds) != 2 || prepared.seeds[0].RequestKey != "https://example.com/a" || prepared.seeds[1].RequestKey != "https://example.org/b" {
		t.Fatalf("seeds=%+v", prepared.seeds)
	}
	if len(prepared.config.AllowedHosts) != 2 {
		t.Fatalf("allowed hosts=%v", prepared.config.AllowedHosts)
	}
	progress, err := service.Progress(context.Background(), prepared.result.CrawlID)
	if err != nil {
		t.Fatal(err)
	}
	if progress.Discovered != 2 {
		t.Fatalf("discovered=%d", progress.Discovered)
	}
}
