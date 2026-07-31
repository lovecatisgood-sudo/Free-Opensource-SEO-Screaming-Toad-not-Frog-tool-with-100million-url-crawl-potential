package database

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/seo-auditor/seo-auditor/internal/contracts"
	"github.com/seo-auditor/seo-auditor/internal/customaudit"
)

func TestCustomAuditDefinitionRoundTrip(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db, err := Open(ctx, filepath.Join(t.TempDir(), "audit.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	frontier := NewFrontier(db, 8)
	defer frontier.Close()
	projectID := contracts.ID("project_custom")
	if err := frontier.CreateProject(ctx, projectID, "Custom"); err != nil {
		t.Fatal(err)
	}
	definition := customaudit.Definition{ID: "titles", Name: "Titles", Enabled: true, Mode: "raw", SelectorKind: "css", Selector: "title", Extraction: customaudit.Extraction{Kind: "text"}, Condition: customaudit.Condition{Kind: "exists"}, Limits: customaudit.DefaultLimits()}
	record, err := frontier.PutCustomAuditDefinition(ctx, projectID, definition)
	if err != nil {
		t.Fatal(err)
	}
	if record.Definition.ID != "titles" {
		t.Fatalf("record=%+v", record)
	}
	items, err := frontier.ListCustomAuditDefinitions(ctx, projectID)
	if err != nil || len(items) != 1 {
		t.Fatalf("items=%+v err=%v", items, err)
	}
	if err := frontier.DeleteCustomAuditDefinition(ctx, projectID, "titles"); err != nil {
		t.Fatal(err)
	}
}
