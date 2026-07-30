package mcpserver

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/seo-auditor/seo-auditor/internal/application"
	"github.com/seo-auditor/seo-auditor/internal/contracts"
	"github.com/seo-auditor/seo-auditor/internal/database"
	"github.com/seo-auditor/seo-auditor/internal/version"
)

type Caller interface {
	Call(context.Context, string, string, any, any) error
}
type service struct {
	caller Caller
	mu     sync.Mutex
	starts map[string]application.CrawlResult
}
type EmptyInput struct{}
type HealthOutput struct {
	Status       string `json:"status" jsonschema:"server readiness state"`
	Name         string `json:"name" jsonschema:"server implementation name"`
	Version      string `json:"version" jsonschema:"server implementation version"`
	APIConnected bool   `json:"api_connected" jsonschema:"whether the local crawler API is reachable"`
	Time         string `json:"time" jsonschema:"current server time in RFC3339 format"`
}
type ProjectCreateInput struct {
	Name string `json:"name" jsonschema:"local project name,1,200"`
}
type PageInput struct {
	Cursor string `json:"cursor,omitempty" jsonschema:"opaque pagination cursor"`
	Limit  int    `json:"limit,omitempty" jsonschema:"page size from 1 to 100"`
}
type ProjectInput struct {
	ProjectID contracts.ID `json:"project_id" jsonschema:"opaque project ID"`
}
type ProfileCreateInput struct {
	ProjectID           contracts.ID `json:"project_id" jsonschema:"opaque project ID"`
	Name                string       `json:"name" jsonschema:"profile name from 1 to 200 characters"`
	SeedURL             string       `json:"seed_url" jsonschema:"public HTTP or HTTPS seed URL"`
	MaximumURLs         int64        `json:"maximum_urls,omitempty" jsonschema:"URL ceiling from 1 to 100000000; defaults to 10000"`
	MaximumDepth        *int         `json:"maximum_depth,omitempty" jsonschema:"maximum crawl depth from 0 to 1000; defaults to 50"`
	AllowSubdomains     bool         `json:"allow_subdomains,omitempty" jsonschema:"include subdomains within the guarded crawl scope"`
	RenderingMode       string       `json:"rendering_mode,omitempty" jsonschema:"raw or rendered; defaults to raw"`
	ResponseCompression string       `json:"response_compression,omitempty" jsonschema:"gzip or disabled; defaults to gzip"`
	ExcludePathRegex    []string     `json:"exclude_path_regex,omitempty" jsonschema:"bounded regular expressions for excluded URL paths"`
}
type ProfileListInput struct {
	ProjectID contracts.ID `json:"project_id" jsonschema:"opaque project ID"`
	Cursor    string       `json:"cursor,omitempty" jsonschema:"opaque pagination cursor"`
	Limit     int          `json:"limit,omitempty" jsonschema:"page size from 1 to 100"`
}
type ScopePreviewInput struct {
	ProjectID     contracts.ID                 `json:"project_id"`
	Configuration contracts.CrawlConfiguration `json:"configuration"`
	URLs          []string                     `json:"urls,omitempty" jsonschema:"up to 100 candidate URLs"`
}
type CrawlStartInput struct {
	ProjectID      contracts.ID `json:"project_id"`
	ProfileID      contracts.ID `json:"profile_id"`
	IdempotencyKey string       `json:"idempotency_key" jsonschema:"caller-generated key from 1 to 128 characters"`
}
type CrawlInput struct {
	CrawlID contracts.ID `json:"crawl_id"`
}
type CrawlPageInput struct {
	CrawlID contracts.ID `json:"crawl_id" jsonschema:"opaque crawl ID"`
	Cursor  string       `json:"cursor,omitempty" jsonschema:"opaque pagination cursor"`
	Limit   int          `json:"limit,omitempty" jsonschema:"page size from 1 to 100"`
}
type CrawlListInput struct {
	ProjectID contracts.ID `json:"project_id"`
	Cursor    string       `json:"cursor,omitempty"`
	Limit     int          `json:"limit,omitempty"`
}
type IssueListInput struct {
	CrawlID  contracts.ID `json:"crawl_id"`
	Cursor   string       `json:"cursor,omitempty"`
	Limit    int          `json:"limit,omitempty"`
	Severity string       `json:"severity,omitempty"`
	RuleID   string       `json:"rule_id,omitempty"`
	Search   string       `json:"search,omitempty"`
}
type IssueExplainInput struct {
	CrawlID contracts.ID `json:"crawl_id"`
	IssueID int64        `json:"issue_id"`
}
type PageListInput struct {
	CrawlID contracts.ID `json:"crawl_id" jsonschema:"opaque crawl ID"`
	Cursor  string       `json:"cursor,omitempty" jsonschema:"opaque pagination cursor"`
	Limit   int          `json:"limit,omitempty" jsonschema:"page size from 1 to 100"`
	Search  string       `json:"search,omitempty" jsonschema:"optional URL or title search up to 500 characters"`
}
type PageGetInput struct {
	CrawlID contracts.ID `json:"crawl_id"`
	PageID  int64        `json:"page_id"`
}
type CompareInput struct {
	BaseCrawlID   contracts.ID `json:"base_crawl_id"`
	TargetCrawlID contracts.ID `json:"target_crawl_id"`
}
type ExportInput struct {
	CrawlID contracts.ID `json:"crawl_id"`
	Dataset string       `json:"dataset" jsonschema:"pages, issues, or workbook"`
	Format  string       `json:"format" jsonschema:"csv, ndjson, or xlsx"`
}
type ArtifactInput struct {
	ArtifactID contracts.ID `json:"artifact_id"`
}
type ControlOutput struct {
	CrawlID contracts.ID `json:"crawl_id"`
	Status  string       `json:"status"`
}
type ScopePreviewOutput struct {
	Decisions []application.ScopeDecision `json:"decisions"`
}

func New(caller Caller) *mcp.Server {
	s := &service{caller: caller, starts: make(map[string]application.CrawlResult)}
	server := mcp.NewServer(&mcp.Implementation{Name: "seo-screaming-toad", Version: version.Version}, nil)
	mcp.AddTool(server, readTool("health_get", "Check connection", "Verify MCP readiness and connectivity to the local SEO Screaming Toad API. Does not contact crawl targets."), s.health)
	mcp.AddTool(server, additiveTool("project_create", "Create project", "Create one local SEO Screaming Toad project."), s.projectCreate)
	mcp.AddTool(server, readTool("project_list", "List projects", "List local projects with bounded opaque-cursor pagination."), s.projectList)
	mcp.AddTool(server, additiveTool("profile_create", "Create crawl profile", "Create a reusable, bounded crawl profile for one project. Target safety policies remain mandatory."), s.profileCreate)
	mcp.AddTool(server, readTool("profile_list", "List crawl profiles", "List reusable crawl profiles for one project with bounded opaque-cursor pagination."), s.profileList)
	mcp.AddTool(server, readTool("crawl_preview_scope", "Preview crawl scope", "Normalize up to 100 candidate URLs and explain whether the supplied profile includes them. Does not fetch targets."), s.previewScope)
	mcp.AddTool(server, openWorldTool("crawl_start", "Start crawl", "Explicitly start a bounded crawl from a stored project/profile. Returns immediately; poll crawl_status. Requires a caller-generated idempotency key.", true), s.crawlStart)
	mcp.AddTool(server, readTool("crawl_status", "Get crawl status", "Read persisted crawl state, counters and terminal reason."), s.crawlStatus)
	mcp.AddTool(server, lifecycleTool("crawl_pause", "Pause crawl", "Request a controlled pause of a running crawl."), s.crawlPause)
	mcp.AddTool(server, openWorldTool("crawl_resume", "Resume crawl", "Resume a paused crawl using its stored bounded configuration.", false), s.crawlResume)
	mcp.AddTool(server, lifecycleTool("crawl_cancel", "Cancel crawl", "Request cancellation of one crawl."), s.crawlCancel)
	mcp.AddTool(server, readTool("crawl_list", "List crawls", "List crawl history for one project with bounded opaque-cursor pagination."), s.crawlList)
	mcp.AddTool(server, readTool("crawl_timeline", "Get crawl timeline", "List persisted crawl lifecycle events with bounded opaque-cursor pagination."), s.crawlTimeline)
	mcp.AddTool(server, readTool("audit_summary", "Get audit summary", "Read response, rendering and issue distributions for one crawl."), s.auditSummary)
	mcp.AddTool(server, readTool("issue_list", "List audit issues", "List versioned audit findings with bounded filters and opaque-cursor pagination."), s.issueList)
	mcp.AddTool(server, readTool("issue_explain", "Explain audit issue", "Return one finding, its versioned rule guidance, evidence and limitations."), s.issueExplain)
	mcp.AddTool(server, readTool("page_list", "List crawled pages", "List crawled pages with bounded search and opaque-cursor pagination."), s.pageList)
	mcp.AddTool(server, readTool("page_get", "Inspect crawled page", "Return one page's bounded raw/rendered extraction, relationships and issue evidence."), s.pageGet)
	mcp.AddTool(server, readTool("link_list", "List discovered links", "List stored crawl-graph links with bounded opaque-cursor pagination."), s.linkList)
	mcp.AddTool(server, readTool("crawl_compare", "Compare crawls", "Compare two crawls and disclose whether their configurations match."), s.crawlCompare)
	mcp.AddTool(server, additiveTool("report_export", "Export audit report", "Create CSV, NDJSON or XLSX output only inside the managed artifact directory. No output path is accepted."), s.reportExport)
	mcp.AddTool(server, additiveTool("diagnostic_create", "Create diagnostic", "Create a metadata-only support artifact for one crawl. Crawled URLs, content, headers and issue evidence are excluded."), s.diagnosticCreate)
	mcp.AddTool(server, readTool("artifact_get", "Get artifact metadata", "Return metadata and the approved managed path for one report or diagnostic artifact."), s.artifactGet)
	return server
}

func readTool(name, title, description string) *mcp.Tool {
	closed := false
	return &mcp.Tool{Name: name, Title: title, Description: description, Annotations: &mcp.ToolAnnotations{Title: title, ReadOnlyHint: true, IdempotentHint: true, OpenWorldHint: &closed}}
}
func additiveTool(name, title, description string) *mcp.Tool {
	destructive, closed := false, false
	return &mcp.Tool{Name: name, Title: title, Description: description, Annotations: &mcp.ToolAnnotations{Title: title, DestructiveHint: &destructive, OpenWorldHint: &closed}}
}
func lifecycleTool(name, title, description string) *mcp.Tool {
	destructive, closed := false, false
	return &mcp.Tool{Name: name, Title: title, Description: description, Annotations: &mcp.ToolAnnotations{Title: title, DestructiveHint: &destructive, IdempotentHint: true, OpenWorldHint: &closed}}
}
func openWorldTool(name, title, description string, idempotent bool) *mcp.Tool {
	destructive, open := false, true
	return &mcp.Tool{Name: name, Title: title, Description: description, Annotations: &mcp.ToolAnnotations{Title: title, DestructiveHint: &destructive, IdempotentHint: idempotent, OpenWorldHint: &open}}
}

func (s *service) requireCaller() error {
	if s.caller == nil {
		return errors.New("local SEO Auditor API client is not configured")
	}
	return nil
}
func (s *service) health(ctx context.Context, _ *mcp.CallToolRequest, _ EmptyInput) (*mcp.CallToolResult, HealthOutput, error) {
	output := HealthOutput{Status: "adapter_ready", Name: "seo-screaming-toad", Version: version.Version, Time: time.Now().UTC().Format(time.RFC3339)}
	if s.caller == nil {
		return nil, output, nil
	}
	var projects contracts.Page[database.ProjectRecord]
	if err := s.caller.Call(ctx, http.MethodGet, "/api/v1/projects?limit=1", nil, &projects); err != nil {
		return nil, output, fmt.Errorf("local crawler API is unavailable: %w", err)
	}
	output.Status, output.APIConnected = "ready", true
	return nil, output, nil
}
func (s *service) profileCreate(ctx context.Context, _ *mcp.CallToolRequest, input ProfileCreateInput) (*mcp.CallToolResult, database.ProfileRecord, error) {
	var output database.ProfileRecord
	limits := contracts.DefaultCrawlLimits()
	limits.MaximumURLs = 10_000
	if input.MaximumURLs != 0 {
		limits.MaximumURLs = input.MaximumURLs
	}
	if input.MaximumDepth != nil {
		limits.MaximumDepth = *input.MaximumDepth
	}
	mode := input.RenderingMode
	if mode == "" {
		mode = "raw"
	}
	compression := input.ResponseCompression
	if compression == "" {
		compression = "gzip"
	}
	configuration := contracts.CrawlConfiguration{SeedURL: input.SeedURL, AllowSubdomains: input.AllowSubdomains, ExcludePathRegex: input.ExcludePathRegex, RenderingMode: mode, ResponseCompression: compression, Limits: limits}
	err := s.caller.Call(ctx, http.MethodPost, "/api/v1/projects/"+url.PathEscape(string(input.ProjectID))+"/profiles", map[string]any{"name": input.Name, "configuration": configuration}, &output)
	return nil, output, err
}
func (s *service) profileList(ctx context.Context, _ *mcp.CallToolRequest, input ProfileListInput) (*mcp.CallToolResult, contracts.Page[database.ProfileRecord], error) {
	var output contracts.Page[database.ProfileRecord]
	if err := bounded(input.Limit); err != nil {
		return nil, output, err
	}
	err := s.caller.Call(ctx, http.MethodGet, "/api/v1/projects/"+url.PathEscape(string(input.ProjectID))+"/profiles?"+pageQuery(input.Cursor, input.Limit), nil, &output)
	return nil, output, err
}
func (s *service) projectCreate(ctx context.Context, _ *mcp.CallToolRequest, input ProjectCreateInput) (*mcp.CallToolResult, database.ProjectRecord, error) {
	var output database.ProjectRecord
	if err := s.requireCaller(); err != nil {
		return nil, output, err
	}
	err := s.caller.Call(ctx, http.MethodPost, "/api/v1/projects", input, &output)
	return nil, output, err
}
func (s *service) projectList(ctx context.Context, _ *mcp.CallToolRequest, input PageInput) (*mcp.CallToolResult, contracts.Page[database.ProjectRecord], error) {
	var output contracts.Page[database.ProjectRecord]
	if err := bounded(input.Limit); err != nil {
		return nil, output, err
	}
	err := s.caller.Call(ctx, http.MethodGet, "/api/v1/projects?"+pageQuery(input.Cursor, input.Limit), nil, &output)
	return nil, output, err
}
func (s *service) previewScope(ctx context.Context, _ *mcp.CallToolRequest, input ScopePreviewInput) (*mcp.CallToolResult, ScopePreviewOutput, error) {
	var output ScopePreviewOutput
	if len(input.URLs) > 100 {
		return nil, output, errors.New("URLs are limited to 100")
	}
	err := s.caller.Call(ctx, http.MethodPost, "/api/v1/projects/"+url.PathEscape(string(input.ProjectID))+"/scope-preview", map[string]any{"configuration": input.Configuration, "urls": input.URLs}, &output)
	return nil, output, err
}
func (s *service) crawlStart(ctx context.Context, _ *mcp.CallToolRequest, input CrawlStartInput) (*mcp.CallToolResult, application.CrawlResult, error) {
	var output application.CrawlResult
	if len(input.IdempotencyKey) < 1 || len(input.IdempotencyKey) > 128 {
		return nil, output, errors.New("idempotency_key must contain 1 to 128 characters")
	}
	s.mu.Lock()
	cacheKey := string(input.ProjectID) + "\x00" + string(input.ProfileID) + "\x00" + input.IdempotencyKey
	cached, found := s.starts[cacheKey]
	s.mu.Unlock()
	if found {
		return nil, cached, nil
	}
	err := s.caller.Call(ctx, http.MethodPost, "/api/v1/projects/"+url.PathEscape(string(input.ProjectID))+"/crawls", map[string]any{"profile_id": input.ProfileID}, &output)
	if err == nil {
		s.mu.Lock()
		s.starts[cacheKey] = output
		s.mu.Unlock()
	}
	return nil, output, err
}
func (s *service) crawlPause(ctx context.Context, _ *mcp.CallToolRequest, input CrawlInput) (*mcp.CallToolResult, ControlOutput, error) {
	output := ControlOutput{CrawlID: input.CrawlID, Status: "pause_requested"}
	err := s.caller.Call(ctx, http.MethodPost, crawlPath(input.CrawlID, "pause"), nil, nil)
	return nil, output, err
}
func (s *service) crawlResume(ctx context.Context, _ *mcp.CallToolRequest, input CrawlInput) (*mcp.CallToolResult, ControlOutput, error) {
	output := ControlOutput{CrawlID: input.CrawlID, Status: "resume_requested"}
	err := s.caller.Call(ctx, http.MethodPost, crawlPath(input.CrawlID, "resume"), nil, nil)
	return nil, output, err
}
func (s *service) crawlStatus(ctx context.Context, _ *mcp.CallToolRequest, input CrawlInput) (*mcp.CallToolResult, contracts.CrawlProgress, error) {
	var output contracts.CrawlProgress
	err := s.caller.Call(ctx, http.MethodGet, crawlPath(input.CrawlID, "status"), nil, &output)
	return nil, output, err
}
func (s *service) crawlTimeline(ctx context.Context, _ *mcp.CallToolRequest, input CrawlPageInput) (*mcp.CallToolResult, contracts.Page[database.CrawlEventRecord], error) {
	var output contracts.Page[database.CrawlEventRecord]
	if err := bounded(input.Limit); err != nil {
		return nil, output, err
	}
	err := s.caller.Call(ctx, http.MethodGet, crawlPath(input.CrawlID, "timeline")+"?"+pageQuery(input.Cursor, input.Limit), nil, &output)
	return nil, output, err
}
func (s *service) crawlCancel(ctx context.Context, _ *mcp.CallToolRequest, input CrawlInput) (*mcp.CallToolResult, ControlOutput, error) {
	output := ControlOutput{CrawlID: input.CrawlID, Status: "cancellation_requested"}
	err := s.caller.Call(ctx, http.MethodPost, crawlPath(input.CrawlID, "cancel"), nil, nil)
	return nil, output, err
}
func (s *service) pageList(ctx context.Context, _ *mcp.CallToolRequest, input PageListInput) (*mcp.CallToolResult, contracts.Page[database.PageRecord], error) {
	var output contracts.Page[database.PageRecord]
	if err := bounded(input.Limit); err != nil {
		return nil, output, err
	}
	if len(input.Search) > 500 {
		return nil, output, errors.New("search is limited to 500 characters")
	}
	query := url.Values{"cursor": {input.Cursor}, "limit": {fmt.Sprint(boundedLimit(input.Limit))}, "search": {input.Search}}
	err := s.caller.Call(ctx, http.MethodGet, crawlPath(input.CrawlID, "pages")+"?"+query.Encode(), nil, &output)
	return nil, output, err
}
func (s *service) crawlList(ctx context.Context, _ *mcp.CallToolRequest, input CrawlListInput) (*mcp.CallToolResult, contracts.Page[contracts.CrawlProgress], error) {
	var output contracts.Page[contracts.CrawlProgress]
	if err := bounded(input.Limit); err != nil {
		return nil, output, err
	}
	err := s.caller.Call(ctx, http.MethodGet, "/api/v1/projects/"+url.PathEscape(string(input.ProjectID))+"/crawls?"+pageQuery(input.Cursor, input.Limit), nil, &output)
	return nil, output, err
}
func (s *service) linkList(ctx context.Context, _ *mcp.CallToolRequest, input CrawlPageInput) (*mcp.CallToolResult, contracts.Page[database.LinkRecord], error) {
	var output contracts.Page[database.LinkRecord]
	if err := bounded(input.Limit); err != nil {
		return nil, output, err
	}
	err := s.caller.Call(ctx, http.MethodGet, crawlPath(input.CrawlID, "links")+"?"+pageQuery(input.Cursor, input.Limit), nil, &output)
	return nil, output, err
}
func (s *service) auditSummary(ctx context.Context, _ *mcp.CallToolRequest, input CrawlInput) (*mcp.CallToolResult, database.AuditSummary, error) {
	var output database.AuditSummary
	err := s.caller.Call(ctx, http.MethodGet, crawlPath(input.CrawlID, "summary"), nil, &output)
	return nil, output, err
}
func (s *service) issueList(ctx context.Context, _ *mcp.CallToolRequest, input IssueListInput) (*mcp.CallToolResult, contracts.Page[database.IssueRecord], error) {
	var output contracts.Page[database.IssueRecord]
	if err := bounded(input.Limit); err != nil {
		return nil, output, err
	}
	query := url.Values{"cursor": {input.Cursor}, "limit": {fmt.Sprint(boundedLimit(input.Limit))}, "severity": {input.Severity}, "rule_id": {input.RuleID}, "search": {input.Search}}
	err := s.caller.Call(ctx, http.MethodGet, crawlPath(input.CrawlID, "issues")+"?"+query.Encode(), nil, &output)
	return nil, output, err
}
func (s *service) issueExplain(ctx context.Context, _ *mcp.CallToolRequest, input IssueExplainInput) (*mcp.CallToolResult, application.IssueExplanation, error) {
	var output application.IssueExplanation
	if input.IssueID < 1 {
		return nil, output, errors.New("issue_id must be positive")
	}
	err := s.caller.Call(ctx, http.MethodGet, fmt.Sprintf("/api/v1/crawls/%s/issues/%d", url.PathEscape(string(input.CrawlID)), input.IssueID), nil, &output)
	return nil, output, err
}
func (s *service) pageGet(ctx context.Context, _ *mcp.CallToolRequest, input PageGetInput) (*mcp.CallToolResult, database.PageDetail, error) {
	var output database.PageDetail
	if input.PageID < 1 {
		return nil, output, errors.New("page_id must be positive")
	}
	err := s.caller.Call(ctx, http.MethodGet, fmt.Sprintf("/api/v1/crawls/%s/pages/%d", url.PathEscape(string(input.CrawlID)), input.PageID), nil, &output)
	return nil, output, err
}
func (s *service) crawlCompare(ctx context.Context, _ *mcp.CallToolRequest, input CompareInput) (*mcp.CallToolResult, database.CrawlComparison, error) {
	var output database.CrawlComparison
	err := s.caller.Call(ctx, http.MethodPost, "/api/v1/comparisons", input, &output)
	return nil, output, err
}
func (s *service) reportExport(ctx context.Context, _ *mcp.CallToolRequest, input ExportInput) (*mcp.CallToolResult, application.Artifact, error) {
	var output application.Artifact
	err := s.caller.Call(ctx, http.MethodPost, "/api/v1/exports", application.ExportRequest{CrawlID: input.CrawlID, Dataset: input.Dataset, Format: input.Format}, &output)
	return nil, output, err
}
func (s *service) diagnosticCreate(ctx context.Context, _ *mcp.CallToolRequest, input CrawlInput) (*mcp.CallToolResult, application.Artifact, error) {
	var output application.Artifact
	err := s.caller.Call(ctx, http.MethodPost, crawlPath(input.CrawlID, "diagnostics"), nil, &output)
	return nil, output, err
}
func (s *service) artifactGet(ctx context.Context, _ *mcp.CallToolRequest, input ArtifactInput) (*mcp.CallToolResult, application.Artifact, error) {
	var output application.Artifact
	err := s.caller.Call(ctx, http.MethodGet, "/api/v1/artifacts/"+url.PathEscape(string(input.ArtifactID)), nil, &output)
	return nil, output, err
}
func bounded(limit int) error {
	if limit < 0 || limit > 100 {
		return errors.New("limit must be between 1 and 100 when supplied")
	}
	return nil
}
func boundedLimit(limit int) int {
	if limit == 0 {
		return 100
	}
	return limit
}
func pageQuery(cursor string, limit int) string {
	return (url.Values{"cursor": {cursor}, "limit": {fmt.Sprint(boundedLimit(limit))}}).Encode()
}
func crawlPath(id contracts.ID, action string) string {
	return "/api/v1/crawls/" + url.PathEscape(string(id)) + "/" + action
}
