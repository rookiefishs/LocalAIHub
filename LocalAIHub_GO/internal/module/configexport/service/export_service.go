package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

type ExportData map[string]interface{}

type ImportOptions struct {
	Mode string // "merge" or "replace"
}

type ImportSummary struct {
	ProvidersCreated     int `json:"providers_created"`
	ProvidersUpdated     int `json:"providers_updated"`
	ProviderKeysCreated  int `json:"provider_keys_created"`
	ProviderKeysSkipped  int `json:"provider_keys_skipped"`
	VirtualModelsCreated int `json:"virtual_models_created"`
	VirtualModelsUpdated int `json:"virtual_models_updated"`
	BindingsCreated      int `json:"bindings_created"`
	BindingsSkipped      int `json:"bindings_skipped"`
	ApiClientsCreated    int `json:"api_clients_created"`
	ApiClientsUpdated    int `json:"api_clients_updated"`
}

type ExportService struct {
	db *sql.DB
}

func NewExportService(db *sql.DB) *ExportService {
	return &ExportService{db: db}
}

func (s *ExportService) Export(ctx context.Context) (ExportData, error) {
	result := make(ExportData)

	providers, err := s.exportProviders(ctx)
	if err != nil {
		return nil, fmt.Errorf("export providers: %w", err)
	}
	result["providers"] = providers

	providerKeys, err := s.exportProviderKeys(ctx)
	if err != nil {
		return nil, fmt.Errorf("export provider_keys: %w", err)
	}
	result["provider_keys"] = providerKeys

	virtualModels, err := s.exportVirtualModels(ctx)
	if err != nil {
		return nil, fmt.Errorf("export virtual_models: %w", err)
	}
	result["virtual_models"] = virtualModels

	bindings, err := s.exportBindings(ctx)
	if err != nil {
		return nil, fmt.Errorf("export bindings: %w", err)
	}
	result["bindings"] = bindings

	apiClients, err := s.exportApiClients(ctx)
	if err != nil {
		return nil, fmt.Errorf("export api_clients: %w", err)
	}
	result["api_clients"] = apiClients

	result["exported_at"] = time.Now().UTC().Format(time.RFC3339)

	return result, nil
}

func (s *ExportService) Import(ctx context.Context, data ExportData, opts ImportOptions) (*ImportSummary, error) {
	summary := &ImportSummary{}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if providers, ok := data["providers"].([]interface{}); ok && len(providers) > 0 {
		if err = s.importProviders(ctx, tx, providers, opts, summary); err != nil {
			return nil, fmt.Errorf("import providers: %w", err)
		}
	}

	if providerKeys, ok := data["provider_keys"].([]interface{}); ok && len(providerKeys) > 0 {
		if err = s.importProviderKeys(ctx, tx, providerKeys, opts, summary); err != nil {
			return nil, fmt.Errorf("import provider_keys: %w", err)
		}
	}

	if virtualModels, ok := data["virtual_models"].([]interface{}); ok && len(virtualModels) > 0 {
		if err = s.importVirtualModels(ctx, tx, virtualModels, opts, summary); err != nil {
			return nil, fmt.Errorf("import virtual_models: %w", err)
		}
	}

	if bindings, ok := data["bindings"].([]interface{}); ok && len(bindings) > 0 {
		if err = s.importBindings(ctx, tx, bindings, opts, summary); err != nil {
			return nil, fmt.Errorf("import bindings: %w", err)
		}
	}

	if apiClients, ok := data["api_clients"].([]interface{}); ok && len(apiClients) > 0 {
		if err = s.importApiClients(ctx, tx, apiClients, opts, summary); err != nil {
			return nil, fmt.Errorf("import api_clients: %w", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}

	return summary, nil
}

func (s *ExportService) exportProviders(ctx context.Context) ([]map[string]interface{}, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, name, provider_type, service_type, base_url, auth_type, timeout_ms,
		       enabled, health_status, last_health_check_at, last_health_message, remark,
		       created_at, updated_at
		FROM provider
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		var id int64
		var name, providerType, serviceType, baseURL, authType string
		var timeoutMs int
		var enabled int
		var healthStatus string
		var lastHealthCheckAt sql.NullTime
		var lastHealthMessage, remark sql.NullString
		var createdAt, updatedAt time.Time

		if err := rows.Scan(&id, &name, &providerType, &serviceType, &baseURL, &authType,
			&timeoutMs, &enabled, &healthStatus, &lastHealthCheckAt, &lastHealthMessage,
			&remark, &createdAt, &updatedAt); err != nil {
			return nil, err
		}

		m := map[string]interface{}{
			"id":                  id,
			"name":                name,
			"provider_type":       providerType,
			"service_type":        serviceType,
			"base_url":            baseURL,
			"auth_type":           authType,
			"timeout_ms":          timeoutMs,
			"enabled":             enabled == 1,
			"health_status":       healthStatus,
			"last_health_message": lastHealthMessage.String,
			"remark":              remark.String,
			"created_at":          createdAt,
			"updated_at":          updatedAt,
		}
		if lastHealthCheckAt.Valid {
			m["last_health_check_at"] = lastHealthCheckAt.Time
		}
		results = append(results, m)
	}
	return results, rows.Err()
}

func (s *ExportService) exportProviderKeys(ctx context.Context) ([]map[string]interface{}, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, provider_id, key_masked, status, priority, fail_count,
		       last_used_at, last_error_at, last_error_message, remark,
		       created_at, updated_at
		FROM provider_key
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		var id, providerID int64
		var keyMasked, status string
		var priority, failCount int
		var lastUsedAt, lastErrorAt, lastErrorMessage sql.NullString
		var remark sql.NullString
		var createdAt, updatedAt time.Time

		if err := rows.Scan(&id, &providerID, &keyMasked, &status, &priority, &failCount,
			&lastUsedAt, &lastErrorAt, &lastErrorMessage, &remark, &createdAt, &updatedAt); err != nil {
			return nil, err
		}

		results = append(results, map[string]interface{}{
			"id":                 id,
			"provider_id":        providerID,
			"key_masked":         keyMasked,
			"secret_encrypted":   "***MASKED***",
			"status":             status,
			"priority":           priority,
			"fail_count":         failCount,
			"last_error_message": lastErrorMessage.String,
			"remark":             remark.String,
			"created_at":         createdAt,
			"updated_at":         updatedAt,
		})
	}
	return results, rows.Err()
}

func (s *ExportService) exportVirtualModels(ctx context.Context) ([]map[string]interface{}, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, model_code, display_name, protocol_family, capability_flags,
		       visible, status, sort_order, description, default_params_json, remark,
		       created_at, updated_at
		FROM virtual_model
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		var id int64
		var modelCode, displayName, protocolFamily string
		var capabilityFlags []byte
		var visible int
		var status string
		var sortOrder int
		var description, remark sql.NullString
		var defaultParamsJSON []byte
		var createdAt, updatedAt time.Time

		if err := rows.Scan(&id, &modelCode, &displayName, &protocolFamily, &capabilityFlags,
			&visible, &status, &sortOrder, &description, &defaultParamsJSON, &remark,
			&createdAt, &updatedAt); err != nil {
			return nil, err
		}

		var caps interface{}
		_ = json.Unmarshal(capabilityFlags, &caps)
		var defaults interface{}
		_ = json.Unmarshal(defaultParamsJSON, &defaults)

		m := map[string]interface{}{
			"id":               id,
			"model_code":       modelCode,
			"display_name":     displayName,
			"protocol_family":  protocolFamily,
			"capability_flags": caps,
			"visible":          visible == 1,
			"status":           status,
			"sort_order":       sortOrder,
			"default_params":   defaults,
			"remark":           remark.String,
			"created_at":       createdAt,
			"updated_at":       updatedAt,
		}
		if description.Valid {
			m["description"] = description.String
		}
		results = append(results, m)
	}
	return results, rows.Err()
}

func (s *ExportService) exportBindings(ctx context.Context) ([]map[string]interface{}, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, virtual_model_id, provider_id, provider_key_id, upstream_model_name,
		       priority, is_same_name, enabled, capability_snapshot_json, param_override_json, remark,
		       created_at, updated_at
		FROM virtual_model_binding
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		var id, virtualModelID, providerID int64
		var providerKeyID sql.NullInt64
		var upstreamModelName string
		var priority int
		var isSameName, enabled int
		var capabilitySnapshot, paramOverride []byte
		var remark sql.NullString
		var createdAt, updatedAt time.Time

		if err := rows.Scan(&id, &virtualModelID, &providerID, &providerKeyID, &upstreamModelName,
			&priority, &isSameName, &enabled, &capabilitySnapshot, &paramOverride, &remark,
			&createdAt, &updatedAt); err != nil {
			return nil, err
		}

		var caps interface{}
		_ = json.Unmarshal(capabilitySnapshot, &caps)
		var params interface{}
		_ = json.Unmarshal(paramOverride, &params)

		m := map[string]interface{}{
			"id":                  id,
			"virtual_model_id":    virtualModelID,
			"provider_id":         providerID,
			"upstream_model_name": upstreamModelName,
			"priority":            priority,
			"is_same_name":        isSameName == 1,
			"enabled":             enabled == 1,
			"capability_snapshot": caps,
			"param_override":      params,
			"remark":              remark.String,
			"created_at":          createdAt,
			"updated_at":          updatedAt,
		}
		if providerKeyID.Valid {
			m["provider_key_id"] = providerKeyID.Int64
		}
		results = append(results, m)
	}
	return results, rows.Err()
}

func (s *ExportService) exportApiClients(ctx context.Context) ([]map[string]interface{}, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, name, key_prefix, api_key_hash, plain_key, status, remark,
		       last_used_at, expires_at, allowed_models_json,
		       created_at, updated_at
		FROM api_client
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		var id int64
		var name, keyPrefix, apiKeyHash, plainKey, status string
		var remark sql.NullString
		var lastUsedAt, expiresAt sql.NullTime
		var allowedModelsJSON []byte
		var createdAt, updatedAt time.Time

		if err := rows.Scan(&id, &name, &keyPrefix, &apiKeyHash, &plainKey, &status, &remark,
			&lastUsedAt, &expiresAt, &allowedModelsJSON, &createdAt, &updatedAt); err != nil {
			return nil, err
		}

		var allowedModels interface{}
		if len(allowedModelsJSON) > 0 {
			_ = json.Unmarshal(allowedModelsJSON, &allowedModels)
		}

		m := map[string]interface{}{
			"id":         id,
			"name":       name,
			"key_prefix": keyPrefix,
			"plain_key":  plainKey,
			"status":     status,
			"remark":     remark.String,
			"created_at": createdAt,
			"updated_at": updatedAt,
		}
		if lastUsedAt.Valid {
			m["last_used_at"] = lastUsedAt.Time
		}
		if expiresAt.Valid {
			m["expires_at"] = expiresAt.Time
		}
		if allowedModels != nil {
			m["allowed_models"] = allowedModels
		}
		results = append(results, m)
	}
	return results, rows.Err()
}

func (s *ExportService) importProviders(ctx context.Context, tx *sql.Tx, providers []interface{}, opts ImportOptions, summary *ImportSummary) error {
	for _, p := range providers {
		pm, ok := p.(map[string]interface{})
		if !ok {
			continue
		}

		name := getString(pm, "name")
		if name == "" {
			continue
		}

		var existingID int64
		err := tx.QueryRowContext(ctx, "SELECT id FROM provider WHERE name = ?", name).Scan(&existingID)
		if err == nil {
			if opts.Mode == "replace" {
				_, _ = tx.ExecContext(ctx, `
					UPDATE provider SET provider_type=?, service_type=?, base_url=?, auth_type=?,
					timeout_ms=?, enabled=?, health_status=?, remark=?
					WHERE id=?`,
					getString(pm, "provider_type"),
					getString(pm, "service_type"),
					getString(pm, "base_url"),
					getString(pm, "auth_type"),
					getInt(pm, "timeout_ms"),
					getBoolInt(pm, "enabled"),
					getString(pm, "health_status"),
					getString(pm, "remark"),
					existingID,
				)
				summary.ProvidersUpdated++
			}
			continue
		}

		_, err = tx.ExecContext(ctx, `
			INSERT INTO provider (name, provider_type, service_type, base_url, auth_type,
				timeout_ms, enabled, health_status, remark)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			name,
			getString(pm, "provider_type"),
			getString(pm, "service_type"),
			getString(pm, "base_url"),
			getString(pm, "auth_type"),
			getInt(pm, "timeout_ms"),
			getBoolInt(pm, "enabled"),
			getString(pm, "health_status"),
			getString(pm, "remark"),
		)
		if err != nil {
			return fmt.Errorf("insert provider %s: %w", name, err)
		}
		summary.ProvidersCreated++
	}
	return nil
}

func (s *ExportService) importProviderKeys(ctx context.Context, tx *sql.Tx, keys []interface{}, opts ImportOptions, summary *ImportSummary) error {
	for _, k := range keys {
		km, ok := k.(map[string]interface{})
		if !ok {
			continue
		}

		secret := getString(km, "secret_encrypted")
		if secret == "" || secret == "***MASKED***" {
			summary.ProviderKeysSkipped++
			continue
		}

		providerID := getInt64(km, "provider_id")
		if providerID == 0 {
			summary.ProviderKeysSkipped++
			continue
		}

		_, err := tx.ExecContext(ctx, `
			INSERT INTO provider_key (provider_id, key_masked, secret_encrypted, status, priority,
				fail_count, last_error_message, remark)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			providerID,
			getString(km, "key_masked"),
			secret,
			getString(km, "status"),
			getInt(km, "priority"),
			getInt(km, "fail_count"),
			getString(km, "last_error_message"),
			getString(km, "remark"),
		)
		if err != nil {
			return fmt.Errorf("insert provider_key: %w", err)
		}
		summary.ProviderKeysCreated++
	}
	return nil
}

func (s *ExportService) importVirtualModels(ctx context.Context, tx *sql.Tx, models []interface{}, opts ImportOptions, summary *ImportSummary) error {
	for _, m := range models {
		vm, ok := m.(map[string]interface{})
		if !ok {
			continue
		}

		modelCode := getString(vm, "model_code")
		if modelCode == "" {
			continue
		}

		capsJSON, _ := json.Marshal(vm["capability_flags"])
		defaultsJSON, _ := json.Marshal(vm["default_params"])

		var existingID int64
		err := tx.QueryRowContext(ctx, "SELECT id FROM virtual_model WHERE model_code = ?", modelCode).Scan(&existingID)
		if err == nil {
			if opts.Mode == "replace" {
				_, _ = tx.ExecContext(ctx, `
					UPDATE virtual_model SET display_name=?, protocol_family=?, capability_flags=?,
					visible=?, status=?, sort_order=?, description=?, default_params_json=?, remark=?
					WHERE id=?`,
					getString(vm, "display_name"),
					getString(vm, "protocol_family"),
					capsJSON,
					getBoolInt(vm, "visible"),
					getString(vm, "status"),
					getInt(vm, "sort_order"),
					getString(vm, "description"),
					defaultsJSON,
					getString(vm, "remark"),
					existingID,
				)
				summary.VirtualModelsUpdated++
			}
			continue
		}

		_, err = tx.ExecContext(ctx, `
			INSERT INTO virtual_model (model_code, display_name, protocol_family, capability_flags,
				visible, status, sort_order, description, default_params_json, remark)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			modelCode,
			getString(vm, "display_name"),
			getString(vm, "protocol_family"),
			capsJSON,
			getBoolInt(vm, "visible"),
			getString(vm, "status"),
			getInt(vm, "sort_order"),
			getString(vm, "description"),
			defaultsJSON,
			getString(vm, "remark"),
		)
		if err != nil {
			return fmt.Errorf("insert virtual_model %s: %w", modelCode, err)
		}
		summary.VirtualModelsCreated++
	}
	return nil
}

func (s *ExportService) importBindings(ctx context.Context, tx *sql.Tx, bindings []interface{}, opts ImportOptions, summary *ImportSummary) error {
	for _, b := range bindings {
		bm, ok := b.(map[string]interface{})
		if !ok {
			continue
		}

		virtualModelID := getInt64(bm, "virtual_model_id")
		providerID := getInt64(bm, "provider_id")
		upstreamModelName := getString(bm, "upstream_model_name")

		if virtualModelID == 0 || providerID == 0 || upstreamModelName == "" {
			summary.BindingsSkipped++
			continue
		}

		capsJSON, _ := json.Marshal(bm["capability_snapshot"])
		paramsJSON, _ := json.Marshal(bm["param_override"])

		var existingID int64
		err := tx.QueryRowContext(ctx, `
			SELECT id FROM virtual_model_binding
			WHERE virtual_model_id=? AND provider_id=? AND upstream_model_name=?`,
			virtualModelID, providerID, upstreamModelName).Scan(&existingID)
		if err == nil {
			summary.BindingsSkipped++
			continue
		}

		providerKeyID := getInt64(bm, "provider_key_id")
		_, err = tx.ExecContext(ctx, `
			INSERT INTO virtual_model_binding (virtual_model_id, provider_id, provider_key_id,
				upstream_model_name, priority, is_same_name, enabled, capability_snapshot_json,
				param_override_json, remark)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			virtualModelID,
			providerID,
			zeroIf(providerKeyID == 0),
			upstreamModelName,
			getInt(bm, "priority"),
			getBoolInt(bm, "is_same_name"),
			getBoolInt(bm, "enabled"),
			capsJSON,
			paramsJSON,
			getString(bm, "remark"),
		)
		if err != nil {
			return fmt.Errorf("insert binding: %w", err)
		}
		summary.BindingsCreated++
	}
	return nil
}

func (s *ExportService) importApiClients(ctx context.Context, tx *sql.Tx, clients []interface{}, opts ImportOptions, summary *ImportSummary) error {
	for _, c := range clients {
		cm, ok := c.(map[string]interface{})
		if !ok {
			continue
		}

		name := getString(cm, "name")
		keyPrefix := getString(cm, "key_prefix")
		if name == "" || keyPrefix == "" {
			continue
		}

		var existingID int64
		err := tx.QueryRowContext(ctx, "SELECT id FROM api_client WHERE key_prefix = ?", keyPrefix).Scan(&existingID)
		if err == nil {
			if opts.Mode == "replace" {
				_, _ = tx.ExecContext(ctx, `
					UPDATE api_client SET name=?, api_key_hash=?, plain_key=?, status=?,
					remark=?, allowed_models_json=?
					WHERE id=?`,
					name,
					getString(cm, "api_key_hash"),
					getString(cm, "plain_key"),
					getString(cm, "status"),
					getString(cm, "remark"),
					getJSONBytes(cm, "allowed_models"),
					existingID,
				)
				summary.ApiClientsUpdated++
			}
			continue
		}

		_, err = tx.ExecContext(ctx, `
			INSERT INTO api_client (name, key_prefix, api_key_hash, plain_key, status, remark,
				allowed_models_json)
			VALUES (?, ?, ?, ?, ?, ?, ?)`,
			name,
			keyPrefix,
			getString(cm, "api_key_hash"),
			getString(cm, "plain_key"),
			getString(cm, "status"),
			getString(cm, "remark"),
			getJSONBytes(cm, "allowed_models"),
		)
		if err != nil {
			return fmt.Errorf("insert api_client %s: %w", name, err)
		}
		summary.ApiClientsCreated++
	}
	return nil
}

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func getInt(m map[string]interface{}, key string) int {
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case int:
			return val
		case float64:
			return int(val)
		case int64:
			return int(val)
		}
	}
	return 0
}

func getInt64(m map[string]interface{}, key string) int64 {
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case int64:
			return val
		case float64:
			return int64(val)
		case int:
			return int64(val)
		}
	}
	return 0
}

func getBoolInt(m map[string]interface{}, key string) int {
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case bool:
			if val {
				return 1
			}
			return 0
		case int:
			return val
		case float64:
			return int(val)
		}
	}
	return 0
}

func getJSONBytes(m map[string]interface{}, key string) []byte {
	if v, ok := m[key]; ok {
		b, _ := json.Marshal(v)
		return b
	}
	return []byte("{}")
}

func zeroIf(condition bool) interface{} {
	if condition {
		return nil
	}
	return 0
}
