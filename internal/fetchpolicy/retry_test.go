package fetchpolicy

import (
	"net/http"
	"testing"
	"time"
)

func TestClassifyRetryHonorsBoundedRetryAfter(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 7, 30, 12, 0, 0, 0, time.UTC)
	for _, test := range []struct {
		value string
		want  time.Duration
	}{
		{"120", 2 * time.Minute},
		{"999999", maximumRetryAfter},
		{now.Add(30 * time.Second).Format(http.TimeFormat), 30 * time.Second},
		{"invalid", 0},
	} {
		header := make(http.Header)
		header.Set("Retry-After", test.value)
		got := ClassifyRetry(http.StatusTooManyRequests, header, nil, now)
		if !got.Retry || got.After != test.want {
			t.Fatalf("Retry-After %q = %+v, want %s", test.value, got, test.want)
		}
	}
}

func TestClassifyRetryRejectsPermanentStatuses(t *testing.T) {
	t.Parallel()

	if got := ClassifyRetry(http.StatusNotFound, nil, nil, time.Now()); got.Retry {
		t.Fatalf("404 was classified retryable: %+v", got)
	}
}
