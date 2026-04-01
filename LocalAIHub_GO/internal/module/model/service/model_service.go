package service

import (
	"context"
	"fmt"

	auditservice "localaihub/localaihub_go/internal/module/audit/service"
	"localaihub/localaihub_go/internal/module/model/repository"
	providerrepo "localaihub/localaihub_go/internal/module/provider/repository"
	providerservice "localaihub/localaihub_go/internal/module/provider/service"
)

type ModelService struct {
	repo            *repository.ModelRepository
	audit           *auditservice.AuditService
	providerService *providerservice.ProviderService
}

func NewModelService(repo *repository.ModelRepository, audit *auditservice.AuditService, providerService *providerservice.ProviderService) *ModelService {
	return &ModelService{repo: repo, audit: audit, providerService: providerService}
}
func (s *ModelService) List(ctx context.Context, page, pageSize int) ([]repository.Model, int, error) {
	return s.repo.List(ctx, page, pageSize)
}
func (s *ModelService) Get(ctx context.Context, id int64) (*repository.Model, error) {
	return s.repo.Get(ctx, id)
}
func (s *ModelService) Create(ctx context.Context, item *repository.Model, ip, userAgent string) (int64, error) {
	id, err := s.repo.Create(ctx, item)
	if err == nil && s.audit != nil {
		s.audit.Log(ctx, "model.create", "virtual_model", &id, map[string]any{"model_code": item.ModelCode}, ip, userAgent)
	}
	return id, err
}
func (s *ModelService) Update(ctx context.Context, item *repository.Model, ip, userAgent string) error {
	err := s.repo.Update(ctx, item)
	if err == nil && s.audit != nil {
		targetID := item.ID
		s.audit.Log(ctx, "model.update", "virtual_model", &targetID, map[string]any{"model_code": item.ModelCode, "visible": item.Visible}, ip, userAgent)
	}
	return err
}
func (s *ModelService) ListBindings(ctx context.Context, modelID int64) ([]repository.Binding, error) {
	return s.repo.ListBindings(ctx, modelID)
}
func (s *ModelService) CreateBinding(ctx context.Context, item *repository.Binding, ip, userAgent string) (int64, error) {
	id, err := s.repo.CreateBinding(ctx, item)
	if err == nil && s.audit != nil {
		s.audit.Log(ctx, "binding.create", "virtual_model_binding", &id, map[string]any{"virtual_model_id": item.VirtualModelID, "provider_id": item.ProviderID}, ip, userAgent)
	}
	return id, err
}

func (s *ModelService) UpdateBinding(ctx context.Context, item *repository.Binding) error {
	err := s.repo.UpdateBinding(ctx, item)
	if err == nil && s.audit != nil {
		targetID := item.ID
		s.audit.Log(ctx, "binding.update", "virtual_model_binding", &targetID, map[string]any{"priority": item.Priority}, "", "")
	}
	return err
}

func (s *ModelService) Delete(ctx context.Context, id int64, ip, userAgent string) error {
	check, err := s.repo.CheckDelete(ctx, id)
	if err != nil {
		return err
	}
	if check.BindingCount > 0 {
		return fmt.Errorf("model is still referenced by %d bindings", check.BindingCount)
	}
	err = s.repo.Delete(ctx, id)
	if err == nil && s.audit != nil {
		targetID := id
		s.audit.Log(ctx, "model.delete", "virtual_model", &targetID, map[string]any{"id": id}, ip, userAgent)
	}
	return err
}

func (s *ModelService) DeleteBinding(ctx context.Context, id int64, ip, userAgent string) error {
	err := s.repo.DeleteBinding(ctx, id)
	if err == nil && s.audit != nil {
		targetID := id
		s.audit.Log(ctx, "binding.delete", "virtual_model_binding", &targetID, map[string]any{"id": id}, ip, userAgent)
	}
	return err
}

func (s *ModelService) TestBinding(ctx context.Context, modelID, bindingID int64) (map[string]any, error) {
	binding, err := s.repo.GetBinding(ctx, bindingID)
	if err != nil {
		return nil, err
	}
	if binding == nil {
		return nil, fmt.Errorf("binding not found")
	}
	if binding.VirtualModelID != modelID {
		return nil, fmt.Errorf("binding does not belong to model")
	}

	provider, err := s.providerService.Get(ctx, binding.ProviderID)
	if err != nil || provider == nil {
		return nil, fmt.Errorf("provider not found")
	}

	testReq := &providerrepo.Provider{
		ID:           provider.ID,
		Name:         provider.Name,
		BaseURL:      provider.BaseURL,
		AuthType:     provider.AuthType,
		ProviderType: provider.ProviderType,
		TimeoutMS:    provider.TimeoutMS,
	}
	return s.providerService.TestConnection(ctx, testReq, "", "")
}
