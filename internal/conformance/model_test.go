package conformance

import (
	"strings"
	"testing"

	"github.com/seo-auditor/seo-auditor/internal/rules"
)

func TestParseManifestRejectsUnknownFieldsAndInvalidVersion(t *testing.T) {
	t.Parallel()
	for _, body := range []string{
		`{"schema_version":2,"id":"fixture.invalid","description":"invalid","cases":[{"id":"case.one","path":"/","status_code":200,"body":""}]}`,
		`{"schema_version":1,"id":"fixture.invalid","description":"invalid","unknown":true,"cases":[{"id":"case.one","path":"/","status_code":200,"body":""}]}`,
	} {
		if _, err := ParseManifest(strings.NewReader(body)); err == nil {
			t.Fatalf("expected manifest to fail: %s", body)
		}
	}
}

func TestCompareCountsMissingAndUnexpectedFindings(t *testing.T) {
	t.Parallel()
	manifest := Manifest{SchemaVersion: 1, ID: "fixture.compare", Description: "comparison", Cases: []FixtureCase{{
		ID: "case.compare", Path: "/", StatusCode: 200,
		ExpectedFindings: []Expectation{{RuleID: "AUD-05", Message: "Page is not indexable", Severity: rules.SeverityError, Classification: ClassificationDeterministic}},
	}}}
	observed := map[string][]ObservedFinding{"case.compare": {{Issue: rules.Issue{RuleID: "AUD-04", Severity: rules.SeverityInfo, Message: "Canonical is absent", Evidence: map[string]any{"canonical_count": 0}}, Classification: ClassificationRecommendation}}}
	report := Compare(manifest, observed, map[string]int{"case.compare": 200}, map[string]string{"case.compare": "https://example.test/"})
	if report.Passed || report.TruePositive != 0 || report.FalsePositive != 1 || report.FalseNegative != 1 || report.Precision != 0 || report.Recall != 0 {
		t.Fatalf("report=%+v", report)
	}
}
