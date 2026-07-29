package crawler

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/seo-auditor/seo-auditor/internal/contracts"
	"github.com/seo-auditor/seo-auditor/internal/database"
	"github.com/seo-auditor/seo-auditor/internal/extractor"
	"github.com/seo-auditor/seo-auditor/internal/fetchpolicy"
	"github.com/seo-auditor/seo-auditor/internal/rules"
)

type Fetcher interface {
	Fetch(context.Context, string) (fetchpolicy.FetchResult, error)
}

type ScopeEvaluator interface {
	Evaluate(fetchpolicy.NormalizedURL) error
}

type Engine struct {
	Frontier        *database.Frontier
	Fetcher         Fetcher
	Scope           ScopeEvaluator
	LeaseTime       time.Duration
	MaxLinksPerPage int
}

type RunRequest struct {
	CrawlID   contracts.ID
	ProjectID contracts.ID
	Limits    contracts.CrawlLimits
	WorkerID  string
}

type workResult struct {
	lease database.Lease
	fetch fetchpolicy.FetchResult
	err   error
}

func (e *Engine) Run(ctx context.Context, request RunRequest) error {
	if e.Frontier == nil || e.Fetcher == nil || e.Scope == nil {
		return errors.New("crawler dependencies are incomplete")
	}
	if err := request.Limits.Validate(); err != nil {
		return err
	}
	if request.WorkerID == "" {
		return errors.New("worker ID is required")
	}
	leaseTime := e.LeaseTime
	if leaseTime <= 0 {
		leaseTime = 2 * time.Minute
	}
	maxLinks := e.MaxLinksPerPage
	if maxLinks <= 0 {
		maxLinks = 10_000
	}
	if err := e.Frontier.SetStatus(ctx, request.CrawlID, []contracts.CrawlStatus{contracts.CrawlPending, contracts.CrawlPaused, contracts.CrawlRunning}, contracts.CrawlRunning, ""); err != nil {
		return err
	}
	runCtx, cancel := context.WithTimeout(ctx, request.Limits.MaximumDuration)
	defer cancel()
	hosts := newHostCoordinator(request.Limits.PerHostConcurrency, request.Limits.MinimumHostDelay)

	for {
		if err := runCtx.Err(); err != nil {
			status := contracts.CrawlLimited
			reason := "duration_limit"
			if ctx.Err() != nil {
				status = contracts.CrawlCancelled
				reason = "context_cancelled"
			}
			_ = e.Frontier.SetStatus(context.Background(), request.CrawlID, []contracts.CrawlStatus{contracts.CrawlRunning}, status, reason)
			return err
		}
		progress, err := e.Frontier.Progress(runCtx, request.CrawlID)
		if err != nil {
			return err
		}
		storage, err := e.Frontier.StorageBytes()
		if err != nil {
			return err
		}
		if storage > request.Limits.MaximumDiskBytes {
			_ = e.Frontier.SetStatus(runCtx, request.CrawlID, []contracts.CrawlStatus{contracts.CrawlRunning}, contracts.CrawlLimited, "disk_limit")
			return fmt.Errorf("crawl disk limit reached")
		}
		switch progress.Status {
		case contracts.CrawlPausing:
			return e.Frontier.SetStatus(runCtx, request.CrawlID, []contracts.CrawlStatus{contracts.CrawlPausing}, contracts.CrawlPaused, "")
		case contracts.CrawlCancelling:
			return e.Frontier.SetStatus(runCtx, request.CrawlID, []contracts.CrawlStatus{contracts.CrawlCancelling}, contracts.CrawlCancelled, "user_cancelled")
		}
		leases, err := e.Frontier.Lease(runCtx, request.CrawlID, request.WorkerID, request.Limits.GlobalConcurrency, leaseTime)
		if err != nil {
			return err
		}
		if len(leases) == 0 {
			if progress.Queued == 0 {
				return e.Frontier.SetStatus(runCtx, request.CrawlID, []contracts.CrawlStatus{contracts.CrawlRunning}, contracts.CrawlCompleted, "")
			}
			timer := time.NewTimer(100 * time.Millisecond)
			select {
			case <-runCtx.Done():
				timer.Stop()
				continue
			case <-timer.C:
				continue
			}
		}
		results := make(chan workResult, len(leases))
		var wait sync.WaitGroup
		for _, lease := range leases {
			lease := lease
			wait.Add(1)
			go func() {
				defer wait.Done()
				release, err := hosts.acquire(runCtx, lease.URL)
				if err != nil {
					results <- workResult{lease: lease, err: err}
					return
				}
				defer release()
				result, err := e.Fetcher.Fetch(runCtx, lease.URL)
				results <- workResult{lease: lease, fetch: result, err: err}
			}()
		}
		wait.Wait()
		close(results)
		for result := range results {
			if err := e.commitResult(runCtx, request, result, maxLinks); err != nil {
				if errors.Is(err, database.ErrURLLimitReached) {
					_ = e.Frontier.SetStatus(runCtx, request.CrawlID, []contracts.CrawlStatus{contracts.CrawlRunning}, contracts.CrawlLimited, "url_limit")
				}
				return err
			}
		}
	}
}

type hostState struct {
	semaphore chan struct{}
	mu        sync.Mutex
	nextStart time.Time
}

type hostCoordinator struct {
	mu       sync.Mutex
	hosts    map[string]*hostState
	parallel int
	delay    time.Duration
}

func newHostCoordinator(parallel int, delay time.Duration) *hostCoordinator {
	return &hostCoordinator{hosts: make(map[string]*hostState), parallel: parallel, delay: delay}
}

func (c *hostCoordinator) acquire(ctx context.Context, raw string) (func(), error) {
	normalized, err := fetchpolicy.NormalizeURL(raw)
	if err != nil {
		return nil, err
	}
	host := normalized.URL.Host
	c.mu.Lock()
	state := c.hosts[host]
	if state == nil {
		state = &hostState{semaphore: make(chan struct{}, c.parallel)}
		c.hosts[host] = state
	}
	c.mu.Unlock()
	select {
	case state.semaphore <- struct{}{}:
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	state.mu.Lock()
	now := time.Now()
	wait := max(time.Duration(0), state.nextStart.Sub(now))
	state.nextStart = now.Add(wait).Add(c.delay)
	state.mu.Unlock()
	if wait > 0 {
		timer := time.NewTimer(wait)
		select {
		case <-ctx.Done():
			timer.Stop()
			<-state.semaphore
			return nil, ctx.Err()
		case <-timer.C:
		}
	}
	return func() { <-state.semaphore }, nil
}

func (e *Engine) commitResult(ctx context.Context, request RunRequest, result workResult, maxLinks int) error {
	if result.err != nil {
		var denied *RobotsDeniedError
		if errors.As(result.err, &denied) {
			return e.Frontier.Skip(ctx, request.CrawlID, result.lease, "robots_disallowed")
		}
		return e.failOrRetry(ctx, request.CrawlID, result.lease, 0, nil, result.err)
	}
	retry := fetchpolicy.ClassifyRetry(result.fetch.StatusCode, result.fetch.Header, nil, time.Now())
	if retry.Retry {
		return e.failOrRetry(ctx, request.CrawlID, result.lease, result.fetch.StatusCode, &retry, nil)
	}
	if err := e.Frontier.CompleteFetch(ctx, request.CrawlID, database.FetchCompletion{
		Lease: result.lease, StatusCode: result.fetch.StatusCode, ContentType: result.fetch.ContentType,
		CompressedBytes: result.fetch.CompressedBytes, DecodedBytes: result.fetch.DecodedBytes,
		StartedAt: result.fetch.StartedAt, FinishedAt: result.fetch.FinishedAt,
	}); err != nil {
		return err
	}
	if !isHTML(result.fetch.ContentType) {
		return nil
	}
	page, err := extractor.Extract(result.fetch.FinalURL, result.fetch.Header, result.fetch.Body)
	if err != nil {
		return nil // malformed HTML is evidence for extraction, not a crawl failure.
	}
	issues := rules.EvaluatePage(rules.PageInput{Page: page, StatusCode: result.fetch.StatusCode, Headers: result.fetch.Header, Depth: result.lease.Depth}, rules.DefaultThresholds())
	if err := e.Frontier.SaveAnalysis(ctx, request.CrawlID, request.ProjectID, result.lease, page, issues); err != nil {
		return err
	}
	if result.lease.Depth >= request.Limits.MaximumDepth {
		return nil
	}
	links := make([]string, 0, min(maxLinks, len(page.Links)))
	for _, link := range page.Links {
		if len(links) >= maxLinks {
			break
		}
		links = append(links, link.URL)
	}
	parent := result.lease.CrawlURLID
	for _, raw := range links {
		normalized, err := fetchpolicy.NormalizeURL(raw)
		if err != nil || e.Scope.Evaluate(normalized) != nil || DetectTrap(normalized) != "" {
			continue
		}
		_, err = e.Frontier.Enqueue(ctx, database.Discovery{
			CrawlID: request.CrawlID, ProjectID: request.ProjectID, URL: normalized,
			Depth: result.lease.Depth + 1, DiscoveryKind: "link", DiscoveredFrom: &parent,
			MaximumURLs: request.Limits.MaximumURLs,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (e *Engine) failOrRetry(ctx context.Context, crawlID contracts.ID, lease database.Lease, status int, decision *fetchpolicy.RetryDecision, fetchErr error) error {
	code := "fetch_failed"
	detail := "request failed"
	var retryAt *time.Time
	if fetchErr != nil {
		detail = fetchErr.Error()
		classified := fetchpolicy.ClassifyRetry(0, nil, fetchErr, time.Now())
		decision = &classified
	} else {
		code = fmt.Sprintf("http_%d", status)
		detail = http.StatusText(status)
	}
	if decision != nil && decision.Retry && lease.Attempt < 3 {
		delay := decision.After
		if delay <= 0 {
			delay = time.Duration(lease.Attempt*lease.Attempt) * time.Second
		}
		when := time.Now().Add(delay)
		retryAt = &when
	}
	return e.Frontier.Fail(ctx, crawlID, lease, code, detail, retryAt)
}

func isHTML(contentType string) bool {
	contentType = strings.ToLower(contentType)
	return strings.Contains(contentType, "text/html") || strings.Contains(contentType, "application/xhtml+xml")
}
