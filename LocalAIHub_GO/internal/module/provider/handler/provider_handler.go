package handler

import (
	"net/http"

	"localaihub/localaihub_go/internal/module/provider/repository"
	"localaihub/localaihub_go/internal/module/provider/service"
	"localaihub/localaihub_go/internal/pkg/httpx"
	"localaihub/localaihub_go/internal/pkg/netx"
	"localaihub/localaihub_go/internal/pkg/response"
)

type ProviderHandler struct{ service *service.ProviderService }

func NewProviderHandler(service *service.ProviderService) *ProviderHandler {
	return &ProviderHandler{service: service}
}

func (h *ProviderHandler) List(w http.ResponseWriter, r *http.Request) {
	page := httpx.ParsePage(r, 1)
	pageSize := httpx.ParsePageSize(r, 10)
	items, total, err := h.service.List(r.Context(), page, pageSize)
	if err != nil {
		response.AdminError(w, r, http.StatusInternalServerError, 500100, "list providers failed")
		return
	}
	response.AdminSuccess(w, r, map[string]any{"items": items, "total": total, "page": page, "page_size": pageSize})
}

func (h *ProviderHandler) Get(w http.ResponseWriter, r *http.Request, id int64) {
	item, err := h.service.Get(r.Context(), id)
	if err != nil {
		response.AdminError(w, r, http.StatusInternalServerError, 500100, "get provider failed")
		return
	}
	if item == nil {
		response.AdminError(w, r, http.StatusNotFound, 404100, "provider not found")
		return
	}
	response.AdminSuccess(w, r, item)
}

func (h *ProviderHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req repository.Provider
	if err := httpx.DecodeJSON(r, &req); err != nil {
		response.AdminError(w, r, http.StatusBadRequest, 400100, "invalid parameters")
		return
	}
	id, err := h.service.Create(r.Context(), &req, netx.ClientIP(r), r.UserAgent())
	if err != nil {
		response.AdminError(w, r, http.StatusInternalServerError, 500100, "create provider failed")
		return
	}
	item, _ := h.service.Get(r.Context(), id)
	response.AdminSuccess(w, r, item)
}

func (h *ProviderHandler) Update(w http.ResponseWriter, r *http.Request, id int64) {
	var req repository.Provider
	if err := httpx.DecodeJSON(r, &req); err != nil {
		response.AdminError(w, r, http.StatusBadRequest, 400100, "invalid parameters")
		return
	}
	req.ID = id
	if err := h.service.Update(r.Context(), &req, netx.ClientIP(r), r.UserAgent()); err != nil {
		response.AdminError(w, r, http.StatusInternalServerError, 500100, "update provider failed")
		return
	}
	item, _ := h.service.Get(r.Context(), id)
	response.AdminSuccess(w, r, item)
}

func (h *ProviderHandler) UpdateStatus(w http.ResponseWriter, r *http.Request, id int64) {
	var req struct {
		Enabled bool `json:"enabled"`
	}
	if err := httpx.DecodeJSON(r, &req); err != nil {
		response.AdminError(w, r, http.StatusBadRequest, 400100, "invalid parameters")
		return
	}
	if err := h.service.UpdateStatus(r.Context(), id, req.Enabled, netx.ClientIP(r), r.UserAgent()); err != nil {
		response.AdminError(w, r, http.StatusInternalServerError, 500100, "update provider status failed")
		return
	}
	item, _ := h.service.Get(r.Context(), id)
	response.AdminSuccess(w, r, item)
}

func (h *ProviderHandler) TestConnection(w http.ResponseWriter, r *http.Request, id int64) {
	item, err := h.service.Get(r.Context(), id)
	if err != nil {
		response.AdminError(w, r, http.StatusInternalServerError, 500100, "get provider failed")
		return
	}
	if item == nil {
		response.AdminError(w, r, http.StatusNotFound, 404100, "provider not found")
		return
	}
	result, err := h.service.TestConnection(r.Context(), item, netx.ClientIP(r), r.UserAgent())
	if err != nil {
		response.AdminError(w, r, http.StatusInternalServerError, 500100, "test provider connection failed")
		return
	}
	response.AdminSuccess(w, r, result)
}

func (h *ProviderHandler) Delete(w http.ResponseWriter, r *http.Request, id int64) {
	if err := h.service.Delete(r.Context(), id, netx.ClientIP(r), r.UserAgent()); err != nil {
		response.AdminError(w, r, http.StatusInternalServerError, 500100, "delete provider failed")
		return
	}
	response.AdminSuccess(w, r, map[string]any{"id": id, "deleted": true})
}
