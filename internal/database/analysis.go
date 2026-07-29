package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/seo-auditor/seo-auditor/internal/contracts"
	"github.com/seo-auditor/seo-auditor/internal/extractor"
	"github.com/seo-auditor/seo-auditor/internal/fetchpolicy"
	"github.com/seo-auditor/seo-auditor/internal/rules"
)

func (f *Frontier) SaveAnalysis(ctx context.Context, crawlID, projectID contracts.ID, lease Lease, page extractor.Page, issues []rules.Issue) error {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	return f.writer.Submit(ctx, func(ctx context.Context, tx *sql.Tx) error {
		var canonical any
		if len(page.Canonicals) > 0 {
			canonical = page.Canonicals[0]
			if normalized, err := fetchpolicy.NormalizeURL(page.Canonicals[0]); err == nil {
				canonical = normalized.RequestKey
			}
		}
		social, _ := json.Marshal(page.Social)
		result, err := tx.ExecContext(ctx, `INSERT INTO page(crawl_url_id, extraction_mode, title, meta_description, canonical_url, robots_directives, language, text_length, content_hash, extracted_at, viewport, html_hash, x_robots_tag, social_json)
            VALUES (?, 'raw', ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, lease.CrawlURLID, page.Title, page.MetaDescription, canonical,
			page.MetaRobots, page.Language, len(page.VisibleText), page.ContentHash, now, page.Viewport, page.HTMLHash, page.XRobotsTag, string(social))
		if err != nil {
			return err
		}
		pageID, err := result.LastInsertId()
		if err != nil {
			return err
		}
		for position, heading := range page.Headings {
			if _, err := tx.ExecContext(ctx, "INSERT INTO heading(page_id, position, level, text) VALUES (?, ?, ?, ?)", pageID, position, heading.Level, heading.Text); err != nil {
				return err
			}
		}
		for _, link := range page.Links {
			targetID, err := ensureURL(ctx, tx, projectID, link.URL, now)
			if err != nil {
				continue
			}
			kind := "external"
			if target, parseErr := fetchpolicy.NormalizeURL(link.URL); parseErr == nil && target.URL.Hostname() == mustHost(page.URL) {
				kind = "internal"
			}
			if _, err := tx.ExecContext(ctx, `INSERT INTO link(crawl_id, source_url_id, target_url_id, raw_target, anchor_text, rel, link_kind, extraction_mode)
                    VALUES (?, ?, ?, ?, ?, ?, ?, 'raw')`, crawlID, lease.URLID, targetID, link.RawURL, link.Text, link.Rel, kind); err != nil {
				return err
			}
		}
		for position, image := range page.Images {
			normalized, err := fetchpolicy.NormalizeURL(image.URL)
			if err != nil {
				continue
			}
			if _, err := tx.ExecContext(ctx, `INSERT INTO image(project_id, request_key, original_url, declared_width, declared_height) VALUES (?, ?, ?, ?, ?)
                    ON CONFLICT(project_id, request_key) DO UPDATE SET declared_width = excluded.declared_width, declared_height = excluded.declared_height`, projectID, normalized.RequestKey, image.URL, image.Width, image.Height); err != nil {
				return err
			}
			var imageID int64
			if err := tx.QueryRowContext(ctx, "SELECT id FROM image WHERE project_id = ? AND request_key = ?", projectID, normalized.RequestKey).Scan(&imageID); err != nil {
				return err
			}
			var alt any
			if image.AltPresent {
				alt = image.Alt
			}
			if _, err := tx.ExecContext(ctx, "INSERT INTO page_image(page_id, image_id, position, alt_text) VALUES (?, ?, ?, ?)", pageID, imageID, position, alt); err != nil {
				return err
			}
		}
		for _, item := range page.Hreflangs {
			if _, err := tx.ExecContext(ctx, "INSERT INTO hreflang(page_id, language_code, target_url, validation_state) VALUES (?, ?, ?, 'page_only')", pageID, item.Language, item.URL); err != nil {
				return err
			}
		}
		for _, item := range page.StructuredData {
			evidence, _ := json.Marshal(item)
			types, _ := json.Marshal(item.Types)
			if _, err := tx.ExecContext(ctx, "INSERT INTO structured_data(page_id, format, type_summary, evidence_json) VALUES (?, ?, ?, ?)", pageID, item.Format, string(types), string(evidence)); err != nil {
				return err
			}
		}
		for _, issue := range issues {
			evidence, _ := json.Marshal(issue.Evidence)
			if _, err := tx.ExecContext(ctx, `INSERT INTO issue(crawl_id, rule_id, rule_version, subject_type, subject_id, severity, evidence_json, created_at) VALUES (?, ?, ?, 'page', ?, ?, ?, ?)`, crawlID, issue.RuleID, issue.RuleVersion, fmt.Sprint(pageID), issue.Severity, string(evidence), now); err != nil {
				return err
			}
		}
		updated, err := tx.ExecContext(ctx, "UPDATE crawl_url SET state = 'analysed', updated_at = ? WHERE id = ? AND state = 'fetched'", now, lease.CrawlURLID)
		if err != nil {
			return err
		}
		if rows, _ := updated.RowsAffected(); rows != 1 {
			return fmt.Errorf("crawl URL is not fetched")
		}
		_, err = tx.ExecContext(ctx, "UPDATE crawl SET analysed_count = analysed_count + 1, updated_at = ? WHERE id = ?", now, crawlID)
		return err
	})
}

func ensureURL(ctx context.Context, tx *sql.Tx, projectID contracts.ID, raw, now string) (int64, error) {
	normalized, err := fetchpolicy.NormalizeURL(raw)
	if err != nil {
		return 0, err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO url(project_id, request_key, original_url, scheme, host, port, path, query, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?) ON CONFLICT(project_id, request_key) DO NOTHING`, projectID, normalized.RequestKey, raw, normalized.URL.Scheme, normalized.URL.Hostname(), normalized.URL.Port(), normalized.URL.EscapedPath(), normalized.URL.RawQuery, now)
	if err != nil {
		return 0, err
	}
	var id int64
	err = tx.QueryRowContext(ctx, "SELECT id FROM url WHERE project_id = ? AND request_key = ?", projectID, normalized.RequestKey).Scan(&id)
	return id, err
}

func mustHost(raw string) string {
	normalized, err := fetchpolicy.NormalizeURL(raw)
	if err != nil {
		return ""
	}
	return normalized.URL.Hostname()
}
