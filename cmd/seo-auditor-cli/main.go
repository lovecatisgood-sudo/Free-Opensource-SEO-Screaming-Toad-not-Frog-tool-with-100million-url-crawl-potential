package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/seo-auditor/seo-auditor/internal/application"
	"github.com/seo-auditor/seo-auditor/internal/config"
	"github.com/seo-auditor/seo-auditor/internal/contracts"
	"github.com/seo-auditor/seo-auditor/internal/database"
	"github.com/seo-auditor/seo-auditor/internal/localclient"
	"github.com/seo-auditor/seo-auditor/internal/version"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	var value any
	var err error
	switch os.Args[1] {
	case "version":
		value = map[string]string{"version": version.Version, "commit": version.Commit}
	case "crawl":
		value, err = runStandaloneCrawl(ctx, os.Args[2:])
	default:
		value, err = runAPICommand(ctx, os.Args[1], os.Args[2:])
	}
	if err != nil {
		_ = json.NewEncoder(os.Stderr).Encode(map[string]string{"error": err.Error()})
		os.Exit(1)
	}
	if value != nil {
		if err := json.NewEncoder(os.Stdout).Encode(value); err != nil {
			os.Exit(1)
		}
	}
}

func apiClient() (*localclient.Client, error) {
	server, err := config.ResolveServer()
	if err != nil {
		return nil, err
	}
	return localclient.New("http://" + net.JoinHostPort(server.Host, strconv.Itoa(server.Port)))
}
func runAPICommand(ctx context.Context, command string, arguments []string) (any, error) {
	client, err := apiClient()
	if err != nil {
		return nil, err
	}
	switch command {
	case "project-create":
		flags := flag.NewFlagSet(command, flag.ContinueOnError)
		name := flags.String("name", "", "project name")
		if err := flags.Parse(arguments); err != nil {
			return nil, err
		}
		if *name == "" {
			return nil, errors.New("--name is required")
		}
		var result database.ProjectRecord
		err = client.Call(ctx, http.MethodPost, "/api/v1/projects", map[string]string{"name": *name}, &result)
		return result, err
	case "project-list":
		var result contracts.Page[database.ProjectRecord]
		err = client.Call(ctx, http.MethodGet, "/api/v1/projects?limit=100", nil, &result)
		return result, err
	case "profile-create":
		flags := flag.NewFlagSet(command, flag.ContinueOnError)
		project := flags.String("project", "", "project ID")
		name := flags.String("name", "Default", "profile name")
		seed := flags.String("url", "", "seed URL")
		maximum := flags.Int64("max-urls", 10_000, "URL ceiling")
		depth := flags.Int("max-depth", 50, "maximum depth")
		subdomains := flags.Bool("allow-subdomains", false, "include subdomains")
		rendered := flags.Bool("rendered", false, "enable rendered mode")
		responseCompression := flags.String("response-compression", "gzip", "response compression: gzip or disabled")
		exclude := flags.String("exclude-path", "", "exclude path regex")
		if err := flags.Parse(arguments); err != nil {
			return nil, err
		}
		if *project == "" || *seed == "" {
			return nil, errors.New("--project and --url are required")
		}
		limits := contracts.DefaultCrawlLimits()
		limits.MaximumURLs = *maximum
		limits.MaximumDepth = *depth
		mode := "raw"
		if *rendered {
			mode = "rendered"
		}
		configuration := contracts.CrawlConfiguration{SeedURL: *seed, AllowSubdomains: *subdomains, RenderingMode: mode, ResponseCompression: *responseCompression, Limits: limits}
		if *exclude != "" {
			configuration.ExcludePathRegex = []string{*exclude}
		}
		var result database.ProfileRecord
		err = client.Call(ctx, http.MethodPost, "/api/v1/projects/"+url.PathEscape(*project)+"/profiles", map[string]any{"name": *name, "configuration": configuration}, &result)
		return result, err
	case "profile-list":
		project, err := requiredFlag(command, arguments, "project")
		if err != nil {
			return nil, err
		}
		var result contracts.Page[database.ProfileRecord]
		err = client.Call(ctx, http.MethodGet, "/api/v1/projects/"+url.PathEscape(project)+"/profiles?limit=100", nil, &result)
		return result, err
	case "scope-preview":
		flags := flag.NewFlagSet(command, flag.ContinueOnError)
		project := flags.String("project", "", "project ID")
		seed := flags.String("url", "", "seed URL")
		candidate := flags.String("candidate", "", "candidate URL, comma-separated")
		if err := flags.Parse(arguments); err != nil {
			return nil, err
		}
		if *project == "" || *seed == "" {
			return nil, errors.New("--project and --url are required")
		}
		configuration := contracts.CrawlConfiguration{SeedURL: *seed, RenderingMode: "raw", Limits: contracts.DefaultCrawlLimits()}
		urls := []string{}
		if *candidate != "" {
			urls = strings.Split(*candidate, ",")
		}
		var result any
		err = client.Call(ctx, http.MethodPost, "/api/v1/projects/"+url.PathEscape(*project)+"/scope-preview", map[string]any{"configuration": configuration, "urls": urls}, &result)
		return result, err
	case "crawl-start":
		flags := flag.NewFlagSet(command, flag.ContinueOnError)
		project := flags.String("project", "", "project ID")
		profile := flags.String("profile", "", "profile ID")
		if err := flags.Parse(arguments); err != nil {
			return nil, err
		}
		if *project == "" || *profile == "" {
			return nil, errors.New("--project and --profile are required")
		}
		var result application.CrawlResult
		err = client.Call(ctx, http.MethodPost, "/api/v1/projects/"+url.PathEscape(*project)+"/crawls", map[string]string{"profile_id": *profile}, &result)
		return result, err
	case "crawl-list":
		project, err := requiredFlag(command, arguments, "project")
		if err != nil {
			return nil, err
		}
		var result contracts.Page[contracts.CrawlProgress]
		err = client.Call(ctx, http.MethodGet, "/api/v1/projects/"+url.PathEscape(project)+"/crawls?limit=100", nil, &result)
		return result, err
	case "crawl-status":
		return getByCrawl(ctx, client, command, arguments, "status")
	case "crawl-timeline":
		return getByCrawl(ctx, client, command, arguments, "timeline?limit=100")
	case "audit-summary":
		return getByCrawl(ctx, client, command, arguments, "summary")
	case "issue-list":
		return getByCrawl(ctx, client, command, arguments, "issues?limit=100")
	case "page-list":
		return getByCrawl(ctx, client, command, arguments, "pages?limit=100")
	case "crawl-pause", "crawl-resume", "crawl-cancel":
		crawl, err := requiredFlag(command, arguments, "crawl")
		if err != nil {
			return nil, err
		}
		action := strings.TrimPrefix(command, "crawl-")
		err = client.Call(ctx, http.MethodPost, "/api/v1/crawls/"+url.PathEscape(crawl)+"/"+action, nil, nil)
		return map[string]string{"crawl_id": crawl, "status": action + "_requested"}, err
	case "issue-explain":
		flags := flag.NewFlagSet(command, flag.ContinueOnError)
		crawl := flags.String("crawl", "", "crawl ID")
		issue := flags.Int64("issue", 0, "issue ID")
		if err := flags.Parse(arguments); err != nil {
			return nil, err
		}
		if *crawl == "" || *issue < 1 {
			return nil, errors.New("--crawl and a positive --issue are required")
		}
		var result application.IssueExplanation
		err = client.Call(ctx, http.MethodGet, fmt.Sprintf("/api/v1/crawls/%s/issues/%d", url.PathEscape(*crawl), *issue), nil, &result)
		return result, err
	case "page-get":
		flags := flag.NewFlagSet(command, flag.ContinueOnError)
		crawl := flags.String("crawl", "", "crawl ID")
		page := flags.Int64("page", 0, "page ID")
		if err := flags.Parse(arguments); err != nil {
			return nil, err
		}
		if *crawl == "" || *page < 1 {
			return nil, errors.New("--crawl and a positive --page are required")
		}
		var result database.PageDetail
		err = client.Call(ctx, http.MethodGet, fmt.Sprintf("/api/v1/crawls/%s/pages/%d", url.PathEscape(*crawl), *page), nil, &result)
		return result, err
	case "crawl-compare":
		flags := flag.NewFlagSet(command, flag.ContinueOnError)
		base := flags.String("base", "", "base crawl ID")
		target := flags.String("target", "", "target crawl ID")
		if err := flags.Parse(arguments); err != nil {
			return nil, err
		}
		if *base == "" || *target == "" {
			return nil, errors.New("--base and --target are required")
		}
		var result database.CrawlComparison
		err = client.Call(ctx, http.MethodPost, "/api/v1/comparisons", map[string]string{"base_crawl_id": *base, "target_crawl_id": *target}, &result)
		return result, err
	case "report-export":
		flags := flag.NewFlagSet(command, flag.ContinueOnError)
		crawl := flags.String("crawl", "", "crawl ID")
		dataset := flags.String("dataset", "workbook", "pages, issues, or workbook")
		format := flags.String("format", "xlsx", "csv, ndjson, or xlsx")
		if err := flags.Parse(arguments); err != nil {
			return nil, err
		}
		if *crawl == "" {
			return nil, errors.New("--crawl is required")
		}
		var result application.Artifact
		err = client.Call(ctx, http.MethodPost, "/api/v1/exports", application.ExportRequest{CrawlID: contracts.ID(*crawl), Dataset: *dataset, Format: *format}, &result)
		return result, err
	case "artifact-get":
		artifact, err := requiredFlag(command, arguments, "artifact")
		if err != nil {
			return nil, err
		}
		var result application.Artifact
		err = client.Call(ctx, http.MethodGet, "/api/v1/artifacts/"+url.PathEscape(artifact), nil, &result)
		return result, err
	case "diagnostic-create":
		crawl, err := requiredFlag(command, arguments, "crawl")
		if err != nil {
			return nil, err
		}
		var result application.Artifact
		err = client.Call(ctx, http.MethodPost, "/api/v1/crawls/"+url.PathEscape(crawl)+"/diagnostics", nil, &result)
		return result, err
	default:
		usage()
		return nil, fmt.Errorf("unknown command %q", command)
	}
}

func requiredFlag(command string, arguments []string, name string) (string, error) {
	flags := flag.NewFlagSet(command, flag.ContinueOnError)
	value := flags.String(name, "", name+" ID")
	if err := flags.Parse(arguments); err != nil {
		return "", err
	}
	if *value == "" {
		return "", fmt.Errorf("--%s is required", name)
	}
	return *value, nil
}
func getByCrawl(ctx context.Context, client *localclient.Client, command string, arguments []string, action string) (any, error) {
	crawl, err := requiredFlag(command, arguments, "crawl")
	if err != nil {
		return nil, err
	}
	var result any
	err = client.Call(ctx, http.MethodGet, "/api/v1/crawls/"+url.PathEscape(crawl)+"/"+action, nil, &result)
	return result, err
}

func runStandaloneCrawl(ctx context.Context, arguments []string) (application.CrawlResult, error) {
	flags := flag.NewFlagSet("crawl", flag.ContinueOnError)
	seed := flags.String("url", "", "public HTTP or HTTPS seed URL")
	seedList := flags.String("urls", "", "comma or newline separated list-mode seed URLs")
	name := flags.String("name", "Audit", "project name")
	maximumURLs := flags.Int64("max-urls", 10_000, "maximum URLs")
	responseCompression := flags.String("response-compression", "gzip", "response compression: gzip or disabled")
	if err := flags.Parse(arguments); err != nil {
		return application.CrawlResult{}, err
	}
	if (*seed == "") == (*seedList == "") {
		return application.CrawlResult{}, errors.New("provide exactly one of --url or --urls")
	}
	var seeds []string
	if *seedList != "" {
		seeds = strings.FieldsFunc(*seedList, func(r rune) bool { return r == ',' || r == '\n' || r == '\r' })
	}
	paths, err := config.ResolvePaths()
	if err != nil {
		return application.CrawlResult{}, err
	}
	if err := paths.Ensure(); err != nil {
		return application.CrawlResult{}, err
	}
	service, err := application.Open(ctx, paths.Data)
	if err != nil {
		return application.CrawlResult{}, err
	}
	defer service.Close()
	limits := contracts.DefaultCrawlLimits()
	limits.MaximumURLs = *maximumURLs
	limits.MaximumDuration = 24 * time.Hour
	return service.Crawl(ctx, application.CrawlRequest{ProjectName: *name, SeedURL: *seed, SeedURLs: seeds, ResponseCompression: *responseCompression, Limits: limits})
}
func usage() {
	fmt.Fprintln(os.Stderr, "usage: seo-auditor-cli <version|crawl|project-create|project-list|profile-create|profile-list|scope-preview|crawl-start|crawl-list|crawl-status|crawl-timeline|crawl-pause|crawl-resume|crawl-cancel|audit-summary|issue-list|issue-explain|page-list|page-get|crawl-compare|report-export|diagnostic-create|artifact-get> [options]")
}
