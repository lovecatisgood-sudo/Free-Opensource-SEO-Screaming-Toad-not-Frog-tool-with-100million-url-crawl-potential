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
	if len(Catalog) != 17 {
		t.Fatalf("catalog size = %d", len(Catalog))
	}
	for index, item := range Catalog {
		expected := "AUD-" + fmt.Sprintf("%02d", index+1)
		if item.ID != expected || item.Version < 1 || item.Remediation == "" || item.Limitations == "" {
			t.Fatalf("invalid metadata: %+v", item)
		}
	}
}

func TestMobileImageAndPDFDiagnosticsRemainClassified(t *testing.T) {
	page := extractor.Page{URL: "https://example.com/", Title: strings.Repeat("T", 40), MetaDescription: strings.Repeat("D", 100), Canonicals: []string{"https://example.com/"}, Headings: []extractor.Heading{{Level: 1, Text: "Heading"}}, AMPURL: "https://example.com/", Images: []extractor.Image{{URL: "https://example.com/a.jpg", AltPresent: true}}}
	issues := EvaluatePage(PageInput{Page: page, StatusCode: 200, Headers: http.Header{}, InSitemap: true}, DefaultThresholds())
	seen := map[string]bool{}
	for _, issue := range issues {
		if issue.RuleID == "AUD-15" || issue.RuleID == "AUD-16" {
			seen[issue.RuleID] = true
		}
	}
	if !seen["AUD-15"] || !seen["AUD-16"] {
		t.Fatalf("missing mobile/image diagnostics: %+v", issues)
	}
	resource := EvaluateResource(ResourceInput{URL: "https://example.com/file.pdf", StatusCode: 200, ContentType: "application/pdf", Body: []byte("%PDF-1.7 /Encrypt"), DecodedBytes: 20})
	if len(resource) != 4 {
		t.Fatalf("PDF diagnostics=%+v", resource)
	}
	for _, issue := range resource {
		if issue.RuleID != "AUD-17" {
			t.Fatalf("unexpected resource issue: %+v", issue)
		}
	}
}

func TestGoogleProfileRequiredAndRecommendedFindings(t *testing.T) {
	t.Parallel()
	page := extractor.Page{
		URL: "https://example.com/product", Title: strings.Repeat("T", 40), MetaDescription: strings.Repeat("D", 100),
		Canonicals: []string{"https://example.com/product"}, Headings: []extractor.Heading{{Level: 1, Text: "Product"}},
		StructuredData: []extractor.StructuredData{{Format: "json-ld", Valid: true, Contexts: []string{"https://schema.org"}, Types: []string{"Product"}, Properties: []string{"name"}, Nodes: []extractor.StructuredNode{{Path: "$", Types: []string{"Product"}, Properties: []string{"name"}}}}},
	}
	issues := EvaluatePage(PageInput{Page: page, StatusCode: 200, Headers: http.Header{"Content-Security-Policy": []string{"default-src 'self'"}}, InSitemap: true}, DefaultThresholds())
	var required, recommended int
	for _, issue := range issues {
		if issue.RuleID != "AUD-14" {
			continue
		}
		if issue.Message == "Required Google search-feature properties are absent" {
			required++
			if issue.Classification != ClassificationDeterministic || issue.Evidence["profile_version"] != "2026-07-30" {
				t.Fatalf("bad required profile issue: %+v", issue)
			}
		}
		if issue.Message == "Recommended Google search-feature properties are absent" {
			recommended++
			if issue.Classification != ClassificationRecommendation {
				t.Fatalf("bad recommended profile issue: %+v", issue)
			}
		}
	}
	if required != 1 || recommended != 1 {
		t.Fatalf("required=%d recommended=%d: %+v", required, recommended, issues)
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
		if issue.RuleVersion < 1 || len(issue.Evidence) == 0 {
			t.Fatalf("invalid issue: %+v", issue)
		}
	}
	for _, id := range []string{"AUD-01", "AUD-02", "AUD-03", "AUD-04", "AUD-05", "AUD-07", "AUD-08", "AUD-09", "AUD-10", "AUD-11", "AUD-12"} {
		if !seen[id] {
			t.Errorf("missing %s", id)
		}
	}
}

func TestStructuredDataVocabularyFindingsUsePinnedVersion(t *testing.T) {
	t.Parallel()
	page := extractor.Page{
		URL:             "https://example.com/",
		Title:           strings.Repeat("T", 40),
		MetaDescription: strings.Repeat("D", 100),
		Canonicals:      []string{"https://example.com/"},
		Headings:        []extractor.Heading{{Level: 1, Text: "Heading"}},
		StructuredData: []extractor.StructuredData{{
			Format: "json-ld", Valid: true, Contexts: []string{"https://schema.org"},
			Types: []string{"Article", "NotARealSchemaType"}, Properties: []string{"headline", "notARealSchemaProperty", "episodes"},
		}},
	}
	issues := EvaluatePage(PageInput{Page: page, StatusCode: 200, Headers: http.Header{"Content-Security-Policy": []string{"default-src 'self'"}}, InSitemap: true}, DefaultThresholds())
	wanted := map[string]bool{
		"Schema.org type is unknown in bundled vocabulary":     false,
		"Schema.org property is unknown in bundled vocabulary": false,
		"Schema.org term is superseded":                        false,
	}
	for _, issue := range issues {
		if _, exists := wanted[issue.Message]; !exists {
			continue
		}
		wanted[issue.Message] = true
		if issue.RuleVersion != 3 || issue.Evidence["vocabulary_version"] != "30.0" {
			t.Fatalf("unversioned vocabulary issue: %+v", issue)
		}
	}
	for message, seen := range wanted {
		if !seen {
			t.Errorf("missing %q: %+v", message, issues)
		}
	}
}

func TestStructuredDataIgnoresExternalVocabularyTerms(t *testing.T) {
	t.Parallel()
	page := extractor.Page{
		URL: "https://example.com/", Title: strings.Repeat("T", 40), MetaDescription: strings.Repeat("D", 100),
		Canonicals: []string{"https://example.com/"}, Headings: []extractor.Heading{{Level: 1, Text: "Heading"}},
		StructuredData: []extractor.StructuredData{{Format: "json-ld", Valid: true, Contexts: []string{"https://example.net/context"}, Types: []string{"external:Thing"}, Properties: []string{"external:value"}}},
	}
	issues := EvaluatePage(PageInput{Page: page, StatusCode: 200, Headers: http.Header{"Content-Security-Policy": []string{"default-src 'self'"}}, InSitemap: true}, DefaultThresholds())
	for _, issue := range issues {
		if strings.Contains(issue.Message, "unknown in bundled vocabulary") {
			t.Fatalf("external vocabulary produced Schema.org finding: %+v", issue)
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

func TestSchemaDomainAndRangeGuidanceUsesTypeHierarchy(t *testing.T) {
	t.Parallel()
	page := extractor.Page{URL: "https://example.com/product", Title: strings.Repeat("T", 40), MetaDescription: strings.Repeat("D", 100), Viewport: "width=device-width", Canonicals: []string{"https://example.com/product"}, Headings: []extractor.Heading{{Level: 1, Text: "Product"}}, StructuredData: []extractor.StructuredData{{Format: "json-ld", Valid: true, Contexts: []string{"https://schema.org"}, Nodes: []extractor.StructuredNode{{Path: "$", Types: []string{"Product"}, Properties: []string{"birthDate", "offers"}}, {Path: "$.offers", Types: []string{"Person"}, Properties: []string{"name"}}}}}}
	issues := EvaluatePage(PageInput{Page: page, StatusCode: 200, Headers: http.Header{}, InSitemap: true}, DefaultThresholds())
	domain, rangeFinding := false, false
	for _, issue := range issues {
		if strings.Contains(issue.Message, "Schema.org domain") {
			domain = true
		}
		if strings.Contains(issue.Message, "Schema.org range") {
			rangeFinding = true
		}
		if (strings.Contains(issue.Message, "Schema.org domain") || strings.Contains(issue.Message, "Schema.org range")) && issue.Classification != ClassificationRecommendation {
			t.Fatalf("relationship classification=%+v", issue)
		}
	}
	if !domain || !rangeFinding {
		t.Fatalf("missing relationship findings: %+v", issues)
	}
}
