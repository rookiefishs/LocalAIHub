package router

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	adminauth "localaihub/localaihub_go/internal/module/auth/handler"
	authservice "localaihub/localaihub_go/internal/module/auth/service"
	clientkeyhandler "localaihub/localaihub_go/internal/module/clientkey/handler"
	gatewayhandler "localaihub/localaihub_go/internal/module/gateway/handler"
	"localaihub/localaihub_go/internal/module/health/handler"
	loghandler "localaihub/localaihub_go/internal/module/log/handler"
	modelhandler "localaihub/localaihub_go/internal/module/model/handler"
	providerhandler "localaihub/localaihub_go/internal/module/provider/handler"
	routehandler "localaihub/localaihub_go/internal/module/route/handler"
	"localaihub/localaihub_go/internal/pkg/appctx"
	"localaihub/localaihub_go/internal/pkg/cors"
	"localaihub/localaihub_go/internal/pkg/requestid"
	"localaihub/localaihub_go/internal/pkg/response"
)

type Handlers struct {
	Auth        *adminauth.AdminAuthHandler
	System      *routehandler.DashboardHandler
	Proxy       *gatewayhandler.ProxyHandler
	Providers   *providerhandler.ProviderHandler
	ProviderKey *providerhandler.ProviderKeyHandler
	Models      *modelhandler.ModelHandler
	ClientKey   *clientkeyhandler.ClientKeyHandler
	Route       *routehandler.RouteHandler
	Logs        *loghandler.LogHandler
}

func New(handlers Handlers, authService *authservice.AuthService, allowedOrigins []string) http.Handler {
	mux := http.NewServeMux()

	healthHandler := handler.NewHealthHandler()
	proxyHandler := handlers.Proxy
	if proxyHandler == nil {
		panic("proxy handler is required")
	}

	mux.HandleFunc("GET /healthz", healthHandler.Healthz)
	mux.HandleFunc("POST /admin/api/v1/auth/login", handlers.Auth.Login)
	mux.Handle("GET /admin/api/v1/auth/me", adminAuthMiddleware(authService, http.HandlerFunc(handlers.Auth.Me)))
	mux.Handle("GET /admin/api/v1/dashboard/overview", adminAuthMiddleware(authService, http.HandlerFunc(handlers.System.DashboardOverview)))
	mux.Handle("GET /admin/api/v1/providers", adminAuthMiddleware(authService, http.HandlerFunc(handlers.Providers.List)))
	mux.Handle("POST /admin/api/v1/providers", adminAuthMiddleware(authService, http.HandlerFunc(handlers.Providers.Create)))
	mux.Handle("GET /admin/api/v1/models", adminAuthMiddleware(authService, http.HandlerFunc(handlers.Models.List)))
	mux.Handle("POST /admin/api/v1/models", adminAuthMiddleware(authService, http.HandlerFunc(handlers.Models.Create)))
	mux.Handle("GET /admin/api/v1/client-keys", adminAuthMiddleware(authService, http.HandlerFunc(handlers.ClientKey.List)))
	mux.Handle("POST /admin/api/v1/client-keys", adminAuthMiddleware(authService, http.HandlerFunc(handlers.ClientKey.Create)))
	mux.Handle("GET /admin/api/v1/routes", adminAuthMiddleware(authService, http.HandlerFunc(handlers.Route.List)))
	mux.Handle("GET /admin/api/v1/logs/requests", adminAuthMiddleware(authService, http.HandlerFunc(handlers.Logs.ListRequestLogs)))
	mux.Handle("GET /admin/api/v1/logs/audit", adminAuthMiddleware(authService, http.HandlerFunc(handlers.Logs.ListAuditLogs)))
	mux.Handle("GET /admin/api/v1/logs/requests/", adminAuthMiddleware(authService, dynamicLogHandler(handlers.Logs)))
	mux.Handle("GET /admin/api/v1/providers/", adminAuthMiddleware(authService, dynamicProviderHandler(handlers.Providers, handlers.ProviderKey)))
	mux.Handle("PUT /admin/api/v1/providers/", adminAuthMiddleware(authService, dynamicProviderHandler(handlers.Providers, handlers.ProviderKey)))
	mux.Handle("DELETE /admin/api/v1/providers/", adminAuthMiddleware(authService, dynamicProviderHandler(handlers.Providers, handlers.ProviderKey)))
	mux.Handle("POST /admin/api/v1/providers/", adminAuthMiddleware(authService, dynamicProviderHandler(handlers.Providers, handlers.ProviderKey)))
	mux.Handle("GET /admin/api/v1/models/", adminAuthMiddleware(authService, dynamicModelHandler(handlers.Models)))
	mux.Handle("PUT /admin/api/v1/models/", adminAuthMiddleware(authService, dynamicModelHandler(handlers.Models)))
	mux.Handle("DELETE /admin/api/v1/models/", adminAuthMiddleware(authService, dynamicModelHandler(handlers.Models)))
	mux.Handle("POST /admin/api/v1/models/", adminAuthMiddleware(authService, dynamicModelHandler(handlers.Models)))
	mux.Handle("GET /admin/api/v1/client-keys/", adminAuthMiddleware(authService, dynamicClientKeyHandler(handlers.ClientKey)))
	mux.Handle("DELETE /admin/api/v1/client-keys/", adminAuthMiddleware(authService, dynamicClientKeyHandler(handlers.ClientKey)))
	mux.Handle("POST /admin/api/v1/client-keys/", adminAuthMiddleware(authService, dynamicClientKeyHandler(handlers.ClientKey)))
	mux.Handle("GET /admin/api/v1/routes/", adminAuthMiddleware(authService, dynamicRouteHandler(handlers)))
	mux.Handle("POST /admin/api/v1/routes/", adminAuthMiddleware(authService, dynamicRouteHandler(handlers)))
	mux.Handle("DELETE /admin/api/v1/routes/", adminAuthMiddleware(authService, dynamicRouteHandler(handlers)))
	mux.HandleFunc("POST /proxy/openai/v1/chat/completions", proxyHandler.OpenAIChatCompletions)
	mux.HandleFunc("POST /proxy/openai/v1/responses", proxyHandler.OpenAIResponses)
	mux.HandleFunc("GET /proxy/openai/v1/models", proxyHandler.OpenAIModels)
	mux.HandleFunc("POST /proxy/openai/v1/", proxyHandler.OpenAIProxy)
	mux.HandleFunc("GET /proxy/openai/v1/", proxyHandler.OpenAIProxy)
	mux.HandleFunc("POST /proxy/anthropic/v1/messages", proxyHandler.AnthropicMessages)
	mux.HandleFunc("GET /proxy/anthropic/v1/models", proxyHandler.AnthropicModels)
	mux.HandleFunc("POST /proxy/gemini/v1beta/models/", proxyHandler.GeminiGeneratePlaceholder)
	mux.HandleFunc("GET /proxy/gemini/v1beta/models", proxyHandler.GeminiModels)

	return cors.Middleware(allowedOrigins, requestid.Middleware(mux))
}

func adminAuthMiddleware(authService *authservice.AuthService, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := bearerToken(r.Header.Get("Authorization"))
		if token == "" {
			response.AdminError(w, r, http.StatusUnauthorized, 401100, "unauthorized")
			return
		}
		session, err := authService.Authenticate(token)
		if err != nil {
			response.AdminError(w, r, http.StatusUnauthorized, 401100, "unauthorized")
			return
		}
		ctx := context.WithValue(r.Context(), appctx.AdminUserIDKey, session.AdminID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func bearerToken(header string) string {
	if !strings.HasPrefix(header, "Bearer ") {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
}

func dynamicProviderHandler(handler *providerhandler.ProviderHandler, keyHandler *providerhandler.ProviderKeyHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/admin/api/v1/providers/")
		if strings.Contains(path, "/keys") {
			parts := strings.Split(strings.Trim(path, "/"), "/")
			if len(parts) >= 2 && parts[1] == "keys" {
				providerID, err := strconv.ParseInt(parts[0], 10, 64)
				if err != nil {
					response.AdminError(w, r, http.StatusBadRequest, 400100, "invalid provider id")
					return
				}
				if len(parts) == 2 {
					if r.Method == http.MethodGet {
						keyHandler.List(w, r, providerID)
						return
					}
					if r.Method == http.MethodPost {
						keyHandler.Create(w, r, providerID)
						return
					}
				}
				if len(parts) == 4 && parts[3] == "status" {
					keyID, err := strconv.ParseInt(parts[2], 10, 64)
					if err != nil {
						response.AdminError(w, r, http.StatusBadRequest, 400100, "invalid provider key id")
						return
					}
					keyHandler.UpdateStatus(w, r, keyID)
					return
				}
				if len(parts) == 3 && r.Method == http.MethodDelete {
					keyID, err := strconv.ParseInt(parts[2], 10, 64)
					if err != nil {
						response.AdminError(w, r, http.StatusBadRequest, 400100, "invalid provider key id")
						return
					}
					keyHandler.Delete(w, r, keyID)
					return
				}
			}
		}
		if strings.HasSuffix(path, "/status") {
			if strings.HasSuffix(path, "/test-connection") {
				http.NotFound(w, r)
				return
			}
			id, err := strconv.ParseInt(strings.Trim(strings.TrimSuffix(path, "/status"), "/"), 10, 64)
			if err != nil {
				response.AdminError(w, r, http.StatusBadRequest, 400100, "invalid provider id")
				return
			}
			handler.UpdateStatus(w, r, id)
			return
		}
		if strings.HasSuffix(path, "/test-connection") {
			id, err := strconv.ParseInt(strings.Trim(strings.TrimSuffix(path, "/test-connection"), "/"), 10, 64)
			if err != nil {
				response.AdminError(w, r, http.StatusBadRequest, 400100, "invalid provider id")
				return
			}
			handler.TestConnection(w, r, id)
			return
		}
		id, err := strconv.ParseInt(strings.Trim(path, "/"), 10, 64)
		if err != nil {
			response.AdminError(w, r, http.StatusBadRequest, 400100, "invalid provider id")
			return
		}
		if r.Method == http.MethodGet {
			handler.Get(w, r, id)
			return
		}
		if r.Method == http.MethodPut {
			handler.Update(w, r, id)
			return
		}
		if r.Method == http.MethodDelete {
			handler.Delete(w, r, id)
			return
		}
		http.NotFound(w, r)
	}
}

func dynamicModelHandler(handler *modelhandler.ModelHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/admin/api/v1/models/")
		if strings.Contains(path, "/bindings") {
			parts := strings.Split(strings.Trim(path, "/"), "/")
			modelID, err := strconv.ParseInt(parts[0], 10, 64)
			if err != nil {
				response.AdminError(w, r, http.StatusBadRequest, 400100, "invalid model id")
				return
			}
			if len(parts) >= 3 {
				bindingID, err := strconv.ParseInt(parts[2], 10, 64)
				if err != nil {
					response.AdminError(w, r, http.StatusBadRequest, 400100, "invalid binding id")
					return
				}
				if r.Method == http.MethodDelete {
					handler.DeleteBinding(w, r, modelID, bindingID)
					return
				}
				if r.Method == http.MethodPut {
					handler.UpdateBinding(w, r, modelID, bindingID)
					return
				}
				if len(parts) >= 4 && parts[3] == "test" && r.Method == http.MethodPost {
					handler.TestBinding(w, r, modelID, bindingID)
					return
				}
			}
			if r.Method == http.MethodGet {
				handler.ListBindings(w, r, modelID)
				return
			}
			if r.Method == http.MethodPost {
				handler.CreateBinding(w, r, modelID)
				return
			}
			http.NotFound(w, r)
			return
		}
		id, err := strconv.ParseInt(strings.Trim(path, "/"), 10, 64)
		if err != nil {
			response.AdminError(w, r, http.StatusBadRequest, 400100, "invalid model id")
			return
		}
		if r.Method == http.MethodGet {
			handler.Get(w, r, id)
			return
		}
		if r.Method == http.MethodPut {
			handler.Update(w, r, id)
			return
		}
		if r.Method == http.MethodDelete {
			handler.Delete(w, r, id)
			return
		}
		http.NotFound(w, r)
	}
}

func dynamicClientKeyHandler(handler *clientkeyhandler.ClientKeyHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/admin/api/v1/client-keys/")
		if strings.HasSuffix(path, "/status") {
			id, err := strconv.ParseInt(strings.Trim(strings.TrimSuffix(path, "/status"), "/"), 10, 64)
			if err != nil {
				response.AdminError(w, r, http.StatusBadRequest, 400100, "invalid client key id")
				return
			}
			handler.UpdateStatus(w, r, id)
			return
		}
		if strings.HasSuffix(path, "/test") {
			id, err := strconv.ParseInt(strings.Trim(strings.TrimSuffix(path, "/test"), "/"), 10, 64)
			if err != nil {
				response.AdminError(w, r, http.StatusBadRequest, 400100, "invalid client key id")
				return
			}
			handler.Test(w, r, id)
			return
		}
		id, err := strconv.ParseInt(strings.Trim(path, "/"), 10, 64)
		if err != nil {
			response.AdminError(w, r, http.StatusBadRequest, 400100, "invalid client key id")
			return
		}
		if r.Method == http.MethodGet {
			handler.Get(w, r, id)
			return
		}
		if r.Method == http.MethodPut {
			handler.Update(w, r, id)
			return
		}
		if r.Method == http.MethodDelete {
			handler.Delete(w, r, id)
			return
		}
		http.NotFound(w, r)
	}
}

func dynamicRouteHandler(handlers Handlers) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/admin/api/v1/routes/")
		if strings.HasSuffix(path, "/switch") {
			id, err := strconv.ParseInt(strings.Trim(strings.TrimSuffix(path, "/switch"), "/"), 10, 64)
			if err != nil {
				response.AdminError(w, r, http.StatusBadRequest, 400100, "invalid route id")
				return
			}
			handlers.Route.Switch(w, r, id)
			return
		}
		if strings.HasSuffix(path, "/unlock") {
			id, err := strconv.ParseInt(strings.Trim(strings.TrimSuffix(path, "/unlock"), "/"), 10, 64)
			if err != nil {
				response.AdminError(w, r, http.StatusBadRequest, 400100, "invalid route id")
				return
			}
			handlers.Route.Unlock(w, r, id)
			return
		}
		id, err := strconv.ParseInt(strings.Trim(path, "/"), 10, 64)
		if err != nil {
			response.AdminError(w, r, http.StatusBadRequest, 400100, "invalid route id")
			return
		}
		if r.Method == http.MethodGet {
			handlers.Route.Get(w, r, id)
			return
		}
		if r.Method == http.MethodDelete {
			handlers.Route.Delete(w, r, id)
			return
		}
		http.NotFound(w, r)
	}
}

func dynamicLogHandler(handler *loghandler.LogHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/admin/api/v1/logs/requests/")
		id, err := strconv.ParseInt(strings.Trim(path, "/"), 10, 64)
		if err != nil {
			response.AdminError(w, r, http.StatusBadRequest, 400100, "invalid request log id")
			return
		}
		handler.GetRequestLog(w, r, id)
	}
}
