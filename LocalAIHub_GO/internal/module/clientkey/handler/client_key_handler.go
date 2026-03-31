package handler

import (
	"net/http"

	"localaihub/localaihub_go/internal/module/clientkey/service"
	"localaihub/localaihub_go/internal/pkg/httpx"
	"localaihub/localaihub_go/internal/pkg/netx"
	"localaihub/localaihub_go/internal/pkg/response"
)

type ClientKeyHandler struct{ service *service.ClientKeyService }

func NewClientKeyHandler(service *service.ClientKeyService) *ClientKeyHandler {
	return &ClientKeyHandler{service: service}
}

func (h *ClientKeyHandler) List(w http.ResponseWriter, r *http.Request) {
	page := httpx.ParsePage(r, 1)
	pageSize := httpx.ParsePageSize(r, 10)
	items, total, err := h.service.List(r.Context(), page, pageSize)
	if err != nil {
		response.AdminError(w, r, http.StatusInternalServerError, 500100, "list client keys failed")
		return
	}
	response.AdminSuccess(w, r, map[string]any{"items": items, "total": total, "page": page, "page_size": pageSize})
}

func (h *ClientKeyHandler) Get(w http.ResponseWriter, r *http.Request, id int64) {
	item, err := h.service.Get(r.Context(), id)
	if err != nil {
		response.AdminError(w, r, http.StatusInternalServerError, 500100, "get client key failed")
		return
	}
	if item == nil {
		response.AdminError(w, r, http.StatusNotFound, 404100, "client key not found")
		return
	}
	allowedModels, _ := h.service.GetAllowedModels(r.Context(), id)
	response.AdminSuccess(w, r, map[string]any{
		"id":             item.ID,
		"name":           item.Name,
		"key_prefix":     item.KeyPrefix,
		"plain_key":      item.PlainKey,
		"status":         item.Status,
		"remark":         item.Remark,
		"last_used_at":   item.LastUsedAt,
		"expires_at":     item.ExpiresAt,
		"allowed_models": allowedModels,
		"created_at":     item.CreatedAt,
		"updated_at":     item.UpdatedAt,
	})
}

func (h *ClientKeyHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name          string  `json:"name"`
		Remark        string  `json:"remark"`
		ExpiresAt     string  `json:"expires_at"`
		AllowedModels []int64 `json:"allowed_models"`
	}
	if err := httpx.DecodeJSON(r, &req); err != nil {
		response.AdminError(w, r, http.StatusBadRequest, 400100, "invalid parameters")
		return
	}
	created, err := h.service.Create(r.Context(), req.Name, req.Remark, req.ExpiresAt, req.AllowedModels, netx.ClientIP(r), r.UserAgent())
	if err != nil {
		response.AdminError(w, r, http.StatusInternalServerError, 500100, "create client key failed")
		return
	}
	response.AdminSuccess(w, r, map[string]any{"id": created.Key.ID, "name": created.Key.Name, "key_prefix": created.Key.KeyPrefix, "plain_key": created.PlainKey, "status": created.Key.Status})
}

func (h *ClientKeyHandler) Update(w http.ResponseWriter, r *http.Request, id int64) {
	var req struct {
		Name          string  `json:"name"`
		Remark        string  `json:"remark"`
		AllowedModels []int64 `json:"allowed_models"`
	}
	if err := httpx.DecodeJSON(r, &req); err != nil {
		response.AdminError(w, r, http.StatusBadRequest, 400100, "invalid parameters")
		return
	}
	if err := h.service.Update(r.Context(), id, req.Name, req.Remark, netx.ClientIP(r), r.UserAgent(), req.AllowedModels); err != nil {
		response.AdminError(w, r, http.StatusInternalServerError, 500100, "update client key failed")
		return
	}
	item, _ := h.service.Get(r.Context(), id)
	response.AdminSuccess(w, r, item)
}

func (h *ClientKeyHandler) Test(w http.ResponseWriter, r *http.Request, id int64) {
	result, err := h.service.Test(r.Context(), id)
	if err != nil {
		response.AdminError(w, r, http.StatusBadRequest, 400100, err.Error())
		return
	}
	response.AdminSuccess(w, r, result)
}

func (h *ClientKeyHandler) UpdateStatus(w http.ResponseWriter, r *http.Request, id int64) {
	var req struct {
		Status string `json:"status"`
	}
	if err := httpx.DecodeJSON(r, &req); err != nil {
		response.AdminError(w, r, http.StatusBadRequest, 400100, "invalid parameters")
		return
	}
	if err := h.service.UpdateStatus(r.Context(), id, req.Status, netx.ClientIP(r), r.UserAgent()); err != nil {
		response.AdminError(w, r, http.StatusInternalServerError, 500100, "update client key status failed")
		return
	}
	item, _ := h.service.Get(r.Context(), id)
	response.AdminSuccess(w, r, item)
}

func (h *ClientKeyHandler) Delete(w http.ResponseWriter, r *http.Request, id int64) {
	if err := h.service.Delete(r.Context(), id, netx.ClientIP(r), r.UserAgent()); err != nil {
		response.AdminError(w, r, http.StatusInternalServerError, 500100, "delete client key failed")
		return
	}
	response.AdminSuccess(w, r, map[string]any{"id": id, "deleted": true})
}
