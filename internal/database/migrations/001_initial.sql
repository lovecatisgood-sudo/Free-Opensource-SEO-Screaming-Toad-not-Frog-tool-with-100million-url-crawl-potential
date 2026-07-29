PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS schema_migration (
    version INTEGER PRIMARY KEY,
    applied_at TEXT NOT NULL
) STRICT;

CREATE TABLE IF NOT EXISTS project (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL CHECK (length(name) BETWEEN 1 AND 200),
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    deleted_at TEXT
) STRICT;

CREATE TABLE IF NOT EXISTS crawl_profile (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL REFERENCES project(id) ON DELETE CASCADE,
    version INTEGER NOT NULL CHECK (version > 0),
    name TEXT NOT NULL,
    config_json TEXT NOT NULL CHECK (json_valid(config_json)),
    created_at TEXT NOT NULL,
    UNIQUE(project_id, version)
) STRICT;

CREATE TABLE IF NOT EXISTS crawl (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL REFERENCES project(id) ON DELETE CASCADE,
    profile_id TEXT REFERENCES crawl_profile(id),
    seed_url TEXT NOT NULL,
    config_json TEXT NOT NULL CHECK (json_valid(config_json)),
    status TEXT NOT NULL CHECK (status IN ('pending','running','pausing','paused','cancelling','cancelled','completed','failed','limit_reached')),
    terminal_reason TEXT NOT NULL DEFAULT '',
    discovered_count INTEGER NOT NULL DEFAULT 0 CHECK (discovered_count >= 0),
    fetched_count INTEGER NOT NULL DEFAULT 0 CHECK (fetched_count >= 0),
    analysed_count INTEGER NOT NULL DEFAULT 0 CHECK (analysed_count >= 0),
    failed_count INTEGER NOT NULL DEFAULT 0 CHECK (failed_count >= 0),
    created_at TEXT NOT NULL,
    started_at TEXT,
    updated_at TEXT NOT NULL,
    finished_at TEXT
) STRICT;

CREATE TABLE IF NOT EXISTS url (
    id INTEGER PRIMARY KEY,
    project_id TEXT NOT NULL REFERENCES project(id) ON DELETE CASCADE,
    request_key TEXT NOT NULL,
    original_url TEXT NOT NULL,
    scheme TEXT NOT NULL,
    host TEXT NOT NULL,
    port TEXT NOT NULL DEFAULT '',
    path TEXT NOT NULL,
    query TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL,
    UNIQUE(project_id, request_key)
) STRICT;

CREATE TABLE IF NOT EXISTS crawl_url (
    id INTEGER PRIMARY KEY,
    crawl_id TEXT NOT NULL REFERENCES crawl(id) ON DELETE CASCADE,
    url_id INTEGER NOT NULL REFERENCES url(id) ON DELETE CASCADE,
    state TEXT NOT NULL CHECK (state IN ('discovered','accepted','queued','leased','fetched','extracted','analysed','skipped','retry_wait','failed')),
    depth INTEGER NOT NULL CHECK (depth >= 0),
    priority INTEGER NOT NULL DEFAULT 0,
    discovered_from_id INTEGER REFERENCES crawl_url(id),
    discovery_kind TEXT NOT NULL,
    attempt_count INTEGER NOT NULL DEFAULT 0 CHECK (attempt_count >= 0),
    lease_owner TEXT,
    lease_expires_at TEXT,
    next_attempt_at TEXT,
    robots_decision TEXT,
    skip_reason TEXT,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    UNIQUE(crawl_id, url_id)
) STRICT;

CREATE INDEX IF NOT EXISTS idx_crawl_url_frontier
ON crawl_url(crawl_id, state, next_attempt_at, priority DESC, depth, id);

CREATE INDEX IF NOT EXISTS idx_crawl_url_lease
ON crawl_url(crawl_id, lease_expires_at) WHERE state = 'leased';

CREATE TABLE IF NOT EXISTS fetch_attempt (
    id INTEGER PRIMARY KEY,
    crawl_url_id INTEGER NOT NULL REFERENCES crawl_url(id) ON DELETE CASCADE,
    attempt INTEGER NOT NULL CHECK (attempt > 0),
    started_at TEXT NOT NULL,
    finished_at TEXT,
    resolved_ip TEXT,
    status_code INTEGER,
    content_type TEXT,
    compressed_bytes INTEGER CHECK (compressed_bytes >= 0),
    decoded_bytes INTEGER CHECK (decoded_bytes >= 0),
    error_code TEXT,
    error_detail TEXT,
    UNIQUE(crawl_url_id, attempt)
) STRICT;

CREATE TABLE IF NOT EXISTS redirect_hop (
    fetch_attempt_id INTEGER NOT NULL REFERENCES fetch_attempt(id) ON DELETE CASCADE,
    hop INTEGER NOT NULL CHECK (hop >= 0),
    source_url TEXT NOT NULL,
    status_code INTEGER NOT NULL,
    target_url TEXT NOT NULL,
    policy_decision TEXT NOT NULL,
    PRIMARY KEY(fetch_attempt_id, hop)
) STRICT;

CREATE TABLE IF NOT EXISTS page (
    id INTEGER PRIMARY KEY,
    crawl_url_id INTEGER NOT NULL UNIQUE REFERENCES crawl_url(id) ON DELETE CASCADE,
    extraction_mode TEXT NOT NULL CHECK (extraction_mode IN ('raw','rendered')),
    title TEXT,
    meta_description TEXT,
    canonical_url TEXT,
    robots_directives TEXT,
    language TEXT,
    text_length INTEGER NOT NULL DEFAULT 0,
    content_hash TEXT,
    extracted_at TEXT NOT NULL
) STRICT;

CREATE TABLE IF NOT EXISTS heading (
    page_id INTEGER NOT NULL REFERENCES page(id) ON DELETE CASCADE,
    position INTEGER NOT NULL,
    level INTEGER NOT NULL CHECK (level BETWEEN 1 AND 6),
    text TEXT NOT NULL,
    PRIMARY KEY(page_id, position)
) STRICT;

CREATE TABLE IF NOT EXISTS link (
    id INTEGER PRIMARY KEY,
    crawl_id TEXT NOT NULL REFERENCES crawl(id) ON DELETE CASCADE,
    source_url_id INTEGER NOT NULL REFERENCES url(id) ON DELETE CASCADE,
    target_url_id INTEGER REFERENCES url(id) ON DELETE SET NULL,
    raw_target TEXT NOT NULL,
    anchor_text TEXT NOT NULL DEFAULT '',
    rel TEXT NOT NULL DEFAULT '',
    link_kind TEXT NOT NULL,
    extraction_mode TEXT NOT NULL CHECK (extraction_mode IN ('raw','rendered'))
) STRICT;

CREATE INDEX IF NOT EXISTS idx_link_source ON link(crawl_id, source_url_id);
CREATE INDEX IF NOT EXISTS idx_link_target ON link(crawl_id, target_url_id);

CREATE TABLE IF NOT EXISTS image (
    id INTEGER PRIMARY KEY,
    project_id TEXT NOT NULL REFERENCES project(id) ON DELETE CASCADE,
    request_key TEXT NOT NULL,
    original_url TEXT NOT NULL,
    UNIQUE(project_id, request_key)
) STRICT;

CREATE TABLE IF NOT EXISTS page_image (
    page_id INTEGER NOT NULL REFERENCES page(id) ON DELETE CASCADE,
    image_id INTEGER NOT NULL REFERENCES image(id) ON DELETE CASCADE,
    position INTEGER NOT NULL,
    alt_text TEXT,
    PRIMARY KEY(page_id, position)
) STRICT;

CREATE TABLE IF NOT EXISTS hreflang (
    id INTEGER PRIMARY KEY,
    page_id INTEGER NOT NULL REFERENCES page(id) ON DELETE CASCADE,
    language_code TEXT NOT NULL,
    target_url TEXT NOT NULL,
    validation_state TEXT NOT NULL
) STRICT;

CREATE TABLE IF NOT EXISTS structured_data (
    id INTEGER PRIMARY KEY,
    page_id INTEGER NOT NULL REFERENCES page(id) ON DELETE CASCADE,
    format TEXT NOT NULL,
    type_summary TEXT NOT NULL DEFAULT '',
    evidence_json TEXT NOT NULL CHECK (json_valid(evidence_json))
) STRICT;

CREATE TABLE IF NOT EXISTS sitemap (
    id INTEGER PRIMARY KEY,
    crawl_id TEXT NOT NULL REFERENCES crawl(id) ON DELETE CASCADE,
    url TEXT NOT NULL,
    status TEXT NOT NULL,
    discovered_from TEXT NOT NULL,
    UNIQUE(crawl_id, url)
) STRICT;

CREATE TABLE IF NOT EXISTS sitemap_entry (
    sitemap_id INTEGER NOT NULL REFERENCES sitemap(id) ON DELETE CASCADE,
    url_id INTEGER NOT NULL REFERENCES url(id) ON DELETE CASCADE,
    last_modified TEXT,
    PRIMARY KEY(sitemap_id, url_id)
) STRICT;

CREATE TABLE IF NOT EXISTS issue (
    id INTEGER PRIMARY KEY,
    crawl_id TEXT NOT NULL REFERENCES crawl(id) ON DELETE CASCADE,
    rule_id TEXT NOT NULL,
    rule_version INTEGER NOT NULL CHECK (rule_version > 0),
    subject_type TEXT NOT NULL,
    subject_id TEXT NOT NULL,
    severity TEXT NOT NULL CHECK (severity IN ('info','warning','error')),
    evidence_json TEXT NOT NULL CHECK (json_valid(evidence_json)),
    created_at TEXT NOT NULL
) STRICT;

CREATE INDEX IF NOT EXISTS idx_issue_query ON issue(crawl_id, severity, rule_id, id);

CREATE TABLE IF NOT EXISTS crawl_event (
    id INTEGER PRIMARY KEY,
    crawl_id TEXT NOT NULL REFERENCES crawl(id) ON DELETE CASCADE,
    event_type TEXT NOT NULL,
    payload_json TEXT NOT NULL CHECK (json_valid(payload_json)),
    created_at TEXT NOT NULL
) STRICT;

CREATE INDEX IF NOT EXISTS idx_crawl_event_cursor ON crawl_event(crawl_id, id);

CREATE TABLE IF NOT EXISTS artifact (
    id TEXT PRIMARY KEY,
    crawl_id TEXT NOT NULL REFERENCES crawl(id) ON DELETE CASCADE,
    format TEXT NOT NULL,
    relative_path TEXT NOT NULL,
    checksum TEXT NOT NULL,
    size_bytes INTEGER NOT NULL CHECK (size_bytes >= 0),
    created_at TEXT NOT NULL,
    expires_at TEXT
) STRICT;

