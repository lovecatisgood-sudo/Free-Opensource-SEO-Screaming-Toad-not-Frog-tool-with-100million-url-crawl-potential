package application

import (
	"context"
	"testing"

	"github.com/seo-auditor/seo-auditor/internal/contracts"
)

func TestProjectsProfilesAndScopePreviewRoundTrip(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	service, err := Open(ctx, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer service.Close()
	project, err := service.CreateProject(ctx, "Example audit")
	if err != nil {
		t.Fatal(err)
	}
	configuration := contracts.CrawlConfiguration{SeedURL: "https://example.com/", AllowedHosts: []string{"example.com"}, ExcludePathRegex: []string{`^/private`}, Limits: contracts.DefaultCrawlLimits()}
	profile, err := service.CreateProfile(ctx, project.ID, "Default", configuration)
	if err != nil {
		t.Fatal(err)
	}
	if profile.Version != 1 || profile.Configuration.UserAgent == "" {
		t.Fatalf("profile=%+v", profile)
	}
	profiles, err := service.ListProfiles(ctx, project.ID, contracts.PageRequest{Limit: 10})
	if err != nil || len(profiles.Items) != 1 {
		t.Fatalf("profiles=%+v err=%v", profiles, err)
	}
	decisions, err := service.PreviewScope(ctx, profile.Configuration, []string{"https://example.com/ok", "https://example.com/private/secret", "https://other.example/"})
	if err != nil {
		t.Fatal(err)
	}
	if !decisions[0].Allowed || decisions[1].Allowed || decisions[2].Allowed {
		t.Fatalf("decisions=%+v", decisions)
	}
	if err := service.ArchiveProject(ctx, project.ID, true); err != nil {
		t.Fatal(err)
	}
	projects, err := service.ListProjects(ctx, contracts.PageRequest{})
	if err != nil || len(projects.Items) != 1 || !projects.Items[0].Archived {
		t.Fatalf("projects=%+v err=%v", projects, err)
	}
}
