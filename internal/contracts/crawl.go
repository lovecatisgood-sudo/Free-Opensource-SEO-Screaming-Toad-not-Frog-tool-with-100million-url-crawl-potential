package contracts

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

type CrawlStatus string

const (
	CrawlPending    CrawlStatus = "pending"
	CrawlRunning    CrawlStatus = "running"
	CrawlPausing    CrawlStatus = "pausing"
	CrawlPaused     CrawlStatus = "paused"
	CrawlCancelling CrawlStatus = "cancelling"
	CrawlCancelled  CrawlStatus = "cancelled"
	CrawlCompleted  CrawlStatus = "completed"
	CrawlFailed     CrawlStatus = "failed"
	CrawlLimited    CrawlStatus = "limit_reached"
)

func (s CrawlStatus) Terminal() bool {
	switch s {
	case CrawlCancelled, CrawlCompleted, CrawlFailed, CrawlLimited:
		return true
	default:
		return false
	}
}

type CrawlLimits struct {
	MaximumURLs        int64         `json:"maximum_urls"`
	MaximumDepth       int           `json:"maximum_depth"`
	MaximumDuration    time.Duration `json:"maximum_duration"`
	MaximumBodyBytes   int64         `json:"maximum_body_bytes"`
	MaximumDiskBytes   int64         `json:"maximum_disk_bytes"`
	GlobalConcurrency  int           `json:"global_concurrency"`
	PerHostConcurrency int           `json:"per_host_concurrency"`
	MinimumHostDelay   time.Duration `json:"minimum_host_delay"`
}

// CrawlConfiguration is the non-secret, reusable input required to reconstruct
// a crawl safely after pause or process restart.
type CrawlConfiguration struct {
	SeedURL           string   `json:"seed_url"`
	AllowedHosts      []string `json:"allowed_hosts"`
	AllowSubdomains   bool     `json:"allow_subdomains"`
	IncludePathRegex  []string `json:"include_path_regex,omitempty"`
	ExcludePathRegex  []string `json:"exclude_path_regex,omitempty"`
	IncludeQueryRegex []string `json:"include_query_regex,omitempty"`
	ExcludeQueryRegex []string `json:"exclude_query_regex,omitempty"`
	UserAgent         string   `json:"user_agent"`
	RenderingMode     string   `json:"rendering_mode"`
	// ResponseCompression controls outbound Accept-Encoding negotiation.
	// Empty and "gzip" request gzip; "disabled" omits the header for servers
	// whose CDN rejects explicit compression negotiation.
	ResponseCompression string `json:"response_compression,omitempty"`
	// NearDuplicateDistance is the maximum SimHash Hamming distance (0-3).
	// Nil selects the default of 3; zero disables near-duplicate findings.
	NearDuplicateDistance *int                          `json:"near_duplicate_distance,omitempty"`
	SegmentSize           int64                         `json:"segment_size,omitempty"`
	RenderedEvidence      RenderedEvidenceConfiguration `json:"rendered_evidence,omitempty"`
	Authentication        AuthenticationConfiguration   `json:"authentication,omitempty"`
	Limits                CrawlLimits                   `json:"limits"`
}

type AuthenticationConfiguration struct {
	Mode                string `json:"mode,omitempty"`
	CredentialReference string `json:"credential_reference,omitempty"`
	Username            string `json:"username,omitempty"`
}

var credentialReferencePattern = regexp.MustCompile(`^secret_[a-zA-Z0-9][a-zA-Z0-9._-]{0,99}$`)

func (c AuthenticationConfiguration) Validate(renderingMode string) error {
	if c.Mode == "" || c.Mode == "none" {
		if c.CredentialReference != "" || c.Username != "" {
			return fmt.Errorf("authentication fields require an authentication mode")
		}
		return nil
	}
	if c.Mode != "bearer" && c.Mode != "basic" && c.Mode != "cookie" {
		return fmt.Errorf("authentication mode must be bearer, basic, or cookie")
	}
	if !credentialReferencePattern.MatchString(c.CredentialReference) {
		return fmt.Errorf("authentication credential reference is invalid")
	}
	if len(c.Username) > 256 || strings.ContainsAny(c.Username, "\r\n:") {
		return fmt.Errorf("authentication username is invalid")
	}
	if c.Mode == "basic" && c.Username == "" {
		return fmt.Errorf("basic authentication requires a username")
	}
	if c.Mode != "basic" && c.Username != "" {
		return fmt.Errorf("username is only supported for basic authentication")
	}
	if renderingMode == "rendered" {
		return fmt.Errorf("authenticated crawling currently supports raw mode only")
	}
	return nil
}

type RenderedEvidenceConfiguration struct {
	RetainDOM         bool  `json:"retain_dom,omitempty"`
	CaptureScreenshot bool  `json:"capture_screenshot,omitempty"`
	RunAccessibility  bool  `json:"run_accessibility,omitempty"`
	MaximumPageBytes  int64 `json:"maximum_page_bytes,omitempty"`
	MaximumCrawlBytes int64 `json:"maximum_crawl_bytes,omitempty"`
	RetentionDays     int   `json:"retention_days,omitempty"`
}

func (c RenderedEvidenceConfiguration) EffectiveMaximumPageBytes() int64 {
	if c.MaximumPageBytes == 0 {
		return 8 << 20
	}
	return c.MaximumPageBytes
}
func (c RenderedEvidenceConfiguration) EffectiveMaximumCrawlBytes() int64 {
	if c.MaximumCrawlBytes == 0 {
		return 1 << 30
	}
	return c.MaximumCrawlBytes
}
func (c RenderedEvidenceConfiguration) EffectiveRetentionDays() int {
	if c.RetentionDays == 0 {
		return 7
	}
	return c.RetentionDays
}

func (c RenderedEvidenceConfiguration) Validate(renderingMode string) error {
	if (c.RetainDOM || c.CaptureScreenshot || c.RunAccessibility) && renderingMode != "rendered" {
		return fmt.Errorf("rendered evidence requires rendered mode")
	}
	if value := c.EffectiveMaximumPageBytes(); value < 1<<20 || value > 32<<20 {
		return fmt.Errorf("maximum rendered artifact bytes per page is outside supported range")
	}
	if value := c.EffectiveMaximumCrawlBytes(); value < 1<<20 || value > 100<<30 {
		return fmt.Errorf("maximum rendered artifact bytes per crawl is outside supported range")
	}
	if value := c.EffectiveRetentionDays(); value < 1 || value > 30 {
		return fmt.Errorf("rendered artifact retention must be between 1 and 30 days")
	}
	return nil
}

func (c CrawlConfiguration) EffectiveResponseCompression() string {
	if c.ResponseCompression == "" {
		return "gzip"
	}
	return c.ResponseCompression
}

func (c CrawlConfiguration) EffectiveNearDuplicateDistance() int {
	if c.NearDuplicateDistance == nil {
		return 3
	}
	return *c.NearDuplicateDistance
}

func (c CrawlConfiguration) EffectiveSegmentSize() int64 {
	if c.SegmentSize == 0 {
		return 100_000
	}
	return c.SegmentSize
}

func DefaultCrawlLimits() CrawlLimits {
	return CrawlLimits{
		MaximumURLs: 100_000, MaximumDepth: 50, MaximumDuration: 24 * time.Hour,
		MaximumBodyBytes: 25 << 20, MaximumDiskBytes: 20 << 30, GlobalConcurrency: 16, PerHostConcurrency: 2,
		MinimumHostDelay: 100 * time.Millisecond,
	}
}

func (l CrawlLimits) Validate() error {
	if l.MaximumURLs < 1 || l.MaximumURLs > 100_000_000 {
		return fmt.Errorf("maximum URLs must be between 1 and 100000000")
	}
	if l.MaximumDepth < 0 || l.MaximumDepth > 1000 {
		return fmt.Errorf("maximum depth must be between 0 and 1000")
	}
	if l.MaximumDuration <= 0 || l.MaximumDuration > 30*24*time.Hour {
		return fmt.Errorf("maximum duration is outside supported range")
	}
	if l.MaximumBodyBytes < 1024 || l.MaximumBodyBytes > 256<<20 {
		return fmt.Errorf("maximum body bytes is outside supported range")
	}
	if l.MaximumDiskBytes < 1<<20 || l.MaximumDiskBytes > 10<<40 {
		return fmt.Errorf("maximum disk bytes is outside supported range")
	}
	if l.GlobalConcurrency < 1 || l.GlobalConcurrency > 512 {
		return fmt.Errorf("global concurrency is outside supported range")
	}
	if l.PerHostConcurrency < 1 || l.PerHostConcurrency > l.GlobalConcurrency {
		return fmt.Errorf("per-host concurrency is outside supported range")
	}
	if l.MinimumHostDelay < 0 || l.MinimumHostDelay > time.Minute {
		return fmt.Errorf("minimum host delay is outside supported range")
	}
	return nil
}

type CrawlProgress struct {
	CrawlID        ID          `json:"crawl_id"`
	Status         CrawlStatus `json:"status"`
	Discovered     int64       `json:"discovered"`
	Queued         int64       `json:"queued"`
	Fetched        int64       `json:"fetched"`
	Analysed       int64       `json:"analysed"`
	Failed         int64       `json:"failed"`
	StartedAt      *time.Time  `json:"started_at,omitempty"`
	UpdatedAt      time.Time   `json:"updated_at"`
	TerminalReason string      `json:"terminal_reason,omitempty"`
}
