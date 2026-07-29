package application

import (
	"context"
	"errors"
	"strings"

	"github.com/seo-auditor/seo-auditor/internal/contracts"
	"github.com/seo-auditor/seo-auditor/internal/database"
	"github.com/seo-auditor/seo-auditor/internal/fetchpolicy"
	"github.com/seo-auditor/seo-auditor/internal/rules"
)

func (s *Service) CreateProject(ctx context.Context, name string) (database.ProjectRecord, error) {
	id, err := contracts.NewID("project")
	if err != nil {
		return database.ProjectRecord{}, err
	}
	if err := s.frontier.CreateProject(ctx, id, strings.TrimSpace(name)); err != nil {
		return database.ProjectRecord{}, err
	}
	return s.frontier.GetProject(ctx, id)
}
func (s *Service) ListProjects(ctx context.Context, page contracts.PageRequest) (contracts.Page[database.ProjectRecord], error) {
	return s.frontier.ListProjects(ctx, page)
}
func (s *Service) RenameProject(ctx context.Context, id contracts.ID, name string) error {
	return s.frontier.RenameProject(ctx, id, strings.TrimSpace(name))
}
func (s *Service) ArchiveProject(ctx context.Context, id contracts.ID, archived bool) error {
	return s.frontier.SetProjectArchived(ctx, id, archived)
}
func (s *Service) TrashProject(ctx context.Context, id contracts.ID) error {
	return s.frontier.TrashProject(ctx, id)
}
func (s *Service) RestoreProject(ctx context.Context, id contracts.ID) error {
	return s.frontier.RestoreProject(ctx, id)
}

func validateConfiguration(configuration contracts.CrawlConfiguration) (contracts.CrawlConfiguration, error) {
	seed, err := fetchpolicy.NormalizeURL(configuration.SeedURL)
	if err != nil {
		return configuration, err
	}
	configuration.SeedURL = seed.RequestKey
	if len(configuration.AllowedHosts) == 0 {
		configuration.AllowedHosts = []string{seed.URL.Hostname()}
	}
	if configuration.UserAgent == "" {
		configuration.UserAgent = UserAgent
	}
	if len(configuration.UserAgent) > 256 || strings.ContainsAny(configuration.UserAgent, "\r\n") {
		return configuration, errors.New("user agent is invalid")
	}
	if configuration.RenderingMode == "" {
		configuration.RenderingMode = "raw"
	}
	if configuration.RenderingMode != "raw" && configuration.RenderingMode != "rendered" {
		return configuration, errors.New("rendering mode must be raw or rendered")
	}
	if distance := configuration.EffectiveNearDuplicateDistance(); distance < 0 || distance > 3 {
		return configuration, errors.New("near-duplicate distance must be between 0 and 3")
	}
	if size := configuration.EffectiveSegmentSize(); size < 10_000 || size > 100_000 {
		return configuration, errors.New("segment size must be between 10000 and 100000 URLs")
	}
	if err := configuration.Limits.Validate(); err != nil {
		return configuration, err
	}
	_, err = fetchpolicy.CompileScope(fetchpolicy.ScopeConfig{AllowedHosts: configuration.AllowedHosts, AllowSubdomains: configuration.AllowSubdomains, IncludePathRegex: configuration.IncludePathRegex, ExcludePathRegex: configuration.ExcludePathRegex, IncludeQueryRegex: configuration.IncludeQueryRegex, ExcludeQueryRegex: configuration.ExcludeQueryRegex})
	return configuration, err
}
func (s *Service) CreateProfile(ctx context.Context, projectID contracts.ID, name string, configuration contracts.CrawlConfiguration) (database.ProfileRecord, error) {
	if err := s.frontier.ProjectExists(ctx, projectID); err != nil {
		return database.ProfileRecord{}, err
	}
	configuration, err := validateConfiguration(configuration)
	if err != nil {
		return database.ProfileRecord{}, err
	}
	id, err := contracts.NewID("profile")
	if err != nil {
		return database.ProfileRecord{}, err
	}
	return s.frontier.CreateProfile(ctx, id, projectID, strings.TrimSpace(name), configuration)
}
func (s *Service) ListProfiles(ctx context.Context, projectID contracts.ID, page contracts.PageRequest) (contracts.Page[database.ProfileRecord], error) {
	return s.frontier.ListProfiles(ctx, projectID, page)
}

type ScopeDecision struct {
	URL           string `json:"url"`
	NormalizedURL string `json:"normalized_url,omitempty"`
	Allowed       bool   `json:"allowed"`
	Reason        string `json:"reason,omitempty"`
}

func (s *Service) PreviewScope(ctx context.Context, configuration contracts.CrawlConfiguration, urls []string) ([]ScopeDecision, error) {
	_ = ctx
	configuration, err := validateConfiguration(configuration)
	if err != nil {
		return nil, err
	}
	if len(urls) == 0 {
		urls = []string{configuration.SeedURL}
	}
	if len(urls) > 100 {
		return nil, errors.New("scope preview is limited to 100 URLs")
	}
	scope, err := fetchpolicy.CompileScope(fetchpolicy.ScopeConfig{AllowedHosts: configuration.AllowedHosts, AllowSubdomains: configuration.AllowSubdomains, IncludePathRegex: configuration.IncludePathRegex, ExcludePathRegex: configuration.ExcludePathRegex, IncludeQueryRegex: configuration.IncludeQueryRegex, ExcludeQueryRegex: configuration.ExcludeQueryRegex})
	if err != nil {
		return nil, err
	}
	result := make([]ScopeDecision, 0, len(urls))
	for _, raw := range urls {
		item := ScopeDecision{URL: raw}
		normalized, normalizeErr := fetchpolicy.NormalizeURL(raw)
		if normalizeErr != nil {
			item.Reason = normalizeErr.Error()
			result = append(result, item)
			continue
		}
		item.NormalizedURL = normalized.RequestKey
		if scopeErr := scope.Evaluate(normalized); scopeErr != nil {
			item.Reason = scopeErr.Error()
		} else {
			item.Allowed = true
		}
		result = append(result, item)
	}
	return result, nil
}

type IssueExplanation struct {
	Issue database.IssueRecord `json:"issue"`
	Rule  rules.Metadata       `json:"rule"`
}

func (s *Service) ExplainIssue(ctx context.Context, crawlID contracts.ID, id int64) (IssueExplanation, error) {
	issue, err := s.frontier.GetIssue(ctx, crawlID, id)
	if err != nil {
		return IssueExplanation{}, err
	}
	for _, metadata := range rules.Catalog {
		if metadata.ID == issue.RuleID {
			return IssueExplanation{Issue: issue, Rule: metadata}, nil
		}
	}
	return IssueExplanation{}, errors.New("rule metadata is unavailable")
}
