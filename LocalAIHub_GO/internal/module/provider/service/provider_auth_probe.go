package service

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"localaihub/localaihub_go/internal/module/provider/repository"
	"localaihub/localaihub_go/internal/pkg/logger"
)

type authProbeResult struct {
	Success          bool
	AuthType         string
	TestedURL        string
	LatencyMs        int
	Message          string
	AutoDetected     bool
	FallbackUsed     bool
	StatusCode       int
	ProviderKeyError string
}

func probeProviderAuth(ctx context.Context, client *http.Client, provider *repository.Provider, secret string) (*authProbeResult, error) {
	start := time.Now()
	logger.Log.Debug().
		Int64("provider_id", provider.ID).
		Str("provider_type", provider.ProviderType).
		Str("base_url", provider.BaseURL).
		Str("configured_auth_type", provider.AuthType).
		Int("timeout_ms", provider.TimeoutMS).
		Bool("has_secret", strings.TrimSpace(secret) != "").
		Msg("starting provider auth probe")
	if strings.TrimSpace(secret) == "" {
		return probeWithoutKey(ctx, client, provider, start)
	}

	preferred := provider.AuthType
	if preferred == "" {
		preferred = "x_api_key"
	}
	authOrder := []string{"x_api_key", "bearer"}
	if preferred == "bearer" {
		authOrder = []string{"bearer", "x_api_key"}
	}

	first, err := tryAuthMode(ctx, client, provider, secret, authOrder[0], start)
	if err == nil && first.Success {
		logger.Log.Debug().
			Int64("provider_id", provider.ID).
			Str("auth_type", first.AuthType).
			Str("tested_url", first.TestedURL).
			Int("status_code", first.StatusCode).
			Bool("fallback_used", first.FallbackUsed).
			Msg("provider auth probe succeeded on first auth mode")
		first.AutoDetected = authOrder[0] != provider.AuthType && provider.AuthType != ""
		return first, nil
	}
	if first != nil {
		logger.Log.Debug().
			Int64("provider_id", provider.ID).
			Str("auth_type", authOrder[0]).
			Str("tested_url", first.TestedURL).
			Int("status_code", first.StatusCode).
			Str("message", first.Message).
			Bool("fallback_used", first.FallbackUsed).
			Msg("provider auth probe first auth mode failed")
	}

	second, secondErr := tryAuthMode(ctx, client, provider, secret, authOrder[1], start)
	if secondErr == nil && second.Success {
		logger.Log.Debug().
			Int64("provider_id", provider.ID).
			Str("auth_type", second.AuthType).
			Str("tested_url", second.TestedURL).
			Int("status_code", second.StatusCode).
			Bool("fallback_used", second.FallbackUsed).
			Msg("provider auth probe succeeded on fallback auth mode")
		second.AutoDetected = authOrder[1] != provider.AuthType
		return second, nil
	}
	if second != nil {
		logger.Log.Warn().
			Int64("provider_id", provider.ID).
			Str("first_auth_type", authOrder[0]).
			Str("second_auth_type", authOrder[1]).
			Str("tested_url", second.TestedURL).
			Int("status_code", second.StatusCode).
			Str("message", second.Message).
			Msg("provider auth probe failed on all auth modes")
	}

	if second != nil {
		return second, nil
	}
	if first != nil {
		return first, nil
	}
	if secondErr != nil {
		return nil, secondErr
	}
	return nil, err
}

func probeWithoutKey(ctx context.Context, client *http.Client, provider *repository.Provider, start time.Time) (*authProbeResult, error) {
	testURL := normalizeURL(normalizedProviderBase(provider.BaseURL, true), provider.ProviderType)
	logger.Log.Debug().Int64("provider_id", provider.ID).Str("tested_url", testURL).Msg("probing provider without key using primary url")
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, testURL, nil)
	resp, err := client.Do(req)
	if err == nil && resp != nil {
		defer resp.Body.Close()
		if resp.StatusCode < 400 {
			return &authProbeResult{
				Success:    true,
				AuthType:   provider.AuthType,
				TestedURL:  testURL,
				LatencyMs:  int(time.Since(start).Milliseconds()),
				Message:    "connection ok (no key configured)",
				StatusCode: resp.StatusCode,
			}, nil
		}
	}

	altURL := normalizedProviderBase(provider.BaseURL, false)
	logger.Log.Debug().Int64("provider_id", provider.ID).Str("tested_url", altURL).Msg("probing provider without key using fallback url")
	altReq, _ := http.NewRequestWithContext(ctx, http.MethodGet, altURL, nil)
	altResp, altErr := client.Do(altReq)
	if altErr == nil && altResp != nil {
		defer altResp.Body.Close()
		if altResp.StatusCode < 400 {
			return &authProbeResult{
				Success:      true,
				AuthType:     provider.AuthType,
				TestedURL:    altURL,
				LatencyMs:    int(time.Since(start).Milliseconds()),
				Message:      "connection ok (no key configured)",
				FallbackUsed: true,
				StatusCode:   altResp.StatusCode,
			}, nil
		}
	}

	message := "connection failed"
	if altErr != nil {
		message = altErr.Error()
	} else if err != nil {
		message = err.Error()
	} else if altResp != nil {
		message = fmt.Sprintf("HTTP %d", altResp.StatusCode)
	} else if resp != nil {
		message = fmt.Sprintf("HTTP %d", resp.StatusCode)
	}

	return &authProbeResult{
		Success:   false,
		AuthType:  provider.AuthType,
		TestedURL: altURL,
		LatencyMs: int(time.Since(start).Milliseconds()),
		Message:   message,
	}, nil
}

func tryAuthMode(ctx context.Context, client *http.Client, provider *repository.Provider, secret, authType string, start time.Time) (*authProbeResult, error) {
	firstURL := testedURLWithSuffixAndAuth(provider, true)
	logger.Log.Debug().
		Int64("provider_id", provider.ID).
		Str("auth_type", authType).
		Str("tested_url", firstURL).
		Str("secret_masked", maskSecret(secret)).
		Msg("trying provider auth mode with primary url")
	resp, err := doProviderAuthRequest(ctx, client, provider, secret, authType, true)
	if err == nil && resp != nil {
		bodyPreview := readResponsePreview(resp)
		logger.Log.Debug().
			Int64("provider_id", provider.ID).
			Str("auth_type", authType).
			Str("tested_url", firstURL).
			Int("status_code", resp.StatusCode).
			Str("response_preview", bodyPreview).
			Msg("provider auth mode primary url responded")
		if resp.StatusCode < 400 {
			resp.Body.Close()
			return &authProbeResult{
				Success:    true,
				AuthType:   authType,
				TestedURL:  firstURL,
				LatencyMs:  int(time.Since(start).Milliseconds()),
				Message:    "connection success",
				StatusCode: resp.StatusCode,
			}, nil
		}
		resp.Body.Close()
	}

	secondURL := testedURLWithSuffixAndAuth(provider, false)
	logger.Log.Debug().
		Int64("provider_id", provider.ID).
		Str("auth_type", authType).
		Str("tested_url", secondURL).
		Msg("trying provider auth mode with fallback url")
	if resp != nil {
		resp.Body.Close()
	}
	resp, err = doProviderAuthRequest(ctx, client, provider, secret, authType, false)
	if err == nil && resp != nil {
		bodyPreview := readResponsePreview(resp)
		logger.Log.Debug().
			Int64("provider_id", provider.ID).
			Str("auth_type", authType).
			Str("tested_url", secondURL).
			Int("status_code", resp.StatusCode).
			Str("response_preview", bodyPreview).
			Msg("provider auth mode fallback url responded")
		if resp.StatusCode < 400 {
			resp.Body.Close()
			return &authProbeResult{
				Success:      true,
				AuthType:     authType,
				TestedURL:    secondURL,
				LatencyMs:    int(time.Since(start).Milliseconds()),
				Message:      "connection success",
				FallbackUsed: true,
				StatusCode:   resp.StatusCode,
			}, nil
		}
		resp.Body.Close()
	}

	message := "connection failed"
	statusCode := 0
	if err != nil {
		logger.Log.Warn().
			Int64("provider_id", provider.ID).
			Str("auth_type", authType).
			Str("tested_url", secondURL).
			Err(err).
			Msg("provider auth mode request failed")
		message = err.Error()
	} else if resp != nil {
		statusCode = resp.StatusCode
		logger.Log.Warn().
			Int64("provider_id", provider.ID).
			Str("auth_type", authType).
			Str("tested_url", secondURL).
			Int("status_code", resp.StatusCode).
			Msg("provider auth mode returned non-success status")
		message = fmt.Sprintf("HTTP %d", resp.StatusCode)
	}

	return &authProbeResult{
		Success:    false,
		AuthType:   authType,
		TestedURL:  secondURL,
		LatencyMs:  int(time.Since(start).Milliseconds()),
		Message:    message,
		StatusCode: statusCode,
	}, nil
}

func doProviderAuthRequest(ctx context.Context, client *http.Client, provider *repository.Provider, secret, authType string, withV1 bool) (*http.Response, error) {
	requestURL := testedURLWithSuffixAndAuth(provider, withV1)
	if provider.ProviderType != "anthropic" {
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
		if authType == "x_api_key" {
			req.Header.Set("x-api-key", secret)
		} else {
			req.Header.Set("Authorization", "Bearer "+secret)
		}
		return client.Do(req)
	}
	payload := `{"model":"gpt-4o-mini","max_tokens":1,"messages":[{"role":"user","content":"hi"}]}`
	if provider.ProviderType == "anthropic" {
		payload = `{"model":"claude-3-haiku-20240307","max_tokens":1,"messages":[{"role":"user","content":"hi"}]}`
	}
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, requestURL, strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	if authType == "x_api_key" {
		req.Header.Set("x-api-key", secret)
	} else {
		req.Header.Set("Authorization", "Bearer "+secret)
	}
	return client.Do(req)
}

func testedURLWithSuffixAndAuth(provider *repository.Provider, withV1 bool) string {
	base := normalizedProviderBase(provider.BaseURL, withV1)
	if provider.ProviderType != "anthropic" {
		return base + "/models"
	}
	if provider.ProviderType == "anthropic" {
		if withV1 {
			return base + "/messages"
		}
		return base + "/messages"
	}
	if withV1 {
		return base + "/chat/completions"
	}
	return base + "/chat/completions"
}

func normalizedProviderBase(baseURL string, withV1 bool) string {
	base := strings.TrimRight(baseURL, "/")
	base = strings.ReplaceAll(base, "/v1/v1", "/v1")
	if withV1 {
		if strings.HasSuffix(base, "/v1") {
			return base
		}
		return base + "/v1"
	}
	if strings.HasSuffix(base, "/v1") {
		return base
	}
	return strings.TrimSuffix(base, "/v1")
}

func readResponsePreview(resp *http.Response) string {
	if resp == nil || resp.Body == nil {
		return ""
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Sprintf("read response body failed: %v", err)
	}
	preview := strings.TrimSpace(string(body))
	if len(preview) > 500 {
		preview = preview[:500]
	}
	resp.Body = io.NopCloser(strings.NewReader(string(body)))
	return preview
}
