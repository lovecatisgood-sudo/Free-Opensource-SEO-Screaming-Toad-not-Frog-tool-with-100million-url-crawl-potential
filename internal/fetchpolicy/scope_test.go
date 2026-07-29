package fetchpolicy

import "testing"

func TestScopeHostPathAndQuery(t *testing.T) {
	t.Parallel()

	scope, err := CompileScope(ScopeConfig{
		AllowedHosts:      []string{"example.com"},
		AllowSubdomains:   true,
		IncludePathRegex:  []string{`^/(docs|products)/`},
		ExcludePathRegex:  []string{`/private/`},
		ExcludeQueryRegex: []string{`(^|&)preview=true(&|$)`},
	})
	if err != nil {
		t.Fatalf("CompileScope: %v", err)
	}
	tests := []struct {
		url     string
		allowed bool
	}{
		{"https://example.com/docs/start", true},
		{"https://www.example.com/products/widget?a=1", true},
		{"https://example.com/docs/private/a", false},
		{"https://example.com/docs/start?preview=true", false},
		{"https://other.example/docs/start", false},
		{"https://example.com/blog/post", false},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.url, func(t *testing.T) {
			t.Parallel()
			normalized, err := NormalizeURL(tt.url)
			if err != nil {
				t.Fatal(err)
			}
			if got := scope.Evaluate(normalized) == nil; got != tt.allowed {
				t.Fatalf("allowed = %v, want %v", got, tt.allowed)
			}
		})
	}
}

func TestScopeRejectsInvalidConfiguration(t *testing.T) {
	t.Parallel()

	for _, config := range []ScopeConfig{
		{},
		{AllowedHosts: []string{"example.com:443"}},
		{AllowedHosts: []string{"example.com"}, IncludePathRegex: []string{"["}},
	} {
		if _, err := CompileScope(config); err == nil {
			t.Fatalf("CompileScope(%+v) unexpectedly succeeded", config)
		}
	}
}
