package service

import (
	"context"
	"fmt"
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
	if check.KeyCount > 0 {
		return fmt.Errorf("provider still has %d keys, please delete keys first", check.KeyCount)
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
	providerKey, secret, err := s.keyService.SelectForRequest(ctx, item.ID)
	if err != nil {
		return nil, err
	}

	result, err := probeProviderAuth(ctx, client, item, secret)
	if err != nil {
		return nil, err
	}

	if result.Success {
		_ = s.repo.UpdateHealth(ctx, item.ID, "healthy", result.Message)
		if result.AuthType != "" && result.AuthType != item.AuthType {
			if err := s.repo.UpdateAuthType(ctx, item.ID, result.AuthType); err == nil {
				item.AuthType = result.AuthType
			}
		}
		if s.audit != nil {
			targetID := item.ID
			details := map[string]any{"success": true, "auth_type": result.AuthType, "tested_url": result.TestedURL, "latency_ms": result.LatencyMs}
			if result.AutoDetected {
				details["auto_detected"] = true
			}
			s.audit.Log(ctx, "provider.test_connection", "provider", &targetID, details, ip, userAgent)
		}
		return map[string]any{"success": true, "latency_ms": result.LatencyMs, "message": result.Message, "tested_url": result.TestedURL, "auth_type": result.AuthType, "auth_auto_detected": result.AutoDetected}, nil
	}

	_ = s.repo.UpdateHealth(ctx, item.ID, "disabled", result.Message)
	if s.audit != nil {
		targetID := item.ID
		s.audit.Log(ctx, "provider.test_connection", "provider", &targetID, map[string]any{"success": false, "error": result.Message, "tested_url": result.TestedURL}, ip, userAgent)
	}
	if providerKey != nil {
		_ = s.keyService.ReportResult(ctx, providerKey.ID, false, result.Message)
	}
	return map[string]any{"success": false, "latency_ms": result.LatencyMs, "message": result.Message, "tested_url": result.TestedURL, "auth_type": result.AuthType, "auth_auto_detected": false}, nil
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

func normalizeURLWithoutV1(baseURL string) string {
	url := strings.TrimRight(baseURL, "/")
	url = strings.ReplaceAll(url, "/v1/v1", "/v1")
	url = strings.ReplaceAll(url, "/v1/", "/")
	url = strings.TrimSuffix(url, "/v1")
	return url
}
