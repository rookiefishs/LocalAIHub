package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	gatewayservice "localaihub/localaihub_go/internal/module/gateway/service"
	"localaihub/localaihub_go/internal/pkg/logger"
	"localaihub/localaihub_go/internal/pkg/response"
)

type ProxyHandler struct {
	service *gatewayservice.GatewayService
}

func NewProxyHandler(service *gatewayservice.GatewayService) *ProxyHandler {
	return &ProxyHandler{service: service}
}

func (h *ProxyHandler) OpenAIChatCompletions(w http.ResponseWriter, r *http.Request) {
	logger.Log.Debug().Str("method", r.Method).Str("path", r.URL.Path).Msg("request received")

	client, err := h.service.AuthenticateClient(r.Context(), r.Header.Get("Authorization"))
	if err != nil {
		logger.Log.Warn().Err(err).Msg("authentication failed")
		response.JSON(w, http.StatusUnauthorized, map[string]any{
			"error": map[string]any{"message": "unauthorized", "type": "authentication_error", "code": "GW401001"},
		})
		return
	}

	body, _ := io.ReadAll(r.Body)
	logger.Log.Debug().Int("body_size", len(body)).Msg("proxy request body received")
	result, err := h.service.ProxyOpenAIRequest(r.Context(), client, r.Method, r.URL.Path, r.URL.RawQuery, body, r.Header)
	if err != nil {
		response.JSON(w, http.StatusBadRequest, map[string]any{
			"error": map[string]any{"message": err.Error(), "type": "gateway_error", "code": "GW422001"},
		})
		return
	}
	writeProxyResponse(w, result)
}

func (h *ProxyHandler) OpenAIResponses(w http.ResponseWriter, r *http.Request) {
	client, err := h.service.AuthenticateClient(r.Context(), r.Header.Get("Authorization"))
	if err != nil {
		response.JSON(w, http.StatusUnauthorized, map[string]any{
			"error": map[string]any{"message": "unauthorized", "type": "authentication_error", "code": "GW401001"},
		})
		return
	}
	body, _ := io.ReadAll(r.Body)
	logger.Log.Debug().Int("body_size", len(body)).Msg("proxy request body received")
	result, err := h.service.ProxyOpenAIRequest(r.Context(), client, r.Method, r.URL.Path, r.URL.RawQuery, body, r.Header)
	if err != nil {
		response.JSON(w, http.StatusBadRequest, map[string]any{
			"error": map[string]any{"message": err.Error(), "type": "gateway_error", "code": "GW422001"},
		})
		return
	}
	writeProxyResponse(w, result)
}

func (h *ProxyHandler) OpenAIProxy(w http.ResponseWriter, r *http.Request) {
	logger.Log.Debug().Str("method", r.Method).Str("path", r.URL.Path).Msg("request received")
	client, err := h.service.AuthenticateClient(r.Context(), r.Header.Get("Authorization"))
	if err != nil {
		response.JSON(w, http.StatusUnauthorized, map[string]any{
			"error": map[string]any{"message": "unauthorized", "type": "authentication_error", "code": "GW401001"},
		})
		return
	}
	body, _ := io.ReadAll(r.Body)
	logger.Log.Debug().Int("body_size", len(body)).Msg("proxy request body received")
	result, err := h.service.ProxyOpenAIRequest(r.Context(), client, r.Method, r.URL.Path, r.URL.RawQuery, body, r.Header)
	if err != nil {
		response.JSON(w, http.StatusBadRequest, map[string]any{
			"error": map[string]any{"message": err.Error(), "type": "gateway_error", "code": "GW422001"},
		})
		return
	}
	writeProxyResponse(w, result)
}

func writeProxyResponse(w http.ResponseWriter, result *gatewayservice.ProxyHTTPResult) {
	for key, values := range result.Headers {
		lower := strings.ToLower(key)
		if lower == "content-length" || lower == "transfer-encoding" || lower == "connection" {
			continue
		}
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
	w.WriteHeader(result.StatusCode)
	_, _ = w.Write(result.Body)
}

func (h *ProxyHandler) OpenAIModels(w http.ResponseWriter, r *http.Request) {
	_, err := h.service.AuthenticateClient(r.Context(), r.Header.Get("Authorization"))
	if err != nil {
		response.JSON(w, http.StatusUnauthorized, map[string]any{
			"error": map[string]any{"message": "unauthorized", "type": "authentication_error", "code": "GW401001"},
		})
		return
	}
	items, err := h.service.ListOpenAIModels(r.Context())
	if err != nil {
		response.JSON(w, http.StatusInternalServerError, map[string]any{
			"error": map[string]any{"message": "list models failed", "type": "server_error", "code": "GW500001"},
		})
		return
	}
	data := make([]map[string]any, 0, len(items))
	for _, item := range items {
		data = append(data, map[string]any{
			"id":       item["model_code"],
			"object":   "model",
			"owned_by": "local-gateway",
		})
	}
	response.JSON(w, http.StatusOK, map[string]any{"object": "list", "data": data})
}

func (h *ProxyHandler) AnthropicMessages(w http.ResponseWriter, r *http.Request) {
	client, err := h.service.AuthenticateClient(r.Context(), r.Header.Get("Authorization"))
	if err != nil {
		client, err = h.service.AuthenticateClient(r.Context(), r.Header.Get("x-api-key"))
		if err != nil {
			response.JSON(w, http.StatusUnauthorized, map[string]any{
				"error": map[string]any{"message": "unauthorized", "type": "authentication_error", "code": "GW401001"},
			})
			return
		}
	}

	var req gatewayservice.AnthropicMessagesRequest
	body, _ := io.ReadAll(r.Body)
	if err := json.Unmarshal(body, &req); err != nil {
		response.JSON(w, http.StatusBadRequest, map[string]any{
			"error": map[string]any{"message": "invalid request body", "type": "invalid_request_error", "code": "GW422001"},
		})
		return
	}

	result, statusCode, err := h.service.ForwardAnthropicMessages(r.Context(), client, req)
	if err != nil {
		response.JSON(w, statusCode, map[string]any{
			"error": map[string]any{"message": err.Error(), "type": "gateway_error", "code": gatewayErrorCode(statusCode)},
		})
		return
	}

	response.JSON(w, statusCode, result)
}

func (h *ProxyHandler) AnthropicModels(w http.ResponseWriter, r *http.Request) {
	response.JSON(w, http.StatusOK, map[string]any{"data": []map[string]any{}})
}

func (h *ProxyHandler) GeminiGeneratePlaceholder(w http.ResponseWriter, r *http.Request) {
	if !strings.Contains(r.URL.Path, ":generateContent") && !strings.Contains(r.URL.Path, ":streamGenerateContent") {
		http.NotFound(w, r)
		return
	}
	response.JSON(w, http.StatusNotImplemented, map[string]any{
		"error": map[string]any{"message": "Gemini generateContent gateway endpoint is scaffolded but not implemented yet", "code": "GW500001"},
	})
}

func (h *ProxyHandler) GeminiModels(w http.ResponseWriter, r *http.Request) {
	response.JSON(w, http.StatusOK, map[string]any{"models": []map[string]any{}})
}

func gatewayErrorCode(statusCode int) string {
	switch statusCode {
	case http.StatusBadRequest:
		return "GW422001"
	case http.StatusUnauthorized:
		return "GW401001"
	case http.StatusNotFound:
		return "GW404001"
	case http.StatusBadGateway:
		return "GW502001"
	default:
		return "GW500001"
	}
}
