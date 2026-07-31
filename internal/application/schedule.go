package application

import (
	"context"
	"errors"
	"time"

	"github.com/seo-auditor/seo-auditor/internal/contracts"
	"github.com/seo-auditor/seo-auditor/internal/database"
)

func (s *Service) CreateSchedule(ctx context.Context, projectID, profileID contracts.ID, name string, intervalSeconds int64, firstRunAt string) (database.ScheduledAuditRecord, error) {
	if intervalSeconds < 900 || intervalSeconds > 2592000 {
		return database.ScheduledAuditRecord{}, errors.New("schedule interval must be between 15 minutes and 30 days")
	}
	next := time.Now().UTC().Add(time.Duration(intervalSeconds) * time.Second)
	if firstRunAt != "" {
		parsed, err := time.Parse(time.RFC3339, firstRunAt)
		if err != nil || parsed.Before(time.Now().Add(-time.Minute)) {
			return database.ScheduledAuditRecord{}, errors.New("first run time must be a valid future RFC3339 timestamp")
		}
		next = parsed
	}
	id, err := contracts.NewID("schedule")
	if err != nil {
		return database.ScheduledAuditRecord{}, err
	}
	record := database.ScheduledAuditRecord{ID: id, ProjectID: projectID, ProfileID: profileID, Name: name, IntervalSeconds: intervalSeconds, Enabled: true, NextRunAt: next.UTC().Format(time.RFC3339Nano)}
	if err := s.frontier.CreateSchedule(ctx, record); err != nil {
		return database.ScheduledAuditRecord{}, err
	}
	page, err := s.frontier.ListSchedules(ctx, projectID, contracts.PageRequest{Limit: 100})
	if err != nil {
		return database.ScheduledAuditRecord{}, err
	}
	for _, item := range page.Items {
		if item.ID == id {
			return item, nil
		}
	}
	return database.ScheduledAuditRecord{}, errors.New("scheduled audit was not found after creation")
}
func (s *Service) ListSchedules(ctx context.Context, projectID contracts.ID, page contracts.PageRequest) (contracts.Page[database.ScheduledAuditRecord], error) {
	return s.frontier.ListSchedules(ctx, projectID, page)
}
func (s *Service) DeleteSchedule(ctx context.Context, projectID, id contracts.ID) error {
	return s.frontier.DeleteSchedule(ctx, projectID, id)
}
func (s *Service) RunDueSchedules(ctx context.Context, now time.Time) error {
	items, err := s.frontier.ClaimDueSchedules(ctx, now, 20)
	if err != nil {
		return err
	}
	for _, item := range items {
		result, startErr := s.StartProfileCrawl(ctx, item.ProjectID, item.ProfileID)
		message := ""
		crawlID := result.CrawlID
		if startErr != nil {
			message = startErr.Error()
		}
		if err := s.frontier.RecordScheduleResult(context.Background(), item.ID, crawlID, message); err != nil {
			return err
		}
	}
	return nil
}
func (s *Service) scheduleLoop() {
	defer s.runs.Done()
	timer := time.NewTimer(30 * time.Second)
	defer timer.Stop()
	for {
		select {
		case <-s.ctx.Done():
			return
		case now := <-timer.C:
			_ = s.RunDueSchedules(s.ctx, now)
			timer.Reset(time.Minute)
		}
	}
}
