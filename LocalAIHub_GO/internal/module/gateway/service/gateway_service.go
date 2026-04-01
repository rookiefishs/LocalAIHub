package service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"localaihub/localaihub_go/internal/module/gateway/repository"
	providerrepo "localaihub/localaihub_go/internal/module/provider/repository"
	providerservice "localaihub/localaihub_go/internal/module/provider/service"
	routeservice "localaihub/localaihub_go/internal/module/route/service"
	"localaihub/localaihub_go/internal/pkg/appctx"
	"localaihub/localaihub_go/internal/pkg/logger"
)

type OpenAIChatCompletionRequest struct {
	Model             string          `json:"model"`
	Messages          []OpenAIMessage `json:"messages"`
	Tools             []any           `json:"tools,omitempty"`
	ToolChoice        any             `json:"tool_choice,omitempty"`
	ParallelToolCalls any             `json:"parallel_tool_calls,omitempty"`
	ResponseFormat    any             `json:"response_format,omitempty"`
	Temperature       *float64        `json:"temperature,omitempty"`
	TopP              *float64        `json:"top_p,omitempty"`
	MaxTokens         *int            `json:"max_tokens,omitempty"`
	Stream            bool            `json:"stream,omitempty"`
	Stop              any             `json:"stop,omitempty"`
	PresencePenalty   *float64        `json:"presence_penalty,omitempty"`
	FrequencyPenalty  *float64        `json:"frequency_penalty,omitempty"`
	User              *string         `json:"user,omitempty"`
}

type OpenAIMessage struct {
	Role       string `json:"role"`
	Content    any    `json:"content"`
	ToolCallID string `json:"tool_call_id,omitempty"`
	ToolCalls  []any  `json:"tool_calls,omitempty"`
	Name       string `json:"name,omitempty"`
}

type AnthropicMessagesRequest struct {
	Model       string             `json:"model"`
	MaxTokens   int                `json:"max_tokens"`
	System      string             `json:"system,omitempty"`
	Messages    []AnthropicMessage `json:"messages"`
	Temperature *float64           `json:"temperature,omitempty"`
	Stream      bool               `json:"stream,omitempty"`
}

type AnthropicMessage struct {
	Role    string `json:"role"`
	Content any    `json:"content"`
}

type GatewayService struct {
	repo               *repository.GatewayRepository
	providerKeyService *providerservice.ProviderKeyService
	routeService       *routeservice.RouteService
	httpClient         *http.Client
}

func NewGatewayService(repo *repository.GatewayRepository, providerKeyService *providerservice.ProviderKeyService, routeService *routeservice.RouteService) *GatewayService {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		MaxIdleConns:    10,
		IdleConnTimeout: 90 * time.Second,
	}
	return &GatewayService{
		repo:               repo,
		providerKeyService: providerKeyService,
		routeService:       routeService,
		httpClient:         &http.Client{Timeout: 120 * time.Second, Transport: tr},
	}
}

func (s *GatewayService) AuthenticateClient(ctx context.Context, authHeader string) (*repository.GatewayClient, error) {
	apiKey := bearerOrRawKey(authHeader)
	if apiKey == "" {
		return nil, fmt.Errorf("missing api key")
	}
	hash := sha256.Sum256([]byte(apiKey))
	hashString := hex.EncodeToString(hash[:])
	logger.Log.Debug().Str("hash", hashString).Msg("authenticating client")
	client, err := s.repo.GetClientByHash(ctx, hashString)
	if err != nil {
		logger.Log.Error().Err(err).Str("hash", hashString).Msg("failed to get client by hash")
		return nil, err
	}
	logger.Log.Debug().Interface("client", client).Msg("client lookup result")
	if client == nil || client.Status != "active" {
		return nil, fmt.Errorf("invalid client key")
	}
	if client.ExpiresAt != nil && client.ExpiresAt.Before(time.Now().UTC()) {
		return nil, fmt.Errorf("client key expired")
	}
	_ = s.repo.TouchClientLastUsed(ctx, client.ID)
	return client, nil
}

func (s *GatewayService) ListOpenAIModels(ctx context.Context) ([]map[string]any, error) {
	return s.repo.ListVisibleModels(ctx)
}

type ProxyHTTPResult struct {
	StatusCode int
	Headers    http.Header
	Body       []byte
}

func (s *GatewayService) ProxyOpenAIRequest(ctx context.Context, client *repository.GatewayClient, method, path, rawQuery string, rawBody []byte, incomingHeaders http.Header) (*ProxyHTTPResult, error) {
	route, err := s.resolveOpenAIRouteForProxy(ctx, client, path, rawBody)
	if err != nil {
		return nil, err
	}
	providerKey, secret, err := s.providerKeyService.SelectForRequest(ctx, route.ProviderID)
	if err != nil || providerKey == nil || secret == "" {
		return nil, fmt.Errorf("no available provider key")
	}
	requestURL := strings.TrimRight(route.BaseURL, "/")
	requestURL = strings.ReplaceAll(requestURL, "/v1/v1", "/v1")
	requestURL = strings.ReplaceAll(requestURL, "/v1/", "/")
	if !strings.HasSuffix(requestURL, "/v1") {
		requestURL += "/v1"
	}
	trimmedPath := strings.TrimPrefix(path, "/proxy/openai/v1")
	requestURL += trimmedPath
	if rawQuery != "" {
		requestURL += "?" + rawQuery
	}
	logger.Log.Debug().Str("method", method).Str("path", path).Str("target_url", requestURL).Str("auth_type", route.AuthType).Str("upstream_model", route.UpstreamModelName).Msg("transparent proxy forwarding")

	upstreamReq, err := http.NewRequestWithContext(ctx, method, requestURL, bytes.NewReader(rawBody))
	if err != nil {
		return nil, err
	}
	copyHeadersForProxy(upstreamReq.Header, incomingHeaders)
	if route.AuthType == "x_api_key" {
		upstreamReq.Header.Set("x-api-key", secret)
		upstreamReq.Header.Del("Authorization")
	} else {
		upstreamReq.Header.Set("Authorization", "Bearer "+secret)
		upstreamReq.Header.Del("x-api-key")
	}
	resp, err := s.httpClient.Do(upstreamReq)
	if err != nil {
		_ = s.providerKeyService.ReportResult(ctx, providerKey.ID, false, err.Error())
		_, _ = s.routeService.RegisterFailure(ctx, route.ProviderID, route.VirtualModelID, err.Error())
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		_ = s.providerKeyService.ReportResult(ctx, providerKey.ID, false, err.Error())
		_, _ = s.routeService.RegisterFailure(ctx, route.ProviderID, route.VirtualModelID, err.Error())
		return nil, err
	}
	if resp.StatusCode >= 400 {
		_ = s.providerKeyService.ReportResult(ctx, providerKey.ID, false, resp.Status)
		_, _ = s.routeService.RegisterFailure(ctx, route.ProviderID, route.VirtualModelID, resp.Status)
		s.logProxyRequest(ctx, client, route, providerKey, method, path, rawBody, resp.StatusCode, false, body, nil, nil)
	} else {
		_ = s.providerKeyService.ReportResult(ctx, providerKey.ID, true, "")
		_ = s.routeService.RegisterSuccess(ctx, route.ProviderID, route.VirtualModelID)
		s.logProxyRequest(ctx, client, route, providerKey, method, path, rawBody, resp.StatusCode, true, body, nil, nil)
	}
	return &ProxyHTTPResult{StatusCode: resp.StatusCode, Headers: resp.Header.Clone(), Body: body}, nil
}

func (s *GatewayService) resolveOpenAIRouteForProxy(ctx context.Context, client *repository.GatewayClient, path string, rawBody []byte) (*repository.ModelRoute, error) {
	logger.Log.Debug().Str("path", path).Str("model_from_body_check", "checking").Msg("resolve route start")
	if !strings.HasPrefix(path, "/proxy/openai") {
		models, err := s.repo.ListVisibleModels(ctx)
		if err != nil || len(models) == 0 {
			return nil, fmt.Errorf("no visible models configured")
		}
		firstCode, _ := models[0]["model_code"].(string)
		logger.Log.Debug().Str("model_code", firstCode).Msg("using first visible model")
		return s.repo.ResolveOpenAIModelRoute(ctx, firstCode)
	}
	modelCode, err := extractModelFromBody(rawBody)
	if err != nil {
		return nil, err
	}
	if modelCode == "" {
		return nil, fmt.Errorf("model is required")
	}
	logger.Log.Debug().Str("model_code", modelCode).Msg("resolving route for model")
	route, err := s.repo.ResolveOpenAIModelRoute(ctx, modelCode)
	if err != nil {
		logger.Log.Error().Err(err).Str("model_code", modelCode).Msg("resolve route error")
		return nil, err
	}
	if route == nil {
		logger.Log.Debug().Str("model_code", modelCode).Msg("no available route found")
		return nil, fmt.Errorf("virtual model not found")
	}
	logger.Log.Debug().Str("model_code", modelCode).Int64("provider_id", route.ProviderID).Str("provider_name", route.ProviderName).Msg("route resolved")
	allowed, err := s.repo.ClientCanAccessModel(ctx, client.ID, route.VirtualModelID)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, fmt.Errorf("client key is not allowed to access this model")
	}
	return route, nil
}

func extractModelFromBody(rawBody []byte) (string, error) {
	var payload map[string]any
	if err := json.Unmarshal(rawBody, &payload); err != nil {
		return "", fmt.Errorf("invalid request body")
	}
	model, _ := payload["model"].(string)
	return model, nil
}

func copyHeadersForProxy(target http.Header, source http.Header) {
	for key, values := range source {
		lower := strings.ToLower(key)
		if lower == "host" || lower == "content-length" || lower == "authorization" || lower == "x-api-key" {
			continue
		}
		for _, value := range values {
			target.Add(key, value)
		}
	}
}

func (s *GatewayService) ForwardOpenAIResponses(ctx context.Context, client *repository.GatewayClient, raw map[string]any) (map[string]any, bool, int, error) {
	model, _ := raw["model"].(string)
	if model == "" {
		return nil, false, http.StatusBadRequest, fmt.Errorf("model is required")
	}
	stream, _ := raw["stream"].(bool)
	route, err := s.repo.ResolveOpenAIModelRoute(ctx, model)
	if err != nil {
		return nil, stream, http.StatusInternalServerError, err
	}
	if route == nil {
		return nil, stream, http.StatusNotFound, fmt.Errorf("virtual model not found")
	}
	allowed, err := s.repo.ClientCanAccessModel(ctx, client.ID, route.VirtualModelID)
	if err != nil {
		return nil, stream, http.StatusInternalServerError, err
	}
	if !allowed {
		return nil, stream, http.StatusForbidden, fmt.Errorf("client key is not allowed to access this model")
	}
	providerKey, secret, err := s.providerKeyService.SelectForRequest(ctx, route.ProviderID)
	if err != nil || providerKey == nil || secret == "" {
		return nil, stream, http.StatusBadGateway, fmt.Errorf("no available provider key")
	}
	repaired := repairResponsesPayload(raw)
	repaired["model"] = route.UpstreamModelName
	repaired["stream"] = false
	body, err := json.Marshal(repaired)
	if err != nil {
		return nil, stream, http.StatusInternalServerError, err
	}
	requestURL := strings.TrimRight(route.BaseURL, "/")
	requestURL = strings.ReplaceAll(requestURL, "/v1/v1", "/v1")
	requestURL = strings.ReplaceAll(requestURL, "/v1/", "/")
	if !strings.HasSuffix(requestURL, "/v1") {
		requestURL += "/v1"
	}
	requestURL += "/responses"
	upstreamReq, err := http.NewRequestWithContext(ctx, http.MethodPost, requestURL, bytes.NewReader(body))
	if err != nil {
		return nil, stream, http.StatusInternalServerError, err
	}
	upstreamReq.Header.Set("Content-Type", "application/json")
	if route.AuthType == "x_api_key" {
		upstreamReq.Header.Set("x-api-key", secret)
	} else {
		upstreamReq.Header.Set("Authorization", "Bearer "+secret)
	}
	resp, err := s.httpClient.Do(upstreamReq)
	if err != nil {
		fallbackResult, fallbackErr := s.forwardResponsesViaChatFallback(ctx, client, raw, err.Error())
		if fallbackErr == nil {
			return fallbackResult, stream, http.StatusOK, nil
		}
		return nil, stream, http.StatusBadGateway, err
	}
	defer resp.Body.Close()
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, stream, http.StatusBadGateway, err
	}
	var mapped map[string]any
	if err := json.Unmarshal(responseBody, &mapped); err != nil {
		return nil, stream, http.StatusBadGateway, err
	}
	if resp.StatusCode >= 400 {
		if message, ok := mapped["error"].(map[string]any); ok {
			if msg, ok := message["message"].(string); ok {
				fallbackResult, fallbackErr := s.forwardResponsesViaChatFallback(ctx, client, raw, msg)
				if fallbackErr == nil {
					return fallbackResult, stream, http.StatusOK, nil
				}
				return nil, stream, resp.StatusCode, fmt.Errorf(msg)
			}
		}
		fallbackResult, fallbackErr := s.forwardResponsesViaChatFallback(ctx, client, raw, fmt.Sprintf("upstream error: %d", resp.StatusCode))
		if fallbackErr == nil {
			return fallbackResult, stream, http.StatusOK, nil
		}
		return nil, stream, resp.StatusCode, fmt.Errorf("upstream error: %d", resp.StatusCode)
	}
	return mapped, stream, resp.StatusCode, nil
}

func (s *GatewayService) forwardResponsesViaChatFallback(ctx context.Context, client *repository.GatewayClient, raw map[string]any, upstreamMessage string) (map[string]any, error) {
	chatReq, err := mapResponsesToChatRequest(raw)
	if err != nil {
		return nil, err
	}
	chatResp, _, runErr := s.ForwardOpenAIChatCompletion(ctx, client, chatReq)
	if runErr != nil {
		return nil, runErr
	}
	return mapChatCompletionToResponses(raw, chatResp, upstreamMessage), nil
}

func repairResponsesPayload(raw map[string]any) map[string]any {
	cloned := make(map[string]any, len(raw))
	for k, v := range raw {
		cloned[k] = v
	}
	input, ok := cloned["input"].([]any)
	if !ok {
		return cloned
	}
	lastCallID := ""
	for _, item := range input {
		obj, ok := item.(map[string]any)
		if !ok {
			continue
		}
		itemType, _ := obj["type"].(string)
		if itemType == "function_call" {
			if callID, ok := obj["call_id"].(string); ok && callID != "" {
				lastCallID = callID
			} else if id, ok := obj["id"].(string); ok && id != "" {
				obj["call_id"] = id
				lastCallID = id
			}
		}
		if itemType == "function_call_output" {
			if _, ok := obj["call_id"].(string); !ok || obj["call_id"] == "" {
				if lastCallID != "" {
					obj["call_id"] = lastCallID
				}
			}
		}
	}
	cloned["input"] = input
	return cloned
}

func mapResponsesToChatRequest(raw map[string]any) (OpenAIChatCompletionRequest, error) {
	model, _ := raw["model"].(string)
	stream, _ := raw["stream"].(bool)
	messages := make([]OpenAIMessage, 0)
	input, _ := raw["input"].([]any)
	for _, item := range input {
		obj, ok := item.(map[string]any)
		if !ok {
			continue
		}
		itemType, _ := obj["type"].(string)
		role, _ := obj["role"].(string)
		switch itemType {
		case "message", "input_text", "output_text", "":
			content := extractResponsesContent(obj)
			if content == "" {
				continue
			}
			if role == "" {
				role = "user"
			}
			messages = append(messages, OpenAIMessage{Role: role, Content: content})
		case "function_call", "function_call_output":
			content := extractResponsesContent(obj)
			if content == "" {
				if output, ok := obj["output"].(string); ok {
					content = output
				} else if arguments, ok := obj["arguments"].(string); ok {
					content = arguments
				}
			}
			if content != "" {
				messages = append(messages, OpenAIMessage{Role: "user", Content: content})
			}
		}
	}
	if len(messages) == 0 {
		if instructions, ok := raw["instructions"].(string); ok && instructions != "" {
			messages = append(messages, OpenAIMessage{Role: "user", Content: instructions})
		}
	}
	if len(messages) == 0 {
		return OpenAIChatCompletionRequest{}, fmt.Errorf("unable to map responses input to chat messages")
	}
	maxTokens := 512
	if value, ok := raw["max_output_tokens"].(float64); ok && int(value) > 0 {
		maxTokens = int(value)
	}
	return OpenAIChatCompletionRequest{
		Model:          model,
		Messages:       messages,
		Tools:          toAnySlice(raw["tools"]),
		ToolChoice:     raw["tool_choice"],
		ResponseFormat: raw["text"],
		MaxTokens:      &maxTokens,
		Stream:         stream,
	}, nil
}

func toAnySlice(value any) []any {
	if items, ok := value.([]any); ok {
		return items
	}
	return nil
}

func extractResponsesContent(obj map[string]any) string {
	if content, ok := obj["content"].(string); ok {
		return content
	}
	if inputText, ok := obj["text"].(string); ok {
		return inputText
	}
	if contentItems, ok := obj["content"].([]any); ok {
		parts := make([]string, 0)
		for _, item := range contentItems {
			part, ok := item.(map[string]any)
			if !ok {
				continue
			}
			if text, ok := part["text"].(string); ok && text != "" {
				parts = append(parts, text)
			}
			if inputText, ok := part["input_text"].(string); ok && inputText != "" {
				parts = append(parts, inputText)
			}
		}
		return strings.Join(parts, "\n")
	}
	return ""
}

func mapChatCompletionToResponses(raw map[string]any, chatResp map[string]any, note string) map[string]any {
	id, _ := chatResp["id"].(string)
	model, _ := raw["model"].(string)
	choices, _ := chatResp["choices"].([]any)
	text := ""
	var output []map[string]any
	if len(choices) > 0 {
		if first, ok := choices[0].(map[string]any); ok {
			if message, ok := first["message"].(map[string]any); ok {
				if content, ok := message["content"].(string); ok {
					text = content
				}
				if toolCalls, ok := message["tool_calls"].([]any); ok && len(toolCalls) > 0 {
					for _, item := range toolCalls {
						call, ok := item.(map[string]any)
						if !ok {
							continue
						}
						function, _ := call["function"].(map[string]any)
						output = append(output, map[string]any{
							"id":        call["id"],
							"type":      "function_call",
							"call_id":   call["id"],
							"name":      function["name"],
							"arguments": function["arguments"],
						})
					}
				}
			}
		}
	}
	outputText := text
	if outputText == "" {
		outputText = ""
	}
	if len(output) == 0 {
		output = []map[string]any{{
			"id":      fmt.Sprintf("msg_%d", time.Now().UnixNano()),
			"type":    "message",
			"role":    "assistant",
			"content": []map[string]any{{"type": "output_text", "text": outputText}},
		}}
	}
	return map[string]any{
		"id":          id,
		"object":      "response",
		"model":       model,
		"status":      "completed",
		"output":      output,
		"output_text": outputText,
		"usage":       chatResp["usage"],
		"metadata":    map[string]any{"fallback": "chat.completions", "note": note},
	}
}

func (s *GatewayService) ForwardAnthropicMessages(ctx context.Context, client *repository.GatewayClient, req AnthropicMessagesRequest) (map[string]any, int, error) {
	if req.Model == "" || len(req.Messages) == 0 || req.MaxTokens <= 0 {
		return nil, http.StatusBadRequest, fmt.Errorf("model, messages and max_tokens are required")
	}
	if req.Stream {
		return nil, http.StatusBadRequest, fmt.Errorf("stream is not supported yet")
	}

	route, err := s.repo.ResolveOpenAIModelRoute(ctx, req.Model)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	if route == nil {
		return nil, http.StatusNotFound, fmt.Errorf("virtual model not found")
	}
	allowed, err := s.repo.ClientCanAccessModel(ctx, client.ID, route.VirtualModelID)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	if !allowed {
		return nil, http.StatusForbidden, fmt.Errorf("client key is not allowed to access this model")
	}

	providerKey, secret, err := s.providerKeyService.SelectForRequest(ctx, route.ProviderID)
	if err != nil {
		return nil, http.StatusBadGateway, fmt.Errorf("no available provider key")
	}

	requestBody := map[string]any{
		"model":      route.UpstreamModelName,
		"max_tokens": req.MaxTokens,
		"messages":   req.Messages,
	}
	if req.System != "" {
		requestBody["system"] = req.System
	}
	if req.Temperature != nil {
		requestBody["temperature"] = *req.Temperature
	}

	body, err := json.Marshal(requestBody)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	requestURL := strings.TrimRight(route.BaseURL, "/") + "/v1/messages"
	requestURL = strings.ReplaceAll(requestURL, "/v1/v1", "/v1")
	upstreamReq, err := http.NewRequestWithContext(ctx, http.MethodPost, requestURL, bytes.NewReader(body))
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	upstreamReq.Header.Set("Content-Type", "application/json")
	upstreamReq.Header.Set("x-api-key", secret)
	upstreamReq.Header.Set("anthropic-version", "2023-06-01")

	resp, err := s.httpClient.Do(upstreamReq)
	if err != nil {
		_ = s.providerKeyService.ReportResult(ctx, providerKey.ID, false, err.Error())
		_, _ = s.routeService.RegisterFailure(ctx, route.ProviderID, route.VirtualModelID, err.Error())
		return nil, http.StatusBadGateway, err
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		_ = s.providerKeyService.ReportResult(ctx, providerKey.ID, false, err.Error())
		_, _ = s.routeService.RegisterFailure(ctx, route.ProviderID, route.VirtualModelID, err.Error())
		return nil, http.StatusBadGateway, err
	}

	var mapped map[string]any
	if err := json.Unmarshal(responseBody, &mapped); err != nil {
		_ = s.providerKeyService.ReportResult(ctx, providerKey.ID, false, err.Error())
		_, _ = s.routeService.RegisterFailure(ctx, route.ProviderID, route.VirtualModelID, err.Error())
		return nil, http.StatusBadGateway, err
	}

	if resp.StatusCode >= 400 {
		_ = s.providerKeyService.ReportResult(ctx, providerKey.ID, false, resp.Status)
		_, _ = s.routeService.RegisterFailure(ctx, route.ProviderID, route.VirtualModelID, resp.Status)
		return nil, resp.StatusCode, fmt.Errorf("anthropic upstream error: %d", resp.StatusCode)
	}

	_ = s.providerKeyService.ReportResult(ctx, providerKey.ID, true, "")
	_ = s.routeService.RegisterSuccess(ctx, route.ProviderID, route.VirtualModelID)
	return mapped, resp.StatusCode, nil
}

func (s *GatewayService) ForwardOpenAIChatCompletion(ctx context.Context, client *repository.GatewayClient, req OpenAIChatCompletionRequest) (map[string]any, int, error) {
	start := time.Now()
	if req.Model == "" || len(req.Messages) == 0 {
		return nil, http.StatusBadRequest, fmt.Errorf("model and messages are required")
	}

	route, err := s.repo.ResolveOpenAIModelRoute(ctx, req.Model)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	if route == nil {
		return nil, http.StatusNotFound, fmt.Errorf("virtual model not found")
	}

	responsePayload, statusCode, err := s.tryRoute(ctx, client, req, route, start)
	if err == nil {
		return responsePayload, statusCode, nil
	}

	if shouldFallback(statusCode, err) {
		fallbackRoute, fallbackErr := s.repo.ResolveOpenAIFallbackRoute(ctx, route.VirtualModelID, derefInt64(route.CurrentBindingID))
		if fallbackErr == nil && fallbackRoute != nil {
			fallbackPayload, fallbackStatus, fallbackRunErr := s.tryRoute(ctx, client, req, fallbackRoute, start)
			if fallbackRunErr == nil {
				if route.CurrentBindingID != nil {
					_ = s.routeService.Switch(ctx, route.VirtualModelID, derefInt64(fallbackRoute.CurrentBindingID), false, nil, "auto switch after fallback success", 0, "", "")
				}
				return fallbackPayload, fallbackStatus, nil
			}
		}
	}

	return nil, statusCode, err
}

func (s *GatewayService) tryRoute(ctx context.Context, client *repository.GatewayClient, req OpenAIChatCompletionRequest, route *repository.ModelRoute, start time.Time) (map[string]any, int, error) {
	open, err := s.routeService.IsCircuitOpen(ctx, route.ProviderID, route.VirtualModelID)
	if err != nil {
		return nil, http.StatusBadGateway, err
	}
	if open {
		return nil, http.StatusBadGateway, fmt.Errorf("provider circuit is open")
	}

	logger.Log.Debug().Int64("provider_id", route.ProviderID).Msg("selecting provider key")
	providerKey, secret, err := s.providerKeyService.SelectForRequest(ctx, route.ProviderID)
	if err != nil {
		logger.Log.Error().Err(err).Int64("provider_id", route.ProviderID).Msg("failed to select provider key")
		return nil, http.StatusBadGateway, fmt.Errorf("no available provider key")
	}

	if providerKey == nil {
		logger.Log.Error().Int64("provider_id", route.ProviderID).Msg("provider key is nil")
		return nil, http.StatusBadGateway, fmt.Errorf("no available provider key")
	}

	logger.Log.Debug().Int64("key_id", providerKey.ID).Str("secret_len", fmt.Sprintf("%d", len(secret))).Msg("provider key selected")

	responsePayload, statusCode, err := s.forwardToOpenAIProvider(ctx, route, providerKey, secret, req)
	latency := int(time.Since(start).Milliseconds())
	if err != nil {
		_, _ = s.routeService.RegisterFailure(ctx, route.ProviderID, route.VirtualModelID, err.Error())
		_ = s.providerKeyService.ReportResult(ctx, providerKey.ID, false, err.Error())
		s.logGatewayRequest(ctx, client, route, providerKey, req, nil, &statusCode, false, &latency, stringPtr(errorCodeForStatus(statusCode)), stringPtr(err.Error()))
		return nil, statusCode, err
	}

	_ = s.routeService.RegisterSuccess(ctx, route.ProviderID, route.VirtualModelID)
	_ = s.providerKeyService.ReportResult(ctx, providerKey.ID, true, "")
	s.logGatewayRequest(ctx, client, route, providerKey, req, responsePayload, &statusCode, true, &latency, nil, nil)
	return responsePayload, statusCode, nil
}

func (s *GatewayService) forwardToOpenAIProvider(ctx context.Context, route *repository.ModelRoute, providerKey *providerrepo.ProviderKey, secret string, req OpenAIChatCompletionRequest) (map[string]any, int, error) {
	upstreamPayload := map[string]any{
		"model":    route.UpstreamModelName,
		"messages": req.Messages,
		"stream":   false,
	}
	if req.Temperature != nil {
		upstreamPayload["temperature"] = *req.Temperature
	}
	if req.TopP != nil {
		upstreamPayload["top_p"] = *req.TopP
	}
	if req.MaxTokens != nil {
		upstreamPayload["max_tokens"] = *req.MaxTokens
	}
	if req.Stop != nil {
		upstreamPayload["stop"] = req.Stop
	}
	if req.PresencePenalty != nil {
		upstreamPayload["presence_penalty"] = *req.PresencePenalty
	}
	if req.FrequencyPenalty != nil {
		upstreamPayload["frequency_penalty"] = *req.FrequencyPenalty
	}
	if req.User != nil {
		upstreamPayload["user"] = *req.User
	}
	if len(req.Tools) > 0 {
		upstreamPayload["tools"] = req.Tools
	}
	if req.ToolChoice != nil {
		upstreamPayload["tool_choice"] = req.ToolChoice
	}
	if req.ParallelToolCalls != nil {
		upstreamPayload["parallel_tool_calls"] = req.ParallelToolCalls
	}
	if req.ResponseFormat != nil {
		upstreamPayload["response_format"] = req.ResponseFormat
	}

	body, err := json.Marshal(upstreamPayload)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	requestURL := strings.TrimRight(route.BaseURL, "/")
	requestURL = strings.ReplaceAll(requestURL, "/v1/v1", "/v1")
	requestURL = strings.ReplaceAll(requestURL, "/v1/", "/")
	if !strings.HasSuffix(requestURL, "/v1") {
		requestURL += "/v1"
	}
	requestURL += "/chat/completions"
	logger.Log.Debug().Str("url", requestURL).Str("auth_type", route.AuthType).Str("upstream_model", route.UpstreamModelName).Msg("forwarding to upstream")

	upstreamReq, err := http.NewRequestWithContext(ctx, http.MethodPost, requestURL, bytes.NewReader(body))
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	upstreamReq.Header.Set("Content-Type", "application/json")
	if route.AuthType == "x_api_key" {
		upstreamReq.Header.Set("x-api-key", secret)
		logger.Log.Debug().Str("auth", "x_api_key").Msg("using x-api-key auth")
	} else {
		upstreamReq.Header.Set("Authorization", "Bearer "+secret)
		logger.Log.Debug().Str("auth", "bearer").Msg("using bearer auth")
	}

	resp, err := s.httpClient.Do(upstreamReq)
	if err != nil {
		return nil, http.StatusBadGateway, err
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, http.StatusBadGateway, err
	}

	logger.Log.Debug().Str("status", fmt.Sprintf("%d", resp.StatusCode)).Str("body", string(responseBody)).Msg("upstream response")

	var mapped map[string]any
	if err := json.Unmarshal(responseBody, &mapped); err != nil {
		return nil, http.StatusBadGateway, err
	}

	if resp.StatusCode >= 400 {
		if message, ok := mapped["error"].(map[string]any); ok {
			if msg, ok := message["message"].(string); ok {
				return nil, resp.StatusCode, fmt.Errorf(msg)
			}
		}
		return nil, resp.StatusCode, fmt.Errorf("upstream error: %d", resp.StatusCode)
	}

	return mapped, resp.StatusCode, nil
}

func (s *GatewayService) logGatewayRequest(ctx context.Context, client *repository.GatewayClient, route *repository.ModelRoute, providerKey *providerrepo.ProviderKey, req OpenAIChatCompletionRequest, respBody map[string]any, statusCode *int, success bool, latency *int, errorCode, errorMessage *string) {
	requestSummary, _ := json.Marshal(map[string]any{
		"message_count": len(req.Messages),
		"stream":        req.Stream,
		"model":         req.Model,
	})
	var responseSummary json.RawMessage
	var promptTokens, completionTokens, totalTokens *int
	if respBody != nil {
		if usage, ok := respBody["usage"].(map[string]any); ok {
			if pt, ok := usage["prompt_tokens"].(float64); ok {
				ptInt := int(pt)
				promptTokens = &ptInt
			}
			if ct, ok := usage["completion_tokens"].(float64); ok {
				ctInt := int(ct)
				completionTokens = &ctInt
			}
			if tt, ok := usage["total_tokens"].(float64); ok {
				ttInt := int(tt)
				totalTokens = &ttInt
			}
		}
		responseSummary, _ = json.Marshal(map[string]any{
			"has_choices":       respBody["choices"] != nil,
			"status":            statusCode,
			"prompt_tokens":     promptTokens,
			"completion_tokens": completionTokens,
			"total_tokens":      totalTokens,
		})
	}
	clientID := client.ID
	virtualModelID := route.VirtualModelID
	virtualModelCode := route.ModelCode
	providerID := route.ProviderID
	providerKeyID := providerKey.ID
	upstreamModelName := route.UpstreamModelName
	if err := s.repo.InsertRequestLog(ctx, repository.RequestLogInput{
		TraceID:             appctx.RequestID(ctx),
		ProtocolType:        "openai",
		ClientID:            &clientID,
		VirtualModelID:      &virtualModelID,
		VirtualModelCode:    &virtualModelCode,
		ProviderID:          &providerID,
		ProviderKeyID:       &providerKeyID,
		UpstreamModelName:   &upstreamModelName,
		RequestSummaryJSON:  requestSummary,
		ResponseSummaryJSON: responseSummary,
		StatusCode:          statusCode,
		Success:             success,
		LatencyMS:           latency,
		PromptTokens:        promptTokens,
		CompletionTokens:    completionTokens,
		TotalTokens:         totalTokens,
		ErrorCode:           errorCode,
		ErrorMessage:        errorMessage,
	}); err != nil {
		logger.Log.Error().Err(err).Msg("failed to insert request log")
	}
}

func (s *GatewayService) logProxyRequest(ctx context.Context, client *repository.GatewayClient, route *repository.ModelRoute, providerKey *providerrepo.ProviderKey, method, path string, rawBody []byte, statusCode int, success bool, respBody []byte, errorCode, errorMessage *string) {
	var modelName string
	if route != nil {
		modelName = route.UpstreamModelName
	}
	requestSummary, _ := json.Marshal(map[string]any{
		"method": method,
		"path":   path,
		"model":  modelName,
	})
	var responseSummary json.RawMessage
	var promptTokens, completionTokens, totalTokens *int
	if respBody != nil {
		var mapped map[string]any
		if err := json.Unmarshal(respBody, &mapped); err == nil {
			if usage, ok := mapped["usage"].(map[string]any); ok {
				parseTokensFromUsage(usage, &promptTokens, &completionTokens, &totalTokens)
			}
		} else {
			if isSSEResponse(respBody) {
				parseTokensFromSSE(respBody, &promptTokens, &completionTokens, &totalTokens)
			} else {
				logger.Log.Warn().Err(err).Msg("failed to parse response body for token extraction")
			}
		}
		responseSummary, _ = json.Marshal(map[string]any{
			"body_size":         len(respBody),
			"status":            statusCode,
			"prompt_tokens":     promptTokens,
			"completion_tokens": completionTokens,
			"total_tokens":      totalTokens,
		})
	}
	var clientID, virtualModelID, providerID, providerKeyID int64
	var virtualModelCode, upstreamModelName string
	if client != nil {
		clientID = client.ID
	}
	if route != nil {
		virtualModelID = route.VirtualModelID
		virtualModelCode = route.ModelCode
		providerID = route.ProviderID
		upstreamModelName = route.UpstreamModelName
	}
	if providerKey != nil {
		providerKeyID = providerKey.ID
	}
	statusCodePtr := statusCode
	if err := s.repo.InsertRequestLog(ctx, repository.RequestLogInput{
		TraceID:             appctx.RequestID(ctx),
		ProtocolType:        "openai",
		ClientID:            &clientID,
		VirtualModelID:      &virtualModelID,
		VirtualModelCode:    &virtualModelCode,
		ProviderID:          &providerID,
		ProviderKeyID:       &providerKeyID,
		UpstreamModelName:   &upstreamModelName,
		RequestSummaryJSON:  requestSummary,
		ResponseSummaryJSON: responseSummary,
		StatusCode:          &statusCodePtr,
		Success:             success,
		PromptTokens:        promptTokens,
		CompletionTokens:    completionTokens,
		TotalTokens:         totalTokens,
		ErrorCode:           errorCode,
		ErrorMessage:        errorMessage,
	}); err != nil {
		logger.Log.Error().Err(err).Msg("failed to insert request log")
	}
}

func bearerOrRawKey(header string) string {
	if strings.HasPrefix(header, "Bearer ") {
		return strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
	}
	return strings.TrimSpace(header)
}

func shouldFallback(statusCode int, err error) bool {
	if err == nil {
		return false
	}
	if statusCode == http.StatusBadGateway || statusCode == http.StatusTooManyRequests || statusCode >= 500 {
		return true
	}
	return strings.Contains(strings.ToLower(err.Error()), "timeout") || strings.Contains(strings.ToLower(err.Error()), "circuit")
}

func errorCodeForStatus(statusCode int) string {
	switch {
	case statusCode == http.StatusTooManyRequests:
		return "GW429001"
	case statusCode >= 500:
		return "GW502001"
	default:
		return "GW500001"
	}
}

func derefInt64(value *int64) int64 {
	if value == nil {
		return 0
	}
	return *value
}

func stringPtr(value string) *string { return &value }

func parseTokensFromUsage(usage map[string]any, promptTokens, completionTokens, totalTokens **int) {
	if pt, ok := usage["prompt_tokens"].(float64); ok {
		ptInt := int(pt)
		*promptTokens = &ptInt
	}
	if ct, ok := usage["completion_tokens"].(float64); ok {
		ctInt := int(ct)
		*completionTokens = &ctInt
	}
	if tt, ok := usage["total_tokens"].(float64); ok {
		ttInt := int(tt)
		*totalTokens = &ttInt
	}
}

func isSSEResponse(body []byte) bool {
	return len(body) > 5 && string(body[:5]) == "data:"
}

func parseTokensFromSSE(body []byte, promptTokens, completionTokens, totalTokens **int) {
	lines := strings.Split(string(body), "\n")
	var lastUsage map[string]any
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		data := strings.TrimPrefix(line, "data:")
		data = strings.TrimSpace(data)
		if data == "" || data == "[DONE]" {
			continue
		}
		var mapped map[string]any
		if err := json.Unmarshal([]byte(data), &mapped); err == nil {
			if usage, ok := mapped["usage"].(map[string]any); ok {
				lastUsage = usage
			}
		}
	}
	if lastUsage != nil {
		parseTokensFromUsage(lastUsage, promptTokens, completionTokens, totalTokens)
	}
}
