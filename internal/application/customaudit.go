package application

import (
	"context"

	"github.com/seo-auditor/seo-auditor/internal/contracts"
	"github.com/seo-auditor/seo-auditor/internal/customaudit"
	"github.com/seo-auditor/seo-auditor/internal/database"
)

func (s *Service) PutCustomAuditDefinition(ctx context.Context, projectID contracts.ID, definition customaudit.Definition) (database.CustomAuditDefinitionRecord, error) {
	return s.frontier.PutCustomAuditDefinition(ctx, projectID, definition)
}
func (s *Service) ListCustomAuditDefinitions(ctx context.Context, projectID contracts.ID) ([]database.CustomAuditDefinitionRecord, error) {
	return s.frontier.ListCustomAuditDefinitions(ctx, projectID)
}
func (s *Service) DeleteCustomAuditDefinition(ctx context.Context, projectID contracts.ID, id string) error {
	return s.frontier.DeleteCustomAuditDefinition(ctx, projectID, id)
}
func (s *Service) PreviewCustomAudit(_ context.Context, definition customaudit.Definition, document []byte) (customaudit.Result, error) {
	return customaudit.Execute(definition, document)
}
func (s *Service) ExportCustomAudits(ctx context.Context, projectID contracts.ID) ([]byte, error) {
	records, err := s.frontier.ListCustomAuditDefinitions(ctx, projectID)
	if err != nil {
		return nil, err
	}
	definitions := make([]customaudit.Definition, 0, len(records))
	for _, record := range records {
		definitions = append(definitions, record.Definition)
	}
	return customaudit.Export(definitions)
}
func (s *Service) ImportCustomAudits(ctx context.Context, projectID contracts.ID, body []byte) ([]database.CustomAuditDefinitionRecord, error) {
	definitions, err := customaudit.Import(body)
	if err != nil {
		return nil, err
	}
	result := make([]database.CustomAuditDefinitionRecord, 0, len(definitions))
	for _, definition := range definitions {
		record, err := s.frontier.PutCustomAuditDefinition(ctx, projectID, definition)
		if err != nil {
			return nil, err
		}
		result = append(result, record)
	}
	return result, nil
}
func (s *Service) ListCustomAuditResults(ctx context.Context, crawlID contracts.ID, definitionID string, page contracts.PageRequest) (contracts.Page[database.CustomAuditResultRecord], error) {
	return s.frontier.ListCustomAuditResults(ctx, crawlID, definitionID, page)
}
