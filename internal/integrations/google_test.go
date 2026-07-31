package integrations

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/seo-auditor/seo-auditor/internal/secretstore"
)

type failingTransport struct{}

func (failingTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	return nil, errors.New(request.URL.String())
}

func TestPageSpeedSeparatesLabEvidenceAndNeverReturnsKey(t *testing.T) {
	secrets := secretstore.NewMemory()
	_ = secrets.Put(context.Background(), "secret_psi", []byte("private-key"))
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("key") != "private-key" {
			t.Error("key missing")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"https://example.com/","lighthouseResult":{"lighthouseVersion":"13.4.1","fetchTime":"2026-07-30T00:00:00Z","finalDisplayedUrl":"https://example.com/","categories":{"performance":{"score":0.91}},"audits":{"largest-contentful-paint":{"numericValue":1234}}}}`))
	}))
	defer server.Close()
	client := Client{HTTP: server.Client(), Secrets: secrets, Endpoints: Endpoints{PageSpeed: server.URL, CrUX: server.URL, SearchConsole: server.URL + "/", GA4: server.URL + "/"}, Now: func() time.Time { return time.Date(2026, 7, 30, 0, 0, 0, 0, time.UTC) }}
	result, err := client.PageSpeed(context.Background(), "https://example.com/", "mobile", "secret_psi")
	if err != nil {
		t.Fatal(err)
	}
	if result.EvidenceSource != "lab" || result.Data.Scores["performance"] != 0.91 || result.Data.Metrics["largest-contentful-paint"] != 1234 {
		t.Fatalf("result=%+v", result)
	}
}
func TestCrUXUsesFieldEvidenceAndBoundedRequest(t *testing.T) {
	secrets := secretstore.NewMemory()
	_ = secrets.Put(context.Background(), "secret_crux", []byte("key"))
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"record":{"key":{"url":"https://example.com/"},"collectionPeriod":{"firstDate":{"year":2026}},"metrics":{"largest_contentful_paint":{"percentiles":{"p75":2000}}}}}`))
	}))
	defer server.Close()
	client := Client{HTTP: server.Client(), Secrets: secrets, Endpoints: Endpoints{CrUX: server.URL, PageSpeed: server.URL, SearchConsole: server.URL + "/", GA4: server.URL + "/"}}
	result, err := client.CrUX(context.Background(), CrUXRequest{URL: "https://example.com/"}, "secret_crux")
	if err != nil {
		t.Fatal(err)
	}
	if result.EvidenceSource != "field" || len(result.Data.Metrics) == 0 {
		t.Fatalf("result=%+v", result)
	}
}

func TestTransportErrorDoesNotExposePageSpeedKey(t *testing.T) {
	secrets := secretstore.NewMemory()
	_ = secrets.Put(context.Background(), "secret_psi", []byte("private-key"))
	client := Client{HTTP: &http.Client{Transport: failingTransport{}}, Secrets: secrets}
	_, err := client.PageSpeed(context.Background(), "https://example.com/", "mobile", "secret_psi")
	if err == nil || strings.Contains(err.Error(), "private-key") || strings.Contains(err.Error(), "example.com") {
		t.Fatalf("unsafe transport error: %v", err)
	}
}

func TestExpiredOAuthCredentialRefreshesAndRotatesInSecretStore(t *testing.T) {
	secrets := secretstore.NewMemory()
	_ = secrets.Put(context.Background(), "secret_oauth", []byte(`{"access_token":"expired","refresh_token":"refresh","client_id":"client","client_secret":"secret","expiry":"2026-07-29T00:00:00Z"}`))
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil || r.Form.Get("refresh_token") != "refresh" || r.Form.Get("client_secret") != "secret" {
			t.Errorf("unexpected OAuth refresh form: %v", r.Form)
		}
		_, _ = w.Write([]byte(`{"access_token":"rotated","expires_in":3600}`))
	}))
	defer server.Close()
	now := time.Date(2026, 7, 30, 0, 0, 0, 0, time.UTC)
	client := Client{HTTP: server.Client(), Secrets: secrets, Endpoints: Endpoints{OAuthToken: server.URL}, Now: func() time.Time { return now }}.prepare()
	token, err := client.bearerToken(context.Background(), "secret_oauth")
	if err != nil || string(token) != "rotated" {
		t.Fatalf("token=%q err=%v", token, err)
	}
	stored, _ := secrets.Get(context.Background(), "secret_oauth")
	if strings.Contains(string(stored), `"access_token":"expired"`) || !strings.Contains(string(stored), `"access_token":"rotated"`) {
		t.Fatalf("credential was not rotated: %s", stored)
	}
}

func TestSearchConsoleRejectsUnsupportedDimensionsAndDataStateBeforeSecretLookup(t *testing.T) {
	client := Client{Secrets: secretstore.NewMemory()}
	base := SearchConsoleRequest{SiteURL: "sc-domain:example.com", StartDate: "2026-07-01", EndDate: "2026-07-30", RowLimit: 10}
	invalidDimension := base
	invalidDimension.Dimensions = []string{"not-a-dimension"}
	if _, err := client.SearchConsole(context.Background(), invalidDimension, "secret_missing"); err == nil || !strings.Contains(err.Error(), "dimension") {
		t.Fatalf("unexpected dimension validation error: %v", err)
	}
	invalidState := base
	invalidState.DataState = "draft"
	if _, err := client.SearchConsole(context.Background(), invalidState, "secret_missing"); err == nil || !strings.Contains(err.Error(), "data state") {
		t.Fatalf("unexpected state validation error: %v", err)
	}
}
