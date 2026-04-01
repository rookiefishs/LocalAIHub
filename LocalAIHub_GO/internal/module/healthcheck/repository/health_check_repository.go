package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type HealthCheckLog struct {
	ID             int64     `json:"id"`
	TargetType     string    `json:"target_type"`
	TargetID       int64     `json:"target_id"`
	TargetName     string    `json:"target_name"`
	CheckStatus    string    `json:"check_status"`
	PreviousStatus string    `json:"previous_status"`
	ErrorMessage   string    `json:"error_message"`
	LatencyMs      int       `json:"latency_ms"`
	CreatedAt      time.Time `json:"created_at"`
}

type HealthCheckRepository struct{ db *sql.DB }

func NewHealthCheckRepository(db *sql.DB) *HealthCheckRepository {
	return &HealthCheckRepository{db: db}
}

func (r *HealthCheckRepository) Create(ctx context.Context, log *HealthCheckLog) (int64, error) {
	result, err := r.db.ExecContext(ctx, `
		INSERT INTO health_check_log (target_type, target_id, target_name, check_status, previous_status, error_message, latency_ms, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, log.TargetType, log.TargetID, log.TargetName, log.CheckStatus, log.PreviousStatus, log.ErrorMessage, log.LatencyMs, time.Now().UTC())
	if err != nil {
		return 0, fmt.Errorf("create health check log: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (r *HealthCheckRepository) List(ctx context.Context, targetType string, targetID int64, page, pageSize int) ([]HealthCheckLog, int, error) {
	var total int
	baseWhere := ""
	if targetType != "" {
		baseWhere = fmt.Sprintf("WHERE target_type = '%s'", targetType)
	}
	if targetID > 0 {
		if baseWhere == "" {
			baseWhere = fmt.Sprintf("WHERE target_id = %d", targetID)
		} else {
			baseWhere += fmt.Sprintf(" AND target_id = %d", targetID)
		}
	}
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM health_check_log %s", baseWhere)
	if err := r.db.QueryRowContext(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	query := fmt.Sprintf(`
		SELECT id, target_type, target_id, target_name, check_status, previous_status, error_message, latency_ms, created_at
		FROM health_check_log %s
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`, baseWhere)

	rows, err := r.db.QueryContext(ctx, query, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]HealthCheckLog, 0)
	for rows.Next() {
		var item HealthCheckLog
		if err := rows.Scan(&item.ID, &item.TargetType, &item.TargetID, &item.TargetName, &item.CheckStatus, &item.PreviousStatus, &item.ErrorMessage, &item.LatencyMs, &item.CreatedAt); err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func (r *HealthCheckRepository) DeleteOlderThan(ctx context.Context, days int) error {
	cutoff := time.Now().UTC().AddDate(0, 0, -days)
	_, err := r.db.ExecContext(ctx, "DELETE FROM health_check_log WHERE created_at < ?", cutoff)
	return err
}
