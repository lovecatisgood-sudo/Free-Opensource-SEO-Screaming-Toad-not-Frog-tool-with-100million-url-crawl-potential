package crawler

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/seo-auditor/seo-auditor/internal/fetchpolicy"
)

type SitemapDiscoveryLimits struct {
	MaximumSitemaps int
	MaximumURLs     int
	MaximumDepth    int
	Document        SitemapLimits
}

func DefaultSitemapDiscoveryLimits() SitemapDiscoveryLimits {
	return SitemapDiscoveryLimits{MaximumSitemaps: 1000, MaximumURLs: 100_000, MaximumDepth: 5, Document: DefaultSitemapLimits()}
}

type SitemapEvidence struct {
	URL        string
	StatusCode int
	Kind       SitemapKind
	Depth      int
	Entries    int
	Error      string
	Locations  []string
}

func DiscoverSitemaps(ctx context.Context, fetcher Fetcher, scope ScopeEvaluator, roots []string, limits SitemapDiscoveryLimits) ([]string, []SitemapEvidence, error) {
	if fetcher == nil || scope == nil || limits.MaximumSitemaps < 1 || limits.MaximumURLs < 1 || limits.MaximumDepth < 0 {
		return nil, nil, errors.New("invalid sitemap discovery configuration")
	}
	type queued struct {
		raw   string
		depth int
	}
	queue := make([]queued, 0, len(roots))
	seenSitemaps := make(map[string]struct{})
	seenURLs := make(map[string]struct{})
	var URLs []string
	var evidence []SitemapEvidence
	for _, root := range roots {
		queue = append(queue, queued{root, 0})
	}
	for len(queue) > 0 {
		item := queue[0]
		queue = queue[1:]
		normalized, err := fetchpolicy.NormalizeURL(item.raw)
		if err != nil || scope.Evaluate(normalized) != nil {
			continue
		}
		if _, exists := seenSitemaps[normalized.RequestKey]; exists {
			continue
		}
		if len(seenSitemaps) >= limits.MaximumSitemaps {
			return URLs, evidence, errors.New("sitemap document limit reached")
		}
		seenSitemaps[normalized.RequestKey] = struct{}{}
		result, err := fetcher.Fetch(ctx, normalized.RequestKey)
		record := SitemapEvidence{URL: normalized.RequestKey, Depth: item.depth}
		if err != nil {
			record.Error = err.Error()
			evidence = append(evidence, record)
			continue
		}
		record.StatusCode = result.StatusCode
		if result.StatusCode == http.StatusNotFound || result.StatusCode == http.StatusGone {
			evidence = append(evidence, record)
			continue
		}
		if result.StatusCode < 200 || result.StatusCode >= 300 {
			record.Error = fmt.Sprintf("HTTP %d", result.StatusCode)
			evidence = append(evidence, record)
			continue
		}
		document, err := ParseSitemap(bytes.NewReader(result.Body), limits.Document)
		if err != nil {
			record.Error = err.Error()
			evidence = append(evidence, record)
			continue
		}
		record.Kind, record.Entries, record.Locations = document.Kind, len(document.Locations), append([]string(nil), document.Locations...)
		evidence = append(evidence, record)
		if document.Kind == SitemapIndex {
			if item.depth >= limits.MaximumDepth {
				return URLs, evidence, errors.New("sitemap index depth limit reached")
			}
			for _, location := range document.Locations {
				queue = append(queue, queued{location, item.depth + 1})
			}
			continue
		}
		for _, location := range document.Locations {
			target, err := fetchpolicy.NormalizeURL(location)
			if err != nil || scope.Evaluate(target) != nil {
				continue
			}
			if _, exists := seenURLs[target.RequestKey]; exists {
				continue
			}
			if len(URLs) >= limits.MaximumURLs {
				return URLs, evidence, errors.New("sitemap URL limit reached")
			}
			seenURLs[target.RequestKey] = struct{}{}
			URLs = append(URLs, target.RequestKey)
		}
	}
	return URLs, evidence, nil
}
