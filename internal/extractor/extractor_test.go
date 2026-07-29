package extractor

import (
	"net/http"
	"testing"
)

func TestExtractTechnicalSEOFields(t *testing.T) {
	t.Parallel()
	body := []byte(`<!doctype html><html lang="en"><head><base href="https://example.com/docs/"><title> Example Title </title><meta name="description" content="A useful description"><meta name="robots" content="index,follow"><meta name="viewport" content="width=device-width"><meta property="og:title" content="Social title"><link rel="canonical" href="../canonical"><link rel="alternate" hreflang="en-GB" href="/uk"><link rel="next" href="page-2"><script type="application/ld+json">{"@type":"Article"}</script></head><body><h1>Example Title</h1><h2>Section</h2><p>Hello visible world.</p><script>hidden words</script><a href="page-2#part" rel="nofollow"> Next page </a><img src="hero.jpg" alt="" width="120" height="80"><div hidden>not visible</div><div itemscope itemtype="https://schema.org/Product"></div><div typeof="BreadcrumbList"></div></body></html>`)
	page, err := Extract("https://example.com/start", http.Header{"X-Robots-Tag": []string{"noarchive"}}, body)
	if err != nil {
		t.Fatal(err)
	}
	if page.Title != "Example Title" || page.MetaDescription != "A useful description" || page.Language != "en" || page.Viewport == "" {
		t.Fatalf("metadata: %+v", page)
	}
	if len(page.Canonicals) != 1 || page.Canonicals[0] != "https://example.com/canonical" {
		t.Fatalf("canonicals: %#v", page.Canonicals)
	}
	if len(page.Headings) != 2 || len(page.Links) != 1 || page.Links[0].URL != "https://example.com/docs/page-2" {
		t.Fatalf("content extraction: %+v", page)
	}
	if len(page.Images) != 1 || !page.Images[0].AltPresent || page.ContentHash == "" || page.WordCount == 0 {
		t.Fatalf("image/text: %+v", page)
	}
	if len(page.StructuredData) != 3 || !page.StructuredData[0].Valid || page.StructuredData[0].Types[0] != "Article" {
		t.Fatalf("structured data: %+v", page.StructuredData)
	}
}

func TestExtractReportsInvalidJSONLD(t *testing.T) {
	t.Parallel()
	page, err := Extract("https://example.com/", nil, []byte(`<script type="application/ld+json">{broken</script>`))
	if err != nil {
		t.Fatal(err)
	}
	if len(page.StructuredData) != 1 || page.StructuredData[0].Valid {
		t.Fatalf("structured data: %+v", page.StructuredData)
	}
}
