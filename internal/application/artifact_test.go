package application

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/seo-auditor/seo-auditor/internal/contracts"
	"github.com/seo-auditor/seo-auditor/internal/fetchpolicy"
)

func TestManagedExportAndBackupStayInsideArtifactDirectory(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	service, err := Open(ctx, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer service.Close()
	projectID := contracts.ID("project_artifact")
	crawlID := contracts.ID("crawl_artifact")
	if err := service.frontier.CreateProject(ctx, projectID, "Artifacts"); err != nil {
		t.Fatal(err)
	}
	seed, _ := fetchpolicy.NormalizeURL("https://example.com/")
	configuration := contracts.CrawlConfiguration{SeedURL: seed.RequestKey, AllowedHosts: []string{seed.URL.Hostname()}, UserAgent: "test", RenderingMode: "raw", Limits: contracts.DefaultCrawlLimits()}
	if err := service.frontier.CreateCrawl(ctx, crawlID, projectID, "", seed, configuration); err != nil {
		t.Fatal(err)
	}
	artifact, err := service.Export(ctx, ExportRequest{CrawlID: crawlID, Dataset: "pages", Format: "csv"})
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Dir(artifact.Path) != service.artifactDir || len(artifact.Checksum) != 64 {
		t.Fatalf("artifact=%+v", artifact)
	}
	if info, err := os.Stat(artifact.Path); err != nil || info.Size() == 0 {
		t.Fatalf("info=%v err=%v", info, err)
	}
	loaded, err := service.Artifact(ctx, artifact.ID)
	if err != nil || loaded.Path != artifact.Path {
		t.Fatalf("loaded=%+v err=%v", loaded, err)
	}
	backup, err := service.Backup(ctx, crawlID)
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Ext(backup.Path) != ".sqlite3" {
		t.Fatalf("backup=%+v", backup)
	}
	diagnostic, err := service.Diagnostic(ctx, crawlID)
	if err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(diagnostic.Path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), `"crawled_content_excluded": true`) || strings.Contains(string(body), "example.com") {
		t.Fatalf("diagnostic content was incomplete or leaked a URL: %s", body)
	}
}

func TestArtifactCleanupRemovesOnlyExpiredAndOrphanedManagedFiles(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	service, err := Open(ctx, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer service.Close()
	orphan := filepath.Join(service.artifactDir, "artifact_00000000000000000000000000000000.csv")
	unmanaged := filepath.Join(service.artifactDir, "operator-note.txt")
	if err := os.WriteFile(orphan, []byte("orphan"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(unmanaged, []byte("keep"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := service.cleanupArtifacts(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(orphan); !os.IsNotExist(err) {
		t.Fatalf("orphaned managed artifact remains: %v", err)
	}
	if _, err := os.Stat(unmanaged); err != nil {
		t.Fatalf("unmanaged file was removed: %v", err)
	}
}
