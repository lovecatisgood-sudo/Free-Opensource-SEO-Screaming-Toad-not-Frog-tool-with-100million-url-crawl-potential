package conformance

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/seo-auditor/seo-auditor/internal/extractor"
	"github.com/seo-auditor/seo-auditor/internal/rules"
)

func RunHTTP(ctx context.Context, manifest Manifest, baseURL string, client *http.Client) (Report, error) {
	if err := manifest.Validate(); err != nil {
		return Report{}, err
	}
	baseURL = strings.TrimRight(baseURL, "/")
	if baseURL == "" {
		return Report{}, errors.New("fixture base URL is required")
	}
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	observed := make(map[string][]ObservedFinding, len(manifest.Cases))
	statuses := make(map[string]int, len(manifest.Cases))
	urls := make(map[string]string, len(manifest.Cases))
	for _, fixture := range manifest.Cases {
		target := baseURL + fixture.Path
		request, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
		if err != nil {
			return Report{}, fmt.Errorf("case %s: create request: %w", fixture.ID, err)
		}
		response, err := client.Do(request)
		if err != nil {
			return Report{}, fmt.Errorf("case %s: fetch fixture: %w", fixture.ID, err)
		}
		body, readErr := io.ReadAll(io.LimitReader(response.Body, 2<<20+1))
		_ = response.Body.Close()
		if readErr != nil {
			return Report{}, fmt.Errorf("case %s: read fixture: %w", fixture.ID, readErr)
		}
		if len(body) > 2<<20 {
			return Report{}, fmt.Errorf("case %s: response exceeds fixture limit", fixture.ID)
		}
		finalURL := response.Request.URL.String()
		documentURL := strings.ReplaceAll(fixture.DocumentURL, "{{BASE_URL}}", baseURL)
		if documentURL == "" {
			documentURL = finalURL
		}
		var issues []rules.Issue
		contentType := strings.ToLower(response.Header.Get("Content-Type"))
		if strings.Contains(contentType, "text/html") || strings.Contains(contentType, "application/xhtml+xml") {
			page, err := extractor.Extract(documentURL, response.Header, body)
			if err != nil {
				return Report{}, fmt.Errorf("case %s: extract fixture: %w", fixture.ID, err)
			}
			issues = rules.EvaluatePage(rules.PageInput{
				Page: page, StatusCode: response.StatusCode, Headers: response.Header,
				Depth: fixture.Depth, Inlinks: fixture.Inlinks, InSitemap: fixture.InSitemap,
				RobotsBlocked: fixture.RobotsBlocked, ExactDuplicateCount: fixture.ExactDuplicateCount,
			}, rules.DefaultThresholds())
		} else {
			issues = rules.EvaluateResource(rules.ResourceInput{URL: documentURL, StatusCode: response.StatusCode, ContentType: contentType, DecodedBytes: int64(len(body)), Body: body})
		}
		for _, issue := range issues {
			observed[fixture.ID] = append(observed[fixture.ID], ObservedFinding{URL: documentURL, Source: SourceRaw, Classification: Classify(issue), Issue: issue})
		}
		statuses[fixture.ID] = response.StatusCode
		urls[fixture.ID] = documentURL
	}
	return Compare(manifest, observed, statuses, urls), nil
}

func (r Report) JSON() ([]byte, error) {
	return json.MarshalIndent(r, "", "  ")
}

func (r Report) Markdown() string {
	var output strings.Builder
	fmt.Fprintf(&output, "# Conformance report: %s\n\n", r.ManifestID)
	fmt.Fprintf(&output, "- Result: **%s**\n", map[bool]string{true: "PASS", false: "FAIL"}[r.Passed])
	fmt.Fprintf(&output, "- Precision: %.4f\n- Recall: %.4f\n", r.Precision, r.Recall)
	fmt.Fprintf(&output, "- True positives: %d\n- False positives: %d\n- False negatives: %d\n\n", r.TruePositive, r.FalsePositive, r.FalseNegative)
	output.WriteString("| Case | HTTP | Matched | Unexpected | Missing | Result |\n")
	output.WriteString("|---|---:|---:|---:|---:|---|\n")
	for _, item := range r.Cases {
		result := "PASS"
		if !item.Passed {
			result = "FAIL"
		}
		fmt.Fprintf(&output, "| %s | %d/%d | %d | %d | %d | %s |\n", item.CaseID, item.ObservedStatusCode, item.ExpectedStatusCode, len(item.Matched), len(item.Unexpected), len(item.Missing), result)
	}
	return output.String()
}
