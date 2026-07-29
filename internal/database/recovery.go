package database

import (
	"context"

	"github.com/seo-auditor/seo-auditor/internal/contracts"
)

type MissingLinkDiscovery struct {
	URL              string
	Depth            int
	DiscoveredFromID int64
}

// MissingLinkDiscoveries returns internal link targets retained in evidence
// but absent from the frontier, which can occur only in databases written by
// builds predating atomic analysis/discovery commits.
func (f *Frontier) MissingLinkDiscoveries(ctx context.Context, crawlID contracts.ID, limit int) ([]MissingLinkDiscovery, error) {
	if limit < 1 || limit > 10_000 {
		limit = 10_000
	}
	rows, err := f.db.QueryContext(ctx, `SELECT DISTINCT target.request_key,source_cu.depth+1,source_cu.id
FROM link l
JOIN url target ON target.id=l.target_url_id
JOIN crawl_url source_cu ON source_cu.crawl_id=l.crawl_id AND source_cu.url_id=l.source_url_id
LEFT JOIN crawl_url target_cu ON target_cu.crawl_id=l.crawl_id AND target_cu.url_id=l.target_url_id
WHERE l.crawl_id=? AND l.link_kind='internal' AND target_cu.id IS NULL
ORDER BY source_cu.id,target.request_key LIMIT ?`, crawlID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []MissingLinkDiscovery
	for rows.Next() {
		var item MissingLinkDiscovery
		if err := rows.Scan(&item.URL, &item.Depth, &item.DiscoveredFromID); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}
