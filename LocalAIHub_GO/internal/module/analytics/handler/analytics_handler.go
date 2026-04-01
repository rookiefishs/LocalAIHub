package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"localaihub/localaihub_go/internal/module/analytics/repository"
	"localaihub/localaihub_go/internal/pkg/logger"
	"localaihub/localaihub_go/internal/pkg/response"
)

type AnalyticsHandler struct {
	repo *repository.AnalyticsRepository
}

func NewAnalyticsHandler(repo *repository.AnalyticsRepository) *AnalyticsHandler {
	return &AnalyticsHandler{repo: repo}
}

func (h *AnalyticsHandler) GetCostStats(w http.ResponseWriter, r *http.Request) {
	hours := 24
	if v := r.URL.Query().Get("hours"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			hours = parsed
		}
	}
	providerID := int64(0)
	if v := r.URL.Query().Get("provider_id"); v != "" {
		if parsed, err := strconv.ParseInt(v, 10, 64); err == nil && parsed > 0 {
			providerID = parsed
		}
	}
	clientID := int64(0)
	if v := r.URL.Query().Get("client_id"); v != "" {
		if parsed, err := strconv.ParseInt(v, 10, 64); err == nil && parsed > 0 {
			clientID = parsed
		}
	}

	totalCost, _ := h.repo.GetCostStats(r.Context(), hours, providerID, clientID)
	byProvider, _ := h.repo.GetCostByProvider(r.Context(), hours, clientID)
	byModel, _ := h.repo.GetCostByModel(r.Context(), hours, providerID, clientID)
	trend, _ := h.repo.GetCostTrend(r.Context(), hours, providerID, clientID)

	response.AdminSuccess(w, r, map[string]any{
		"total_cost":  totalCost["total_cost"],
		"by_provider": byProvider,
		"by_model":    byModel,
		"trend":       trend,
	})
}

func (h *AnalyticsHandler) GetTokenStats(w http.ResponseWriter, r *http.Request) {
	hours := 24
	if v := r.URL.Query().Get("hours"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			hours = parsed
		}
	}
	providerID := int64(0)
	if v := r.URL.Query().Get("provider_id"); v != "" {
		if parsed, err := strconv.ParseInt(v, 10, 64); err == nil && parsed > 0 {
			providerID = parsed
		}
	}
	clientID := int64(0)
	if v := r.URL.Query().Get("client_id"); v != "" {
		if parsed, err := strconv.ParseInt(v, 10, 64); err == nil && parsed > 0 {
			clientID = parsed
		}
	}

	tokens, _ := h.repo.GetTokenStats(r.Context(), hours, providerID, clientID)
	trend, _ := h.repo.GetTokenTrend(r.Context(), hours, providerID, clientID)

	response.AdminSuccess(w, r, map[string]any{
		"total_prompt_tokens":     tokens["prompt_tokens"],
		"total_completion_tokens": tokens["completion_tokens"],
		"total_tokens":            tokens["total_tokens"],
		"trend":                   trend,
	})
}

func (h *AnalyticsHandler) GetComparison(w http.ResponseWriter, r *http.Request) {
	hours := 24
	if v := r.URL.Query().Get("hours"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			hours = parsed
		}
	}
	clientID := int64(0)
	if v := r.URL.Query().Get("client_id"); v != "" {
		if parsed, err := strconv.ParseInt(v, 10, 64); err == nil && parsed > 0 {
			clientID = parsed
		}
	}

	requestComp, _ := h.repo.GetRequestComparison(r.Context(), hours, clientID)
	tokenComp, _ := h.repo.GetTokenComparison(r.Context(), hours, clientID)
	costComp, _ := h.repo.GetCostComparison(r.Context(), hours, clientID)

	response.AdminSuccess(w, r, map[string]any{
		"request_count": requestComp,
		"total_tokens":  tokenComp,
		"total_cost":    costComp,
	})
}

func (h *AnalyticsHandler) ListModelPricing(w http.ResponseWriter, r *http.Request) {
	page := 1
	pageSize := 20
	if v := r.URL.Query().Get("page"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			page = parsed
		}
	}
	if v := r.URL.Query().Get("page_size"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			pageSize = parsed
		}
	}

	items, total, err := h.repo.ListModelPricing(r.Context(), page, pageSize)
	if err != nil {
		logger.Log.Error().Err(err).Msg("failed to list model pricing")
		response.AdminError(w, r, http.StatusInternalServerError, 500100, "list model pricing failed")
		return
	}

	response.AdminSuccess(w, r, map[string]any{
		"items": items,
		"total": total,
		"page":  page,
	})
}

func (h *AnalyticsHandler) CreateModelPricing(w http.ResponseWriter, r *http.Request) {
	var item repository.ModelPricing
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		response.AdminError(w, r, http.StatusBadRequest, 400100, "invalid parameters")
		return
	}

	id, err := h.repo.CreateModelPricing(r.Context(), &item)
	if err != nil {
		logger.Log.Error().Err(err).Msg("failed to create model pricing")
		response.AdminError(w, r, http.StatusInternalServerError, 500100, "create model pricing failed")
		return
	}

	response.AdminSuccess(w, r, map[string]any{"id": id})
}

func (h *AnalyticsHandler) UpdateModelPricing(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		response.AdminError(w, r, http.StatusBadRequest, 400100, "invalid parameters")
		return
	}

	existing, err := h.repo.GetModelPricing(r.Context(), id)
	if err != nil {
		response.AdminError(w, r, http.StatusNotFound, 404100, "model pricing not found")
		return
	}

	var item repository.ModelPricing
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		response.AdminError(w, r, http.StatusBadRequest, 400100, "invalid parameters")
		return
	}
	item.ID = existing.ID
	item.CreatedAt = existing.CreatedAt

	if err := h.repo.UpdateModelPricing(r.Context(), &item); err != nil {
		logger.Log.Error().Err(err).Msg("failed to update model pricing")
		response.AdminError(w, r, http.StatusInternalServerError, 500100, "update model pricing failed")
		return
	}

	response.AdminSuccess(w, r, nil)
}

func (h *AnalyticsHandler) DeleteModelPricing(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		response.AdminError(w, r, http.StatusBadRequest, 400100, "invalid parameters")
		return
	}

	if err := h.repo.DeleteModelPricing(r.Context(), id); err != nil {
		logger.Log.Error().Err(err).Msg("failed to delete model pricing")
		response.AdminError(w, r, http.StatusInternalServerError, 500100, "delete model pricing failed")
		return
	}

	response.AdminSuccess(w, r, nil)
}
