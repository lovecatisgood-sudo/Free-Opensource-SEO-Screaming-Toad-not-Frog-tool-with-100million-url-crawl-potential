package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http/httptest"
	"os"

	"github.com/seo-auditor/seo-auditor/internal/conformance"
	"github.com/seo-auditor/seo-auditor/internal/rules"
	"github.com/seo-auditor/seo-auditor/internal/testfixtures/sites"
)

func main() {
	format := flag.String("format", "markdown", "report format: markdown or json")
	flag.Parse()
	manifest, err := sites.CoreRules()
	if err != nil {
		fatal(err)
	}
	handler, err := sites.Handler(manifest)
	if err != nil {
		fatal(err)
	}
	server := httptest.NewServer(handler)
	defer server.Close()
	report, err := conformance.RunHTTP(context.Background(), manifest, server.URL, server.Client())
	if err != nil {
		fatal(err)
	}
	coverage := conformance.BuildCoverage(rules.Catalog, manifest)
	switch *format {
	case "markdown":
		fmt.Print(report.Markdown())
		fmt.Print("\n")
		fmt.Print(coverage.Markdown())
	case "json":
		body, marshalErr := json.MarshalIndent(struct {
			Report   conformance.Report         `json:"report"`
			Coverage conformance.CoverageReport `json:"coverage"`
		}{Report: report, Coverage: coverage}, "", "  ")
		if marshalErr != nil {
			fatal(marshalErr)
		}
		fmt.Println(string(body))
	default:
		fatal(fmt.Errorf("unsupported format %q", *format))
	}
	if !report.Passed || !coverage.Passed() {
		os.Exit(1)
	}
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "conformance:", err)
	os.Exit(2)
}
