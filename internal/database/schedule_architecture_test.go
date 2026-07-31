package database

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/seo-auditor/seo-auditor/internal/contracts"
)

func TestScheduledAuditLifecycle(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db, err := Open(ctx, filepath.Join(t.TempDir(), "schedule.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	frontier := NewFrontier(db, 8)
	defer frontier.Close()
	projectID := contracts.ID("project_schedule")
	profileID := contracts.ID("profile_schedule")
	if err := frontier.CreateProject(ctx, projectID, "Schedules"); err != nil {
		t.Fatal(err)
	}
	configuration := contracts.CrawlConfiguration{SeedURL: "https://example.com/", AllowedHosts: []string{"example.com"}, Limits: contracts.DefaultCrawlLimits()}
	if _, err := frontier.CreateProfile(ctx, profileID, projectID, "Daily", configuration); err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 7, 30, 12, 0, 0, 0, time.UTC)
	record := ScheduledAuditRecord{ID: "schedule_daily", ProjectID: projectID, ProfileID: profileID, Name: "Daily", IntervalSeconds: 86400, NextRunAt: now.Add(-time.Minute).Format(time.RFC3339Nano)}
	if err := frontier.CreateSchedule(ctx, record); err != nil {
		t.Fatal(err)
	}
	claimed, err := frontier.ClaimDueSchedules(ctx, now, 10)
	if err != nil || len(claimed) != 1 || claimed[0].ID != record.ID {
		t.Fatalf("claimed=%+v err=%v", claimed, err)
	}
	claimed, err = frontier.ClaimDueSchedules(ctx, now, 10)
	if err != nil || len(claimed) != 0 {
		t.Fatalf("schedule was claimed twice: %+v err=%v", claimed, err)
	}
	page, err := frontier.ListSchedules(ctx, projectID, contracts.PageRequest{Limit: 10})
	if err != nil || len(page.Items) != 1 || page.Items[0].NextRunAt <= record.NextRunAt {
		t.Fatalf("page=%+v err=%v", page, err)
	}
	if err := frontier.DeleteSchedule(ctx, projectID, record.ID); err != nil {
		t.Fatal(err)
	}
}

func TestArchitectureBuildsBoundedInternalGraph(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	db, err := Open(ctx, filepath.Join(t.TempDir(), "architecture.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	frontier := NewFrontier(db, 8)
	defer frontier.Close()
	stamp := "2026-07-30T12:00:00Z"
	statements := []string{
		`INSERT INTO project(id,name,created_at,updated_at) VALUES('project_arch','Architecture','` + stamp + `','` + stamp + `')`,
		`INSERT INTO crawl(id,project_id,seed_url,config_json,status,created_at,updated_at) VALUES('crawl_arch','project_arch','https://example.com/','{}','completed','` + stamp + `','` + stamp + `')`,
		`INSERT INTO url(id,project_id,request_key,original_url,scheme,host,path,created_at) VALUES(1,'project_arch','https://example.com/','https://example.com/','https','example.com','/','` + stamp + `')`,
		`INSERT INTO url(id,project_id,request_key,original_url,scheme,host,path,created_at) VALUES(2,'project_arch','https://example.com/guides/a','https://example.com/guides/a','https','example.com','/guides/a','` + stamp + `')`,
		`INSERT INTO crawl_url(id,crawl_id,url_id,state,depth,discovery_kind,created_at,updated_at) VALUES(1,'crawl_arch',1,'analysed',0,'seed','` + stamp + `','` + stamp + `')`,
		`INSERT INTO crawl_url(id,crawl_id,url_id,state,depth,discovery_kind,created_at,updated_at) VALUES(2,'crawl_arch',2,'analysed',1,'link','` + stamp + `','` + stamp + `')`,
		`INSERT INTO fetch_attempt(crawl_url_id,attempt,started_at,status_code) VALUES(1,1,'` + stamp + `',200)`,
		`INSERT INTO fetch_attempt(crawl_url_id,attempt,started_at,status_code) VALUES(2,1,'` + stamp + `',200)`,
		`INSERT INTO link(crawl_id,source_url_id,target_url_id,raw_target,link_kind,extraction_mode) VALUES('crawl_arch',1,2,'/guides/a','internal','raw')`,
	}
	for _, statement := range statements {
		if _, err := db.SQL().ExecContext(ctx, statement); err != nil {
			t.Fatal(err)
		}
	}
	graph, err := frontier.Architecture(ctx, "crawl_arch", 10, 10)
	if err != nil || len(graph.Nodes) != 2 || len(graph.Edges) != 1 {
		t.Fatalf("graph=%+v err=%v", graph, err)
	}
	if graph.Nodes[0].Segment != "/guides/" || graph.Nodes[0].Inlinks != 1 {
		t.Fatalf("unexpected ranked node: %+v", graph.Nodes[0])
	}
}
