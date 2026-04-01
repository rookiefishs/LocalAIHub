package repository

import (
	"context"
	"database/sql"
	"time"
)

type ModelPricing struct {
	ID                   int64     `json:"id"`
	ModelCode            string    `json:"model_code"`
	ProviderID           *int64    `json:"provider_id"`
	PromptPricePer1k     float64   `json:"prompt_price_per_1k"`
	CompletionPricePer1k float64   `json:"completion_price_per_1k"`
	Currency             string    `json:"currency"`
	Enabled              bool      `json:"enabled"`
	Remark               *string   `json:"remark"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

type AnalyticsRepository struct {
	db *sql.DB
}

func NewAnalyticsRepository(db *sql.DB) *AnalyticsRepository {
	return &AnalyticsRepository{db: db}
}

func (r *AnalyticsRepository) GetCostStats(ctx context.Context, hours int, providerID, clientID int64) (map[string]float64, error) {
	var totalCost float64
	query := `
		SELECT COALESCE(SUM(
			COALESCE(rl.prompt_tokens, 0) * COALESCE(mp.prompt_price_per_1k, 0) / 1000 +
			COALESCE(rl.completion_tokens, 0) * COALESCE(mp.completion_price_per_1k, 0) / 1000
		), 0) as total_cost
		FROM request_log rl
		LEFT JOIN model_pricing mp ON (mp.model_code = rl.virtual_model_code OR mp.model_code = rl.upstream_model_name)
		WHERE rl.created_at >= DATE_SUB(UTC_TIMESTAMP(), INTERVAL ? HOUR)`
	args := []any{hours}

	if providerID > 0 {
		query += " AND rl.provider_id = ?"
		args = append(args, providerID)
	}
	if clientID > 0 {
		query += " AND rl.client_id = ?"
		args = append(args, clientID)
	}

	err := r.db.QueryRowContext(ctx, query, args...).Scan(&totalCost)
	return map[string]float64{"total_cost": totalCost}, err
}

func (r *AnalyticsRepository) GetCostByProvider(ctx context.Context, hours int, clientID int64) ([]map[string]any, error) {
	query := `
		SELECT 
			p.id as provider_id,
			p.name as provider_name,
			COALESCE(SUM(rl.prompt_tokens), 0) as prompt_tokens,
			COALESCE(SUM(rl.completion_tokens), 0) as completion_tokens,
			COALESCE(SUM(rl.total_tokens), 0) as total_tokens,
			COALESCE(SUM(
				COALESCE(rl.prompt_tokens, 0) * COALESCE(mp.prompt_price_per_1k, 0) / 1000 +
				COALESCE(rl.completion_tokens, 0) * COALESCE(mp.completion_price_per_1k, 0) / 1000
			), 0) as cost
		FROM request_log rl
		LEFT JOIN provider p ON rl.provider_id = p.id
		LEFT JOIN model_pricing mp ON (mp.model_code = rl.virtual_model_code OR mp.model_code = rl.upstream_model_name)
		WHERE rl.created_at >= DATE_SUB(UTC_TIMESTAMP(), INTERVAL ? HOUR)`
	args := []any{hours}

	if clientID > 0 {
		query += " AND rl.client_id = ?"
		args = append(args, clientID)
	}
	query += " GROUP BY p.id, p.name ORDER BY cost DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []map[string]any
	for rows.Next() {
		var pID sql.NullInt64
		var pName sql.NullString
		var promptTokens, completionTokens, totalTokens int64
		var cost float64
		if err := rows.Scan(&pID, &pName, &promptTokens, &completionTokens, &totalTokens, &cost); err != nil {
			return nil, err
		}
		results = append(results, map[string]any{
			"provider_id":       pID.Int64,
			"provider_name":     pName.String,
			"prompt_tokens":     promptTokens,
			"completion_tokens": completionTokens,
			"total_tokens":      totalTokens,
			"cost":              cost,
		})
	}
	return results, nil
}

func (r *AnalyticsRepository) GetCostByModel(ctx context.Context, hours int, providerID, clientID int64) ([]map[string]any, error) {
	query := `
		SELECT 
			COALESCE(rl.virtual_model_code, rl.upstream_model_name, 'unknown') as model_code,
			COALESCE(SUM(rl.prompt_tokens), 0) as prompt_tokens,
			COALESCE(SUM(rl.completion_tokens), 0) as completion_tokens,
			COALESCE(SUM(rl.total_tokens), 0) as total_tokens,
			COALESCE(SUM(
				COALESCE(rl.prompt_tokens, 0) * COALESCE(mp.prompt_price_per_1k, 0) / 1000 +
				COALESCE(rl.completion_tokens, 0) * COALESCE(mp.completion_price_per_1k, 0) / 1000
			), 0) as cost
		FROM request_log rl
		LEFT JOIN model_pricing mp ON (mp.model_code = rl.virtual_model_code OR mp.model_code = rl.upstream_model_name)
		WHERE rl.created_at >= DATE_SUB(UTC_TIMESTAMP(), INTERVAL ? HOUR)`
	args := []any{hours}

	if providerID > 0 {
		query += " AND rl.provider_id = ?"
		args = append(args, providerID)
	}
	if clientID > 0 {
		query += " AND rl.client_id = ?"
		args = append(args, clientID)
	}
	query += " GROUP BY model_code ORDER BY cost DESC LIMIT 20"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []map[string]any
	for rows.Next() {
		var modelCode string
		var promptTokens, completionTokens, totalTokens int64
		var cost float64
		if err := rows.Scan(&modelCode, &promptTokens, &completionTokens, &totalTokens, &cost); err != nil {
			return nil, err
		}
		results = append(results, map[string]any{
			"model_code":        modelCode,
			"prompt_tokens":     promptTokens,
			"completion_tokens": completionTokens,
			"total_tokens":      totalTokens,
			"cost":              cost,
		})
	}
	return results, nil
}

func (r *AnalyticsRepository) GetCostTrend(ctx context.Context, hours int, providerID, clientID int64) ([]map[string]any, error) {
	groupBy := "DATE(rl.created_at)"
	if hours <= 24 {
		groupBy = "DATE_FORMAT(rl.created_at, '%Y-%m-%d %H:00')"
	}

	query := `
		SELECT 
			` + groupBy + ` as period,
			COALESCE(SUM(
				COALESCE(rl.prompt_tokens, 0) * COALESCE(mp.prompt_price_per_1k, 0) / 1000 +
				COALESCE(rl.completion_tokens, 0) * COALESCE(mp.completion_price_per_1k, 0) / 1000
			), 0) as cost,
			COALESCE(SUM(rl.total_tokens), 0) as tokens
		FROM request_log rl
		LEFT JOIN model_pricing mp ON (mp.model_code = rl.virtual_model_code OR mp.model_code = rl.upstream_model_name)
		WHERE rl.created_at >= DATE_SUB(UTC_TIMESTAMP(), INTERVAL ? HOUR)`
	args := []any{hours}

	if providerID > 0 {
		query += " AND rl.provider_id = ?"
		args = append(args, providerID)
	}
	if clientID > 0 {
		query += " AND rl.client_id = ?"
		args = append(args, clientID)
	}
	query += " GROUP BY " + groupBy + " ORDER BY period ASC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []map[string]any
	for rows.Next() {
		var period string
		var cost, tokens float64
		if err := rows.Scan(&period, &cost, &tokens); err != nil {
			return nil, err
		}
		results = append(results, map[string]any{
			"period": period,
			"cost":   cost,
			"tokens": tokens,
		})
	}
	return results, nil
}

func (r *AnalyticsRepository) GetTokenStats(ctx context.Context, hours int, providerID, clientID int64) (map[string]int64, error) {
	var promptTokens, completionTokens, totalTokens int64
	query := `SELECT COALESCE(SUM(prompt_tokens),0), COALESCE(SUM(completion_tokens),0), COALESCE(SUM(total_tokens),0) FROM request_log WHERE created_at >= DATE_SUB(UTC_TIMESTAMP(), INTERVAL ? HOUR)`
	args := []any{hours}

	if providerID > 0 {
		query += " AND provider_id = ?"
		args = append(args, providerID)
	}
	if clientID > 0 {
		query += " AND client_id = ?"
		args = append(args, clientID)
	}

	err := r.db.QueryRowContext(ctx, query, args...).Scan(&promptTokens, &completionTokens, &totalTokens)
	return map[string]int64{
		"prompt_tokens":     promptTokens,
		"completion_tokens": completionTokens,
		"total_tokens":      totalTokens,
	}, err
}

func (r *AnalyticsRepository) GetTokenTrend(ctx context.Context, hours int, providerID, clientID int64) ([]map[string]any, error) {
	groupBy := "DATE(created_at)"
	if hours <= 24 {
		groupBy = "DATE_FORMAT(created_at, '%Y-%m-%d %H:00')"
	}

	query := `
		SELECT 
			` + groupBy + ` as period,
			COALESCE(SUM(prompt_tokens), 0) as prompt_tokens,
			COALESCE(SUM(completion_tokens), 0) as completion_tokens,
			COALESCE(SUM(total_tokens), 0) as total_tokens
		FROM request_log
		WHERE created_at >= DATE_SUB(UTC_TIMESTAMP(), INTERVAL ? HOUR)`
	args := []any{hours}

	if providerID > 0 {
		query += " AND provider_id = ?"
		args = append(args, providerID)
	}
	if clientID > 0 {
		query += " AND client_id = ?"
		args = append(args, clientID)
	}
	query += " GROUP BY " + groupBy + " ORDER BY period ASC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []map[string]any
	for rows.Next() {
		var period string
		var promptTokens, completionTokens, totalTokens int64
		if err := rows.Scan(&period, &promptTokens, &completionTokens, &totalTokens); err != nil {
			return nil, err
		}
		results = append(results, map[string]any{
			"period":            period,
			"prompt_tokens":     promptTokens,
			"completion_tokens": completionTokens,
			"total_tokens":      totalTokens,
		})
	}
	return results, nil
}

func (r *AnalyticsRepository) GetRequestComparison(ctx context.Context, hours int, clientID int64) (map[string]any, error) {
	var currentCount, previousCount int64
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM request_log WHERE created_at >= DATE_SUB(UTC_TIMESTAMP(), INTERVAL ? HOUR)`, hours).Scan(&currentCount)
	if err != nil {
		return nil, err
	}

	prevHours := hours * 2
	err = r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM request_log WHERE created_at >= DATE_SUB(UTC_TIMESTAMP(), INTERVAL ? HOUR) AND created_at < DATE_SUB(UTC_TIMESTAMP(), INTERVAL ? HOUR)`, prevHours, hours).Scan(&previousCount)
	if err != nil {
		previousCount = 0
	}

	return buildComparison("request_count", float64(currentCount), float64(previousCount)), nil
}

func (r *AnalyticsRepository) GetTokenComparison(ctx context.Context, hours int, clientID int64) (map[string]any, error) {
	var currentTokens, previousTokens int64
	err := r.db.QueryRowContext(ctx, `SELECT COALESCE(SUM(total_tokens),0) FROM request_log WHERE created_at >= DATE_SUB(UTC_TIMESTAMP(), INTERVAL ? HOUR)`, hours).Scan(&currentTokens)
	if err != nil {
		return nil, err
	}

	prevHours := hours * 2
	err = r.db.QueryRowContext(ctx, `SELECT COALESCE(SUM(total_tokens),0) FROM request_log WHERE created_at >= DATE_SUB(UTC_TIMESTAMP(), INTERVAL ? HOUR) AND created_at < DATE_SUB(UTC_TIMESTAMP(), INTERVAL ? HOUR)`, prevHours, hours).Scan(&previousTokens)
	if err != nil {
		previousTokens = 0
	}

	return buildComparison("total_tokens", float64(currentTokens), float64(previousTokens)), nil
}

func (r *AnalyticsRepository) GetCostComparison(ctx context.Context, hours int, clientID int64) (map[string]any, error) {
	var currentCost, previousCost float64

	err := r.db.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(
			COALESCE(rl.prompt_tokens, 0) * COALESCE(mp.prompt_price_per_1k, 0) / 1000 +
			COALESCE(rl.completion_tokens, 0) * COALESCE(mp.completion_price_per_1k, 0) / 1000
		), 0)
		FROM request_log rl
		LEFT JOIN model_pricing mp ON (mp.model_code = rl.virtual_model_code OR mp.model_code = rl.upstream_model_name)
		WHERE rl.created_at >= DATE_SUB(UTC_TIMESTAMP(), INTERVAL ? HOUR)`, hours).Scan(&currentCost)
	if err != nil {
		return nil, err
	}

	prevHours := hours * 2
	err = r.db.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(
			COALESCE(rl.prompt_tokens, 0) * COALESCE(mp.prompt_price_per_1k, 0) / 1000 +
			COALESCE(rl.completion_tokens, 0) * COALESCE(mp.completion_price_per_1k, 0) / 1000
		), 0)
		FROM request_log rl
		LEFT JOIN model_pricing mp ON (mp.model_code = rl.virtual_model_code OR mp.model_code = rl.upstream_model_name)
		WHERE rl.created_at >= DATE_SUB(UTC_TIMESTAMP(), INTERVAL ? HOUR) AND rl.created_at < DATE_SUB(UTC_TIMESTAMP(), INTERVAL ? HOUR)`, prevHours, hours).Scan(&previousCost)
	if err != nil {
		previousCost = 0
	}

	return buildComparison("total_cost", currentCost, previousCost), nil
}

func buildComparison(key string, current, previous float64) map[string]any {
	var change = current - previous
	var changeRate float64
	var direction string
	if previous > 0 {
		changeRate = (change / previous) * 100
	}
	if change > 0 {
		direction = "up"
	} else if change < 0 {
		direction = "down"
	} else {
		direction = "flat"
	}
	return map[string]any{
		key:           current,
		"previous":    previous,
		"change":      change,
		"change_rate": changeRate,
		"direction":   direction,
	}
}

func (r *AnalyticsRepository) ListModelPricing(ctx context.Context, page, pageSize int) ([]ModelPricing, int, error) {
	var total int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM model_pricing`).Scan(&total); err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	rows, err := r.db.QueryContext(ctx, `SELECT id, model_code, provider_id, prompt_price_per_1k, completion_price_per_1k, currency, enabled, remark, created_at, updated_at FROM model_pricing ORDER BY id LIMIT ? OFFSET ?`, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []ModelPricing
	for rows.Next() {
		var item ModelPricing
		var providerID sql.NullInt64
		var remark sql.NullString
		if err := rows.Scan(&item.ID, &item.ModelCode, &providerID, &item.PromptPricePer1k, &item.CompletionPricePer1k, &item.Currency, &item.Enabled, &remark, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, 0, err
		}
		if providerID.Valid {
			item.ProviderID = &providerID.Int64
		}
		if remark.Valid {
			item.Remark = &remark.String
		}
		items = append(items, item)
	}
	return items, total, nil
}

func (r *AnalyticsRepository) GetModelPricing(ctx context.Context, id int64) (*ModelPricing, error) {
	var item ModelPricing
	var providerID sql.NullInt64
	var remark sql.NullString
	err := r.db.QueryRowContext(ctx, `SELECT id, model_code, provider_id, prompt_price_per_1k, completion_price_per_1k, currency, enabled, remark, created_at, updated_at FROM model_pricing WHERE id = ?`, id).Scan(&item.ID, &item.ModelCode, &providerID, &item.PromptPricePer1k, &item.CompletionPricePer1k, &item.Currency, &item.Enabled, &remark, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		return nil, err
	}
	if providerID.Valid {
		item.ProviderID = &providerID.Int64
	}
	if remark.Valid {
		item.Remark = &remark.String
	}
	return &item, nil
}

func (r *AnalyticsRepository) CreateModelPricing(ctx context.Context, item *ModelPricing) (int64, error) {
	result, err := r.db.ExecContext(ctx, `INSERT INTO model_pricing (model_code, provider_id, prompt_price_per_1k, completion_price_per_1k, currency, enabled, remark) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		item.ModelCode, item.ProviderID, item.PromptPricePer1k, item.CompletionPricePer1k, item.Currency, item.Enabled, item.Remark)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (r *AnalyticsRepository) UpdateModelPricing(ctx context.Context, item *ModelPricing) error {
	_, err := r.db.ExecContext(ctx, `UPDATE model_pricing SET model_code=?, provider_id=?, prompt_price_per_1k=?, completion_price_per_1k=?, currency=?, enabled=?, remark=? WHERE id=?`,
		item.ModelCode, item.ProviderID, item.PromptPricePer1k, item.CompletionPricePer1k, item.Currency, item.Enabled, item.Remark, item.ID)
	return err
}

func (r *AnalyticsRepository) DeleteModelPricing(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM model_pricing WHERE id = ?`, id)
	return err
}
