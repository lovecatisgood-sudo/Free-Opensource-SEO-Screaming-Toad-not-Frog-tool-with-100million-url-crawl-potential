package fetchpolicy

import (
	"compress/gzip"
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"net/url"
	"strings"
	"testing"
	"time"
)

func fixtureFetcher(t *testing.T, handler http.Handler, limits FetchLimits) (*Fetcher, string) {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	serverURL, err := url.Parse(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	fixtureAddress := serverURL.Host
	resolver := staticResolver{netip.MustParseAddr("93.184.216.34")}
	scope, err := CompileScope(ScopeConfig{AllowedHosts: []string{"fixture.example"}})
	if err != nil {
		t.Fatal(err)
	}
	guard := &Guard{Resolver: resolver, Scope: scope}
	transport := NewHTTPTransportWithDial(resolver, func(ctx context.Context, network, approved string) (net.Conn, error) {
		if !strings.HasPrefix(approved, "93.184.216.34:") {
			t.Fatalf("dial received unapproved address %q", approved)
		}
		return (&net.Dialer{}).DialContext(ctx, network, fixtureAddress)
	})
	fetcher, err := NewFetcher(guard, transport, limits, "SEOAuditor-Test/1.0")
	if err != nil {
		t.Fatal(err)
	}
	return fetcher, "http://fixture.example:" + serverURL.Port()
}

func TestFetcherFollowsRevalidatedRelativeRedirect(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	mux.HandleFunc("/start", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Location", "/final")
		w.WriteHeader(http.StatusFound)
	})
	mux.HandleFunc("/final", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-Agent") != "SEOAuditor-Test/1.0" {
			t.Errorf("unexpected user agent: %q", r.Header.Get("User-Agent"))
		}
		_, _ = w.Write([]byte("ok"))
	})
	fetcher, base := fixtureFetcher(t, mux, DefaultFetchLimits())
	result, err := fetcher.Fetch(context.Background(), base+"/start")
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	if string(result.Body) != "ok" || len(result.Redirects) != 1 || result.StatusCode != http.StatusOK {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestFetcherCanOmitCompressionNegotiation(t *testing.T) {
	t.Parallel()

	limits := DefaultFetchLimits()
	limits.OmitAcceptEncoding = true
	fetcher, base := fixtureFetcher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if value := r.Header.Get("Accept-Encoding"); value != "" {
			t.Errorf("Accept-Encoding=%q, want omitted", value)
		}
		_, _ = w.Write([]byte("identity response"))
	}), limits)
	result, err := fetcher.Fetch(context.Background(), base)
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	if string(result.Body) != "identity response" {
		t.Fatalf("body=%q", result.Body)
	}
}

func TestFetcherRequestsGzipByDefault(t *testing.T) {
	t.Parallel()

	fetcher, base := fixtureFetcher(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if value := r.Header.Get("Accept-Encoding"); value != "gzip" {
			t.Errorf("Accept-Encoding=%q, want gzip", value)
		}
		_, _ = w.Write([]byte("ok"))
	}), DefaultFetchLimits())
	if _, err := fetcher.Fetch(context.Background(), base); err != nil {
		t.Fatalf("Fetch: %v", err)
	}
}

func TestFetcherBlocksRedirectToPrivateTarget(t *testing.T) {
	t.Parallel()

	fetcher, base := fixtureFetcher(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Location", "http://127.0.0.1/private")
		w.WriteHeader(http.StatusFound)
	}), DefaultFetchLimits())
	if _, err := fetcher.Fetch(context.Background(), base); err == nil || !strings.Contains(err.Error(), "outside crawl scope") {
		t.Fatalf("expected blocked redirect, got %v", err)
	}
}

func TestFetcherEnforcesDecodedLimit(t *testing.T) {
	t.Parallel()

	limits := DefaultFetchLimits()
	limits.MaximumCompressedBytes = 128
	limits.MaximumDecodedBytes = 32
	fetcher, base := fixtureFetcher(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(strings.Repeat("x", 64)))
	}), limits)
	if _, err := fetcher.Fetch(context.Background(), base); err == nil || !strings.Contains(err.Error(), "decoded response exceeds") {
		t.Fatalf("expected decoded limit error, got %v", err)
	}
}

func TestFetcherEnforcesCompressionRatio(t *testing.T) {
	t.Parallel()

	limits := DefaultFetchLimits()
	limits.MaximumCompressedBytes = 1024
	limits.MaximumDecodedBytes = 4096
	limits.MaximumCompressionRatio = 2
	fetcher, base := fixtureFetcher(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Encoding", "gzip")
		writer := gzip.NewWriter(w)
		_, _ = writer.Write([]byte(strings.Repeat("a", 2048)))
		_ = writer.Close()
	}), limits)
	if _, err := fetcher.Fetch(context.Background(), base); err == nil || !strings.Contains(err.Error(), "compression ratio") {
		t.Fatalf("expected compression ratio error, got %v", err)
	}
}

func TestFetcherEnforcesDeadline(t *testing.T) {
	t.Parallel()

	limits := DefaultFetchLimits()
	limits.TotalTimeout = 20 * time.Millisecond
	fetcher, base := fixtureFetcher(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(100 * time.Millisecond)
		fmt.Fprint(w, "late")
	}), limits)
	if _, err := fetcher.Fetch(context.Background(), base); err == nil {
		t.Fatal("expected deadline error")
	}
}

func TestRequestCredentialsAreHostBound(t *testing.T) {
	credentials := RequestCredentials{Header: "Authorization", Value: "Bearer secret", AllowedHosts: []string{"example.com"}, AllowSubdomains: true}
	if !credentials.allows("example.com") || !credentials.allows("www.example.com") {
		t.Fatal("credentials should apply to the configured host boundary")
	}
	if credentials.allows("example.com.evil.test") || credentials.allows("evil.test") {
		t.Fatal("credentials escaped the configured host boundary")
	}
}
