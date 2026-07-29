package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/seo-auditor/seo-auditor/internal/application"
	"github.com/seo-auditor/seo-auditor/internal/config"
	"github.com/seo-auditor/seo-auditor/internal/contracts"
	"github.com/seo-auditor/seo-auditor/internal/version"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	var err error
	switch os.Args[1] {
	case "version":
		err = json.NewEncoder(os.Stdout).Encode(map[string]string{"version": version.Version, "commit": version.Commit})
	case "crawl":
		err = runCrawl(os.Args[2:])
	default:
		usage()
		os.Exit(2)
	}
	if err != nil {
		_ = json.NewEncoder(os.Stderr).Encode(map[string]string{"error": err.Error()})
		os.Exit(1)
	}
}

func runCrawl(arguments []string) error {
	flags := flag.NewFlagSet("crawl", flag.ContinueOnError)
	seed := flags.String("url", "", "public HTTP or HTTPS seed URL")
	name := flags.String("name", "Audit", "project name")
	maximumURLs := flags.Int64("max-urls", 10_000, "maximum URLs to discover")
	maximumDepth := flags.Int("max-depth", 50, "maximum link depth")
	concurrency := flags.Int("concurrency", 16, "global request concurrency")
	perHost := flags.Int("per-host", 2, "per-host request concurrency")
	allowSubdomains := flags.Bool("allow-subdomains", false, "include subdomains of the seed host")
	if err := flags.Parse(arguments); err != nil {
		return err
	}
	if *seed == "" {
		return fmt.Errorf("--url is required")
	}
	paths, err := config.ResolvePaths()
	if err != nil {
		return err
	}
	if err := paths.Ensure(); err != nil {
		return err
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	service, err := application.Open(ctx, paths.Data)
	if err != nil {
		return err
	}
	defer service.Close()
	limits := contracts.DefaultCrawlLimits()
	limits.MaximumURLs = *maximumURLs
	limits.MaximumDepth = *maximumDepth
	limits.GlobalConcurrency = *concurrency
	limits.PerHostConcurrency = *perHost
	limits.MaximumDuration = 24 * time.Hour
	result, err := service.Crawl(ctx, application.CrawlRequest{
		ProjectName: *name, SeedURL: *seed, AllowSubdomains: *allowSubdomains, Limits: limits,
	})
	if encodeErr := json.NewEncoder(os.Stdout).Encode(result); encodeErr != nil && err == nil {
		return encodeErr
	}
	return err
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: seo-auditor-cli <version|crawl> [options]")
}
