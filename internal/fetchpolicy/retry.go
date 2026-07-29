package fetchpolicy

import (
	"errors"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const maximumRetryAfter = 15 * time.Minute

type RetryDecision struct {
	Retry  bool
	After  time.Duration
	Reason string
}

func ClassifyRetry(status int, header http.Header, requestErr error, now time.Time) RetryDecision {
	if requestErr != nil {
		var networkError net.Error
		if errors.As(requestErr, &networkError) && (networkError.Timeout() || networkError.Temporary()) {
			return RetryDecision{Retry: true, Reason: "transient_network"}
		}
		return RetryDecision{Reason: "permanent_network"}
	}
	switch status {
	case http.StatusRequestTimeout, http.StatusTooEarly, http.StatusTooManyRequests,
		http.StatusInternalServerError, http.StatusBadGateway, http.StatusServiceUnavailable,
		http.StatusGatewayTimeout:
		return RetryDecision{Retry: true, After: parseRetryAfter(header.Get("Retry-After"), now), Reason: "transient_http"}
	default:
		return RetryDecision{Reason: "status_not_retryable"}
	}
}

func parseRetryAfter(value string, now time.Time) time.Duration {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0
	}
	if seconds, err := strconv.ParseUint(value, 10, 32); err == nil {
		return min(time.Duration(seconds)*time.Second, maximumRetryAfter)
	}
	when, err := http.ParseTime(value)
	if err != nil || !when.After(now) {
		return 0
	}
	return min(when.Sub(now), maximumRetryAfter)
}
