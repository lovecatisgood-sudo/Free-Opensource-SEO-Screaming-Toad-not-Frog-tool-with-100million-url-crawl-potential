package conformance_test

import (
	"strings"
	"testing"

	"github.com/seo-auditor/seo-auditor/internal/conformance"
	"github.com/seo-auditor/seo-auditor/internal/rules"
	"github.com/seo-auditor/seo-auditor/internal/testfixtures/sites"
)

func TestCoreCoverageCatalogHasPositiveAndCleanControlForEveryRule(t *testing.T) {
	manifest, err := sites.CoreRules()
	if err != nil {
		t.Fatal(err)
	}
	report := conformance.BuildCoverage(rules.Catalog, manifest)
	if !report.Passed() || report.BaselineCovered != len(rules.Catalog) {
		t.Fatalf("coverage=%+v", report)
	}
	if markdown := report.Markdown(); !strings.Contains(markdown, "AUD-17") || !strings.Contains(markdown, "Baseline-covered: 17") {
		t.Fatalf("markdown=%s", markdown)
	}
	if _, err := report.JSON(); err != nil {
		t.Fatal(err)
	}
}
