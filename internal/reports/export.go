package reports

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/seo-auditor/seo-auditor/internal/contracts"
	"github.com/seo-auditor/seo-auditor/internal/database"
)

type QuerySource interface {
	ListPages(context.Context, contracts.ID, contracts.PageRequest) (contracts.Page[database.PageRecord], error)
	ListIssues(context.Context, contracts.ID, contracts.PageRequest) (contracts.Page[database.IssueRecord], error)
}

func PagesCSV(ctx context.Context, source QuerySource, crawlID contracts.ID, output io.Writer) error {
	writer := csv.NewWriter(output)
	if err := writer.Write([]string{"url", "status_code", "depth", "raw_title", "raw_meta_description", "raw_canonical", "raw_robots", "raw_language", "raw_text_length", "raw_content_hash", "raw_extraction_mode", "render_status", "rendered_title", "rendered_meta_description", "rendered_canonical", "rendered_content_hash"}); err != nil {
		return err
	}
	cursor := ""
	for {
		page, err := source.ListPages(ctx, crawlID, contracts.PageRequest{Cursor: cursor, Limit: 1000})
		if err != nil {
			return err
		}
		for _, item := range page.Items {
			row := []string{item.URL, fmt.Sprint(item.StatusCode), fmt.Sprint(item.Depth), item.Title, item.MetaDescription, item.CanonicalURL, item.RobotsDirectives, item.Language, fmt.Sprint(item.TextLength), item.ContentHash, item.ExtractionMode, item.RenderStatus, item.RenderedTitle, item.RenderedMetaDescription, item.RenderedCanonicalURL, item.RenderedContentHash}
			for index := range row {
				row[index] = spreadsheetSafe(row[index])
			}
			if err := writer.Write(row); err != nil {
				return err
			}
		}
		if page.NextCursor == "" {
			break
		}
		cursor = page.NextCursor
	}
	writer.Flush()
	return writer.Error()
}

func IssuesNDJSON(ctx context.Context, source QuerySource, crawlID contracts.ID, output io.Writer) error {
	encoder := json.NewEncoder(output)
	cursor := ""
	for {
		page, err := source.ListIssues(ctx, crawlID, contracts.PageRequest{Cursor: cursor, Limit: 1000})
		if err != nil {
			return err
		}
		for _, item := range page.Items {
			if err := encoder.Encode(item); err != nil {
				return err
			}
		}
		if page.NextCursor == "" {
			return nil
		}
		cursor = page.NextCursor
	}
}

func PagesNDJSON(ctx context.Context, source QuerySource, crawlID contracts.ID, output io.Writer) error {
	encoder := json.NewEncoder(output)
	cursor := ""
	for {
		page, err := source.ListPages(ctx, crawlID, contracts.PageRequest{Cursor: cursor, Limit: 1000})
		if err != nil {
			return err
		}
		for _, item := range page.Items {
			if err := encoder.Encode(item); err != nil {
				return err
			}
		}
		if page.NextCursor == "" {
			return nil
		}
		cursor = page.NextCursor
	}
}

func IssuesCSV(ctx context.Context, source QuerySource, crawlID contracts.ID, output io.Writer) error {
	writer := csv.NewWriter(output)
	if err := writer.Write([]string{"id", "rule_id", "rule_version", "subject_type", "subject_id", "severity", "classification", "evidence_source", "evidence_json", "created_at"}); err != nil {
		return err
	}
	cursor := ""
	for {
		page, err := source.ListIssues(ctx, crawlID, contracts.PageRequest{Cursor: cursor, Limit: 1000})
		if err != nil {
			return err
		}
		for _, item := range page.Items {
			row := []string{fmt.Sprint(item.ID), item.RuleID, fmt.Sprint(item.RuleVersion), item.SubjectType, item.SubjectID, item.Severity, item.Classification, item.EvidenceSource, item.EvidenceJSON, item.CreatedAt}
			for index := range row {
				row[index] = spreadsheetSafe(row[index])
			}
			if err := writer.Write(row); err != nil {
				return err
			}
		}
		if page.NextCursor == "" {
			break
		}
		cursor = page.NextCursor
	}
	writer.Flush()
	return writer.Error()
}

func spreadsheetSafe(value string) string {
	trimmed := strings.TrimLeft(value, "\t\r\n ")
	if trimmed != "" && strings.ContainsRune("=+-@", rune(trimmed[0])) {
		return "'" + value
	}
	return value
}
