package crawler

import (
	"strings"
	"testing"
)

func TestParseSitemapURLSet(t *testing.T) {
	t.Parallel()

	document, err := ParseSitemap(strings.NewReader(`<?xml version="1.0"?>
      <urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
        <url><loc>https://example.com/a</loc></url>
        <url><loc>https://example.com/b?x=1&amp;y=2</loc></url>
      </urlset>`), DefaultSitemapLimits())
	if err != nil {
		t.Fatal(err)
	}
	if document.Kind != SitemapURLSet || len(document.Locations) != 2 || document.Locations[1] != "https://example.com/b?x=1&y=2" {
		t.Fatalf("unexpected document: %+v", document)
	}
}

func TestParseSitemapIndexAndLimits(t *testing.T) {
	t.Parallel()

	body := `<sitemapindex><sitemap><loc>https://example.com/one.xml</loc></sitemap><sitemap><loc>https://example.com/two.xml</loc></sitemap></sitemapindex>`
	document, err := ParseSitemap(strings.NewReader(body), SitemapLimits{MaximumBytes: int64(len(body)), MaximumEntries: 2})
	if err != nil || document.Kind != SitemapIndex {
		t.Fatalf("document = %+v, error = %v", document, err)
	}
	if _, err := ParseSitemap(strings.NewReader(body), SitemapLimits{MaximumBytes: int64(len(body)), MaximumEntries: 1}); err == nil {
		t.Fatal("expected entry limit error")
	}
	if _, err := ParseSitemap(strings.NewReader(body), SitemapLimits{MaximumBytes: 10, MaximumEntries: 2}); err == nil {
		t.Fatal("expected byte limit error")
	}
}

func TestParseSitemapRejectsMalformedXML(t *testing.T) {
	t.Parallel()
	if _, err := ParseSitemap(strings.NewReader(`<urlset><url></urlset>`), DefaultSitemapLimits()); err == nil {
		t.Fatal("expected malformed XML error")
	}
}
