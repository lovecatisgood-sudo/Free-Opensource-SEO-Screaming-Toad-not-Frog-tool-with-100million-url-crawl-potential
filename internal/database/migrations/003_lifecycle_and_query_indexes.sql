ALTER TABLE crawl ADD COLUMN deleted_at TEXT;
ALTER TABLE project ADD COLUMN archived_at TEXT;

CREATE INDEX IF NOT EXISTS idx_crawl_project_history ON crawl(project_id, deleted_at, created_at, id);
CREATE INDEX IF NOT EXISTS idx_page_content_hash ON page(content_hash) WHERE content_hash IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_page_title ON page(title) WHERE title IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_sitemap_entry_url ON sitemap_entry(url_id, sitemap_id);
CREATE INDEX IF NOT EXISTS idx_artifact_expiry ON artifact(expires_at) WHERE expires_at IS NOT NULL;
