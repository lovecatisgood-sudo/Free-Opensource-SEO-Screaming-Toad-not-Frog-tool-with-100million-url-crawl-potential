package fetchpolicy

import "testing"

func TestNormalizeURL(t *testing.T) {
	t.Parallel()

	got, err := NormalizeURL("HTTPS://Example.COM:443/a/../b/?z=2&a=2&a=1#fragment")
	if err != nil {
		t.Fatalf("NormalizeURL: %v", err)
	}
	want := "https://example.com/b/?a=1&a=2&z=2"
	if got.RequestKey != want {
		t.Fatalf("RequestKey = %q, want %q", got.RequestKey, want)
	}
}

func TestNormalizeURLRejectsUnsafeForms(t *testing.T) {
	t.Parallel()

	for _, raw := range []string{"file:///etc/passwd", "https://user:pass@example.com/", "https:///missing"} {
		if _, err := NormalizeURL(raw); err == nil {
			t.Errorf("NormalizeURL(%q) unexpectedly succeeded", raw)
		}
	}
}

func TestNormalizeURLUsesASCIIDomainIdentity(t *testing.T) {
	t.Parallel()

	got, err := NormalizeURL("https://bücher.example/")
	if err != nil {
		t.Fatalf("NormalizeURL: %v", err)
	}
	if got.RequestKey != "https://xn--bcher-kva.example/" {
		t.Fatalf("unexpected request key: %q", got.RequestKey)
	}
}
