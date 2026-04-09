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
		"token":         session.Token,
		"refresh_token": session.RefreshToken,
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

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type registerRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (h *AdminAuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		response.AdminError(w, r, http.StatusBadRequest, 400100, "invalid parameters")
		return
	}

	session, err := h.authService.Register(r.Context(), req.Username, req.Password, netx.ClientIP(r), r.UserAgent())
	if err != nil {
		if err == service.ErrUsernameTaken {
			response.AdminError(w, r, http.StatusConflict, 409100, "username already taken")
			return
		}
		if err == service.ErrRegistrationClosed {
			response.AdminError(w, r, http.StatusForbidden, 403100, "registration is closed")
			return
		}
		response.AdminError(w, r, http.StatusBadRequest, 400101, err.Error())
		return
	}

	response.AdminSuccess(w, r, map[string]any{
		"user": map[string]any{
			"id":       session.AdminID,
			"username": session.Username,
		},
		"token":         session.Token,
		"refresh_token": session.RefreshToken,
	})
}

func (h *AdminAuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req refreshRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		response.AdminError(w, r, http.StatusBadRequest, 400100, "invalid parameters")
		return
	}

	session, err := h.authService.RefreshToken(req.RefreshToken)
	if err != nil {
		response.AdminError(w, r, http.StatusUnauthorized, 401102, "invalid or expired refresh token")
		return
	}

	response.AdminSuccess(w, r, map[string]any{
		"user": map[string]any{
			"id":       session.AdminID,
			"username": session.Username,
		},
		"token":         session.Token,
		"refresh_token": session.RefreshToken,
	})
}
