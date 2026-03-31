package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

type AuditLog struct {
	AdminUserID       int64
	Action            string
	TargetType        string
	TargetID          *int64
	ChangeSummaryJSON json.RawMessage
	IPAddress         string
	UserAgent         string
	RequestID         string
}

type AuditRepository struct{ db *sql.DB }

func NewAuditRepository(db *sql.DB) *AuditRepository { return &AuditRepository{db: db} }

func (r *AuditRepository) Create(ctx context.Context, item AuditLog) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO audit_log (admin_user_id, action, target_type, target_id, change_summary_json, ip_address, user_agent, request_id, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`, item.AdminUserID, item.Action, item.TargetType, nullableInt64(item.TargetID), normalizeJSON(item.ChangeSummaryJSON), nullableString(item.IPAddress), nullableString(item.UserAgent), nullableString(item.RequestID), time.Now().UTC())
	if err != nil {
		return fmt.Errorf("insert audit log: %w", err)
	}
	return nil
}

func normalizeJSON(raw json.RawMessage) any {
	if len(raw) == 0 {
		return nil
	}
	return string(raw)
}

func nullableInt64(value *int64) any {
	if value == nil {
		return nil
	}
	return *value
}

func nullableString(value string) any {
	if value == "" {
		return nil
	}
	return value
}
