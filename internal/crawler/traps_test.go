package crawler

import (
	"strings"
	"testing"

	"github.com/seo-auditor/seo-auditor/internal/fetchpolicy"
)

func TestDetectTrap(t *testing.T) {
	t.Parallel()

	tests := []struct{ raw, reason string }{
		{"https://example.com/normal/path?a=1", ""},
		{"https://example.com/" + strings.Repeat("same/", 10), "repeated_path_segment"},
		{"https://example.com/?" + strings.Repeat("a=1&", 51), "excessive_query_parameters"},
	}
	for _, test := range tests {
		target, err := fetchpolicy.NormalizeURL(test.raw)
		if err != nil {
			t.Fatal(err)
		}
		if got := DetectTrap(target); got != test.reason {
			t.Errorf("DetectTrap(%q) = %q, want %q", test.raw, got, test.reason)
		}
	}
}
