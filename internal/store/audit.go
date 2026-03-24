package store

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// AuditParams holds the fields for one audit_logs row.
type AuditParams struct {
	SchoolID  *uuid.UUID
	UserID    *uuid.UUID
	Action    string // INSERT | UPDATE | ARCHIVE | LOGIN | CLOSE
	TableName string
	RecordID  string
	NewValues map[string]any
}

// InsertAudit writes one row to audit_logs.
// audit_logs is INSERT-ONLY (triggers block UPDATE/DELETE).
func InsertAudit(ctx context.Context, tx pgx.Tx, p AuditParams) error {
	nv := toJSONB(p.NewValues)
	_, err := tx.Exec(ctx, `
		INSERT INTO audit_logs
		       (school_id, user_id, action, table_name, record_id, new_values)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, p.SchoolID, p.UserID, p.Action, p.TableName, p.RecordID, nv)
	if err != nil {
		return fmt.Errorf("store.InsertAudit: %w", err)
	}
	return nil
}

// toJSONB converts a Go map to a value pgx can bind as JSONB.
// Returns nil (SQL NULL) when m is nil or empty.
func toJSONB(m map[string]any) any {
	if len(m) == 0 {
		return nil
	}
	return m
}
