package handler

import (
	"net/http"
	"strconv"
	"time"

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
	requestCount, err := h.gatewayRepo.CountRequests(r.Context(), hours, clientID)
	if err != nil {
		logger.Log.Error().Err(err).Msg("failed to count requests")
	}
	successCount, err := h.gatewayRepo.CountSuccessRequests(r.Context(), hours, clientID)
	if err != nil {
		logger.Log.Error().Err(err).Msg("failed to count success requests")
	}
	avgLatency, err := h.gatewayRepo.AvgLatency(r.Context(), hours, clientID)
	if err != nil {
		logger.Log.Error().Err(err).Msg("failed to calculate average latency")
	}
	activeUpstreams, err := h.providerRepo.CountActive(r.Context())
	if err != nil {
		logger.Log.Error().Err(err).Msg("failed to count active upstreams")
	}
	debugSessions, err := h.gatewayRepo.CountDebugSessions24h(r.Context())
	if err != nil {
		logger.Log.Error().Err(err).Msg("failed to count debug sessions")
	}
	promptTokens, completionTokens, totalTokens, err := h.gatewayRepo.SumTokens(r.Context(), hours, clientID)
	if err != nil {
		logger.Log.Error().Err(err).Msg("failed to sum tokens")
	}
	trendData, err := h.gatewayRepo.GetRequestTrend(r.Context(), hours, clientID)
	if err != nil {
		logger.Log.Error().Err(err).Msg("failed to get request trend")
	}
	trendData = fillHourlyTrendData(trendData, hours)
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
			trendDataByKey = fillHourlyKeyTrendData(trendDataByKey, hours)
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
		"success_count":          successCount,
		"failure_count":          maxInt64(requestCount-successCount, 0),
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

func fillHourlyTrendData(items []gatewayrepo.HourlyStat, hours int) []gatewayrepo.HourlyStat {
	if hours <= 0 {
		return items
	}
	byHour := make(map[string]gatewayrepo.HourlyStat, len(items))
	for _, item := range items {
		byHour[item.Hour] = item
	}
	loc := time.FixedZone("CST", 8*3600)
	now := time.Now().UTC().In(loc).Truncate(time.Hour)
	start := now.Add(time.Duration(-(hours - 1)) * time.Hour)
	filled := make([]gatewayrepo.HourlyStat, 0, hours)
	for current := start; !current.After(now); current = current.Add(time.Hour) {
		key := current.Format("2006-01-02 15:00")
		if item, ok := byHour[key]; ok {
			filled = append(filled, item)
			continue
		}
		filled = append(filled, gatewayrepo.HourlyStat{Hour: key})
	}
	return filled
}

func fillHourlyKeyTrendData(items []gatewayrepo.KeyTrend, hours int) []gatewayrepo.KeyTrend {
	if hours <= 0 || len(items) == 0 {
		return items
	}
	keyNames := make([]string, 0)
	seenKeys := make(map[string]struct{})
	byKeyHour := make(map[string]gatewayrepo.KeyTrend, len(items))
	for _, item := range items {
		if _, ok := seenKeys[item.KeyName]; !ok {
			seenKeys[item.KeyName] = struct{}{}
			keyNames = append(keyNames, item.KeyName)
		}
		byKeyHour[item.KeyName+"|"+item.Hour] = item
	}
	loc := time.FixedZone("CST", 8*3600)
	now := time.Now().UTC().In(loc).Truncate(time.Hour)
	start := now.Add(time.Duration(-(hours - 1)) * time.Hour)
	filled := make([]gatewayrepo.KeyTrend, 0, hours*len(keyNames))
	for current := start; !current.After(now); current = current.Add(time.Hour) {
		hourKey := current.Format("2006-01-02 15:00")
		for _, keyName := range keyNames {
			lookupKey := keyName + "|" + hourKey
			if item, ok := byKeyHour[lookupKey]; ok {
				filled = append(filled, item)
				continue
			}
			filled = append(filled, gatewayrepo.KeyTrend{Hour: hourKey, KeyName: keyName})
		}
	}
	return filled
}

func maxInt64(value int64, floor int64) int64 {
	if value < floor {
		return floor
	}
	return value
}
