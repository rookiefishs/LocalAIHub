package service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"errors"
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
	"localaihub/localaihub_go/internal/pkg/netx"
)

type UpstreamError struct {
	StatusCode int
	Message    string
}

func (e *UpstreamError) Error() string {
	return fmt.Sprintf("upstream error [%d]: %s", e.StatusCode, e.Message)
}

func (e *UpstreamError) Unwrap() error {
	return errors.New(e.Message)
}

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

type GeminiGenerateContentRequest struct {
	Contents          []map[string]any `json:"contents"`
	SystemInstruction any              `json:"system_instruction,omitempty"`
	GenerationConfig  any              `json:"generation_config,omitempty"`
	Tools             any              `json:"tools,omitempty"`
	ToolConfig        any              `json:"tool_config,omitempty"`
	SafetySettings    any              `json:"safety_settings,omitempty"`
	CachedContent     string           `json:"cached_content,omitempty"`
}

type GatewayService struct {
	repo               *repository.GatewayRepository
	providerKeyService *providerservice.ProviderKeyService
	routeService       *routeservice.RouteService
	httpClient         *http.Client
	audit              interface {
		Log(ctx context.Context, action, targetType string, targetID *int64, details map[string]any, ip, userAgent string)
	}
}

func NewGatewayService(repo *repository.GatewayRepository, providerKeyService *providerservice.ProviderKeyService, routeService *routeservice.RouteService, audit interface {
	Log(ctx context.Context, action, targetType string, targetID *int64, details map[string]any, ip, userAgent string)
}, allowInsecureTLS bool) *GatewayService {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: allowInsecureTLS},
		MaxIdleConns:    10,
		IdleConnTimeout: 90 * time.Second,
	}
	return &GatewayService{
		repo:               repo,
		providerKeyService: providerKeyService,
		routeService:       routeService,
		httpClient:         &http.Client{Timeout: 120 * time.Second, Transport: tr},
		audit:              audit,
	}
}

func (s *GatewayService) AuthenticateClient(ctx context.Context, authHeader string) (*repository.GatewayClient, error) {
	return s.authenticateClientWithRequest(ctx, authHeader, "", "")
}

func (s *GatewayService) AuthenticateClientWithRequest(ctx context.Context, authHeader, ip, userAgent string) (*repository.GatewayClient, error) {
	return s.authenticateClientWithRequest(ctx, authHeader, ip, userAgent)
}

func (s *GatewayService) authenticateClientWithRequest(ctx context.Context, authHeader, ip, userAgent string) (*repository.GatewayClient, error) {
	apiKey := bearerOrRawKey(authHeader)
	if apiKey == "" {
		s.logAuthFailure(ctx, nil, "missing_api_key", ip, userAgent)
		return nil, fmt.Errorf("missing api key (缺少API Key)")
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
	if client == nil {
		s.logAuthFailure(ctx, nil, "invalid_client_key", ip, userAgent)
		return nil, fmt.Errorf("invalid client key (无效的API Key)")
	}
	if client.Status != "active" {
		s.logAuthFailure(ctx, client, "client_disabled", ip, userAgent)
		return nil, fmt.Errorf("invalid client key (无效的API Key)")
	}
	if client.ExpiresAt != nil && client.ExpiresAt.Before(time.Now().UTC()) {
		s.logAuthFailure(ctx, client, "client_key_expired", ip, userAgent)
		return nil, fmt.Errorf("client key expired (API Key已过期)")
	}
	if client.QuotaDisabledAt != nil {
		s.logAuthFailure(ctx, client, "quota_disabled", ip, userAgent)
		return nil, fmt.Errorf("api key disabled due to quota exceeded (API Key因配额超限被禁用)")
	}
	allowed, reason, err := s.repo.CheckAndIncrementUsage(ctx, client.ID, client.DailyRequestLimit, client.MonthlyRequestLimit, client.DailyTokenLimit, client.MonthlyTokenLimit)
	if err != nil {
		logger.Log.Error().Err(err).Int64("client_id", client.ID).Msg("failed to check and increment usage")
		return nil, fmt.Errorf("internal error")
	}
	if !allowed {
		s.logAuthFailure(ctx, client, "quota_exceeded", ip, userAgent)
		if disErr := s.repo.DisableClient(ctx, client.ID); disErr != nil {
			logger.Log.Error().Err(disErr).Msg("failed to disable client due to quota exceeded")
		}
		return nil, fmt.Errorf("quota exceeded: %s", reason)
	}
	if err := s.repo.TouchClientLastUsed(ctx, client.ID); err != nil {
		logger.Log.Error().Err(err).Int64("client_id", client.ID).Msg("failed to touch client last used")
	}
	return client, nil
}

func (s *GatewayService) AuthenticateClientForTest(ctx context.Context, authHeader string) (*repository.GatewayClient, error) {
	apiKey := bearerOrRawKey(authHeader)
	if apiKey == "" {
		return nil, fmt.Errorf("missing api key (缺少API Key)")
	}
	hash := sha256.Sum256([]byte(apiKey))
	hashString := hex.EncodeToString(hash[:])
	client, err := s.repo.GetClientByHash(ctx, hashString)
	if err != nil {
		logger.Log.Error().Err(err).Str("hash", hashString).Msg("failed to get client by hash for test")
		return nil, fmt.Errorf("invalid client key (无效的API Key)")
	}
	if client == nil {
		return nil, fmt.Errorf("invalid client key (无效的API Key)")
	}
	if client.Status != "active" {
		return nil, fmt.Errorf("invalid client key (API Key已被禁用)")
	}
	if client.ExpiresAt != nil && client.ExpiresAt.Before(time.Now().UTC()) {
		return nil, fmt.Errorf("client key expired (API Key已过期)")
	}
	if err := s.repo.TouchClientLastUsed(ctx, client.ID); err != nil {
		logger.Log.Error().Err(err).Int64("client_id", client.ID).Msg("failed to touch client last used for test")
	}
	return client, nil
}

func (s *GatewayService) UpdateClientStatusAfterTest(ctx context.Context, clientID int64, success bool) error {
	if success {
		return s.repo.SetClientStatus(ctx, clientID, "active", true)
	}
	return s.repo.SetClientStatus(ctx, clientID, "disabled", false)
}

func (s *GatewayService) logAuthFailure(ctx context.Context, client *repository.GatewayClient, reason, ip, userAgent string) {
	if s.audit == nil {
		return
	}
	var targetID *int64
	payload := map[string]any{"reason": reason}
	if client != nil {
		targetID = &client.ID
		payload["client_id"] = client.ID
		payload["client_name"] = client.Name
		payload["key_prefix"] = client.KeyPrefix
	}
	s.audit.Log(ctx, "gateway.auth_failed", "api_client", targetID, payload, ip, userAgent)
}

func ClientIPFromRequest(r *http.Request) string {
	return netx.ClientIP(r)
}

func (s *GatewayService) checkAndEnforceQuota(ctx context.Context, client *repository.GatewayClient) error {
	exceeded := false
	reason := ""

	if client.DailyRequestLimit != nil && client.CurrentDailyRequests >= *client.DailyRequestLimit {
		exceeded = true
		reason = "daily request limit exceeded"
	} else if client.MonthlyRequestLimit != nil && client.CurrentMonthlyRequests >= *client.MonthlyRequestLimit {
		exceeded = true
		reason = "monthly request limit exceeded"
	} else if client.DailyTokenLimit != nil && client.CurrentDailyTokens >= *client.DailyTokenLimit {
		exceeded = true
		reason = "daily token limit exceeded"
	} else if client.MonthlyTokenLimit != nil && client.CurrentMonthlyTokens >= *client.MonthlyTokenLimit {
		exceeded = true
		reason = "monthly token limit exceeded"
	}

	if exceeded {
		logger.Log.Warn().Int64("client_id", client.ID).Str("reason", reason).Msg("quota exceeded, disabling client")
		if err := s.repo.DisableClient(ctx, client.ID); err != nil {
			logger.Log.Error().Err(err).Msg("failed to disable client due to quota exceeded")
		}
		return fmt.Errorf("quota exceeded: %s", reason)
	}
	return nil
}

func (s *GatewayService) IncrementClientUsage(ctx context.Context, clientID int64, tokens int) error {
	return s.repo.IncrementClientUsage(ctx, clientID, tokens)
}

func (s *GatewayService) ListOpenAIModels(ctx context.Context) ([]map[string]any, error) {
	return s.repo.ListVisibleModels(ctx)
}

func (s *GatewayService) ListGeminiModels(ctx context.Context) ([]map[string]any, error) {
	return s.repo.ListVisibleModels(ctx)
}

type ProxyHTTPResult struct {
	StatusCode  int
	Headers     http.Header
	Body        []byte
	StreamBody  io.ReadCloser
	IsStream    bool
	AfterStream func([]byte)
}

type proxyAttemptResult struct {
	result      *ProxyHTTPResult
	retryable   bool
	route       *repository.ModelRoute
	providerKey *providerrepo.ProviderKey
	errorBody   []byte
	errorMsg    string
	statusCode  int
	method      string
	path        string
	rawBody     []byte
	latency     int
}

type routeAttemptOutcome struct {
	statusCode  int
	err         error
	providerKey *providerrepo.ProviderKey
	retryable   bool
}

func shouldFallbackByMessage(message string) bool {
	message = strings.ToLower(message)
	return strings.Contains(message, "timeout") || strings.Contains(message, "circuit") || strings.Contains(message, "connection") || strings.Contains(message, "reset") || strings.Contains(message, "eof") || strings.Contains(message, "no available provider key") || strings.Contains(message, "invalid api key") || strings.Contains(message, "incorrect api key") || strings.Contains(message, "expired api key") || strings.Contains(message, "api key")
}

func shouldTryNextProviderKey(statusCode int, err error) bool {
	if err == nil {
		return false
	}
	if statusCode == http.StatusUnauthorized || statusCode == http.StatusForbidden || statusCode == http.StatusTooManyRequests || statusCode == http.StatusBadGateway || statusCode >= 500 {
		return true
	}
	return shouldFallbackByMessage(err.Error())
}

func (s *GatewayService) shouldPersistAutoSwitch(ctx context.Context, virtualModelID int64) bool {
	state, err := s.routeService.Get(ctx, virtualModelID)
	if err != nil || state == nil {
		return true
	}
	if state.ManualLocked {
		return false
	}
	if state.LockUntil != nil && state.LockUntil.After(time.Now().UTC()) {
		return false
	}
	return true
}

func (s *GatewayService) executeRouteCandidateChain(ctx context.Context, route *repository.ModelRoute, runner func(candidateRoute *repository.ModelRoute) routeAttemptOutcome, switchReason string) routeAttemptOutcome {
	attemptedBindingIDs := make([]int64, 0, 4)
	if route.CurrentBindingID != nil {
		attemptedBindingIDs = append(attemptedBindingIDs, *route.CurrentBindingID)
	}

	outcome := runner(route)
	if outcome.err == nil {
		return outcome
	}
	if !shouldFallback(outcome.statusCode, outcome.err) {
		return outcome
	}

	fallbackRoutes, fallbackErr := s.repo.ListOpenAIFallbackRoutes(ctx, route.VirtualModelID, attemptedBindingIDs)
	if fallbackErr != nil {
		return routeAttemptOutcome{statusCode: http.StatusInternalServerError, err: fallbackErr}
	}
	for _, fallbackRoute := range fallbackRoutes {
		fallbackRoute := fallbackRoute
		if fallbackRoute.CurrentBindingID != nil {
			attemptedBindingIDs = append(attemptedBindingIDs, *fallbackRoute.CurrentBindingID)
		}
		fallbackOutcome := runner(&fallbackRoute)
		if fallbackOutcome.err == nil {
			if route.CurrentBindingID != nil && fallbackRoute.CurrentBindingID != nil && *route.CurrentBindingID != *fallbackRoute.CurrentBindingID && s.shouldPersistAutoSwitch(ctx, route.VirtualModelID) {
				if switchErr := s.routeService.AutoSwitchCurrentBinding(ctx, route.VirtualModelID, *route.CurrentBindingID, *fallbackRoute.CurrentBindingID, switchReason); switchErr != nil {
					logger.Log.Error().Err(switchErr).Int64("virtual_model_id", route.VirtualModelID).Int64("from_binding_id", *route.CurrentBindingID).Int64("to_binding_id", *fallbackRoute.CurrentBindingID).Msg("failed to auto switch route after fallback success")
				}
			}
			return fallbackOutcome
		}
		outcome = fallbackOutcome
		if !shouldFallback(fallbackOutcome.statusCode, fallbackOutcome.err) {
			break
		}
	}
	return outcome
}

	func (s *GatewayService) ProxyOpenAIRequest(ctx context.Context, client *repository.GatewayClient, method, path, rawQuery string, rawBody []byte, incomingHeaders http.Header) (*ProxyHTTPResult, error) {
	route, err := s.resolveOpenAIRouteForProxy(ctx, client, path, rawBody)
	if err != nil {
		return nil, err
	}

	attemptedBindingIDs := make([]int64, 0, 4)
	if route.CurrentBindingID != nil {
		attemptedBindingIDs = append(attemptedBindingIDs, *route.CurrentBindingID)
	}

	attempt, err := s.executeProxyAttempt(ctx, client, route, method, path, rawQuery, rawBody, incomingHeaders)
	if err != nil {
		return nil, err
	}
	if attempt.result != nil {
		return attempt.result, nil
	}
	if !attempt.retryable {
		return nil, buildProxyAttemptError(attempt)
	}

	fallbackRoutes, err := s.repo.ListOpenAIFallbackRoutes(ctx, route.VirtualModelID, attemptedBindingIDs)
	if err != nil {
		return nil, err
	}
	for _, fallbackRoute := range fallbackRoutes {
		fallbackRoute := fallbackRoute
		if fallbackRoute.CurrentBindingID != nil {
			attemptedBindingIDs = append(attemptedBindingIDs, *fallbackRoute.CurrentBindingID)
		}
		fallbackAttempt, runErr := s.executeProxyAttempt(ctx, client, &fallbackRoute, method, path, rawQuery, rawBody, incomingHeaders)
		if runErr != nil {
			return nil, runErr
		}
		if fallbackAttempt.result != nil {
			if route.CurrentBindingID != nil && fallbackRoute.CurrentBindingID != nil && *route.CurrentBindingID != *fallbackRoute.CurrentBindingID && s.shouldPersistAutoSwitch(ctx, route.VirtualModelID) {
				if switchErr := s.routeService.AutoSwitchCurrentBinding(ctx, route.VirtualModelID, *route.CurrentBindingID, *fallbackRoute.CurrentBindingID, "auto switch after proxy fallback success"); switchErr != nil {
					logger.Log.Error().Err(switchErr).Int64("virtual_model_id", route.VirtualModelID).Int64("from_binding_id", *route.CurrentBindingID).Int64("to_binding_id", *fallbackRoute.CurrentBindingID).Msg("failed to auto switch route after proxy fallback success")
				}
			}
			if fallbackAttempt.providerKey != nil && fallbackAttempt.latency > 0 {
				s.logProxyRequest(ctx, client, &fallbackRoute, fallbackAttempt.providerKey, fallbackAttempt.method, fallbackAttempt.path, fallbackAttempt.rawBody, fallbackAttempt.result.StatusCode, true, fallbackAttempt.result.Body, &fallbackAttempt.latency, nil, nil)
			}
			return fallbackAttempt.result, nil
		}
		attempt = fallbackAttempt
		if !fallbackAttempt.retryable {
			break
		}
	}

	return nil, buildProxyAttemptError(attempt)
}

func (s *GatewayService) executeProxyAttempt(ctx context.Context, client *repository.GatewayClient, route *repository.ModelRoute, method, path, rawQuery string, rawBody []byte, incomingHeaders http.Header) (*proxyAttemptResult, error) {
	start := time.Now()
	candidates, err := s.providerKeyService.ListCandidatesForRequest(ctx, route.ProviderID)
	if err != nil {
		msg := "no available provider key"
		if _, registerErr := s.routeService.RegisterFailure(ctx, route.ProviderID, route.VirtualModelID, msg); registerErr != nil {
			logger.Log.Error().Err(registerErr).Int64("provider_id", route.ProviderID).Int64("virtual_model_id", route.VirtualModelID).Msg("failed to register proxy route failure for missing provider key")
		}
		return &proxyAttemptResult{retryable: true, route: route, statusCode: http.StatusBadGateway, errorMsg: msg}, nil
	}
	if len(candidates) == 0 {
		msg := "no available provider key"
		if _, registerErr := s.routeService.RegisterFailure(ctx, route.ProviderID, route.VirtualModelID, msg); registerErr != nil {
			logger.Log.Error().Err(registerErr).Int64("provider_id", route.ProviderID).Int64("virtual_model_id", route.VirtualModelID).Msg("failed to register proxy route failure for missing provider key")
		}
		return &proxyAttemptResult{retryable: true, route: route, statusCode: http.StatusBadGateway, errorMsg: msg}, nil
	}

	requestURL := buildOpenAIProxyURL(route.BaseURL, path, rawQuery)
	logger.Log.Debug().Str("method", method).Str("path", path).Str("target_url", requestURL).Str("auth_type", route.AuthType).Str("upstream_model", route.UpstreamModelName).Msg("transparent proxy forwarding")

	rewrittenBody := replaceModelInBody(rawBody, route.UpstreamModelName)
	var lastAttempt *proxyAttemptResult
	for _, candidate := range candidates {
		providerKey := candidate.Key
		secret := candidate.Secret
		if providerKey == nil || secret == "" {
			continue
		}

		upstreamReq, reqErr := http.NewRequestWithContext(ctx, method, requestURL, bytes.NewReader(rewrittenBody))
		if reqErr != nil {
			return nil, reqErr
		}
		copyHeadersForProxy(upstreamReq.Header, incomingHeaders)
		if route.AuthType == "x_api_key" {
			upstreamReq.Header.Set("x-api-key", secret)
			upstreamReq.Header.Del("Authorization")
		} else {
			upstreamReq.Header.Set("Authorization", "Bearer "+secret)
			upstreamReq.Header.Del("x-api-key")
		}

		resp, reqErr := s.httpClient.Do(upstreamReq)
		if reqErr != nil {
			if reportErr := s.providerKeyService.ReportResult(ctx, providerKey.ID, false, reqErr.Error()); reportErr != nil {
				logger.Log.Error().Err(reportErr).Int64("provider_key_id", providerKey.ID).Msg("failed to report provider key result for proxy request error")
			}
			lastAttempt = &proxyAttemptResult{retryable: isRetryableProxyError(http.StatusBadGateway, reqErr), route: route, providerKey: providerKey, statusCode: http.StatusBadGateway, errorMsg: reqErr.Error(), method: method, path: path, rawBody: rewrittenBody}
			if shouldTryNextProviderKey(http.StatusBadGateway, reqErr) {
				continue
			}
			if _, registerErr := s.routeService.RegisterFailure(ctx, route.ProviderID, route.VirtualModelID, reqErr.Error()); registerErr != nil {
				logger.Log.Error().Err(registerErr).Int64("provider_id", route.ProviderID).Int64("virtual_model_id", route.VirtualModelID).Msg("failed to register proxy route failure")
			}
			return lastAttempt, nil
		}

		if isStreamingResponse(resp.Header) {
			streamResult, streamErr := s.handleStreamingProxyAttempt(ctx, client, route, providerKey, method, path, rewrittenBody, resp, start)
			if streamErr != nil {
				return nil, streamErr
			}
			if streamResult != nil && streamResult.result != nil {
				return streamResult, nil
			}
			lastAttempt = streamResult
			if streamResult != nil && streamResult.retryable && shouldTryNextProviderKey(streamResult.statusCode, buildProxyAttemptError(streamResult)) {
				continue
			}
			return streamResult, nil
		}

		body, readErr := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if readErr != nil {
			if reportErr := s.providerKeyService.ReportResult(ctx, providerKey.ID, false, readErr.Error()); reportErr != nil {
				logger.Log.Error().Err(reportErr).Int64("provider_key_id", providerKey.ID).Msg("failed to report provider key result after proxy read error")
			}
			lastAttempt = &proxyAttemptResult{retryable: true, route: route, providerKey: providerKey, statusCode: http.StatusBadGateway, errorMsg: readErr.Error(), method: method, path: path, rawBody: rewrittenBody}
			if shouldTryNextProviderKey(http.StatusBadGateway, readErr) {
				continue
			}
			if _, registerErr := s.routeService.RegisterFailure(ctx, route.ProviderID, route.VirtualModelID, readErr.Error()); registerErr != nil {
				logger.Log.Error().Err(registerErr).Int64("provider_id", route.ProviderID).Int64("virtual_model_id", route.VirtualModelID).Msg("failed to register proxy route failure after read error")
			}
			return lastAttempt, nil
		}

		if resp.StatusCode >= 400 {
			upstreamErrMsg := parseUpstreamError(body)
			if reportErr := s.providerKeyService.ReportResult(ctx, providerKey.ID, false, resp.Status); reportErr != nil {
				logger.Log.Error().Err(reportErr).Int64("provider_key_id", providerKey.ID).Msg("failed to report provider key result for proxy upstream failure")
			}
			lastAttempt = &proxyAttemptResult{retryable: isRetryableProxyError(resp.StatusCode, &UpstreamError{StatusCode: resp.StatusCode, Message: upstreamErrMsg}), route: route, providerKey: providerKey, errorBody: body, errorMsg: upstreamErrMsg, statusCode: resp.StatusCode, method: method, path: path, rawBody: rewrittenBody}
			latency := int(time.Since(start).Milliseconds())
			s.logProxyRequest(ctx, client, route, providerKey, method, path, rewrittenBody, resp.StatusCode, false, body, &latency, nil, &upstreamErrMsg)
			if shouldTryNextProviderKey(resp.StatusCode, &UpstreamError{StatusCode: resp.StatusCode, Message: upstreamErrMsg}) {
				continue
			}
			if _, registerErr := s.routeService.RegisterFailure(ctx, route.ProviderID, route.VirtualModelID, resp.Status); registerErr != nil {
				logger.Log.Error().Err(registerErr).Int64("provider_id", route.ProviderID).Int64("virtual_model_id", route.VirtualModelID).Msg("failed to register proxy route failure for upstream status")
			}
			return lastAttempt, nil
		}

		if reportErr := s.providerKeyService.ReportResult(ctx, providerKey.ID, true, ""); reportErr != nil {
			logger.Log.Error().Err(reportErr).Int64("provider_key_id", providerKey.ID).Msg("failed to report provider key result for proxy success")
		}
		if registerErr := s.routeService.RegisterSuccess(ctx, route.ProviderID, route.VirtualModelID); registerErr != nil {
			logger.Log.Error().Err(registerErr).Int64("provider_id", route.ProviderID).Int64("virtual_model_id", route.VirtualModelID).Msg("failed to register proxy route success")
		}
		latency := int(time.Since(start).Milliseconds())
		s.logProxyRequest(ctx, client, route, providerKey, method, path, rewrittenBody, resp.StatusCode, true, body, &latency, nil, nil)
		return &proxyAttemptResult{
			result:  &ProxyHTTPResult{StatusCode: resp.StatusCode, Headers: resp.Header.Clone(), Body: body},
			method:  method,
			path:    path,
			rawBody: rewrittenBody,
			latency: latency,
			providerKey: providerKey,
		}, nil
	}

	if lastAttempt == nil {
		lastAttempt = &proxyAttemptResult{retryable: true, route: route, statusCode: http.StatusBadGateway, errorMsg: "no available provider key"}
	}
	if _, registerErr := s.routeService.RegisterFailure(ctx, route.ProviderID, route.VirtualModelID, lastAttempt.errorMsg); registerErr != nil {
		logger.Log.Error().Err(registerErr).Int64("provider_id", route.ProviderID).Int64("virtual_model_id", route.VirtualModelID).Msg("failed to register proxy route final failure")
	}
	return lastAttempt, nil
}

func (s *GatewayService) handleStreamingProxyAttempt(ctx context.Context, client *repository.GatewayClient, route *repository.ModelRoute, providerKey *providerrepo.ProviderKey, method, path string, rawBody []byte, resp *http.Response, start time.Time) (*proxyAttemptResult, error) {
	buffer := make([]byte, 32*1024)
	n, readErr := resp.Body.Read(buffer)
	if readErr != nil && !errors.Is(readErr, io.EOF) {
		_ = resp.Body.Close()
		if reportErr := s.providerKeyService.ReportResult(ctx, providerKey.ID, false, readErr.Error()); reportErr != nil {
			logger.Log.Error().Err(reportErr).Int64("provider_key_id", providerKey.ID).Msg("failed to report provider key result after proxy stream prelude error")
		}
		if _, registerErr := s.routeService.RegisterFailure(ctx, route.ProviderID, route.VirtualModelID, readErr.Error()); registerErr != nil {
			logger.Log.Error().Err(registerErr).Int64("provider_id", route.ProviderID).Int64("virtual_model_id", route.VirtualModelID).Msg("failed to register proxy route failure after stream prelude error")
		}
		return &proxyAttemptResult{retryable: true, route: route, providerKey: providerKey, statusCode: http.StatusBadGateway, errorMsg: readErr.Error()}, nil
	}
	prefix := append([]byte(nil), buffer[:n]...)
	if resp.StatusCode >= 400 {
		remainingBody, err := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			remainingBody = nil
		}
		body := append(prefix, remainingBody...)
		upstreamErrMsg := parseUpstreamError(body)
		latency := int(time.Since(start).Milliseconds())
		if reportErr := s.providerKeyService.ReportResult(ctx, providerKey.ID, false, resp.Status); reportErr != nil {
			logger.Log.Error().Err(reportErr).Int64("provider_key_id", providerKey.ID).Msg("failed to report provider key result for proxy stream upstream failure")
		}
		if _, registerErr := s.routeService.RegisterFailure(ctx, route.ProviderID, route.VirtualModelID, resp.Status); registerErr != nil {
			logger.Log.Error().Err(registerErr).Int64("provider_id", route.ProviderID).Int64("virtual_model_id", route.VirtualModelID).Msg("failed to register proxy route failure for stream upstream status")
		}
		s.logProxyRequest(ctx, client, route, providerKey, method, path, rawBody, resp.StatusCode, false, body, &latency, nil, &upstreamErrMsg)
		return &proxyAttemptResult{retryable: isRetryableProxyError(resp.StatusCode, &UpstreamError{StatusCode: resp.StatusCode, Message: upstreamErrMsg}), route: route, providerKey: providerKey, errorBody: body, errorMsg: upstreamErrMsg, statusCode: resp.StatusCode}, nil
	}
	if reportErr := s.providerKeyService.ReportResult(ctx, providerKey.ID, true, ""); reportErr != nil {
		logger.Log.Error().Err(reportErr).Int64("provider_key_id", providerKey.ID).Msg("failed to report provider key success for proxy stream")
	}
	if registerErr := s.routeService.RegisterSuccess(ctx, route.ProviderID, route.VirtualModelID); registerErr != nil {
		logger.Log.Error().Err(registerErr).Int64("provider_id", route.ProviderID).Int64("virtual_model_id", route.VirtualModelID).Msg("failed to register proxy route success for stream")
	}
	mergedBody := io.NopCloser(io.MultiReader(bytes.NewReader(prefix), resp.Body))
	latencyVal := int(time.Since(start).Milliseconds())
	return &proxyAttemptResult{
		result: &ProxyHTTPResult{
			StatusCode: resp.StatusCode,
			Headers:    resp.Header.Clone(),
			StreamBody: mergedBody,
			IsStream:   true,
			AfterStream: func(streamBody []byte) {
				latency := int(time.Since(start).Milliseconds())
				s.logProxyRequest(ctx, client, route, providerKey, method, path, rawBody, resp.StatusCode, true, streamBody, &latency, nil, nil)
			},
		},
		method:  method,
		path:    path,
		rawBody: rawBody,
		latency: latencyVal,
	}, nil
}

func buildOpenAIProxyURL(baseURL, path, rawQuery string) string {
	requestURL := strings.TrimRight(baseURL, "/")
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
	return requestURL
}

func isRetryableProxyError(statusCode int, err error) bool {
	if err == nil {
		return false
	}
	if statusCode == http.StatusTooManyRequests || statusCode == http.StatusBadGateway || statusCode >= 500 {
		return true
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "timeout") || strings.Contains(message, "connection") || strings.Contains(message, "circuit") || strings.Contains(message, "reset") || strings.Contains(message, "eof")
}

func buildProxyAttemptError(attempt *proxyAttemptResult) error {
	if attempt == nil {
		return fmt.Errorf("proxy request failed")
	}
	if attempt.statusCode >= 400 {
		return &UpstreamError{StatusCode: attempt.statusCode, Message: attempt.errorMsg}
	}
	if attempt.errorMsg != "" {
		return fmt.Errorf(attempt.errorMsg)
	}
	return fmt.Errorf("proxy request failed")
}

func reportProviderKeySelectionFailure(err error) error {
	return err
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

	result, statusCode, err := s.executeOpenAICandidateChain(ctx, route, func(candidateRoute *repository.ModelRoute) (map[string]any, int, error) {
		return s.tryResponsesRoute(ctx, client, raw, candidateRoute)
	}, "auto switch after responses fallback success")
	if err == nil {
		return result, stream, statusCode, nil
	}

	fallbackResult, fallbackErr := s.forwardResponsesViaChatFallback(ctx, client, raw, err.Error())
	if fallbackErr == nil {
		return fallbackResult, stream, http.StatusOK, nil
	}
	return nil, stream, statusCode, err
}

func (s *GatewayService) executeOpenAICandidateChain(ctx context.Context, route *repository.ModelRoute, runner func(candidateRoute *repository.ModelRoute) (map[string]any, int, error), switchReason string) (map[string]any, int, error) {
	attempedResult := s.executeRouteCandidateChain(ctx, route, func(candidateRoute *repository.ModelRoute) routeAttemptOutcome {
		_, statusCode, err := runner(candidateRoute)
		if err == nil {
			return routeAttemptOutcome{statusCode: statusCode, err: nil}
		}
		return routeAttemptOutcome{statusCode: statusCode, err: err, retryable: shouldFallback(statusCode, err)}
	}, switchReason)
	if attempedResult.err == nil {
		result, statusCode, err := runner(route)
		if err == nil {
			return result, statusCode, nil
		}
		return nil, statusCode, err
	}

	attemptedBindingIDs := make([]int64, 0, 4)
	if route.CurrentBindingID != nil {
		attemptedBindingIDs = append(attemptedBindingIDs, *route.CurrentBindingID)
	}

	result, statusCode, err := runner(route)
	if err == nil {
		return result, statusCode, nil
	}
	if !shouldFallback(statusCode, err) {
		return nil, statusCode, err
	}

	fallbackRoutes, fallbackErr := s.repo.ListOpenAIFallbackRoutes(ctx, route.VirtualModelID, attemptedBindingIDs)
	if fallbackErr != nil {
		return nil, http.StatusInternalServerError, fallbackErr
	}
	for _, fallbackRoute := range fallbackRoutes {
		fallbackRoute := fallbackRoute
		if fallbackRoute.CurrentBindingID != nil {
			attemptedBindingIDs = append(attemptedBindingIDs, *fallbackRoute.CurrentBindingID)
		}
		fallbackResult, fallbackStatus, fallbackRunErr := runner(&fallbackRoute)
		if fallbackRunErr == nil {
			if route.CurrentBindingID != nil && fallbackRoute.CurrentBindingID != nil && *route.CurrentBindingID != *fallbackRoute.CurrentBindingID && s.shouldPersistAutoSwitch(ctx, route.VirtualModelID) {
				if switchErr := s.routeService.AutoSwitchCurrentBinding(ctx, route.VirtualModelID, *route.CurrentBindingID, *fallbackRoute.CurrentBindingID, switchReason); switchErr != nil {
					logger.Log.Error().Err(switchErr).Int64("virtual_model_id", route.VirtualModelID).Int64("from_binding_id", *route.CurrentBindingID).Int64("to_binding_id", *fallbackRoute.CurrentBindingID).Msg("failed to auto switch route after fallback success")
				}
			}
			return fallbackResult, fallbackStatus, nil
		}
		statusCode = fallbackStatus
		err = fallbackRunErr
		if !shouldFallback(fallbackStatus, fallbackRunErr) {
			break
		}
	}

	return nil, statusCode, err
}

func (s *GatewayService) tryResponsesRoute(ctx context.Context, client *repository.GatewayClient, raw map[string]any, route *repository.ModelRoute) (map[string]any, int, error) {
	candidates, err := s.providerKeyService.ListCandidatesForRequest(ctx, route.ProviderID)
	if err != nil || len(candidates) == 0 {
		return nil, http.StatusBadGateway, fmt.Errorf("no available provider key")
	}
	repaired := repairResponsesPayload(raw)
	repaired["model"] = route.UpstreamModelName
	repaired["stream"] = false
	body, err := json.Marshal(repaired)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	requestURL := strings.TrimRight(route.BaseURL, "/")
	requestURL = strings.ReplaceAll(requestURL, "/v1/v1", "/v1")
	requestURL = strings.ReplaceAll(requestURL, "/v1/", "/")
	if !strings.HasSuffix(requestURL, "/v1") {
		requestURL += "/v1"
	}
	requestURL += "/responses"

	var lastErr error
	var lastStatusCode = http.StatusBadGateway
	for _, candidate := range candidates {
		providerKey := candidate.Key
		secret := candidate.Secret
		if providerKey == nil || secret == "" {
			continue
		}
		upstreamReq, err := http.NewRequestWithContext(ctx, http.MethodPost, requestURL, bytes.NewReader(body))
		if err != nil {
			return nil, http.StatusInternalServerError, err
		}
		upstreamReq.Header.Set("Content-Type", "application/json")
		if route.AuthType == "x_api_key" {
			upstreamReq.Header.Set("x-api-key", secret)
		} else {
			upstreamReq.Header.Set("Authorization", "Bearer "+secret)
		}
		resp, err := s.httpClient.Do(upstreamReq)
		if err != nil {
			lastErr = err
			lastStatusCode = http.StatusBadGateway
			if reportErr := s.providerKeyService.ReportResult(ctx, providerKey.ID, false, err.Error()); reportErr != nil {
				logger.Log.Error().Err(reportErr).Int64("provider_key_id", providerKey.ID).Msg("failed to report responses provider key failure")
			}
			if shouldTryNextProviderKey(lastStatusCode, err) {
				continue
			}
			break
		}
		responseBody, err := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			lastErr = err
			lastStatusCode = http.StatusBadGateway
			if reportErr := s.providerKeyService.ReportResult(ctx, providerKey.ID, false, err.Error()); reportErr != nil {
				logger.Log.Error().Err(reportErr).Int64("provider_key_id", providerKey.ID).Msg("failed to report responses read failure")
			}
			if shouldTryNextProviderKey(lastStatusCode, err) {
				continue
			}
			break
		}
		var mapped map[string]any
		if err := json.Unmarshal(responseBody, &mapped); err != nil {
			lastErr = err
			lastStatusCode = http.StatusBadGateway
			if reportErr := s.providerKeyService.ReportResult(ctx, providerKey.ID, false, err.Error()); reportErr != nil {
				logger.Log.Error().Err(reportErr).Int64("provider_key_id", providerKey.ID).Msg("failed to report responses parse failure")
			}
			if shouldTryNextProviderKey(lastStatusCode, err) {
				continue
			}
			break
		}
		if resp.StatusCode >= 400 {
			upstreamErrMsg := parseUpstreamError(responseBody)
			lastErr = &UpstreamError{StatusCode: resp.StatusCode, Message: upstreamErrMsg}
			lastStatusCode = resp.StatusCode
			if reportErr := s.providerKeyService.ReportResult(ctx, providerKey.ID, false, resp.Status); reportErr != nil {
				logger.Log.Error().Err(reportErr).Int64("provider_key_id", providerKey.ID).Msg("failed to report responses upstream error")
			}
			logger.Log.Error().
				Int("status_code", resp.StatusCode).
				Str("upstream_error", upstreamErrMsg).
				Str("model_code", route.ModelCode).
				Str("provider_url", route.BaseURL).
				Msg("openai responses upstream request failed")
			if shouldTryNextProviderKey(lastStatusCode, lastErr) {
				continue
			}
			break
		}
		if reportErr := s.providerKeyService.ReportResult(ctx, providerKey.ID, true, ""); reportErr != nil {
			logger.Log.Error().Err(reportErr).Int64("provider_key_id", providerKey.ID).Msg("failed to report responses success")
		}
		if registerErr := s.routeService.RegisterSuccess(ctx, route.ProviderID, route.VirtualModelID); registerErr != nil {
			logger.Log.Error().Err(registerErr).Int64("provider_id", route.ProviderID).Int64("virtual_model_id", route.VirtualModelID).Msg("failed to register responses route success")
		}
		return mapped, resp.StatusCode, nil
	}

	if lastErr == nil {
		lastErr = fmt.Errorf("no available provider key")
	}
	if _, registerErr := s.routeService.RegisterFailure(ctx, route.ProviderID, route.VirtualModelID, lastErr.Error()); registerErr != nil {
		logger.Log.Error().Err(registerErr).Int64("provider_id", route.ProviderID).Int64("virtual_model_id", route.VirtualModelID).Msg("failed to register responses route final failure")
	}
	return nil, lastStatusCode, lastErr
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

func (s *GatewayService) ForwardAnthropicMessages(ctx context.Context, client *repository.GatewayClient, req AnthropicMessagesRequest) (*ProxyHTTPResult, int, error) {
	start := time.Now()
	if req.Model == "" || len(req.Messages) == 0 || req.MaxTokens <= 0 {
		return nil, http.StatusBadRequest, fmt.Errorf("model, messages and max_tokens are required")
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
	if req.Stream {
		requestBody["stream"] = true
	}

	body, err := json.Marshal(requestBody)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	requestURL := strings.TrimRight(route.BaseURL, "/") + "/v1/messages"
	requestURL = strings.ReplaceAll(requestURL, "/v1/v1", "/v1")
	candidates, err := s.providerKeyService.ListCandidatesForRequest(ctx, route.ProviderID)
	if err != nil || len(candidates) == 0 {
		return nil, http.StatusBadGateway, fmt.Errorf("no available provider key")
	}

	var lastErr error
	lastStatusCode := http.StatusBadGateway
	for _, candidate := range candidates {
		providerKey := candidate.Key
		secret := candidate.Secret
		if providerKey == nil || secret == "" {
			continue
		}
		upstreamReq, err := http.NewRequestWithContext(ctx, http.MethodPost, requestURL, bytes.NewReader(body))
		if err != nil {
			return nil, http.StatusInternalServerError, err
		}
		upstreamReq.Header.Set("Content-Type", "application/json")
		upstreamReq.Header.Set("x-api-key", secret)
		upstreamReq.Header.Set("anthropic-version", "2023-06-01")

		resp, err := s.httpClient.Do(upstreamReq)
		if err != nil {
			lastErr = err
			lastStatusCode = http.StatusBadGateway
			if reportErr := s.providerKeyService.ReportResult(ctx, providerKey.ID, false, err.Error()); reportErr != nil {
				logger.Log.Error().Err(reportErr).Int64("provider_key_id", providerKey.ID).Msg("failed to report anthropic provider key failure")
			}
			if shouldTryNextProviderKey(lastStatusCode, err) {
				continue
			}
			break
		}
		if req.Stream && isStreamingResponse(resp.Header) {
			if reportErr := s.providerKeyService.ReportResult(ctx, providerKey.ID, true, ""); reportErr != nil {
				logger.Log.Error().Err(reportErr).Int64("provider_key_id", providerKey.ID).Msg("failed to report anthropic stream success")
			}
			if registerErr := s.routeService.RegisterSuccess(ctx, route.ProviderID, route.VirtualModelID); registerErr != nil {
				logger.Log.Error().Err(registerErr).Int64("provider_id", route.ProviderID).Int64("virtual_model_id", route.VirtualModelID).Msg("failed to register anthropic stream route success")
			}
			return &ProxyHTTPResult{
				StatusCode: resp.StatusCode,
				Headers:    resp.Header.Clone(),
				StreamBody: resp.Body,
				IsStream:   true,
				AfterStream: func(streamBody []byte) {
					latency := int(time.Since(start).Milliseconds())
					s.logProxyRequestWithProtocol(ctx, client, route, providerKey, http.MethodPost, requestURL, body, resp.StatusCode, true, streamBody, &latency, nil, nil, "anthropic")
				},
			}, resp.StatusCode, nil
		}

		responseBody, err := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			lastErr = err
			lastStatusCode = http.StatusBadGateway
			if reportErr := s.providerKeyService.ReportResult(ctx, providerKey.ID, false, err.Error()); reportErr != nil {
				logger.Log.Error().Err(reportErr).Int64("provider_key_id", providerKey.ID).Msg("failed to report anthropic read failure")
			}
			if shouldTryNextProviderKey(lastStatusCode, err) {
				continue
			}
			break
		}

		var mapped map[string]any
		if err := json.Unmarshal(responseBody, &mapped); err != nil {
			lastErr = err
			lastStatusCode = http.StatusBadGateway
			if reportErr := s.providerKeyService.ReportResult(ctx, providerKey.ID, false, err.Error()); reportErr != nil {
				logger.Log.Error().Err(reportErr).Int64("provider_key_id", providerKey.ID).Msg("failed to report anthropic parse failure")
			}
			if shouldTryNextProviderKey(lastStatusCode, err) {
				continue
			}
			break
		}

		if resp.StatusCode >= 400 {
			upstreamErrMsg := parseUpstreamError(responseBody)
			lastErr = &UpstreamError{StatusCode: resp.StatusCode, Message: upstreamErrMsg}
			lastStatusCode = resp.StatusCode
			if reportErr := s.providerKeyService.ReportResult(ctx, providerKey.ID, false, resp.Status); reportErr != nil {
				logger.Log.Error().Err(reportErr).Int64("provider_key_id", providerKey.ID).Msg("failed to report anthropic upstream error")
			}
			logger.Log.Error().
				Int("status_code", resp.StatusCode).
				Str("upstream_error", upstreamErrMsg).
				Str("model_code", route.ModelCode).
				Str("provider_url", route.BaseURL).
				Msg("anthropic upstream request failed")
			if shouldTryNextProviderKey(lastStatusCode, lastErr) {
				continue
			}
			break
		}

		if reportErr := s.providerKeyService.ReportResult(ctx, providerKey.ID, true, ""); reportErr != nil {
			logger.Log.Error().Err(reportErr).Int64("provider_key_id", providerKey.ID).Msg("failed to report anthropic success")
		}
		if registerErr := s.routeService.RegisterSuccess(ctx, route.ProviderID, route.VirtualModelID); registerErr != nil {
			logger.Log.Error().Err(registerErr).Int64("provider_id", route.ProviderID).Int64("virtual_model_id", route.VirtualModelID).Msg("failed to register anthropic route success")
		}
		return &ProxyHTTPResult{StatusCode: resp.StatusCode, Headers: resp.Header.Clone(), Body: responseBody}, resp.StatusCode, nil
	}

	if lastErr == nil {
		lastErr = fmt.Errorf("no available provider key")
	}
	return nil, lastStatusCode, lastErr
}

func (s *GatewayService) ForwardGeminiGenerateContent(ctx context.Context, client *repository.GatewayClient, modelCode string, req GeminiGenerateContentRequest, stream bool) (*ProxyHTTPResult, int, error) {
	start := time.Now()
	if modelCode == "" || len(req.Contents) == 0 {
		return nil, http.StatusBadRequest, fmt.Errorf("model and contents are required")
	}

	route, err := s.repo.ResolveOpenAIModelRoute(ctx, modelCode)
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
	action := "generateContent"
	if stream {
		action = "streamGenerateContent"
	}
	payload := map[string]any{
		"contents": req.Contents,
	}
	if req.SystemInstruction != nil {
		payload["system_instruction"] = req.SystemInstruction
	}
	if req.GenerationConfig != nil {
		payload["generation_config"] = req.GenerationConfig
	}
	if req.Tools != nil {
		payload["tools"] = req.Tools
	}
	if req.ToolConfig != nil {
		payload["tool_config"] = req.ToolConfig
	}
	if req.SafetySettings != nil {
		payload["safety_settings"] = req.SafetySettings
	}
	if req.CachedContent != "" {
		payload["cached_content"] = req.CachedContent
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	baseURL := strings.TrimRight(route.BaseURL, "/")
	baseURL = strings.ReplaceAll(baseURL, "/v1beta/v1beta", "/v1beta")
	baseURL = strings.ReplaceAll(baseURL, "/v1beta/", "/")
	baseURL = strings.ReplaceAll(baseURL, "/v1/v1beta", "/v1beta")
	if !strings.HasSuffix(baseURL, "/v1beta") {
		baseURL += "/v1beta"
	}
	requestURL := fmt.Sprintf("%s/models/%s:%s", baseURL, route.UpstreamModelName, action)
	candidates, err := s.providerKeyService.ListCandidatesForRequest(ctx, route.ProviderID)
	if err != nil || len(candidates) == 0 {
		return nil, http.StatusBadGateway, fmt.Errorf("no available provider key")
	}

	var lastErr error
	lastStatusCode := http.StatusBadGateway
	for _, candidate := range candidates {
		providerKey := candidate.Key
		secret := candidate.Secret
		if providerKey == nil || secret == "" {
			continue
		}
		upstreamReq, err := http.NewRequestWithContext(ctx, http.MethodPost, requestURL, bytes.NewReader(body))
		if err != nil {
			return nil, http.StatusInternalServerError, err
		}
		upstreamReq.Header.Set("Content-Type", "application/json")
		upstreamReq.Header.Set("x-goog-api-key", secret)
		resp, err := s.httpClient.Do(upstreamReq)
		if err != nil {
			lastErr = err
			lastStatusCode = http.StatusBadGateway
			if reportErr := s.providerKeyService.ReportResult(ctx, providerKey.ID, false, err.Error()); reportErr != nil {
				logger.Log.Error().Err(reportErr).Int64("provider_key_id", providerKey.ID).Msg("failed to report gemini provider key failure")
			}
			if shouldTryNextProviderKey(lastStatusCode, err) {
				continue
			}
			break
		}
		responseBody, err := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			lastErr = err
			lastStatusCode = http.StatusBadGateway
			if reportErr := s.providerKeyService.ReportResult(ctx, providerKey.ID, false, err.Error()); reportErr != nil {
				logger.Log.Error().Err(reportErr).Int64("provider_key_id", providerKey.ID).Msg("failed to report gemini read failure")
			}
			if shouldTryNextProviderKey(lastStatusCode, err) {
				continue
			}
			break
		}
		if resp.StatusCode >= 400 {
			upstreamErrMsg := parseUpstreamError(responseBody)
			lastErr = &UpstreamError{StatusCode: resp.StatusCode, Message: upstreamErrMsg}
			lastStatusCode = resp.StatusCode
			if reportErr := s.providerKeyService.ReportResult(ctx, providerKey.ID, false, resp.Status); reportErr != nil {
				logger.Log.Error().Err(reportErr).Int64("provider_key_id", providerKey.ID).Msg("failed to report gemini upstream error")
			}
			logger.Log.Error().
				Int("status_code", resp.StatusCode).
				Str("upstream_error", upstreamErrMsg).
				Str("model_code", route.ModelCode).
				Str("provider_url", route.BaseURL).
				Msg("gemini upstream request failed")
			latency := int(time.Since(start).Milliseconds())
			s.logProxyRequestWithProtocol(ctx, client, route, providerKey, http.MethodPost, requestURL, body, resp.StatusCode, false, responseBody, &latency, stringPtr(errorCodeForStatus(resp.StatusCode)), &upstreamErrMsg, "gemini")
			if shouldTryNextProviderKey(lastStatusCode, lastErr) {
				continue
			}
			break
		}
		if reportErr := s.providerKeyService.ReportResult(ctx, providerKey.ID, true, ""); reportErr != nil {
			logger.Log.Error().Err(reportErr).Int64("provider_key_id", providerKey.ID).Msg("failed to report gemini success")
		}
		if registerErr := s.routeService.RegisterSuccess(ctx, route.ProviderID, route.VirtualModelID); registerErr != nil {
			logger.Log.Error().Err(registerErr).Int64("provider_id", route.ProviderID).Int64("virtual_model_id", route.VirtualModelID).Msg("failed to register gemini route success")
		}
		latency := int(time.Since(start).Milliseconds())
		s.logProxyRequestWithProtocol(ctx, client, route, providerKey, http.MethodPost, requestURL, body, resp.StatusCode, true, responseBody, &latency, nil, nil, "gemini")
		return &ProxyHTTPResult{StatusCode: resp.StatusCode, Headers: resp.Header.Clone(), Body: responseBody}, resp.StatusCode, nil
	}

	if lastErr == nil {
		lastErr = fmt.Errorf("no available provider key")
	}
	if _, registerErr := s.routeService.RegisterFailure(ctx, route.ProviderID, route.VirtualModelID, lastErr.Error()); registerErr != nil {
		logger.Log.Error().Err(registerErr).Int64("provider_id", route.ProviderID).Int64("virtual_model_id", route.VirtualModelID).Msg("failed to register gemini route final failure")
	}
	return nil, lastStatusCode, lastErr
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
	allowed, err := s.repo.ClientCanAccessModel(ctx, client.ID, route.VirtualModelID)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	if !allowed {
		return nil, http.StatusForbidden, fmt.Errorf("client key is not allowed to access this model")
	}

	return s.executeOpenAICandidateChain(ctx, route, func(candidateRoute *repository.ModelRoute) (map[string]any, int, error) {
		return s.tryRoute(ctx, client, req, candidateRoute, start)
	}, "auto switch after chat fallback success")
}

func (s *GatewayService) tryRoute(ctx context.Context, client *repository.GatewayClient, req OpenAIChatCompletionRequest, route *repository.ModelRoute, start time.Time) (map[string]any, int, error) {
	open, err := s.routeService.IsCircuitOpen(ctx, route.ProviderID, route.VirtualModelID)
	if err != nil {
		return nil, http.StatusBadGateway, err
	}
	if open {
		return nil, http.StatusBadGateway, fmt.Errorf("provider circuit is open")
	}

	logger.Log.Debug().Int64("provider_id", route.ProviderID).Msg("selecting provider keys")
	candidates, err := s.providerKeyService.ListCandidatesForRequest(ctx, route.ProviderID)
	if err != nil || len(candidates) == 0 {
		logger.Log.Error().Err(err).Int64("provider_id", route.ProviderID).Msg("failed to list provider key candidates")
		return nil, http.StatusBadGateway, fmt.Errorf("no available provider key")
	}

	var lastErr error
	lastStatusCode := http.StatusBadGateway
	for _, candidate := range candidates {
		providerKey := candidate.Key
		secret := candidate.Secret
		if providerKey == nil || secret == "" {
			continue
		}

		logger.Log.Debug().Int64("key_id", providerKey.ID).Str("secret_len", fmt.Sprintf("%d", len(secret))).Msg("provider key candidate selected")
		responsePayload, statusCode, runErr := s.forwardToOpenAIProvider(ctx, route, providerKey, secret, req)
		latency := int(time.Since(start).Milliseconds())
		if runErr != nil {
			lastErr = runErr
			lastStatusCode = statusCode
			if reportErr := s.providerKeyService.ReportResult(ctx, providerKey.ID, false, runErr.Error()); reportErr != nil {
				logger.Log.Error().Err(reportErr).Int64("provider_key_id", providerKey.ID).Msg("failed to report provider key failure in tryRoute")
			}
			s.logGatewayRequest(ctx, client, route, providerKey, req, nil, &statusCode, false, &latency, stringPtr(errorCodeForStatus(statusCode)), stringPtr(runErr.Error()))
			if shouldTryNextProviderKey(statusCode, runErr) {
				continue
			}
			break
		}

		if registerErr := s.routeService.RegisterSuccess(ctx, route.ProviderID, route.VirtualModelID); registerErr != nil {
			logger.Log.Error().Err(registerErr).Int64("provider_id", route.ProviderID).Int64("virtual_model_id", route.VirtualModelID).Msg("failed to register route success in tryRoute")
		}
		if reportErr := s.providerKeyService.ReportResult(ctx, providerKey.ID, true, ""); reportErr != nil {
			logger.Log.Error().Err(reportErr).Int64("provider_key_id", providerKey.ID).Msg("failed to report provider key success in tryRoute")
		}
		s.logGatewayRequest(ctx, client, route, providerKey, req, responsePayload, &statusCode, true, &latency, nil, nil)
		return responsePayload, statusCode, nil
	}

	if lastErr == nil {
		lastErr = fmt.Errorf("no available provider key")
	}
	if _, registerErr := s.routeService.RegisterFailure(ctx, route.ProviderID, route.VirtualModelID, lastErr.Error()); registerErr != nil {
		logger.Log.Error().Err(registerErr).Int64("provider_id", route.ProviderID).Int64("virtual_model_id", route.VirtualModelID).Msg("failed to register route failure in tryRoute")
	}
	return nil, lastStatusCode, lastErr
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
	var requestMessages []string
	for _, msg := range req.Messages {
		if msg.Role == "user" || msg.Role == "system" {
			content := summarizeMessageContent(msg.Content)
			if content == "" {
				continue
			}
			requestMessages = append(requestMessages, fmt.Sprintf("[%s] %s", msg.Role, truncatePreview(content, 500)))
		}
	}
	requestSummary, _ := json.Marshal(map[string]any{
		"messages":        requestMessages,
		"stream":          req.Stream,
		"model":           req.Model,
		"message_count":   len(req.Messages),
		"content_preview": summarizeMessagesPreview(requestMessages),
	})
	var responseSummary json.RawMessage
	var promptTokens, completionTokens, totalTokens *int
	if respBody != nil {
		extractTokensFromMappedResponse("openai", respBody, &promptTokens, &completionTokens, &totalTokens)
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
	requestedModel := req.Model
	providerID := route.ProviderID
	providerKeyID := providerKey.ID
	upstreamModelName := route.UpstreamModelName
	bindingID := route.CurrentBindingID
	if err := s.repo.InsertRequestLog(ctx, repository.RequestLogInput{
		TraceID:             appctx.RequestID(ctx),
		ProtocolType:        "openai",
		ClientID:            &clientID,
		VirtualModelID:      &virtualModelID,
		VirtualModelCode:    &virtualModelCode,
		RequestedModel:      &requestedModel,
		BindingID:           bindingID,
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
	if client != nil && success && totalTokens != nil && *totalTokens > 0 {
		if err := s.repo.IncrementClientUsage(ctx, client.ID, *totalTokens); err != nil {
			logger.Log.Error().Err(err).Msg("failed to increment client usage")
		}
	}
}

func (s *GatewayService) logProxyRequest(ctx context.Context, client *repository.GatewayClient, route *repository.ModelRoute, providerKey *providerrepo.ProviderKey, method, path string, rawBody []byte, statusCode int, success bool, respBody []byte, latency *int, errorCode, errorMessage *string) {
	s.logProxyRequestWithProtocol(ctx, client, route, providerKey, method, path, rawBody, statusCode, success, respBody, latency, errorCode, errorMessage, inferProxyProtocol(path))
}

func (s *GatewayService) logProxyRequestWithProtocol(ctx context.Context, client *repository.GatewayClient, route *repository.ModelRoute, providerKey *providerrepo.ProviderKey, method, path string, rawBody []byte, statusCode int, success bool, respBody []byte, latency *int, errorCode, errorMessage *string, protocol string) {
	var modelName string
	if route != nil {
		modelName = route.UpstreamModelName
	}
	requestSummaryData := buildProxyRequestSummary(protocol, method, path, modelName, rawBody)
	requestSummary, _ := json.Marshal(requestSummaryData)
	requestedModel, _ := requestSummaryData["model"].(string)
	var responseSummary json.RawMessage
	var promptTokens, completionTokens, totalTokens *int
	if respBody != nil {
		var mapped map[string]any
		if err := json.Unmarshal(respBody, &mapped); err == nil {
			extractTokensFromMappedResponse(protocol, mapped, &promptTokens, &completionTokens, &totalTokens)
		} else {
			if isSSEResponse(respBody) {
				parseTokensFromSSE(respBody, protocol, &promptTokens, &completionTokens, &totalTokens)
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
	var virtualModelCode string
	var bindingID *int64
	var upstreamModelName string
	if client != nil {
		clientID = client.ID
	}
	if route != nil {
		virtualModelID = route.VirtualModelID
		virtualModelCode = route.ModelCode
		bindingID = route.CurrentBindingID
		providerID = route.ProviderID
		upstreamModelName = route.UpstreamModelName
	}
	if providerKey != nil {
		providerKeyID = providerKey.ID
	}
	statusCodePtr := statusCode
	var requestedModelPtr *string
	if requestedModel != "" {
		requestedModelPtr = &requestedModel
	}
	if err := s.repo.InsertRequestLog(ctx, repository.RequestLogInput{
		TraceID:             appctx.RequestID(ctx),
		ProtocolType:        protocol,
		ClientID:            &clientID,
		VirtualModelID:      &virtualModelID,
		VirtualModelCode:    &virtualModelCode,
		RequestedModel:      requestedModelPtr,
		BindingID:           bindingID,
		ProviderID:          &providerID,
		ProviderKeyID:       &providerKeyID,
		UpstreamModelName:   &upstreamModelName,
		RequestSummaryJSON:  requestSummary,
		ResponseSummaryJSON: responseSummary,
		StatusCode:          &statusCodePtr,
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
	if client != nil && success && totalTokens != nil && *totalTokens > 0 {
		if err := s.repo.IncrementClientUsage(ctx, client.ID, *totalTokens); err != nil {
			logger.Log.Error().Err(err).Msg("failed to increment client usage")
		}
	}
}

func bearerOrRawKey(header string) string {
	if strings.HasPrefix(header, "Bearer ") {
		return strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
	}
	return strings.TrimSpace(header)
}

func buildProxyRequestSummary(protocol, method, path, fallbackModel string, rawBody []byte) map[string]any {
	summary := map[string]any{
		"method":    method,
		"path":      path,
		"model":     fallbackModel,
		"body_size": len(rawBody),
	}
	if len(rawBody) == 0 {
		return summary
	}
	switch protocol {
	case "anthropic":
		applyAnthropicSummary(summary, rawBody)
	case "gemini":
		applyGeminiSummary(summary, rawBody)
	default:
		applyOpenAISummary(summary, rawBody)
	}
	if preview, _ := summary["content_preview"].(string); preview == "" {
		if messages, ok := summary["messages"].([]string); ok {
			summary["content_preview"] = summarizeMessagesPreview(messages)
		}
	}
	if model, _ := summary["model"].(string); model == "" {
		summary["model"] = fallbackModel
	}
	return summary
}

func applyOpenAISummary(summary map[string]any, rawBody []byte) {
	var req OpenAIChatCompletionRequest
	if err := json.Unmarshal(rawBody, &req); err != nil {
		return
	}
	if req.Model != "" {
		summary["model"] = req.Model
	}
	summary["stream"] = req.Stream
	summary["message_count"] = len(req.Messages)
	messages := make([]string, 0, len(req.Messages))
	for _, msg := range req.Messages {
		if msg.Role != "user" && msg.Role != "system" && msg.Role != "assistant" {
			continue
		}
		content := summarizeMessageContent(msg.Content)
		if content == "" {
			continue
		}
		messages = append(messages, fmt.Sprintf("[%s] %s", msg.Role, truncatePreview(content, 500)))
	}
	if len(messages) > 0 {
		summary["messages"] = messages
		summary["content_preview"] = summarizeMessagesPreview(messages)
	}
}

func applyAnthropicSummary(summary map[string]any, rawBody []byte) {
	var req AnthropicMessagesRequest
	if err := json.Unmarshal(rawBody, &req); err != nil {
		return
	}
	if req.Model != "" {
		summary["model"] = req.Model
	}
	summary["stream"] = req.Stream
	messageCount := len(req.Messages)
	if strings.TrimSpace(req.System) != "" {
		messageCount++
	}
	summary["message_count"] = messageCount
	messages := make([]string, 0, messageCount)
	if system := truncatePreview(strings.TrimSpace(req.System), 500); system != "" {
		messages = append(messages, fmt.Sprintf("[system] %s", system))
	}
	for _, msg := range req.Messages {
		content := summarizeMessageContent(msg.Content)
		if content == "" {
			continue
		}
		messages = append(messages, fmt.Sprintf("[%s] %s", msg.Role, truncatePreview(content, 500)))
	}
	if len(messages) > 0 {
		summary["messages"] = messages
		summary["content_preview"] = summarizeMessagesPreview(messages)
	}
}

func applyGeminiSummary(summary map[string]any, rawBody []byte) {
	var req GeminiGenerateContentRequest
	if err := json.Unmarshal(rawBody, &req); err != nil {
		return
	}
	messages := make([]string, 0, len(req.Contents)+1)
	if system := summarizeMessageContent(req.SystemInstruction); system != "" {
		messages = append(messages, fmt.Sprintf("[system] %s", truncatePreview(system, 500)))
	}
	for _, content := range req.Contents {
		role, _ := content["role"].(string)
		if role == "" {
			role = "user"
		}
		text := summarizeMessageContent(content)
		if text == "" {
			continue
		}
		messages = append(messages, fmt.Sprintf("[%s] %s", role, truncatePreview(text, 500)))
	}
	if len(messages) > 0 {
		summary["messages"] = messages
		summary["message_count"] = len(messages)
		summary["content_preview"] = summarizeMessagesPreview(messages)
	}
}

func summarizeMessageContent(content any) string {
	switch v := content.(type) {
	case string:
		return strings.TrimSpace(v)
	case []any:
		parts := make([]string, 0, len(v))
		for _, item := range v {
			if text := summarizeMessageContent(item); text != "" {
				parts = append(parts, text)
			}
		}
		return strings.Join(parts, " ")
	case map[string]any:
		if text, _ := v["text"].(string); strings.TrimSpace(text) != "" {
			return strings.TrimSpace(text)
		}
		if input, _ := v["input_text"].(string); strings.TrimSpace(input) != "" {
			return strings.TrimSpace(input)
		}
		if parts, ok := v["parts"].([]any); ok {
			return summarizeMessageContent(parts)
		}
		if content, ok := v["content"]; ok {
			return summarizeMessageContent(content)
		}
		return ""
	default:
		return ""
	}
}

func summarizeMessagesPreview(messages []string) string {
	if len(messages) == 0 {
		return ""
	}
	joined := strings.Join(messages, " | ")
	return truncatePreview(joined, 300)
}

func truncatePreview(value string, limit int) string {
	value = strings.TrimSpace(value)
	if value == "" || limit <= 0 {
		return ""
	}
	runes := []rune(value)
	if len(runes) <= limit {
		return value
	}
	return string(runes[:limit]) + "..."
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

func extractTokensFromMappedResponse(protocol string, mapped map[string]any, promptTokens, completionTokens, totalTokens **int) {
	if usage, ok := mapped["usage"].(map[string]any); ok {
		if protocol == "anthropic" {
			parseAnthropicTokensFromUsage(usage, promptTokens, completionTokens, totalTokens)
		} else {
			parseCompatibleTokensFromUsage(protocol, usage, promptTokens, completionTokens, totalTokens)
		}
		return
	}
	if usage, ok := mapped["usageMetadata"].(map[string]any); ok {
		parseGeminiTokensFromUsage(usage, promptTokens, completionTokens, totalTokens)
	}
}

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

func parseCompatibleTokensFromUsage(protocol string, usage map[string]any, promptTokens, completionTokens, totalTokens **int) {
	if protocol == "anthropic" {
		parseAnthropicTokensFromUsage(usage, promptTokens, completionTokens, totalTokens)
		return
	}
	parseTokensFromUsage(usage, promptTokens, completionTokens, totalTokens)
	if *promptTokens != nil || *completionTokens != nil || *totalTokens != nil {
		return
	}
	parseAnthropicTokensFromUsage(usage, promptTokens, completionTokens, totalTokens)
}

func parseGeminiTokensFromUsage(usage map[string]any, promptTokens, completionTokens, totalTokens **int) {
	if pt, ok := usage["promptTokenCount"].(float64); ok {
		ptInt := int(pt)
		*promptTokens = &ptInt
	}
	if ct, ok := usage["candidatesTokenCount"].(float64); ok {
		ctInt := int(ct)
		*completionTokens = &ctInt
	}
	if tt, ok := usage["totalTokenCount"].(float64); ok {
		ttInt := int(tt)
		*totalTokens = &ttInt
	}
}

func parseAnthropicTokensFromUsage(usage map[string]any, promptTokens, completionTokens, totalTokens **int) {
	total := 0
	if pt, ok := usage["input_tokens"].(float64); ok {
		ptInt := int(pt)
		*promptTokens = &ptInt
		total += ptInt
	}
	if ct, ok := usage["output_tokens"].(float64); ok {
		ctInt := int(ct)
		*completionTokens = &ctInt
		total += ctInt
	}
	if total > 0 {
		totalInt := total
		*totalTokens = &totalInt
	}
}

func inferProxyProtocol(path string) string {
	if strings.Contains(path, "/anthropic/") || strings.HasSuffix(path, "/messages") {
		return "anthropic"
	}
	if strings.Contains(path, "/gemini/") {
		return "gemini"
	}
	return "openai"
}

func isSSEResponse(body []byte) bool {
	trimmed := bytes.TrimSpace(body)
	if len(trimmed) == 0 {
		return false
	}
	return bytes.HasPrefix(trimmed, []byte("data:")) || bytes.HasPrefix(trimmed, []byte("event:")) || bytes.Contains(trimmed, []byte("\ndata:")) || bytes.Contains(trimmed, []byte("\nevent:"))
}

func isStreamingResponse(headers http.Header) bool {
	contentType := strings.ToLower(headers.Get("Content-Type"))
	return strings.Contains(contentType, "text/event-stream")
}

func extractTokensFromSSEEvent(protocol string, mapped map[string]any, promptTokens, completionTokens, totalTokens **int) bool {
	extractTokensFromMappedResponse(protocol, mapped, promptTokens, completionTokens, totalTokens)
	if *promptTokens != nil || *completionTokens != nil || *totalTokens != nil {
		return true
	}
	if protocol == "anthropic" {
		extractTokensFromAnthropicStreamEvent(mapped, promptTokens, completionTokens, totalTokens)
	} else {
		extractTokensFromResponseEvent(mapped, promptTokens, completionTokens, totalTokens)
	}
	return *promptTokens != nil || *completionTokens != nil || *totalTokens != nil
}

func extractTokensFromResponseEvent(mapped map[string]any, promptTokens, completionTokens, totalTokens **int) {
	response, ok := mapped["response"].(map[string]any)
	if !ok {
		return
	}
	extractTokensFromMappedResponse("openai", response, promptTokens, completionTokens, totalTokens)
}

func extractTokensFromAnthropicStreamEvent(mapped map[string]any, promptTokens, completionTokens, totalTokens **int) {
	message, ok := mapped["message"].(map[string]any)
	if !ok {
		return
	}
	extractTokensFromMappedResponse("anthropic", message, promptTokens, completionTokens, totalTokens)
}

func parseTokensFromSSE(body []byte, protocol string, promptTokens, completionTokens, totalTokens **int) {
	lines := strings.Split(string(body), "\n")
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
			extractTokensFromSSEEvent(protocol, mapped, promptTokens, completionTokens, totalTokens)
		}
	}
}

func replaceModelInBody(body []byte, upstreamModelName string) []byte {
	if len(body) == 0 || upstreamModelName == "" {
		return body
	}
	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		return body
	}
	if model, ok := payload["model"].(string); ok && model != "" {
		payload["model"] = upstreamModelName
		newBody, err := json.Marshal(payload)
		if err != nil {
			return body
		}
		return newBody
	}
	return body
}

func parseUpstreamError(body []byte) string {
	if len(body) == 0 {
		return "empty response"
	}
	var respMap map[string]any
	if err := json.Unmarshal(body, &respMap); err != nil {
		return string(body)
	}
	if errObj, ok := respMap["error"].(map[string]any); ok {
		if msg, ok := errObj["message"].(string); ok {
			return msg
		}
		if errSlice, ok := errObj["errors"].([]any); ok && len(errSlice) > 0 {
			if firstErr, ok := errSlice[0].(map[string]any); ok {
				if msg, ok := firstErr["message"].(string); ok {
					return msg
				}
			}
		}
	}
	if msg, ok := respMap["message"].(string); ok {
		return msg
	}
	return string(body)
}
