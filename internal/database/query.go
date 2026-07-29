package database

import (
	"context"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/seo-auditor/seo-auditor/internal/contracts"
)

type AuditSummary struct {
	CrawlID           contracts.ID          `json:"crawl_id"`
	Status            contracts.CrawlStatus `json:"status"`
	Discovered        int64                 `json:"discovered"`
	Fetched           int64                 `json:"fetched"`
	Analysed          int64                 `json:"analysed"`
	Failed            int64                 `json:"failed"`
	IssuesBySeverity  map[string]int64      `json:"issues_by_severity"`
	ResponsesByClass  map[string]int64      `json:"responses_by_class"`
	RenderingByStatus map[string]int64      `json:"rendering_by_status,omitempty"`
}

type IssueRecord struct {
	ID           int64  `json:"id"`
	RuleID       string `json:"rule_id"`
	RuleVersion  int    `json:"rule_version"`
	SubjectType  string `json:"subject_type"`
	SubjectID    string `json:"subject_id"`
	Severity     string `json:"severity"`
	EvidenceJSON string `json:"evidence_json"`
	CreatedAt    string `json:"created_at"`
}
type PageRecord struct {
	ID                      int64  `json:"id"`
	URL                     string `json:"url"`
	Title                   string `json:"title"`
	MetaDescription         string `json:"meta_description"`
	CanonicalURL            string `json:"canonical_url"`
	RobotsDirectives        string `json:"robots_directives"`
	Language                string `json:"language"`
	ContentHash             string `json:"content_hash"`
	StatusCode              int    `json:"status_code"`
	Depth                   int    `json:"depth"`
	TextLength              int    `json:"text_length"`
	ExtractionMode          string `json:"extraction_mode"`
	RenderStatus            string `json:"render_status,omitempty"`
	RenderedTitle           string `json:"rendered_title,omitempty"`
	RenderedMetaDescription string `json:"rendered_meta_description,omitempty"`
	RenderedCanonicalURL    string `json:"rendered_canonical_url,omitempty"`
	RenderedContentHash     string `json:"rendered_content_hash,omitempty"`
}

func (f *Frontier) Summary(ctx context.Context, crawlID contracts.ID) (AuditSummary, error) {
	summary := AuditSummary{CrawlID: crawlID, IssuesBySeverity: map[string]int64{}, ResponsesByClass: map[string]int64{}, RenderingByStatus: map[string]int64{}}
	if err := f.db.QueryRowContext(ctx, `SELECT status, discovered_count, fetched_count, analysed_count, failed_count FROM crawl WHERE id = ?`, crawlID).Scan(&summary.Status, &summary.Discovered, &summary.Fetched, &summary.Analysed, &summary.Failed); err != nil {
		return summary, err
	}
	rows, err := f.db.QueryContext(ctx, "SELECT severity, count(*) FROM issue WHERE crawl_id = ? GROUP BY severity", crawlID)
	if err != nil {
		return summary, err
	}
	for rows.Next() {
		var key string
		var count int64
		if err := rows.Scan(&key, &count); err != nil {
			rows.Close()
			return summary, err
		}
		summary.IssuesBySeverity[key] = count
	}
	if err := rows.Close(); err != nil {
		return summary, err
	}
	rows, err = f.db.QueryContext(ctx, `SELECT CASE WHEN status_code IS NULL THEN 'none' ELSE printf('%dxx', status_code / 100) END, count(*) FROM fetch_attempt WHERE crawl_url_id IN (SELECT id FROM crawl_url WHERE crawl_id = ?) AND finished_at IS NOT NULL GROUP BY 1`, crawlID)
	if err != nil {
		return summary, err
	}
	defer rows.Close()
	for rows.Next() {
		var key string
		var count int64
		if err := rows.Scan(&key, &count); err != nil {
			return summary, err
		}
		summary.ResponsesByClass[key] = count
	}
	if err := rows.Err(); err != nil {
		return summary, err
	}
	if err := rows.Close(); err != nil {
		return summary, err
	}
	renderRows, err := f.db.QueryContext(ctx, `SELECT status,count(*) FROM rendered_page WHERE crawl_url_id IN (SELECT id FROM crawl_url WHERE crawl_id=?) GROUP BY status`, crawlID)
	if err != nil {
		return summary, err
	}
	defer renderRows.Close()
	for renderRows.Next() {
		var status string
		var count int64
		if err := renderRows.Scan(&status, &count); err != nil {
			return summary, err
		}
		summary.RenderingByStatus[status] = count
	}
	return summary, renderRows.Err()
}

func (f *Frontier) ListIssues(ctx context.Context, crawlID contracts.ID, page contracts.PageRequest) (contracts.Page[IssueRecord], error) {
	if err := validateQuery(page, true); err != nil {
		return contracts.Page[IssueRecord]{}, err
	}
	search := likeSearch(page.Search)
	return listKeyset(ctx, page, func(after int64, limit int) (*sql.Rows, error) {
		return f.db.QueryContext(ctx, `SELECT id, rule_id, rule_version, subject_type, subject_id, severity, evidence_json, created_at FROM issue WHERE crawl_id = ? AND id > ? AND (?='' OR severity=?) AND (?='' OR rule_id=?) AND (?='' OR evidence_json LIKE ? ESCAPE '\') ORDER BY id LIMIT ?`, crawlID, after, page.Severity, page.Severity, page.RuleID, page.RuleID, page.Search, search, limit)
	}, func(rows *sql.Rows) (IssueRecord, error) {
		var item IssueRecord
		err := rows.Scan(&item.ID, &item.RuleID, &item.RuleVersion, &item.SubjectType, &item.SubjectID, &item.Severity, &item.EvidenceJSON, &item.CreatedAt)
		return item, err
	}, func(item IssueRecord) int64 { return item.ID })
}
func (f *Frontier) GetIssue(ctx context.Context, crawlID contracts.ID, id int64) (IssueRecord, error) {
	var item IssueRecord
	err := f.db.QueryRowContext(ctx, `SELECT id,rule_id,rule_version,subject_type,subject_id,severity,evidence_json,created_at FROM issue WHERE crawl_id=? AND id=?`, crawlID, id).Scan(&item.ID, &item.RuleID, &item.RuleVersion, &item.SubjectType, &item.SubjectID, &item.Severity, &item.EvidenceJSON, &item.CreatedAt)
	return item, err
}

func (f *Frontier) ListPages(ctx context.Context, crawlID contracts.ID, page contracts.PageRequest) (contracts.Page[PageRecord], error) {
	if err := validateQuery(page, false); err != nil {
		return contracts.Page[PageRecord]{}, err
	}
	search := likeSearch(page.Search)
	return listKeyset(ctx, page, func(after int64, limit int) (*sql.Rows, error) {
		return f.db.QueryContext(ctx, `SELECT p.id, u.request_key, COALESCE(fa.status_code, 0), cu.depth, COALESCE(p.title,''), COALESCE(p.meta_description,''), COALESCE(p.canonical_url,''), COALESCE(p.robots_directives,''), COALESCE(p.language,''), p.text_length, COALESCE(p.content_hash,''),p.extraction_mode,COALESCE(rp.status,''),COALESCE(rp.title,''),COALESCE(rp.meta_description,''),COALESCE(rp.canonical_url,''),COALESCE(rp.content_hash,'') FROM page p JOIN crawl_url cu ON cu.id = p.crawl_url_id JOIN url u ON u.id = cu.url_id LEFT JOIN fetch_attempt fa ON fa.crawl_url_id = cu.id AND fa.attempt = cu.attempt_count LEFT JOIN rendered_page rp ON rp.crawl_url_id=cu.id WHERE cu.crawl_id = ? AND p.id > ? AND (?='' OR u.request_key LIKE ? ESCAPE '\' OR COALESCE(p.title,'') LIKE ? ESCAPE '\' OR COALESCE(rp.title,'') LIKE ? ESCAPE '\') ORDER BY p.id LIMIT ?`, crawlID, after, page.Search, search, search, search, limit)
	}, func(rows *sql.Rows) (PageRecord, error) {
		var item PageRecord
		err := rows.Scan(&item.ID, &item.URL, &item.StatusCode, &item.Depth, &item.Title, &item.MetaDescription, &item.CanonicalURL, &item.RobotsDirectives, &item.Language, &item.TextLength, &item.ContentHash, &item.ExtractionMode, &item.RenderStatus, &item.RenderedTitle, &item.RenderedMetaDescription, &item.RenderedCanonicalURL, &item.RenderedContentHash)
		return item, err
	}, func(item PageRecord) int64 { return item.ID })
}

func validateQuery(page contracts.PageRequest, issues bool) error {
	if page.Sort != "" && page.Sort != "id" {
		return errors.New("sort must be id")
	}
	if len(page.Search) > 500 || len(page.RuleID) > 50 {
		return errors.New("filter is too long")
	}
	if !issues && (page.Severity != "" || page.RuleID != "") {
		return errors.New("issue filters are not valid for pages")
	}
	if page.Severity != "" && page.Severity != "info" && page.Severity != "warning" && page.Severity != "error" {
		return errors.New("severity is invalid")
	}
	return nil
}
func likeSearch(value string) string {
	value = strings.ReplaceAll(value, `\`, `\\`)
	value = strings.ReplaceAll(value, "%", `\%`)
	value = strings.ReplaceAll(value, "_", `\_`)
	return "%" + value + "%"
}

func listKeyset[T any](ctx context.Context, request contracts.PageRequest, query func(int64, int) (*sql.Rows, error), scan func(*sql.Rows) (T, error), id func(T) int64) (contracts.Page[T], error) {
	_ = ctx
	after, err := decodeCursor(request.Cursor)
	if err != nil {
		return contracts.Page[T]{}, err
	}
	limit := request.BoundedLimit()
	rows, err := query(after, limit+1)
	if err != nil {
		return contracts.Page[T]{}, err
	}
	defer rows.Close()
	result := contracts.Page[T]{Items: make([]T, 0, limit)}
	for rows.Next() {
		item, err := scan(rows)
		if err != nil {
			return result, err
		}
		if len(result.Items) == limit {
			result.NextCursor = encodeCursor(id(result.Items[len(result.Items)-1]))
			break
		}
		result.Items = append(result.Items, item)
	}
	return result, rows.Err()
}

func encodeCursor(id int64) string {
	return base64.RawURLEncoding.EncodeToString([]byte(strconv.FormatInt(id, 10)))
}
func decodeCursor(cursor string) (int64, error) {
	if cursor == "" {
		return 0, nil
	}
	decoded, err := base64.RawURLEncoding.DecodeString(cursor)
	if err != nil {
		return 0, errors.New("invalid cursor")
	}
	value, err := strconv.ParseInt(string(decoded), 10, 64)
	if err != nil || value < 0 {
		return 0, fmt.Errorf("invalid cursor")
	}
	return value, nil
}

func (f *Frontier) ListCrawls(ctx context.Context, projectID contracts.ID, page contracts.PageRequest) (contracts.Page[contracts.CrawlProgress], error) {
	limit := page.BoundedLimit()
	after, err := decodeCursor(page.Cursor)
	if err != nil {
		return contracts.Page[contracts.CrawlProgress]{}, err
	}
	rows, err := f.db.QueryContext(ctx, `SELECT rowid, id, status, discovered_count, fetched_count, analysed_count, failed_count, updated_at, terminal_reason FROM crawl WHERE project_id = ? AND deleted_at IS NULL AND rowid > ? ORDER BY rowid LIMIT ?`, projectID, after, limit+1)
	if err != nil {
		return contracts.Page[contracts.CrawlProgress]{}, err
	}
	defer rows.Close()
	result := contracts.Page[contracts.CrawlProgress]{Items: make([]contracts.CrawlProgress, 0, limit)}
	var last int64
	for rows.Next() {
		var item contracts.CrawlProgress
		var updated string
		var rowid int64
		if err := rows.Scan(&rowid, &item.CrawlID, &item.Status, &item.Discovered, &item.Fetched, &item.Analysed, &item.Failed, &updated, &item.TerminalReason); err != nil {
			return result, err
		}
		if len(result.Items) == limit {
			result.NextCursor = encodeCursor(last)
			break
		}
		item.UpdatedAt, _ = time.Parse(time.RFC3339Nano, updated)
		result.Items = append(result.Items, item)
		last = rowid
	}
	return result, rows.Err()
}
