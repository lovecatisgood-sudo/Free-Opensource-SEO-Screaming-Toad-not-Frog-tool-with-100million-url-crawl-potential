CREATE TABLE campaign_segment (
    id INTEGER PRIMARY KEY,
    crawl_id TEXT NOT NULL REFERENCES crawl(id) ON DELETE CASCADE,
    sequence INTEGER NOT NULL CHECK (sequence >= 0),
    start_analysed INTEGER NOT NULL CHECK (start_analysed >= 0),
    end_analysed INTEGER CHECK (end_analysed >= start_analysed),
    status TEXT NOT NULL CHECK (status IN ('active','completed')),
    storage_bytes INTEGER NOT NULL DEFAULT 0 CHECK (storage_bytes >= 0),
    checksum TEXT,
    started_at TEXT NOT NULL,
    completed_at TEXT,
    UNIQUE(crawl_id, sequence)
) STRICT;

CREATE INDEX idx_campaign_segment_history ON campaign_segment(crawl_id, sequence);
