package rules

import (
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/seo-auditor/seo-auditor/internal/extractor"
)

type Severity string

const (
	SeverityInfo    Severity = "info"
	SeverityWarning Severity = "warning"
	SeverityError   Severity = "error"
)

type Metadata struct {
	ID, Title, Category, Remediation, Limitations string
	Version                                       int
	DefaultSeverity                               Severity
}
type Issue struct {
	RuleID      string         `json:"rule_id"`
	RuleVersion int            `json:"rule_version"`
	Severity    Severity       `json:"severity"`
	Message     string         `json:"message"`
	Evidence    map[string]any `json:"evidence"`
}
type Thresholds struct{ MinimumTitle, MaximumTitle, MinimumDescription, MaximumDescription, DeepPageDepth int }

func DefaultThresholds() Thresholds { return Thresholds{30, 60, 70, 160, 4} }

type PageInput struct {
	Page                     extractor.Page
	StatusCode               int
	Headers                  http.Header
	Depth, Inlinks           int
	InSitemap, RobotsBlocked bool
	ExactDuplicateCount      int
}

var Catalog = []Metadata{
	{"AUD-01", "Response and redirects", "response", "Repair failing URLs and update internal links to stable 2xx targets.", "A failed response can be temporary and should be confirmed.", 1, SeverityError},
	{"AUD-02", "Titles and descriptions", "metadata", "Write unique, descriptive titles and meta descriptions within configured editorial thresholds.", "Length thresholds are diagnostics, not ranking guarantees.", 1, SeverityWarning},
	{"AUD-03", "Primary headings", "content", "Provide one descriptive H1 that reflects the page topic.", "Some templates intentionally use multiple semantic H1 elements.", 1, SeverityWarning},
	{"AUD-04", "Canonical signals", "indexability", "Provide one valid canonical URL aligned with the intended indexable target.", "Canonical tags are hints and require graph-level validation.", 1, SeverityWarning},
	{"AUD-05", "Indexability", "indexability", "Remove unintended blocking directives or response conditions.", "Search engines may apply additional indexing policies.", 1, SeverityError},
	{"AUD-06", "Sitemap coverage", "discovery", "Include intended indexable canonical pages in a valid sitemap.", "Sitemap inclusion is not required for indexing.", 1, SeverityInfo},
	{"AUD-07", "Robots coverage", "discovery", "Review robots rules and ensure blocked coverage is intentional.", "Blocked pages cannot be fully audited.", 1, SeverityWarning},
	{"AUD-08", "Duplicate content", "content", "Consolidate, canonicalize, or differentiate duplicate pages.", "Exact text hashes do not capture every semantic duplicate.", 1, SeverityWarning},
	{"AUD-09", "Hreflang", "international", "Use valid language/region codes and reciprocal reachable targets.", "This page-level check is supplemented by crawl graph validation.", 1, SeverityWarning},
	{"AUD-10", "Image alternatives", "images", "Add appropriate alt attributes; use empty alt only for decorative images.", "The crawler cannot determine visual intent without human review.", 1, SeverityWarning},
	{"AUD-11", "Internal architecture", "links", "Improve crawl depth and contextual internal links where appropriate.", "Low inlink counts can be intentional for utility pages.", 1, SeverityWarning},
	{"AUD-12", "Transport observations", "security", "Remove mixed content and consider appropriate defensive response headers.", "Security headers are technical observations, not ranking factors.", 1, SeverityWarning},
}

var languagePattern = regexp.MustCompile(`(?i)^(x-default|[a-z]{2,3}(-[a-z]{2}|-[0-9]{3})?)$`)

func EvaluatePage(input PageInput, thresholds Thresholds) []Issue {
	metadata := make(map[string]Metadata, len(Catalog))
	for _, item := range Catalog {
		metadata[item.ID] = item
	}
	var issues []Issue
	add := func(id string, severity Severity, message string, evidence map[string]any) {
		issues = append(issues, Issue{id, metadata[id].Version, severity, message, evidence})
	}
	if input.StatusCode < 200 || input.StatusCode >= 400 {
		add("AUD-01", SeverityError, "Page response is not successful", map[string]any{"status_code": input.StatusCode})
	}
	titleLength := len([]rune(strings.TrimSpace(input.Page.Title)))
	if titleLength == 0 {
		add("AUD-02", SeverityError, "Title is missing or empty", map[string]any{"field": "title", "length": 0})
	} else if titleLength < thresholds.MinimumTitle {
		add("AUD-02", SeverityWarning, "Title is shorter than the configured threshold", map[string]any{"field": "title", "length": titleLength, "minimum": thresholds.MinimumTitle})
	} else if titleLength > thresholds.MaximumTitle {
		add("AUD-02", SeverityWarning, "Title is longer than the configured threshold", map[string]any{"field": "title", "length": titleLength, "maximum": thresholds.MaximumTitle})
	}
	descriptionLength := len([]rune(strings.TrimSpace(input.Page.MetaDescription)))
	if descriptionLength == 0 {
		add("AUD-02", SeverityWarning, "Meta description is missing or empty", map[string]any{"field": "meta_description", "length": 0})
	} else if descriptionLength < thresholds.MinimumDescription {
		add("AUD-02", SeverityInfo, "Meta description is shorter than the configured threshold", map[string]any{"field": "meta_description", "length": descriptionLength, "minimum": thresholds.MinimumDescription})
	} else if descriptionLength > thresholds.MaximumDescription {
		add("AUD-02", SeverityWarning, "Meta description is longer than the configured threshold", map[string]any{"field": "meta_description", "length": descriptionLength, "maximum": thresholds.MaximumDescription})
	}
	var h1 []string
	for _, heading := range input.Page.Headings {
		if heading.Level == 1 {
			h1 = append(h1, heading.Text)
		}
	}
	if len(h1) == 0 {
		add("AUD-03", SeverityWarning, "H1 is missing", map[string]any{"h1_count": 0})
	} else if len(h1) > 1 {
		add("AUD-03", SeverityWarning, "Page has multiple H1 headings", map[string]any{"h1_count": len(h1)})
	}
	if len(h1) > 0 && strings.EqualFold(strings.TrimSpace(h1[0]), strings.TrimSpace(input.Page.Title)) {
		add("AUD-03", SeverityInfo, "Title and H1 are identical", map[string]any{"title": input.Page.Title, "h1": h1[0]})
	}
	if len(input.Page.Canonicals) == 0 {
		add("AUD-04", SeverityInfo, "Canonical is absent", map[string]any{"canonical_count": 0})
	} else if len(input.Page.Canonicals) > 1 {
		add("AUD-04", SeverityError, "Page has conflicting canonical elements", map[string]any{"canonicals": input.Page.Canonicals})
	} else if parsed, err := url.Parse(input.Page.Canonicals[0]); err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
		add("AUD-04", SeverityError, "Canonical target is invalid", map[string]any{"canonical": input.Page.Canonicals[0]})
	}
	directives := strings.ToLower(input.Page.MetaRobots + "," + input.Page.XRobotsTag)
	noindex := directiveContains(directives, "noindex")
	if input.StatusCode != http.StatusOK || noindex {
		add("AUD-05", SeverityError, "Page is not indexable", map[string]any{"status_code": input.StatusCode, "noindex": noindex})
	}
	if !input.InSitemap && input.StatusCode == http.StatusOK && !noindex {
		add("AUD-06", SeverityInfo, "Indexable page was not observed in a sitemap", map[string]any{"in_sitemap": false})
	}
	if input.RobotsBlocked {
		add("AUD-07", SeverityWarning, "URL is blocked by robots.txt", map[string]any{"robots_blocked": true})
	}
	if input.ExactDuplicateCount > 0 {
		add("AUD-08", SeverityWarning, "Page has exact content duplicates", map[string]any{"content_hash": input.Page.ContentHash, "duplicate_count": input.ExactDuplicateCount})
	}
	for _, item := range input.Page.Hreflangs {
		if !languagePattern.MatchString(item.Language) {
			add("AUD-09", SeverityWarning, "Hreflang language or region is invalid", map[string]any{"hreflang": item.Language, "target": item.URL})
		}
	}
	missingAlt := 0
	for _, image := range input.Page.Images {
		if !image.AltPresent {
			missingAlt++
		}
	}
	if missingAlt > 0 {
		add("AUD-10", SeverityWarning, "Images are missing alt attributes", map[string]any{"missing_alt": missingAlt, "image_count": len(input.Page.Images)})
	}
	if input.Depth > thresholds.DeepPageDepth {
		add("AUD-11", SeverityWarning, "Page is deeper than the configured threshold", map[string]any{"depth": input.Depth, "threshold": thresholds.DeepPageDepth})
	}
	if input.Depth > 0 && input.Inlinks == 0 {
		add("AUD-11", SeverityWarning, "Page has no recorded internal inlinks", map[string]any{"inlinks": 0})
	}
	nofollow := 0
	for _, link := range input.Page.Links {
		if directiveContains(link.Rel, "nofollow") {
			nofollow++
		}
	}
	if nofollow > 0 {
		add("AUD-11", SeverityInfo, "Page contains internal nofollow links", map[string]any{"nofollow_links": nofollow})
	}
	if parsed, _ := url.Parse(input.Page.URL); parsed != nil && parsed.Scheme == "https" {
		mixed := 0
		for _, link := range input.Page.Links {
			if strings.HasPrefix(strings.ToLower(link.URL), "http://") {
				mixed++
			}
		}
		for _, image := range input.Page.Images {
			if strings.HasPrefix(strings.ToLower(image.URL), "http://") {
				mixed++
			}
		}
		if mixed > 0 {
			add("AUD-12", SeverityError, "HTTPS page references HTTP resources", map[string]any{"mixed_content_references": mixed})
		}
		if input.Headers.Get("Content-Security-Policy") == "" {
			add("AUD-12", SeverityInfo, "Content-Security-Policy header was not observed", map[string]any{"header": "Content-Security-Policy"})
		}
	}
	return issues
}

func directiveContains(value, wanted string) bool {
	for _, part := range strings.FieldsFunc(strings.ToLower(value), func(r rune) bool { return r == ',' || r == ' ' || r == ';' }) {
		if part == wanted {
			return true
		}
	}
	return false
}
