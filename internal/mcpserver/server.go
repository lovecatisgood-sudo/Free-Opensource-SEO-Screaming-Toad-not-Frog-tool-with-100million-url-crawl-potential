package mcpserver

import (
	"context"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/seo-auditor/seo-auditor/internal/version"
)

type HealthInput struct{}

type HealthOutput struct {
	Status  string `json:"status" jsonschema:"server readiness state"`
	Name    string `json:"name" jsonschema:"server implementation name"`
	Version string `json:"version" jsonschema:"server implementation version"`
	Time    string `json:"time" jsonschema:"current server time in RFC3339 format"`
}

func New() *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "seo-auditor",
		Version: version.Version,
	}, nil)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "health_get",
		Description: "Return SEO Auditor MCP readiness and version information. This read-only tool performs no network or filesystem operations.",
	}, health)
	return server
}

func health(context.Context, *mcp.CallToolRequest, HealthInput) (*mcp.CallToolResult, HealthOutput, error) {
	return nil, HealthOutput{
		Status:  "ready",
		Name:    "seo-auditor",
		Version: version.Version,
		Time:    time.Now().UTC().Format(time.RFC3339),
	}, nil
}
