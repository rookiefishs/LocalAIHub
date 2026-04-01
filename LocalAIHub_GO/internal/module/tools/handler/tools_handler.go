package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	gatewayservice "localaihub/localaihub_go/internal/module/gateway/service"
	"localaihub/localaihub_go/internal/pkg/response"
)

type ToolsHandler struct {
	gatewayService *gatewayservice.GatewayService
}

func NewToolsHandler(gatewayService *gatewayservice.GatewayService) *ToolsHandler {
	return &ToolsHandler{gatewayService: gatewayService}
}

type TestRequestInput struct {
	APIKey      string                         `json:"api_key"`
	Model       string                         `json:"model"`
	Messages    []gatewayservice.OpenAIMessage `json:"messages"`
	Stream      bool                           `json:"stream"`
	Temperature *float64                       `json:"temperature,omitempty"`
	MaxTokens   *int                           `json:"max_tokens,omitempty"`
}

func (h *ToolsHandler) TestRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.AdminError(w, r, http.StatusMethodNotAllowed, 405100, "method not allowed")
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		response.AdminError(w, r, http.StatusBadRequest, 400100, "invalid request body")
		return
	}
	defer r.Body.Close()

	var input TestRequestInput
	if err := json.Unmarshal(body, &input); err != nil {
		response.AdminError(w, r, http.StatusBadRequest, 400100, "invalid json")
		return
	}

	if input.APIKey == "" || input.Model == "" || len(input.Messages) == 0 {
		response.AdminError(w, r, http.StatusBadRequest, 400101, "api_key, model and messages are required")
		return
	}

	authHeader := "Bearer " + input.APIKey
	client, err := h.gatewayService.AuthenticateClientForTest(r.Context(), authHeader)
	if err != nil {
		response.AdminError(w, r, http.StatusUnauthorized, 401100, "invalid api key: "+err.Error())
		return
	}

	start := time.Now()
	req := gatewayservice.OpenAIChatCompletionRequest{
		Model:       input.Model,
		Messages:    input.Messages,
		Stream:      input.Stream,
		Temperature: input.Temperature,
		MaxTokens:   input.MaxTokens,
	}

	resp, statusCode, err := h.gatewayService.ForwardOpenAIChatCompletion(r.Context(), client, req)
	latency := int(time.Since(start).Milliseconds())

	if err != nil {
		_ = h.gatewayService.UpdateClientStatusAfterTest(r.Context(), client.ID, false)
		response.AdminSuccess(w, r, map[string]any{
			"success":     false,
			"model":       input.Model,
			"latency_ms":  latency,
			"error":       err.Error(),
			"status_code": statusCode,
			"key_status":  "disabled",
		})
		return
	}

	var promptTokens, completionTokens, totalTokens int
	if usage, ok := resp["usage"].(map[string]any); ok {
		if pt, ok := usage["prompt_tokens"].(float64); ok {
			promptTokens = int(pt)
		}
		if ct, ok := usage["completion_tokens"].(float64); ok {
			completionTokens = int(ct)
		}
		if tt, ok := usage["total_tokens"].(float64); ok {
			totalTokens = int(tt)
		}
	}

	_ = h.gatewayService.UpdateClientStatusAfterTest(r.Context(), client.ID, true)

	response.AdminSuccess(w, r, map[string]any{
		"success":           true,
		"model":             input.Model,
		"latency_ms":        latency,
		"prompt_tokens":     promptTokens,
		"completion_tokens": completionTokens,
		"total_tokens":      totalTokens,
		"key_status":        "active",
		"response":          resp,
	})
}
