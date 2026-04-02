package service

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"time"

	clientkeyrepo "localaihub/localaihub_go/internal/module/clientkey/repository"
	hc "localaihub/localaihub_go/internal/module/healthcheck/repository"
	providerrepo "localaihub/localaihub_go/internal/module/provider/repository"
	"localaihub/localaihub_go/internal/pkg/logger"
)

type HealthCheckService struct {
	repo            *hc.HealthCheckRepository
	db              *sql.DB
	providerRepo    *providerrepo.ProviderRepository
	providerKeyRepo *providerrepo.ProviderKeyRepository
	clientKeyRepo   *clientkeyrepo.ClientKeyRepository
	intervalMinutes int
	enabled         bool
	audit           interface {
		Log(ctx context.Context, action, targetType string, targetID *int64, details map[string]any, ip, userAgent string)
	}
}

func NewHealthCheckService(
	repo *hc.HealthCheckRepository,
	db *sql.DB,
	providerRepo *providerrepo.ProviderRepository,
	providerKeyRepo *providerrepo.ProviderKeyRepository,
	clientKeyRepo *clientkeyrepo.ClientKeyRepository,
	audit interface {
		Log(ctx context.Context, action, targetType string, targetID *int64, details map[string]any, ip, userAgent string)
	},
) *HealthCheckService {
	return &HealthCheckService{
		repo:            repo,
		db:              db,
		providerRepo:    providerRepo,
		providerKeyRepo: providerKeyRepo,
		clientKeyRepo:   clientKeyRepo,
		audit:           audit,
	}
}

func (s *HealthCheckService) SetConfig(enabled bool, intervalMinutes int) {
	s.enabled = enabled
	s.intervalMinutes = intervalMinutes
	logger.Log.Info().Bool("enabled", enabled).Int("interval", intervalMinutes).Msg("health check config updated")
}

func (s *HealthCheckService) Start(ctx context.Context) {
	if !s.enabled {
		logger.Log.Info().Msg("health check is disabled")
		return
	}

	ticker := time.NewTicker(time.Duration(s.intervalMinutes) * time.Minute)
	go func() {
		logger.Log.Info().Msg("health check started")
		s.checkAll(ctx)
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				logger.Log.Info().Msg("health check stopped")
				return
			case <-ticker.C:
				s.checkAll(ctx)
			}
		}
	}()
}

func (s *HealthCheckService) checkAll(ctx context.Context) {
	logger.Log.Info().Msg("starting health check cycle")
	s.checkProviders(ctx)
	s.checkProviderKeys(ctx)
	s.cleanOldLogs(ctx)
	logger.Log.Info().Msg("health check cycle completed")
}

func (s *HealthCheckService) checkProviders(ctx context.Context) {
	providers, err := s.providerRepo.ListEnabled(ctx)
	if err != nil {
		logger.Log.Error().Err(err).Msg("failed to list providers for health check")
		return
	}

	for _, p := range providers {
		s.checkProvider(ctx, p)
	}
}

func (s *HealthCheckService) checkProvider(ctx context.Context, p providerrepo.Provider) {
	start := time.Now()
	err := s.testProviderConnection(ctx, &p)
	latencyMs := int(time.Since(start).Milliseconds())

	prevStatus := "enabled"
	newStatus := "enabled"
	errMsg := ""

	if err != nil {
		errMsg = err.Error()
		newStatus = "disabled"
	} else {
		newStatus = "enabled"
	}

	if prevStatus != newStatus {
		log := &hc.HealthCheckLog{
			TargetType:     "provider",
			TargetID:       p.ID,
			TargetName:     p.Name,
			CheckStatus:    newStatus,
			PreviousStatus: prevStatus,
			ErrorMessage:   errMsg,
			LatencyMs:      latencyMs,
		}
		s.repo.Create(ctx, log)
		if s.audit != nil {
			id := p.ID
			action := "health_check.provider.disabled"
			if newStatus == "enabled" {
				action = "health_check.provider.enabled"
			}
			s.audit.Log(ctx, action, "provider", &id, map[string]any{"name": p.Name, "latency_ms": latencyMs}, "", "")
		}
		logger.Log.Info().Int64("provider_id", p.ID).Str("prev", prevStatus).Str("new", newStatus).Msg("provider health check status changed")
	}
}

func (s *HealthCheckService) checkProviderKeys(ctx context.Context) {
	keys, err := s.providerKeyRepo.ListEnabled(ctx)
	if err != nil {
		logger.Log.Error().Err(err).Msg("failed to list provider keys for health check")
		return
	}

	for _, k := range keys {
		s.checkProviderKey(ctx, k)
	}
}

func (s *HealthCheckService) checkProviderKey(ctx context.Context, k providerrepo.ProviderKey) {
	provider, err := s.providerRepo.GetByID(ctx, k.ProviderID)
	if err != nil || provider == nil {
		return
	}

	start := time.Now()
	err = s.testProviderConnection(ctx, provider)
	latencyMs := int(time.Since(start).Milliseconds())

	prevStatus := "enabled"
	newStatus := "enabled"
	errMsg := ""

	if err != nil {
		errMsg = err.Error()
		newStatus = "disabled"
		s.providerKeyRepo.UpdateStatus(ctx, k.ID, "disabled")
	} else {
		status, _ := s.providerKeyRepo.GetStatus(ctx, k.ID)
		if status != "enabled" {
			newStatus = "enabled"
			s.providerKeyRepo.UpdateStatus(ctx, k.ID, "enabled")
		}
	}

	if prevStatus != newStatus {
		log := &hc.HealthCheckLog{
			TargetType:     "provider_key",
			TargetID:       k.ID,
			TargetName:     k.KeyMasked,
			CheckStatus:    newStatus,
			PreviousStatus: prevStatus,
			ErrorMessage:   errMsg,
			LatencyMs:      latencyMs,
		}
		s.repo.Create(ctx, log)
		if s.audit != nil {
			id := k.ID
			action := "health_check.provider_key.disabled"
			if newStatus == "enabled" {
				action = "health_check.provider_key.enabled"
			}
			s.audit.Log(ctx, action, "provider_key", &id, map[string]any{"provider_id": k.ProviderID, "latency_ms": latencyMs}, "", "")
		}
		logger.Log.Info().Int64("provider_key_id", k.ID).Str("prev", prevStatus).Str("new", newStatus).Msg("provider key health check status changed")
	}
}

func (s *HealthCheckService) checkClientKeys(ctx context.Context) {
	keys, err := s.clientKeyRepo.ListActive(ctx)
	if err != nil {
		logger.Log.Error().Err(err).Msg("failed to list client keys for health check")
		return
	}

	for _, k := range keys {
		s.checkClientKey(ctx, k)
	}
}

func (s *HealthCheckService) checkClientKey(ctx context.Context, k clientkeyrepo.ClientKey) {
	start := time.Now()
	err := s.testClientKeyConnection(ctx, k)
	latencyMs := int(time.Since(start).Milliseconds())

	prevStatus := "active"
	newStatus := "active"
	errMsg := ""

	if err != nil {
		errMsg = err.Error()
		newStatus = "disabled"
		s.clientKeyRepo.UpdateStatus(ctx, k.ID, "disabled")
	} else {
		status, _ := s.clientKeyRepo.GetStatus(ctx, k.ID)
		if status != "active" {
			newStatus = "active"
			s.clientKeyRepo.UpdateStatus(ctx, k.ID, "active")
		}
	}

	if prevStatus != newStatus {
		log := &hc.HealthCheckLog{
			TargetType:     "client_key",
			TargetID:       k.ID,
			TargetName:     k.Name,
			CheckStatus:    newStatus,
			PreviousStatus: prevStatus,
			ErrorMessage:   errMsg,
			LatencyMs:      latencyMs,
		}
		s.repo.Create(ctx, log)
		if s.audit != nil {
			id := k.ID
			action := "health_check.client_key.disabled"
			if newStatus == "active" {
				action = "health_check.client_key.enabled"
			}
			s.audit.Log(ctx, action, "api_client", &id, map[string]any{"name": k.Name, "latency_ms": latencyMs}, "", "")
		}
		logger.Log.Info().Int64("client_key_id", k.ID).Str("prev", prevStatus).Str("new", newStatus).Msg("client key health check status changed")
	}
}

func (s *HealthCheckService) cleanOldLogs(ctx context.Context) {
	retentionDays := 30
	err := s.repo.DeleteOlderThan(ctx, retentionDays)
	if err != nil {
		logger.Log.Error().Err(err).Int("days", retentionDays).Msg("failed to clean old health check logs")
	}
}

func (s *HealthCheckService) testProviderConnection(ctx context.Context, p *providerrepo.Provider) error {
	keys, err := s.providerRepo.ListKeys(ctx, p.ID)
	if err != nil {
		return err
	}

	if len(keys) == 0 {
		return fmt.Errorf("no provider key")
	}

	key := keys[0]
	client := &http.Client{Timeout: time.Duration(p.TimeoutMS) * time.Millisecond}
	secret := key.SecretEncrypted

	resp, err := s.doHealthCheckRequest(ctx, client, p, secret, true)
	if err == nil && resp != nil && resp.StatusCode < 400 {
		resp.Body.Close()
		return nil
	}
	if resp != nil {
		resp.Body.Close()
	}

	resp, err = s.doHealthCheckRequest(ctx, client, p, secret, false)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("upstream error: %d", resp.StatusCode)
	}
	return nil
}

func (s *HealthCheckService) doHealthCheckRequest(ctx context.Context, client *http.Client, p *providerrepo.Provider, secret string, withV1 bool) (*http.Response, error) {
	baseURL := strings.TrimRight(p.BaseURL, "/")
	if !withV1 {
		baseURL = strings.TrimSuffix(baseURL, "/v1")
	}
	testURL := baseURL
	if p.ProviderType == "anthropic" {
		if withV1 {
			testURL += "/v1/messages"
		} else {
			testURL += "/messages"
		}
	} else {
		if withV1 {
			testURL += "/v1/chat/completions"
		} else {
			testURL += "/chat/completions"
		}
	}
	req, _ := http.NewRequestWithContext(ctx, "POST", testURL, strings.NewReader(`{"model":"gpt-4o-mini","max_tokens":1,"messages":[{"role":"user","content":"hi"}]}`))
	req.Header.Set("Content-Type", "application/json")
	if p.AuthType == "x_api_key" {
		req.Header.Set("x-api-key", secret)
	} else {
		req.Header.Set("Authorization", "Bearer "+secret)
	}
	return client.Do(req)
}

func (s *HealthCheckService) testClientKeyConnection(ctx context.Context, k clientkeyrepo.ClientKey) error {
	if k.ExpiresAt != nil && k.ExpiresAt.Before(time.Now().UTC()) {
		return fmt.Errorf("key expired")
	}
	return nil
}
