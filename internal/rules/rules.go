package rules

import (
	"bytes"
	"net/http"
	"net/url"
	"regexp"
	"slices"
	"strings"

	"github.com/seo-auditor/seo-auditor/internal/extractor"
	"github.com/seo-auditor/seo-auditor/internal/ruledata/googleprofiles"
	"github.com/seo-auditor/seo-auditor/internal/ruledata/schemaorg"
)

type Severity string

const (
	SeverityInfo    Severity = "info"
	SeverityWarning Severity = "warning"
	SeverityError   Severity = "error"
)

type Classification string

const (
	ClassificationDeterministic  Classification = "deterministic"
	ClassificationRecommendation Classification = "recommendation"
	ClassificationReview         Classification = "review"
	ClassificationInformation    Classification = "information"
)

type Metadata struct {
	ID              string   `json:"id"`
	Title           string   `json:"title"`
	Category        string   `json:"category"`
	Remediation     string   `json:"remediation"`
	Limitations     string   `json:"limitations"`
	Version         int      `json:"version"`
	DefaultSeverity Severity `json:"default_severity"`
}

type ResourceInput struct {
	URL          string
	StatusCode   int
	ContentType  string
	DecodedBytes int64
	Body         []byte
}

func EvaluateResource(input ResourceInput) []Issue {
	metadata := make(map[string]Metadata, len(Catalog))
	for _, item := range Catalog {
		metadata[item.ID] = item
	}
	var issues []Issue
	add := func(id string, severity Severity, message string, evidence map[string]any) {
		issue := Issue{RuleID: id, RuleVersion: metadata[id].Version, Severity: severity, Message: message, Evidence: evidence}
		issue.Classification = Classify(issue)
		issues = append(issues, issue)
	}
	contentType := strings.ToLower(strings.Split(input.ContentType, ";")[0])
	if input.StatusCode < 200 || input.StatusCode >= 400 {
		add("AUD-01", SeverityError, "Resource response is not successful", map[string]any{"status_code": input.StatusCode, "content_type": contentType})
	}
	if strings.HasPrefix(contentType, "image/") {
		if input.DecodedBytes > 500<<10 {
			add("AUD-16", SeverityWarning, "Image resource exceeds the review-size threshold", map[string]any{"decoded_bytes": input.DecodedBytes, "threshold_bytes": 500 << 10, "content_type": contentType})
		}
		if contentType == "image/bmp" || contentType == "image/tiff" {
			add("AUD-16", SeverityWarning, "Image uses a legacy web delivery format", map[string]any{"content_type": contentType})
		}
	}
	isPDF := contentType == "application/pdf" || bytes.HasPrefix(input.Body, []byte("%PDF-"))
	if isPDF {
		if !bytes.Contains(input.Body, []byte("/Title")) {
			add("AUD-17", SeverityWarning, "PDF title metadata was not observed", map[string]any{"check": "info_title", "decoded_bytes": input.DecodedBytes})
		}
		if !bytes.Contains(input.Body, []byte("/Lang")) {
			add("AUD-17", SeverityWarning, "PDF document language was not observed", map[string]any{"check": "document_language"})
		}
		if !bytes.Contains(input.Body, []byte("/Marked true")) && !bytes.Contains(input.Body, []byte("/Marked(true)")) {
			add("AUD-17", SeverityWarning, "Tagged PDF marker was not observed", map[string]any{"check": "marked_content"})
		}
		if bytes.Contains(input.Body, []byte("/Encrypt")) {
			add("AUD-17", SeverityInfo, "PDF encryption dictionary was observed", map[string]any{"encrypted": true})
		}
	}
	return issues
}

type Issue struct {
	RuleID         string         `json:"rule_id"`
	RuleVersion    int            `json:"rule_version"`
	Severity       Severity       `json:"severity"`
	Message        string         `json:"message"`
	Evidence       map[string]any `json:"evidence"`
	Classification Classification `json:"classification"`
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
	{"AUD-13", "Structured data and Schema.org vocabulary", "structured-data", "Repair invalid JSON-LD, use a valid Schema.org context, and replace unknown or superseded Schema.org terms.", "Vocabulary and advisory domain/range checks use the bundled Schema.org release and do not establish Google rich-result eligibility.", 3, SeverityError},
	{"AUD-14", "Google search-feature profiles", "structured-data", "Complete the documented required properties and review applicable recommended properties for the detected search feature.", "These local diagnostics reflect a pinned documentation review and do not guarantee eligibility, indexing, ranking, or rich-result display.", 1, SeverityError},
	{"AUD-15", "Mobile, AMP and alternate signals", "mobile", "Provide a responsive viewport and ensure AMP/mobile alternate links do not self-reference or conflict with canonical intent.", "These document-level checks do not emulate search-engine mobile indexing or validate reciprocal alternates across uncrawled pages.", 1, SeverityWarning},
	{"AUD-16", "Advanced image delivery", "images", "Declare image dimensions and review oversized or incorrectly served image resources.", "Declared dimensions and transfer bytes do not measure visual quality, compression efficiency, or layout behavior in every viewport.", 1, SeverityWarning},
	{"AUD-17", "PDF search readiness", "documents", "Add useful PDF metadata, language and tagged-document structure where the format and audience require it.", "PDF byte-pattern checks are bounded diagnostics, not a full PDF parser or accessibility conformance assessment.", 1, SeverityWarning},
}

var languagePattern = regexp.MustCompile(`(?i)^(x-default|[a-z]{2,3}(-[a-z]{2}|-[0-9]{3})?)$`)

func EvaluatePage(input PageInput, thresholds Thresholds) []Issue {
	metadata := make(map[string]Metadata, len(Catalog))
	for _, item := range Catalog {
		metadata[item.ID] = item
	}
	var issues []Issue
	add := func(id string, severity Severity, message string, evidence map[string]any) {
		issue := Issue{RuleID: id, RuleVersion: metadata[id].Version, Severity: severity, Message: message, Evidence: evidence}
		issue.Classification = Classify(issue)
		issues = append(issues, issue)
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
	missingDimensions := 0
	for _, image := range input.Page.Images {
		if image.Width == 0 || image.Height == 0 {
			missingDimensions++
		}
	}
	if missingDimensions > 0 {
		add("AUD-16", SeverityWarning, "Images omit declared width or height", map[string]any{"missing_dimensions": missingDimensions, "image_count": len(input.Page.Images)})
	}
	viewport := strings.ToLower(strings.ReplaceAll(input.Page.Viewport, " ", ""))
	if viewport == "" {
		add("AUD-15", SeverityWarning, "Responsive viewport declaration is absent", map[string]any{"viewport": ""})
	} else if !strings.Contains(viewport, "width=device-width") {
		add("AUD-15", SeverityWarning, "Viewport does not declare device width", map[string]any{"viewport": input.Page.Viewport})
	}
	if input.Page.AMPURL != "" && sameURL(input.Page.AMPURL, input.Page.URL) {
		add("AUD-15", SeverityError, "AMP alternate self-references the current page", map[string]any{"amp_url": input.Page.AMPURL})
	}
	if input.Page.MobileAlternate != "" && sameURL(input.Page.MobileAlternate, input.Page.URL) {
		add("AUD-15", SeverityError, "Mobile alternate self-references the current page", map[string]any{"mobile_alternate": input.Page.MobileAlternate})
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
	for position, item := range input.Page.StructuredData {
		if !item.Valid {
			add("AUD-13", SeverityError, "JSON-LD syntax is invalid", map[string]any{"format": item.Format, "block_index": position, "parser_error": item.Error})
			continue
		}
		if len(item.StructuralErrors) > 0 {
			add("AUD-13", SeverityError, "JSON-LD structure is invalid", map[string]any{"format": item.Format, "block_index": position, "errors": item.StructuralErrors, "evidence_truncated": item.EvidenceTruncated})
		}
		if item.Format == "json-ld" && hasCompactSchemaType(item.Types) && !hasSchemaContext(item.Contexts) {
			add("AUD-13", SeverityWarning, "Compact structured-data types have no observed Schema.org context", map[string]any{"format": item.Format, "block_index": position, "types": item.Types, "contexts": item.Contexts, "evidence_truncated": item.EvidenceTruncated})
		}
		registry, err := schemaorg.Default()
		if err != nil {
			continue
		}
		schemaContext := hasExplicitSchemaContext(item.Contexts)
		for _, term := range item.Types {
			if !schemaContext && !isExplicitSchemaTerm(term) {
				continue
			}
			name, definition, known := registry.Type(term)
			if !known {
				if normalized, eligible := schemaorg.NormalizeTerm(term); eligible {
					add("AUD-13", SeverityError, "Schema.org type is unknown in bundled vocabulary", schemaVocabularyEvidence(item, position, normalized))
				}
				continue
			}
			if definition.SupersededBy != "" {
				evidence := schemaVocabularyEvidence(item, position, name)
				evidence["replacement"] = definition.SupersededBy
				add("AUD-13", SeverityWarning, "Schema.org term is superseded", evidence)
			}
		}
		for _, term := range item.Properties {
			if !schemaContext && !isExplicitSchemaTerm(term) {
				continue
			}
			name, definition, known := registry.Property(term)
			if !known {
				if normalized, eligible := schemaorg.NormalizeTerm(term); eligible {
					add("AUD-13", SeverityError, "Schema.org property is unknown in bundled vocabulary", schemaVocabularyEvidence(item, position, normalized))
				}
				continue
			}
			if definition.SupersededBy != "" {
				evidence := schemaVocabularyEvidence(item, position, name)
				evidence["replacement"] = definition.SupersededBy
				add("AUD-13", SeverityWarning, "Schema.org term is superseded", evidence)
			}
		}
		for _, finding := range evaluateSchemaRelationships(registry, item, position) {
			add("AUD-13", finding.severity, finding.message, finding.evidence)
		}
		for _, finding := range evaluateGoogleProfiles(item, position) {
			add("AUD-14", finding.severity, finding.message, finding.evidence)
		}
	}
	return issues
}

func Classify(issue Issue) Classification {
	if issue.Severity == SeverityInfo {
		return ClassificationInformation
	}
	switch issue.RuleID {
	case "AUD-01", "AUD-05", "AUD-07", "AUD-09", "AUD-10":
		return ClassificationDeterministic
	case "AUD-13":
		if issue.Message == "Schema.org term is superseded" || strings.Contains(issue.Message, "Schema.org domain") || strings.Contains(issue.Message, "Schema.org range") {
			return ClassificationRecommendation
		}
		return ClassificationDeterministic
	case "AUD-14":
		if issue.Message == "Recommended Google search-feature properties are absent" {
			return ClassificationRecommendation
		}
		return ClassificationDeterministic
	case "AUD-15":
		if strings.Contains(issue.Message, "self-references") {
			return ClassificationDeterministic
		}
		return ClassificationRecommendation
	case "AUD-16", "AUD-17":
		return ClassificationRecommendation
	case "AUD-04":
		if issue.Message == "Canonical is absent" {
			return ClassificationRecommendation
		}
		return ClassificationDeterministic
	case "AUD-02":
		if issue.Message == "Title is missing or empty" || issue.Message == "Meta description is missing or empty" {
			return ClassificationDeterministic
		}
		return ClassificationRecommendation
	case "AUD-03":
		if issue.Message == "H1 is missing" || issue.Message == "Page has multiple H1 headings" {
			return ClassificationDeterministic
		}
		return ClassificationRecommendation
	case "AUD-08":
		if issue.Message == "Page has exact content duplicates" {
			return ClassificationDeterministic
		}
		return ClassificationRecommendation
	case "AUD-11":
		return ClassificationRecommendation
	case "AUD-12":
		if issue.Message == "HTTPS page references HTTP resources" {
			return ClassificationDeterministic
		}
		return ClassificationInformation
	default:
		return ClassificationReview
	}
}

func evaluateSchemaRelationships(registry *schemaorg.Registry, item extractor.StructuredData, blockIndex int) []profileFinding {
	if item.Format != "json-ld" || !item.Valid || len(item.Nodes) == 0 {
		return nil
	}
	var findings []profileFinding
	for _, node := range item.Nodes {
		if len(node.Types) == 0 {
			continue
		}
		for _, property := range node.Properties {
			name, definition, known := registry.Property(property)
			if !known {
				continue
			}
			if len(definition.Domains) > 0 && !anyTypeMatches(registry, node.Types, definition.Domains) {
				findings = append(findings, profileFinding{SeverityWarning, "Schema.org domain guidance may not match the node type", map[string]any{"block_index": blockIndex, "node_path": node.Path, "property": name, "node_types": node.Types, "expected_domains": definition.Domains, "vocabulary_version": registry.Metadata.Version}})
			}
			if len(definition.Ranges) == 0 {
				continue
			}
			prefix := node.Path + "." + property
			for _, child := range item.Nodes {
				if child.Path == prefix || strings.HasPrefix(child.Path, prefix+"[") {
					if len(child.Types) > 0 && !anyTypeMatches(registry, child.Types, definition.Ranges) {
						findings = append(findings, profileFinding{SeverityWarning, "Schema.org range guidance may not match the nested node type", map[string]any{"block_index": blockIndex, "node_path": child.Path, "property": name, "nested_types": child.Types, "expected_ranges": definition.Ranges, "vocabulary_version": registry.Metadata.Version}})
					}
				}
			}
		}
	}
	return findings
}
func anyTypeMatches(registry *schemaorg.Registry, types, expected []string) bool {
	for _, actual := range types {
		for _, wanted := range expected {
			if registry.IsA(actual, wanted) {
				return true
			}
		}
	}
	return false
}

func sameURL(first, second string) bool {
	a, errA := url.Parse(first)
	b, errB := url.Parse(second)
	if errA != nil || errB != nil {
		return false
	}
	a.Fragment, b.Fragment = "", ""
	return a.String() == b.String()
}

type profileFinding struct {
	severity Severity
	message  string
	evidence map[string]any
}

func evaluateGoogleProfiles(item extractor.StructuredData, blockIndex int) []profileFinding {
	if item.Format != "json-ld" || !item.Valid || len(item.Nodes) == 0 {
		return nil
	}
	bundle, err := googleprofiles.Default()
	if err != nil {
		return nil
	}
	var findings []profileFinding
	for _, profile := range bundle.Profiles {
		for _, node := range item.Nodes {
			if !nodeHasAnyType(node, profile.AppliesTo) {
				continue
			}
			missing := missingProperties(node.Properties, profile.Required)
			if len(missing) > 0 {
				findings = append(findings, profileFinding{SeverityError, "Required Google search-feature properties are absent", googleProfileEvidence(bundle, profile, item, node.Path, blockIndex, missing)})
			}
			missingRecommended := missingProperties(node.Properties, profile.Recommended)
			if len(missingRecommended) > 0 {
				findings = append(findings, profileFinding{SeverityWarning, "Recommended Google search-feature properties are absent", googleProfileEvidence(bundle, profile, item, node.Path, blockIndex, missingRecommended)})
			}
			for _, nested := range profile.Nested {
				// Avoid a second nested error while the parent itself is incomplete.
				if len(missing) > 0 {
					continue
				}
				count, incomplete := countNestedNodes(item.Nodes, nested)
				if count < nested.MinimumCount || len(incomplete) > 0 {
					evidence := googleProfileEvidence(bundle, profile, item, node.Path, blockIndex, incomplete)
					evidence["nested_types"] = nested.Types
					evidence["minimum_count"] = nested.MinimumCount
					evidence["observed_count"] = count
					findings = append(findings, profileFinding{SeverityError, "Nested Google search-feature items are incomplete", evidence})
				}
			}
		}
	}
	return findings
}

func nodeHasAnyType(node extractor.StructuredNode, types []string) bool {
	wanted := make(map[string]struct{}, len(types))
	for _, value := range types {
		wanted[value] = struct{}{}
	}
	for _, value := range node.Types {
		if name, ok := schemaorg.NormalizeTerm(value); ok {
			if _, exists := wanted[name]; exists {
				return true
			}
		}
	}
	return false
}

func missingProperties(observed, required []string) []string {
	set := make(map[string]struct{}, len(observed))
	for _, value := range observed {
		if name, ok := schemaorg.NormalizeTerm(value); ok {
			set[name] = struct{}{}
		}
	}
	var missing []string
	for _, value := range required {
		if _, exists := set[value]; !exists {
			missing = append(missing, value)
		}
	}
	return missing
}

func countNestedNodes(nodes []extractor.StructuredNode, requirement googleprofiles.NestedRequirement) (int, []string) {
	count := 0
	missingSet := make(map[string]struct{})
	for _, node := range nodes {
		if !nodeHasAnyType(node, requirement.Types) {
			continue
		}
		count++
		for _, value := range missingProperties(node.Properties, requirement.Required) {
			missingSet[value] = struct{}{}
		}
	}
	missing := make([]string, 0, len(missingSet))
	for value := range missingSet {
		missing = append(missing, value)
	}
	slices.Sort(missing)
	return count, missing
}

func googleProfileEvidence(bundle googleprofiles.Bundle, profile googleprofiles.Profile, item extractor.StructuredData, path string, blockIndex int, missing []string) map[string]any {
	return map[string]any{
		"profile_id":         profile.ID,
		"profile_title":      profile.Title,
		"profile_version":    bundle.Metadata.Version,
		"source_urls":        bundle.Metadata.Sources,
		"format":             item.Format,
		"block_index":        blockIndex,
		"node_path":          path,
		"missing_properties": missing,
		"evidence_truncated": item.EvidenceTruncated,
		"eligibility_notice": bundle.Metadata.Limitations,
	}
}

func hasCompactSchemaType(types []string) bool {
	for _, value := range types {
		value = strings.TrimSpace(value)
		if value != "" && !strings.Contains(value, ":") && !strings.HasPrefix(value, "http://") && !strings.HasPrefix(value, "https://") {
			return true
		}
	}
	return false
}

func hasSchemaContext(contexts []string) bool {
	for _, value := range contexts {
		normalized := strings.TrimRight(strings.ToLower(strings.TrimSpace(value)), "/")
		if normalized == "https://schema.org" || normalized == "http://schema.org" || value == "[inline context]" {
			return true
		}
	}
	return false
}

func hasExplicitSchemaContext(contexts []string) bool {
	for _, value := range contexts {
		normalized := strings.TrimRight(strings.ToLower(strings.TrimSpace(value)), "/")
		if normalized == "https://schema.org" || normalized == "http://schema.org" {
			return true
		}
	}
	return false
}

func isExplicitSchemaTerm(value string) bool {
	normalized := strings.ToLower(strings.TrimSpace(value))
	return strings.HasPrefix(normalized, "schema:") || strings.HasPrefix(normalized, "https://schema.org/") || strings.HasPrefix(normalized, "http://schema.org/")
}

func schemaVocabularyEvidence(item extractor.StructuredData, position int, term string) map[string]any {
	registry, _ := schemaorg.Default()
	version := "unknown"
	if registry != nil {
		version = registry.Metadata.Version
	}
	return map[string]any{
		"format":             item.Format,
		"block_index":        position,
		"term":               term,
		"vocabulary":         "Schema.org",
		"vocabulary_version": version,
		"evidence_truncated": item.EvidenceTruncated,
	}
}

func directiveContains(value, wanted string) bool {
	for _, part := range strings.FieldsFunc(strings.ToLower(value), func(r rune) bool { return r == ',' || r == ' ' || r == ';' }) {
		if part == wanted {
			return true
		}
	}
	return false
}
