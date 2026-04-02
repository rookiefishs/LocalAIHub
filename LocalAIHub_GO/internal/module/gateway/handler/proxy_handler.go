package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	gatewayservice "localaihub/localaihub_go/internal/module/gateway/service"
	"localaihub/localaihub_go/internal/pkg/logger"
)

type ProxyHandler struct {
	service *gatewayservice.GatewayService
}

func NewProxyHandler(service *gatewayservice.GatewayService) *ProxyHandler {
	return &ProxyHandler{service: service}
}

func (h *ProxyHandler) OpenAIChatCompletions(w http.ResponseWriter, r *http.Request) {
	logger.Log.Debug().Str("method", r.Method).Str("path", r.URL.Path).Msg("request received")

	client, err := h.service.AuthenticateClientWithRequest(r.Context(), r.Header.Get("Authorization"), gatewayservice.ClientIPFromRequest(r), r.UserAgent())
	if err != nil {
		logger.Log.Warn().Err(err).Msg("authentication failed")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{"message": err.Error(), "type": "authentication_error", "code": "GW401001"},
		})
		return
	}

	body, _ := io.ReadAll(r.Body)
	logger.Log.Debug().Int("body_size", len(body)).Msg("proxy request body received")
	result, err := h.service.ProxyOpenAIRequest(r.Context(), client, r.Method, r.URL.Path, r.URL.RawQuery, body, r.Header)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{"message": err.Error(), "type": "gateway_error", "code": "GW422001"},
		})
		return
	}
	writeProxyResponse(w, result)
}

func (h *ProxyHandler) OpenAIResponses(w http.ResponseWriter, r *http.Request) {
	client, err := h.service.AuthenticateClientWithRequest(r.Context(), r.Header.Get("Authorization"), gatewayservice.ClientIPFromRequest(r), r.UserAgent())
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{"message": err.Error(), "type": "authentication_error", "code": "GW401001"},
		})
		return
	}
	body, _ := io.ReadAll(r.Body)
	logger.Log.Debug().Int("body_size", len(body)).Msg("proxy request body received")
	result, err := h.service.ProxyOpenAIRequest(r.Context(), client, r.Method, r.URL.Path, r.URL.RawQuery, body, r.Header)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{"message": err.Error(), "type": "gateway_error", "code": "GW422001"},
		})
		return
	}
	writeProxyResponse(w, result)
}

func (h *ProxyHandler) OpenAIProxy(w http.ResponseWriter, r *http.Request) {
	logger.Log.Debug().Str("method", r.Method).Str("path", r.URL.Path).Msg("request received")
	client, err := h.service.AuthenticateClientWithRequest(r.Context(), r.Header.Get("Authorization"), gatewayservice.ClientIPFromRequest(r), r.UserAgent())
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{"message": err.Error(), "type": "authentication_error", "code": "GW401001"},
		})
		return
	}
	body, _ := io.ReadAll(r.Body)
	logger.Log.Debug().Int("body_size", len(body)).Msg("proxy request body received")
	result, err := h.service.ProxyOpenAIRequest(r.Context(), client, r.Method, r.URL.Path, r.URL.RawQuery, body, r.Header)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
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
	if result.IsStream && result.StreamBody != nil {
		defer result.StreamBody.Close()
		if flusher, ok := w.(http.Flusher); ok {
			w.WriteHeader(result.StatusCode)
			buffer := make([]byte, 32*1024)
			captured := make([]byte, 0, 64*1024)
			for {
				n, err := result.StreamBody.Read(buffer)
				if n > 0 {
					captured = append(captured, buffer[:n]...)
					_, _ = w.Write(buffer[:n])
					flusher.Flush()
				}
				if err != nil {
					break
				}
			}
			if result.AfterStream != nil {
				result.AfterStream(captured)
			}
			return
		}
	}
	w.WriteHeader(result.StatusCode)
	_, _ = w.Write(result.Body)
}

func (h *ProxyHandler) OpenAIModels(w http.ResponseWriter, r *http.Request) {
	_, err := h.service.AuthenticateClient(r.Context(), r.Header.Get("Authorization"))
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{"message": err.Error(), "type": "authentication_error", "code": "GW401001"},
		})
		return
	}
	items, err := h.service.ListOpenAIModels(r.Context())
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{"message": "获取模型列表失败", "type": "server_error", "code": "GW500001"},
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
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"object": "list", "data": data})
}

func (h *ProxyHandler) AnthropicMessages(w http.ResponseWriter, r *http.Request) {
	client, err := h.service.AuthenticateClient(r.Context(), r.Header.Get("Authorization"))
	if err != nil {
		client, err = h.service.AuthenticateClientWithRequest(r.Context(), r.Header.Get("x-api-key"), gatewayservice.ClientIPFromRequest(r), r.UserAgent())
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]any{
				"error": map[string]any{"message": err.Error(), "type": "authentication_error", "code": "GW401001"},
			})
			return
		}
	}

	var req gatewayservice.AnthropicMessagesRequest
	body, _ := io.ReadAll(r.Body)
	if err := json.Unmarshal(body, &req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{"message": "无效的请求体", "type": "invalid_request_error", "code": "GW422001"},
		})
		return
	}

	result, statusCode, err := h.service.ForwardAnthropicMessages(r.Context(), client, req)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{"message": err.Error(), "type": "gateway_error", "code": gatewayErrorCode(statusCode)},
		})
		return
	}
	writeProxyResponse(w, result)
}

func (h *ProxyHandler) AnthropicModels(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"data": []map[string]any{}})
}

func (h *ProxyHandler) GeminiGeneratePlaceholder(w http.ResponseWriter, r *http.Request) {
	if !strings.Contains(r.URL.Path, ":generateContent") && !strings.Contains(r.URL.Path, ":streamGenerateContent") {
		http.NotFound(w, r)
		return
	}
	client, err := h.service.AuthenticateClientWithRequest(r.Context(), r.Header.Get("Authorization"), gatewayservice.ClientIPFromRequest(r), r.UserAgent())
	if err != nil {
		client, err = h.service.AuthenticateClientWithRequest(r.Context(), r.Header.Get("x-api-key"), gatewayservice.ClientIPFromRequest(r), r.UserAgent())
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]any{
				"error": map[string]any{"message": err.Error(), "type": "authentication_error", "code": "GW401001"},
			})
			return
		}
	}
	modelCode, stream, parseErr := parseGeminiPath(r.URL.Path)
	if parseErr != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{"message": parseErr.Error(), "type": "invalid_request_error", "code": "GW422001"},
		})
		return
	}
	var req gatewayservice.GeminiGenerateContentRequest
	body, _ := io.ReadAll(r.Body)
	if err := json.Unmarshal(body, &req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{"message": "无效的请求体", "type": "invalid_request_error", "code": "GW422001"},
		})
		return
	}
	result, statusCode, err := h.service.ForwardGeminiGenerateContent(r.Context(), client, modelCode, req, stream)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{"message": err.Error(), "type": "gateway_error", "code": gatewayErrorCode(statusCode)},
		})
		return
	}
	writeProxyResponse(w, result)
}

func (h *ProxyHandler) GeminiModels(w http.ResponseWriter, r *http.Request) {
	_, err := h.service.AuthenticateClientWithRequest(r.Context(), r.Header.Get("Authorization"), gatewayservice.ClientIPFromRequest(r), r.UserAgent())
	if err != nil {
		_, err = h.service.AuthenticateClientWithRequest(r.Context(), r.Header.Get("x-api-key"), gatewayservice.ClientIPFromRequest(r), r.UserAgent())
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]any{
				"error": map[string]any{"message": err.Error(), "type": "authentication_error", "code": "GW401001"},
			})
			return
		}
	}
	items, err := h.service.ListGeminiModels(r.Context())
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{"message": "获取 Gemini 模型列表失败", "type": "server_error", "code": "GW500001"},
		})
		return
	}
	models := make([]map[string]any, 0, len(items))
	for _, item := range items {
		code, _ := item["model_code"].(string)
		models = append(models, map[string]any{
			"name":                       fmt.Sprintf("models/%s", code),
			"displayName":                item["display_name"],
			"description":                item["description"],
			"version":                    "local",
			"supportedGenerationMethods": []string{"generateContent", "streamGenerateContent"},
		})
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"models": models})
}

func parseGeminiPath(path string) (string, bool, error) {
	prefix := "/proxy/gemini/v1beta/models/"
	trimmed := strings.TrimPrefix(path, prefix)
	if trimmed == path || trimmed == "" {
		return "", false, fmt.Errorf("invalid gemini model path")
	}
	if strings.HasSuffix(trimmed, ":generateContent") {
		return strings.TrimSuffix(trimmed, ":generateContent"), false, nil
	}
	if strings.HasSuffix(trimmed, ":streamGenerateContent") {
		return strings.TrimSuffix(trimmed, ":streamGenerateContent"), true, nil
	}
	return "", false, fmt.Errorf("unsupported gemini action")
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
