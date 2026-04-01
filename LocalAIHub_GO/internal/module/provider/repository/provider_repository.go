package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type Provider struct {
	ID                int64      `json:"id"`
	Name              string     `json:"name"`
	ProviderType      string     `json:"provider_type"`
	ServiceType       string     `json:"service_type"`
	BaseURL           string     `json:"base_url"`
	AuthType          string     `json:"auth_type"`
	TimeoutMS         int        `json:"timeout_ms"`
	Enabled           bool       `json:"enabled"`
	HealthStatus      string     `json:"health_status"`
	LastHealthCheckAt *time.Time `json:"last_health_check_at,omitempty"`
	LastHealthMessage *string    `json:"last_health_message,omitempty"`
	Remark            *string    `json:"remark,omitempty"`
	NewKey            string     `json:"new_key,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

type ProviderRepository struct{ db *sql.DB }

type ProviderDeleteCheck struct {
	BindingCount int64 `json:"binding_count"`
	KeyCount     int64 `json:"key_count"`
}

func NewProviderRepository(db *sql.DB) *ProviderRepository { return &ProviderRepository{db: db} }

func (r *ProviderRepository) List(ctx context.Context, page, pageSize int) ([]Provider, int, error) {
	var total int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM provider`).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count providers: %w", err)
	}
	offset := (page - 1) * pageSize
	rows, err := r.db.QueryContext(ctx, `SELECT id, name, provider_type, service_type, base_url, auth_type, timeout_ms, enabled, health_status, last_health_check_at, last_health_message, remark, created_at, updated_at FROM provider ORDER BY id DESC LIMIT ? OFFSET ?`, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list providers: %w", err)
	}
	defer rows.Close()
	items := make([]Provider, 0)
	for rows.Next() {
		var item Provider
		var lastCheck sql.NullTime
		var lastMessage, remark sql.NullString
		if err := rows.Scan(&item.ID, &item.Name, &item.ProviderType, &item.ServiceType, &item.BaseURL, &item.AuthType, &item.TimeoutMS, &item.Enabled, &item.HealthStatus, &lastCheck, &lastMessage, &remark, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan provider: %w", err)
		}
		if lastCheck.Valid {
			item.LastHealthCheckAt = &lastCheck.Time
		}
		if lastMessage.Valid {
			item.LastHealthMessage = &lastMessage.String
		}
		if remark.Valid {
			item.Remark = &remark.String
		}
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func (r *ProviderRepository) GetByID(ctx context.Context, id int64) (*Provider, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, name, provider_type, service_type, base_url, auth_type, timeout_ms, enabled, health_status, last_health_check_at, last_health_message, remark, created_at, updated_at FROM provider WHERE id = ?`, id)
	var item Provider
	var lastCheck sql.NullTime
	var lastMessage, remark sql.NullString
	if err := row.Scan(&item.ID, &item.Name, &item.ProviderType, &item.ServiceType, &item.BaseURL, &item.AuthType, &item.TimeoutMS, &item.Enabled, &item.HealthStatus, &lastCheck, &lastMessage, &remark, &item.CreatedAt, &item.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get provider by id: %w", err)
	}
	if lastCheck.Valid {
		item.LastHealthCheckAt = &lastCheck.Time
	}
	if lastMessage.Valid {
		item.LastHealthMessage = &lastMessage.String
	}
	if remark.Valid {
		item.Remark = &remark.String
	}
	return &item, nil
}

func (r *ProviderRepository) Create(ctx context.Context, item *Provider) (int64, error) {
	result, err := r.db.ExecContext(ctx, `INSERT INTO provider (name, provider_type, service_type, base_url, auth_type, timeout_ms, enabled, health_status, remark, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, item.Name, item.ProviderType, item.ServiceType, item.BaseURL, item.AuthType, item.TimeoutMS, item.Enabled, item.HealthStatus, item.Remark, time.Now().UTC(), time.Now().UTC())
	if err != nil {
		return 0, fmt.Errorf("create provider: %w", err)
	}
	return result.LastInsertId()
}

func (r *ProviderRepository) Update(ctx context.Context, item *Provider) error {
	_, err := r.db.ExecContext(ctx, `UPDATE provider SET name = ?, provider_type = ?, service_type = ?, base_url = ?, auth_type = ?, timeout_ms = ?, enabled = ?, health_status = ?, remark = ?, updated_at = ? WHERE id = ?`, item.Name, item.ProviderType, item.ServiceType, item.BaseURL, item.AuthType, item.TimeoutMS, item.Enabled, item.HealthStatus, item.Remark, time.Now().UTC(), item.ID)
	if err != nil {
		return fmt.Errorf("update provider: %w", err)
	}
	return nil
}

func (r *ProviderRepository) UpdateStatus(ctx context.Context, id int64, enabled bool) error {
	_, err := r.db.ExecContext(ctx, `UPDATE provider SET enabled = ?, updated_at = ? WHERE id = ?`, enabled, time.Now().UTC(), id)
	if err != nil {
		return fmt.Errorf("update provider status: %w", err)
	}
	return nil
}

func (r *ProviderRepository) CountActive(ctx context.Context) (int64, error) {
	var count int64
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM provider WHERE enabled = 1`).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (r *ProviderRepository) UpdateHealth(ctx context.Context, id int64, status, message string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE provider SET health_status = ?, last_health_check_at = ?, last_health_message = ?, updated_at = ? WHERE id = ?`, status, time.Now().UTC(), message, time.Now().UTC(), id)
	if err != nil {
		return fmt.Errorf("update provider health: %w", err)
	}
	return nil
}

func (r *ProviderRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM provider WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete provider: %w", err)
	}
	return nil
}

func (r *ProviderRepository) CheckDelete(ctx context.Context, id int64) (*ProviderDeleteCheck, error) {
	var result ProviderDeleteCheck
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM virtual_model_binding WHERE provider_id = ?`, id).Scan(&result.BindingCount); err != nil {
		return nil, fmt.Errorf("count provider bindings: %w", err)
	}
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM provider_key WHERE provider_id = ?`, id).Scan(&result.KeyCount); err != nil {
		return nil, fmt.Errorf("count provider keys: %w", err)
	}
	return &result, nil
}

func (r *ProviderRepository) ListEnabled(ctx context.Context) ([]Provider, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, name, provider_type, service_type, base_url, auth_type, timeout_ms, enabled, health_status, last_health_check_at, last_health_message, remark, created_at, updated_at FROM provider WHERE enabled = 1`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]Provider, 0)
	for rows.Next() {
		var item Provider
		var lastCheck sql.NullTime
		var lastMessage, remark sql.NullString
		if err := rows.Scan(&item.ID, &item.Name, &item.ProviderType, &item.ServiceType, &item.BaseURL, &item.AuthType, &item.TimeoutMS, &item.Enabled, &item.HealthStatus, &lastCheck, &lastMessage, &remark, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		if lastCheck.Valid {
			item.LastHealthCheckAt = &lastCheck.Time
		}
		if lastMessage.Valid {
			item.LastHealthMessage = &lastMessage.String
		}
		if remark.Valid {
			item.Remark = &remark.String
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *ProviderRepository) GetEnabled(ctx context.Context) (bool, error) {
	var enabled bool
	err := r.db.QueryRowContext(ctx, `SELECT enabled FROM provider WHERE id = ?`).Scan(&enabled)
	return enabled, err
}

func (r *ProviderRepository) ListKeys(ctx context.Context, providerID int64) ([]ProviderKey, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, provider_id, key_masked, secret_encrypted, status, priority, fail_count, last_used_at, last_error_at, last_error_message, remark, created_at, updated_at FROM provider_key WHERE provider_id = ? AND status = 'enabled' ORDER BY priority ASC`, providerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]ProviderKey, 0)
	for rows.Next() {
		item, err := scanProviderKey(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *item)
	}
	return items, rows.Err()
}
