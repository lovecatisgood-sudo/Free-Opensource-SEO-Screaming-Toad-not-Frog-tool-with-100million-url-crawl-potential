CREATE TABLE IF NOT EXISTS rendered_page (
    id INTEGER PRIMARY KEY,
    crawl_url_id INTEGER NOT NULL UNIQUE REFERENCES crawl_url(id) ON DELETE CASCADE,
    status TEXT NOT NULL CHECK (status IN ('completed','blocked','failed')),
    error_code TEXT NOT NULL DEFAULT '',
    final_url TEXT NOT NULL DEFAULT '',
    request_count INTEGER NOT NULL DEFAULT 0 CHECK (request_count >= 0),
    transferred_bytes INTEGER NOT NULL DEFAULT 0 CHECK (transferred_bytes >= 0),
    title TEXT,
    meta_description TEXT,
    canonical_url TEXT,
    robots_directives TEXT,
    language TEXT,
    text_length INTEGER NOT NULL DEFAULT 0 CHECK (text_length >= 0),
    content_hash TEXT,
    html_hash TEXT,
    headings_json TEXT NOT NULL DEFAULT '[]' CHECK (json_valid(headings_json)),
    images_json TEXT NOT NULL DEFAULT '[]' CHECK (json_valid(images_json)),
    hreflang_json TEXT NOT NULL DEFAULT '[]' CHECK (json_valid(hreflang_json)),
    structured_data_json TEXT NOT NULL DEFAULT '[]' CHECK (json_valid(structured_data_json)),
    social_json TEXT NOT NULL DEFAULT '{}' CHECK (json_valid(social_json)),
    rendered_at TEXT NOT NULL
) STRICT;

CREATE TABLE IF NOT EXISTS render_difference (
    crawl_url_id INTEGER NOT NULL REFERENCES crawl_url(id) ON DELETE CASCADE,
    field TEXT NOT NULL,
    raw_value TEXT NOT NULL,
    rendered_value TEXT NOT NULL,
    PRIMARY KEY(crawl_url_id, field)
) STRICT;

CREATE INDEX IF NOT EXISTS idx_rendered_page_status ON rendered_page(status, id);
