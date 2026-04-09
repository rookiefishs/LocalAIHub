package service

import (
	"context"
	"encoding/json"

	"localaihub/localaihub_go/internal/module/audit/repository"
	"localaihub/localaihub_go/internal/pkg/appctx"
	"localaihub/localaihub_go/internal/pkg/logger"
)

type AuditService struct{ repo *repository.AuditRepository }

func NewAuditService(repo *repository.AuditRepository) *AuditService {
	return &AuditService{repo: repo}
}

func (s *AuditService) Log(ctx context.Context, action, targetType string, targetID *int64, payload map[string]any, ip, userAgent string) {
	raw, err := json.Marshal(payload)
	if err != nil {
		logger.Log.Error().Err(err).Str("action", action).Str("target_type", targetType).Msg("failed to marshal audit payload")
		raw = []byte("{}")
	}

	adminUserID := appctx.AdminUserID(ctx)
	if adminUserID == 0 {
		if targetType == "admin_user" && targetID != nil && *targetID > 0 {
			adminUserID = *targetID
		} else {
			logger.Log.Warn().
				Str("action", action).
				Str("target_type", targetType).
				Msg("skip audit log because operator admin user id is missing")
			return
		}
	}

	if err := s.repo.Create(ctx, repository.AuditLog{
		AdminUserID:       adminUserID,
		Action:            action,
		TargetType:        targetType,
		TargetID:          targetID,
		ChangeSummaryJSON: raw,
		IPAddress:         ip,
		UserAgent:         userAgent,
		RequestID:         appctx.RequestID(ctx),
	}); err != nil {
		logger.Log.Error().Err(err).Str("action", action).Str("target_type", targetType).Msg("failed to create audit log")
	}
}
