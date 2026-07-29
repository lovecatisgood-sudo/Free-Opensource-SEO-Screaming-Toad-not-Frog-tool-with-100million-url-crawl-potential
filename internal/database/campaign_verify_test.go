package database

import (
	"context"
	"testing"

	"github.com/seo-auditor/seo-auditor/internal/contracts"
)

func TestVerifyCampaignRejectsIncompleteCampaign(t *testing.T) {
	t.Parallel()
	frontier, _, crawlID := testFrontier(t)
	result, err := frontier.VerifyCampaign(context.Background(), crawlID)
	if err != nil {
		t.Fatal(err)
	}
	if result.Passed || result.Status != contracts.CrawlPending {
		t.Fatalf("verification=%+v", result)
	}
}
