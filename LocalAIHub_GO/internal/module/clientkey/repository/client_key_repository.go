package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type ClientKey struct {
	ID            int64      `json:"id"`
	Name          string     `json:"name"`
	KeyPrefix     string     `json:"key_prefix"`
	APIKeyHash    string     `json:"-"`
	PlainKey      string     `json:"plain_key,omitempty"`
	Status        string     `json:"status"`
	Remark        *string    `json:"remark,omitempty"`
	LastUsedAt    *time.Time `json:"last_used_at,omitempty"`
	ExpiresAt     *time.Time `json:"expires_at,omitempty"`
	AllowedModels []int64    `json:"allowed_models,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

type ClientKeyRepository struct{ db *sql.DB }

func NewClientKeyRepository(db *sql.DB) *ClientKeyRepository { return &ClientKeyRepository{db: db} }

func (r *ClientKeyRepository) List(ctx context.Context, page, pageSize int) ([]ClientKey, int, error) {
	var total int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM api_client`).Scan(&total); err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	rows, err := r.db.QueryContext(ctx, `SELECT id, name, key_prefix, api_key_hash, plain_key, status, remark, last_used_at, expires_at, created_at, updated_at FROM api_client ORDER BY id DESC LIMIT ? OFFSET ?`, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	items := make([]ClientKey, 0)
	for rows.Next() {
		var item ClientKey
		var remark sql.NullString
		var lastUsed, expires sql.NullTime
		if err := rows.Scan(&item.ID, &item.Name, &item.KeyPrefix, &item.APIKeyHash, &item.PlainKey, &item.Status, &remark, &lastUsed, &expires, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, 0, err
		}
		if remark.Valid {
			item.Remark = &remark.String
		}
		if lastUsed.Valid {
			item.LastUsedAt = &lastUsed.Time
		}
		if expires.Valid {
			item.ExpiresAt = &expires.Time
		}
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func (r *ClientKeyRepository) Get(ctx context.Context, id int64) (*ClientKey, error) {
	var item ClientKey
	var remark sql.NullString
	var lastUsed, expires sql.NullTime
	err := r.db.QueryRowContext(ctx, `SELECT id, name, key_prefix, api_key_hash, plain_key, status, remark, last_used_at, expires_at, created_at, updated_at FROM api_client WHERE id = ?`, id).Scan(&item.ID, &item.Name, &item.KeyPrefix, &item.APIKeyHash, &item.PlainKey, &item.Status, &remark, &lastUsed, &expires, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if remark.Valid {
		item.Remark = &remark.String
	}
	if lastUsed.Valid {
		item.LastUsedAt = &lastUsed.Time
	}
	if expires.Valid {
		item.ExpiresAt = &expires.Time
	}
	return &item, nil
}

func (r *ClientKeyRepository) Create(ctx context.Context, item *ClientKey) (int64, error) {
	result, err := r.db.ExecContext(ctx, `INSERT INTO api_client (name, key_prefix, api_key_hash, plain_key, status, remark, expires_at, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`, item.Name, item.KeyPrefix, item.APIKeyHash, item.PlainKey, item.Status, item.Remark, item.ExpiresAt, time.Now().UTC(), time.Now().UTC())
	if err != nil {
		return 0, fmt.Errorf("create client key: %w", err)
	}
	return result.LastInsertId()
}

func (r *ClientKeyRepository) AssignModels(ctx context.Context, clientID int64, modelIDs []int64) error {
	for _, modelID := range modelIDs {
		_, err := r.db.ExecContext(ctx, `INSERT IGNORE INTO api_client_model (client_id, virtual_model_id, created_at) VALUES (?, ?, ?)`, clientID, modelID, time.Now().UTC())
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *ClientKeyRepository) GetAllowedModels(ctx context.Context, clientID int64) ([]int64, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT virtual_model_id FROM api_client_model WHERE client_id = ?`, clientID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var models []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		models = append(models, id)
	}
	return models, rows.Err()
}

func (r *ClientKeyRepository) UpdateStatus(ctx context.Context, id int64, status string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE api_client SET status = ?, updated_at = ? WHERE id = ?`, status, time.Now().UTC(), id)
	return err
}

func (r *ClientKeyRepository) Update(ctx context.Context, id int64, name, remark string) error {
	var remarkPtr *string
	if remark != "" {
		remarkPtr = &remark
	}
	_, err := r.db.ExecContext(ctx, `UPDATE api_client SET name = ?, remark = ?, updated_at = ? WHERE id = ?`, name, remarkPtr, time.Now().UTC(), id)
	return err
}

func (r *ClientKeyRepository) ReplaceAllowedModels(ctx context.Context, clientID int64, modelIDs []int64) error {
	if _, err := r.db.ExecContext(ctx, `DELETE FROM api_client_model WHERE client_id = ?`, clientID); err != nil {
		return err
	}
	for _, modelID := range modelIDs {
		if _, err := r.db.ExecContext(ctx, `INSERT INTO api_client_model (client_id, virtual_model_id, created_at) VALUES (?, ?, ?)`, clientID, modelID, time.Now().UTC()); err != nil {
			return err
		}
	}
	return nil
}

func (r *ClientKeyRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM api_client WHERE id = ?`, id)
	return err
}
