package service

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"strings"

	auditservice "localaihub/localaihub_go/internal/module/audit/service"
	"localaihub/localaihub_go/internal/module/provider/repository"
	"localaihub/localaihub_go/internal/pkg/logger"
)

type ProviderKeyService struct {
	repo          *repository.ProviderKeyRepository
	encryptionKey string
	audit         *auditservice.AuditService
}

func NewProviderKeyService(repo *repository.ProviderKeyRepository, encryptionKey string, audit *auditservice.AuditService) *ProviderKeyService {
	return &ProviderKeyService{repo: repo, encryptionKey: encryptionKey, audit: audit}
}

func (s *ProviderKeyService) ListByProviderID(ctx context.Context, providerID int64) ([]repository.ProviderKey, error) {
	return s.repo.ListByProviderID(ctx, providerID)
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
		return nil, "", nil
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
