package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/seo-auditor/seo-auditor/internal/contracts"
	"github.com/seo-auditor/seo-auditor/internal/customaudit"
)

type CustomAuditDefinitionRecord struct {
	ProjectID  contracts.ID           `json:"project_id"`
	Definition customaudit.Definition `json:"definition"`
	CreatedAt  string                 `json:"created_at"`
	UpdatedAt  string                 `json:"updated_at"`
}

type CustomAuditResultRecord struct {
	ID              int64        `json:"id"`
	CrawlID         contracts.ID `json:"crawl_id"`
	PageID          int64        `json:"page_id"`
	DefinitionID    string       `json:"definition_id"`
	SchemaVersion   int          `json:"definition_schema_version"`
	Mode            string       `json:"mode"`
	Values          []string     `json:"values"`
	MatchCount      int          `json:"match_count"`
	ConditionMet    bool         `json:"condition_met"`
	Finding         bool         `json:"finding"`
	Truncated       bool         `json:"truncated"`
	FindingSeverity string       `json:"finding_severity,omitempty"`
	FindingMessage  string       `json:"finding_message,omitempty"`
	CreatedAt       string       `json:"created_at"`
	rowID           int64
}

func (f *Frontier) PutCustomAuditDefinition(ctx context.Context, projectID contracts.ID, definition customaudit.Definition) (CustomAuditDefinitionRecord, error) {
	if err := definition.NormalizeAndValidate(); err != nil {
		return CustomAuditDefinitionRecord{}, err
	}
	if err := f.ProjectExists(ctx, projectID); err != nil {
		return CustomAuditDefinitionRecord{}, err
	}
	body, _ := json.Marshal(definition)
	now := time.Now().UTC().Format(time.RFC3339Nano)
	err := f.writer.Submit(ctx, func(ctx context.Context, tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, `INSERT INTO custom_audit_definition(id,project_id,schema_version,name,enabled,definition_json,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?) ON CONFLICT(project_id,id) DO UPDATE SET schema_version=excluded.schema_version,name=excluded.name,enabled=excluded.enabled,definition_json=excluded.definition_json,updated_at=excluded.updated_at`, definition.ID, projectID, definition.SchemaVersion, definition.Name, definition.Enabled, string(body), now, now)
		return err
	})
	if err != nil {
		return CustomAuditDefinitionRecord{}, err
	}
	return f.GetCustomAuditDefinition(ctx, projectID, definition.ID)
}

func (f *Frontier) GetCustomAuditDefinition(ctx context.Context, projectID contracts.ID, id string) (CustomAuditDefinitionRecord, error) {
	var record CustomAuditDefinitionRecord
	var body string
	err := f.db.QueryRowContext(ctx, `SELECT project_id,definition_json,created_at,updated_at FROM custom_audit_definition WHERE project_id=? AND id=?`, projectID, id).Scan(&record.ProjectID, &body, &record.CreatedAt, &record.UpdatedAt)
	if err != nil {
		return record, err
	}
	err = json.Unmarshal([]byte(body), &record.Definition)
	return record, err
}

func (f *Frontier) ListCustomAuditDefinitions(ctx context.Context, projectID contracts.ID) ([]CustomAuditDefinitionRecord, error) {
	rows, err := f.db.QueryContext(ctx, `SELECT project_id,definition_json,created_at,updated_at FROM custom_audit_definition WHERE project_id=? ORDER BY id LIMIT 100`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []CustomAuditDefinitionRecord
	for rows.Next() {
		var item CustomAuditDefinitionRecord
		var body string
		if err := rows.Scan(&item.ProjectID, &body, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(body), &item.Definition); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func (f *Frontier) DeleteCustomAuditDefinition(ctx context.Context, projectID contracts.ID, id string) error {
	return f.writer.Submit(ctx, func(ctx context.Context, tx *sql.Tx) error {
		result, err := tx.ExecContext(ctx, `DELETE FROM custom_audit_definition WHERE project_id=? AND id=?`, projectID, id)
		if err != nil {
			return err
		}
		n, _ := result.RowsAffected()
		if n != 1 {
			return sql.ErrNoRows
		}
		return nil
	})
}

func (f *Frontier) ListCustomAuditResults(ctx context.Context, crawlID contracts.ID, definitionID string, page contracts.PageRequest) (contracts.Page[CustomAuditResultRecord], error) {
	return listKeyset(ctx, page, func(after int64, limit int) (*sql.Rows, error) {
		if definitionID == "" {
			return f.db.QueryContext(ctx, `SELECT id,crawl_id,page_id,definition_id,definition_schema_version,mode,values_json,match_count,condition_met,finding,truncated,COALESCE(finding_severity,''),COALESCE(finding_message,''),created_at FROM custom_audit_result WHERE crawl_id=? AND id>? ORDER BY id LIMIT ?`, crawlID, after, limit)
		}
		return f.db.QueryContext(ctx, `SELECT id,crawl_id,page_id,definition_id,definition_schema_version,mode,values_json,match_count,condition_met,finding,truncated,COALESCE(finding_severity,''),COALESCE(finding_message,''),created_at FROM custom_audit_result WHERE crawl_id=? AND definition_id=? AND id>? ORDER BY id LIMIT ?`, crawlID, definitionID, after, limit)
	}, func(rows *sql.Rows) (CustomAuditResultRecord, error) {
		var item CustomAuditResultRecord
		var values string
		err := rows.Scan(&item.ID, &item.CrawlID, &item.PageID, &item.DefinitionID, &item.SchemaVersion, &item.Mode, &values, &item.MatchCount, &item.ConditionMet, &item.Finding, &item.Truncated, &item.FindingSeverity, &item.FindingMessage, &item.CreatedAt)
		if err == nil {
			err = json.Unmarshal([]byte(values), &item.Values)
		}
		item.rowID = item.ID
		return item, err
	}, func(item CustomAuditResultRecord) int64 { return item.rowID })
}

func (f *Frontier) SaveCustomAuditResults(ctx context.Context, crawlID contracts.ID, crawlURLID int64, definitions []customaudit.Definition, results []customaudit.Result) error {
	definitionMap := make(map[string]customaudit.Definition, len(definitions))
	for _, definition := range definitions {
		definitionMap[definition.ID] = definition
	}
	return f.writer.Submit(ctx, func(ctx context.Context, tx *sql.Tx) error {
		var pageID int64
		if err := tx.QueryRowContext(ctx, `SELECT id FROM page WHERE crawl_url_id=?`, crawlURLID).Scan(&pageID); err != nil {
			return err
		}
		return saveCustomAuditResultsTx(ctx, tx, crawlID, pageID, definitionMap, results)
	})
}

func saveCustomAuditResultsTx(ctx context.Context, tx *sql.Tx, crawlID contracts.ID, pageID int64, definitions map[string]customaudit.Definition, results []customaudit.Result) error {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	for _, result := range results {
		definition, ok := definitions[result.DefinitionID]
		if !ok {
			return errors.New("custom-audit result has no matching definition")
		}
		if definition.SchemaVersion == 0 {
			definition.SchemaVersion = customaudit.SchemaVersion
		}
		values, _ := json.Marshal(result.Values)
		var severity, message any
		if definition.Finding != nil {
			severity = definition.Finding.Severity
			message = definition.Finding.Message
		}
		_, err := tx.ExecContext(ctx, `INSERT INTO custom_audit_result(crawl_id,page_id,definition_id,definition_schema_version,mode,values_json,match_count,condition_met,finding,truncated,finding_severity,finding_message,created_at) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?) ON CONFLICT(page_id,definition_id,mode) DO UPDATE SET values_json=excluded.values_json,match_count=excluded.match_count,condition_met=excluded.condition_met,finding=excluded.finding,truncated=excluded.truncated,finding_severity=excluded.finding_severity,finding_message=excluded.finding_message,created_at=excluded.created_at`, crawlID, pageID, result.DefinitionID, definition.SchemaVersion, result.Mode, string(values), result.MatchCount, result.ConditionMet, result.Finding, result.Truncated, severity, message, now)
		if err != nil {
			return err
		}
	}
	return nil
}
