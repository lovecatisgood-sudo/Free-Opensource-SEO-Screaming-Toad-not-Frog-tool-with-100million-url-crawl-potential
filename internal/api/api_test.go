package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/seo-auditor/seo-auditor/internal/application"
	"github.com/seo-auditor/seo-auditor/internal/contracts"
	"github.com/seo-auditor/seo-auditor/internal/database"
)

type fakeBackend struct{ cancelled bool }

func (*fakeBackend) StartCrawl(context.Context, application.CrawlRequest) (application.CrawlResult, error) {
	return application.CrawlResult{CrawlID: "crawl_new"}, nil
}

func (*fakeBackend) Progress(context.Context, contracts.ID) (contracts.CrawlProgress, error) {
	return contracts.CrawlProgress{Status: contracts.CrawlRunning}, nil
}
func (*fakeBackend) Summary(context.Context, contracts.ID) (database.AuditSummary, error) {
	return database.AuditSummary{Status: contracts.CrawlRunning}, nil
}
func (*fakeBackend) ListPages(context.Context, contracts.ID, contracts.PageRequest) (contracts.Page[database.PageRecord], error) {
	return contracts.Page[database.PageRecord]{Items: []database.PageRecord{}}, nil
}
func (*fakeBackend) ListIssues(context.Context, contracts.ID, contracts.PageRequest) (contracts.Page[database.IssueRecord], error) {
	return contracts.Page[database.IssueRecord]{Items: []database.IssueRecord{}}, nil
}
func (b *fakeBackend) Cancel(context.Context, contracts.ID) error { b.cancelled = true; return nil }

func TestSessionAndMutationSecurity(t *testing.T) {
	t.Parallel()
	backend := &fakeBackend{}
	handler, err := New(backend, "http://127.0.0.1:7331")
	if err != nil {
		t.Fatal(err)
	}
	bad := httptest.NewRequest(http.MethodPost, "http://127.0.0.1:7331/api/v1/session", nil)
	bad.Header.Set("Origin", "http://evil.example")
	badResult := httptest.NewRecorder()
	handler.ServeHTTP(badResult, bad)
	if badResult.Code != http.StatusForbidden {
		t.Fatalf("bad origin=%d", badResult.Code)
	}
	bootstrap := httptest.NewRequest(http.MethodPost, "http://127.0.0.1:7331/api/v1/session", nil)
	bootstrap.Header.Set("Origin", "http://127.0.0.1:7331")
	bootstrapResult := httptest.NewRecorder()
	handler.ServeHTTP(bootstrapResult, bootstrap)
	if bootstrapResult.Code != http.StatusOK {
		t.Fatalf("bootstrap=%d", bootstrapResult.Code)
	}
	cookie := bootstrapResult.Result().Cookies()[0]
	cancel := httptest.NewRequest(http.MethodPost, "http://127.0.0.1:7331/api/v1/crawls/crawl_test/cancel", nil)
	cancel.AddCookie(cookie)
	cancel.Header.Set("Origin", "http://127.0.0.1:7331")
	cancel.Header.Set("X-CSRF-Token", handler.csrfToken)
	result := httptest.NewRecorder()
	handler.ServeHTTP(result, cancel)
	if result.Code != http.StatusNoContent || !backend.cancelled {
		t.Fatalf("cancel=%d cancelled=%v", result.Code, backend.cancelled)
	}
}

func TestPaginationIsBounded(t *testing.T) {
	t.Parallel()
	handler, _ := New(&fakeBackend{}, "http://127.0.0.1:7331")
	request := httptest.NewRequest(http.MethodGet, "http://127.0.0.1:7331/api/v1/crawls/id/pages?limit=1001", nil)
	request.AddCookie(&http.Cookie{Name: sessionCookie, Value: handler.sessionToken})
	result := httptest.NewRecorder()
	handler.ServeHTTP(result, request)
	if result.Code != http.StatusBadRequest {
		t.Fatalf("status=%d", result.Code)
	}
}
