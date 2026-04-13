package service

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	auditservice "localaihub/localaihub_go/internal/module/audit/service"
	"localaihub/localaihub_go/internal/module/provider/repository"
	"localaihub/localaihub_go/internal/pkg/logger"
)

type ProviderKeyCandidate struct {
	Key    *repository.ProviderKey
	Secret string
}

type ProviderKeyService struct {
	repo           *repository.ProviderKeyRepository
	providerRepo   *repository.ProviderRepository
	encryptionKey  string
	audit          *auditservice.AuditService
	defaultTimeout time.Duration
}
func NewProviderKeyService(repo *repository.ProviderKeyRepository, providerRepo *repository.ProviderRepository, encryptionKey string, audit *auditservice.AuditService, defaultTimeout time.Duration) *ProviderKeyService {
	return &ProviderKeyService{repo: repo, providerRepo: providerRepo, encryptionKey: encryptionKey, audit: audit, defaultTimeout: defaultTimeout}
}

func (s *ProviderKeyService) ListByProviderID(ctx context.Context, providerID int64) ([]repository.ProviderKey, error) {
	items, err := s.repo.ListByProviderID(ctx, providerID)
	if err != nil {
		return nil, err
	}
	for i := range items {
		plain, err := decodeSecret(items[i].SecretEncrypted, s.encryptionKey)
		if err == nil {
			items[i].PlainKey = plain
		}
	}
	return items, nil
}

func (s *ProviderKeyService) Create(ctx context.Context, providerID int64, secret string, priority int, remark, ip, userAgent string) (int64, error) {
	masked := maskSecret(secret)
	encoded, err := encodeSecret(secret, s.encryptionKey)
	if err != nil {
		return 0, err
	}
	var remarkPtr *string
	if remark != "" {
		remarkPtr = &remark
	}
	item := &repository.ProviderKey{ProviderID: providerID, KeyMasked: masked, SecretEncrypted: encoded, Status: "enabled", Priority: priority, Remark: remarkPtr}
	id, err := s.repo.Create(ctx, item)
	if err == nil && s.audit != nil {
		s.audit.Log(ctx, "provider_key.create", "provider_key", &id, map[string]any{"provider_id": providerID, "key_masked": masked}, ip, userAgent)
	}
	return id, err
}

func (s *ProviderKeyService) Update(ctx context.Context, id int64, secret string, priority int, remark, ip, userAgent string) error {
	masked := maskSecret(secret)
	encoded, err := encodeSecret(secret, s.encryptionKey)
	if err != nil {
		return err
	}
	err = s.repo.Update(ctx, id, masked, encoded, priority, remark)
	if err == nil && s.audit != nil {
		targetID := id
		s.audit.Log(ctx, "provider_key.update", "provider_key", &targetID, map[string]any{"id": id}, ip, userAgent)
	}
	return err
}

func (s *ProviderKeyService) UpdatePriority(ctx context.Context, id int64, priority int) error {
	return s.repo.UpdatePriority(ctx, id, priority)
}

func (s *ProviderKeyService) UpdateStatus(ctx context.Context, id int64, status, ip, userAgent string) error {
	err := s.repo.UpdateStatus(ctx, id, status)
	if err == nil && s.audit != nil {
		targetID := id
		s.audit.Log(ctx, "provider_key.update_status", "provider_key", &targetID, map[string]any{"status": status}, ip, userAgent)
	}
	return err
}

func (s *ProviderKeyService) SelectForRequest(ctx context.Context, providerID int64) (*repository.ProviderKey, string, error) {
	item, err := s.repo.FirstActiveByProviderID(ctx, providerID)
	if err != nil {
		logger.Log.Error().Err(err).Int64("provider_id", providerID).Msg("select provider key error")
		return nil, "", err
	}
	if item == nil {
		logger.Log.Debug().Int64("provider_id", providerID).Msg("no provider key found, checking DB")

		rows, err := s.repo.ListByProviderID(ctx, providerID)
		if err == nil && len(rows) > 0 {
			for _, k := range rows {
				logger.Log.Debug().Int64("provider_id", providerID).Int64("key_id", k.ID).Str("status", k.Status).Msg("available key")
			}
		}
		return nil, "", fmt.Errorf("no available provider key for provider_id %d", providerID)
	}
	logger.Log.Debug().Int64("provider_id", providerID).Int64("key_id", item.ID).Str("status", item.Status).Msg("provider key found")
	secret, err := decodeSecret(item.SecretEncrypted, s.encryptionKey)
	if err != nil {
		return nil, "", err
	}
	return item, secret, nil
}


func (s *ProviderKeyService) ReportResult(ctx context.Context, id int64, success bool, errorMessage string) error {
	return s.repo.MarkResult(ctx, id, success, errorMessage)
}

func (s *ProviderKeyService) ListCandidatesForRequest(ctx context.Context, providerID int64) ([]ProviderKeyCandidate, error) {
	items, err := s.repo.ListEnabledByProviderID(ctx, providerID)
	if err != nil {
		logger.Log.Error().Err(err).Int64("provider_id", providerID).Msg("list enabled provider keys error")
		return nil, err
	}
	candidates := make([]ProviderKeyCandidate, 0, len(items))
	for i := range items {
		secret, err := decodeSecret(items[i].SecretEncrypted, s.encryptionKey)
		if err != nil {
			logger.Log.Error().Err(err).Int64("provider_id", providerID).Int64("key_id", items[i].ID).Msg("decode provider key secret failed")
			continue
		}
		item := items[i]
		candidates = append(candidates, ProviderKeyCandidate{Key: &item, Secret: secret})
	}
	return candidates, nil
}


func (s *ProviderKeyService) Delete(ctx context.Context, id int64, ip, userAgent string) error {
	err := s.repo.Delete(ctx, id)
	if err == nil && s.audit != nil {
		targetID := id
		s.audit.Log(ctx, "provider_key.delete", "provider_key", &targetID, map[string]any{"id": id}, ip, userAgent)
	}
	return err
}

func maskSecret(secret string) string {
	if len(secret) <= 8 {
		return "****"
	}
	return secret[:4] + "****" + secret[len(secret)-4:]
}

func encodeSecret(secret, key string) (string, error) {
	block, err := aes.NewCipher(normalizeAESKey(key))
	if err != nil {
		return "", fmt.Errorf("new aes cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("new gcm: %w", err)
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("read nonce: %w", err)
	}
	cipherText := gcm.Seal(nil, nonce, []byte(secret), []byte("provider_key:v1"))
	payload := append(nonce, cipherText...)
	return "gcm:v1:" + base64.StdEncoding.EncodeToString(payload), nil
}

func decodeSecret(encoded, key string) (string, error) {
	if strings.HasPrefix(encoded, "gcm:v1:") {
		return decodeSecretGCM(strings.TrimPrefix(encoded, "gcm:v1:"), key)
	}
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", fmt.Errorf("decode provider secret: %w", err)
	}
	payload := string(decoded)
	prefix := key + ":"
	if !strings.HasPrefix(payload, prefix) {
		return "", fmt.Errorf("provider secret prefix mismatch")
	}
	return strings.TrimPrefix(payload, prefix), nil
}

func decodeSecretGCM(encoded, key string) (string, error) {
	payload, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", fmt.Errorf("decode gcm payload: %w", err)
	}
	block, err := aes.NewCipher(normalizeAESKey(key))
	if err != nil {
		return "", fmt.Errorf("new aes cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("new gcm: %w", err)
	}
	if len(payload) < gcm.NonceSize() {
		return "", fmt.Errorf("invalid gcm payload")
	}
	nonce := payload[:gcm.NonceSize()]
	cipherText := payload[gcm.NonceSize():]
	plain, err := gcm.Open(nil, nonce, cipherText, []byte("provider_key:v1"))
	if err != nil {
		return "", fmt.Errorf("gcm open: %w", err)
	}
	return string(plain), nil
}

func normalizeAESKey(key string) []byte {
	buf := make([]byte, 32)
	copy(buf, []byte(key))
	return buf
}

func (s *ProviderKeyService) TestConnection(ctx context.Context, keyID int64) (map[string]any, error) {
	logger.Log.Debug().Int64("key_id", keyID).Msg("starting provider key connection test")
	key, err := s.repo.GetByID(ctx, keyID)
	if err != nil || key == nil {
		logger.Log.Warn().Int64("key_id", keyID).Err(err).Msg("provider key not found for test")
		return nil, fmt.Errorf("key not found")
	}
	logger.Log.Debug().
		Int64("provider_id", key.ProviderID).
		Int64("key_id", key.ID).
		Str("key_masked", key.KeyMasked).
		Str("key_status", key.Status).
		Int("priority", key.Priority).
		Msg("loaded provider key for connection test")

	provider, err := s.providerRepo.GetByID(ctx, key.ProviderID)
	if err != nil || provider == nil {
		logger.Log.Warn().Int64("provider_id", key.ProviderID).Int64("key_id", key.ID).Err(err).Msg("provider not found for key test")
		return nil, fmt.Errorf("provider not found")
	}
	logger.Log.Debug().
		Int64("provider_id", provider.ID).
		Int64("key_id", key.ID).
		Str("provider_type", provider.ProviderType).
		Str("base_url", provider.BaseURL).
		Str("configured_auth_type", provider.AuthType).
		Int("timeout_ms", provider.TimeoutMS).
		Msg("loaded provider for key connection test")

	secret, err := decodeSecret(key.SecretEncrypted, s.encryptionKey)
	if err != nil {
		logger.Log.Error().Int64("provider_id", key.ProviderID).Int64("key_id", key.ID).Err(err).Msg("failed to decode provider key secret")
		return nil, fmt.Errorf("decode secret failed")
	}
	logger.Log.Debug().
		Int64("provider_id", key.ProviderID).
		Int64("key_id", key.ID).
		Str("secret_masked", maskSecret(secret)).
		Msg("decoded provider key secret for connection test")

	timeout := time.Duration(provider.TimeoutMS) * time.Millisecond
	if timeout == 0 {
		timeout = s.defaultTimeout
	}
	client := &http.Client{Timeout: timeout}
	result, err := probeProviderAuth(ctx, client, provider, secret)
	if err != nil {
		logger.Log.Error().Int64("provider_id", provider.ID).Int64("key_id", key.ID).Err(err).Msg("provider auth probe failed with error during key test")
		return nil, err
	}

	if result.Success {
		logger.Log.Info().
			Int64("provider_id", provider.ID).
			Int64("key_id", key.ID).
			Str("auth_type", result.AuthType).
			Str("tested_url", result.TestedURL).
			Int("latency_ms", result.LatencyMs).
			Bool("auth_auto_detected", result.AutoDetected).
			Msg("provider key connection test succeeded")
		if result.AuthType != "" && result.AuthType != provider.AuthType {
			if err := s.providerRepo.UpdateAuthType(ctx, provider.ID, result.AuthType); err == nil {
				logger.Log.Info().
					Int64("provider_id", provider.ID).
					Int64("key_id", key.ID).
					Str("from_auth_type", provider.AuthType).
					Str("to_auth_type", result.AuthType).
					Msg("provider auth type auto updated after key test")
				provider.AuthType = result.AuthType
			}
		}
		if err := s.repo.MarkResult(ctx, key.ID, true, ""); err != nil {
			logger.Log.Error().Err(err).Int64("key_id", key.ID).Msg("failed to mark provider key test success result")
		}
		return map[string]any{"success": true, "tested_url": result.TestedURL, "auth_type": result.AuthType, "auth_auto_detected": result.AutoDetected}, nil
	}

	logger.Log.Warn().
		Int64("provider_id", provider.ID).
		Int64("key_id", key.ID).
		Str("auth_type", result.AuthType).
		Str("tested_url", result.TestedURL).
		Int("latency_ms", result.LatencyMs).
		Int("status_code", result.StatusCode).
		Str("message", result.Message).
		Msg("provider key connection test failed")
	if err := s.repo.MarkResult(ctx, key.ID, false, result.Message); err != nil {
		logger.Log.Error().Err(err).Int64("key_id", key.ID).Msg("failed to mark provider key test failure result")
	}
	return map[string]any{"success": false, "error": result.Message, "tested_url": result.TestedURL, "auth_type": result.AuthType, "auth_auto_detected": false}, nil
}
