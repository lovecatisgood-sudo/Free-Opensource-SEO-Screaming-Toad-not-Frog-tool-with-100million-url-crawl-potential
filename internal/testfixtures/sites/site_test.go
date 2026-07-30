package sites_test

import (
	"context"
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/seo-auditor/seo-auditor/internal/conformance"
	"github.com/seo-auditor/seo-auditor/internal/testfixtures/sites"
)

func TestCoreIndexabilityFixturePassesConformance(t *testing.T) {
	manifest, err := sites.CoreIndexability()
	if err != nil {
		t.Fatal(err)
	}
	handler, err := sites.Handler(manifest)
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(handler)
	defer server.Close()
	report, err := conformance.RunHTTP(context.Background(), manifest, server.URL, server.Client())
	if err != nil {
		t.Fatal(err)
	}
	if !report.Passed || report.Precision != 1 || report.Recall != 1 {
		t.Fatalf("report failed:\n%s", report.Markdown())
	}
	if len(report.Cases) != 4 || report.TruePositive != 4 || report.FalsePositive != 0 || report.FalseNegative != 0 {
		t.Fatalf("unexpected report: %+v", report)
	}
	if markdown := report.Markdown(); !strings.Contains(markdown, "**PASS**") || !strings.Contains(markdown, "canonical.conflicting") {
		t.Fatalf("markdown=%s", markdown)
	}
	if _, err := report.JSON(); err != nil {
		t.Fatal(err)
	}
}

func TestCoreRulesFixtureCoversEveryCatalogRule(t *testing.T) {
	manifest, err := sites.CoreRules()
	if err != nil {
		t.Fatal(err)
	}
	handler, err := sites.Handler(manifest)
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(handler)
	defer server.Close()
	report, err := conformance.RunHTTP(context.Background(), manifest, server.URL, server.Client())
	if err != nil {
		t.Fatal(err)
	}
	if !report.Passed || report.Precision != 1 || report.Recall != 1 {
		t.Fatalf("report failed:\n%s", report.Markdown())
	}
	covered := make(map[string]bool)
	for _, item := range report.Cases {
		for _, finding := range item.Matched {
			covered[finding.Issue.RuleID] = true
		}
	}
	for index := 1; index <= 14; index++ {
		id := fmt.Sprintf("AUD-%02d", index)
		if !covered[id] {
			t.Errorf("missing positive fixture for %s", id)
		}
	}
}
