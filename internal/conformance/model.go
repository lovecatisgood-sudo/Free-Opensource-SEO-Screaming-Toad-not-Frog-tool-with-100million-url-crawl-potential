package conformance

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strings"

	"github.com/seo-auditor/seo-auditor/internal/rules"
)

const CurrentSchemaVersion = 1

type Classification = rules.Classification

const (
	ClassificationDeterministic  = rules.ClassificationDeterministic
	ClassificationRecommendation = rules.ClassificationRecommendation
	ClassificationReview         = rules.ClassificationReview
	ClassificationInformation    = rules.ClassificationInformation
)

type EvidenceSource string

const (
	SourceRaw      EvidenceSource = "raw"
	SourceRendered EvidenceSource = "rendered"
	SourceGraph    EvidenceSource = "graph"
	SourceSitemap  EvidenceSource = "sitemap"
	SourceAPI      EvidenceSource = "external_api"
	SourceLab      EvidenceSource = "lab"
	SourceField    EvidenceSource = "field"
)

type Expectation struct {
	RuleID         string         `json:"rule_id"`
	Message        string         `json:"message,omitempty"`
	Severity       rules.Severity `json:"severity,omitempty"`
	Classification Classification `json:"classification,omitempty"`
	Evidence       map[string]any `json:"evidence,omitempty"`
}

type FixtureCase struct {
	ID                  string            `json:"id"`
	Path                string            `json:"path"`
	DocumentURL         string            `json:"document_url,omitempty"`
	StatusCode          int               `json:"status_code"`
	Headers             map[string]string `json:"headers,omitempty"`
	Body                string            `json:"body"`
	InSitemap           bool              `json:"in_sitemap"`
	RobotsBlocked       bool              `json:"robots_blocked,omitempty"`
	Depth               int               `json:"depth,omitempty"`
	Inlinks             int               `json:"inlinks,omitempty"`
	ExactDuplicateCount int               `json:"exact_duplicate_count,omitempty"`
	ExpectedFindings    []Expectation     `json:"expected_findings"`
	ExpectedAbsent      []Expectation     `json:"expected_absent,omitempty"`
}

type Manifest struct {
	SchemaVersion int           `json:"schema_version"`
	ID            string        `json:"id"`
	Description   string        `json:"description"`
	Cases         []FixtureCase `json:"cases"`
}

type ObservedFinding struct {
	URL            string         `json:"url"`
	Source         EvidenceSource `json:"source"`
	Classification Classification `json:"classification"`
	Issue          rules.Issue    `json:"issue"`
}

type MissingFinding struct {
	Expected Expectation `json:"expected"`
}

type CaseResult struct {
	CaseID             string            `json:"case_id"`
	URL                string            `json:"url"`
	ExpectedStatusCode int               `json:"expected_status_code"`
	ObservedStatusCode int               `json:"observed_status_code"`
	Matched            []ObservedFinding `json:"matched"`
	Unexpected         []ObservedFinding `json:"unexpected"`
	Missing            []MissingFinding  `json:"missing"`
	Passed             bool              `json:"passed"`
}

type Report struct {
	SchemaVersion int          `json:"schema_version"`
	ManifestID    string       `json:"manifest_id"`
	Cases         []CaseResult `json:"cases"`
	TruePositive  int          `json:"true_positive"`
	FalsePositive int          `json:"false_positive"`
	FalseNegative int          `json:"false_negative"`
	Precision     float64      `json:"precision"`
	Recall        float64      `json:"recall"`
	Passed        bool         `json:"passed"`
}

var stableIDPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9._-]{2,127}$`)

func ParseManifest(reader io.Reader) (Manifest, error) {
	decoder := json.NewDecoder(io.LimitReader(reader, 8<<20))
	decoder.DisallowUnknownFields()
	var manifest Manifest
	if err := decoder.Decode(&manifest); err != nil {
		return Manifest{}, fmt.Errorf("decode conformance manifest: %w", err)
	}
	if err := manifest.Validate(); err != nil {
		return Manifest{}, err
	}
	return manifest, nil
}

func (m Manifest) Validate() error {
	if m.SchemaVersion != CurrentSchemaVersion {
		return fmt.Errorf("unsupported conformance schema version %d", m.SchemaVersion)
	}
	if !stableIDPattern.MatchString(m.ID) {
		return errors.New("manifest ID is invalid")
	}
	if strings.TrimSpace(m.Description) == "" {
		return errors.New("manifest description is required")
	}
	if len(m.Cases) == 0 || len(m.Cases) > 10_000 {
		return errors.New("manifest must contain between 1 and 10000 cases")
	}
	seen := make(map[string]struct{}, len(m.Cases))
	for index, item := range m.Cases {
		if !stableIDPattern.MatchString(item.ID) {
			return fmt.Errorf("case %d ID is invalid", index)
		}
		if _, exists := seen[item.ID]; exists {
			return fmt.Errorf("duplicate case ID %q", item.ID)
		}
		seen[item.ID] = struct{}{}
		if !strings.HasPrefix(item.Path, "/") || strings.HasPrefix(item.Path, "//") {
			return fmt.Errorf("case %q path must be root-relative", item.ID)
		}
		if item.DocumentURL != "" && !strings.HasPrefix(item.DocumentURL, "http://") && !strings.HasPrefix(item.DocumentURL, "https://") {
			return fmt.Errorf("case %q document URL must be HTTP or HTTPS", item.ID)
		}
		if item.StatusCode < 100 || item.StatusCode > 599 {
			return fmt.Errorf("case %q status code is invalid", item.ID)
		}
		if len(item.Body) > 2<<20 {
			return fmt.Errorf("case %q body exceeds fixture limit", item.ID)
		}
		for _, group := range [][]Expectation{item.ExpectedFindings, item.ExpectedAbsent} {
			for _, expected := range group {
				if !regexp.MustCompile(`^AUD-[0-9]{2}$`).MatchString(expected.RuleID) {
					return fmt.Errorf("case %q has invalid rule ID %q", item.ID, expected.RuleID)
				}
				if expected.Classification != "" && !validClassification(expected.Classification) {
					return fmt.Errorf("case %q has invalid classification %q", item.ID, expected.Classification)
				}
			}
		}
	}
	return nil
}

func validClassification(c Classification) bool {
	switch c {
	case ClassificationDeterministic, ClassificationRecommendation, ClassificationReview, ClassificationInformation:
		return true
	default:
		return false
	}
}

func Classify(issue rules.Issue) Classification {
	return rules.Classify(issue)
}

func Compare(manifest Manifest, observed map[string][]ObservedFinding, statuses map[string]int, urls map[string]string) Report {
	report := Report{SchemaVersion: CurrentSchemaVersion, ManifestID: manifest.ID, Passed: true}
	for _, fixture := range manifest.Cases {
		actual := append([]ObservedFinding(nil), observed[fixture.ID]...)
		sortFindings(actual)
		used := make([]bool, len(actual))
		result := CaseResult{CaseID: fixture.ID, URL: urls[fixture.ID], ExpectedStatusCode: fixture.StatusCode, ObservedStatusCode: statuses[fixture.ID], Passed: statuses[fixture.ID] == fixture.StatusCode}
		for _, expected := range fixture.ExpectedFindings {
			matched := -1
			for index, finding := range actual {
				if !used[index] && matches(expected, finding) {
					matched = index
					break
				}
			}
			if matched < 0 {
				result.Missing = append(result.Missing, MissingFinding{Expected: expected})
				report.FalseNegative++
				result.Passed = false
				continue
			}
			used[matched] = true
			result.Matched = append(result.Matched, actual[matched])
			report.TruePositive++
		}
		for index, finding := range actual {
			if used[index] {
				continue
			}
			result.Unexpected = append(result.Unexpected, finding)
			report.FalsePositive++
			result.Passed = false
		}
		for _, absent := range fixture.ExpectedAbsent {
			for _, finding := range actual {
				if matches(absent, finding) {
					result.Passed = false
				}
			}
		}
		if !result.Passed {
			report.Passed = false
		}
		report.Cases = append(report.Cases, result)
	}
	if denominator := report.TruePositive + report.FalsePositive; denominator > 0 {
		report.Precision = float64(report.TruePositive) / float64(denominator)
	} else {
		report.Precision = 1
	}
	if denominator := report.TruePositive + report.FalseNegative; denominator > 0 {
		report.Recall = float64(report.TruePositive) / float64(denominator)
	} else {
		report.Recall = 1
	}
	return report
}

func matches(expected Expectation, finding ObservedFinding) bool {
	if expected.RuleID != finding.Issue.RuleID {
		return false
	}
	if expected.Message != "" && expected.Message != finding.Issue.Message {
		return false
	}
	if expected.Severity != "" && expected.Severity != finding.Issue.Severity {
		return false
	}
	if expected.Classification != "" && expected.Classification != finding.Classification {
		return false
	}
	return evidenceContains(finding.Issue.Evidence, expected.Evidence)
}

func evidenceContains(actual, expected map[string]any) bool {
	if len(expected) == 0 {
		return true
	}
	actualJSON, err := json.Marshal(actual)
	if err != nil {
		return false
	}
	expectedJSON, err := json.Marshal(expected)
	if err != nil {
		return false
	}
	var normalizedActual, normalizedExpected map[string]any
	if json.Unmarshal(actualJSON, &normalizedActual) != nil || json.Unmarshal(expectedJSON, &normalizedExpected) != nil {
		return false
	}
	for key, wanted := range normalizedExpected {
		got, exists := normalizedActual[key]
		if !exists || fmt.Sprint(got) != fmt.Sprint(wanted) {
			return false
		}
	}
	return true
}

func sortFindings(items []ObservedFinding) {
	sort.Slice(items, func(i, j int) bool {
		left, right := items[i], items[j]
		if left.Issue.RuleID != right.Issue.RuleID {
			return left.Issue.RuleID < right.Issue.RuleID
		}
		if left.Issue.Message != right.Issue.Message {
			return left.Issue.Message < right.Issue.Message
		}
		return left.Issue.Severity < right.Issue.Severity
	})
}
