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

type RenderMetadata struct {
	Status           string
	ErrorCode        string
	FinalURL         string
	RequestCount     int
	TransferredBytes int64
}

// SaveRenderFailure preserves renderer coverage even when raw extraction
// remains the audit fallback.
func (f *Frontier) SaveRenderFailure(ctx context.Context, lease Lease, metadata RenderMetadata) error {
	if metadata.Status != "blocked" && metadata.Status != "failed" {
		return fmt.Errorf("render failure status is invalid")
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	return f.writer.Submit(ctx, func(ctx context.Context, tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, `INSERT INTO rendered_page(crawl_url_id,status,error_code,final_url,request_count,transferred_bytes,rendered_at)
			VALUES (?,?,?,?,?,?,?)
			ON CONFLICT(crawl_url_id) DO UPDATE SET status=excluded.status,error_code=excluded.error_code,final_url=excluded.final_url,request_count=excluded.request_count,transferred_bytes=excluded.transferred_bytes,rendered_at=excluded.rendered_at`,
			lease.CrawlURLID, metadata.Status, metadata.ErrorCode, metadata.FinalURL, metadata.RequestCount, metadata.TransferredBytes, now)
		return err
	})
}

func (f *Frontier) SaveRenderedAnalysis(ctx context.Context, crawlID, projectID contracts.ID, lease Lease, raw, rendered extractor.Page, issues []rules.Issue, metadata RenderMetadata) error {
	if metadata.Status != "completed" {
		return fmt.Errorf("completed render metadata is required")
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	return f.writer.Submit(ctx, func(ctx context.Context, tx *sql.Tx) error {
		canonical := firstCanonical(rendered)
		headings, _ := json.Marshal(rendered.Headings)
		images, _ := json.Marshal(rendered.Images)
		hreflang, _ := json.Marshal(rendered.Hreflangs)
		structured, _ := json.Marshal(rendered.StructuredData)
		social, _ := json.Marshal(rendered.Social)
		_, err := tx.ExecContext(ctx, `INSERT INTO rendered_page(crawl_url_id,status,error_code,final_url,request_count,transferred_bytes,title,meta_description,canonical_url,robots_directives,language,text_length,content_hash,html_hash,headings_json,images_json,hreflang_json,structured_data_json,social_json,rendered_at)
			VALUES (?,'completed','',?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)
			ON CONFLICT(crawl_url_id) DO UPDATE SET status='completed',error_code='',final_url=excluded.final_url,request_count=excluded.request_count,transferred_bytes=excluded.transferred_bytes,title=excluded.title,meta_description=excluded.meta_description,canonical_url=excluded.canonical_url,robots_directives=excluded.robots_directives,language=excluded.language,text_length=excluded.text_length,content_hash=excluded.content_hash,html_hash=excluded.html_hash,headings_json=excluded.headings_json,images_json=excluded.images_json,hreflang_json=excluded.hreflang_json,structured_data_json=excluded.structured_data_json,social_json=excluded.social_json,rendered_at=excluded.rendered_at`,
			lease.CrawlURLID, metadata.FinalURL, metadata.RequestCount, metadata.TransferredBytes,
			rendered.Title, rendered.MetaDescription, canonical, rendered.MetaRobots, rendered.Language,
			len(rendered.VisibleText), rendered.ContentHash, rendered.HTMLHash, string(headings), string(images), string(hreflang), string(structured), string(social), now)
		if err != nil {
			return err
		}
		var renderedPageID int64
		if err := tx.QueryRowContext(ctx, "SELECT id FROM rendered_page WHERE crawl_url_id=?", lease.CrawlURLID).Scan(&renderedPageID); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, "DELETE FROM render_difference WHERE crawl_url_id=?", lease.CrawlURLID); err != nil {
			return err
		}
		for field, values := range renderDifferences(raw, rendered) {
			if values[0] == values[1] {
				continue
			}
			if _, err := tx.ExecContext(ctx, "INSERT INTO render_difference(crawl_url_id,field,raw_value,rendered_value) VALUES (?,?,?,?)", lease.CrawlURLID, field, values[0], values[1]); err != nil {
				return err
			}
		}
		for _, link := range rendered.Links {
			targetID, err := ensureURL(ctx, tx, projectID, link.URL, now)
			if err != nil {
				continue
			}
			kind := "external"
			if target, parseErr := fetchpolicy.NormalizeURL(link.URL); parseErr == nil && target.URL.Hostname() == mustHost(rendered.URL) {
				kind = "internal"
			}
			if _, err := tx.ExecContext(ctx, `INSERT INTO link(crawl_id,source_url_id,target_url_id,raw_target,anchor_text,rel,link_kind,extraction_mode) VALUES (?,?,?,?,?,?,?,'rendered')`, crawlID, lease.URLID, targetID, link.RawURL, link.Text, link.Rel, kind); err != nil {
				return err
			}
		}
		for _, issue := range issues {
			if issue.Evidence == nil {
				issue.Evidence = map[string]any{}
			}
			issue.Evidence["extraction_mode"] = "rendered"
			evidence, _ := json.Marshal(issue.Evidence)
			if _, err := tx.ExecContext(ctx, `INSERT INTO issue(crawl_id,rule_id,rule_version,subject_type,subject_id,severity,evidence_json,created_at) VALUES (?,?,?,'rendered_page',?,?,?,?)`, crawlID, issue.RuleID, issue.RuleVersion, fmt.Sprint(renderedPageID), issue.Severity, string(evidence), now); err != nil {
				return err
			}
		}
		return nil
	})
}

func firstCanonical(page extractor.Page) any {
	if len(page.Canonicals) == 0 {
		return nil
	}
	if normalized, err := fetchpolicy.NormalizeURL(page.Canonicals[0]); err == nil {
		return normalized.RequestKey
	}
	return page.Canonicals[0]
}

func renderDifferences(raw, rendered extractor.Page) map[string][2]string {
	return map[string][2]string{
		"title":            {raw.Title, rendered.Title},
		"meta_description": {raw.MetaDescription, rendered.MetaDescription},
		"canonical":        {canonicalString(raw), canonicalString(rendered)},
		"robots":           {raw.MetaRobots, rendered.MetaRobots},
		"language":         {raw.Language, rendered.Language},
		"content_hash":     {raw.ContentHash, rendered.ContentHash},
		"text_length":      {fmt.Sprint(len(raw.VisibleText)), fmt.Sprint(len(rendered.VisibleText))},
	}
}

func canonicalString(page extractor.Page) string {
	value := firstCanonical(page)
	if value == nil {
		return ""
	}
	return fmt.Sprint(value)
}
