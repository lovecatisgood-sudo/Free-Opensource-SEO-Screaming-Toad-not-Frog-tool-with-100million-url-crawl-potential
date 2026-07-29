package application

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/seo-auditor/seo-auditor/internal/contracts"
	"github.com/seo-auditor/seo-auditor/internal/database"
)

func (s *Service) Backup(ctx context.Context, crawlID contracts.ID) (Artifact, error) {
	id, err := contracts.NewID("artifact")
	if err != nil {
		return Artifact{}, err
	}
	name := string(id) + ".sqlite3"
	path := filepath.Join(s.artifactDir, name)
	if err := s.db.Verify(ctx); err != nil {
		return Artifact{}, err
	}
	if err := s.db.Backup(ctx, path); err != nil {
		return Artifact{}, err
	}
	file, err := os.Open(path)
	if err != nil {
		_ = os.Remove(path)
		return Artifact{}, err
	}
	hash := sha256.New()
	size, err := io.Copy(hash, file)
	closeErr := file.Close()
	if err != nil {
		_ = os.Remove(path)
		return Artifact{}, err
	}
	if closeErr != nil {
		_ = os.Remove(path)
		return Artifact{}, closeErr
	}
	now := time.Now().UTC()
	record := database.ArtifactRecord{ID: id, CrawlID: crawlID, Format: "sqlite3-backup", RelativePath: name, Checksum: hex.EncodeToString(hash.Sum(nil)), SizeBytes: size, CreatedAt: now.Format(time.RFC3339Nano), ExpiresAt: now.Add(30 * 24 * time.Hour).Format(time.RFC3339Nano)}
	if err := s.frontier.RecordArtifact(ctx, record); err != nil {
		_ = os.Remove(path)
		return Artifact{}, err
	}
	return Artifact{ArtifactRecord: record, Path: path}, nil
}

func (s *Service) TrashCrawl(ctx context.Context, id contracts.ID) error {
	return s.frontier.TrashCrawl(ctx, id)
}
func (s *Service) RestoreCrawl(ctx context.Context, id contracts.ID) error {
	return s.frontier.RestoreCrawl(ctx, id)
}
