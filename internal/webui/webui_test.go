package webui

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandlerServesNoIndexAppWithSecurityHeaders(t *testing.T) {
	t.Parallel()
	request := httptest.NewRequest(http.MethodGet, "http://127.0.0.1/", nil)
	result := httptest.NewRecorder()
	Handler().ServeHTTP(result, request)
	if result.Code != http.StatusOK {
		t.Fatalf("status=%d", result.Code)
	}
	if !strings.Contains(result.Body.String(), "noindex") {
		t.Fatal("embedded UI must remain excluded from search indexing")
	}
	if got := result.Header().Get("Content-Security-Policy"); !strings.Contains(got, "frame-ancestors 'none'") {
		t.Fatalf("CSP=%q", got)
	}
}

func TestHandlerUsesSPAFallback(t *testing.T) {
	t.Parallel()
	request := httptest.NewRequest(http.MethodGet, "http://127.0.0.1/crawls/example", nil)
	result := httptest.NewRecorder()
	Handler().ServeHTTP(result, request)
	if result.Code != http.StatusOK || !strings.Contains(result.Body.String(), "SEO Screaming Toad") {
		t.Fatalf("status=%d body=%q", result.Code, result.Body.String())
	}
}
