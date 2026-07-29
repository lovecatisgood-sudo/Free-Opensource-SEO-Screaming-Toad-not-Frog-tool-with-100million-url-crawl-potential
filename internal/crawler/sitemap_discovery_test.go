package crawler

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/seo-auditor/seo-auditor/internal/fetchpolicy"
)

type sitemapFetcher map[string]fetchpolicy.FetchResult

func (f sitemapFetcher) Fetch(_ context.Context, raw string) (fetchpolicy.FetchResult, error) {
	result := f[raw]
	result.FinalURL = raw
	result.Header = http.Header{"Content-Type": []string{"application/xml"}}
	result.ContentType = "application/xml"
	result.StartedAt, result.FinishedAt = time.Now(), time.Now()
	return result, nil
}

func TestDiscoverSitemapsRecursesAndDeduplicates(t *testing.T) {
	t.Parallel()
	scope, _ := fetchpolicy.CompileScope(fetchpolicy.ScopeConfig{AllowedHosts: []string{"example.com"}})
	fetcher := sitemapFetcher{
		"https://example.com/sitemap.xml": {StatusCode: 200, Body: []byte(`<sitemapindex><sitemap><loc>https://example.com/a.xml</loc></sitemap><sitemap><loc>https://other.com/x.xml</loc></sitemap></sitemapindex>`)},
		"https://example.com/a.xml":       {StatusCode: 200, Body: []byte(`<urlset><url><loc>https://example.com/a</loc></url><url><loc>https://example.com/a#fragment</loc></url><url><loc>https://example.com/b</loc></url></urlset>`)},
	}
	URLs, evidence, err := DiscoverSitemaps(context.Background(), fetcher, scope, []string{"https://example.com/sitemap.xml"}, DefaultSitemapDiscoveryLimits())
	if err != nil {
		t.Fatal(err)
	}
	if len(URLs) != 2 || len(evidence) != 2 || URLs[0] != "https://example.com/a" || URLs[1] != "https://example.com/b" {
		t.Fatalf("urls=%#v evidence=%#v", URLs, evidence)
	}
}
