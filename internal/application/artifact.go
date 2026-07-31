package application

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/seo-auditor/seo-auditor/internal/contracts"
	"github.com/seo-auditor/seo-auditor/internal/database"
	"github.com/seo-auditor/seo-auditor/internal/reports"
)

type ExportRequest struct {
	CrawlID contracts.ID `json:"crawl_id"`
	Dataset string       `json:"dataset"`
	Format  string       `json:"format"`
}
type Artifact struct {
	database.ArtifactRecord
	Path string `json:"path"`
}

func (s *Service) Export(ctx context.Context, request ExportRequest) (Artifact, error) {
	if request.CrawlID == "" {
		return Artifact{}, errors.New("crawl ID is required")
	}
	extension := request.Format
	if request.Format == "xlsx" {
		request.Dataset = "workbook"
	}
	if (request.Dataset != "pages" && request.Dataset != "issues" && request.Dataset != "workbook") || (request.Format != "csv" && request.Format != "ndjson" && request.Format != "xlsx") || (request.Dataset == "workbook" && request.Format != "xlsx") {
		return Artifact{}, errors.New("unsupported export dataset or format")
	}
	id, err := contracts.NewID("artifact")
	if err != nil {
		return Artifact{}, err
	}
	temporary, err := os.CreateTemp(s.artifactDir, ".export-*")
	if err != nil {
		return Artifact{}, err
	}
	temporaryName := temporary.Name()
	committed := false
	defer func() {
		_ = temporary.Close()
		if !committed {
			_ = os.Remove(temporaryName)
		}
	}()
	if err := temporary.Chmod(0o600); err != nil {
		return Artifact{}, err
	}
	hash := sha256.New()
	writer := io.MultiWriter(temporary, hash)
	switch {
	case request.Format == "xlsx":
		err = reports.WorkbookXLSX(ctx, s.frontier, request.CrawlID, writer)
	case request.Dataset == "pages" && request.Format == "csv":
		err = reports.PagesCSV(ctx, s.frontier, request.CrawlID, writer)
	case request.Dataset == "pages" && request.Format == "ndjson":
		err = reports.PagesNDJSON(ctx, s.frontier, request.CrawlID, writer)
	case request.Dataset == "issues" && request.Format == "csv":
		err = reports.IssuesCSV(ctx, s.frontier, request.CrawlID, writer)
	case request.Dataset == "issues" && request.Format == "ndjson":
		err = reports.IssuesNDJSON(ctx, s.frontier, request.CrawlID, writer)
	}
	if err != nil {
		return Artifact{}, err
	}
	if err := temporary.Sync(); err != nil {
		return Artifact{}, err
	}
	if err := temporary.Close(); err != nil {
		return Artifact{}, err
	}
	name := string(id) + "." + extension
	destination := filepath.Join(s.artifactDir, name)
	if err := os.Rename(temporaryName, destination); err != nil {
		return Artifact{}, err
	}
	info, err := os.Stat(destination)
	if err != nil {
		return Artifact{}, err
	}
	now := time.Now().UTC()
	record := database.ArtifactRecord{ID: id, CrawlID: request.CrawlID, Format: request.Format, RelativePath: name, Checksum: hex.EncodeToString(hash.Sum(nil)), SizeBytes: info.Size(), CreatedAt: now.Format(time.RFC3339Nano), ExpiresAt: now.Add(7 * 24 * time.Hour).Format(time.RFC3339Nano)}
	if err := s.frontier.RecordArtifact(ctx, record); err != nil {
		_ = os.Remove(destination)
		return Artifact{}, err
	}
	committed = true
	return Artifact{ArtifactRecord: record, Path: destination}, nil
}

func (s *Service) Artifact(ctx context.Context, id contracts.ID) (Artifact, error) {
	record, err := s.frontier.GetArtifact(ctx, id)
	if err != nil {
		return Artifact{}, err
	}
	if filepath.Base(record.RelativePath) != record.RelativePath {
		return Artifact{}, errors.New("invalid managed artifact path")
	}
	path := filepath.Join(s.artifactDir, record.RelativePath)
	if _, err := os.Stat(path); err != nil {
		return Artifact{}, fmt.Errorf("artifact unavailable: %w", err)
	}
	return Artifact{ArtifactRecord: record, Path: path}, nil
}

func (s *Service) cleanupArtifacts(ctx context.Context) error {
	paths, err := s.frontier.DeleteExpiredArtifacts(ctx, time.Now().UTC())
	if err != nil {
		return err
	}
	for _, name := range paths {
		if filepath.Base(name) == name {
			_ = os.Remove(filepath.Join(s.artifactDir, name))
		}
	}
	owned, err := s.frontier.ArtifactRelativePaths(ctx)
	if err != nil {
		return err
	}
	entries, err := os.ReadDir(s.artifactDir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		name := entry.Name()
		if entry.Type().IsRegular() && strings.HasPrefix(name, "artifact_") {
			if _, exists := owned[name]; !exists {
				_ = os.Remove(filepath.Join(s.artifactDir, name))
			}
		}
	}
	return nil
}
