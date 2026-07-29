package campaign

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/seo-auditor/seo-auditor/internal/contracts"
	"github.com/seo-auditor/seo-auditor/internal/database"
	"github.com/seo-auditor/seo-auditor/internal/fetchpolicy"
	"github.com/seo-auditor/seo-auditor/internal/segmentstore"
)

func TestCoordinatorCommitsFileAndMetadataIdempotently(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db, err := database.Open(ctx, filepath.Join(t.TempDir(), "campaign.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	frontier := database.NewFrontier(db, 32)
	defer frontier.Close()
	projectID, crawlID := contracts.ID("project_campaign"), contracts.ID("crawl_campaign")
	if err := frontier.CreateProject(ctx, projectID, "Campaign"); err != nil {
		t.Fatal(err)
	}
	seed, _ := fetchpolicy.NormalizeURL("https://example.com/")
	configuration := contracts.CrawlConfiguration{SeedURL: seed.RequestKey, AllowedHosts: []string{"example.com"}, RenderingMode: "raw", UserAgent: "test", Limits: contracts.DefaultCrawlLimits()}
	if err := frontier.CreateCrawl(ctx, crawlID, projectID, "", seed, configuration); err != nil {
		t.Fatal(err)
	}
	store, err := segmentstore.Open(filepath.Join(t.TempDir(), "objects"))
	if err != nil {
		t.Fatal(err)
	}
	coordinator := Coordinator{Frontier: frontier, Segments: store}
	request := SegmentRequest{CrawlID: crawlID, Sequence: 0, MaximumBytes: 1024, RowCount: 2, SchemaVersion: 1, RuleVersion: 1, Body: strings.NewReader("row1\nrow2\n")}
	first, err := coordinator.CommitSegment(ctx, request)
	if err != nil || !first.Inserted {
		t.Fatalf("first=%+v err=%v", first, err)
	}
	request.Body = strings.NewReader("row1\nrow2\n")
	replay, err := coordinator.CommitSegment(ctx, request)
	if err != nil || replay.Inserted || replay.Checksum != first.Checksum {
		t.Fatalf("replay=%+v err=%v", replay, err)
	}
}
