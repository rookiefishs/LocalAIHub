package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	gatewayrepo "localaihub/localaihub_go/internal/module/gateway/repository"
	logrepo "localaihub/localaihub_go/internal/module/log/repository"
	logservice "localaihub/localaihub_go/internal/module/log/service"
	"localaihub/localaihub_go/internal/pkg/logger"
	"localaihub/localaihub_go/internal/pkg/response"
)

type LogHandler struct{ service *logservice.LogService }

func NewLogHandler(service *logservice.LogService) *LogHandler { return &LogHandler{service: service} }

func (h *LogHandler) ListRequestLogs(w http.ResponseWriter, r *http.Request) {
	logger.Log.Debug().Str("handler", "ListRequestLogs").Msg("handler called")
	limit := parseLimit(r, 10)
	page := parsePage(r, 1)
	filters := gatewayrepo.RequestLogFilters{Limit: limit, Page: page}
	if clientID := parseInt64Query(r, "client_id"); clientID != nil {
		filters.ClientID = clientID
	}
	if virtualModel := r.URL.Query().Get("virtual_model_code"); virtualModel != "" {
		filters.VirtualModelCode = virtualModel
	}
	if success := parseBoolQuery(r, "success"); success != nil {
		filters.Success = success
	}
	if timeRange := r.URL.Query().Get("time_range"); timeRange != "" {
		filters.TimeRange = timeRange
	} else {
		if start := parseTimeQuery(r, "start_time"); start != nil {
			filters.StartTime = start
		}
		if end := parseTimeQuery(r, "end_time"); end != nil {
			filters.EndTime = end
		}
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
	if keyword := r.URL.Query().Get("keyword"); keyword != "" {
		filters.Keyword = keyword
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

func (h *LogHandler) GetAuditLog(w http.ResponseWriter, r *http.Request, id int64) {
	item, err := h.service.GetAuditLog(r.Context(), id)
	if err != nil {
		response.AdminError(w, r, http.StatusInternalServerError, 500100, "get audit log failed")
		return
	}
	if item == nil {
		response.AdminError(w, r, http.StatusNotFound, 404100, "audit log not found")
		return
	}
	response.AdminSuccess(w, r, item)
}

func (h *LogHandler) ExportAuditLogs(w http.ResponseWriter, r *http.Request) {
	filters := logrepo.AuditLogFilters{Limit: 1000, Page: 1}
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
	if keyword := r.URL.Query().Get("keyword"); keyword != "" {
		filters.Keyword = keyword
	}
	if start := parseTimeQuery(r, "start_time"); start != nil {
		filters.StartTime = start
	}
	if end := parseTimeQuery(r, "end_time"); end != nil {
		filters.EndTime = end
	}
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="audit-logs-%s.csv"`, time.Now().UTC().Format("20060102-150405")))
	if err := h.service.ExportAuditLogsCSV(r.Context(), filters, w); err != nil {
		response.AdminError(w, r, http.StatusInternalServerError, 500100, "export audit logs failed")
		return
	}
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
