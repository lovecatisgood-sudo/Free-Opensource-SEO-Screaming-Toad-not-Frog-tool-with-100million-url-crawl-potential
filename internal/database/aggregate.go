package database

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/seo-auditor/seo-auditor/internal/contracts"
)

// FinalizeAudit evaluates rules that require a complete crawl graph. It is
// idempotent so interrupted finalization can be retried safely.
func (f *Frontier) FinalizeAudit(ctx context.Context, crawlID contracts.ID, deepPageDepth int) error {
	if crawlID == "" || deepPageDepth < 0 {
		return errors.New("crawl ID and non-negative depth threshold are required")
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	return f.writer.Submit(ctx, func(ctx context.Context, tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx, `DELETE FROM issue WHERE crawl_id = ? AND (
rule_id IN ('AUD-06','AUD-07','AUD-08','AUD-11') OR
(rule_id='AUD-01' AND subject_type='link') OR
(rule_id='AUD-02' AND json_type(evidence_json,'$.duplicate_count') IS NOT NULL) OR
(rule_id='AUD-04' AND json_type(evidence_json,'$.target_status') IS NOT NULL) OR
(rule_id='AUD-10' AND json_type(evidence_json,'$.status_code') IS NOT NULL))`, crawlID); err != nil {
			return err
		}
		statements := []struct {
			query string
			args  []any
		}{
			{`INSERT INTO issue(crawl_id,rule_id,rule_version,subject_type,subject_id,severity,evidence_json,created_at)
SELECT ?, 'AUD-01', 1, 'link', CAST(l.id AS TEXT), 'error', json_object('source_url',su.request_key,'target_url',tu.request_key,'status_code',COALESCE(fa.status_code,0)), ?
FROM link l JOIN url su ON su.id=l.source_url_id JOIN url tu ON tu.id=l.target_url_id
LEFT JOIN crawl_url cu ON cu.crawl_id=l.crawl_id AND cu.url_id=l.target_url_id
LEFT JOIN fetch_attempt fa ON fa.crawl_url_id=cu.id AND fa.attempt=cu.attempt_count
WHERE l.crawl_id=? AND l.link_kind='internal' AND (cu.id IS NULL OR cu.state='failed' OR COALESCE(fa.status_code,0)>=400)`, []any{crawlID, now, crawlID}},
			{`INSERT INTO issue(crawl_id,rule_id,rule_version,subject_type,subject_id,severity,evidence_json,created_at)
SELECT ?, 'AUD-02', 1, 'page', CAST(p.id AS TEXT), 'warning', json_object('field','title','value',p.title,'duplicate_count',d.n), ?
FROM page p JOIN crawl_url cu ON cu.id=p.crawl_url_id JOIN
(SELECT p2.title value,count(*) n FROM page p2 JOIN crawl_url cu2 ON cu2.id=p2.crawl_url_id WHERE cu2.crawl_id=? AND trim(COALESCE(p2.title,''))<>'' GROUP BY p2.title HAVING count(*)>1) d ON d.value=p.title
WHERE cu.crawl_id=?`, []any{crawlID, now, crawlID, crawlID}},
			{`INSERT INTO issue(crawl_id,rule_id,rule_version,subject_type,subject_id,severity,evidence_json,created_at)
SELECT ?, 'AUD-02', 1, 'page', CAST(p.id AS TEXT), 'warning', json_object('field','meta_description','value',p.meta_description,'duplicate_count',d.n), ?
FROM page p JOIN crawl_url cu ON cu.id=p.crawl_url_id JOIN
(SELECT p2.meta_description value,count(*) n FROM page p2 JOIN crawl_url cu2 ON cu2.id=p2.crawl_url_id WHERE cu2.crawl_id=? AND trim(COALESCE(p2.meta_description,''))<>'' GROUP BY p2.meta_description HAVING count(*)>1) d ON d.value=p.meta_description
WHERE cu.crawl_id=?`, []any{crawlID, now, crawlID, crawlID}},
			{`INSERT INTO issue(crawl_id,rule_id,rule_version,subject_type,subject_id,severity,evidence_json,created_at)
SELECT ?, 'AUD-04', 1, 'page', CAST(p.id AS TEXT), 'error', json_object('canonical',p.canonical_url,'target_status',COALESCE(fa.status_code,0),'target_state',COALESCE(tcu.state,'not_crawled')), ?
FROM page p JOIN crawl_url cu ON cu.id=p.crawl_url_id
LEFT JOIN url target ON target.project_id=(SELECT project_id FROM crawl WHERE id=?) AND target.request_key=p.canonical_url
LEFT JOIN crawl_url tcu ON tcu.crawl_id=? AND tcu.url_id=target.id
LEFT JOIN fetch_attempt fa ON fa.crawl_url_id=tcu.id AND fa.attempt=tcu.attempt_count
WHERE cu.crawl_id=? AND p.canonical_url IS NOT NULL AND (tcu.id IS NULL OR tcu.state='failed' OR COALESCE(fa.status_code,0)<>200)`, []any{crawlID, now, crawlID, crawlID, crawlID}},
			{`INSERT INTO issue(crawl_id,rule_id,rule_version,subject_type,subject_id,severity,evidence_json,created_at)
SELECT ?, 'AUD-06', 1, 'page', CAST(p.id AS TEXT), 'info', json_object('in_sitemap',json('false'),'url',u.request_key), ?
FROM page p JOIN crawl_url cu ON cu.id=p.crawl_url_id JOIN url u ON u.id=cu.url_id
LEFT JOIN sitemap_entry se ON se.url_id=u.id AND se.sitemap_id IN (SELECT id FROM sitemap WHERE crawl_id=?)
LEFT JOIN fetch_attempt fa ON fa.crawl_url_id=cu.id AND fa.attempt=cu.attempt_count
WHERE cu.crawl_id=? AND se.url_id IS NULL AND fa.status_code=200 AND lower(COALESCE(p.robots_directives,'')||','||COALESCE(p.x_robots_tag,'')) NOT LIKE '%noindex%'`, []any{crawlID, now, crawlID, crawlID}},
			{`INSERT INTO issue(crawl_id,rule_id,rule_version,subject_type,subject_id,severity,evidence_json,created_at)
SELECT ?, 'AUD-06', 1, 'sitemap', CAST(id AS TEXT), 'error', json_object('sitemap_url',url,'status',status), ? FROM sitemap WHERE crawl_id=? AND status<>'ok'`, []any{crawlID, now, crawlID}},
			{`INSERT INTO issue(crawl_id,rule_id,rule_version,subject_type,subject_id,severity,evidence_json,created_at)
SELECT ?, 'AUD-06', 1, 'url', CAST(u.id AS TEXT), 'warning', json_object('url',u.request_key,'status_code',COALESCE(fa.status_code,0),'state',COALESCE(cu.state,'not_crawled')), ?
FROM sitemap_entry se JOIN sitemap s ON s.id=se.sitemap_id JOIN url u ON u.id=se.url_id
LEFT JOIN crawl_url cu ON cu.crawl_id=s.crawl_id AND cu.url_id=u.id LEFT JOIN fetch_attempt fa ON fa.crawl_url_id=cu.id AND fa.attempt=cu.attempt_count
WHERE s.crawl_id=? AND (cu.id IS NULL OR cu.state='failed' OR COALESCE(fa.status_code,0)<>200)`, []any{crawlID, now, crawlID}},
			{`INSERT INTO issue(crawl_id,rule_id,rule_version,subject_type,subject_id,severity,evidence_json,created_at)
SELECT ?, 'AUD-07', 1, 'url', CAST(cu.id AS TEXT), 'warning', json_object('url',u.request_key,'robots_decision','disallowed'), ?
FROM crawl_url cu JOIN url u ON u.id=cu.url_id WHERE cu.crawl_id=? AND cu.robots_decision='disallowed'`, []any{crawlID, now, crawlID}},
			{`INSERT INTO issue(crawl_id,rule_id,rule_version,subject_type,subject_id,severity,evidence_json,created_at)
SELECT ?, 'AUD-08', 1, 'page', CAST(p.id AS TEXT), 'warning', json_object('content_hash',p.content_hash,'duplicate_count',d.n-1), ?
FROM page p JOIN crawl_url cu ON cu.id=p.crawl_url_id JOIN
(SELECT p2.content_hash value,count(*) n FROM page p2 JOIN crawl_url cu2 ON cu2.id=p2.crawl_url_id WHERE cu2.crawl_id=? AND COALESCE(p2.content_hash,'')<>'' GROUP BY p2.content_hash HAVING count(*)>1) d ON d.value=p.content_hash
WHERE cu.crawl_id=?`, []any{crawlID, now, crawlID, crawlID}},
			{`INSERT INTO issue(crawl_id,rule_id,rule_version,subject_type,subject_id,severity,evidence_json,created_at)
SELECT ?, 'AUD-10', 1, 'image', CAST(i.id AS TEXT), 'error', json_object('image_url',i.request_key,'status_code',COALESCE(fa.status_code,0),'state',COALESCE(cu.state,'not_crawled')), ?
FROM image i JOIN page_image pi ON pi.image_id=i.id JOIN page p ON p.id=pi.page_id JOIN crawl_url source ON source.id=p.crawl_url_id
LEFT JOIN url u ON u.project_id=i.project_id AND u.request_key=i.request_key LEFT JOIN crawl_url cu ON cu.crawl_id=? AND cu.url_id=u.id
LEFT JOIN fetch_attempt fa ON fa.crawl_url_id=cu.id AND fa.attempt=cu.attempt_count
WHERE source.crawl_id=? AND (cu.id IS NULL OR cu.state='failed' OR COALESCE(fa.status_code,0)>=400)
GROUP BY i.id`, []any{crawlID, now, crawlID, crawlID}},
			{`INSERT INTO issue(crawl_id,rule_id,rule_version,subject_type,subject_id,severity,evidence_json,created_at)
SELECT ?, 'AUD-11', 1, 'page', CAST(p.id AS TEXT), 'warning', json_object('depth',cu.depth,'threshold',?), ?
FROM page p JOIN crawl_url cu ON cu.id=p.crawl_url_id WHERE cu.crawl_id=? AND cu.depth>?`, []any{crawlID, deepPageDepth, now, crawlID, deepPageDepth}},
			{`INSERT INTO issue(crawl_id,rule_id,rule_version,subject_type,subject_id,severity,evidence_json,created_at)
SELECT ?, 'AUD-11', 1, 'page', CAST(p.id AS TEXT), 'warning', json_object('inlinks',0), ?
FROM page p JOIN crawl_url cu ON cu.id=p.crawl_url_id
WHERE cu.crawl_id=? AND cu.depth>0 AND NOT EXISTS (SELECT 1 FROM link l WHERE l.crawl_id=? AND l.target_url_id=cu.url_id AND l.link_kind='internal')`, []any{crawlID, now, crawlID, crawlID}},
			{`INSERT INTO issue(crawl_id,rule_id,rule_version,subject_type,subject_id,severity,evidence_json,created_at)
SELECT ?, 'AUD-11', 1, 'page', CAST(p.id AS TEXT), 'info', json_object('nofollow_links',count(*)), ?
FROM page p JOIN crawl_url cu ON cu.id=p.crawl_url_id JOIN link l ON l.crawl_id=cu.crawl_id AND l.source_url_id=cu.url_id
WHERE cu.crawl_id=? AND l.link_kind='internal' AND (' '||lower(replace(l.rel,',',' '))||' ') LIKE '% nofollow %' GROUP BY p.id`, []any{crawlID, now, crawlID}},
		}
		for _, statement := range statements {
			if _, err := tx.ExecContext(ctx, statement.query, statement.args...); err != nil {
				return err
			}
		}
		return nil
	})
}
