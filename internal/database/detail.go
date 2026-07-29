package database

import (
	"context"
	"database/sql"

	"github.com/seo-auditor/seo-auditor/internal/contracts"
)

type HeadingRecord struct {
	Position int    `json:"position"`
	Level    int    `json:"level"`
	Text     string `json:"text"`
}
type LinkRecord struct {
	ID         int64  `json:"id"`
	SourceURL  string `json:"source_url"`
	TargetURL  string `json:"target_url"`
	RawTarget  string `json:"raw_target"`
	AnchorText string `json:"anchor_text"`
	Rel        string `json:"rel"`
	Kind       string `json:"kind"`
}
type ImageRecord struct {
	URL        string `json:"url"`
	AltText    string `json:"alt_text"`
	AltPresent bool   `json:"alt_present"`
	Width      int    `json:"width,omitempty"`
	Height     int    `json:"height,omitempty"`
}
type HreflangRecord struct {
	Language        string `json:"language"`
	TargetURL       string `json:"target_url"`
	ValidationState string `json:"validation_state"`
}
type StructuredDataRecord struct {
	Format       string `json:"format"`
	TypeSummary  string `json:"type_summary"`
	EvidenceJSON string `json:"evidence_json"`
}
type PageDetail struct {
	Page           PageRecord             `json:"page"`
	Headings       []HeadingRecord        `json:"headings"`
	Inlinks        []LinkRecord           `json:"inlinks"`
	Outlinks       []LinkRecord           `json:"outlinks"`
	Images         []ImageRecord          `json:"images"`
	Hreflangs      []HreflangRecord       `json:"hreflang"`
	StructuredData []StructuredDataRecord `json:"structured_data"`
	Issues         []IssueRecord          `json:"issues"`
}

func (f *Frontier) GetPage(ctx context.Context, crawlID contracts.ID, pageID int64) (PageDetail, error) {
	var result PageDetail
	err := f.db.QueryRowContext(ctx, `SELECT p.id,u.request_key,COALESCE(fa.status_code,0),cu.depth,COALESCE(p.title,''),COALESCE(p.meta_description,''),COALESCE(p.canonical_url,''),COALESCE(p.robots_directives,''),COALESCE(p.language,''),p.text_length,COALESCE(p.content_hash,'')
FROM page p JOIN crawl_url cu ON cu.id=p.crawl_url_id JOIN url u ON u.id=cu.url_id LEFT JOIN fetch_attempt fa ON fa.crawl_url_id=cu.id AND fa.attempt=cu.attempt_count WHERE cu.crawl_id=? AND p.id=?`, crawlID, pageID).Scan(
		&result.Page.ID, &result.Page.URL, &result.Page.StatusCode, &result.Page.Depth, &result.Page.Title, &result.Page.MetaDescription, &result.Page.CanonicalURL, &result.Page.RobotsDirectives, &result.Page.Language, &result.Page.TextLength, &result.Page.ContentHash)
	if err != nil {
		return result, err
	}
	result.Headings = []HeadingRecord{}
	result.Inlinks = []LinkRecord{}
	result.Outlinks = []LinkRecord{}
	result.Images = []ImageRecord{}
	result.Hreflangs = []HreflangRecord{}
	result.StructuredData = []StructuredDataRecord{}
	result.Issues = []IssueRecord{}
	if err := scanRows(ctx, f.db, `SELECT position,level,text FROM heading WHERE page_id=? ORDER BY position`, []any{pageID}, func(rows *sql.Rows) error {
		var item HeadingRecord
		if err := rows.Scan(&item.Position, &item.Level, &item.Text); err != nil {
			return err
		}
		result.Headings = append(result.Headings, item)
		return nil
	}); err != nil {
		return result, err
	}
	if err := scanRows(ctx, f.db, `SELECT l.id,su.request_key,COALESCE(tu.request_key,''),l.raw_target,l.anchor_text,l.rel,l.link_kind FROM link l JOIN url su ON su.id=l.source_url_id LEFT JOIN url tu ON tu.id=l.target_url_id WHERE l.crawl_id=? AND l.source_url_id=(SELECT cu.url_id FROM page p JOIN crawl_url cu ON cu.id=p.crawl_url_id WHERE p.id=?) ORDER BY l.id LIMIT 10000`, []any{crawlID, pageID}, func(rows *sql.Rows) error {
		var item LinkRecord
		if err := rows.Scan(&item.ID, &item.SourceURL, &item.TargetURL, &item.RawTarget, &item.AnchorText, &item.Rel, &item.Kind); err != nil {
			return err
		}
		result.Outlinks = append(result.Outlinks, item)
		return nil
	}); err != nil {
		return result, err
	}
	if err := scanRows(ctx, f.db, `SELECT l.id,su.request_key,COALESCE(tu.request_key,''),l.raw_target,l.anchor_text,l.rel,l.link_kind FROM link l JOIN url su ON su.id=l.source_url_id LEFT JOIN url tu ON tu.id=l.target_url_id WHERE l.crawl_id=? AND l.target_url_id=(SELECT cu.url_id FROM page p JOIN crawl_url cu ON cu.id=p.crawl_url_id WHERE p.id=?) ORDER BY l.id LIMIT 10000`, []any{crawlID, pageID}, func(rows *sql.Rows) error {
		var item LinkRecord
		if err := rows.Scan(&item.ID, &item.SourceURL, &item.TargetURL, &item.RawTarget, &item.AnchorText, &item.Rel, &item.Kind); err != nil {
			return err
		}
		result.Inlinks = append(result.Inlinks, item)
		return nil
	}); err != nil {
		return result, err
	}
	if err := scanRows(ctx, f.db, `SELECT i.request_key,COALESCE(pi.alt_text,''),pi.alt_text IS NOT NULL,COALESCE(i.declared_width,0),COALESCE(i.declared_height,0) FROM page_image pi JOIN image i ON i.id=pi.image_id WHERE pi.page_id=? ORDER BY pi.position LIMIT 10000`, []any{pageID}, func(rows *sql.Rows) error {
		var item ImageRecord
		if err := rows.Scan(&item.URL, &item.AltText, &item.AltPresent, &item.Width, &item.Height); err != nil {
			return err
		}
		result.Images = append(result.Images, item)
		return nil
	}); err != nil {
		return result, err
	}
	if err := scanRows(ctx, f.db, `SELECT language_code,target_url,validation_state FROM hreflang WHERE page_id=? ORDER BY id LIMIT 10000`, []any{pageID}, func(rows *sql.Rows) error {
		var item HreflangRecord
		if err := rows.Scan(&item.Language, &item.TargetURL, &item.ValidationState); err != nil {
			return err
		}
		result.Hreflangs = append(result.Hreflangs, item)
		return nil
	}); err != nil {
		return result, err
	}
	if err := scanRows(ctx, f.db, `SELECT format,type_summary,evidence_json FROM structured_data WHERE page_id=? ORDER BY id LIMIT 1000`, []any{pageID}, func(rows *sql.Rows) error {
		var item StructuredDataRecord
		if err := rows.Scan(&item.Format, &item.TypeSummary, &item.EvidenceJSON); err != nil {
			return err
		}
		result.StructuredData = append(result.StructuredData, item)
		return nil
	}); err != nil {
		return result, err
	}
	if err := scanRows(ctx, f.db, `SELECT id,rule_id,rule_version,subject_type,subject_id,severity,evidence_json,created_at FROM issue WHERE crawl_id=? AND subject_type='page' AND subject_id=CAST(? AS TEXT) ORDER BY id LIMIT 1000`, []any{crawlID, pageID}, func(rows *sql.Rows) error {
		var item IssueRecord
		if err := rows.Scan(&item.ID, &item.RuleID, &item.RuleVersion, &item.SubjectType, &item.SubjectID, &item.Severity, &item.EvidenceJSON, &item.CreatedAt); err != nil {
			return err
		}
		result.Issues = append(result.Issues, item)
		return nil
	}); err != nil {
		return result, err
	}
	return result, nil
}

func (f *Frontier) ListLinks(ctx context.Context, crawlID contracts.ID, page contracts.PageRequest) (contracts.Page[LinkRecord], error) {
	return listKeyset(ctx, page, func(after int64, limit int) (*sql.Rows, error) {
		return f.db.QueryContext(ctx, `SELECT l.id,su.request_key,COALESCE(tu.request_key,''),l.raw_target,l.anchor_text,l.rel,l.link_kind FROM link l JOIN url su ON su.id=l.source_url_id LEFT JOIN url tu ON tu.id=l.target_url_id WHERE l.crawl_id=? AND l.id>? ORDER BY l.id LIMIT ?`, crawlID, after, limit)
	}, func(rows *sql.Rows) (LinkRecord, error) {
		var item LinkRecord
		err := rows.Scan(&item.ID, &item.SourceURL, &item.TargetURL, &item.RawTarget, &item.AnchorText, &item.Rel, &item.Kind)
		return item, err
	}, func(item LinkRecord) int64 { return item.ID })
}

type rowQueryer interface {
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
}

func scanRows(ctx context.Context, db rowQueryer, query string, args []any, scan func(*sql.Rows) error) error {
	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		if err := scan(rows); err != nil {
			return err
		}
	}
	return rows.Err()
}

type CrawlComparison struct {
	BaseCrawlID        contracts.ID `json:"base_crawl_id"`
	TargetCrawlID      contracts.ID `json:"target_crawl_id"`
	ConfigurationMatch bool         `json:"configuration_match"`
	AddedPages         int64        `json:"added_pages"`
	RemovedPages       int64        `json:"removed_pages"`
	ChangedPages       int64        `json:"changed_pages"`
	NewIssues          int64        `json:"new_issues"`
	FixedIssues        int64        `json:"fixed_issues"`
}

func (f *Frontier) CompareCrawls(ctx context.Context, baseID, targetID contracts.ID) (CrawlComparison, error) {
	result := CrawlComparison{BaseCrawlID: baseID, TargetCrawlID: targetID}
	var baseConfig, targetConfig string
	if err := f.db.QueryRowContext(ctx, `SELECT config_json FROM crawl WHERE id=?`, baseID).Scan(&baseConfig); err != nil {
		return result, err
	}
	if err := f.db.QueryRowContext(ctx, `SELECT config_json FROM crawl WHERE id=?`, targetID).Scan(&targetConfig); err != nil {
		return result, err
	}
	result.ConfigurationMatch = baseConfig == targetConfig
	queries := []struct {
		target *int64
		sql    string
	}{
		{&result.AddedPages, `SELECT count(*) FROM page p JOIN crawl_url cu ON cu.id=p.crawl_url_id JOIN url u ON u.id=cu.url_id WHERE cu.crawl_id=? AND NOT EXISTS(SELECT 1 FROM page bp JOIN crawl_url bcu ON bcu.id=bp.crawl_url_id JOIN url bu ON bu.id=bcu.url_id WHERE bcu.crawl_id=? AND bu.request_key=u.request_key)`},
		{&result.RemovedPages, `SELECT count(*) FROM page p JOIN crawl_url cu ON cu.id=p.crawl_url_id JOIN url u ON u.id=cu.url_id WHERE cu.crawl_id=? AND NOT EXISTS(SELECT 1 FROM page tp JOIN crawl_url tcu ON tcu.id=tp.crawl_url_id JOIN url tu ON tu.id=tcu.url_id WHERE tcu.crawl_id=? AND tu.request_key=u.request_key)`},
		{&result.ChangedPages, `SELECT count(*) FROM page tp JOIN crawl_url tcu ON tcu.id=tp.crawl_url_id JOIN url tu ON tu.id=tcu.url_id JOIN crawl_url bcu ON bcu.crawl_id=? JOIN url bu ON bu.id=bcu.url_id AND bu.request_key=tu.request_key JOIN page bp ON bp.crawl_url_id=bcu.id LEFT JOIN fetch_attempt tf ON tf.crawl_url_id=tcu.id AND tf.attempt=tcu.attempt_count LEFT JOIN fetch_attempt bf ON bf.crawl_url_id=bcu.id AND bf.attempt=bcu.attempt_count WHERE tcu.crawl_id=? AND (COALESCE(tp.title,'')<>COALESCE(bp.title,'') OR COALESCE(tp.canonical_url,'')<>COALESCE(bp.canonical_url,'') OR COALESCE(tp.content_hash,'')<>COALESCE(bp.content_hash,'') OR COALESCE(tf.status_code,0)<>COALESCE(bf.status_code,0))`},
		{&result.NewIssues, `SELECT count(*) FROM issue i WHERE i.crawl_id=? AND NOT EXISTS(SELECT 1 FROM issue b WHERE b.crawl_id=? AND b.rule_id=i.rule_id AND b.subject_type=i.subject_type AND b.evidence_json=i.evidence_json)`},
		{&result.FixedIssues, `SELECT count(*) FROM issue i WHERE i.crawl_id=? AND NOT EXISTS(SELECT 1 FROM issue t WHERE t.crawl_id=? AND t.rule_id=i.rule_id AND t.subject_type=i.subject_type AND t.evidence_json=i.evidence_json)`},
	}
	for index, item := range queries {
		first, second := targetID, baseID
		if index == 1 || index == 4 {
			first, second = baseID, targetID
		}
		if index == 2 {
			first, second = baseID, targetID
		}
		if err := f.db.QueryRowContext(ctx, item.sql, first, second).Scan(item.target); err != nil {
			return result, err
		}
	}
	return result, nil
}
