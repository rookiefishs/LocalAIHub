package service

import (
	"context"
	"encoding/json"

	"localaihub/localaihub_go/internal/module/audit/repository"
	"localaihub/localaihub_go/internal/pkg/appctx"
)

type AuditService struct{ repo *repository.AuditRepository }

func NewAuditService(repo *repository.AuditRepository) *AuditService {
	return &AuditService{repo: repo}
}

func (s *AuditService) Log(ctx context.Context, action, targetType string, targetID *int64, payload map[string]any, ip, userAgent string) {
	raw, _ := json.Marshal(payload)
	_ = s.repo.Create(ctx, repository.AuditLog{
		AdminUserID:       appctx.AdminUserID(ctx),
		Action:            action,
		TargetType:        targetType,
		TargetID:          targetID,
		ChangeSummaryJSON: raw,
		IPAddress:         ip,
		UserAgent:         userAgent,
		RequestID:         appctx.RequestID(ctx),
	})
}
