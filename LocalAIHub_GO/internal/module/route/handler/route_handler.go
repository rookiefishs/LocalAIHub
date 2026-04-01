package handler

import (
	"net/http"
	"strings"
	"time"

	"localaihub/localaihub_go/internal/module/route/service"
	"localaihub/localaihub_go/internal/pkg/appctx"
	"localaihub/localaihub_go/internal/pkg/httpx"
	"localaihub/localaihub_go/internal/pkg/netx"
	"localaihub/localaihub_go/internal/pkg/response"
)

type RouteHandler struct{ service *service.RouteService }

func NewRouteHandler(service *service.RouteService) *RouteHandler {
	return &RouteHandler{service: service}
}

func (h *RouteHandler) List(w http.ResponseWriter, r *http.Request) {
	page := httpx.ParsePage(r, 1)
	pageSize := httpx.ParsePageSize(r, 10)
	items, total, err := h.service.List(r.Context(), page, pageSize)
	if err != nil {
		response.AdminError(w, r, http.StatusInternalServerError, 500100, "list routes failed")
		return
	}
	response.AdminSuccess(w, r, map[string]any{"items": items, "total": total, "page": page, "page_size": pageSize})
}

func (h *RouteHandler) Get(w http.ResponseWriter, r *http.Request, virtualModelID int64) {
	item, err := h.service.Get(r.Context(), virtualModelID)
	if err != nil {
		response.AdminError(w, r, http.StatusInternalServerError, 500100, "get route failed")
		return
	}
	if item == nil {
		response.AdminError(w, r, http.StatusNotFound, 404100, "route not found")
		return
	}
	response.AdminSuccess(w, r, item)
}

func (h *RouteHandler) Switch(w http.ResponseWriter, r *http.Request, virtualModelID int64) {
	var req struct {
		TargetBindingID int64   `json:"target_binding_id"`
		ManualLock      bool    `json:"manual_lock"`
		LockUntil       *string `json:"lock_until"`
		Reason          string  `json:"reason"`
	}
	if err := httpx.DecodeJSON(r, &req); err != nil {
		response.AdminError(w, r, http.StatusBadRequest, 400100, "invalid parameters")
		return
	}
	var lockUntil *time.Time
	if req.LockUntil != nil && *req.LockUntil != "" {
		parsed, err := time.Parse(time.RFC3339, *req.LockUntil)
		if err != nil {
			response.AdminError(w, r, http.StatusBadRequest, 400100, "invalid lock_until")
			return
		}
		lockUntil = &parsed
	}
	if err := h.service.Switch(r.Context(), virtualModelID, req.TargetBindingID, req.ManualLock, lockUntil, req.Reason, appctx.AdminUserID(r.Context()), netx.ClientIP(r), r.UserAgent()); err != nil {
		response.AdminError(w, r, http.StatusInternalServerError, 500100, "switch route failed")
		return
	}
	item, _ := h.service.Get(r.Context(), virtualModelID)
	response.AdminSuccess(w, r, item)
}

func (h *RouteHandler) Unlock(w http.ResponseWriter, r *http.Request, virtualModelID int64) {
	if err := h.service.Unlock(r.Context(), virtualModelID, netx.ClientIP(r), r.UserAgent()); err != nil {
		response.AdminError(w, r, http.StatusInternalServerError, 500100, "unlock route failed")
		return
	}
	item, _ := h.service.Get(r.Context(), virtualModelID)
	response.AdminSuccess(w, r, item)
}

func (h *RouteHandler) Delete(w http.ResponseWriter, r *http.Request, virtualModelID int64) {
	if err := h.service.Delete(r.Context(), virtualModelID, netx.ClientIP(r), r.UserAgent()); err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "cannot delete route") {
			response.AdminError(w, r, http.StatusBadRequest, 500100, "delete route failed: model still has bindings")
			return
		}
		response.AdminError(w, r, http.StatusInternalServerError, 500100, "delete route failed")
		return
	}
	response.AdminSuccess(w, r, map[string]any{"id": virtualModelID, "deleted": true})
}
