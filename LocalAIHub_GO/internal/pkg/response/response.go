package response

import (
	"encoding/json"
	"net/http"

	"localaihub/localaihub_go/internal/pkg/appctx"
)

func JSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func AdminSuccess(w http.ResponseWriter, r *http.Request, data any) {
	JSON(w, http.StatusOK, map[string]any{
		"code":       0,
		"message":    "ok",
		"data":       data,
		"request_id": appctx.RequestID(r.Context()),
	})
}

func AdminError(w http.ResponseWriter, r *http.Request, status int, code int, message string) {
	JSON(w, status, map[string]any{
		"code":       code,
		"message":    message,
		"data":       nil,
		"request_id": appctx.RequestID(r.Context()),
	})
}
