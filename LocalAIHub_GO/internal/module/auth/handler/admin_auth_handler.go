package handler

import (
	"net/http"

	"localaihub/localaihub_go/internal/module/auth/service"
	"localaihub/localaihub_go/internal/pkg/appctx"
	"localaihub/localaihub_go/internal/pkg/httpx"
	"localaihub/localaihub_go/internal/pkg/netx"
	"localaihub/localaihub_go/internal/pkg/response"
)

type AdminAuthHandler struct {
	authService *service.AuthService
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func NewAdminAuthHandler(authService *service.AuthService) *AdminAuthHandler {
	return &AdminAuthHandler{authService: authService}
}

func (h *AdminAuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		response.AdminError(w, r, http.StatusBadRequest, 400100, "invalid parameters")
		return
	}

	session, err := h.authService.Login(r.Context(), req.Username, req.Password, netx.ClientIP(r), r.UserAgent())
	if err != nil {
		response.AdminError(w, r, http.StatusUnauthorized, 401101, "invalid username or password")
		return
	}

	response.AdminSuccess(w, r, map[string]any{
		"user": map[string]any{
			"id":       session.AdminID,
			"username": session.Username,
		},
		"token": session.Token,
	})
}

func (h *AdminAuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	adminID := appctx.AdminUserID(r.Context())
	admin, err := h.authService.CurrentAdmin(r.Context(), adminID)
	if err != nil || admin == nil {
		response.AdminError(w, r, http.StatusUnauthorized, 401100, "unauthorized")
		return
	}

	response.AdminSuccess(w, r, map[string]any{
		"user": map[string]any{
			"id":       admin.ID,
			"username": admin.Username,
			"status":   admin.Status,
		},
	})
}
