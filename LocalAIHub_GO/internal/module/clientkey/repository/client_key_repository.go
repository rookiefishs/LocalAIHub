package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type ClientKey struct {
	ID                     int64      `json:"id"`
	Name                   string     `json:"name"`
	KeyPrefix              string     `json:"key_prefix"`
	APIKeyHash             string     `json:"-"`
	PlainKey               string     `json:"plain_key,omitempty"`
	Status                 string     `json:"status"`
	Remark                 *string    `json:"remark,omitempty"`
	LastUsedAt             *time.Time `json:"last_used_at,omitempty"`
	ExpiresAt              *time.Time `json:"expires_at,omitempty"`
	AllowedModels          []int64    `json:"allowed_models,omitempty"`
	RequestCount           int        `json:"request_count,omitempty"`
	TotalTokens            int        `json:"total_tokens,omitempty"`
	DailyRequestLimit      *int64     `json:"daily_request_limit,omitempty"`
	MonthlyRequestLimit    *int64     `json:"monthly_request_limit,omitempty"`
	DailyTokenLimit        *int64     `json:"daily_token_limit,omitempty"`
	MonthlyTokenLimit      *int64     `json:"monthly_token_limit,omitempty"`
	CurrentDailyRequests   int64      `json:"current_daily_requests"`
	CurrentMonthlyRequests int64      `json:"current_monthly_requests"`
	CurrentDailyTokens     int64      `json:"current_daily_tokens"`
	CurrentMonthlyTokens   int64      `json:"current_monthly_tokens"`
	QuotaResetAt           *string    `json:"quota_reset_at,omitempty"`
	QuotaDisabledAt        *time.Time `json:"quota_disabled_at,omitempty"`
	CreatedAt              time.Time  `json:"created_at"`
	UpdatedAt              time.Time  `json:"updated_at"`
}

type ClientKeyRepository struct{ db *sql.DB }

func NewClientKeyRepository(db *sql.DB) *ClientKeyRepository { return &ClientKeyRepository{db: db} }

func (r *ClientKeyRepository) List(ctx context.Context, page, pageSize int) ([]ClientKey, int, error) {
	var total int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM api_client`).Scan(&total); err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	rows, err := r.db.QueryContext(ctx, `SELECT id, name, key_prefix, api_key_hash, plain_key, status, remark, last_used_at, expires_at, daily_request_limit, monthly_request_limit, daily_token_limit, monthly_token_limit, current_daily_requests, current_monthly_requests, current_daily_tokens, current_monthly_tokens, quota_reset_at, quota_disabled_at, created_at, updated_at FROM api_client ORDER BY id DESC LIMIT ? OFFSET ?`, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	items := make([]ClientKey, 0)
	for rows.Next() {
		var item ClientKey
		var remark, quotaResetAt sql.NullString
		var lastUsed, expires, quotaDisabled sql.NullTime
		if err := rows.Scan(&item.ID, &item.Name, &item.KeyPrefix, &item.APIKeyHash, &item.PlainKey, &item.Status, &remark, &lastUsed, &expires, &item.DailyRequestLimit, &item.MonthlyRequestLimit, &item.DailyTokenLimit, &item.MonthlyTokenLimit, &item.CurrentDailyRequests, &item.CurrentMonthlyRequests, &item.CurrentDailyTokens, &item.CurrentMonthlyTokens, &quotaResetAt, &quotaDisabled, &item.CreatedAt, &item.UpdatedAt); err != nil {
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
		if quotaResetAt.Valid {
			item.QuotaResetAt = &quotaResetAt.String
		}
		if quotaDisabled.Valid {
			item.QuotaDisabledAt = &quotaDisabled.Time
		}
		items = append(items, item)
	}

	if len(items) > 0 {
		idStr := make([]string, len(items))
		for i, it := range items {
			idStr[i] = strconv.FormatInt(it.ID, 10)
		}
		idsComma := strings.Join(idStr, ",")
		statsRows, err := r.db.QueryContext(ctx, fmt.Sprintf(`SELECT client_id, COUNT(*) as request_count, COALESCE(SUM(total_tokens), 0) as total_tokens FROM request_log WHERE client_id IN (%s) GROUP BY client_id`, idsComma))
		if err == nil {
			defer statsRows.Close()
			statsMap := make(map[int64]struct {
				RequestCount int
				TotalTokens  int
			})
			for statsRows.Next() {
				var cid, rc, tt int64
				statsRows.Scan(&cid, &rc, &tt)
				statsMap[cid] = struct {
					RequestCount int
					TotalTokens  int
				}{RequestCount: int(rc), TotalTokens: int(tt)}
			}
			for i := range items {
				if st, ok := statsMap[items[i].ID]; ok {
					items[i].RequestCount = st.RequestCount
					items[i].TotalTokens = st.TotalTokens
				}
			}
		}
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

func (r *ClientKeyRepository) ListActive(ctx context.Context) ([]ClientKey, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, name, key_prefix, api_key_hash, plain_key, status, remark, last_used_at, expires_at, created_at, updated_at FROM api_client WHERE status = 'active'`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]ClientKey, 0)
	for rows.Next() {
		var item ClientKey
		var remark sql.NullString
		var lastUsed, expires sql.NullTime
		if err := rows.Scan(&item.ID, &item.Name, &item.KeyPrefix, &item.APIKeyHash, &item.PlainKey, &item.Status, &remark, &lastUsed, &expires, &item.CreatedAt, &item.UpdatedAt); err != nil {
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
		items = append(items, item)
	}
	return items, rows.Err()
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

func (r *ClientKeyRepository) GetByID(ctx context.Context, id int64) (*ClientKey, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, name, key_prefix, api_key_hash, plain_key, status, remark, last_used_at, expires_at, daily_request_limit, monthly_request_limit, daily_token_limit, monthly_token_limit, current_daily_requests, current_monthly_requests, current_daily_tokens, current_monthly_tokens, quota_reset_at, quota_disabled_at, created_at, updated_at FROM api_client WHERE id = ?`, id)
	var item ClientKey
	var remark, quotaResetAt sql.NullString
	var lastUsed, expires, quotaDisabled sql.NullTime
	if err := row.Scan(&item.ID, &item.Name, &item.KeyPrefix, &item.APIKeyHash, &item.PlainKey, &item.Status, &remark, &lastUsed, &expires, &item.DailyRequestLimit, &item.MonthlyRequestLimit, &item.DailyTokenLimit, &item.MonthlyTokenLimit, &item.CurrentDailyRequests, &item.CurrentMonthlyRequests, &item.CurrentDailyTokens, &item.CurrentMonthlyTokens, &quotaResetAt, &quotaDisabled, &item.CreatedAt, &item.UpdatedAt); err != nil {
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
	if quotaResetAt.Valid {
		item.QuotaResetAt = &quotaResetAt.String
	}
	if quotaDisabled.Valid {
		item.QuotaDisabledAt = &quotaDisabled.Time
	}
	return &item, nil
}

func (r *ClientKeyRepository) GetStatus(ctx context.Context, keyID int64) (string, error) {
	var status string
	err := r.db.QueryRowContext(ctx, `SELECT status FROM api_client WHERE id = ?`, keyID).Scan(&status)
	return status, err
}

func (r *ClientKeyRepository) UpdateStatus(ctx context.Context, keyID int64, status string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE api_client SET status = ?, updated_at = ? WHERE id = ?`, status, time.Now().UTC(), keyID)
	return err
}

func (r *ClientKeyRepository) UpdateQuota(ctx context.Context, keyID int64, dailyReq, monthlyReq, dailyToken, monthlyToken *int64) error {
	_, err := r.db.ExecContext(ctx, `UPDATE api_client SET daily_request_limit = ?, monthly_request_limit = ?, daily_token_limit = ?, monthly_token_limit = ?, updated_at = ? WHERE id = ?`,
		dailyReq, monthlyReq, dailyToken, monthlyToken, time.Now().UTC(), keyID)
	return err
}

func (r *ClientKeyRepository) IncrementUsage(ctx context.Context, keyID int64, tokens int) error {
	_, err := r.db.ExecContext(ctx, `UPDATE api_client SET current_daily_requests = current_daily_requests + 1, current_monthly_requests = current_monthly_requests + 1, current_daily_tokens = current_daily_tokens + ?, current_monthly_tokens = current_monthly_tokens + ?, updated_at = ? WHERE id = ?`,
		tokens, tokens, time.Now().UTC(), keyID)
	return err
}

func (r *ClientKeyRepository) GetQuotaUsage(ctx context.Context, keyID int64) (*ClientKey, error) {
	var item ClientKey
	var remark, quotaResetAt sql.NullString
	var lastUsed, expires, quotaDisabled sql.NullTime
	err := r.db.QueryRowContext(ctx, `SELECT id, name, key_prefix, api_key_hash, plain_key, status, remark, last_used_at, expires_at, daily_request_limit, monthly_request_limit, daily_token_limit, monthly_token_limit, current_daily_requests, current_monthly_requests, current_daily_tokens, current_monthly_tokens, quota_reset_at, quota_disabled_at, created_at, updated_at FROM api_client WHERE id = ?`, keyID).Scan(
		&item.ID, &item.Name, &item.KeyPrefix, &item.APIKeyHash, &item.PlainKey, &item.Status, &remark, &lastUsed, &expires, &item.DailyRequestLimit, &item.MonthlyRequestLimit, &item.DailyTokenLimit, &item.MonthlyTokenLimit, &item.CurrentDailyRequests, &item.CurrentMonthlyRequests, &item.CurrentDailyTokens, &item.CurrentMonthlyTokens, &quotaResetAt, &quotaDisabled, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		return nil, err
	}
	if remark.Valid {
		item.Remark = &remark.String
	}
	if quotaResetAt.Valid {
		item.QuotaResetAt = &quotaResetAt.String
	}
	if quotaDisabled.Valid {
		item.QuotaDisabledAt = &quotaDisabled.Time
	}
	return &item, nil
}

func (r *ClientKeyRepository) ResetDailyQuota(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, `UPDATE api_client SET current_daily_requests = 0, current_daily_tokens = 0, quota_reset_at = CURDATE() WHERE status = 'active'`)
	return err
}

func (r *ClientKeyRepository) ResetMonthlyQuota(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, `UPDATE api_client SET current_monthly_requests = 0, current_monthly_tokens = 0 WHERE status = 'active'`)
	return err
}
