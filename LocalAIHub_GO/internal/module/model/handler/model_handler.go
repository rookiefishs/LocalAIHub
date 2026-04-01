package handler

import (
	"net/http"
	"strings"

	"localaihub/localaihub_go/internal/module/model/repository"
	"localaihub/localaihub_go/internal/module/model/service"
	"localaihub/localaihub_go/internal/pkg/httpx"
	"localaihub/localaihub_go/internal/pkg/netx"
	"localaihub/localaihub_go/internal/pkg/response"
)

type ModelHandler struct{ service *service.ModelService }

func NewModelHandler(service *service.ModelService) *ModelHandler {
	return &ModelHandler{service: service}
}

func (h *ModelHandler) List(w http.ResponseWriter, r *http.Request) {
	page := httpx.ParsePage(r, 1)
	pageSize := httpx.ParsePageSize(r, 10)
	items, total, err := h.service.List(r.Context(), page, pageSize)
	if err != nil {
		response.AdminError(w, r, http.StatusInternalServerError, 500100, "list models failed")
		return
	}
	response.AdminSuccess(w, r, map[string]any{"items": items, "total": total, "page": page, "page_size": pageSize})
}

func (h *ModelHandler) Get(w http.ResponseWriter, r *http.Request, id int64) {
	item, err := h.service.Get(r.Context(), id)
	if err != nil {
		response.AdminError(w, r, http.StatusInternalServerError, 500100, "get model failed")
		return
	}
	if item == nil {
		response.AdminError(w, r, http.StatusNotFound, 404100, "model not found")
		return
	}
	response.AdminSuccess(w, r, item)
}

func (h *ModelHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req repository.Model
	if err := httpx.DecodeJSON(r, &req); err != nil {
		response.AdminError(w, r, http.StatusBadRequest, 400100, "invalid parameters")
		return
	}
	id, err := h.service.Create(r.Context(), &req, netx.ClientIP(r), r.UserAgent())
	if err != nil {
		response.AdminError(w, r, http.StatusInternalServerError, 500100, "create model failed")
		return
	}
	item, _ := h.service.Get(r.Context(), id)
	response.AdminSuccess(w, r, item)
}

func (h *ModelHandler) Update(w http.ResponseWriter, r *http.Request, id int64) {
	var req repository.Model
	if err := httpx.DecodeJSON(r, &req); err != nil {
		response.AdminError(w, r, http.StatusBadRequest, 400100, "invalid parameters")
		return
	}
	req.ID = id
	if err := h.service.Update(r.Context(), &req, netx.ClientIP(r), r.UserAgent()); err != nil {
		response.AdminError(w, r, http.StatusInternalServerError, 500100, "update model failed")
		return
	}
	item, _ := h.service.Get(r.Context(), id)
	response.AdminSuccess(w, r, item)
}

func (h *ModelHandler) ListBindings(w http.ResponseWriter, r *http.Request, modelID int64) {
	items, err := h.service.ListBindings(r.Context(), modelID)
	if err != nil {
		response.AdminError(w, r, http.StatusInternalServerError, 500100, "list bindings failed")
		return
	}
	page := httpx.ParsePage(r, 1)
	pageSize := httpx.ParsePageSize(r, 10)
	start, end := httpx.Paginate(len(items), page, pageSize)
	response.AdminSuccess(w, r, map[string]any{"items": items[start:end], "total": len(items), "page": page, "page_size": pageSize})
}

func (h *ModelHandler) CreateBinding(w http.ResponseWriter, r *http.Request, modelID int64) {
	var req repository.Binding
	if err := httpx.DecodeJSON(r, &req); err != nil {
		response.AdminError(w, r, http.StatusBadRequest, 400100, "invalid parameters")
		return
	}
	req.VirtualModelID = modelID
	id, err := h.service.CreateBinding(r.Context(), &req, netx.ClientIP(r), r.UserAgent())
	if err != nil {
		response.AdminError(w, r, http.StatusInternalServerError, 500100, "create binding failed")
		return
	}
	response.AdminSuccess(w, r, map[string]any{"id": id})
}

func (h *ModelHandler) UpdateBinding(w http.ResponseWriter, r *http.Request, modelID, bindingID int64) {
	var req repository.Binding
	if err := httpx.DecodeJSON(r, &req); err != nil {
		response.AdminError(w, r, http.StatusBadRequest, 400100, "invalid parameters")
		return
	}
	req.ID = bindingID
	req.VirtualModelID = modelID
	if err := h.service.UpdateBinding(r.Context(), &req); err != nil {
		response.AdminError(w, r, http.StatusInternalServerError, 500100, "update binding failed")
		return
	}
	response.AdminSuccess(w, r, map[string]any{"id": bindingID, "updated": true})
}

func (h *ModelHandler) Delete(w http.ResponseWriter, r *http.Request, id int64) {
	if err := h.service.Delete(r.Context(), id, netx.ClientIP(r), r.UserAgent()); err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "still referenced by") {
			response.AdminError(w, r, http.StatusBadRequest, 500100, "delete model failed: model is still referenced by bindings")
			return
		}
		response.AdminError(w, r, http.StatusInternalServerError, 500100, "delete model failed")
		return
	}
	response.AdminSuccess(w, r, map[string]any{"id": id, "deleted": true})
}

func (h *ModelHandler) DeleteBinding(w http.ResponseWriter, r *http.Request, modelID, bindingID int64) {
	if err := h.service.DeleteBinding(r.Context(), bindingID, netx.ClientIP(r), r.UserAgent()); err != nil {
		response.AdminError(w, r, http.StatusInternalServerError, 500100, "delete binding failed")
		return
	}
	response.AdminSuccess(w, r, map[string]any{"id": bindingID, "deleted": true})
}

func (h *ModelHandler) TestBinding(w http.ResponseWriter, r *http.Request, modelID, bindingID int64) {
	result, err := h.service.TestBinding(r.Context(), modelID, bindingID)
	if err != nil {
		response.AdminError(w, r, http.StatusInternalServerError, 500100, err.Error())
		return
	}
	response.AdminSuccess(w, r, result)
}
