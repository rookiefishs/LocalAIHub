package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"
)

type AuditLogRecord struct {
	ID                int64           `json:"id"`
	AdminUserID       int64           `json:"admin_user_id"`
	Action            string          `json:"action"`
	TargetType        string          `json:"target_type"`
	TargetID          *int64          `json:"target_id,omitempty"`
	ChangeSummaryJSON json.RawMessage `json:"change_summary_json,omitempty"`
	IPAddress         *string         `json:"ip_address,omitempty"`
	UserAgent         *string         `json:"user_agent,omitempty"`
	RequestID         *string         `json:"request_id,omitempty"`
	CreatedAt         time.Time       `json:"created_at"`
}

type AuditLogFilters struct {
	AdminUserID *int64
	Action      string
	TargetType  string
	TargetID    *int64
	StartTime   *time.Time
	EndTime     *time.Time
	Limit       int
	Page        int
}

type LogRepository struct{ db *sql.DB }

func NewLogRepository(db *sql.DB) *LogRepository { return &LogRepository{db: db} }

func (r *LogRepository) ListAuditLogs(ctx context.Context, filters AuditLogFilters) ([]AuditLogRecord, int, error) {
	countQuery := `SELECT COUNT(*) FROM audit_log WHERE 1=1`
	query := `SELECT id, admin_user_id, action, target_type, target_id, change_summary_json, ip_address, user_agent, request_id, created_at FROM audit_log WHERE 1=1`
	args := make([]any, 0)
	if filters.AdminUserID != nil {
		countQuery += ` AND admin_user_id = ?`
		query += ` AND admin_user_id = ?`
		args = append(args, *filters.AdminUserID)
	}
	if filters.Action != "" {
		countQuery += ` AND action = ?`
		query += ` AND action = ?`
		args = append(args, filters.Action)
	}
	if filters.TargetType != "" {
		countQuery += ` AND target_type = ?`
		query += ` AND target_type = ?`
		args = append(args, filters.TargetType)
	}
	if filters.TargetID != nil {
		countQuery += ` AND target_id = ?`
		query += ` AND target_id = ?`
		args = append(args, *filters.TargetID)
	}
	if filters.StartTime != nil {
		countQuery += ` AND created_at >= ?`
		query += ` AND created_at >= ?`
		args = append(args, *filters.StartTime)
	}
	if filters.EndTime != nil {
		countQuery += ` AND created_at <= ?`
		query += ` AND created_at <= ?`
		args = append(args, *filters.EndTime)
	}
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return []AuditLogRecord{}, 0, nil
	}
	offset := (filters.Page - 1) * filters.Limit
	query += ` ORDER BY created_at DESC LIMIT ? OFFSET ?`
	args = append(args, filters.Limit, offset)
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	items := make([]AuditLogRecord, 0)
	for rows.Next() {
		var item AuditLogRecord
		var targetID sql.NullInt64
		var changeSummary sql.NullString
		var ipAddress, userAgent, requestID sql.NullString
		if err := rows.Scan(&item.ID, &item.AdminUserID, &item.Action, &item.TargetType, &targetID, &changeSummary, &ipAddress, &userAgent, &requestID, &item.CreatedAt); err != nil {
			return nil, 0, err
		}
		if targetID.Valid {
			v := targetID.Int64
			item.TargetID = &v
		}
		if changeSummary.Valid {
			item.ChangeSummaryJSON = json.RawMessage(changeSummary.String)
		}
		if ipAddress.Valid {
			v := ipAddress.String
			item.IPAddress = &v
		}
		if userAgent.Valid {
			v := userAgent.String
			item.UserAgent = &v
		}
		if requestID.Valid {
			v := requestID.String
			item.RequestID = &v
		}
		items = append(items, item)
	}
	return items, total, rows.Err()
}
