package api

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/seo-auditor/seo-auditor/internal/application"
	"github.com/seo-auditor/seo-auditor/internal/contracts"
	"github.com/seo-auditor/seo-auditor/internal/database"
	"github.com/seo-auditor/seo-auditor/internal/version"
)

const sessionCookie = "seo_auditor_session"

type Backend interface {
	StartCrawl(context.Context, application.CrawlRequest) (application.CrawlResult, error)
	Progress(context.Context, contracts.ID) (contracts.CrawlProgress, error)
	Summary(context.Context, contracts.ID) (database.AuditSummary, error)
	ListPages(context.Context, contracts.ID, contracts.PageRequest) (contracts.Page[database.PageRecord], error)
	ListIssues(context.Context, contracts.ID, contracts.PageRequest) (contracts.Page[database.IssueRecord], error)
	ListLinks(context.Context, contracts.ID, contracts.PageRequest) (contracts.Page[database.LinkRecord], error)
	ListEvents(context.Context, contracts.ID, contracts.PageRequest) (contracts.Page[database.CrawlEventRecord], error)
	GetPage(context.Context, contracts.ID, int64) (database.PageDetail, error)
	CompareCrawls(context.Context, contracts.ID, contracts.ID) (database.CrawlComparison, error)
	Export(context.Context, application.ExportRequest) (application.Artifact, error)
	Artifact(context.Context, contracts.ID) (application.Artifact, error)
	Backup(context.Context, contracts.ID) (application.Artifact, error)
	Diagnostic(context.Context, contracts.ID) (application.Artifact, error)
	TrashCrawl(context.Context, contracts.ID) error
	RestoreCrawl(context.Context, contracts.ID) error
	CreateProject(context.Context, string) (database.ProjectRecord, error)
	ListProjects(context.Context, contracts.PageRequest) (contracts.Page[database.ProjectRecord], error)
	RenameProject(context.Context, contracts.ID, string) error
	ArchiveProject(context.Context, contracts.ID, bool) error
	TrashProject(context.Context, contracts.ID) error
	RestoreProject(context.Context, contracts.ID) error
	CreateProfile(context.Context, contracts.ID, string, contracts.CrawlConfiguration) (database.ProfileRecord, error)
	ListProfiles(context.Context, contracts.ID, contracts.PageRequest) (contracts.Page[database.ProfileRecord], error)
	PreviewScope(context.Context, contracts.CrawlConfiguration, []string) ([]application.ScopeDecision, error)
	StartProfileCrawl(context.Context, contracts.ID, contracts.ID) (application.CrawlResult, error)
	ListCrawls(context.Context, contracts.ID, contracts.PageRequest) (contracts.Page[contracts.CrawlProgress], error)
	ExplainIssue(context.Context, contracts.ID, int64) (application.IssueExplanation, error)
	Cancel(context.Context, contracts.ID) error
	Pause(context.Context, contracts.ID) error
	Resume(context.Context, contracts.ID) error
}

type Handler struct {
	backend                         Backend
	sessionToken, csrfToken, origin string
}

func New(backend Backend, origin string) (*Handler, error) {
	if backend == nil || origin == "" {
		return nil, errors.New("API backend and origin are required")
	}
	session, err := randomToken()
	if err != nil {
		return nil, err
	}
	csrf, err := randomToken()
	if err != nil {
		return nil, err
	}
	return &Handler{backend: backend, sessionToken: session, csrfToken: csrf, origin: strings.TrimSuffix(origin, "/")}, nil
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")
	if r.URL.Path == "/api/v1/health" && r.Method == http.MethodGet {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ready", "version": version.Version})
		return
	}
	if r.URL.Path == "/api/v1/openapi.json" && r.Method == http.MethodGet {
		serveOpenAPI(w)
		return
	}
	if r.URL.Path == "/api/v1/session" && r.Method == http.MethodPost {
		h.bootstrap(w, r)
		return
	}
	if !h.authenticated(r) {
		writeError(w, http.StatusUnauthorized, "unauthorized", "local session is required")
		return
	}
	if r.URL.Path == "/api/v1/projects" {
		if r.Method == http.MethodGet {
			page, err := pageRequest(r)
			if err != nil {
				writeError(w, http.StatusBadRequest, "invalid_argument", err.Error())
				return
			}
			value, err := h.backend.ListProjects(r.Context(), page)
			respond(w, value, err)
			return
		}
		if r.Method == http.MethodPost {
			if !h.mutationAllowed(r) {
				writeError(w, http.StatusForbidden, "forbidden", "origin or CSRF validation failed")
				return
			}
			var input struct {
				Name string `json:"name"`
			}
			if err := decodeBody(w, r, 16<<10, &input); err != nil {
				writeError(w, http.StatusBadRequest, "invalid_argument", "invalid project request")
				return
			}
			value, err := h.backend.CreateProject(r.Context(), input.Name)
			if err != nil {
				writeError(w, http.StatusBadRequest, "invalid_argument", err.Error())
				return
			}
			writeJSON(w, http.StatusCreated, value)
			return
		}
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		return
	}
	projectParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(projectParts) >= 4 && projectParts[0] == "api" && projectParts[1] == "v1" && projectParts[2] == "projects" {
		if h.serveProjectRoute(w, r, projectParts) {
			return
		}
	}
	if r.URL.Path == "/api/v1/crawls" && r.Method == http.MethodPost {
		if !h.mutationAllowed(r) {
			writeError(w, http.StatusForbidden, "forbidden", "origin or CSRF validation failed")
			return
		}
		var input struct {
			URL                 string   `json:"url"`
			URLs                []string `json:"urls"`
			Name                string   `json:"name"`
			AllowSubdomains     bool     `json:"allow_subdomains"`
			MaximumURLs         int64    `json:"maximum_urls"`
			MaximumDepth        *int     `json:"maximum_depth"`
			Concurrency         *int     `json:"concurrency"`
			PerHost             *int     `json:"per_host"`
			ResponseCompression string   `json:"response_compression"`
		}
		r.Body = http.MaxBytesReader(w, r.Body, 2<<20)
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if err := decodeOne(decoder, &input); err != nil {
			writeError(w, http.StatusBadRequest, "invalid_argument", "invalid crawl request")
			return
		}
		limits := contracts.DefaultCrawlLimits()
		if input.MaximumURLs != 0 {
			limits.MaximumURLs = input.MaximumURLs
		}
		if input.MaximumDepth != nil {
			limits.MaximumDepth = *input.MaximumDepth
		}
		if input.Concurrency != nil {
			limits.GlobalConcurrency = *input.Concurrency
		}
		if input.PerHost != nil {
			limits.PerHostConcurrency = *input.PerHost
		}
		limits.MaximumDuration = 24 * time.Hour
		if input.URL == "" && len(input.URLs) == 0 {
			writeError(w, http.StatusBadRequest, "invalid_argument", "url or urls is required")
			return
		}
		result, err := h.backend.StartCrawl(r.Context(), application.CrawlRequest{ProjectName: input.Name, SeedURL: input.URL, SeedURLs: input.URLs, AllowSubdomains: input.AllowSubdomains, ResponseCompression: input.ResponseCompression, Limits: limits})
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid_argument", err.Error())
			return
		}
		writeJSON(w, http.StatusAccepted, result)
		return
	}
	if r.URL.Path == "/api/v1/comparisons" && r.Method == http.MethodPost {
		if !h.mutationAllowed(r) {
			writeError(w, http.StatusForbidden, "forbidden", "origin or CSRF validation failed")
			return
		}
		var input struct {
			BaseCrawlID   contracts.ID `json:"base_crawl_id"`
			TargetCrawlID contracts.ID `json:"target_crawl_id"`
		}
		r.Body = http.MaxBytesReader(w, r.Body, 16<<10)
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if err := decodeOne(decoder, &input); err != nil || input.BaseCrawlID == "" || input.TargetCrawlID == "" {
			writeError(w, http.StatusBadRequest, "invalid_argument", "two crawl IDs are required")
			return
		}
		value, err := h.backend.CompareCrawls(r.Context(), input.BaseCrawlID, input.TargetCrawlID)
		respond(w, value, err)
		return
	}
	if r.URL.Path == "/api/v1/exports" && r.Method == http.MethodPost {
		if !h.mutationAllowed(r) {
			writeError(w, http.StatusForbidden, "forbidden", "origin or CSRF validation failed")
			return
		}
		var input application.ExportRequest
		r.Body = http.MaxBytesReader(w, r.Body, 16<<10)
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if err := decodeOne(decoder, &input); err != nil {
			writeError(w, http.StatusBadRequest, "invalid_argument", "invalid export request")
			return
		}
		value, err := h.backend.Export(r.Context(), input)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid_argument", err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, value)
		return
	}
	if strings.HasPrefix(r.URL.Path, "/api/v1/artifacts/") && r.Method == http.MethodGet {
		artifactParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(artifactParts) != 4 && len(artifactParts) != 5 {
			writeError(w, http.StatusNotFound, "not_found", "route not found")
			return
		}
		value, err := h.backend.Artifact(r.Context(), contracts.ID(artifactParts[3]))
		if err != nil {
			respond(w, nil, err)
			return
		}
		if len(artifactParts) == 5 && artifactParts[4] == "download" {
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename=%q`, value.RelativePath))
			http.ServeFile(w, r, value.Path)
			return
		}
		writeJSON(w, http.StatusOK, value)
		return
	}
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) == 6 && parts[0] == "api" && parts[1] == "v1" && parts[2] == "crawls" && parts[4] == "pages" && r.Method == http.MethodGet {
		pageID, err := strconv.ParseInt(parts[5], 10, 64)
		if err != nil || pageID < 1 {
			writeError(w, http.StatusBadRequest, "invalid_argument", "page ID is invalid")
			return
		}
		value, err := h.backend.GetPage(r.Context(), contracts.ID(parts[3]), pageID)
		respond(w, value, err)
		return
	}
	if len(parts) == 6 && parts[0] == "api" && parts[1] == "v1" && parts[2] == "crawls" && parts[4] == "issues" && r.Method == http.MethodGet {
		issueID, err := strconv.ParseInt(parts[5], 10, 64)
		if err != nil || issueID < 1 {
			writeError(w, http.StatusBadRequest, "invalid_argument", "issue ID is invalid")
			return
		}
		value, err := h.backend.ExplainIssue(r.Context(), contracts.ID(parts[3]), issueID)
		respond(w, value, err)
		return
	}
	if len(parts) != 5 || parts[0] != "api" || parts[1] != "v1" || parts[2] != "crawls" {
		writeError(w, http.StatusNotFound, "not_found", "route not found")
		return
	}
	crawlID := contracts.ID(parts[3])
	action := parts[4]
	if r.Method == http.MethodGet && action == "events" {
		h.serveEvents(w, r, crawlID)
		return
	}
	if r.Method == http.MethodPost {
		if !h.mutationAllowed(r) {
			writeError(w, http.StatusForbidden, "forbidden", "origin or CSRF validation failed")
			return
		}
		var mutation func(context.Context, contracts.ID) error
		switch action {
		case "cancel":
			mutation = h.backend.Cancel
		case "pause":
			mutation = h.backend.Pause
		case "resume":
			mutation = h.backend.Resume
		case "trash":
			mutation = h.backend.TrashCrawl
		case "restore":
			mutation = h.backend.RestoreCrawl
		case "backup":
			value, err := h.backend.Backup(r.Context(), crawlID)
			if err != nil {
				writeError(w, http.StatusConflict, "conflict", err.Error())
				return
			}
			writeJSON(w, http.StatusCreated, value)
			return
		case "diagnostics":
			value, err := h.backend.Diagnostic(r.Context(), crawlID)
			if err != nil {
				respond(w, nil, err)
				return
			}
			writeJSON(w, http.StatusCreated, value)
			return
		}
		if mutation != nil {
			if err := mutation(r.Context(), crawlID); err != nil {
				writeError(w, http.StatusConflict, "conflict", err.Error())
				return
			}
			w.WriteHeader(http.StatusNoContent)
			return
		}
	}
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		return
	}
	page, err := pageRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_argument", err.Error())
		return
	}
	switch action {
	case "status":
		value, err := h.backend.Progress(r.Context(), crawlID)
		respond(w, value, err)
	case "summary":
		value, err := h.backend.Summary(r.Context(), crawlID)
		respond(w, value, err)
	case "pages":
		value, err := h.backend.ListPages(r.Context(), crawlID, page)
		respond(w, value, err)
	case "issues":
		value, err := h.backend.ListIssues(r.Context(), crawlID, page)
		respond(w, value, err)
	case "links":
		value, err := h.backend.ListLinks(r.Context(), crawlID, page)
		respond(w, value, err)
	case "timeline":
		value, err := h.backend.ListEvents(r.Context(), crawlID, page)
		respond(w, value, err)
	default:
		writeError(w, http.StatusNotFound, "not_found", "route not found")
	}
}

func (h *Handler) serveProjectRoute(w http.ResponseWriter, r *http.Request, parts []string) bool {
	projectID := contracts.ID(parts[3])
	if len(parts) == 4 && (r.Method == http.MethodPatch || r.Method == http.MethodDelete) {
		if !h.mutationAllowed(r) {
			writeError(w, http.StatusForbidden, "forbidden", "origin or CSRF validation failed")
			return true
		}
		if r.Method == http.MethodDelete {
			if err := h.backend.TrashProject(r.Context(), projectID); err != nil {
				writeError(w, http.StatusConflict, "conflict", err.Error())
				return true
			}
			w.WriteHeader(http.StatusNoContent)
			return true
		}
		var input struct {
			Name     *string `json:"name"`
			Archived *bool   `json:"archived"`
		}
		if err := decodeBody(w, r, 16<<10, &input); err != nil {
			writeError(w, http.StatusBadRequest, "invalid_argument", "invalid project update")
			return true
		}
		if input.Name == nil && input.Archived == nil {
			writeError(w, http.StatusBadRequest, "invalid_argument", "no project change supplied")
			return true
		}
		if input.Name != nil {
			if err := h.backend.RenameProject(r.Context(), projectID, *input.Name); err != nil {
				writeError(w, http.StatusBadRequest, "invalid_argument", err.Error())
				return true
			}
		}
		if input.Archived != nil {
			if err := h.backend.ArchiveProject(r.Context(), projectID, *input.Archived); err != nil {
				writeError(w, http.StatusBadRequest, "invalid_argument", err.Error())
				return true
			}
		}
		w.WriteHeader(http.StatusNoContent)
		return true
	}
	if len(parts) == 5 && parts[4] == "restore" && r.Method == http.MethodPost {
		if !h.mutationAllowed(r) {
			writeError(w, http.StatusForbidden, "forbidden", "origin or CSRF validation failed")
			return true
		}
		if err := h.backend.RestoreProject(r.Context(), projectID); err != nil {
			writeError(w, http.StatusConflict, "conflict", err.Error())
			return true
		}
		w.WriteHeader(http.StatusNoContent)
		return true
	}
	if len(parts) == 5 && parts[4] == "profiles" {
		if r.Method == http.MethodGet {
			page, err := pageRequest(r)
			if err != nil {
				writeError(w, http.StatusBadRequest, "invalid_argument", err.Error())
				return true
			}
			value, err := h.backend.ListProfiles(r.Context(), projectID, page)
			respond(w, value, err)
			return true
		}
		if r.Method == http.MethodPost {
			if !h.mutationAllowed(r) {
				writeError(w, http.StatusForbidden, "forbidden", "origin or CSRF validation failed")
				return true
			}
			var input struct {
				Name          string                       `json:"name"`
				Configuration contracts.CrawlConfiguration `json:"configuration"`
			}
			if err := decodeBody(w, r, 128<<10, &input); err != nil {
				writeError(w, http.StatusBadRequest, "invalid_argument", "invalid profile request")
				return true
			}
			if input.Configuration.Limits.MaximumURLs == 0 {
				input.Configuration.Limits = contracts.DefaultCrawlLimits()
			}
			value, err := h.backend.CreateProfile(r.Context(), projectID, input.Name, input.Configuration)
			if err != nil {
				writeError(w, http.StatusBadRequest, "invalid_argument", err.Error())
				return true
			}
			writeJSON(w, http.StatusCreated, value)
			return true
		}
	}
	if len(parts) == 5 && parts[4] == "scope-preview" && r.Method == http.MethodPost {
		if !h.mutationAllowed(r) {
			writeError(w, http.StatusForbidden, "forbidden", "origin or CSRF validation failed")
			return true
		}
		var input struct {
			Configuration contracts.CrawlConfiguration `json:"configuration"`
			URLs          []string                     `json:"urls"`
		}
		if err := decodeBody(w, r, 128<<10, &input); err != nil {
			writeError(w, http.StatusBadRequest, "invalid_argument", "invalid scope preview")
			return true
		}
		if input.Configuration.Limits.MaximumURLs == 0 {
			input.Configuration.Limits = contracts.DefaultCrawlLimits()
		}
		value, err := h.backend.PreviewScope(r.Context(), input.Configuration, input.URLs)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid_argument", err.Error())
			return true
		}
		writeJSON(w, http.StatusOK, map[string]any{"decisions": value})
		return true
	}
	if len(parts) == 5 && parts[4] == "crawls" && r.Method == http.MethodGet {
		page, err := pageRequest(r)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid_argument", err.Error())
			return true
		}
		value, err := h.backend.ListCrawls(r.Context(), projectID, page)
		respond(w, value, err)
		return true
	}
	if len(parts) == 5 && parts[4] == "crawls" && r.Method == http.MethodPost {
		if !h.mutationAllowed(r) {
			writeError(w, http.StatusForbidden, "forbidden", "origin or CSRF validation failed")
			return true
		}
		var input struct {
			ProfileID contracts.ID `json:"profile_id"`
		}
		if err := decodeBody(w, r, 16<<10, &input); err != nil || input.ProfileID == "" {
			writeError(w, http.StatusBadRequest, "invalid_argument", "profile ID is required")
			return true
		}
		value, err := h.backend.StartProfileCrawl(r.Context(), projectID, input.ProfileID)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid_argument", err.Error())
			return true
		}
		writeJSON(w, http.StatusAccepted, value)
		return true
	}
	return false
}

func decodeBody(w http.ResponseWriter, r *http.Request, maximum int64, value any) error {
	r.Body = http.MaxBytesReader(w, r.Body, maximum)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decodeOne(decoder, value)
}

func (h *Handler) serveEvents(w http.ResponseWriter, r *http.Request, crawlID contracts.ID) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		writeError(w, http.StatusInternalServerError, "unavailable", "streaming is unavailable")
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	lastID := r.Header.Get("Last-Event-ID")
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		progress, err := h.backend.Progress(r.Context(), crawlID)
		if err != nil {
			_, _ = fmt.Fprintf(w, "event: error\ndata: {\"error\":\"status unavailable\"}\n\n")
			flusher.Flush()
			return
		}
		eventID := strconv.FormatInt(progress.UpdatedAt.UnixNano(), 10)
		if eventID != lastID {
			body, _ := json.Marshal(progress)
			_ = http.NewResponseController(w).SetWriteDeadline(time.Now().Add(15 * time.Second))
			_, _ = fmt.Fprintf(w, "id: %s\nevent: progress\ndata: %s\n\n", eventID, body)
			flusher.Flush()
			lastID = eventID
		}
		if progress.Status.Terminal() {
			return
		}
		select {
		case <-r.Context().Done():
			return
		case <-ticker.C:
		}
	}
}

func decodeOne(decoder *json.Decoder, value any) error {
	if err := decoder.Decode(value); err != nil {
		return err
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		if err == nil {
			return errors.New("request contains more than one JSON value")
		}
		return err
	}
	return nil
}

func (h *Handler) bootstrap(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Origin") != "" && r.Header.Get("Origin") != h.origin {
		writeError(w, http.StatusForbidden, "forbidden", "origin is not allowed")
		return
	}
	http.SetCookie(w, &http.Cookie{Name: sessionCookie, Value: h.sessionToken, Path: "/api/", HttpOnly: true, SameSite: http.SameSiteStrictMode, Secure: false, MaxAge: 8 * 60 * 60})
	writeJSON(w, http.StatusOK, map[string]string{"csrf_token": h.csrfToken})
}
func (h *Handler) authenticated(r *http.Request) bool {
	cookie, err := r.Cookie(sessionCookie)
	return err == nil && constantEqual(cookie.Value, h.sessionToken)
}
func (h *Handler) mutationAllowed(r *http.Request) bool {
	return r.Header.Get("Origin") == h.origin && constantEqual(r.Header.Get("X-CSRF-Token"), h.csrfToken)
}
func constantEqual(a, b string) bool {
	return len(a) == len(b) && subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}
func randomToken() (string, error) {
	value := make([]byte, 32)
	if _, err := rand.Read(value); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(value), nil
}
func pageRequest(r *http.Request) (contracts.PageRequest, error) {
	result := contracts.PageRequest{Cursor: r.URL.Query().Get("cursor"), Search: r.URL.Query().Get("search"), Sort: r.URL.Query().Get("sort"), Severity: r.URL.Query().Get("severity"), RuleID: r.URL.Query().Get("rule_id")}
	if len(result.Search) > 500 || len(result.RuleID) > 50 {
		return result, errors.New("filter is too long")
	}
	if result.Sort != "" && result.Sort != "id" {
		return result, errors.New("sort must be id")
	}
	if raw := r.URL.Query().Get("limit"); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value < 1 || value > contracts.MaximumPageSize {
			return result, fmt.Errorf("limit must be between 1 and %d", contracts.MaximumPageSize)
		}
		result.Limit = value
	}
	return result, nil
}
func respond(w http.ResponseWriter, value any, err error) {
	if err != nil {
		if errors.Is(err, context.Canceled) {
			writeError(w, 499, "cancelled", "request was cancelled")
			return
		}
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, string(contracts.CodeNotFound), "resource was not found")
			return
		}
		var appError *contracts.AppError
		if errors.As(err, &appError) {
			status := map[contracts.ErrorCode]int{
				contracts.CodeInvalidArgument: http.StatusBadRequest,
				contracts.CodeNotFound:        http.StatusNotFound,
				contracts.CodeConflict:        http.StatusConflict,
				contracts.CodeLimitReached:    http.StatusTooManyRequests,
				contracts.CodeTargetBlocked:   http.StatusForbidden,
				contracts.CodeUnavailable:     http.StatusServiceUnavailable,
			}[appError.Code]
			if status == 0 {
				status = http.StatusInternalServerError
			}
			writeError(w, status, string(appError.Code), appError.Message)
			return
		}
		writeError(w, http.StatusInternalServerError, "internal", "request failed")
		return
	}
	writeJSON(w, http.StatusOK, value)
}
func writeJSON(w http.ResponseWriter, status int, value any) {
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]any{"error": map[string]string{"code": code, "message": message}})
}
