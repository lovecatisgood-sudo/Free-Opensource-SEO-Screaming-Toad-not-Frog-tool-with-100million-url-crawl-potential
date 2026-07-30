package application

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/seo-auditor/seo-auditor/internal/contracts"
	"github.com/seo-auditor/seo-auditor/internal/database"
	"github.com/seo-auditor/seo-auditor/internal/integrations"
)

type SecretStatus struct {
	Reference string `json:"reference"`
	Available bool   `json:"available"`
}

func (s *Service) PutSecret(ctx context.Context, reference string, value []byte) (SecretStatus, error) {
	if s.secrets == nil {
		return SecretStatus{}, errors.New("secure credential store is unavailable")
	}
	if err := s.secrets.Put(ctx, reference, value); err != nil {
		return SecretStatus{}, err
	}
	return SecretStatus{Reference: reference, Available: true}, nil
}
func (s *Service) DeleteSecret(ctx context.Context, reference string) error {
	if s.secrets == nil {
		return errors.New("secure credential store is unavailable")
	}
	return s.secrets.Delete(ctx, reference)
}
func (s *Service) SecretStoreStatus(ctx context.Context) (SecretStatus, error) {
	if s.secrets == nil {
		return SecretStatus{Available: false}, errors.New("secure credential store is unavailable")
	}
	err := s.secrets.Available(ctx)
	return SecretStatus{Available: err == nil}, err
}

type IntegrationContext struct {
	ProjectID           contracts.ID `json:"project_id"`
	CrawlID             contracts.ID `json:"crawl_id,omitempty"`
	CredentialReference string       `json:"credential_reference"`
}

func (s *Service) RunPageSpeed(ctx context.Context, scope IntegrationContext, target, strategy string) (database.IntegrationObservationRecord, error) {
	value, err := s.integrations.PageSpeed(ctx, target, strategy, scope.CredentialReference)
	if err != nil {
		return database.IntegrationObservationRecord{}, err
	}
	return s.storeObservation(ctx, scope, value.Provider, value.EvidenceSource, value.ProfileVersion, value.Scope, value.Freshness, value.ObservedAt, value)
}
func (s *Service) RunCrUX(ctx context.Context, scope IntegrationContext, input integrations.CrUXRequest) (database.IntegrationObservationRecord, error) {
	value, err := s.integrations.CrUX(ctx, input, scope.CredentialReference)
	if err != nil {
		return database.IntegrationObservationRecord{}, err
	}
	return s.storeObservation(ctx, scope, value.Provider, value.EvidenceSource, value.ProfileVersion, value.Scope, value.Freshness, value.ObservedAt, value)
}
func (s *Service) RunSearchConsole(ctx context.Context, scope IntegrationContext, input integrations.SearchConsoleRequest) (database.IntegrationObservationRecord, error) {
	value, err := s.integrations.SearchConsole(ctx, input, scope.CredentialReference)
	if err != nil {
		return database.IntegrationObservationRecord{}, err
	}
	return s.storeObservation(ctx, scope, value.Provider, value.EvidenceSource, value.ProfileVersion, value.Scope, value.Freshness, value.ObservedAt, value)
}
func (s *Service) RunGA4(ctx context.Context, scope IntegrationContext, input integrations.GA4Request) (database.IntegrationObservationRecord, error) {
	value, err := s.integrations.GA4(ctx, input, scope.CredentialReference)
	if err != nil {
		return database.IntegrationObservationRecord{}, err
	}
	return s.storeObservation(ctx, scope, value.Provider, value.EvidenceSource, value.ProfileVersion, value.Scope, value.Freshness, value.ObservedAt, value)
}
func (s *Service) ListIntegrationObservations(ctx context.Context, projectID contracts.ID, provider string, page contracts.PageRequest) (contracts.Page[database.IntegrationObservationRecord], error) {
	return s.frontier.ListIntegrationObservations(ctx, projectID, provider, page)
}
func (s *Service) storeObservation(ctx context.Context, scope IntegrationContext, provider, source, version, target, freshness, observed string, value any) (database.IntegrationObservationRecord, error) {
	if err := s.frontier.ProjectExists(ctx, scope.ProjectID); err != nil {
		return database.IntegrationObservationRecord{}, err
	}
	id, err := contracts.NewID("observation")
	if err != nil {
		return database.IntegrationObservationRecord{}, err
	}
	body, err := json.Marshal(value)
	if err != nil {
		return database.IntegrationObservationRecord{}, err
	}
	record := database.IntegrationObservationRecord{ID: id, ProjectID: scope.ProjectID, CrawlID: scope.CrawlID, Provider: provider, EvidenceSource: source, ProfileVersion: version, Scope: target, Freshness: freshness, Result: body, ObservedAt: observed, CreatedAt: time.Now().UTC().Format(time.RFC3339Nano)}
	if err := s.frontier.RecordIntegrationObservation(ctx, record); err != nil {
		return database.IntegrationObservationRecord{}, err
	}
	return record, nil
}
