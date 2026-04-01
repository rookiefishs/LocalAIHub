package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

type Model struct {
	ID              int64           `json:"id"`
	ModelCode       string          `json:"model_code"`
	DisplayName     string          `json:"display_name"`
	ProtocolFamily  string          `json:"protocol_family"`
	CapabilityFlags json.RawMessage `json:"capability_flags"`
	Visible         bool            `json:"visible"`
	Status          string          `json:"status"`
	SortOrder       int             `json:"sort_order"`
	Description     *string         `json:"description,omitempty"`
	DefaultParams   json.RawMessage `json:"default_params_json"`
	Remark          *string         `json:"remark,omitempty"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

type Binding struct {
	ID                     int64           `json:"id"`
	VirtualModelID         int64           `json:"virtual_model_id"`
	ProviderID             int64           `json:"provider_id"`
	ProviderKeyID          *int64          `json:"provider_key_id,omitempty"`
	UpstreamModelName      string          `json:"upstream_model_name"`
	Priority               int             `json:"priority"`
	IsSameName             bool            `json:"is_same_name"`
	Enabled                bool            `json:"enabled"`
	CapabilitySnapshotJSON json.RawMessage `json:"capability_snapshot_json"`
	ParamOverrideJSON      json.RawMessage `json:"param_override_json"`
	Remark                 *string         `json:"remark,omitempty"`
	CreatedAt              time.Time       `json:"created_at"`
	UpdatedAt              time.Time       `json:"updated_at"`
}

type ModelRepository struct{ db *sql.DB }

func NewModelRepository(db *sql.DB) *ModelRepository { return &ModelRepository{db: db} }

func (r *ModelRepository) List(ctx context.Context, page, pageSize int) ([]Model, int, error) {
	var total int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM virtual_model`).Scan(&total); err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	rows, err := r.db.QueryContext(ctx, `SELECT id, model_code, display_name, protocol_family, capability_flags, visible, status, sort_order, description, default_params_json, remark, created_at, updated_at FROM virtual_model ORDER BY sort_order ASC, id DESC LIMIT ? OFFSET ?`, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	items := make([]Model, 0)
	for rows.Next() {
		var item Model
		var desc, remark sql.NullString
		if err := rows.Scan(&item.ID, &item.ModelCode, &item.DisplayName, &item.ProtocolFamily, &item.CapabilityFlags, &item.Visible, &item.Status, &item.SortOrder, &desc, &item.DefaultParams, &remark, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, 0, err
		}
		if desc.Valid {
			item.Description = &desc.String
		}
		if remark.Valid {
			item.Remark = &remark.String
		}
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func (r *ModelRepository) Get(ctx context.Context, id int64) (*Model, error) {
	var item Model
	var desc, remark sql.NullString
	err := r.db.QueryRowContext(ctx, `SELECT id, model_code, display_name, protocol_family, capability_flags, visible, status, sort_order, description, default_params_json, remark, created_at, updated_at FROM virtual_model WHERE id = ?`, id).Scan(&item.ID, &item.ModelCode, &item.DisplayName, &item.ProtocolFamily, &item.CapabilityFlags, &item.Visible, &item.Status, &item.SortOrder, &desc, &item.DefaultParams, &remark, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if desc.Valid {
		item.Description = &desc.String
	}
	if remark.Valid {
		item.Remark = &remark.String
	}
	return &item, nil
}

func (r *ModelRepository) Create(ctx context.Context, item *Model) (int64, error) {
	result, err := r.db.ExecContext(ctx, `INSERT INTO virtual_model (model_code, display_name, protocol_family, capability_flags, visible, status, sort_order, description, default_params_json, remark, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, item.ModelCode, item.DisplayName, item.ProtocolFamily, normalizeJSON(item.CapabilityFlags, `[]`), item.Visible, item.Status, item.SortOrder, item.Description, normalizeJSON(item.DefaultParams, `{}`), item.Remark, time.Now().UTC(), time.Now().UTC())
	if err != nil {
		return 0, fmt.Errorf("create model: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	_, _ = r.db.ExecContext(ctx, `INSERT INTO route_state (virtual_model_id, current_binding_id, route_status, manual_locked, updated_at) VALUES (?, NULL, 'normal', 0, ?) ON DUPLICATE KEY UPDATE updated_at = VALUES(updated_at)`, id, time.Now().UTC())
	return id, nil
}

func (r *ModelRepository) Update(ctx context.Context, item *Model) error {
	_, err := r.db.ExecContext(ctx, `UPDATE virtual_model SET model_code = ?, display_name = ?, protocol_family = ?, capability_flags = ?, visible = ?, status = ?, sort_order = ?, description = ?, default_params_json = ?, remark = ?, updated_at = ? WHERE id = ?`, item.ModelCode, item.DisplayName, item.ProtocolFamily, normalizeJSON(item.CapabilityFlags, `[]`), item.Visible, item.Status, item.SortOrder, item.Description, normalizeJSON(item.DefaultParams, `{}`), item.Remark, time.Now().UTC(), item.ID)
	return err
}

func (r *ModelRepository) ListBindings(ctx context.Context, modelID int64) ([]Binding, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, virtual_model_id, provider_id, provider_key_id, upstream_model_name, priority, is_same_name, enabled, capability_snapshot_json, param_override_json, remark, created_at, updated_at FROM virtual_model_binding WHERE virtual_model_id = ? ORDER BY priority ASC, id DESC`, modelID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]Binding, 0)
	for rows.Next() {
		var item Binding
		var providerKeyID sql.NullInt64
		var remark sql.NullString
		if err := rows.Scan(&item.ID, &item.VirtualModelID, &item.ProviderID, &providerKeyID, &item.UpstreamModelName, &item.Priority, &item.IsSameName, &item.Enabled, &item.CapabilitySnapshotJSON, &item.ParamOverrideJSON, &remark, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		if providerKeyID.Valid {
			item.ProviderKeyID = &providerKeyID.Int64
		}
		if remark.Valid {
			item.Remark = &remark.String
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *ModelRepository) CreateBinding(ctx context.Context, item *Binding) (int64, error) {
	result, err := r.db.ExecContext(ctx, `INSERT INTO virtual_model_binding (virtual_model_id, provider_id, provider_key_id, upstream_model_name, priority, is_same_name, enabled, capability_snapshot_json, param_override_json, remark, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, item.VirtualModelID, item.ProviderID, item.ProviderKeyID, item.UpstreamModelName, item.Priority, item.IsSameName, item.Enabled, normalizeJSON(item.CapabilitySnapshotJSON, `{}`), normalizeJSON(item.ParamOverrideJSON, `{}`), item.Remark, time.Now().UTC(), time.Now().UTC())
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	_, _ = r.db.ExecContext(ctx, `UPDATE route_state SET current_binding_id = IFNULL(current_binding_id, ?), route_status = CASE WHEN current_binding_id IS NULL THEN 'normal' ELSE route_status END, updated_at = ? WHERE virtual_model_id = ?`, id, time.Now().UTC(), item.VirtualModelID)
	return id, nil
}

func (r *ModelRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM virtual_model WHERE id = ?`, id)
	return err
}

func (r *ModelRepository) DeleteBinding(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM virtual_model_binding WHERE id = ?`, id)
	return err
}

func (r *ModelRepository) UpdateBinding(ctx context.Context, item *Binding) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE virtual_model_binding 
		SET provider_id = ?, provider_key_id = ?, upstream_model_name = ?, priority = ?, is_same_name = ?, enabled = ?
		WHERE id = ?
	`, item.ProviderID, item.ProviderKeyID, item.UpstreamModelName, item.Priority, item.IsSameName, item.Enabled, item.ID)
	return err
}

func (r *ModelRepository) GetBinding(ctx context.Context, bindingID int64) (*Binding, error) {
	var item Binding
	var providerKeyID sql.NullInt64
	var remark sql.NullString
	err := r.db.QueryRowContext(ctx, `SELECT id, virtual_model_id, provider_id, provider_key_id, upstream_model_name, priority, is_same_name, enabled, capability_snapshot_json, param_override_json, remark, created_at, updated_at FROM virtual_model_binding WHERE id = ?`, bindingID).Scan(&item.ID, &item.VirtualModelID, &item.ProviderID, &providerKeyID, &item.UpstreamModelName, &item.Priority, &item.IsSameName, &item.Enabled, &item.CapabilitySnapshotJSON, &item.ParamOverrideJSON, &remark, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if providerKeyID.Valid {
		item.ProviderKeyID = &providerKeyID.Int64
	}
	if remark.Valid {
		item.Remark = &remark.String
	}
	return &item, nil
}

func normalizeJSON(raw json.RawMessage, fallback string) string {
	if len(raw) == 0 {
		return fallback
	}
	return string(raw)
}
