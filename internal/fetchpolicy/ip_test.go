package fetchpolicy

import (
	"net/netip"
	"testing"
)

func TestClassifyIP(t *testing.T) {
	t.Parallel()

	tests := []struct {
		ip      string
		allowed bool
	}{
		{"8.8.8.8", true},
		{"2606:4700:4700::1111", true},
		{"127.0.0.1", false},
		{"10.0.0.1", false},
		{"169.254.169.254", false},
		{"100.64.0.1", false},
		{"192.0.2.1", false},
		{"::1", false},
		{"::ffff:127.0.0.1", false},
		{"fc00::1", false},
		{"fe80::1", false},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.ip, func(t *testing.T) {
			t.Parallel()
			got := ClassifyIP(netip.MustParseAddr(tt.ip))
			if got.Allowed != tt.allowed {
				t.Fatalf("ClassifyIP(%s) = %+v", tt.ip, got)
			}
		})
	}
}
