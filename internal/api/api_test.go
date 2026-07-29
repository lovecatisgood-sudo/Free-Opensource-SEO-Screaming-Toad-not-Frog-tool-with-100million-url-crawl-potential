package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/seo-auditor/seo-auditor/internal/application"
	"github.com/seo-auditor/seo-auditor/internal/contracts"
	"github.com/seo-auditor/seo-auditor/internal/database"
)

func TestOpenAPIDocumentCoversPublicRoutes(t *testing.T) {
	t.Parallel()
	handler, _ := New(&fakeBackend{}, "http://127.0.0.1:7331")
	request := httptest.NewRequest(http.MethodGet, "http://127.0.0.1:7331/api/v1/openapi.json", nil)
	result := httptest.NewRecorder()
	handler.ServeHTTP(result, request)
	if result.Code != http.StatusOK {
		t.Fatalf("status=%d", result.Code)
	}
	var document struct {
		OpenAPI string                     `json:"openapi"`
		Paths   map[string]json.RawMessage `json:"paths"`
	}
	if err := json.Unmarshal(result.Body.Bytes(), &document); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"/openapi.json", "/projects/{projectId}/restore", "/crawls/{crawlId}/timeline", "/crawls/{crawlId}/issues/{issueId}", "/crawls/{crawlId}/diagnostics"} {
		if _, exists := document.Paths[path]; !exists {
			t.Errorf("OpenAPI path missing: %s", path)
		}
	}
	if document.OpenAPI != "3.1.0" {
		t.Fatalf("OpenAPI version=%q", document.OpenAPI)
	}
}

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
func (*fakeBackend) ListLinks(context.Context, contracts.ID, contracts.PageRequest) (contracts.Page[database.LinkRecord], error) {
	return contracts.Page[database.LinkRecord]{Items: []database.LinkRecord{}}, nil
}
func (*fakeBackend) ListEvents(context.Context, contracts.ID, contracts.PageRequest) (contracts.Page[database.CrawlEventRecord], error) {
	return contracts.Page[database.CrawlEventRecord]{Items: []database.CrawlEventRecord{}}, nil
}
func (*fakeBackend) GetPage(context.Context, contracts.ID, int64) (database.PageDetail, error) {
	return database.PageDetail{}, nil
}
func (*fakeBackend) CompareCrawls(context.Context, contracts.ID, contracts.ID) (database.CrawlComparison, error) {
	return database.CrawlComparison{}, nil
}
func (*fakeBackend) Export(context.Context, application.ExportRequest) (application.Artifact, error) {
	return application.Artifact{}, nil
}
func (*fakeBackend) Artifact(context.Context, contracts.ID) (application.Artifact, error) {
	return application.Artifact{}, nil
}
func (*fakeBackend) Backup(context.Context, contracts.ID) (application.Artifact, error) {
	return application.Artifact{}, nil
}
func (*fakeBackend) Diagnostic(context.Context, contracts.ID) (application.Artifact, error) {
	return application.Artifact{}, nil
}
func (*fakeBackend) TrashCrawl(context.Context, contracts.ID) error   { return nil }
func (*fakeBackend) RestoreCrawl(context.Context, contracts.ID) error { return nil }
func (*fakeBackend) CreateProject(context.Context, string) (database.ProjectRecord, error) {
	return database.ProjectRecord{}, nil
}
func (*fakeBackend) ListProjects(context.Context, contracts.PageRequest) (contracts.Page[database.ProjectRecord], error) {
	return contracts.Page[database.ProjectRecord]{Items: []database.ProjectRecord{}}, nil
}
func (*fakeBackend) RenameProject(context.Context, contracts.ID, string) error { return nil }
func (*fakeBackend) ArchiveProject(context.Context, contracts.ID, bool) error  { return nil }
func (*fakeBackend) TrashProject(context.Context, contracts.ID) error          { return nil }
func (*fakeBackend) RestoreProject(context.Context, contracts.ID) error        { return nil }
func (*fakeBackend) CreateProfile(context.Context, contracts.ID, string, contracts.CrawlConfiguration) (database.ProfileRecord, error) {
	return database.ProfileRecord{}, nil
}
func (*fakeBackend) ListProfiles(context.Context, contracts.ID, contracts.PageRequest) (contracts.Page[database.ProfileRecord], error) {
	return contracts.Page[database.ProfileRecord]{Items: []database.ProfileRecord{}}, nil
}
func (*fakeBackend) PreviewScope(context.Context, contracts.CrawlConfiguration, []string) ([]application.ScopeDecision, error) {
	return []application.ScopeDecision{}, nil
}
func (*fakeBackend) StartProfileCrawl(context.Context, contracts.ID, contracts.ID) (application.CrawlResult, error) {
	return application.CrawlResult{}, nil
}
func (*fakeBackend) ListCrawls(context.Context, contracts.ID, contracts.PageRequest) (contracts.Page[contracts.CrawlProgress], error) {
	return contracts.Page[contracts.CrawlProgress]{Items: []contracts.CrawlProgress{}}, nil
}
func (*fakeBackend) ExplainIssue(context.Context, contracts.ID, int64) (application.IssueExplanation, error) {
	return application.IssueExplanation{}, nil
}
func (b *fakeBackend) Cancel(context.Context, contracts.ID) error { b.cancelled = true; return nil }
func (*fakeBackend) Pause(context.Context, contracts.ID) error    { return nil }
func (*fakeBackend) Resume(context.Context, contracts.ID) error   { return nil }

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

func TestUnknownSortIsRejectedByBackendContract(t *testing.T) {
	t.Parallel()
	handler, _ := New(&fakeBackend{}, "http://127.0.0.1:7331")
	request := httptest.NewRequest(http.MethodGet, "http://127.0.0.1:7331/api/v1/crawls/id/pages?sort=arbitrary", nil)
	request.AddCookie(&http.Cookie{Name: sessionCookie, Value: handler.sessionToken})
	result := httptest.NewRecorder()
	handler.ServeHTTP(result, request)
	if result.Code != http.StatusBadRequest {
		t.Fatalf("status=%d", result.Code)
	}
}

func TestRespondMapsSafeDomainErrors(t *testing.T) {
	t.Parallel()
	for name, test := range map[string]struct {
		err    error
		status int
	}{
		"missing": {sql.ErrNoRows, http.StatusNotFound},
		"blocked": {&contracts.AppError{Code: contracts.CodeTargetBlocked, Message: "target rejected"}, http.StatusForbidden},
		"unknown": {context.DeadlineExceeded, http.StatusInternalServerError},
	} {
		t.Run(name, func(t *testing.T) {
			result := httptest.NewRecorder()
			respond(result, nil, test.err)
			if result.Code != test.status {
				t.Fatalf("status=%d want=%d body=%s", result.Code, test.status, result.Body.String())
			}
		})
	}
}
