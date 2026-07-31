CREATE TABLE custom_audit_definition (
    id TEXT NOT NULL,
    project_id TEXT NOT NULL REFERENCES project(id) ON DELETE CASCADE,
    schema_version INTEGER NOT NULL CHECK (schema_version > 0),
    name TEXT NOT NULL CHECK (length(name) BETWEEN 1 AND 200),
    enabled INTEGER NOT NULL CHECK (enabled IN (0,1)),
    definition_json TEXT NOT NULL CHECK (json_valid(definition_json)),
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    PRIMARY KEY(project_id,id)
) STRICT;

CREATE TABLE custom_audit_result (
    id INTEGER PRIMARY KEY,
    crawl_id TEXT NOT NULL REFERENCES crawl(id) ON DELETE CASCADE,
    page_id INTEGER NOT NULL REFERENCES page(id) ON DELETE CASCADE,
    definition_id TEXT NOT NULL,
    definition_schema_version INTEGER NOT NULL CHECK (definition_schema_version > 0),
    mode TEXT NOT NULL CHECK (mode IN ('raw','rendered')),
    values_json TEXT NOT NULL CHECK (json_valid(values_json)),
    match_count INTEGER NOT NULL CHECK (match_count >= 0),
    condition_met INTEGER NOT NULL CHECK (condition_met IN (0,1)),
    finding INTEGER NOT NULL CHECK (finding IN (0,1)),
    truncated INTEGER NOT NULL CHECK (truncated IN (0,1)),
    finding_severity TEXT,
    finding_message TEXT,
    created_at TEXT NOT NULL,
    UNIQUE(page_id,definition_id,mode)
) STRICT;

CREATE INDEX idx_custom_audit_result_query ON custom_audit_result(crawl_id,definition_id,condition_met,id);
