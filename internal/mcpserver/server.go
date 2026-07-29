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
	Status  string `json:"status" jsonschema:"server readiness state"`
	Name    string `json:"name" jsonschema:"server implementation name"`
	Version string `json:"version" jsonschema:"server implementation version"`
	Time    string `json:"time" jsonschema:"current server time in RFC3339 format"`
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
	server := mcp.NewServer(&mcp.Implementation{Name: "seo-auditor", Version: version.Version}, nil)
	mcp.AddTool(server, &mcp.Tool{Name: "health_get", Description: "Return MCP adapter readiness. Read-only and does not contact crawl targets."}, s.health)
	mcp.AddTool(server, &mcp.Tool{Name: "project_create", Description: "Create one local SEO Auditor project. This is an explicit mutation."}, s.projectCreate)
	mcp.AddTool(server, &mcp.Tool{Name: "project_list", Description: "List local projects with bounded opaque-cursor pagination. Read-only."}, s.projectList)
	mcp.AddTool(server, &mcp.Tool{Name: "crawl_preview_scope", Description: "Normalize up to 100 candidate URLs and explain whether the supplied crawl profile includes them. Read-only; does not fetch targets."}, s.previewScope)
	mcp.AddTool(server, &mcp.Tool{Name: "crawl_start", Description: "Explicitly start a bounded crawl from a stored project/profile. Returns immediately with an opaque crawl ID; poll crawl_status."}, s.crawlStart)
	mcp.AddTool(server, &mcp.Tool{Name: "crawl_status", Description: "Read persisted crawl state and counters. Read-only."}, s.crawlStatus)
	mcp.AddTool(server, &mcp.Tool{Name: "crawl_cancel", Description: "Explicitly request cancellation of one crawl. This is a mutation."}, s.crawlCancel)
	mcp.AddTool(server, &mcp.Tool{Name: "crawl_list", Description: "List crawl history for one project with bounded opaque-cursor pagination. Read-only."}, s.crawlList)
	mcp.AddTool(server, &mcp.Tool{Name: "audit_summary", Description: "Read response and issue distributions for one crawl. Read-only."}, s.auditSummary)
	mcp.AddTool(server, &mcp.Tool{Name: "issue_list", Description: "List audit findings with bounded filters and opaque-cursor pagination. Read-only."}, s.issueList)
	mcp.AddTool(server, &mcp.Tool{Name: "issue_explain", Description: "Return one finding, its versioned rule guidance, evidence, and limitations. Read-only."}, s.issueExplain)
	mcp.AddTool(server, &mcp.Tool{Name: "page_get", Description: "Return one page's bounded extraction, relationship, and issue evidence. Read-only."}, s.pageGet)
	mcp.AddTool(server, &mcp.Tool{Name: "crawl_compare", Description: "Compare two crawls and disclose whether their configurations match. Read-only."}, s.crawlCompare)
	mcp.AddTool(server, &mcp.Tool{Name: "report_export", Description: "Create a report only inside SEO Auditor's managed artifact directory. No output path is accepted."}, s.reportExport)
	mcp.AddTool(server, &mcp.Tool{Name: "artifact_get", Description: "Return metadata and the approved managed path for one report artifact. Read-only."}, s.artifactGet)
	return server
}

func (s *service) requireCaller() error {
	if s.caller == nil {
		return errors.New("local SEO Auditor API client is not configured")
	}
	return nil
}
func (*service) health(context.Context, *mcp.CallToolRequest, EmptyInput) (*mcp.CallToolResult, HealthOutput, error) {
	return nil, HealthOutput{Status: "ready", Name: "seo-auditor", Version: version.Version, Time: time.Now().UTC().Format(time.RFC3339)}, nil
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
	cached, found := s.starts[input.IdempotencyKey]
	s.mu.Unlock()
	if found {
		return nil, cached, nil
	}
	err := s.caller.Call(ctx, http.MethodPost, "/api/v1/projects/"+url.PathEscape(string(input.ProjectID))+"/crawls", map[string]any{"profile_id": input.ProfileID}, &output)
	if err == nil {
		s.mu.Lock()
		s.starts[input.IdempotencyKey] = output
		s.mu.Unlock()
	}
	return nil, output, err
}
func (s *service) crawlStatus(ctx context.Context, _ *mcp.CallToolRequest, input CrawlInput) (*mcp.CallToolResult, contracts.CrawlProgress, error) {
	var output contracts.CrawlProgress
	err := s.caller.Call(ctx, http.MethodGet, crawlPath(input.CrawlID, "status"), nil, &output)
	return nil, output, err
}
func (s *service) crawlCancel(ctx context.Context, _ *mcp.CallToolRequest, input CrawlInput) (*mcp.CallToolResult, ControlOutput, error) {
	output := ControlOutput{CrawlID: input.CrawlID, Status: "cancellation_requested"}
	err := s.caller.Call(ctx, http.MethodPost, crawlPath(input.CrawlID, "cancel"), nil, nil)
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
