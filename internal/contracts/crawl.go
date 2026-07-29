package contracts

import "time"

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
