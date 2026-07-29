ALTER TABLE page ADD COLUMN similarity_hash TEXT;

CREATE TABLE page_similarity_band (
    page_id INTEGER NOT NULL REFERENCES page(id) ON DELETE CASCADE,
    band INTEGER NOT NULL CHECK (band BETWEEN 0 AND 3),
    value TEXT NOT NULL,
    PRIMARY KEY(page_id, band)
) STRICT;

CREATE INDEX idx_page_similarity_band ON page_similarity_band(band, value, page_id);
