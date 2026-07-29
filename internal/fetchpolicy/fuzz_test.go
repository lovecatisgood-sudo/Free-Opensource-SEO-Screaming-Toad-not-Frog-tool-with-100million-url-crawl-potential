package fetchpolicy

import (
	"net/netip"
	"testing"
)

func FuzzNormalizeURL(f *testing.F) {
	for _, seed := range []string{
		"https://example.com/",
		"http://example.com:80/a?b=c",
		"https://user:password@example.com/",
		"http://[::ffff:127.0.0.1]/",
		"file:///etc/passwd",
	} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, raw string) {
		got, err := NormalizeURL(raw)
		if err != nil {
			return
		}
		if got.RequestKey == "" || got.URL == nil {
			t.Fatal("successful normalization returned an empty identity")
		}
	})
}

func FuzzClassifyIP(f *testing.F) {
	for _, seed := range []string{"8.8.8.8", "127.0.0.1", "::1", "::ffff:10.0.0.1", "not-an-ip"} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, raw string) {
		addr, err := netip.ParseAddr(raw)
		if err != nil {
			return
		}
		decision := ClassifyIP(addr)
		if decision.Reason == "" {
			t.Fatal("classification reason is required")
		}
	})
}

