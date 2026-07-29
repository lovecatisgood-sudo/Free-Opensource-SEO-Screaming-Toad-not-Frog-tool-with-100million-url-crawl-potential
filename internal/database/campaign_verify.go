package database

import (
	"context"
	"fmt"

	"github.com/seo-auditor/seo-auditor/internal/contracts"
)

type CampaignVerification struct {
	CrawlID               contracts.ID          `json:"crawl_id"`
	Status                contracts.CrawlStatus `json:"status"`
	Discovered            int64                 `json:"discovered"`
	CommittedPages        int64                 `json:"committed_pages"`
	UniqueURLIdentities   int64                 `json:"unique_url_identities"`
	Outstanding           int64                 `json:"outstanding"`
	Links                 int64                 `json:"links"`
	Issues                int64                 `json:"issues"`
	CompletedSegments     int64                 `json:"completed_segments"`
	InvalidSegments       int64                 `json:"invalid_segments"`
	MissingRequiredFields int64                 `json:"missing_required_fields"`
	DatabaseIntegrity     string                `json:"database_integrity"`
	Passed                bool                  `json:"passed"`
}

// VerifyCampaign reconciles durable campaign invariants without trusting
// application counters alone.
func (f *Frontier) VerifyCampaign(ctx context.Context, crawlID contracts.ID) (CampaignVerification, error) {
	result := CampaignVerification{CrawlID: crawlID, DatabaseIntegrity: "ok"}
	if err := f.owner.Verify(ctx); err != nil {
		result.DatabaseIntegrity = "failed"
	}
	if err := f.db.QueryRowContext(ctx, `SELECT status,discovered_count FROM crawl WHERE id=?`, crawlID).Scan(&result.Status, &result.Discovered); err != nil {
		return result, err
	}
	queries := []struct {
		target *int64
		query  string
	}{
		{&result.CommittedPages, `SELECT count(*) FROM page p JOIN crawl_url cu ON cu.id=p.crawl_url_id WHERE cu.crawl_id=?`},
		{&result.UniqueURLIdentities, `SELECT count(DISTINCT url_id) FROM crawl_url WHERE crawl_id=?`},
		{&result.Outstanding, `SELECT count(*) FROM crawl_url WHERE crawl_id=? AND state NOT IN ('analysed','skipped','failed')`},
		{&result.Links, `SELECT count(*) FROM link WHERE crawl_id=?`},
		{&result.Issues, `SELECT count(*) FROM issue WHERE crawl_id=?`},
		{&result.CompletedSegments, `SELECT count(*) FROM campaign_segment WHERE crawl_id=? AND status='completed'`},
		{&result.MissingRequiredFields, `SELECT count(*) FROM page p JOIN crawl_url cu ON cu.id=p.crawl_url_id WHERE cu.crawl_id=? AND (COALESCE(p.content_hash,'')='' OR COALESCE(p.similarity_hash,'')='' OR COALESCE(p.title,'')='')`},
	}
	for _, item := range queries {
		if err := f.db.QueryRowContext(ctx, item.query, crawlID).Scan(item.target); err != nil {
			return result, err
		}
	}
	segments, err := f.ListSegments(ctx, crawlID)
	if err != nil {
		return result, err
	}
	for _, segment := range segments {
		if segment.Status != "completed" || segment.EndAnalysed == nil || segment.Checksum != segmentChecksum(crawlID, segment.Sequence, segment.StartAnalysed, *segment.EndAnalysed, segment.StorageBytes) {
			result.InvalidSegments++
		}
	}
	result.Passed = result.DatabaseIntegrity == "ok" && result.Status == contracts.CrawlCompleted && result.Discovered == result.CommittedPages && result.Discovered == result.UniqueURLIdentities && result.Outstanding == 0 && result.InvalidSegments == 0 && result.MissingRequiredFields == 0
	return result, nil
}

func (v CampaignVerification) Error() error {
	if v.Passed {
		return nil
	}
	return fmt.Errorf("campaign invariant verification failed")
}
