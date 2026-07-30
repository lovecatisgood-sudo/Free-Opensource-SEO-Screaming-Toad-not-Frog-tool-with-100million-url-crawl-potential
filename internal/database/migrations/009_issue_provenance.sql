ALTER TABLE issue ADD COLUMN classification TEXT NOT NULL DEFAULT 'review'
CHECK (classification IN ('deterministic','recommendation','review','information'));

ALTER TABLE issue ADD COLUMN evidence_source TEXT NOT NULL DEFAULT 'raw'
CHECK (evidence_source IN ('raw','rendered','graph','sitemap','external_api','lab','field'));

UPDATE issue SET
classification = CASE
    WHEN severity='info' THEN 'information'
    WHEN rule_id IN ('AUD-01','AUD-05','AUD-07','AUD-09','AUD-10','AUD-13') THEN 'deterministic'
    WHEN rule_id='AUD-04' AND json_type(evidence_json,'$.canonical_count') IS NOT NULL THEN 'recommendation'
    WHEN rule_id='AUD-04' THEN 'deterministic'
    WHEN rule_id='AUD-02' AND COALESCE(json_extract(evidence_json,'$.length'),-1)=0 THEN 'deterministic'
    WHEN rule_id='AUD-02' THEN 'recommendation'
    WHEN rule_id='AUD-03' AND json_type(evidence_json,'$.h1_count') IS NOT NULL THEN 'deterministic'
    WHEN rule_id='AUD-03' THEN 'recommendation'
    WHEN rule_id='AUD-08' AND json_type(evidence_json,'$.hamming_distance') IS NULL THEN 'deterministic'
    WHEN rule_id IN ('AUD-08','AUD-11') THEN 'recommendation'
    WHEN rule_id='AUD-12' THEN 'deterministic'
    ELSE 'review'
END,
evidence_source = CASE
    WHEN subject_type='rendered_page' THEN 'rendered'
    WHEN rule_id='AUD-06' OR subject_type='sitemap' THEN 'sitemap'
    WHEN subject_type IN ('link','image','url') OR rule_id IN ('AUD-08','AUD-11') THEN 'graph'
    ELSE 'raw'
END;

CREATE INDEX IF NOT EXISTS idx_issue_classification
ON issue(crawl_id, classification, evidence_source, id);
