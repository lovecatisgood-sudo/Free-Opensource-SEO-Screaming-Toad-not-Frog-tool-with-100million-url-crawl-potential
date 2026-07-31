package application

import (
	"context"

	"github.com/seo-auditor/seo-auditor/internal/contracts"
	"github.com/seo-auditor/seo-auditor/internal/database"
)

func (s *Service) Architecture(ctx context.Context, crawlID contracts.ID, nodeLimit, edgeLimit int) (database.ArchitectureGraph, error) {
	return s.frontier.Architecture(ctx, crawlID, nodeLimit, edgeLimit)
}
