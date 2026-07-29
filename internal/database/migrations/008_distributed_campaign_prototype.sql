CREATE TABLE coordinator_lease (
    crawl_id TEXT PRIMARY KEY REFERENCES crawl(id) ON DELETE CASCADE,
    owner TEXT NOT NULL,
    epoch INTEGER NOT NULL CHECK (epoch > 0),
    expires_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
) STRICT;

CREATE TABLE host_owner_lease (
    crawl_id TEXT NOT NULL REFERENCES crawl(id) ON DELETE CASCADE,
    host TEXT NOT NULL,
    owner TEXT NOT NULL,
    epoch INTEGER NOT NULL CHECK (epoch > 0),
    expires_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    PRIMARY KEY(crawl_id, host)
) STRICT;

CREATE TABLE immutable_result_segment (
    id INTEGER PRIMARY KEY,
    crawl_id TEXT NOT NULL REFERENCES crawl(id) ON DELETE CASCADE,
    sequence INTEGER NOT NULL CHECK (sequence >= 0),
    object_key TEXT NOT NULL,
    checksum TEXT NOT NULL CHECK (length(checksum) = 64),
    row_count INTEGER NOT NULL CHECK (row_count > 0),
    schema_version INTEGER NOT NULL CHECK (schema_version > 0),
    rule_version INTEGER NOT NULL CHECK (rule_version > 0),
    created_at TEXT NOT NULL,
    UNIQUE(crawl_id, sequence),
    UNIQUE(crawl_id, object_key)
) STRICT;

CREATE INDEX idx_result_segment_verify ON immutable_result_segment(crawl_id, sequence, checksum);
