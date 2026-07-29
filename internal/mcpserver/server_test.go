package mcpserver

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/seo-auditor/seo-auditor/internal/application"
	"github.com/seo-auditor/seo-auditor/internal/contracts"
	"github.com/seo-auditor/seo-auditor/internal/database"
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
	if !ok || !strings.Contains(text.Text, `"status":"ready"`) {
		t.Fatalf("unexpected content: %#v", result.Content[0])
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
	case *ScopePreviewOutput:
		*value = ScopePreviewOutput{Decisions: []application.ScopeDecision{{URL: "https://example.com/", Allowed: true}}}
	case *contracts.CrawlProgress:
		*value = contracts.CrawlProgress{CrawlID: "crawl_mcp", Status: contracts.CrawlCompleted, Analysed: 1}
	case *database.AuditSummary:
		*value = database.AuditSummary{CrawlID: "crawl_mcp", Status: contracts.CrawlCompleted, Analysed: 1, IssuesBySeverity: map[string]int64{}, ResponsesByClass: map[string]int64{}}
	case *contracts.Page[database.IssueRecord]:
		*value = contracts.Page[database.IssueRecord]{Items: []database.IssueRecord{{ID: 1, RuleID: "AUD-02", EvidenceJSON: `{}`}}}
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
	steps := []*mcp.CallToolParams{{Name: "crawl_preview_scope", Arguments: map[string]any{"project_id": "project_mcp", "configuration": configuration, "urls": []string{"https://example.com/"}}}, {Name: "crawl_start", Arguments: map[string]any{"project_id": "project_mcp", "profile_id": "profile_mcp", "idempotency_key": "scenario"}}, {Name: "crawl_status", Arguments: map[string]any{"crawl_id": "crawl_mcp"}}, {Name: "audit_summary", Arguments: map[string]any{"crawl_id": "crawl_mcp"}}, {Name: "issue_list", Arguments: map[string]any{"crawl_id": "crawl_mcp", "limit": 10}}, {Name: "issue_explain", Arguments: map[string]any{"crawl_id": "crawl_mcp", "issue_id": 1}}, {Name: "report_export", Arguments: map[string]any{"crawl_id": "crawl_mcp", "dataset": "workbook", "format": "xlsx"}}, {Name: "artifact_get", Arguments: map[string]any{"artifact_id": "artifact_mcp"}}}
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
	for _, name := range []string{"project_create", "project_list", "crawl_preview_scope", "crawl_start", "crawl_status", "crawl_cancel", "crawl_list", "audit_summary", "issue_list", "issue_explain", "page_get", "crawl_compare", "report_export", "artifact_get"} {
		if !names[name] {
			t.Errorf("missing tool %s", name)
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
