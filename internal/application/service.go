package application

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/seo-auditor/seo-auditor/internal/config"
	"github.com/seo-auditor/seo-auditor/internal/contracts"
	"github.com/seo-auditor/seo-auditor/internal/crawler"
	"github.com/seo-auditor/seo-auditor/internal/database"
	"github.com/seo-auditor/seo-auditor/internal/fetchpolicy"
	"github.com/seo-auditor/seo-auditor/internal/renderer"
)

const UserAgent = "SEOAuditor/2.0 (+https://github.com/seo-auditor/seo-auditor)"

type Service struct {
	db             *database.DB
	frontier       *database.Frontier
	ctx            context.Context
	cancel         context.CancelFunc
	runs           sync.WaitGroup
	mu             sync.Mutex
	active         map[contracts.ID]bool
	artifactDir    string
	rendererConfig config.Renderer
}

func Open(ctx context.Context, dataDirectory string) (*Service, error) {
	if dataDirectory == "" {
		return nil, errors.New("data directory is required")
	}
	db, err := database.Open(ctx, filepath.Join(dataDirectory, "seo-auditor.db"))
	if err != nil {
		return nil, err
	}
	rendererConfig, err := config.ResolveRenderer()
	if err != nil {
		_ = db.Close()
		return nil, err
	}
	serviceCtx, cancel := context.WithCancel(ctx)
	artifactDir := filepath.Join(dataDirectory, "artifacts")
	if err := os.MkdirAll(artifactDir, 0o700); err != nil {
		_ = db.Close()
		cancel()
		return nil, err
	}
	service := &Service{db: db, frontier: database.NewFrontier(db, 1024), ctx: serviceCtx, cancel: cancel, active: make(map[contracts.ID]bool), artifactDir: artifactDir, rendererConfig: rendererConfig}
	if err := service.frontier.RecoverInterruptedCrawls(ctx); err != nil {
		service.frontier.Close()
		_ = db.Close()
		cancel()
		return nil, err
	}
	if err := service.cleanupArtifacts(ctx); err != nil {
		service.frontier.Close()
		_ = db.Close()
		cancel()
		return nil, err
	}
	return service, nil
}

func (s *Service) Close() error {
	s.cancel()
	s.runs.Wait()
	s.frontier.Close()
	return s.db.Close()
}

type CrawlRequest struct {
	ProjectName                                                              string
	SeedURL                                                                  string
	SeedURLs                                                                 []string
	AllowSubdomains                                                          bool
	IncludePathRegex, ExcludePathRegex, IncludeQueryRegex, ExcludeQueryRegex []string
	Limits                                                                   contracts.CrawlLimits
	ResponseCompression                                                      string
}

type CrawlResult struct {
	ProjectID contracts.ID            `json:"project_id"`
	CrawlID   contracts.ID            `json:"crawl_id"`
	Progress  contracts.CrawlProgress `json:"progress"`
}

type preparedCrawl struct {
	result     CrawlResult
	config     contracts.CrawlConfiguration
	seed       fetchpolicy.NormalizedURL
	seeds      []fetchpolicy.NormalizedURL
	scope      *fetchpolicy.Scope
	rawFetcher *fetchpolicy.Fetcher
	robots     *crawler.RobotsService
	renderer   *renderer.Supervisor
	transport  *http.Transport
}

func (s *Service) Crawl(ctx context.Context, request CrawlRequest) (CrawlResult, error) {
	prepared, err := s.prepare(ctx, request)
	if err != nil {
		return CrawlResult{}, err
	}
	err = s.run(ctx, prepared)
	progress, progressErr := s.frontier.Progress(context.Background(), prepared.result.CrawlID)
	prepared.result.Progress = progress
	if err != nil {
		return prepared.result, err
	}
	if progressErr != nil {
		return CrawlResult{}, progressErr
	}
	return prepared.result, nil
}

func (s *Service) StartCrawl(ctx context.Context, request CrawlRequest) (CrawlResult, error) {
	prepared, err := s.prepare(ctx, request)
	if err != nil {
		return CrawlResult{}, err
	}
	if err := s.launch(prepared); err != nil {
		return CrawlResult{}, err
	}
	return prepared.result, nil
}

func (s *Service) StartProfileCrawl(ctx context.Context, projectID, profileID contracts.ID) (CrawlResult, error) {
	profile, err := s.frontier.GetProfile(ctx, projectID, profileID)
	if err != nil {
		return CrawlResult{}, err
	}
	prepared, err := s.buildPrepared(ctx, CrawlResult{}, profile.Configuration)
	if err != nil {
		return CrawlResult{}, err
	}
	prepared, err = s.persistPrepared(ctx, prepared, projectID, profileID)
	if err != nil {
		return CrawlResult{}, err
	}
	if err := s.launch(prepared); err != nil {
		return CrawlResult{}, err
	}
	return prepared.result, nil
}

func (s *Service) launch(prepared preparedCrawl) error {
	s.mu.Lock()
	if s.active[prepared.result.CrawlID] {
		s.mu.Unlock()
		prepared.transport.CloseIdleConnections()
		return errors.New("crawl is already active")
	}
	s.active[prepared.result.CrawlID] = true
	s.mu.Unlock()
	s.runs.Add(1)
	go func() {
		defer s.runs.Done()
		defer func() {
			s.mu.Lock()
			delete(s.active, prepared.result.CrawlID)
			s.mu.Unlock()
		}()
		if runErr := s.run(s.ctx, prepared); runErr != nil {
			_ = s.frontier.SetStatus(context.Background(), prepared.result.CrawlID, []contracts.CrawlStatus{contracts.CrawlPending, contracts.CrawlRunning}, contracts.CrawlFailed, "background_error")
		}
	}()
	return nil
}

func (s *Service) prepare(ctx context.Context, request CrawlRequest) (preparedCrawl, error) {
	if request.ProjectName == "" {
		request.ProjectName = "Audit"
	}
	if err := request.Limits.Validate(); err != nil {
		return preparedCrawl{}, err
	}
	rawSeeds := request.SeedURLs
	if len(rawSeeds) == 0 {
		rawSeeds = []string{request.SeedURL}
	}
	if len(rawSeeds) < 1 || len(rawSeeds) > 10_000 {
		return preparedCrawl{}, errors.New("list mode requires between 1 and 10000 seed URLs")
	}
	seeds := make([]fetchpolicy.NormalizedURL, 0, len(rawSeeds))
	hosts := make([]string, 0, len(rawSeeds))
	seenURL, seenHost := make(map[string]struct{}), make(map[string]struct{})
	for _, raw := range rawSeeds {
		normalized, err := fetchpolicy.NormalizeURL(raw)
		if err != nil {
			return preparedCrawl{}, fmt.Errorf("invalid list seed: %w", err)
		}
		if _, exists := seenURL[normalized.RequestKey]; exists {
			continue
		}
		seenURL[normalized.RequestKey] = struct{}{}
		seeds = append(seeds, normalized)
		if _, exists := seenHost[normalized.URL.Hostname()]; !exists {
			seenHost[normalized.URL.Hostname()] = struct{}{}
			hosts = append(hosts, normalized.URL.Hostname())
		}
	}
	seed := seeds[0]
	configuration := contracts.CrawlConfiguration{
		SeedURL: seed.RequestKey, AllowedHosts: hosts, AllowSubdomains: request.AllowSubdomains,
		IncludePathRegex: request.IncludePathRegex, ExcludePathRegex: request.ExcludePathRegex,
		IncludeQueryRegex: request.IncludeQueryRegex, ExcludeQueryRegex: request.ExcludeQueryRegex,
		UserAgent: UserAgent, RenderingMode: "raw", ResponseCompression: request.ResponseCompression, Limits: request.Limits,
	}
	var err error
	configuration, err = validateConfiguration(configuration)
	if err != nil {
		return preparedCrawl{}, err
	}
	result := CrawlResult{}
	prepared, err := s.buildPrepared(ctx, result, configuration)
	if err != nil {
		return preparedCrawl{}, err
	}
	prepared.seeds = seeds
	projectID, err := contracts.NewID("project")
	if err != nil {
		return preparedCrawl{}, err
	}
	if err := s.frontier.CreateProject(ctx, projectID, request.ProjectName); err != nil {
		prepared.transport.CloseIdleConnections()
		return preparedCrawl{}, err
	}
	return s.persistPrepared(ctx, prepared, projectID, "")
}

func (s *Service) persistPrepared(ctx context.Context, prepared preparedCrawl, projectID, profileID contracts.ID) (preparedCrawl, error) {
	crawlID, err := contracts.NewID("crawl")
	if err != nil {
		prepared.transport.CloseIdleConnections()
		return preparedCrawl{}, err
	}
	seed := prepared.seed
	configuration := prepared.config
	if err := s.frontier.CreateCrawl(ctx, crawlID, projectID, profileID, seed, configuration); err != nil {
		prepared.transport.CloseIdleConnections()
		return preparedCrawl{}, err
	}
	seeds := prepared.seeds
	if len(seeds) == 0 {
		seeds = []fetchpolicy.NormalizedURL{seed}
	}
	discoveries := make([]database.Discovery, 0, len(seeds))
	kind := "seed"
	if len(seeds) > 1 {
		kind = "list"
	}
	for _, item := range seeds {
		discoveries = append(discoveries, database.Discovery{CrawlID: crawlID, ProjectID: projectID, URL: item, Depth: 0, DiscoveryKind: kind, MaximumURLs: configuration.Limits.MaximumURLs})
	}
	if _, err := s.frontier.EnqueueBatch(ctx, discoveries); err != nil {
		prepared.transport.CloseIdleConnections()
		return preparedCrawl{}, err
	}
	progress, err := s.frontier.Progress(ctx, crawlID)
	if err != nil {
		prepared.transport.CloseIdleConnections()
		return preparedCrawl{}, err
	}
	prepared.result = CrawlResult{ProjectID: projectID, CrawlID: crawlID, Progress: progress}
	return prepared, nil
}

func (s *Service) buildPrepared(ctx context.Context, result CrawlResult, configuration contracts.CrawlConfiguration) (preparedCrawl, error) {
	if err := configuration.Limits.Validate(); err != nil {
		return preparedCrawl{}, err
	}
	seed, err := fetchpolicy.NormalizeURL(configuration.SeedURL)
	if err != nil {
		return preparedCrawl{}, err
	}
	scope, err := fetchpolicy.CompileScope(fetchpolicy.ScopeConfig{AllowedHosts: configuration.AllowedHosts, AllowSubdomains: configuration.AllowSubdomains, IncludePathRegex: configuration.IncludePathRegex, ExcludePathRegex: configuration.ExcludePathRegex, IncludeQueryRegex: configuration.IncludeQueryRegex, ExcludeQueryRegex: configuration.ExcludeQueryRegex})
	if err != nil {
		return preparedCrawl{}, err
	}
	resolver := fetchpolicy.SystemResolver{}
	guard := &fetchpolicy.Guard{Resolver: resolver, Scope: scope}
	if _, err := guard.Validate(ctx, seed.RequestKey); err != nil {
		return preparedCrawl{}, fmt.Errorf("seed rejected: %w", err)
	}
	transport := fetchpolicy.NewHTTPTransport(resolver)
	fetchLimits := fetchpolicy.DefaultFetchLimits()
	fetchLimits.MaximumDecodedBytes = configuration.Limits.MaximumBodyBytes
	fetchLimits.MaximumCompressedBytes = min(fetchLimits.MaximumCompressedBytes, configuration.Limits.MaximumBodyBytes)
	fetchLimits.OmitAcceptEncoding = configuration.EffectiveResponseCompression() == "disabled"
	rawFetcher, err := fetchpolicy.NewFetcher(guard, transport, fetchLimits, configuration.UserAgent)
	if err != nil {
		return preparedCrawl{}, err
	}
	robots, err := crawler.NewRobotsService(rawFetcher, "SEOAuditor", 12*time.Hour)
	if err != nil {
		return preparedCrawl{}, err
	}
	var renderSupervisor *renderer.Supervisor
	if configuration.RenderingMode == "rendered" {
		if !s.rendererConfig.Enabled {
			transport.CloseIdleConnections()
			return preparedCrawl{}, errors.New("rendered mode is unavailable; configure the trusted renderer worker at startup")
		}
		renderSupervisor = &renderer.Supervisor{
			NodeBinary: s.rendererConfig.NodeBinary, ScriptPath: s.rendererConfig.ScriptPath,
			BrowserPath: s.rendererConfig.BrowserPath, ContainerSandbox: s.rendererConfig.ContainerSandbox,
			Fetcher: rawFetcher,
		}
	}
	return preparedCrawl{result: result, config: configuration, seed: seed, seeds: []fetchpolicy.NormalizedURL{seed}, scope: scope, rawFetcher: rawFetcher, robots: robots, renderer: renderSupervisor, transport: transport}, nil
}

func (s *Service) run(ctx context.Context, prepared preparedCrawl) error {
	defer prepared.transport.CloseIdleConnections()
	if err := s.frontier.SetStatus(ctx, prepared.result.CrawlID, []contracts.CrawlStatus{contracts.CrawlPending, contracts.CrawlPaused, contracts.CrawlRunning}, contracts.CrawlRunning, ""); err != nil {
		return err
	}
	robotsSitemaps, err := prepared.robots.Sitemaps(ctx, prepared.seed.RequestKey)
	if err != nil {
		return err
	}
	commonSitemap := prepared.seed.URL.Scheme + "://" + prepared.seed.URL.Host + "/sitemap.xml"
	roots := append(robotsSitemaps, commonSitemap)
	discoveryLimits := crawler.DefaultSitemapDiscoveryLimits()
	discoveryLimits.MaximumURLs = int(min(prepared.config.Limits.MaximumURLs, int64(discoveryLimits.MaximumURLs)))
	sitemapURLs, sitemapEvidence, err := crawler.DiscoverSitemaps(ctx, prepared.rawFetcher, prepared.scope, roots, discoveryLimits)
	if err != nil && len(sitemapEvidence) == 0 {
		return err
	}
	records := make([]database.SitemapRecord, 0, len(sitemapEvidence))
	for _, item := range sitemapEvidence {
		status := "ok"
		if item.Error != "" {
			status = "error: " + item.Error
		} else if item.StatusCode != 0 && (item.StatusCode < 200 || item.StatusCode >= 300) {
			status = fmt.Sprintf("http_%d", item.StatusCode)
		}
		records = append(records, database.SitemapRecord{URL: item.URL, Status: status, DiscoveredFrom: "robots_or_common", Entries: item.Locations})
	}
	if err := s.frontier.RecordSitemaps(ctx, prepared.result.CrawlID, prepared.result.ProjectID, records); err != nil {
		return err
	}
	for _, raw := range sitemapURLs {
		target, normalizeErr := fetchpolicy.NormalizeURL(raw)
		if normalizeErr != nil {
			continue
		}
		if _, enqueueErr := s.frontier.Enqueue(ctx, database.Discovery{CrawlID: prepared.result.CrawlID, ProjectID: prepared.result.ProjectID, URL: target, Depth: 0, DiscoveryKind: "sitemap", MaximumURLs: prepared.config.Limits.MaximumURLs}); enqueueErr != nil && !errors.Is(enqueueErr, database.ErrURLLimitReached) {
			return enqueueErr
		}
	}
	progress, err := s.frontier.Progress(ctx, prepared.result.CrawlID)
	if err != nil {
		return err
	}
	if progress.Status == contracts.CrawlCancelling {
		return s.frontier.SetStatus(ctx, prepared.result.CrawlID, []contracts.CrawlStatus{contracts.CrawlCancelling}, contracts.CrawlCancelled, "user_cancelled")
	}
	if progress.Status == contracts.CrawlPausing {
		return s.frontier.SetStatus(ctx, prepared.result.CrawlID, []contracts.CrawlStatus{contracts.CrawlPausing}, contracts.CrawlPaused, "")
	}
	engine := &crawler.Engine{
		Frontier: s.frontier,
		Fetcher:  crawler.RobotsEnforcingFetcher{Base: prepared.rawFetcher, Robots: prepared.robots},
		Scope:    prepared.scope, LeaseTime: 2 * time.Minute, MaxLinksPerPage: 10_000,
		Renderer: prepared.renderer,
	}
	return engine.Run(ctx, crawler.RunRequest{
		CrawlID: prepared.result.CrawlID, ProjectID: prepared.result.ProjectID, Limits: prepared.config.Limits, WorkerID: "local", RenderingMode: prepared.config.RenderingMode, NearDuplicateDistance: prepared.config.EffectiveNearDuplicateDistance(), SegmentSize: prepared.config.EffectiveSegmentSize(),
	})
}

func (s *Service) Progress(ctx context.Context, crawlID contracts.ID) (contracts.CrawlProgress, error) {
	return s.frontier.Progress(ctx, crawlID)
}

func (s *Service) Summary(ctx context.Context, crawlID contracts.ID) (database.AuditSummary, error) {
	return s.frontier.Summary(ctx, crawlID)
}
func (s *Service) ListPages(ctx context.Context, crawlID contracts.ID, page contracts.PageRequest) (contracts.Page[database.PageRecord], error) {
	return s.frontier.ListPages(ctx, crawlID, page)
}
func (s *Service) ListIssues(ctx context.Context, crawlID contracts.ID, page contracts.PageRequest) (contracts.Page[database.IssueRecord], error) {
	return s.frontier.ListIssues(ctx, crawlID, page)
}
func (s *Service) ListLinks(ctx context.Context, crawlID contracts.ID, page contracts.PageRequest) (contracts.Page[database.LinkRecord], error) {
	return s.frontier.ListLinks(ctx, crawlID, page)
}
func (s *Service) ListEvents(ctx context.Context, crawlID contracts.ID, page contracts.PageRequest) (contracts.Page[database.CrawlEventRecord], error) {
	return s.frontier.ListEvents(ctx, crawlID, page)
}
func (s *Service) GetPage(ctx context.Context, crawlID contracts.ID, pageID int64) (database.PageDetail, error) {
	return s.frontier.GetPage(ctx, crawlID, pageID)
}
func (s *Service) CompareCrawls(ctx context.Context, baseID, targetID contracts.ID) (database.CrawlComparison, error) {
	return s.frontier.CompareCrawls(ctx, baseID, targetID)
}
func (s *Service) ListCrawls(ctx context.Context, projectID contracts.ID, page contracts.PageRequest) (contracts.Page[contracts.CrawlProgress], error) {
	return s.frontier.ListCrawls(ctx, projectID, page)
}
func (s *Service) Cancel(ctx context.Context, crawlID contracts.ID) error {
	return s.frontier.RequestCancel(ctx, crawlID)
}

func (s *Service) Pause(ctx context.Context, crawlID contracts.ID) error {
	return s.frontier.RequestPause(ctx, crawlID)
}

func (s *Service) Resume(ctx context.Context, crawlID contracts.ID) error {
	stored, err := s.frontier.LoadCrawl(ctx, crawlID)
	if err != nil {
		return err
	}
	if stored.Status != contracts.CrawlPaused {
		return fmt.Errorf("crawl must be paused to resume")
	}
	prepared, err := s.buildPrepared(ctx, CrawlResult{ProjectID: stored.ProjectID, CrawlID: stored.CrawlID}, stored.Configuration)
	if err != nil {
		return err
	}
	return s.launch(prepared)
}
