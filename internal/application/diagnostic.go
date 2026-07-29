package application

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/seo-auditor/seo-auditor/internal/contracts"
	"github.com/seo-auditor/seo-auditor/internal/database"
	"github.com/seo-auditor/seo-auditor/internal/version"
)

type diagnosticBundle struct {
	GeneratedAt       string                      `json:"generated_at"`
	Version           string                      `json:"version"`
	Commit            string                      `json:"commit"`
	CrawlID           contracts.ID                `json:"crawl_id"`
	Summary           database.AuditSummary       `json:"summary"`
	RenderingMode     string                      `json:"rendering_mode"`
	Limits            contracts.CrawlLimits       `json:"limits"`
	AllowedHostCount  int                         `json:"allowed_host_count"`
	DatabaseIntegrity string                      `json:"database_integrity"`
	Events            []database.CrawlEventRecord `json:"events"`
	ContentExcluded   bool                        `json:"crawled_content_excluded"`
}

// Diagnostic creates a managed metadata-only artifact. It intentionally
// excludes URLs, page content, headers, cookies, and extracted evidence.
func (s *Service) Diagnostic(ctx context.Context, crawlID contracts.ID) (Artifact, error) {
	if crawlID == "" {
		return Artifact{}, errors.New("crawl ID is required")
	}
	stored, err := s.frontier.LoadCrawl(ctx, crawlID)
	if err != nil {
		return Artifact{}, err
	}
	summary, err := s.frontier.Summary(ctx, crawlID)
	if err != nil {
		return Artifact{}, err
	}
	events := make([]database.CrawlEventRecord, 0, 128)
	cursor := ""
	for len(events) < 10_000 {
		page, err := s.frontier.ListEvents(ctx, crawlID, contracts.PageRequest{Cursor: cursor, Limit: 1000})
		if err != nil {
			return Artifact{}, err
		}
		events = append(events, page.Items...)
		if page.NextCursor == "" {
			break
		}
		cursor = page.NextCursor
	}
	integrity := "ok"
	if err := s.db.Verify(ctx); err != nil {
		integrity = "failed"
	}
	bundle := diagnosticBundle{
		GeneratedAt: time.Now().UTC().Format(time.RFC3339Nano), Version: version.Version, Commit: version.Commit,
		CrawlID: crawlID, Summary: summary, RenderingMode: stored.Configuration.RenderingMode,
		Limits: stored.Configuration.Limits, AllowedHostCount: len(stored.Configuration.AllowedHosts),
		DatabaseIntegrity: integrity, Events: events, ContentExcluded: true,
	}
	return s.writeManagedArtifact(ctx, crawlID, "diagnostic-json", "diagnostic.json", 7*24*time.Hour, func(writer io.Writer) error {
		encoder := json.NewEncoder(writer)
		encoder.SetIndent("", "  ")
		return encoder.Encode(bundle)
	})
}

func (s *Service) writeManagedArtifact(ctx context.Context, crawlID contracts.ID, format, extension string, ttl time.Duration, write func(io.Writer) error) (Artifact, error) {
	id, err := contracts.NewID("artifact")
	if err != nil {
		return Artifact{}, err
	}
	temporary, err := os.CreateTemp(s.artifactDir, ".artifact-*")
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
	if err := write(io.MultiWriter(temporary, hash)); err != nil {
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
		_ = os.Remove(destination)
		return Artifact{}, err
	}
	now := time.Now().UTC()
	record := database.ArtifactRecord{ID: id, CrawlID: crawlID, Format: format, RelativePath: name, Checksum: hex.EncodeToString(hash.Sum(nil)), SizeBytes: info.Size(), CreatedAt: now.Format(time.RFC3339Nano), ExpiresAt: now.Add(ttl).Format(time.RFC3339Nano)}
	if err := s.frontier.RecordArtifact(ctx, record); err != nil {
		_ = os.Remove(destination)
		return Artifact{}, err
	}
	committed = true
	return Artifact{ArtifactRecord: record, Path: destination}, nil
}
