package rules

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/seo-auditor/seo-auditor/internal/extractor"
)

func TestCatalogHasStableMustRuleMetadata(t *testing.T) {
	t.Parallel()
	if len(Catalog) != 13 {
		t.Fatalf("catalog size = %d", len(Catalog))
	}
	for index, item := range Catalog {
		expected := "AUD-" + fmt.Sprintf("%02d", index+1)
		if item.ID != expected || item.Version < 1 || item.Remediation == "" || item.Limitations == "" {
			t.Fatalf("invalid metadata: %+v", item)
		}
	}
}

func TestStructuredDataSyntaxAndContextFindingsAreSeparate(t *testing.T) {
	t.Parallel()
	page := extractor.Page{
		URL:             "https://example.com/",
		Title:           strings.Repeat("T", 40),
		MetaDescription: strings.Repeat("D", 100),
		Canonicals:      []string{"https://example.com/"},
		Headings:        []extractor.Heading{{Level: 1, Text: "Heading"}},
		StructuredData: []extractor.StructuredData{
			{Format: "json-ld", Valid: false, Error: "invalid character"},
			{Format: "json-ld", Valid: true, Types: []string{"Article"}},
			{Format: "json-ld", Valid: true, Types: []string{"Product"}, Contexts: []string{"https://schema.org"}},
			{Format: "json-ld", Valid: true, StructuralErrors: []string{"@type must be a string or an array of strings"}},
		},
	}
	issues := EvaluatePage(PageInput{Page: page, StatusCode: 200, Headers: http.Header{"Content-Security-Policy": []string{"default-src 'self'"}}, InSitemap: true}, DefaultThresholds())
	var syntax, structure, context int
	for _, issue := range issues {
		if issue.RuleID != "AUD-13" {
			continue
		}
		switch issue.Message {
		case "JSON-LD syntax is invalid":
			syntax++
		case "JSON-LD structure is invalid":
			structure++
		case "Compact structured-data types have no observed Schema.org context":
			context++
		}
	}
	if syntax != 1 || structure != 1 || context != 1 {
		t.Fatalf("AUD-13 findings syntax=%d structure=%d context=%d: %+v", syntax, structure, context, issues)
	}
}

func TestEvaluatePageEmitsTypedEvidence(t *testing.T) {
	t.Parallel()
	page := extractor.Page{URL: "https://example.com/deep", Title: "Short", MetaDescription: "tiny", MetaRobots: "noindex", ContentHash: "abc", Headings: []extractor.Heading{{Level: 1, Text: "Short"}, {Level: 1, Text: "Again"}}, Canonicals: []string{"bad"}, Hreflangs: []extractor.Hreflang{{Language: "wrong_value", URL: "https://example.com/"}}, Images: []extractor.Image{{URL: "http://example.com/a.jpg"}}, Links: []extractor.Link{{URL: "http://example.com/", Rel: "nofollow"}}}
	issues := EvaluatePage(PageInput{Page: page, StatusCode: 404, Headers: http.Header{}, Depth: 8, Inlinks: 0, RobotsBlocked: true, ExactDuplicateCount: 2}, DefaultThresholds())
	seen := make(map[string]bool)
	for _, issue := range issues {
		seen[issue.RuleID] = true
		if issue.RuleVersion != 1 || len(issue.Evidence) == 0 {
			t.Fatalf("invalid issue: %+v", issue)
		}
	}
	for _, id := range []string{"AUD-01", "AUD-02", "AUD-03", "AUD-04", "AUD-05", "AUD-07", "AUD-08", "AUD-09", "AUD-10", "AUD-11", "AUD-12"} {
		if !seen[id] {
			t.Errorf("missing %s", id)
		}
	}
}

func TestHealthyPageAvoidsErrorSeverity(t *testing.T) {
	t.Parallel()
	page := extractor.Page{URL: "https://example.com/", Title: strings.Repeat("T", 40), MetaDescription: strings.Repeat("D", 100), Canonicals: []string{"https://example.com/"}, Headings: []extractor.Heading{{Level: 1, Text: "Different heading"}}}
	issues := EvaluatePage(PageInput{Page: page, StatusCode: 200, Headers: http.Header{"Content-Security-Policy": []string{"default-src 'self'"}}, InSitemap: true}, DefaultThresholds())
	for _, issue := range issues {
		if issue.Severity == SeverityError {
			t.Fatalf("unexpected error issue: %+v", issue)
		}
	}
}
