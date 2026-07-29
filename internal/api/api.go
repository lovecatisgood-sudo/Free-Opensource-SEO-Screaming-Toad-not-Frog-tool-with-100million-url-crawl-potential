package api

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
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
	Cancel(context.Context, contracts.ID) error
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
	if r.URL.Path == "/api/v1/session" && r.Method == http.MethodPost {
		h.bootstrap(w, r)
		return
	}
	if !h.authenticated(r) {
		writeError(w, http.StatusUnauthorized, "unauthorized", "local session is required")
		return
	}
	if r.URL.Path == "/api/v1/crawls" && r.Method == http.MethodPost {
		if !h.mutationAllowed(r) {
			writeError(w, http.StatusForbidden, "forbidden", "origin or CSRF validation failed")
			return
		}
		var input struct {
			URL             string `json:"url"`
			Name            string `json:"name"`
			AllowSubdomains bool   `json:"allow_subdomains"`
			MaximumURLs     int64  `json:"maximum_urls"`
			MaximumDepth    *int   `json:"maximum_depth"`
			Concurrency     *int   `json:"concurrency"`
			PerHost         *int   `json:"per_host"`
		}
		r.Body = http.MaxBytesReader(w, r.Body, 64<<10)
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
		result, err := h.backend.StartCrawl(r.Context(), application.CrawlRequest{ProjectName: input.Name, SeedURL: input.URL, AllowSubdomains: input.AllowSubdomains, Limits: limits})
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid_argument", err.Error())
			return
		}
		writeJSON(w, http.StatusAccepted, result)
		return
	}
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) != 5 || parts[0] != "api" || parts[1] != "v1" || parts[2] != "crawls" {
		writeError(w, http.StatusNotFound, "not_found", "route not found")
		return
	}
	crawlID := contracts.ID(parts[3])
	action := parts[4]
	if r.Method == http.MethodPost {
		if !h.mutationAllowed(r) {
			writeError(w, http.StatusForbidden, "forbidden", "origin or CSRF validation failed")
			return
		}
		if action == "cancel" {
			if err := h.backend.Cancel(r.Context(), crawlID); err != nil {
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
	default:
		writeError(w, http.StatusNotFound, "not_found", "route not found")
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
	result := contracts.PageRequest{Cursor: r.URL.Query().Get("cursor")}
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
			writeError(w, 499, "cancelled", err.Error())
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
