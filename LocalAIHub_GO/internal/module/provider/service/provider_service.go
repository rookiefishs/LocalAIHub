package service

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"localaihub/localaihub_go/internal/module/provider/repository"
	"localaihub/localaihub_go/internal/pkg/logger"
)

type ProviderService struct {
	repo       *repository.ProviderRepository
	keyService *ProviderKeyService
	audit      interface {
		Log(ctx context.Context, action, targetType string, targetID *int64, details map[string]any, ip, userAgent string)
	}
}

func NewProviderService(repo *repository.ProviderRepository, keyService *ProviderKeyService, audit interface {
	Log(ctx context.Context, action, targetType string, targetID *int64, details map[string]any, ip, userAgent string)
}) *ProviderService {
	return &ProviderService{repo: repo, keyService: keyService, audit: audit}
}

func (s *ProviderService) List(ctx context.Context, page, pageSize int) ([]repository.Provider, int, error) {
	return s.repo.List(ctx, page, pageSize)
}
func (s *ProviderService) Get(ctx context.Context, id int64) (*repository.Provider, error) {
	return s.repo.GetByID(ctx, id)
}
func (s *ProviderService) Create(ctx context.Context, item *repository.Provider, ip, userAgent string) (int64, error) {
	id, err := s.repo.Create(ctx, item)
	if err != nil {
		return id, err
	}
	if item.NewKey != "" && s.keyService != nil {
		_, err = s.keyService.Create(ctx, id, item.NewKey, 1, "", ip, userAgent)
		if err != nil {
			logger.Log.Error().Int64("provider_id", id).Err(err).Msg("failed to create initial provider key")
		}
	}
	if err == nil && s.audit != nil {
		s.audit.Log(ctx, "provider.create", "provider", &id, map[string]any{"name": item.Name, "provider_type": item.ProviderType}, ip, userAgent)
	}
	return id, err
}
func (s *ProviderService) Update(ctx context.Context, item *repository.Provider, ip, userAgent string) error {
	err := s.repo.Update(ctx, item)
	if err == nil && s.audit != nil {
		targetID := item.ID
		s.audit.Log(ctx, "provider.update", "provider", &targetID, map[string]any{"name": item.Name, "enabled": item.Enabled}, ip, userAgent)
	}
	return err
}
func (s *ProviderService) UpdateStatus(ctx context.Context, id int64, enabled bool, ip, userAgent string) error {
	err := s.repo.UpdateStatus(ctx, id, enabled)
	if err == nil && s.audit != nil {
		targetID := id
		s.audit.Log(ctx, "provider.update_status", "provider", &targetID, map[string]any{"enabled": enabled}, ip, userAgent)
	}
	return err
}
func (s *ProviderService) CountActive(ctx context.Context) (int64, error) {
	return s.repo.CountActive(ctx)
}

func (s *ProviderService) Delete(ctx context.Context, id int64, ip, userAgent string) error {
	check, err := s.repo.CheckDelete(ctx, id)
	if err != nil {
		return err
	}
	if check.BindingCount > 0 {
		return fmt.Errorf("provider is still referenced by %d bindings", check.BindingCount)
	}
	err = s.repo.Delete(ctx, id)
	if err == nil && s.audit != nil {
		targetID := id
		s.audit.Log(ctx, "provider.delete", "provider", &targetID, map[string]any{"id": id}, ip, userAgent)
	}
	return err
}

func (s *ProviderService) TestConnection(ctx context.Context, item *repository.Provider, ip, userAgent string) (map[string]any, error) {
	client := &http.Client{Timeout: time.Duration(item.TimeoutMS) * time.Millisecond}
	start := time.Now()
	testURL := normalizeURL(item.BaseURL, item.ProviderType)

	logger.Log.Info().Int64("provider_id", item.ID).Str("test_url", testURL).Msg("testing provider connection")

	var req *http.Request
	var secret string

	providerKey, secret, err := s.keyService.SelectForRequest(ctx, item.ID)
	if err != nil || providerKey == nil || secret == "" {
		logger.Log.Info().Int64("provider_id", item.ID).Msg("no provider key found, using GET request")
		req, _ = http.NewRequestWithContext(ctx, "GET", testURL, nil)
	} else {
		logger.Log.Info().Int64("provider_id", item.ID).Bool("has_key", true).Str("auth_type", item.AuthType).Msg("using provider key for test")
		chatURL := normalizeURLForChat(item.BaseURL, item.ProviderType)
		if item.ProviderType == "anthropic" {
			reqBody := `{"model":"claude-3-haiku-20240307","max_tokens":1,"messages":[{"role":"user","text":"hi"}]}`
			req, _ = http.NewRequestWithContext(ctx, "POST", chatURL, strings.NewReader(reqBody))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("x-api-key", secret)
		} else {
			reqBody := `{"model":"gpt-4o-mini","max_tokens":1,"messages":[{"role":"user","content":"hi"}]}`
			req, _ = http.NewRequestWithContext(ctx, "POST", chatURL+"/chat/completions", strings.NewReader(reqBody))
			req.Header.Set("Content-Type", "application/json")
			if item.AuthType == "x_api_key" {
				req.Header.Set("x-api-key", secret)
			} else {
				req.Header.Set("Authorization", "Bearer "+secret)
			}
		}
	}

	resp, err := client.Do(req)
	logger.Log.Info().Int64("provider_id", item.ID).Err(err).Str("url", req.URL.String()).Msg("request executed")
	if err != nil {
		logger.Log.Error().Int64("provider_id", item.ID).Err(err).Str("url", testURL).Msg("provider connection failed")
		_ = s.repo.UpdateHealth(ctx, item.ID, "degraded", err.Error())
		if s.audit != nil {
			targetID := item.ID
			s.audit.Log(ctx, "provider.test_connection", "provider", &targetID, map[string]any{"success": false, "error": err.Error()}, ip, userAgent)
		}
		return map[string]any{"success": false, "latency_ms": int(time.Since(start).Milliseconds()), "message": err.Error(), "tested_url": testURL}, nil
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	logger.Log.Info().Int64("provider_id", item.ID).Str("url", testURL).Int("status_code", resp.StatusCode).Str("response", string(body)).Msg("provider connection response")

	healthStatus := "healthy"
	message := "connection success"
	if resp.StatusCode >= 400 {
		if secret == "" {
			message = "connection ok (no key configured)"
		} else {
			healthStatus = "degraded"
			message = resp.Status
		}
	}
	_ = s.repo.UpdateHealth(ctx, item.ID, healthStatus, message)
	if s.audit != nil {
		targetID := item.ID
		s.audit.Log(ctx, "provider.test_connection", "provider", &targetID, map[string]any{"success": resp.StatusCode < 400, "status_code": resp.StatusCode}, ip, userAgent)
	}
	return map[string]any{"success": resp.StatusCode < 400, "latency_ms": int(time.Since(start).Milliseconds()), "message": message, "tested_url": testURL}, nil
}

func normalizeURL(baseURL, providerType string) string {
	url := strings.TrimRight(baseURL, "/")
	url = strings.ReplaceAll(url, "/v1/v1", "/v1")
	url = strings.ReplaceAll(url, "/v1/", "/")

	switch providerType {
	case "openai", "proxy":
		if !strings.HasSuffix(url, "/v1") {
			url += "/v1"
		}
		url += "/models"
	case "anthropic":
		url += "/v1/models"
	}
	return url
}

func normalizeURLForChat(baseURL, providerType string) string {
	url := strings.TrimRight(baseURL, "/")
	url = strings.ReplaceAll(url, "/v1/v1", "/v1")
	url = strings.ReplaceAll(url, "/v1/", "/")

	if !strings.HasSuffix(url, "/v1") {
		url += "/v1"
	}
	return url
}
