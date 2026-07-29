package contracts

const (
	DefaultPageSize = 100
	MaximumPageSize = 1000
)

type PageRequest struct {
	Cursor   string `json:"cursor,omitempty"`
	Limit    int    `json:"limit,omitempty"`
	Search   string `json:"search,omitempty"`
	Sort     string `json:"sort,omitempty"`
	Severity string `json:"severity,omitempty"`
	RuleID   string `json:"rule_id,omitempty"`
}

func (p PageRequest) BoundedLimit() int {
	if p.Limit <= 0 {
		return DefaultPageSize
	}
	if p.Limit > MaximumPageSize {
		return MaximumPageSize
	}
	return p.Limit
}

type Page[T any] struct {
	Items      []T    `json:"items"`
	NextCursor string `json:"next_cursor,omitempty"`
}
