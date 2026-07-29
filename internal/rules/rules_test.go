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
	if len(Catalog) != 12 {
		t.Fatalf("catalog size = %d", len(Catalog))
	}
	for index, item := range Catalog {
		expected := "AUD-" + fmt.Sprintf("%02d", index+1)
		if item.ID != expected || item.Version < 1 || item.Remediation == "" || item.Limitations == "" {
			t.Fatalf("invalid metadata: %+v", item)
		}
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
