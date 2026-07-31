ALTER TABLE rendered_page ADD COLUMN engine_version TEXT NOT NULL DEFAULT '';
ALTER TABLE rendered_page ADD COLUMN viewport TEXT NOT NULL DEFAULT '';
ALTER TABLE rendered_page ADD COLUMN screenshot_status TEXT NOT NULL DEFAULT 'not_requested' CHECK (screenshot_status IN ('not_requested','stored','truncated'));
ALTER TABLE rendered_page ADD COLUMN console_count INTEGER NOT NULL DEFAULT 0 CHECK (console_count >= 0);
ALTER TABLE rendered_page ADD COLUMN resource_failure_count INTEGER NOT NULL DEFAULT 0 CHECK (resource_failure_count >= 0);
ALTER TABLE rendered_page ADD COLUMN accessibility_count INTEGER NOT NULL DEFAULT 0 CHECK (accessibility_count >= 0);

CREATE TABLE render_console (
    crawl_url_id INTEGER NOT NULL REFERENCES crawl_url(id) ON DELETE CASCADE,
    position INTEGER NOT NULL,
    level TEXT NOT NULL,
    message TEXT NOT NULL,
    PRIMARY KEY(crawl_url_id,position)
) STRICT;

CREATE TABLE render_resource_failure (
    crawl_url_id INTEGER NOT NULL REFERENCES crawl_url(id) ON DELETE CASCADE,
    position INTEGER NOT NULL,
    resource_type TEXT NOT NULL,
    url TEXT NOT NULL,
    error_code TEXT NOT NULL,
    PRIMARY KEY(crawl_url_id,position)
) STRICT;

CREATE TABLE accessibility_finding (
    crawl_url_id INTEGER NOT NULL REFERENCES crawl_url(id) ON DELETE CASCADE,
    position INTEGER NOT NULL,
    rule_id TEXT NOT NULL,
    impact TEXT NOT NULL,
    tags_json TEXT NOT NULL CHECK (json_valid(tags_json)),
    target TEXT NOT NULL,
    html TEXT NOT NULL,
    help TEXT NOT NULL,
    engine_version TEXT NOT NULL,
    PRIMARY KEY(crawl_url_id,position)
) STRICT;

CREATE TABLE page_artifact (
    artifact_id TEXT PRIMARY KEY REFERENCES artifact(id) ON DELETE CASCADE,
    crawl_url_id INTEGER NOT NULL REFERENCES crawl_url(id) ON DELETE CASCADE,
    kind TEXT NOT NULL CHECK (kind IN ('rendered_dom','viewport_screenshot')),
    mime_type TEXT NOT NULL,
    viewport TEXT NOT NULL DEFAULT '',
    engine_version TEXT NOT NULL DEFAULT '',
    UNIQUE(crawl_url_id,kind)
) STRICT;

CREATE INDEX idx_page_artifact_page ON page_artifact(crawl_url_id,kind);
