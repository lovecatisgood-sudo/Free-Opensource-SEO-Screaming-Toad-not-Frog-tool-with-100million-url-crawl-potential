package api

import (
	"encoding/json"
	"net/http"
)

var openAPIDocument = map[string]any{
	"openapi": "3.1.0",
	"info":    map[string]any{"title": "SEO Auditor local API", "version": "2.0.0", "description": "Loopback-only API. Mutations require the local session cookie, exact Origin, and X-CSRF-Token."},
	"servers": []map[string]string{{"url": "http://127.0.0.1:7331/api/v1"}},
	"paths": map[string]any{
		"/health":                             map[string]any{"get": operation("Read local service health")},
		"/openapi.json":                       map[string]any{"get": operation("Read this OpenAPI document")},
		"/session":                            map[string]any{"post": operation("Bootstrap a local browser session")},
		"/projects":                           map[string]any{"get": operation("List projects"), "post": operation("Create a local project")},
		"/projects/{projectId}":               map[string]any{"patch": operation("Rename or archive a project"), "delete": operation("Move a project to trash")},
		"/projects/{projectId}/restore":       map[string]any{"post": operation("Restore a trashed project")},
		"/projects/{projectId}/profiles":      map[string]any{"get": operation("List reusable crawl profiles"), "post": operation("Create a reusable crawl profile")},
		"/projects/{projectId}/scope-preview": map[string]any{"post": operation("Preview normalized scope decisions")},
		"/projects/{projectId}/crawls":        map[string]any{"post": operation("Start a crawl from a stored profile")},
		"/crawls":                             map[string]any{"post": operation("Start a bounded crawl")},
		"/crawls/{crawlId}/status":            map[string]any{"get": operation("Read crawl progress")},
		"/crawls/{crawlId}/events":            map[string]any{"get": operation("Stream crawl progress as server-sent events")},
		"/crawls/{crawlId}/timeline":          map[string]any{"get": operation("List persisted crawl lifecycle events")},
		"/crawls/{crawlId}/summary":           map[string]any{"get": operation("Read audit summary")},
		"/crawls/{crawlId}/pages":             map[string]any{"get": operation("List pages with bounded keyset pagination")},
		"/crawls/{crawlId}/pages/{pageId}":    map[string]any{"get": operation("Read page evidence and relationships")},
		"/crawls/{crawlId}/issues":            map[string]any{"get": operation("List issues with bounded keyset pagination")},
		"/crawls/{crawlId}/issues/{issueId}":  map[string]any{"get": operation("Explain one issue with rule guidance and evidence")},
		"/crawls/{crawlId}/links":             map[string]any{"get": operation("List links with bounded keyset pagination")},
		"/crawls/{crawlId}/pause":             map[string]any{"post": operation("Pause a running crawl")},
		"/crawls/{crawlId}/resume":            map[string]any{"post": operation("Resume a paused crawl")},
		"/crawls/{crawlId}/cancel":            map[string]any{"post": operation("Cancel a crawl")},
		"/crawls/{crawlId}/trash":             map[string]any{"post": operation("Move a terminal crawl to trash")},
		"/crawls/{crawlId}/restore":           map[string]any{"post": operation("Restore a trashed crawl")},
		"/crawls/{crawlId}/backup":            map[string]any{"post": operation("Create a verified managed SQLite backup")},
		"/crawls/{crawlId}/diagnostics":       map[string]any{"post": operation("Create a metadata-only diagnostic artifact")},
		"/comparisons":                        map[string]any{"post": operation("Compare two crawls")},
		"/exports":                            map[string]any{"post": operation("Create a managed report artifact")},
		"/artifacts/{artifactId}":             map[string]any{"get": operation("Read managed artifact metadata")},
		"/artifacts/{artifactId}/download":    map[string]any{"get": operation("Download a managed artifact")},
	},
}

func operation(summary string) map[string]any {
	return map[string]any{"summary": summary, "responses": map[string]any{"200": map[string]string{"description": "Success"}, "400": map[string]string{"description": "Invalid request"}, "401": map[string]string{"description": "Local session required"}}}
}
func serveOpenAPI(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(openAPIDocument)
}
