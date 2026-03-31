package handler

import (
	"net/http"
	"strconv"
	"time"

	gatewayrepo "localaihub/localaihub_go/internal/module/gateway/repository"
	logrepo "localaihub/localaihub_go/internal/module/log/repository"
	logservice "localaihub/localaihub_go/internal/module/log/service"
	"localaihub/localaihub_go/internal/pkg/response"
)

type LogHandler struct{ service *logservice.LogService }

func NewLogHandler(service *logservice.LogService) *LogHandler { return &LogHandler{service: service} }

func (h *LogHandler) ListRequestLogs(w http.ResponseWriter, r *http.Request) {
	limit := parseLimit(r, 10)
	page := parsePage(r, 1)
	filters := gatewayrepo.RequestLogFilters{Limit: limit, Page: page}
	if traceID := r.URL.Query().Get("trace_id"); traceID != "" {
		filters.TraceID = traceID
	}
	if clientID := parseInt64Query(r, "client_id"); clientID != nil {
		filters.ClientID = clientID
	}
	if providerID := parseInt64Query(r, "provider_id"); providerID != nil {
		filters.ProviderID = providerID
	}
	if virtualModel := r.URL.Query().Get("virtual_model_code"); virtualModel != "" {
		filters.VirtualModelCode = virtualModel
	}
	if success := parseBoolQuery(r, "success"); success != nil {
		filters.Success = success
	}
	if start := parseTimeQuery(r, "start_time"); start != nil {
		filters.StartTime = start
	}
	if end := parseTimeQuery(r, "end_time"); end != nil {
		filters.EndTime = end
	}
	items, total, err := h.service.ListRequestLogs(r.Context(), filters)
	if err != nil {
		response.AdminError(w, r, http.StatusInternalServerError, 500100, "list request logs failed")
		return
	}
	response.AdminSuccess(w, r, map[string]any{"items": items, "total": total, "page": page, "page_size": limit})
}

func (h *LogHandler) ListAuditLogs(w http.ResponseWriter, r *http.Request) {
	limit := parseLimit(r, 10)
	page := parsePage(r, 1)
	filters := logrepo.AuditLogFilters{Limit: limit, Page: page}
	if adminUserID := parseInt64Query(r, "admin_user_id"); adminUserID != nil {
		filters.AdminUserID = adminUserID
	}
	if action := r.URL.Query().Get("action"); action != "" {
		filters.Action = action
	}
	if targetType := r.URL.Query().Get("target_type"); targetType != "" {
		filters.TargetType = targetType
	}
	if targetID := parseInt64Query(r, "target_id"); targetID != nil {
		filters.TargetID = targetID
	}
	if start := parseTimeQuery(r, "start_time"); start != nil {
		filters.StartTime = start
	}
	if end := parseTimeQuery(r, "end_time"); end != nil {
		filters.EndTime = end
	}
	items, total, err := h.service.ListAuditLogs(r.Context(), filters)
	if err != nil {
		response.AdminError(w, r, http.StatusInternalServerError, 500100, "list audit logs failed")
		return
	}
	response.AdminSuccess(w, r, map[string]any{"items": items, "total": total, "page": page, "page_size": limit})
}

func (h *LogHandler) GetRequestLog(w http.ResponseWriter, r *http.Request, id int64) {
	item, err := h.service.GetRequestLog(r.Context(), id)
	if err != nil {
		response.AdminError(w, r, http.StatusInternalServerError, 500100, "get request log failed")
		return
	}
	if item == nil {
		response.AdminError(w, r, http.StatusNotFound, 404100, "request log not found")
		return
	}
	response.AdminSuccess(w, r, item)
}

func parseLimit(r *http.Request, fallback int) int {
	value := r.URL.Query().Get("page_size")
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 || parsed > 200 {
		return fallback
	}
	return parsed
}

func parsePage(r *http.Request, fallback int) int {
	value := r.URL.Query().Get("page")
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}

func parseInt64Query(r *http.Request, key string) *int64 {
	value := r.URL.Query().Get(key)
	if value == "" {
		return nil
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return nil
	}
	return &parsed
}

func parseBoolQuery(r *http.Request, key string) *bool {
	value := r.URL.Query().Get(key)
	if value == "" {
		return nil
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return nil
	}
	return &parsed
}

func parseTimeQuery(r *http.Request, key string) *time.Time {
	value := r.URL.Query().Get(key)
	if value == "" {
		return nil
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return nil
	}
	return &parsed
}
