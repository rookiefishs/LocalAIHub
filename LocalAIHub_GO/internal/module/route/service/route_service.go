package service

import (
	"context"
	"time"

	auditservice "localaihub/localaihub_go/internal/module/audit/service"
	"localaihub/localaihub_go/internal/module/route/repository"
)

type RouteService struct {
	repo  *repository.RouteRepository
	audit *auditservice.AuditService
}

func NewRouteService(repo *repository.RouteRepository, audit *auditservice.AuditService) *RouteService {
	return &RouteService{repo: repo, audit: audit}
}
func (s *RouteService) List(ctx context.Context, page, pageSize int) ([]repository.RouteState, int, error) {
	return s.repo.List(ctx, page, pageSize)
}
func (s *RouteService) Get(ctx context.Context, virtualModelID int64) (*repository.RouteState, error) {
	return s.repo.GetByVirtualModelID(ctx, virtualModelID)
}
func (s *RouteService) Switch(ctx context.Context, virtualModelID, bindingID int64, manualLock bool, lockUntil *time.Time, reason string, adminID int64, ip, userAgent string) error {
	err := s.repo.Switch(ctx, virtualModelID, bindingID, manualLock, lockUntil, reason, adminID)
	if err == nil && s.audit != nil {
		targetID := virtualModelID
		s.audit.Log(ctx, "route.switch", "route_state", &targetID, map[string]any{"binding_id": bindingID, "manual_lock": manualLock, "reason": reason}, ip, userAgent)
	}
	return err
}
func (s *RouteService) Unlock(ctx context.Context, virtualModelID int64, ip, userAgent string) error {
	err := s.repo.Unlock(ctx, virtualModelID)
	if err == nil && s.audit != nil {
		targetID := virtualModelID
		s.audit.Log(ctx, "route.unlock", "route_state", &targetID, map[string]any{}, ip, userAgent)
	}
	return err
}
func (s *RouteService) CountOpenCircuits(ctx context.Context) (int64, error) {
	return s.repo.CountOpenCircuits(ctx)
}

func (s *RouteService) RegisterFailure(ctx context.Context, providerID, virtualModelID int64, reason string) (bool, error) {
	return s.repo.RegisterFailure(ctx, providerID, virtualModelID, reason)
}

func (s *RouteService) RegisterSuccess(ctx context.Context, providerID, virtualModelID int64) error {
	return s.repo.RegisterSuccess(ctx, providerID, virtualModelID)
}

func (s *RouteService) IsCircuitOpen(ctx context.Context, providerID, virtualModelID int64) (bool, error) {
	return s.repo.IsCircuitOpen(ctx, providerID, virtualModelID)
}

func (s *RouteService) Delete(ctx context.Context, virtualModelID int64, ip, userAgent string) error {
	err := s.repo.Delete(ctx, virtualModelID)
	if err == nil && s.audit != nil {
		targetID := virtualModelID
		s.audit.Log(ctx, "route.delete", "route_state", &targetID, map[string]any{"virtual_model_id": virtualModelID}, ip, userAgent)
	}
	return err
}
