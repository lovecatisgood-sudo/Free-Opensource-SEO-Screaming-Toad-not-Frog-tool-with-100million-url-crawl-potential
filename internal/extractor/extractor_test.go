package extractor

import (
	"math/bits"
	"net/http"
	"strconv"
	"testing"
)

func TestExtractTechnicalSEOFields(t *testing.T) {
	t.Parallel()
	body := []byte(`<!doctype html><html lang="en"><head><base href="https://example.com/docs/"><title> Example Title </title><meta name="description" content="A useful description"><meta name="robots" content="index,follow"><meta name="viewport" content="width=device-width"><meta property="og:title" content="Social title"><link rel="canonical" href="../canonical"><link rel="alternate" hreflang="en-GB" href="/uk"><link rel="next" href="page-2"><script type="application/ld+json">{"@context":"https://schema.org","@type":"Article","headline":"Example"}</script></head><body><h1>Example Title</h1><h2>Section</h2><p>Hello visible world.</p><script>hidden words</script><a href="page-2#part" rel="nofollow"> Next page </a><img src="hero.jpg" alt="" width="120" height="80"><div hidden>not visible</div><div itemscope itemtype="https://schema.org/Product"></div><div typeof="BreadcrumbList"></div></body></html>`)
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
	if len(page.Images) != 1 || !page.Images[0].AltPresent || page.ContentHash == "" || page.SimilarityHash == "" || page.WordCount == 0 {
		t.Fatalf("image/text: %+v", page)
	}
	if len(page.StructuredData) != 3 || !page.StructuredData[0].Valid || page.StructuredData[0].Types[0] != "Article" {
		t.Fatalf("structured data: %+v", page.StructuredData)
	}
	if len(page.StructuredData[0].Contexts) != 1 || page.StructuredData[0].Contexts[0] != "https://schema.org" || len(page.StructuredData[0].Properties) != 1 || page.StructuredData[0].Properties[0] != "headline" {
		t.Fatalf("structured evidence: %+v", page.StructuredData[0])
	}
	if len(page.StructuredData[0].Nodes) != 1 || page.StructuredData[0].Nodes[0].Path != "$" || page.StructuredData[0].Nodes[0].Properties[0] != "headline" {
		t.Fatalf("structured nodes: %+v", page.StructuredData[0].Nodes)
	}
}

func TestExtractReportsInvalidJSONLDStructure(t *testing.T) {
	t.Parallel()
	page, err := Extract("https://example.com/", nil, []byte(`<script type="application/ld+json">{"@context":7,"@type":["Article",7]}</script>`))
	if err != nil {
		t.Fatal(err)
	}
	if len(page.StructuredData) != 1 || len(page.StructuredData[0].StructuralErrors) != 2 {
		t.Fatalf("structured data: %+v", page.StructuredData)
	}
}

func TestExtractDoesNotTreatInlineContextAliasesAsProperties(t *testing.T) {
	t.Parallel()
	page, err := Extract("https://example.com/", nil, []byte(`<script type="application/ld+json">{"@context":{"schema":"https://schema.org/","headline":"schema:headline"},"@type":"schema:Article","headline":"Example"}</script>`))
	if err != nil {
		t.Fatal(err)
	}
	if len(page.StructuredData) != 1 {
		t.Fatalf("structured data: %+v", page.StructuredData)
	}
	properties := page.StructuredData[0].Properties
	if len(properties) != 1 || properties[0] != "headline" {
		t.Fatalf("context aliases leaked into properties: %#v", properties)
	}
}

func TestSimilarityHashIsDeterministicAndLocal(t *testing.T) {
	t.Parallel()
	base := similarityHash("one two three four five six seven eight")
	close := similarityHash("one two three four five six seven nine")
	far := similarityHash("completely unrelated vocabulary about another subject")
	if base == "" || base != similarityHash("one two three four five six seven eight") {
		t.Fatal("similarity hash is empty or non-deterministic")
	}
	if hammingHex(base, close) >= hammingHex(base, far) {
		t.Fatalf("near text should be closer: close=%d far=%d", hammingHex(base, close), hammingHex(base, far))
	}
}

func hammingHex(left, right string) int {
	leftValue, _ := strconv.ParseUint(left, 16, 64)
	rightValue, _ := strconv.ParseUint(right, 16, 64)
	return bits.OnesCount64(leftValue ^ rightValue)
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

func TestExtractMobileAMPAndResponsiveImageSignals(t *testing.T) {
	page, err := Extract("https://example.com/page", nil, []byte(`<html><head><link rel="amphtml" href="/amp"><link rel="alternate" media="only screen and (max-width: 640px)" href="/mobile"></head><body><img src="/hero.webp" alt="Hero" loading="lazy" srcset="/hero-2x.webp 2x"></body></html>`))
	if err != nil {
		t.Fatal(err)
	}
	if page.AMPURL != "https://example.com/amp" || page.MobileAlternate != "https://example.com/mobile" || len(page.Images) != 1 || page.Images[0].Loading != "lazy" || !page.Images[0].Srcset {
		t.Fatalf("signals=%+v", page)
	}
}
