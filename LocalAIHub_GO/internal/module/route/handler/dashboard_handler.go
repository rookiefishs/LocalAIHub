package handler

import (
	"net/http"
	"strconv"

	gatewayrepo "localaihub/localaihub_go/internal/module/gateway/repository"
	providerrepo "localaihub/localaihub_go/internal/module/provider/repository"
	"localaihub/localaihub_go/internal/module/route/service"
	"localaihub/localaihub_go/internal/pkg/logger"
	"localaihub/localaihub_go/internal/pkg/response"
)

type DashboardHandler struct {
	routeService *service.RouteService
	gatewayRepo  *gatewayrepo.GatewayRepository
	providerRepo *providerrepo.ProviderRepository
}

func NewDashboardHandler(routeService *service.RouteService, gatewayRepo *gatewayrepo.GatewayRepository, providerRepo *providerrepo.ProviderRepository) *DashboardHandler {
	return &DashboardHandler{routeService: routeService, gatewayRepo: gatewayRepo, providerRepo: providerRepo}
}

func (h *DashboardHandler) DashboardOverview(w http.ResponseWriter, r *http.Request) {
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
	openCircuits, _ := h.routeService.CountOpenCircuits(r.Context())
	requestCount, _ := h.gatewayRepo.CountRequests(r.Context(), hours, clientID)
	successCount, _ := h.gatewayRepo.CountSuccessRequests(r.Context(), hours, clientID)
	avgLatency, _ := h.gatewayRepo.AvgLatency(r.Context(), hours, clientID)
	activeUpstreams, _ := h.providerRepo.CountActive(r.Context())
	debugSessions, _ := h.gatewayRepo.CountDebugSessions24h(r.Context())
	promptTokens, completionTokens, totalTokens, _ := h.gatewayRepo.SumTokens(r.Context(), hours, clientID)
	trendData, err := h.gatewayRepo.GetRequestTrend(r.Context(), hours, clientID)
	if err != nil {
		logger.Log.Error().Err(err).Msg("failed to get request trend")
	}
	modelDistribution, err := h.gatewayRepo.GetModelDistribution(r.Context(), hours, clientID)
	if err != nil {
		logger.Log.Error().Err(err).Msg("failed to get model distribution")
	}
	logger.Log.Debug().Interface("trendData", trendData).Msg("request trend data")
	successRate := 0.0
	if requestCount > 0 {
		successRate = float64(successCount) / float64(requestCount)
	}

	var keyStats []map[string]any
	var keyTrendData []map[string]any
	var keyModelDist []map[string]any
	if clientID == 0 {
		stats, err := h.gatewayRepo.GetKeyStats(r.Context(), hours)
		if err != nil {
			logger.Log.Error().Err(err).Msg("failed to get key stats")
		}
		for _, s := range stats {
			keyStats = append(keyStats, map[string]any{
				"key_name":      s.KeyName,
				"request_count": s.RequestCount,
				"total_tokens":  s.TotalTokens,
				"success_rate":  s.SuccessRate,
			})
		}

		trendDataByKey, err := h.gatewayRepo.GetRequestTrendByKey(r.Context(), hours)
		if err != nil {
			logger.Log.Error().Err(err).Msg("failed to get key trend")
		} else {
			for _, t := range trendDataByKey {
				keyTrendData = append(keyTrendData, map[string]any{
					"hour":     t.Hour,
					"key_name": t.KeyName,
					"count":    t.Count,
					"tokens":   t.Tokens,
				})
			}
		}

		modelDistByKey, err := h.gatewayRepo.GetModelDistributionByKey(r.Context(), hours)
		if err != nil {
			logger.Log.Error().Err(err).Msg("failed to get key model dist")
		} else {
			for _, m := range modelDistByKey {
				keyModelDist = append(keyModelDist, map[string]any{
					"key_name":   m.KeyName,
					"model_code": m.ModelCode,
					"count":      m.Count,
				})
			}
		}
	}

	response.AdminSuccess(w, r, map[string]any{
		"request_count":          requestCount,
		"success_rate":           successRate,
		"avg_latency_ms":         avgLatency,
		"active_upstream_count":  activeUpstreams,
		"open_circuit_count":     openCircuits,
		"debug_session_count":    debugSessions,
		"prompt_tokens":          promptTokens,
		"completion_tokens":      completionTokens,
		"total_tokens":           totalTokens,
		"request_trend":          trendData,
		"model_distribution":     modelDistribution,
		"key_stats":              keyStats,
		"key_trend":              keyTrendData,
		"key_model_distribution": keyModelDist,
	})
}
