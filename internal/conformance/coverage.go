package conformance

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/seo-auditor/seo-auditor/internal/rules"
)

type CoverageStatus string

const (
	CoverageBaseline CoverageStatus = "baseline_covered"
	CoveragePartial  CoverageStatus = "partial"
	CoverageMissing  CoverageStatus = "missing"
)

type CoverageEntry struct {
	RuleID            string         `json:"rule_id"`
	Title             string         `json:"title"`
	Status            CoverageStatus `json:"status"`
	PositiveCases     []string       `json:"positive_cases"`
	CleanControlCases []string       `json:"clean_control_cases"`
}

type CoverageReport struct {
	SchemaVersion   int             `json:"schema_version"`
	Entries         []CoverageEntry `json:"entries"`
	BaselineCovered int             `json:"baseline_covered"`
	Partial         int             `json:"partial"`
	Missing         int             `json:"missing"`
}

func BuildCoverage(catalog []rules.Metadata, manifests ...Manifest) CoverageReport {
	result := CoverageReport{SchemaVersion: CurrentSchemaVersion}
	for _, metadata := range catalog {
		entry := CoverageEntry{RuleID: metadata.ID, Title: metadata.Title}
		positiveSet, controlSet := make(map[string]struct{}), make(map[string]struct{})
		for _, manifest := range manifests {
			for _, fixture := range manifest.Cases {
				for _, expected := range fixture.ExpectedFindings {
					if expected.RuleID == metadata.ID {
						positiveSet[manifest.ID+"/"+fixture.ID] = struct{}{}
					}
				}
				for _, absent := range fixture.ExpectedAbsent {
					if absent.RuleID == metadata.ID {
						controlSet[manifest.ID+"/"+fixture.ID] = struct{}{}
					}
				}
			}
		}
		entry.PositiveCases = sortedKeys(positiveSet)
		entry.CleanControlCases = sortedKeys(controlSet)
		switch {
		case len(entry.PositiveCases) > 0 && len(entry.CleanControlCases) > 0:
			entry.Status = CoverageBaseline
			result.BaselineCovered++
		case len(entry.PositiveCases) > 0 || len(entry.CleanControlCases) > 0:
			entry.Status = CoveragePartial
			result.Partial++
		default:
			entry.Status = CoverageMissing
			result.Missing++
		}
		result.Entries = append(result.Entries, entry)
	}
	return result
}

func (r CoverageReport) Passed() bool { return r.Partial == 0 && r.Missing == 0 }

func (r CoverageReport) JSON() ([]byte, error) { return json.MarshalIndent(r, "", "  ") }

func (r CoverageReport) Markdown() string {
	var output strings.Builder
	output.WriteString("## Rule-family coverage\n\n")
	fmt.Fprintf(&output, "- Baseline-covered: %d\n- Partial: %d\n- Missing: %d\n\n", r.BaselineCovered, r.Partial, r.Missing)
	output.WriteString("| Rule | Title | Status | Positive cases | Clean controls |\n")
	output.WriteString("|---|---|---|---:|---:|\n")
	for _, item := range r.Entries {
		fmt.Fprintf(&output, "| %s | %s | %s | %d | %d |\n", item.RuleID, item.Title, item.Status, len(item.PositiveCases), len(item.CleanControlCases))
	}
	return output.String()
}

func sortedKeys(values map[string]struct{}) []string {
	result := make([]string, 0, len(values))
	for value := range values {
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}
