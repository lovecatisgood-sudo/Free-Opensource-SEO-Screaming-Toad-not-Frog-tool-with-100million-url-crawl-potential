package mcpserver

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/seo-auditor/seo-auditor/internal/api"
	"github.com/seo-auditor/seo-auditor/internal/application"
	"github.com/seo-auditor/seo-auditor/internal/contracts"
	"github.com/seo-auditor/seo-auditor/internal/database"
	"github.com/seo-auditor/seo-auditor/internal/localclient"
)

func TestHealthToolOverMCPTransport(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	clientTransport, serverTransport := mcp.NewInMemoryTransports()
	serverSession, err := New(nil).Connect(ctx, serverTransport, nil)
	if err != nil {
		t.Fatalf("server connect: %v", err)
	}
	t.Cleanup(func() { _ = serverSession.Close() })

	client := mcp.NewClient(&mcp.Implementation{Name: "seo-auditor-test", Version: "1"}, nil)
	clientSession, err := client.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatalf("client connect: %v", err)
	}
	t.Cleanup(func() { _ = clientSession.Close() })

	result, err := clientSession.CallTool(ctx, &mcp.CallToolParams{Name: "health_get", Arguments: map[string]any{}})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}
	if result.IsError || len(result.Content) != 1 {
		t.Fatalf("unexpected result: %+v", result)
	}
	text, ok := result.Content[0].(*mcp.TextContent)
	if !ok || !strings.Contains(text.Text, `"status":"adapter_ready"`) || !strings.Contains(text.Text, `"name":"seo-screaming-toad"`) {
		t.Fatalf("unexpected content: %#v", result.Content[0])
	}
}

func TestMCPConnectsThroughAuthenticatedLocalAPI(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	applicationService, err := application.Open(ctx, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer applicationService.Close()
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	origin := "http://" + listener.Addr().String()
	handler, err := api.New(applicationService, origin)
	if err != nil {
		t.Fatal(err)
	}
	httpServer := &http.Server{Handler: handler}
	go func() { _ = httpServer.Serve(listener) }()
	defer httpServer.Close()
	caller, err := localclient.New(origin)
	if err != nil {
		t.Fatal(err)
	}
	clientTransport, serverTransport := mcp.NewInMemoryTransports()
	serverSession, err := New(caller).Connect(ctx, serverTransport, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer serverSession.Close()
	clientSession, err := mcp.NewClient(&mcp.Implementation{Name: "local-api-integration", Version: "1"}, nil).Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer clientSession.Close()
	for _, call := range []*mcp.CallToolParams{
		{Name: "health_get", Arguments: map[string]any{}},
		{Name: "project_create", Arguments: map[string]any{"name": "Agent workspace"}},
	} {
		result, err := clientSession.CallTool(ctx, call)
		if err != nil || result.IsError {
			t.Fatalf("tool=%s result=%+v err=%v", call.Name, result, err)
		}
		if call.Name == "health_get" {
			text := result.Content[0].(*mcp.TextContent).Text
			if !strings.Contains(text, `"api_connected":true`) {
				t.Fatalf("health=%s", text)
			}
		}
	}
	projects, err := clientSession.CallTool(ctx, &mcp.CallToolParams{Name: "project_list", Arguments: map[string]any{"limit": 10}})
	if err != nil || projects.IsError || len(projects.Content) != 1 {
		t.Fatalf("project_list result=%+v err=%v", projects, err)
	}
	projectText := projects.Content[0].(*mcp.TextContent).Text
	var projectPage contracts.Page[database.ProjectRecord]
	if err := json.Unmarshal([]byte(projectText), &projectPage); err != nil || len(projectPage.Items) != 1 {
		t.Fatalf("project list decode: %v body=%s", err, projectText)
	}
	projectID := projectPage.Items[0].ID
	for _, call := range []*mcp.CallToolParams{
		{Name: "profile_create", Arguments: map[string]any{"project_id": projectID, "name": "Agent default", "seed_url": "https://example.com/", "maximum_urls": 100}},
		{Name: "profile_list", Arguments: map[string]any{"project_id": projectID, "limit": 10}},
	} {
		result, err := clientSession.CallTool(ctx, call)
		if err != nil || result.IsError {
			t.Fatalf("tool=%s result=%+v err=%v", call.Name, result, err)
		}
	}
}

type fakeCaller struct {
	mu    sync.Mutex
	calls int
}

func (f *fakeCaller) Call(_ context.Context, method, path string, input, output any) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls++
	switch value := output.(type) {
	case *application.CrawlResult:
		*value = application.CrawlResult{CrawlID: "crawl_mcp", ProjectID: "project_mcp"}
	case *database.ProjectRecord:
		*value = database.ProjectRecord{ID: "project_mcp", Name: "MCP"}
	case *database.ProfileRecord:
		*value = database.ProfileRecord{ID: "profile_mcp", ProjectID: "project_mcp", Name: "Default"}
	case *contracts.Page[database.ProfileRecord]:
		*value = contracts.Page[database.ProfileRecord]{Items: []database.ProfileRecord{{ID: "profile_mcp", ProjectID: "project_mcp", Name: "Default"}}}
	case *ScopePreviewOutput:
		*value = ScopePreviewOutput{Decisions: []application.ScopeDecision{{URL: "https://example.com/", Allowed: true}}}
	case *contracts.CrawlProgress:
		*value = contracts.CrawlProgress{CrawlID: "crawl_mcp", Status: contracts.CrawlCompleted, Analysed: 1}
	case *database.AuditSummary:
		*value = database.AuditSummary{CrawlID: "crawl_mcp", Status: contracts.CrawlCompleted, Analysed: 1, IssuesBySeverity: map[string]int64{}, ResponsesByClass: map[string]int64{}}
	case *contracts.Page[database.IssueRecord]:
		*value = contracts.Page[database.IssueRecord]{Items: []database.IssueRecord{{ID: 1, RuleID: "AUD-02", EvidenceJSON: `{}`}}}
	case *contracts.Page[database.PageRecord]:
		*value = contracts.Page[database.PageRecord]{Items: []database.PageRecord{{ID: 1, URL: "https://example.com/", StatusCode: 200}}}
	case *contracts.Page[database.LinkRecord]:
		*value = contracts.Page[database.LinkRecord]{Items: []database.LinkRecord{{ID: 1, SourceURL: "https://example.com/", TargetURL: "https://example.com/about"}}}
	case *contracts.Page[database.CrawlEventRecord]:
		*value = contracts.Page[database.CrawlEventRecord]{Items: []database.CrawlEventRecord{{ID: 1, CrawlID: "crawl_mcp", Event: "status_changed"}}}
	case *database.PageDetail:
		*value = database.PageDetail{Page: database.PageRecord{ID: 1, URL: "https://example.com/", StatusCode: 200}}
	case *application.IssueExplanation:
		*value = application.IssueExplanation{Issue: database.IssueRecord{ID: 1, RuleID: "AUD-02"}}
	case *application.Artifact:
		*value = application.Artifact{ArtifactRecord: database.ArtifactRecord{ID: "artifact_mcp", CrawlID: "crawl_mcp", RelativePath: "artifact_mcp.xlsx"}}
	}
	return nil
}

func TestScriptedMCPAuditScenario(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	clientTransport, serverTransport := mcp.NewInMemoryTransports()
	serverSession, err := New(&fakeCaller{}).Connect(ctx, serverTransport, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer serverSession.Close()
	clientSession, err := mcp.NewClient(&mcp.Implementation{Name: "scenario", Version: "1"}, nil).Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer clientSession.Close()
	limits := contracts.DefaultCrawlLimits()
	configuration := contracts.CrawlConfiguration{SeedURL: "https://example.com/", AllowedHosts: []string{"example.com"}, UserAgent: "SEOAuditor/test", RenderingMode: "raw", Limits: limits}
	steps := []*mcp.CallToolParams{
		{Name: "project_create", Arguments: map[string]any{"name": "MCP audit"}},
		{Name: "profile_create", Arguments: map[string]any{"project_id": "project_mcp", "name": "Default", "seed_url": "https://example.com/", "maximum_urls": 100}},
		{Name: "profile_list", Arguments: map[string]any{"project_id": "project_mcp", "limit": 10}},
		{Name: "crawl_preview_scope", Arguments: map[string]any{"project_id": "project_mcp", "configuration": configuration, "urls": []string{"https://example.com/"}}},
		{Name: "crawl_start", Arguments: map[string]any{"project_id": "project_mcp", "profile_id": "profile_mcp", "idempotency_key": "scenario"}},
		{Name: "crawl_status", Arguments: map[string]any{"crawl_id": "crawl_mcp"}},
		{Name: "crawl_pause", Arguments: map[string]any{"crawl_id": "crawl_mcp"}},
		{Name: "crawl_resume", Arguments: map[string]any{"crawl_id": "crawl_mcp"}},
		{Name: "crawl_cancel", Arguments: map[string]any{"crawl_id": "crawl_mcp"}},
		{Name: "crawl_timeline", Arguments: map[string]any{"crawl_id": "crawl_mcp", "limit": 10}},
		{Name: "audit_summary", Arguments: map[string]any{"crawl_id": "crawl_mcp"}},
		{Name: "page_list", Arguments: map[string]any{"crawl_id": "crawl_mcp", "limit": 10}},
		{Name: "page_get", Arguments: map[string]any{"crawl_id": "crawl_mcp", "page_id": 1}},
		{Name: "link_list", Arguments: map[string]any{"crawl_id": "crawl_mcp", "limit": 10}},
		{Name: "issue_list", Arguments: map[string]any{"crawl_id": "crawl_mcp", "limit": 10}},
		{Name: "issue_explain", Arguments: map[string]any{"crawl_id": "crawl_mcp", "issue_id": 1}},
		{Name: "report_export", Arguments: map[string]any{"crawl_id": "crawl_mcp", "dataset": "workbook", "format": "xlsx"}},
		{Name: "diagnostic_create", Arguments: map[string]any{"crawl_id": "crawl_mcp"}},
		{Name: "artifact_get", Arguments: map[string]any{"artifact_id": "artifact_mcp"}},
	}
	for _, step := range steps {
		result, err := clientSession.CallTool(ctx, step)
		if err != nil || result.IsError {
			t.Fatalf("tool=%s result=%+v err=%v", step.Name, result, err)
		}
	}
}

func TestRequiredToolsAndIdempotentStartOverMCP(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	caller := &fakeCaller{}
	clientTransport, serverTransport := mcp.NewInMemoryTransports()
	serverSession, err := New(caller).Connect(ctx, serverTransport, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer serverSession.Close()
	client := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "1"}, nil)
	clientSession, err := client.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer clientSession.Close()
	listed, err := clientSession.ListTools(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	names := map[string]bool{}
	for _, tool := range listed.Tools {
		names[tool.Name] = true
	}
	for _, name := range []string{"project_create", "project_list", "profile_create", "profile_list", "crawl_preview_scope", "crawl_start", "crawl_status", "crawl_pause", "crawl_resume", "crawl_cancel", "crawl_list", "crawl_timeline", "audit_summary", "issue_list", "issue_explain", "page_list", "page_get", "link_list", "architecture_get", "schedule_create", "schedule_list", "schedule_delete", "crawl_compare", "report_export", "diagnostic_create", "artifact_get", "custom_audit_put", "custom_audit_list", "custom_audit_preview", "custom_audit_results", "integration_observation_list", "pagespeed_run", "crux_run", "search_console_run", "ga4_run"} {
		if !names[name] {
			t.Errorf("missing tool %s", name)
		}
	}
	for _, tool := range listed.Tools {
		if tool.Name == "audit_summary" && (tool.Annotations == nil || !tool.Annotations.ReadOnlyHint) {
			t.Fatal("audit_summary must advertise read-only semantics")
		}
		if tool.Name == "crawl_start" && (tool.Annotations == nil || tool.Annotations.OpenWorldHint == nil || !*tool.Annotations.OpenWorldHint || !tool.Annotations.IdempotentHint) {
			t.Fatal("crawl_start must advertise open-world and idempotent semantics")
		}
		if tool.Name == "pagespeed_run" && (tool.Annotations == nil || tool.Annotations.OpenWorldHint == nil || !*tool.Annotations.OpenWorldHint || tool.Annotations.IdempotentHint) {
			t.Fatal("pagespeed_run must advertise open-world and non-idempotent semantics")
		}
	}
	unknown, err := clientSession.CallTool(ctx, &mcp.CallToolParams{Name: "crawl_status", Arguments: map[string]any{"crawl_id": "crawl_mcp", "unexpected": true}})
	if err == nil && !unknown.IsError {
		t.Fatal("unknown MCP fields must be rejected")
	}
	arguments := map[string]any{"project_id": "project_mcp", "profile_id": "profile_mcp", "idempotency_key": "same-operation"}
	for range 2 {
		result, err := clientSession.CallTool(ctx, &mcp.CallToolParams{Name: "crawl_start", Arguments: arguments})
		if err != nil || result.IsError {
			t.Fatalf("start result=%+v err=%v", result, err)
		}
	}
	caller.mu.Lock()
	calls := caller.calls
	caller.mu.Unlock()
	if calls != 1 {
		t.Fatalf("API calls=%d want=1", calls)
	}
}
