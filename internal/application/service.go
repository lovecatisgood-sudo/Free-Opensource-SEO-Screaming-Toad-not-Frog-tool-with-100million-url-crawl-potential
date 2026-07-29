package application

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"time"

	"github.com/seo-auditor/seo-auditor/internal/contracts"
	"github.com/seo-auditor/seo-auditor/internal/crawler"
	"github.com/seo-auditor/seo-auditor/internal/database"
	"github.com/seo-auditor/seo-auditor/internal/fetchpolicy"
)

const UserAgent = "SEOAuditor/0.1 (+https://github.com/seo-auditor/seo-auditor)"

type Service struct {
	db       *database.DB
	frontier *database.Frontier
}

func Open(ctx context.Context, dataDirectory string) (*Service, error) {
	if dataDirectory == "" {
		return nil, errors.New("data directory is required")
	}
	db, err := database.Open(ctx, filepath.Join(dataDirectory, "seo-auditor.db"))
	if err != nil {
		return nil, err
	}
	return &Service{db: db, frontier: database.NewFrontier(db, 1024)}, nil
}

func (s *Service) Close() error {
	s.frontier.Close()
	return s.db.Close()
}

type CrawlRequest struct {
	ProjectName     string
	SeedURL         string
	AllowSubdomains bool
	Limits          contracts.CrawlLimits
}

type CrawlResult struct {
	ProjectID contracts.ID            `json:"project_id"`
	CrawlID   contracts.ID            `json:"crawl_id"`
	Progress  contracts.CrawlProgress `json:"progress"`
}

func (s *Service) Crawl(ctx context.Context, request CrawlRequest) (CrawlResult, error) {
	if request.ProjectName == "" {
		request.ProjectName = "Audit"
	}
	if err := request.Limits.Validate(); err != nil {
		return CrawlResult{}, err
	}
	seed, err := fetchpolicy.NormalizeURL(request.SeedURL)
	if err != nil {
		return CrawlResult{}, err
	}
	scope, err := fetchpolicy.CompileScope(fetchpolicy.ScopeConfig{
		AllowedHosts: []string{seed.URL.Hostname()}, AllowSubdomains: request.AllowSubdomains,
	})
	if err != nil {
		return CrawlResult{}, err
	}
	resolver := fetchpolicy.SystemResolver{}
	guard := &fetchpolicy.Guard{Resolver: resolver, Scope: scope}
	if _, err := guard.Validate(ctx, seed.RequestKey); err != nil {
		return CrawlResult{}, fmt.Errorf("seed rejected: %w", err)
	}
	transport := fetchpolicy.NewHTTPTransport(resolver)
	defer transport.CloseIdleConnections()
	fetchLimits := fetchpolicy.DefaultFetchLimits()
	fetchLimits.MaximumDecodedBytes = request.Limits.MaximumBodyBytes
	fetchLimits.MaximumCompressedBytes = min(fetchLimits.MaximumCompressedBytes, request.Limits.MaximumBodyBytes)
	rawFetcher, err := fetchpolicy.NewFetcher(guard, transport, fetchLimits, UserAgent)
	if err != nil {
		return CrawlResult{}, err
	}
	robots, err := crawler.NewRobotsService(rawFetcher, "SEOAuditor", 12*time.Hour)
	if err != nil {
		return CrawlResult{}, err
	}
	projectID, err := contracts.NewID("project")
	if err != nil {
		return CrawlResult{}, err
	}
	crawlID, err := contracts.NewID("crawl")
	if err != nil {
		return CrawlResult{}, err
	}
	if err := s.frontier.CreateProject(ctx, projectID, request.ProjectName); err != nil {
		return CrawlResult{}, err
	}
	if err := s.frontier.CreateCrawl(ctx, crawlID, projectID, seed, request.Limits); err != nil {
		return CrawlResult{}, err
	}
	if _, err := s.frontier.Enqueue(ctx, database.Discovery{
		CrawlID: crawlID, ProjectID: projectID, URL: seed, Depth: 0,
		DiscoveryKind: "seed", MaximumURLs: request.Limits.MaximumURLs,
	}); err != nil {
		return CrawlResult{}, err
	}
	engine := &crawler.Engine{
		Frontier: s.frontier,
		Fetcher:  crawler.RobotsEnforcingFetcher{Base: rawFetcher, Robots: robots},
		Scope:    scope, LeaseTime: 2 * time.Minute, MaxLinksPerPage: 10_000,
	}
	err = engine.Run(ctx, crawler.RunRequest{
		CrawlID: crawlID, ProjectID: projectID, Limits: request.Limits, WorkerID: "local",
	})
	progress, progressErr := s.frontier.Progress(context.Background(), crawlID)
	if err != nil {
		return CrawlResult{ProjectID: projectID, CrawlID: crawlID, Progress: progress}, err
	}
	if progressErr != nil {
		return CrawlResult{}, progressErr
	}
	return CrawlResult{ProjectID: projectID, CrawlID: crawlID, Progress: progress}, nil
}

func (s *Service) Progress(ctx context.Context, crawlID contracts.ID) (contracts.CrawlProgress, error) {
	return s.frontier.Progress(ctx, crawlID)
}
