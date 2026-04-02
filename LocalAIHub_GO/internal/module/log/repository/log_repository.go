package repository

import (
	"context"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"
)

type AuditLogRecord struct {
	ID                int64           `json:"id"`
	AdminUserID       int64           `json:"admin_user_id"`
	AdminUsername     string          `json:"admin_username"`
	Action            string          `json:"action"`
	TargetType        string          `json:"target_type"`
	TargetID          *int64          `json:"target_id,omitempty"`
	TargetName        string          `json:"target_name,omitempty"`
	ChangeSummary     string          `json:"change_summary,omitempty"`
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
	Keyword     string
	StartTime   *time.Time
	EndTime     *time.Time
	Limit       int
	Page        int
}

type LogRepository struct{ db *sql.DB }

func NewLogRepository(db *sql.DB) *LogRepository { return &LogRepository{db: db} }

func (r *LogRepository) ListAuditLogs(ctx context.Context, filters AuditLogFilters) ([]AuditLogRecord, int, error) {
	whereClause, args := buildAuditWhere(filters)
	countQuery := `SELECT COUNT(*) FROM audit_log a LEFT JOIN admin_user au ON au.id = a.admin_user_id ` + whereClause
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return []AuditLogRecord{}, 0, nil
	}
	offset := (filters.Page - 1) * filters.Limit
	query := auditSelectSQL() + whereClause + ` ORDER BY a.created_at DESC LIMIT ? OFFSET ?`
	queryArgs := append(append([]any{}, args...), filters.Limit, offset)
	rows, err := r.db.QueryContext(ctx, query, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	items := make([]AuditLogRecord, 0)
	for rows.Next() {
		item, err := scanAuditLog(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func (r *LogRepository) GetAuditLogByID(ctx context.Context, id int64) (*AuditLogRecord, error) {
	query := auditSelectSQL() + ` WHERE a.id = ? LIMIT 1`
	row := r.db.QueryRowContext(ctx, query, id)
	item, err := scanAuditLog(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

func (r *LogRepository) ExportAuditLogsCSV(ctx context.Context, filters AuditLogFilters, writer io.Writer) error {
	query := auditSelectSQL() + buildAuditWhereClause(filters) + ` ORDER BY a.created_at DESC`
	rows, err := r.db.QueryContext(ctx, query, buildAuditArgs(filters)...)
	if err != nil {
		return err
	}
	defer rows.Close()

	csvWriter := csv.NewWriter(writer)
	if err := csvWriter.Write([]string{"ID", "时间", "操作人", "动作", "对象类型", "对象ID", "对象名称", "摘要", "IP", "RequestID"}); err != nil {
		return err
	}
	for rows.Next() {
		item, scanErr := scanAuditLog(rows)
		if scanErr != nil {
			return scanErr
		}
		targetID := ""
		if item.TargetID != nil {
			targetID = fmt.Sprintf("%d", *item.TargetID)
		}
		ip := ""
		if item.IPAddress != nil {
			ip = *item.IPAddress
		}
		requestID := ""
		if item.RequestID != nil {
			requestID = *item.RequestID
		}
		if err := csvWriter.Write([]string{
			fmt.Sprintf("%d", item.ID),
			item.CreatedAt.Format(time.RFC3339),
			item.AdminUsername,
			item.Action,
			item.TargetType,
			targetID,
			item.TargetName,
			item.ChangeSummary,
			ip,
			requestID,
		}); err != nil {
			return err
		}
	}
	csvWriter.Flush()
	if err := rows.Err(); err != nil {
		return err
	}
	return csvWriter.Error()
}

func auditSelectSQL() string {
	return `SELECT a.id, a.admin_user_id, COALESCE(au.username, CONCAT('admin#', a.admin_user_id)) AS admin_username, a.action, a.target_type, a.target_id, a.change_summary_json, a.ip_address, a.user_agent, a.request_id, a.created_at FROM audit_log a LEFT JOIN admin_user au ON au.id = a.admin_user_id`
}

func buildAuditWhere(filters AuditLogFilters) (string, []any) {
	return buildAuditWhereClause(filters), buildAuditArgs(filters)
}

func buildAuditWhereClause(filters AuditLogFilters) string {
	whereClause, _ := buildAuditWhereParts(filters)
	return whereClause
}

func buildAuditArgs(filters AuditLogFilters) []any {
	_, args := buildAuditWhereParts(filters)
	return args
}

func buildAuditWhereParts(filters AuditLogFilters) (string, []any) {
	clauses := []string{"WHERE 1=1"}
	args := make([]any, 0)
	if filters.AdminUserID != nil {
		clauses = append(clauses, `AND a.admin_user_id = ?`)
		args = append(args, *filters.AdminUserID)
	}
	if filters.Action != "" {
		clauses = append(clauses, `AND a.action = ?`)
		args = append(args, filters.Action)
	}
	if filters.TargetType != "" {
		clauses = append(clauses, `AND a.target_type = ?`)
		args = append(args, filters.TargetType)
	}
	if filters.TargetID != nil {
		clauses = append(clauses, `AND a.target_id = ?`)
		args = append(args, *filters.TargetID)
	}
	if filters.StartTime != nil {
		clauses = append(clauses, `AND a.created_at >= ?`)
		args = append(args, *filters.StartTime)
	}
	if filters.EndTime != nil {
		clauses = append(clauses, `AND a.created_at <= ?`)
		args = append(args, *filters.EndTime)
	}
	if keyword := strings.TrimSpace(filters.Keyword); keyword != "" {
		like := "%" + keyword + "%"
		clauses = append(clauses, `AND (au.username LIKE ? OR a.action LIKE ? OR a.target_type LIKE ? OR CAST(a.target_id AS CHAR) LIKE ? OR a.request_id LIKE ? OR a.ip_address LIKE ? OR CAST(a.change_summary_json AS CHAR) LIKE ?)`)
		args = append(args, like, like, like, like, like, like, like)
	}
	return " " + strings.Join(clauses, " "), args
}

type auditScanner interface {
	Scan(dest ...any) error
}

func scanAuditLog(scanner auditScanner) (AuditLogRecord, error) {
	var item AuditLogRecord
	var targetID sql.NullInt64
	var changeSummary sql.NullString
	var ipAddress, userAgent, requestID sql.NullString
	if err := scanner.Scan(&item.ID, &item.AdminUserID, &item.AdminUsername, &item.Action, &item.TargetType, &targetID, &changeSummary, &ipAddress, &userAgent, &requestID, &item.CreatedAt); err != nil {
		return item, err
	}
	if targetID.Valid {
		v := targetID.Int64
		item.TargetID = &v
	}
	if changeSummary.Valid {
		item.ChangeSummaryJSON = json.RawMessage(changeSummary.String)
		item.TargetName = extractTargetName(item.ChangeSummaryJSON)
		item.ChangeSummary = buildChangeSummary(item.Action, item.TargetType, item.TargetID, item.TargetName, item.ChangeSummaryJSON)
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
	return item, nil
}

func extractTargetName(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		return ""
	}
	for _, key := range []string{"name", "username", "model_code", "key_prefix", "tested_url"} {
		if value, ok := payload[key].(string); ok && strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func buildChangeSummary(action, targetType string, targetID *int64, targetName string, raw json.RawMessage) string {
	base := action
	if targetType != "" {
		base += " " + targetType
	}
	if targetName != "" {
		return base + " " + targetName
	}
	if targetID != nil {
		return fmt.Sprintf("%s #%d", base, *targetID)
	}
	if len(raw) == 0 {
		return base
	}
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		return base
	}
	for _, key := range []string{"status", "result", "reason", "mode"} {
		if value, ok := payload[key].(string); ok && strings.TrimSpace(value) != "" {
			return base + " - " + value
		}
	}
	return base
}
