CREATE TABLE scheduled_audit (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL REFERENCES project(id) ON DELETE CASCADE,
    profile_id TEXT NOT NULL REFERENCES crawl_profile(id) ON DELETE CASCADE,
    name TEXT NOT NULL CHECK (length(name) BETWEEN 1 AND 200),
    interval_seconds INTEGER NOT NULL CHECK (interval_seconds BETWEEN 900 AND 2592000),
    enabled INTEGER NOT NULL DEFAULT 1 CHECK (enabled IN (0,1)),
    next_run_at TEXT NOT NULL,
    last_run_at TEXT NOT NULL DEFAULT '',
    last_crawl_id TEXT REFERENCES crawl(id) ON DELETE SET NULL,
    last_error TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
) STRICT;

CREATE INDEX idx_scheduled_audit_due ON scheduled_audit(enabled,next_run_at,id);
CREATE INDEX idx_scheduled_audit_project ON scheduled_audit(project_id,id);
