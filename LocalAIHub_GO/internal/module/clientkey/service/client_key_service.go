package service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	auditservice "localaihub/localaihub_go/internal/module/audit/service"
	"localaihub/localaihub_go/internal/module/clientkey/repository"
	gatewayrepo "localaihub/localaihub_go/internal/module/gateway/repository"
	providerservice "localaihub/localaihub_go/internal/module/provider/service"
	"localaihub/localaihub_go/internal/pkg/random"
)

type CreatedClientKey struct {
	Key      *repository.ClientKey `json:"key"`
	PlainKey string                `json:"plain_key"`
}

type ClientKeyService struct {
	repo               *repository.ClientKeyRepository
	gatewayRepo        *gatewayrepo.GatewayRepository
	providerKeyService *providerservice.ProviderKeyService
	audit              *auditservice.AuditService
}

func NewClientKeyService(repo *repository.ClientKeyRepository, gatewayRepo *gatewayrepo.GatewayRepository, providerKeyService *providerservice.ProviderKeyService, audit *auditservice.AuditService) *ClientKeyService {
	return &ClientKeyService{repo: repo, gatewayRepo: gatewayRepo, providerKeyService: providerKeyService, audit: audit}
}

func (s *ClientKeyService) List(ctx context.Context, page, pageSize int) ([]repository.ClientKey, int, error) {
	return s.repo.List(ctx, page, pageSize)
}
func (s *ClientKeyService) Get(ctx context.Context, id int64) (*repository.ClientKey, error) {
	return s.repo.Get(ctx, id)
}

func (s *ClientKeyService) GetAllowedModels(ctx context.Context, clientID int64) ([]int64, error) {
	return s.repo.GetAllowedModels(ctx, clientID)
}

func (s *ClientKeyService) Create(ctx context.Context, name, remark, expiresAt string, allowedModels []int64, ip, userAgent string) (*CreatedClientKey, error) {
	plainSuffix, err := random.Hex(16)
	if err != nil {
		return nil, err
	}
	plainKey := "ak_live_" + plainSuffix
	prefix := plainKey[:12]
	hash := sha256.Sum256([]byte(plainKey))
	hashString := hex.EncodeToString(hash[:])
	var remarkPtr *string
	if remark != "" {
		remarkPtr = &remark
	}
	var expiresAtPtr *time.Time
	if expiresAt != "" {
		if expiresAt == "permanent" {
			expiresAtPtr = nil
		} else {
			days := 0
			if _, err := fmt.Sscanf(expiresAt, "%d", &days); err == nil && days > 0 {
				t := time.Now().UTC().AddDate(0, 0, days)
				expiresAtPtr = &t
			} else {
				t, err := time.Parse("2006-01-02", expiresAt)
				if err == nil {
					expiresAtPtr = &t
				}
			}
		}
	}
	item := &repository.ClientKey{Name: name, KeyPrefix: prefix, APIKeyHash: hashString, PlainKey: plainKey, Status: "active", Remark: remarkPtr, ExpiresAt: expiresAtPtr}
	id, err := s.repo.Create(ctx, item)
	if err != nil {
		return nil, err
	}
	created, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if len(allowedModels) > 0 {
		if err := s.repo.AssignModels(ctx, id, allowedModels); err != nil {
			return nil, err
		}
		allowed, _ := s.repo.GetAllowedModels(ctx, id)
		created.AllowedModels = allowed
	}
	if s.audit != nil {
		s.audit.Log(ctx, "client_key.create", "api_client", &id, map[string]any{"name": name, "key_prefix": created.KeyPrefix}, ip, userAgent)
	}
	if _, testErr := s.Test(ctx, id); testErr != nil {
		_ = s.repo.UpdateStatus(ctx, id, "disabled")
		created.Status = "disabled"
	}
	return &CreatedClientKey{Key: created, PlainKey: plainKey}, nil
}

func (s *ClientKeyService) UpdateStatus(ctx context.Context, id int64, status, ip, userAgent string) error {
	if status == "active" {
		if _, err := s.Test(ctx, id); err != nil {
			_ = s.repo.UpdateStatus(ctx, id, "disabled")
			return err
		}
	}
	err := s.repo.UpdateStatus(ctx, id, status)
	if err == nil && s.audit != nil {
		targetID := id
		s.audit.Log(ctx, "client_key.update_status", "api_client", &targetID, map[string]any{"status": status}, ip, userAgent)
	}
	return err
}

func (s *ClientKeyService) Update(ctx context.Context, id int64, name, remark, ip, userAgent string, allowedModels []int64) error {
	err := s.repo.Update(ctx, id, name, remark)
	if err != nil {
		return err
	}
	if err := s.repo.ReplaceAllowedModels(ctx, id, allowedModels); err != nil {
		return err
	}
	item, getErr := s.repo.Get(ctx, id)
	if getErr == nil && item != nil && item.Status == "active" {
		if _, testErr := s.Test(ctx, id); testErr != nil {
			_ = s.repo.UpdateStatus(ctx, id, "disabled")
			return testErr
		}
	}
	if err == nil && s.audit != nil {
		targetID := id
		s.audit.Log(ctx, "client_key.update", "api_client", &targetID, map[string]any{"name": name}, ip, userAgent)
	}
	return err
}

func (s *ClientKeyService) Test(ctx context.Context, id int64) (map[string]any, error) {
	item, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, fmt.Errorf("client key not found")
	}
	allowedModels, err := s.repo.GetAllowedModels(ctx, id)
	if err != nil {
		return nil, err
	}
	if len(allowedModels) == 0 {
		visibleModels, listErr := s.gatewayRepo.ListVisibleModels(ctx)
		if listErr != nil {
			return nil, listErr
		}
		if len(visibleModels) == 0 {
			return nil, fmt.Errorf("no available model configured")
		}
		if firstID, ok := visibleModels[0]["id"].(int64); ok {
			allowedModels = []int64{firstID}
		} else {
			return nil, fmt.Errorf("no testable model found")
		}
	}
	modelID := allowedModels[0]
	modelCode, route, err := s.resolveTestRoute(ctx, modelID)
	if err != nil {
		return nil, err
	}
	providerKey, secret, err := s.providerKeyService.SelectForRequest(ctx, route.ProviderID)
	if err != nil || providerKey == nil || secret == "" {
		return nil, fmt.Errorf("no available provider key")
	}
	payload := map[string]any{
		"model":      route.UpstreamModelName,
		"messages":   []map[string]string{{"role": "user", "content": "hi"}},
		"max_tokens": 5,
		"stream":     false,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	requestURL := strings.TrimRight(route.BaseURL, "/")
	requestURL = strings.ReplaceAll(requestURL, "/v1/v1", "/v1")
	requestURL = strings.ReplaceAll(requestURL, "/v1/", "/")
	if !strings.HasSuffix(requestURL, "/v1") {
		requestURL += "/v1"
	}
	requestURL += "/chat/completions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, requestURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if route.AuthType == "x_api_key" {
		req.Header.Set("x-api-key", secret)
	} else {
		req.Header.Set("Authorization", "Bearer "+secret)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("upstream error: %s", strings.TrimSpace(string(respBody)))
	}
	return map[string]any{"ok": true, "model": modelCode, "url": requestURL}, nil
}

func (s *ClientKeyService) resolveTestRoute(ctx context.Context, modelID int64) (string, *gatewayrepo.ModelRoute, error) {
	code, err := s.gatewayRepo.GetVisibleModelCodeByID(ctx, modelID)
	if err != nil {
		return "", nil, err
	}
	if code == "" {
		return "", nil, fmt.Errorf("allowed model not found")
	}
	route, routeErr := s.gatewayRepo.ResolveOpenAIModelRoute(ctx, code)
	if routeErr != nil {
		return "", nil, routeErr
	}
	if route == nil {
		return "", nil, fmt.Errorf("virtual model route not found")
	}
	return code, route, nil
}

func (s *ClientKeyService) Delete(ctx context.Context, id int64, ip, userAgent string) error {
	err := s.repo.Delete(ctx, id)
	if err == nil && s.audit != nil {
		targetID := id
		s.audit.Log(ctx, "client_key.delete", "api_client", &targetID, map[string]any{"id": id}, ip, userAgent)
	}
	return err
}

func (s *ClientKeyService) GetQuotaUsage(ctx context.Context, id int64) (*repository.ClientKey, error) {
	return s.repo.GetQuotaUsage(ctx, id)
}

func (s *ClientKeyService) UpdateQuota(ctx context.Context, id int64, dailyReq, monthlyReq, dailyToken, monthlyToken *int64) error {
	err := s.repo.UpdateQuota(ctx, id, dailyReq, monthlyReq, dailyToken, monthlyToken)
	if err == nil && s.audit != nil {
		targetID := id
		s.audit.Log(ctx, "client_key.update_quota", "api_client", &targetID, map[string]any{
			"daily_request_limit":   dailyReq,
			"monthly_request_limit": monthlyReq,
			"daily_token_limit":     dailyToken,
			"monthly_token_limit":   monthlyToken,
		}, "", "")
	}
	return err
}
