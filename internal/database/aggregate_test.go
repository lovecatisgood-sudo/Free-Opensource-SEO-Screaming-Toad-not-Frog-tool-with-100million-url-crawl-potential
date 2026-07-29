package database

import (
	"context"
	"testing"

	"github.com/seo-auditor/seo-auditor/internal/fetchpolicy"
)

func TestFinalizeAuditComputesGraphRulesAndIsIdempotent(t *testing.T) {
	t.Parallel()
	frontier, projectID, crawlID := testFrontier(t)
	ctx := context.Background()
	for index, raw := range []string{"https://example.com/a", "https://example.com/b"} {
		normalized, err := fetchpolicy.NormalizeURL(raw)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := frontier.Enqueue(ctx, Discovery{CrawlID: crawlID, ProjectID: projectID, URL: normalized, Depth: index, DiscoveryKind: "test", MaximumURLs: 10}); err != nil {
			t.Fatal(err)
		}
	}
	rows, err := frontier.db.QueryContext(ctx, `SELECT cu.id,u.id,u.request_key FROM crawl_url cu JOIN url u ON u.id=cu.url_id WHERE cu.crawl_id=? ORDER BY cu.id`, crawlID)
	if err != nil {
		t.Fatal(err)
	}
	type item struct {
		crawlURLID, urlID int64
		url               string
	}
	var items []item
	for rows.Next() {
		var value item
		if err := rows.Scan(&value.crawlURLID, &value.urlID, &value.url); err != nil {
			t.Fatal(err)
		}
		items = append(items, value)
	}
	if err := rows.Close(); err != nil {
		t.Fatal(err)
	}
	for _, value := range items {
		if _, err := frontier.db.ExecContext(ctx, `UPDATE crawl_url SET state='analysed',attempt_count=1 WHERE id=?`, value.crawlURLID); err != nil {
			t.Fatal(err)
		}
		if _, err := frontier.db.ExecContext(ctx, `INSERT INTO fetch_attempt(crawl_url_id,attempt,started_at,finished_at,status_code,content_type) VALUES (?,1,'2026-01-01T00:00:00Z','2026-01-01T00:00:01Z',200,'text/html')`, value.crawlURLID); err != nil {
			t.Fatal(err)
		}
		if _, err := frontier.db.ExecContext(ctx, `INSERT INTO page(crawl_url_id,extraction_mode,title,meta_description,robots_directives,language,text_length,content_hash,extracted_at,social_json) VALUES (?,'raw','Duplicate title','','','en',10,'same-hash','2026-01-01T00:00:01Z','{}')`, value.crawlURLID); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := frontier.db.ExecContext(ctx, `INSERT INTO link(crawl_id,source_url_id,target_url_id,raw_target,anchor_text,rel,link_kind,extraction_mode) VALUES (?,?,?,?,?,'','internal','raw')`, crawlID, items[0].urlID, items[1].urlID, items[1].url, "B"); err != nil {
		t.Fatal(err)
	}
	if _, err := frontier.db.ExecContext(ctx, `INSERT INTO sitemap(crawl_id,url,status,discovered_from) VALUES (?,'https://example.com/sitemap.xml','ok','test')`, crawlID); err != nil {
		t.Fatal(err)
	}
	if _, err := frontier.db.ExecContext(ctx, `INSERT INTO sitemap_entry(sitemap_id,url_id) VALUES ((SELECT id FROM sitemap WHERE crawl_id=?),?)`, crawlID, items[0].urlID); err != nil {
		t.Fatal(err)
	}

	assertCounts := func() {
		t.Helper()
		for rule, wanted := range map[string]int{"AUD-02": 2, "AUD-06": 1, "AUD-08": 2, "AUD-11": 0} {
			var got int
			if err := frontier.db.QueryRowContext(ctx, `SELECT count(*) FROM issue WHERE crawl_id=? AND rule_id=?`, crawlID, rule).Scan(&got); err != nil {
				t.Fatal(err)
			}
			if got != wanted {
				t.Fatalf("%s count=%d want=%d", rule, got, wanted)
			}
		}
	}
	if err := frontier.FinalizeAudit(ctx, crawlID, 4); err != nil {
		t.Fatal(err)
	}
	assertCounts()
	if err := frontier.FinalizeAudit(ctx, crawlID, 4); err != nil {
		t.Fatal(err)
	}
	assertCounts()
}

func TestFinalizeAuditValidatesCanonicalHreflangAndSitemapGraphs(t *testing.T) {
	t.Parallel()
	frontier, projectID, crawlID := testFrontier(t)
	ctx := context.Background()
	for index, raw := range []string{"https://example.com/a", "https://example.com/b"} {
		normalized, _ := fetchpolicy.NormalizeURL(raw)
		if _, err := frontier.Enqueue(ctx, Discovery{CrawlID: crawlID, ProjectID: projectID, URL: normalized, Depth: index, DiscoveryKind: "test", MaximumURLs: 10}); err != nil {
			t.Fatal(err)
		}
	}
	rows, err := frontier.db.QueryContext(ctx, `SELECT cu.id,u.id,u.request_key FROM crawl_url cu JOIN url u ON u.id=cu.url_id WHERE cu.crawl_id=? ORDER BY u.request_key`, crawlID)
	if err != nil {
		t.Fatal(err)
	}
	type record struct {
		crawlURLID, urlID int64
		url               string
	}
	var records []record
	for rows.Next() {
		var item record
		if err := rows.Scan(&item.crawlURLID, &item.urlID, &item.url); err != nil {
			t.Fatal(err)
		}
		records = append(records, item)
	}
	_ = rows.Close()
	for index, item := range records {
		if _, err := frontier.db.ExecContext(ctx, `UPDATE crawl_url SET state='analysed',attempt_count=1 WHERE id=?`, item.crawlURLID); err != nil {
			t.Fatal(err)
		}
		if _, err := frontier.db.ExecContext(ctx, `INSERT INTO fetch_attempt(crawl_url_id,attempt,started_at,finished_at,status_code,content_type) VALUES (?,1,'2026-01-01T00:00:00Z','2026-01-01T00:00:01Z',200,'text/html')`, item.crawlURLID); err != nil {
			t.Fatal(err)
		}
		canonical := "https://example.com/b"
		robots := ""
		if index == 1 {
			canonical = "https://example.com/a"
			robots = "noindex"
		}
		if _, err := frontier.db.ExecContext(ctx, `INSERT INTO page(crawl_url_id,extraction_mode,title,meta_description,canonical_url,robots_directives,language,text_length,content_hash,extracted_at,social_json) VALUES (?,'raw',?,'description',?,?,'en',100,?,'2026-01-01T00:00:01Z','{}')`, item.crawlURLID, item.url, canonical, robots, item.url); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := frontier.db.ExecContext(ctx, `INSERT INTO hreflang(page_id,language_code,target_url,validation_state) VALUES ((SELECT p.id FROM page p WHERE p.crawl_url_id=?),'en','https://example.com/b','page_only')`, records[0].crawlURLID); err != nil {
		t.Fatal(err)
	}
	if _, err := frontier.db.ExecContext(ctx, `INSERT INTO sitemap(crawl_id,url,status,discovered_from) VALUES (?,'https://example.com/sitemap.xml','ok','test')`, crawlID); err != nil {
		t.Fatal(err)
	}
	if _, err := frontier.db.ExecContext(ctx, `INSERT INTO sitemap_entry(sitemap_id,url_id) VALUES ((SELECT id FROM sitemap WHERE crawl_id=?),?)`, crawlID, records[1].urlID); err != nil {
		t.Fatal(err)
	}
	if err := frontier.FinalizeAudit(ctx, crawlID, 4); err != nil {
		t.Fatal(err)
	}
	for rule, minimum := range map[string]int{"AUD-04": 1, "AUD-06": 1, "AUD-09": 1} {
		var count int
		if err := frontier.db.QueryRowContext(ctx, `SELECT count(*) FROM issue WHERE crawl_id=? AND rule_id=?`, crawlID, rule).Scan(&count); err != nil {
			t.Fatal(err)
		}
		if count < minimum {
			t.Fatalf("%s count=%d want at least %d", rule, count, minimum)
		}
	}
}
