package reports

import (
	"archive/zip"
	"bytes"
	"context"
	"io"
	"strings"
	"testing"

	"github.com/seo-auditor/seo-auditor/internal/contracts"
	"github.com/seo-auditor/seo-auditor/internal/database"
)

func TestSpreadsheetSafeEscapesFormulaPrefixes(t *testing.T) {
	t.Parallel()
	for _, value := range []string{"=cmd()", "+1", "-2", "@SUM(A1)", "  =hidden"} {
		if got := spreadsheetSafe(value); got[0] != '\'' {
			t.Errorf("spreadsheetSafe(%q) = %q", value, got)
		}
	}
	if got := spreadsheetSafe("ordinary"); got != "ordinary" {
		t.Fatalf("ordinary = %q", got)
	}
}

type exportFixture struct{}

func (exportFixture) ListPages(context.Context, contracts.ID, contracts.PageRequest) (contracts.Page[database.PageRecord], error) {
	return contracts.Page[database.PageRecord]{Items: []database.PageRecord{{ID: 1, URL: "https://example.com/", Title: "=unsafe", StatusCode: 200}}}, nil
}
func (exportFixture) ListIssues(context.Context, contracts.ID, contracts.PageRequest) (contracts.Page[database.IssueRecord], error) {
	return contracts.Page[database.IssueRecord]{Items: []database.IssueRecord{{ID: 1, RuleID: "AUD-01", Severity: "error", EvidenceJSON: `{"status_code":500}`}}}, nil
}

func TestWorkbookIsValidZipWithSafeInlineCells(t *testing.T) {
	t.Parallel()
	var output bytes.Buffer
	if err := WorkbookXLSX(context.Background(), exportFixture{}, "crawl_test", &output); err != nil {
		t.Fatal(err)
	}
	archive, err := zip.NewReader(bytes.NewReader(output.Bytes()), int64(output.Len()))
	if err != nil {
		t.Fatal(err)
	}
	foundPages := false
	for _, file := range archive.File {
		if file.Name != "xl/worksheets/sheet1.xml" {
			continue
		}
		foundPages = true
		reader, err := file.Open()
		if err != nil {
			t.Fatal(err)
		}
		body, err := io.ReadAll(reader)
		_ = reader.Close()
		if err != nil {
			t.Fatal(err)
		}
		text := string(body)
		if !strings.Contains(text, "&#39;=unsafe") || strings.Contains(text, `<f>`) {
			t.Fatalf("unsafe worksheet: %s", text)
		}
	}
	if !foundPages {
		t.Fatal("pages worksheet missing")
	}
}
