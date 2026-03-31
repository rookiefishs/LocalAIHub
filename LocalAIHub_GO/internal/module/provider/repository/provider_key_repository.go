package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"localaihub/localaihub_go/internal/pkg/logger"
)

var Log = logger.Log

type ProviderKey struct {
	ID               int64      `json:"id"`
	ProviderID       int64      `json:"provider_id"`
	KeyMasked        string     `json:"key_masked"`
	SecretEncrypted  string     `json:"-"`
	Status           string     `json:"status"`
	Priority         int        `json:"priority"`
	FailCount        int        `json:"fail_count"`
	LastUsedAt       *time.Time `json:"last_used_at,omitempty"`
	LastErrorAt      *time.Time `json:"last_error_at,omitempty"`
	LastErrorMessage *string    `json:"last_error_message,omitempty"`
	Remark           *string    `json:"remark,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

type ProviderKeyRepository struct{ db *sql.DB }

func NewProviderKeyRepository(db *sql.DB) *ProviderKeyRepository {
	return &ProviderKeyRepository{db: db}
}

func (r *ProviderKeyRepository) ListByProviderID(ctx context.Context, providerID int64) ([]ProviderKey, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, provider_id, key_masked, secret_encrypted, status, priority, fail_count, last_used_at, last_error_at, last_error_message, remark, created_at, updated_at FROM provider_key WHERE provider_id = ? ORDER BY priority ASC, id DESC`, providerID)
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

func (r *ProviderKeyRepository) GetByID(ctx context.Context, id int64) (*ProviderKey, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, provider_id, key_masked, secret_encrypted, status, priority, fail_count, last_used_at, last_error_at, last_error_message, remark, created_at, updated_at FROM provider_key WHERE id = ?`, id)
	return scanProviderKey(row)
}

func (r *ProviderKeyRepository) FirstActiveByProviderID(ctx context.Context, providerID int64) (*ProviderKey, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, provider_id, key_masked, status FROM provider_key WHERE provider_id = ?`, providerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var id, providerID int
		var status, keyMasked string
		rows.Scan(&id, &providerID, &keyMasked, &status)
		Log.Debug().Int("provider_id", providerID).Int("key_id", id).Str("status", status).Str("key_masked", keyMasked).Msg("found key")
		count++
	}
	Log.Debug().Int64("provider_id", providerID).Int("total_keys", count).Msg("key search result")

	row := r.db.QueryRowContext(ctx, `SELECT id, provider_id, key_masked, secret_encrypted, status, priority, fail_count, last_used_at, last_error_at, last_error_message, remark, created_at, updated_at FROM provider_key WHERE provider_id = ? AND status = 'enabled' ORDER BY priority ASC, id ASC LIMIT 1`, providerID)
	return scanProviderKey(row)
}

func (r *ProviderKeyRepository) Create(ctx context.Context, item *ProviderKey) (int64, error) {
	result, err := r.db.ExecContext(ctx, `INSERT INTO provider_key (provider_id, key_masked, secret_encrypted, status, priority, fail_count, remark, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`, item.ProviderID, item.KeyMasked, item.SecretEncrypted, item.Status, item.Priority, item.FailCount, item.Remark, time.Now().UTC(), time.Now().UTC())
	if err != nil {
		return 0, fmt.Errorf("create provider key: %w", err)
	}
	return result.LastInsertId()
}

func (r *ProviderKeyRepository) UpdateStatus(ctx context.Context, id int64, status string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE provider_key SET status = ?, updated_at = ? WHERE id = ?`, status, time.Now().UTC(), id)
	return err
}

func (r *ProviderKeyRepository) MarkResult(ctx context.Context, id int64, success bool, errorMessage string) error {
	if success {
		_, err := r.db.ExecContext(ctx, `UPDATE provider_key SET last_used_at = ?, fail_count = 0, last_error_at = NULL, last_error_message = NULL, updated_at = ? WHERE id = ?`, time.Now().UTC(), time.Now().UTC(), id)
		return err
	}
	_, err := r.db.ExecContext(ctx, `UPDATE provider_key SET fail_count = fail_count + 1, last_error_at = ?, last_error_message = ?, updated_at = ? WHERE id = ?`, time.Now().UTC(), errorMessage, time.Now().UTC(), id)
	return err
}

func (r *ProviderKeyRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM provider_key WHERE id = ?`, id)
	return err
}

func scanProviderKey(scanner interface{ Scan(dest ...any) error }) (*ProviderKey, error) {
	var item ProviderKey
	var lastUsed, lastError sql.NullTime
	var lastErrorMessage, remark sql.NullString
	err := scanner.Scan(&item.ID, &item.ProviderID, &item.KeyMasked, &item.SecretEncrypted, &item.Status, &item.Priority, &item.FailCount, &lastUsed, &lastError, &lastErrorMessage, &remark, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if lastUsed.Valid {
		item.LastUsedAt = &lastUsed.Time
	}
	if lastError.Valid {
		item.LastErrorAt = &lastError.Time
	}
	if lastErrorMessage.Valid {
		item.LastErrorMessage = &lastErrorMessage.String
	}
	if remark.Valid {
		item.Remark = &remark.String
	}
	return &item, nil
}
