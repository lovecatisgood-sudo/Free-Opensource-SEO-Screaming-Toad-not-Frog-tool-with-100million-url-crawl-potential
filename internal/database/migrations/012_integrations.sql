CREATE TABLE integration_observation (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL REFERENCES project(id) ON DELETE CASCADE,
    crawl_id TEXT REFERENCES crawl(id) ON DELETE CASCADE,
    provider TEXT NOT NULL CHECK (provider IN ('pagespeed-insights','chrome-ux-report','google-search-console','google-analytics-4','lighthouse')),
    evidence_source TEXT NOT NULL CHECK (evidence_source IN ('external_api','lab','field')),
    profile_version TEXT NOT NULL,
    scope TEXT NOT NULL,
    freshness TEXT NOT NULL DEFAULT '',
    result_json TEXT NOT NULL CHECK (json_valid(result_json)),
    observed_at TEXT NOT NULL,
    created_at TEXT NOT NULL
) STRICT;

CREATE INDEX idx_integration_observation_query ON integration_observation(project_id,provider,created_at,id);
CREATE INDEX idx_integration_observation_crawl ON integration_observation(crawl_id,provider,id);
