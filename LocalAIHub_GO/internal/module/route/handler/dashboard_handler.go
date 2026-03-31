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
	openCircuits, _ := h.routeService.CountOpenCircuits(r.Context())
	requestCount, _ := h.gatewayRepo.CountRequests(r.Context(), hours)
	successCount, _ := h.gatewayRepo.CountSuccessRequests(r.Context(), hours)
	avgLatency, _ := h.gatewayRepo.AvgLatency(r.Context(), hours)
	activeUpstreams, _ := h.providerRepo.CountActive(r.Context())
	debugSessions, _ := h.gatewayRepo.CountDebugSessions24h(r.Context())
	promptTokens, completionTokens, totalTokens, _ := h.gatewayRepo.SumTokens(r.Context(), hours)
	trendData, err := h.gatewayRepo.GetRequestTrend(r.Context(), hours)
	if err != nil {
		logger.Log.Error().Err(err).Msg("failed to get request trend")
	}
	modelDistribution, err := h.gatewayRepo.GetModelDistribution(r.Context(), hours)
	if err != nil {
		logger.Log.Error().Err(err).Msg("failed to get model distribution")
	}
	logger.Log.Debug().Interface("trendData", trendData).Msg("request trend data")
	successRate := 0.0
	if requestCount > 0 {
		successRate = float64(successCount) / float64(requestCount)
	}
	response.AdminSuccess(w, r, map[string]any{
		"request_count":         requestCount,
		"success_rate":          successRate,
		"avg_latency_ms":        avgLatency,
		"active_upstream_count": activeUpstreams,
		"open_circuit_count":    openCircuits,
		"debug_session_count":   debugSessions,
		"prompt_tokens":         promptTokens,
		"completion_tokens":     completionTokens,
		"total_tokens":          totalTokens,
		"request_trend":         trendData,
		"model_distribution":    modelDistribution,
	})
}
