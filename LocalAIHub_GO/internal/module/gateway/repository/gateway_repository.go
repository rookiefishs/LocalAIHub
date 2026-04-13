package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"localaihub/localaihub_go/internal/pkg/logger"
)

var Log = logger.Log

type GatewayClient struct {
	ID                     int64
	Name                   string
	KeyPrefix              string
	APIKeyHash             string
	Status                 string
	ExpiresAt              *time.Time
	DailyRequestLimit      *int64
	MonthlyRequestLimit    *int64
	DailyTokenLimit        *int64
	MonthlyTokenLimit      *int64
	CurrentDailyRequests   int64
	CurrentMonthlyRequests int64
	CurrentDailyTokens     int64
	CurrentMonthlyTokens   int64
	QuotaDisabledAt        *time.Time
}

type ModelRoute struct {
	VirtualModelID    int64
	ModelCode         string
	DisplayName       string
	CurrentBindingID  *int64
	BindingID         int64
	ProviderID        int64
	ProviderName      string
	BaseURL           string
	AuthType          string
	UpstreamModelName string
	ProviderKeyID     *int64
	BindingPriority   int
	BindingEnabled    bool
	DefaultParamsJSON json.RawMessage
}

type RequestLogInput struct {
	TraceID             string
	ProtocolType        string
	ClientID            *int64
	VirtualModelID      *int64
	VirtualModelCode    *string
	RequestedModel      *string
	BindingID           *int64
	ProviderID          *int64
	ProviderKeyID       *int64
	UpstreamModelName   *string
	RequestSummaryJSON  json.RawMessage
	ResponseSummaryJSON json.RawMessage
	StatusCode          *int
	Success             bool
	LatencyMS           *int
	PromptTokens        *int
	CompletionTokens    *int
	TotalTokens         *int
	ErrorCode           *string
	ErrorMessage        *string
	IsDebugLogged       bool
}

type RequestLogRecord struct {
	ID               int64     `json:"id"`
	TraceID          string    `json:"trace_id"`
	ProtocolType     string    `json:"protocol_type"`
	ClientID         *int64    `json:"client_id,omitempty"`
	KeyName          string    `json:"key_name,omitempty"`
	VirtualModelID   *int64    `json:"virtual_model_id,omitempty"`
	VirtualModelCode *string   `json:"virtual_model_code,omitempty"`
	VirtualModelName *string   `json:"virtual_model_name,omitempty"`
	RequestedModel   *string   `json:"requested_model,omitempty"`
	BindingID        *int64    `json:"binding_id,omitempty"`
	RouteName        *string   `json:"route_name,omitempty"`
	ProviderID       *int64    `json:"provider_id,omitempty"`
	ProviderName     *string   `json:"provider_name,omitempty"`
	UpstreamModel    *string   `json:"upstream_model_name,omitempty"`
	StatusCode       *int      `json:"status_code,omitempty"`
	Success          bool      `json:"success"`
	LatencyMS        *int      `json:"latency_ms,omitempty"`
	PromptTokens     *int      `json:"prompt_tokens,omitempty"`
	CompletionTokens *int      `json:"completion_tokens,omitempty"`
	TotalTokens      *int      `json:"total_tokens,omitempty"`
	ErrorCode        *string   `json:"error_code,omitempty"`
	ErrorMessage     *string   `json:"error_message,omitempty"`
	RequestSummary   *string   `json:"request_summary,omitempty"`
	ResponseSummary  *string   `json:"response_summary,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
}

type RequestLogFilters struct {
	ClientID         *int64
	VirtualModelCode string
	Success          *bool
	TimeRange        string
	StartTime        *time.Time
	EndTime          *time.Time
	Limit            int
	Page             int
}

type GatewayRepository struct{ db *sql.DB }

func NewGatewayRepository(db *sql.DB) *GatewayRepository { return &GatewayRepository{db: db} }

func (r *GatewayRepository) CountRequests24h(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM request_log WHERE created_at >= DATE_SUB(NOW(), INTERVAL 24 HOUR)`).Scan(&count)
	return count, err
}

func (r *GatewayRepository) CountRequests(ctx context.Context, hours int, clientID int64) (int64, error) {
	if clientID > 0 {
		var count int64
		err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM request_log WHERE client_id = ? AND created_at >= DATE_SUB(NOW(), INTERVAL ? HOUR)`, clientID, hours).Scan(&count)
		return count, err
	}
	var count int64
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM request_log WHERE created_at >= DATE_SUB(NOW(), INTERVAL ? HOUR)`, hours).Scan(&count)
	return count, err
}

func (r *GatewayRepository) CountSuccessRequests(ctx context.Context, hours int, clientID int64) (int64, error) {
	if clientID > 0 {
		var count int64
		err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM request_log WHERE client_id = ? AND success = 1 AND created_at >= DATE_SUB(NOW(), INTERVAL ? HOUR)`, clientID, hours).Scan(&count)
		return count, err
	}
	var count int64
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM request_log WHERE success = 1 AND created_at >= DATE_SUB(NOW(), INTERVAL ? HOUR)`, hours).Scan(&count)
	return count, err
}

func (r *GatewayRepository) AvgLatency(ctx context.Context, hours int, clientID int64) (int64, error) {
	var avg sql.NullFloat64
	var err error
	if clientID > 0 {
		err = r.db.QueryRowContext(ctx, `SELECT AVG(latency_ms) FROM request_log WHERE client_id = ? AND latency_ms IS NOT NULL AND created_at >= DATE_SUB(NOW(), INTERVAL ? HOUR)`, clientID, hours).Scan(&avg)
	} else {
		err = r.db.QueryRowContext(ctx, `SELECT AVG(latency_ms) FROM request_log WHERE latency_ms IS NOT NULL AND created_at >= DATE_SUB(NOW(), INTERVAL ? HOUR)`, hours).Scan(&avg)
	}
	if err != nil {
		return 0, err
	}
	if !avg.Valid {
		return 0, nil
	}
	return int64(avg.Float64), nil
}

func (r *GatewayRepository) SumTokens(ctx context.Context, hours int, clientID int64) (prompt, completion, total int64, err error) {
	if clientID > 0 {
		err = r.db.QueryRowContext(ctx, `SELECT COALESCE(SUM(prompt_tokens),0), COALESCE(SUM(completion_tokens),0), COALESCE(SUM(total_tokens),0) FROM request_log WHERE client_id = ? AND created_at >= DATE_SUB(NOW(), INTERVAL ? HOUR)`, clientID, hours).Scan(&prompt, &completion, &total)
		return
	}
	err = r.db.QueryRowContext(ctx, `SELECT COALESCE(SUM(prompt_tokens),0), COALESCE(SUM(completion_tokens),0), COALESCE(SUM(total_tokens),0) FROM request_log WHERE created_at >= DATE_SUB(NOW(), INTERVAL ? HOUR)`, hours).Scan(&prompt, &completion, &total)
	return
}

func (r *GatewayRepository) SumTokens24h(ctx context.Context) (prompt, completion, total int64, err error) {
	err = r.db.QueryRowContext(ctx, `SELECT COALESCE(SUM(prompt_tokens),0), COALESCE(SUM(completion_tokens),0), COALESCE(SUM(total_tokens),0) FROM request_log WHERE created_at >= DATE_SUB(NOW(), INTERVAL 24 HOUR)`).Scan(&prompt, &completion, &total)
	return
}

type HourlyStat struct {
	Hour             string  `json:"hour"`
	Count            int64   `json:"count"`
	Success          int64   `json:"success"`
	AvgLatency       float64 `json:"avg_latency"`
	PromptTokens     int64   `json:"prompt_tokens"`
	CompletionTokens int64   `json:"completion_tokens"`
	TotalTokens      int64   `json:"total_tokens"`
}

type ModelStat struct {
	ModelCode string `json:"model_code"`
	Count     int64  `json:"count"`
}

func dashboardTrendBucketSQL(column string, hours int) (selectExpr, groupBy string) {
	if hours >= 168 {
		selectExpr = "DATE_FORMAT(" + column + ", '%Y-%m-%d 00:00')"
		groupBy = "DATE_FORMAT(" + column + ", '%Y-%m-%d')"
		return
	}
	if hours > 72 {
		selectExpr = "CONCAT(DATE_FORMAT(" + column + ", '%Y-%m-%d '), LPAD(FLOOR(HOUR(" + column + ") / 6) * 6, 2, '0'), ':00')"
		groupBy = "DATE_FORMAT(" + column + ", '%Y-%m-%d'), FLOOR(HOUR(" + column + ") / 6)"
		return
	}
	selectExpr = "DATE_FORMAT(" + column + ", '%Y-%m-%d %H:00')"
	groupBy = selectExpr
	return
}

func (r *GatewayRepository) GetRequestTrend(ctx context.Context, hours int, clientID int64) ([]HourlyStat, error) {
	args := []any{hours}
	bucketExpr, groupBy := dashboardTrendBucketSQL("created_at", hours)

	query := `
		SELECT
			` + bucketExpr + ` as hour,
			COUNT(*) as total_count,
			COALESCE(SUM(success), 0) as success_count,
			COALESCE(AVG(latency_ms), 0) as avg_latency,
			COALESCE(SUM(prompt_tokens), 0) as prompt_tokens,
			COALESCE(SUM(completion_tokens), 0) as completion_tokens,
			COALESCE(SUM(total_tokens), 0) as total_tokens
		FROM request_log
		WHERE created_at >= DATE_SUB(NOW(), INTERVAL ? HOUR)`
	if clientID > 0 {
		query += " AND client_id = ?"
		args = append(args, clientID)
	}
	query += `
		GROUP BY ` + groupBy + `
		ORDER BY hour ASC`
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []HourlyStat
	for rows.Next() {
		var stat HourlyStat
		if err := rows.Scan(&stat.Hour, &stat.Count, &stat.Success, &stat.AvgLatency, &stat.PromptTokens, &stat.CompletionTokens, &stat.TotalTokens); err != nil {
			return nil, err
		}
		results = append(results, stat)
	}
	return results, nil
}

type KeyTrend struct {
	Hour    string `json:"hour"`
	KeyName string `json:"key_name"`
	Count   int64  `json:"count"`
	Tokens  int64  `json:"tokens"`
}

func (r *GatewayRepository) GetRequestTrendByKey(ctx context.Context, hours int) ([]KeyTrend, error) {
	bucketExpr, groupBy := dashboardTrendBucketSQL("rl.created_at", hours)
	query := `
		SELECT
			` + bucketExpr + ` as hour,
			ac.name as key_name,
			COUNT(*) as count,
			COALESCE(SUM(rl.total_tokens), 0) as tokens
		FROM request_log rl
		INNER JOIN api_client ac ON rl.client_id = ac.id
		WHERE rl.created_at >= DATE_SUB(NOW(), INTERVAL ? HOUR)
		GROUP BY ` + groupBy + `, ac.name
		ORDER BY hour ASC, count DESC
	`
	rows, err := r.db.QueryContext(ctx, query, hours)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []KeyTrend
	for rows.Next() {
		var trend KeyTrend
		if err := rows.Scan(&trend.Hour, &trend.KeyName, &trend.Count, &trend.Tokens); err != nil {
			return nil, err
		}
		results = append(results, trend)
	}
	return results, rows.Err()
}

func (r *GatewayRepository) GetModelDistribution(ctx context.Context, hours int, clientID int64) ([]ModelStat, error) {
	args := []any{hours}
	query := `
		SELECT 
			COALESCE(virtual_model_code, upstream_model_name, 'unknown') as model_code,
			COUNT(*) as count
		FROM request_log 
		WHERE created_at >= DATE_SUB(NOW(), INTERVAL ? HOUR)`
	if clientID > 0 {
		query += " AND client_id = ?"
		args = append(args, clientID)
	}
	query += `
		GROUP BY model_code
		ORDER BY count DESC
		LIMIT 10`
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []ModelStat
	for rows.Next() {
		var stat ModelStat
		if err := rows.Scan(&stat.ModelCode, &stat.Count); err != nil {
			return nil, err
		}
		results = append(results, stat)
	}
	return results, nil
}

type KeyModelStat struct {
	KeyName   string `json:"key_name"`
	ModelCode string `json:"model_code"`
	Count     int64  `json:"count"`
}

func (r *GatewayRepository) GetModelDistributionByKey(ctx context.Context, hours int) ([]KeyModelStat, error) {
	query := `
		SELECT 
			COALESCE(ac.name, '未知') as key_name,
			COALESCE(rl.virtual_model_code, rl.upstream_model_name, 'unknown') as model_code,
			COUNT(*) as count
		FROM request_log rl
		LEFT JOIN api_client ac ON rl.client_id = ac.id
		WHERE rl.created_at >= DATE_SUB(NOW(), INTERVAL ? HOUR)
		GROUP BY key_name, model_code
		ORDER BY count DESC
		LIMIT 20
	`
	rows, err := r.db.QueryContext(ctx, query, hours)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []KeyModelStat
	for rows.Next() {
		var stat KeyModelStat
		if err := rows.Scan(&stat.KeyName, &stat.ModelCode, &stat.Count); err != nil {
			return nil, err
		}
		results = append(results, stat)
	}
	return results, rows.Err()
}

func (r *GatewayRepository) CountSuccessRequests24h(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM request_log WHERE success = 1 AND created_at >= DATE_SUB(NOW(), INTERVAL 24 HOUR)`).Scan(&count)
	return count, err
}

func (r *GatewayRepository) AvgLatency24h(ctx context.Context) (int64, error) {
	var avg sql.NullFloat64
	err := r.db.QueryRowContext(ctx, `SELECT AVG(latency_ms) FROM request_log WHERE latency_ms IS NOT NULL AND created_at >= DATE_SUB(NOW(), INTERVAL 24 HOUR)`).Scan(&avg)
	if err != nil {
		return 0, err
	}
	if !avg.Valid {
		return 0, nil
	}
	return int64(avg.Float64), nil
}

func (r *GatewayRepository) CountDebugSessions24h(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM request_log WHERE is_debug_logged = 1 AND created_at >= DATE_SUB(NOW(), INTERVAL 24 HOUR)`).Scan(&count)
	return count, err
}

func (r *GatewayRepository) GetClientByHash(ctx context.Context, hash string) (*GatewayClient, error) {
	var item GatewayClient
	var expiresAt, quotaDisabledAt sql.NullTime
	err := r.db.QueryRowContext(ctx, `SELECT id, name, key_prefix, api_key_hash, status, expires_at, daily_request_limit, monthly_request_limit, daily_token_limit, monthly_token_limit, current_daily_requests, current_monthly_requests, current_daily_tokens, current_monthly_tokens, quota_disabled_at FROM api_client WHERE api_key_hash = ? LIMIT 1`, hash).Scan(&item.ID, &item.Name, &item.KeyPrefix, &item.APIKeyHash, &item.Status, &expiresAt, &item.DailyRequestLimit, &item.MonthlyRequestLimit, &item.DailyTokenLimit, &item.MonthlyTokenLimit, &item.CurrentDailyRequests, &item.CurrentMonthlyRequests, &item.CurrentDailyTokens, &item.CurrentMonthlyTokens, &quotaDisabledAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if expiresAt.Valid {
		item.ExpiresAt = &expiresAt.Time
	}
	if quotaDisabledAt.Valid {
		item.QuotaDisabledAt = &quotaDisabledAt.Time
	}
	return &item, nil
}

func (r *GatewayRepository) ClientCanAccessModel(ctx context.Context, clientID, virtualModelID int64) (bool, error) {
	var total int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM api_client_model WHERE client_id = ?`, clientID).Scan(&total); err != nil {
		return false, err
	}
	if total == 0 {
		return true, nil
	}
	var allowed int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM api_client_model WHERE client_id = ? AND virtual_model_id = ?`, clientID, virtualModelID).Scan(&allowed)
	if err != nil {
		return false, err
	}
	return allowed > 0, nil
}

func (r *GatewayRepository) TouchClientLastUsed(ctx context.Context, clientID int64) error {
	_, err := r.db.ExecContext(ctx, `UPDATE api_client SET last_used_at = ?, updated_at = ? WHERE id = ?`, time.Now().UTC(), time.Now().UTC(), clientID)
	return err
}

func (r *GatewayRepository) IncrementClientUsage(ctx context.Context, clientID int64, tokens int) error {
	_, err := r.db.ExecContext(ctx, `UPDATE api_client SET current_daily_requests = current_daily_requests + 1, current_monthly_requests = current_monthly_requests + 1, current_daily_tokens = current_daily_tokens + ?, current_monthly_tokens = current_monthly_tokens + ?, updated_at = ? WHERE id = ?`,
		tokens, tokens, time.Now().UTC(), clientID)
	return err
}

func (r *GatewayRepository) CheckAndIncrementUsage(ctx context.Context, clientID int64, dailyReqLimit, monthlyReqLimit, dailyTokenLimit, monthlyTokenLimit *int64) (bool, string, error) {
	var currentDailyRequests, currentMonthlyRequests, currentDailyTokens, currentMonthlyTokens int64
	err := r.db.QueryRowContext(ctx, `SELECT current_daily_requests, current_monthly_requests, current_daily_tokens, current_monthly_tokens FROM api_client WHERE id = ?`, clientID).Scan(&currentDailyRequests, &currentMonthlyRequests, &currentDailyTokens, &currentMonthlyTokens)
	if err != nil {
		return false, "", err
	}

	if dailyReqLimit != nil && currentDailyRequests >= *dailyReqLimit {
		return false, "daily request limit exceeded", nil
	}
	if monthlyReqLimit != nil && currentMonthlyRequests >= *monthlyReqLimit {
		return false, "monthly request limit exceeded", nil
	}
	if dailyTokenLimit != nil && currentDailyTokens >= *dailyTokenLimit {
		return false, "daily token limit exceeded", nil
	}
	if monthlyTokenLimit != nil && currentMonthlyTokens >= *monthlyTokenLimit {
		return false, "monthly token limit exceeded", nil
	}

	_, err = r.db.ExecContext(ctx, `UPDATE api_client SET current_daily_requests = current_daily_requests + 1, current_monthly_requests = current_monthly_requests + 1, updated_at = ? WHERE id = ?`,
		time.Now().UTC(), clientID)
	if err != nil {
		return false, "", err
	}
	return true, "", nil
}

func (r *GatewayRepository) DisableClient(ctx context.Context, clientID int64) error {
	_, err := r.db.ExecContext(ctx, `UPDATE api_client SET status = 'disabled', quota_disabled_at = ?, updated_at = ? WHERE id = ?`, time.Now().UTC(), time.Now().UTC(), clientID)
	return err
}

func (r *GatewayRepository) SetClientStatus(ctx context.Context, clientID int64, status string, clearQuotaDisabled bool) error {
	if clearQuotaDisabled {
		_, err := r.db.ExecContext(ctx, `UPDATE api_client SET status = ?, quota_disabled_at = NULL, updated_at = ? WHERE id = ?`, status, time.Now().UTC(), clientID)
		return err
	}
	_, err := r.db.ExecContext(ctx, `UPDATE api_client SET status = ?, updated_at = ? WHERE id = ?`, status, time.Now().UTC(), clientID)
	return err
}

func (r *GatewayRepository) ListVisibleModels(ctx context.Context) ([]map[string]any, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, model_code, display_name FROM virtual_model WHERE visible = 1 AND status = 'active' ORDER BY sort_order ASC, id ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]map[string]any, 0)
	for rows.Next() {
		var id int64
		var code, displayName string
		if err := rows.Scan(&id, &code, &displayName); err != nil {
			return nil, err
		}
		items = append(items, map[string]any{"id": id, "model_code": code, "display_name": displayName})
	}
	return items, rows.Err()
}

func (r *GatewayRepository) GetVisibleModelCodeByID(ctx context.Context, id int64) (string, error) {
	var code string
	err := r.db.QueryRowContext(ctx, `SELECT model_code FROM virtual_model WHERE id = ? AND visible = 1 AND status = 'active' LIMIT 1`, id).Scan(&code)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", err
	}
	return code, nil
}

func (r *GatewayRepository) ResolveOpenAIModelRoute(ctx context.Context, modelCode string) (*ModelRoute, error) {
	route, err := r.resolveOpenAIModelRoute(ctx, modelCode, true)
	if err != nil {
		return nil, err
	}
	if route == nil {
		route, err = r.resolveOpenAIModelRoute(ctx, modelCode, false)
		if err != nil {
			return nil, err
		}
	}
	return route, nil
}

func (r *GatewayRepository) resolveOpenAIModelRoute(ctx context.Context, modelCode string, requireCurrentBinding bool) (*ModelRoute, error) {
	var query string
	var args []any

	if requireCurrentBinding {
		query = `
			SELECT
				vm.id,
				vm.model_code,
				vm.display_name,
				rs.current_binding_id,
				p.id,
				p.name,
				p.base_url,
				p.auth_type,
				vmb.upstream_model_name,
				vmb.provider_key_id,
				vmb.priority,
				vmb.enabled,
				vm.default_params_json
			FROM virtual_model vm
			INNER JOIN route_state rs ON rs.virtual_model_id = vm.id
			INNER JOIN virtual_model_binding vmb ON vmb.id = rs.current_binding_id
			INNER JOIN provider p ON p.id = vmb.provider_id
			WHERE vm.model_code = ? AND vm.visible = 1 AND vm.status = 'active' AND p.enabled = 1 AND vmb.enabled = 1
			LIMIT 1`
		args = []any{modelCode}
	} else {
		query = `
			SELECT
				vm.id,
				vm.model_code,
				vm.display_name,
				vmb.id as binding_id,
				p.id,
				p.name,
				p.base_url,
				p.auth_type,
				vmb.upstream_model_name,
				vmb.provider_key_id,
				vmb.priority,
				vmb.enabled,
				vm.default_params_json
			FROM virtual_model vm
			INNER JOIN virtual_model_binding vmb ON vmb.virtual_model_id = vm.id
			INNER JOIN provider p ON p.id = vmb.provider_id
			WHERE vm.model_code = ? AND vm.visible = 1 AND vm.status = 'active' AND p.enabled = 1 AND vmb.enabled = 1
			ORDER BY vmb.priority ASC
			LIMIT 1`
		args = []any{modelCode}
	}

	row := r.db.QueryRowContext(ctx, query, args...)
	var item ModelRoute
	var currentBindingID sql.NullInt64
	var providerKeyID sql.NullInt64
	var bindingID sql.NullInt64
	err := row.Scan(&item.VirtualModelID, &item.ModelCode, &item.DisplayName, &currentBindingID, &item.ProviderID, &item.ProviderName, &item.BaseURL, &item.AuthType, &item.UpstreamModelName, &providerKeyID, &item.BindingPriority, &item.BindingEnabled, &item.DefaultParamsJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("resolve openai model route: %w", err)
	}
	if currentBindingID.Valid {
		item.CurrentBindingID = &currentBindingID.Int64
	} else if bindingID.Valid {
		item.CurrentBindingID = &bindingID.Int64
	}
	if providerKeyID.Valid {
		item.ProviderKeyID = &providerKeyID.Int64
	}

	return &item, nil
}

func (r *GatewayRepository) DiagnoseModelRoute(ctx context.Context, modelCode string) error {
	Log.Debug().Str("model_code", modelCode).Msg("running diagnose")

	var vmID int64
	var vmCode string
	var vmVisible int
	var vmStatus string
	err := r.db.QueryRowContext(ctx, `SELECT id, model_code, visible, status FROM virtual_model WHERE model_code = ?`, modelCode).Scan(&vmID, &vmCode, &vmVisible, &vmStatus)
	if err != nil {
		Log.Debug().Str("model_code", modelCode).Str("error", err.Error()).Msg("virtual model not found in DB")
		return err
	}
	Log.Debug().Str("model_code", modelCode).Int64("vm_id", vmID).Int("visible", vmVisible).Str("status", vmStatus).Msg("virtual model found")

	var currentBindingID sql.NullInt64
	var routeStatus string
	err = r.db.QueryRowContext(ctx, `SELECT current_binding_id, route_status FROM route_state WHERE virtual_model_id = ?`, vmID).Scan(&currentBindingID, &routeStatus)
	if err != nil {
		Log.Debug().Str("model_code", modelCode).Str("error", err.Error()).Msg("route state not found")
		return err
	}
	Log.Debug().Str("model_code", modelCode).Int64("current_binding_id", currentBindingID.Int64).Str("route_status", routeStatus).Msg("route state found")

	if !currentBindingID.Valid {
		Log.Debug().Str("model_code", modelCode).Msg("current_binding_id is NULL")
		return nil
	}

	var providerID int64
	var bindingEnabled bool
	err = r.db.QueryRowContext(ctx, `SELECT provider_id, enabled FROM virtual_model_binding WHERE id = ?`, currentBindingID.Int64).Scan(&providerID, &bindingEnabled)
	if err != nil {
		Log.Debug().Str("model_code", modelCode).Str("error", err.Error()).Msg("binding not found")
		return err
	}
	Log.Debug().Str("model_code", modelCode).Int64("provider_id", providerID).Bool("binding_enabled", bindingEnabled).Msg("binding found")

	var providerEnabled bool
	err = r.db.QueryRowContext(ctx, `SELECT enabled FROM provider WHERE id = ?`, providerID).Scan(&providerEnabled)
	if err != nil {
		Log.Debug().Str("model_code", modelCode).Str("error", err.Error()).Msg("provider not found")
		return err
	}
	Log.Debug().Str("model_code", modelCode).Int64("provider_id", providerID).Bool("provider_enabled", providerEnabled).Msg("provider found")

	Log.Debug().Bool("vm_visible", vmVisible == 1).Bool("vm_active", vmStatus == "active").Bool("binding_enabled", bindingEnabled).Bool("provider_enabled", providerEnabled).Msg("all conditions")

	return nil
}

func (r *GatewayRepository) ResolveOpenAIFallbackRoute(ctx context.Context, virtualModelID int64, excludeBindingID int64) (*ModelRoute, error) {
	query := `
		SELECT
			vm.id,
			vm.model_code,
			vm.display_name,
			vmb.id,
			p.id,
			p.name,
			p.base_url,
			p.auth_type,
			vmb.upstream_model_name,
			vmb.provider_key_id,
			vmb.priority,
			vmb.enabled,
			vm.default_params_json
		FROM virtual_model vm
		INNER JOIN virtual_model_binding vmb ON vmb.virtual_model_id = vm.id
		INNER JOIN provider p ON p.id = vmb.provider_id
		LEFT JOIN circuit_breaker_state cbs ON cbs.provider_id = p.id AND cbs.virtual_model_id = vm.id
		WHERE vm.id = ?
		  AND vmb.enabled = 1
		  AND p.enabled = 1
		  AND vmb.id <> ?
		  AND (cbs.state IS NULL OR cbs.state <> 'open')
		ORDER BY vmb.priority ASC, vmb.id ASC
		LIMIT 1`
	row := r.db.QueryRowContext(ctx, query, virtualModelID, excludeBindingID)
	var item ModelRoute
	var currentBindingID sql.NullInt64
	var providerKeyID sql.NullInt64
	err := row.Scan(&item.VirtualModelID, &item.ModelCode, &item.DisplayName, &currentBindingID, &item.ProviderID, &item.ProviderName, &item.BaseURL, &item.AuthType, &item.UpstreamModelName, &providerKeyID, &item.BindingPriority, &item.BindingEnabled, &item.DefaultParamsJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("resolve openai fallback route: %w", err)
	}
	if currentBindingID.Valid {
		item.CurrentBindingID = &currentBindingID.Int64
	}
	if providerKeyID.Valid {
		item.ProviderKeyID = &providerKeyID.Int64
	}
	return &item, nil
}

func (r *GatewayRepository) ListOpenAIFallbackRoutes(ctx context.Context, virtualModelID int64, excludeBindingIDs []int64) ([]ModelRoute, error) {
	query := `
		SELECT
			vm.id,
			vm.model_code,
			vm.display_name,
			vmb.id,
			p.id,
			p.name,
			p.base_url,
			p.auth_type,
			vmb.upstream_model_name,
			vmb.provider_key_id,
			vmb.priority,
			vmb.enabled,
			vm.default_params_json
		FROM virtual_model vm
		INNER JOIN virtual_model_binding vmb ON vmb.virtual_model_id = vm.id
		INNER JOIN provider p ON p.id = vmb.provider_id
		LEFT JOIN circuit_breaker_state cbs ON cbs.provider_id = p.id AND cbs.virtual_model_id = vm.id
		WHERE vm.id = ?
		  AND vmb.enabled = 1
		  AND p.enabled = 1
		  AND (cbs.state IS NULL OR cbs.state <> 'open')`
	args := []any{virtualModelID}
	if len(excludeBindingIDs) > 0 {
		query += ` AND vmb.id NOT IN (` + inClausePlaceholders(len(excludeBindingIDs)) + `)`
		for _, id := range excludeBindingIDs {
			args = append(args, id)
		}
	}
	query += ` ORDER BY vmb.priority ASC, vmb.id ASC`
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list openai fallback routes: %w", err)
	}
	defer rows.Close()
	items := make([]ModelRoute, 0)
	for rows.Next() {
		var item ModelRoute
		var bindingID sql.NullInt64
		var providerKeyID sql.NullInt64
		if err := rows.Scan(&item.VirtualModelID, &item.ModelCode, &item.DisplayName, &bindingID, &item.ProviderID, &item.ProviderName, &item.BaseURL, &item.AuthType, &item.UpstreamModelName, &providerKeyID, &item.BindingPriority, &item.BindingEnabled, &item.DefaultParamsJSON); err != nil {
			return nil, fmt.Errorf("scan openai fallback route: %w", err)
		}
		if bindingID.Valid {
			item.CurrentBindingID = &bindingID.Int64
		}
		if providerKeyID.Valid {
			item.ProviderKeyID = &providerKeyID.Int64
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func inClausePlaceholders(count int) string {
	if count <= 0 {
		return ""
	}
	parts := make([]string, count)
	for i := range parts {
		parts[i] = "?"
	}
	return strings.Join(parts, ",")
}

func (r *GatewayRepository) InsertRequestLog(ctx context.Context, input RequestLogInput) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO request_log (trace_id, protocol_type, client_id, virtual_model_id, virtual_model_code, requested_model, binding_id, provider_id, provider_key_id, upstream_model_name, request_summary_json, response_summary_json, status_code, success, latency_ms, prompt_tokens, completion_tokens, total_tokens, error_code, error_message, is_debug_logged, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, input.TraceID, input.ProtocolType, nullableInt64(input.ClientID), nullableInt64(input.VirtualModelID), nullableString(input.VirtualModelCode), nullableString(input.RequestedModel), nullableInt64(input.BindingID), nullableInt64(input.ProviderID), nullableInt64(input.ProviderKeyID), nullableString(input.UpstreamModelName), normalizeJSON(input.RequestSummaryJSON), normalizeJSON(input.ResponseSummaryJSON), nullableInt(input.StatusCode), input.Success, nullableInt(input.LatencyMS), nullableInt(input.PromptTokens), nullableInt(input.CompletionTokens), nullableInt(input.TotalTokens), nullableString(input.ErrorCode), nullableString(input.ErrorMessage), input.IsDebugLogged, time.Now())
	return err
}

func (r *GatewayRepository) ListRequestLogs(ctx context.Context, filters RequestLogFilters) ([]RequestLogRecord, int, error) {
	Log.Debug().Str("query", "ListRequestLogs").Msg("starting request logs query")
	countQuery := `SELECT COUNT(*) FROM request_log WHERE 1=1`
	query := `SELECT rl.id, rl.trace_id, rl.protocol_type, rl.client_id, COALESCE(ac.name, '') as key_name, rl.virtual_model_id, rl.virtual_model_code, vm.display_name, rl.requested_model, rl.binding_id, vmb.upstream_model_name as route_name, rl.provider_id, p.name as provider_name, rl.upstream_model_name, rl.status_code, rl.success, rl.latency_ms, rl.prompt_tokens, rl.completion_tokens, rl.total_tokens, rl.error_code, rl.error_message, rl.request_summary_json, rl.created_at FROM request_log rl LEFT JOIN api_client ac ON rl.client_id = ac.id LEFT JOIN virtual_model vm ON rl.virtual_model_id = vm.id LEFT JOIN virtual_model_binding vmb ON rl.binding_id = vmb.id LEFT JOIN provider p ON rl.provider_id = p.id WHERE 1=1`
	args := make([]any, 0)
	if filters.ClientID != nil {
		countQuery += ` AND client_id = ?`
		query += ` AND rl.client_id = ?`
		args = append(args, *filters.ClientID)
	}
	if filters.VirtualModelCode != "" {
		countQuery += ` AND virtual_model_code = ?`
		query += ` AND rl.virtual_model_code = ?`
		args = append(args, filters.VirtualModelCode)
	}
	if filters.Success != nil {
		countQuery += ` AND success = ?`
		query += ` AND rl.success = ?`
		args = append(args, *filters.Success)
	}
	if filters.TimeRange != "" {
		hours := 24
		switch filters.TimeRange {
		case "1h":
			hours = 1
		case "6h":
			hours = 6
		case "1d":
			hours = 24
		case "3d":
			hours = 72
		case "7d":
			hours = 168
		}
		countQuery += ` AND created_at >= DATE_SUB(NOW(), INTERVAL ? HOUR)`
		query += ` AND rl.created_at >= DATE_SUB(NOW(), INTERVAL ? HOUR)`
		args = append(args, hours)
	} else if filters.StartTime != nil {
		countQuery += ` AND created_at >= ?`
		query += ` AND rl.created_at >= ?`
		args = append(args, *filters.StartTime)
	}
	if filters.EndTime != nil {
		countQuery += ` AND created_at <= ?`
		query += ` AND rl.created_at <= ?`
		args = append(args, *filters.EndTime)
	}
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return []RequestLogRecord{}, 0, nil
	}
	offset := (filters.Page - 1) * filters.Limit
	query += ` ORDER BY rl.created_at DESC LIMIT ? OFFSET ?`
	args = append(args, filters.Limit, offset)
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	items := make([]RequestLogRecord, 0)
	for rows.Next() {
		var item RequestLogRecord
		var clientID, virtualModelID, bindingID, providerID sql.NullInt64
		var keyName, modelCode, virtualModelName, requestedModel, routeName, providerName, upstreamModel, errorCode, errorMessage, requestSummary sql.NullString
		var statusCode, latency, promptTokens, completionTokens, totalTokens sql.NullInt64
		if err := rows.Scan(&item.ID, &item.TraceID, &item.ProtocolType, &clientID, &keyName, &virtualModelID, &modelCode, &virtualModelName, &requestedModel, &bindingID, &routeName, &providerID, &providerName, &upstreamModel, &statusCode, &item.Success, &latency, &promptTokens, &completionTokens, &totalTokens, &errorCode, &errorMessage, &requestSummary, &item.CreatedAt); err != nil {
			return nil, 0, err
		}
		if clientID.Valid {
			v := clientID.Int64
			item.ClientID = &v
		}
		if keyName.Valid {
			item.KeyName = keyName.String
		}
		if virtualModelID.Valid {
			v := virtualModelID.Int64
			item.VirtualModelID = &v
		}
		if modelCode.Valid {
			v := modelCode.String
			item.VirtualModelCode = &v
		}
		if virtualModelName.Valid {
			v := virtualModelName.String
			item.VirtualModelName = &v
		}
		if requestedModel.Valid {
			v := requestedModel.String
			item.RequestedModel = &v
		}
		if bindingID.Valid {
			v := bindingID.Int64
			item.BindingID = &v
		}
		if routeName.Valid {
			v := routeName.String
			item.RouteName = &v
		}
		if providerID.Valid {
			v := providerID.Int64
			item.ProviderID = &v
		}
		if providerName.Valid {
			v := providerName.String
			item.ProviderName = &v
		}
		if upstreamModel.Valid {
			v := upstreamModel.String
			item.UpstreamModel = &v
		}
		if statusCode.Valid {
			v := int(statusCode.Int64)
			item.StatusCode = &v
		}
		if latency.Valid {
			v := int(latency.Int64)
			item.LatencyMS = &v
		}
		if promptTokens.Valid {
			v := int(promptTokens.Int64)
			item.PromptTokens = &v
		}
		if completionTokens.Valid {
			v := int(completionTokens.Int64)
			item.CompletionTokens = &v
		}
		if totalTokens.Valid {
			v := int(totalTokens.Int64)
			item.TotalTokens = &v
		}
		if errorCode.Valid {
			v := errorCode.String
			item.ErrorCode = &v
		}
		if errorMessage.Valid {
			v := errorMessage.String
			item.ErrorMessage = &v
		}
		if requestSummary.Valid {
			v := requestSummary.String
			item.RequestSummary = &v
		}
		items = append(items, item)
	}
	Log.Debug().Int("total_items", len(items)).Str("query", "ListRequestLogs").Msg("request logs query completed")
	if len(items) > 0 {
		Log.Debug().Str("first_item_key_name", items[0].KeyName).Msg("first item key name")
	}
	return items, total, rows.Err()
}

func (r *GatewayRepository) GetRequestLogByID(ctx context.Context, id int64) (*RequestLogRecord, error) {
	row := r.db.QueryRowContext(ctx, `SELECT rl.id, rl.trace_id, rl.protocol_type, rl.client_id, COALESCE(ac.name, ''), rl.virtual_model_id, rl.virtual_model_code, vm.display_name, rl.requested_model, rl.binding_id, vmb.upstream_model_name as route_name, rl.provider_id, p.name as provider_name, rl.upstream_model_name, rl.status_code, rl.success, rl.latency_ms, rl.prompt_tokens, rl.completion_tokens, rl.total_tokens, rl.error_code, rl.error_message, rl.request_summary_json, rl.response_summary_json, rl.created_at FROM request_log rl LEFT JOIN api_client ac ON rl.client_id = ac.id LEFT JOIN virtual_model vm ON rl.virtual_model_id = vm.id LEFT JOIN virtual_model_binding vmb ON rl.binding_id = vmb.id LEFT JOIN provider p ON rl.provider_id = p.id WHERE rl.id = ? LIMIT 1`, id)
	var item RequestLogRecord
	var clientID, virtualModelID, bindingID, providerID sql.NullInt64
	var keyName, modelCode, virtualModelName, requestedModel, routeName, providerName, upstreamModel, errorCode, errorMessage, requestSummary, responseSummary sql.NullString
	var statusCode, latency, promptTokens, completionTokens, totalTokens sql.NullInt64
	if err := row.Scan(&item.ID, &item.TraceID, &item.ProtocolType, &clientID, &keyName, &virtualModelID, &modelCode, &virtualModelName, &requestedModel, &bindingID, &routeName, &providerID, &providerName, &upstreamModel, &statusCode, &item.Success, &latency, &promptTokens, &completionTokens, &totalTokens, &errorCode, &errorMessage, &requestSummary, &responseSummary, &item.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if clientID.Valid {
		v := clientID.Int64
		item.ClientID = &v
	}
	if keyName.Valid {
		item.KeyName = keyName.String
	}
	if virtualModelID.Valid {
		v := virtualModelID.Int64
		item.VirtualModelID = &v
	}
	if modelCode.Valid {
		v := modelCode.String
		item.VirtualModelCode = &v
	}
	if virtualModelName.Valid {
		v := virtualModelName.String
		item.VirtualModelName = &v
	}
	if requestedModel.Valid {
		v := requestedModel.String
		item.RequestedModel = &v
	}
	if bindingID.Valid {
		v := bindingID.Int64
		item.BindingID = &v
	}
	if routeName.Valid {
		v := routeName.String
		item.RouteName = &v
	}
	if providerID.Valid {
		v := providerID.Int64
		item.ProviderID = &v
	}
	if providerName.Valid {
		v := providerName.String
		item.ProviderName = &v
	}
	if upstreamModel.Valid {
		v := upstreamModel.String
		item.UpstreamModel = &v
	}
	if statusCode.Valid {
		v := int(statusCode.Int64)
		item.StatusCode = &v
	}
	if latency.Valid {
		v := int(latency.Int64)
		item.LatencyMS = &v
	}
	if promptTokens.Valid {
		v := int(promptTokens.Int64)
		item.PromptTokens = &v
	}
	if completionTokens.Valid {
		v := int(completionTokens.Int64)
		item.CompletionTokens = &v
	}
	if totalTokens.Valid {
		v := int(totalTokens.Int64)
		item.TotalTokens = &v
	}
	if errorCode.Valid {
		v := errorCode.String
		item.ErrorCode = &v
	}
	if errorMessage.Valid {
		v := errorMessage.String
		item.ErrorMessage = &v
	}
	if requestSummary.Valid {
		v := requestSummary.String
		item.RequestSummary = &v
	}
	if responseSummary.Valid {
		v := responseSummary.String
		item.ResponseSummary = &v
	}
	return &item, nil
}

func normalizeJSON(raw json.RawMessage) any {
	if len(raw) == 0 {
		return nil
	}
	return string(raw)
}

func nullableInt64(value *int64) any {
	if value == nil {
		return nil
	}
	return *value
}

func nullableInt(value *int) any {
	if value == nil {
		return nil
	}
	return *value
}

func nullableString(value *string) any {
	if value == nil {
		return nil
	}
	return *value
}

type KeyStat struct {
	KeyName      string  `json:"key_name"`
	RequestCount int64   `json:"request_count"`
	TotalTokens  int64   `json:"total_tokens"`
	SuccessRate  float64 `json:"success_rate"`
}

func (r *GatewayRepository) GetKeyStats(ctx context.Context, hours int) ([]KeyStat, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT 
			ac.name as key_name,
			COUNT(rl.id) as request_count,
			COALESCE(SUM(rl.total_tokens), 0) as total_tokens,
			COALESCE(AVG(CASE WHEN rl.success = 1 THEN 1.0 ELSE 0.0 END), 0) as success_rate
		FROM api_client ac
		LEFT JOIN request_log rl ON ac.id = rl.client_id AND rl.created_at >= DATE_SUB(NOW(), INTERVAL ? HOUR)
		WHERE ac.status = 'active'
		GROUP BY ac.id, ac.name
		ORDER BY request_count DESC
	`, hours)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []KeyStat
	for rows.Next() {
		var stat KeyStat
		if err := rows.Scan(&stat.KeyName, &stat.RequestCount, &stat.TotalTokens, &stat.SuccessRate); err != nil {
			return nil, err
		}
		results = append(results, stat)
	}
	return results, rows.Err()
}
