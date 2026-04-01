package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"localaihub/localaihub_go/internal/module/configexport/service"
	"localaihub/localaihub_go/internal/pkg/response"
)

type ConfigExportHandler struct {
	svc *service.ExportService
}

func NewConfigExportHandler(svc *service.ExportService) *ConfigExportHandler {
	return &ConfigExportHandler{svc: svc}
}

func (h *ConfigExportHandler) Export(w http.ResponseWriter, r *http.Request) {
	data, err := h.svc.Export(r.Context())
	if err != nil {
		response.AdminError(w, r, http.StatusInternalServerError, 500100, "export failed: "+err.Error())
		return
	}

	result := map[string]interface{}{
		"version":     "1.0",
		"exported_at": time.Now().UTC().Format(time.RFC3339),
		"exported_by": "admin",
	}
	for k, v := range data {
		result[k] = v
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", `attachment; filename="localaihub-config-export.json"`)
	json.NewEncoder(w).Encode(result)
}

type ImportRequest struct {
	Config  service.ExportData `json:"config"`
	Options *ImportOptions     `json:"options"`
}

type ImportOptions struct {
	OverwriteExisting bool `json:"overwrite_existing"`
	SkipInvalid       bool `json:"skip_invalid"`
	DryRun            bool `json:"dry_run"`
}

func (h *ConfigExportHandler) Import(w http.ResponseWriter, r *http.Request) {
	var req ImportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.AdminError(w, r, http.StatusBadRequest, 400100, "invalid request body")
		return
	}

	if req.Config == nil {
		response.AdminError(w, r, http.StatusBadRequest, 400101, "config is required")
		return
	}

	opts := service.ImportOptions{Mode: "replace"}
	if req.Options != nil && !req.Options.OverwriteExisting {
		opts.Mode = "merge"
	}

	if req.Options != nil && req.Options.DryRun {
		summary := &service.ImportSummary{}
		if providers, ok := req.Config["providers"].([]interface{}); ok {
			for range providers {
				summary.ProvidersCreated++
			}
		}
		if models, ok := req.Config["virtual_models"].([]interface{}); ok {
			for range models {
				summary.VirtualModelsCreated++
			}
		}
		if bindings, ok := req.Config["bindings"].([]interface{}); ok {
			for range bindings {
				summary.BindingsCreated++
			}
		}
		if clients, ok := req.Config["api_clients"].([]interface{}); ok {
			for range clients {
				summary.ApiClientsCreated++
			}
		}
		response.AdminSuccess(w, r, map[string]interface{}{
			"dry_run": true,
			"summary": summary,
			"errors":  []string{},
		})
		return
	}

	summary, err := h.svc.Import(r.Context(), req.Config, opts)
	if err != nil {
		response.AdminError(w, r, http.StatusInternalServerError, 500100, "import failed: "+err.Error())
		return
	}

	response.AdminSuccess(w, r, map[string]interface{}{
		"dry_run": false,
		"summary": summary,
		"errors":  []string{},
	})
}
