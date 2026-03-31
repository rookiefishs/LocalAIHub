package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type RouteState struct {
	ID               int64      `json:"id"`
	VirtualModelID   int64      `json:"virtual_model_id"`
	CurrentBindingID *int64     `json:"current_binding_id,omitempty"`
	RouteStatus      string     `json:"route_status"`
	ManualLocked     bool       `json:"manual_locked"`
	LockUntil        *time.Time `json:"lock_until,omitempty"`
	LastSwitchReason *string    `json:"last_switch_reason,omitempty"`
	LastSwitchAt     *time.Time `json:"last_switch_at,omitempty"`
	UpdatedAt        time.Time  `json:"updated_at"`
	ModelCode        string     `json:"model_code,omitempty"`
}

type RouteRepository struct{ db *sql.DB }

func NewRouteRepository(db *sql.DB) *RouteRepository { return &RouteRepository{db: db} }

func (r *RouteRepository) List(ctx context.Context, page, pageSize int) ([]RouteState, int, error) {
	var total int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM route_state`).Scan(&total); err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	rows, err := r.db.QueryContext(ctx, `SELECT rs.id, rs.virtual_model_id, rs.current_binding_id, rs.route_status, rs.manual_locked, rs.lock_until, rs.last_switch_reason, rs.last_switch_at, rs.updated_at, vm.model_code FROM route_state rs INNER JOIN virtual_model vm ON vm.id = rs.virtual_model_id ORDER BY rs.updated_at DESC LIMIT ? OFFSET ?`, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	items := make([]RouteState, 0)
	for rows.Next() {
		item, err := scanRouteState(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, *item)
	}
	return items, total, rows.Err()
}

func (r *RouteRepository) GetByVirtualModelID(ctx context.Context, virtualModelID int64) (*RouteState, error) {
	row := r.db.QueryRowContext(ctx, `SELECT rs.id, rs.virtual_model_id, rs.current_binding_id, rs.route_status, rs.manual_locked, rs.lock_until, rs.last_switch_reason, rs.last_switch_at, rs.updated_at, vm.model_code FROM route_state rs INNER JOIN virtual_model vm ON vm.id = rs.virtual_model_id WHERE rs.virtual_model_id = ? LIMIT 1`, virtualModelID)
	return scanRouteState(row)
}

func (r *RouteRepository) Switch(ctx context.Context, virtualModelID int64, bindingID int64, manualLock bool, lockUntil *time.Time, reason string, adminID int64) error {
	now := time.Now().UTC()
	status := "switched"
	if manualLock {
		status = "manual_locked"
	}
	_, err := r.db.ExecContext(ctx, `UPDATE route_state SET current_binding_id = ?, route_status = ?, manual_locked = ?, lock_until = ?, last_switch_reason = ?, last_switch_at = ?, updated_at = ? WHERE virtual_model_id = ?`, bindingID, status, manualLock, lockUntil, reason, now, now, virtualModelID)
	if err != nil {
		return fmt.Errorf("update route state: %w", err)
	}
	_, err = r.db.ExecContext(ctx, `INSERT INTO route_switch_log (virtual_model_id, from_binding_id, to_binding_id, trigger_type, operator_admin_id, reason, created_at) VALUES (?, NULL, ?, 'manual', ?, ?, ?)`, virtualModelID, bindingID, adminID, reason, now)
	if err != nil {
		return fmt.Errorf("insert route switch log: %w", err)
	}
	return nil
}

func (r *RouteRepository) Unlock(ctx context.Context, virtualModelID int64) error {
	_, err := r.db.ExecContext(ctx, `UPDATE route_state SET manual_locked = 0, route_status = 'normal', lock_until = NULL, updated_at = ? WHERE virtual_model_id = ?`, time.Now().UTC(), virtualModelID)
	return err
}

func (r *RouteRepository) CountOpenCircuits(ctx context.Context) (int64, error) {
	var count int64
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM circuit_breaker_state WHERE state = 'open'`).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (r *RouteRepository) RegisterFailure(ctx context.Context, providerID, virtualModelID int64, reason string) (bool, error) {
	_, err := r.db.ExecContext(ctx, `INSERT INTO circuit_breaker_state (provider_id, virtual_model_id, state, failure_count, success_count, last_reason, updated_at) VALUES (?, ?, 'closed', 0, 0, NULL, ?) ON DUPLICATE KEY UPDATE updated_at = VALUES(updated_at)`, providerID, virtualModelID, time.Now().UTC())
	if err != nil {
		return false, err
	}
	_, err = r.db.ExecContext(ctx, `UPDATE circuit_breaker_state SET failure_count = failure_count + 1, last_failure_at = ?, last_reason = ?, state = CASE WHEN failure_count + 1 >= 5 THEN 'open' ELSE state END, next_retry_at = CASE WHEN failure_count + 1 >= 5 THEN DATE_ADD(?, INTERVAL 30 SECOND) ELSE next_retry_at END, updated_at = ? WHERE provider_id = ? AND virtual_model_id = ?`, time.Now().UTC(), reason, time.Now().UTC(), time.Now().UTC(), providerID, virtualModelID)
	if err != nil {
		return false, err
	}
	var state string
	if err := r.db.QueryRowContext(ctx, `SELECT state FROM circuit_breaker_state WHERE provider_id = ? AND virtual_model_id = ?`, providerID, virtualModelID).Scan(&state); err != nil {
		return false, err
	}
	return state == "open", nil
}

func (r *RouteRepository) RegisterSuccess(ctx context.Context, providerID, virtualModelID int64) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO circuit_breaker_state (provider_id, virtual_model_id, state, failure_count, success_count, updated_at) VALUES (?, ?, 'closed', 0, 0, ?) ON DUPLICATE KEY UPDATE state = 'closed', failure_count = 0, success_count = success_count + 1, updated_at = VALUES(updated_at)`, providerID, virtualModelID, time.Now().UTC())
	return err
}

func (r *RouteRepository) IsCircuitOpen(ctx context.Context, providerID, virtualModelID int64) (bool, error) {
	var state string
	var nextRetry sql.NullTime
	err := r.db.QueryRowContext(ctx, `SELECT state, next_retry_at FROM circuit_breaker_state WHERE provider_id = ? AND virtual_model_id = ? LIMIT 1`, providerID, virtualModelID).Scan(&state, &nextRetry)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	if state != "open" {
		return false, nil
	}
	if nextRetry.Valid && nextRetry.Time.Before(time.Now().UTC()) {
		_, _ = r.db.ExecContext(ctx, `UPDATE circuit_breaker_state SET state = 'half_open', updated_at = ? WHERE provider_id = ? AND virtual_model_id = ?`, time.Now().UTC(), providerID, virtualModelID)
		return false, nil
	}
	return true, nil
}

func (r *RouteRepository) Delete(ctx context.Context, virtualModelID int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM route_state WHERE virtual_model_id = ?`, virtualModelID)
	return err
}

func scanRouteState(scanner interface{ Scan(dest ...any) error }) (*RouteState, error) {
	var item RouteState
	var currentBindingID sql.NullInt64
	var lockUntil, lastSwitchAt sql.NullTime
	var lastSwitchReason sql.NullString
	err := scanner.Scan(&item.ID, &item.VirtualModelID, &currentBindingID, &item.RouteStatus, &item.ManualLocked, &lockUntil, &lastSwitchReason, &lastSwitchAt, &item.UpdatedAt, &item.ModelCode)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if currentBindingID.Valid {
		item.CurrentBindingID = &currentBindingID.Int64
	}
	if lockUntil.Valid {
		item.LockUntil = &lockUntil.Time
	}
	if lastSwitchAt.Valid {
		item.LastSwitchAt = &lastSwitchAt.Time
	}
	if lastSwitchReason.Valid {
		item.LastSwitchReason = &lastSwitchReason.String
	}
	return &item, nil
}
