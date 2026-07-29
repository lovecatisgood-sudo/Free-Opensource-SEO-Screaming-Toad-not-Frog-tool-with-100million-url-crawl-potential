package contracts

import (
	"fmt"
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
