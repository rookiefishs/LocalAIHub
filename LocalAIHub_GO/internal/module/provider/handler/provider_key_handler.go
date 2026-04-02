package handler

import (
	"net/http"

	providerservice "localaihub/localaihub_go/internal/module/provider/service"
	"localaihub/localaihub_go/internal/pkg/httpx"
	"localaihub/localaihub_go/internal/pkg/netx"
	"localaihub/localaihub_go/internal/pkg/response"
)

type ProviderKeyHandler struct {
	service *providerservice.ProviderKeyService
}

func NewProviderKeyHandler(service *providerservice.ProviderKeyService) *ProviderKeyHandler {
	return &ProviderKeyHandler{service: service}
}

func (h *ProviderKeyHandler) List(w http.ResponseWriter, r *http.Request, providerID int64) {
	items, err := h.service.ListByProviderID(r.Context(), providerID)
	if err != nil {
		response.AdminError(w, r, http.StatusInternalServerError, 500100, "list provider keys failed")
		return
	}
	page := httpx.ParsePage(r, 1)
	pageSize := httpx.ParsePageSize(r, 10)
	start, end := httpx.Paginate(len(items), page, pageSize)
	response.AdminSuccess(w, r, map[string]any{"items": items[start:end], "total": len(items), "page": page, "page_size": pageSize})
}

func (h *ProviderKeyHandler) Create(w http.ResponseWriter, r *http.Request, providerID int64) {
	var req struct {
		Secret   string `json:"secret"`
		Priority int    `json:"priority"`
		Remark   string `json:"remark"`
	}
	if err := httpx.DecodeJSON(r, &req); err != nil {
		response.AdminError(w, r, http.StatusBadRequest, 400100, "invalid parameters")
		return
	}
	id, err := h.service.Create(r.Context(), providerID, req.Secret, req.Priority, req.Remark, netx.ClientIP(r), r.UserAgent())
	if err != nil {
		response.AdminError(w, r, http.StatusInternalServerError, 500100, "create provider key failed")
		return
	}
	response.AdminSuccess(w, r, map[string]any{"id": id})
}

func (h *ProviderKeyHandler) Update(w http.ResponseWriter, r *http.Request, keyID int64) {
	var req struct {
		Secret   string `json:"secret"`
		Priority int    `json:"priority"`
		Remark   string `json:"remark"`
	}
	if err := httpx.DecodeJSON(r, &req); err != nil {
		response.AdminError(w, r, http.StatusBadRequest, 400100, "invalid parameters")
		return
	}
	if err := h.service.Update(r.Context(), keyID, req.Secret, req.Priority, req.Remark, netx.ClientIP(r), r.UserAgent()); err != nil {
		response.AdminError(w, r, http.StatusInternalServerError, 500100, "update provider key failed")
		return
	}
	response.AdminSuccess(w, r, map[string]any{"id": keyID})
}

func (h *ProviderKeyHandler) UpdateStatus(w http.ResponseWriter, r *http.Request, keyID int64) {
	var req struct {
		Status string `json:"status"`
	}
	if err := httpx.DecodeJSON(r, &req); err != nil {
		response.AdminError(w, r, http.StatusBadRequest, 400100, "invalid parameters")
		return
	}
	if err := h.service.UpdateStatus(r.Context(), keyID, req.Status, netx.ClientIP(r), r.UserAgent()); err != nil {
		response.AdminError(w, r, http.StatusInternalServerError, 500100, "update provider key status failed")
		return
	}
	response.AdminSuccess(w, r, map[string]any{"id": keyID, "status": req.Status})
}

func (h *ProviderKeyHandler) Delete(w http.ResponseWriter, r *http.Request, keyID int64) {
	if err := h.service.Delete(r.Context(), keyID, netx.ClientIP(r), r.UserAgent()); err != nil {
		response.AdminError(w, r, http.StatusInternalServerError, 500100, "delete provider key failed")
		return
	}
	response.AdminSuccess(w, r, map[string]any{"id": keyID, "deleted": true})
}

func (h *ProviderKeyHandler) Test(w http.ResponseWriter, r *http.Request, keyID int64) {
	result, err := h.service.TestConnection(r.Context(), keyID)
	if err != nil {
		response.AdminError(w, r, http.StatusInternalServerError, 500100, err.Error())
		return
	}
	response.AdminSuccess(w, r, result)
}

func (h *ProviderKeyHandler) UpdatePriority(w http.ResponseWriter, r *http.Request, keyID int64) {
	var req struct {
		Priority int `json:"priority"`
	}
	if err := httpx.DecodeJSON(r, &req); err != nil {
		response.AdminError(w, r, http.StatusBadRequest, 400100, "invalid parameters")
		return
	}
	if err := h.service.UpdatePriority(r.Context(), keyID, req.Priority); err != nil {
		response.AdminError(w, r, http.StatusInternalServerError, 500100, "update priority failed")
		return
	}
	response.AdminSuccess(w, r, map[string]any{"id": keyID, "priority": req.Priority})
}
