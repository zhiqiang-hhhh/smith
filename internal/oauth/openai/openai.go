// Package openai provides functions to handle OpenAI OAuth device flow
// authentication, following the same flow as OpenAI's Codex CLI.
package openai

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/zhiqiang-hhhh/smith/internal/oauth"
)

const (
	clientID      = "app_EMoamEEZ73f0CkXaXp7hrann"
	defaultIssuer = "https://auth.openai.com"
	userAgent     = "Smith/1.0"
	// MaxPollTimeout is the maximum time to wait for user authorization.
	MaxPollTimeout = 15 * time.Minute
)

// DeviceCode contains the response from requesting a device code.
type DeviceCode struct {
	DeviceAuthID    string `json:"device_auth_id"`
	UserCode        string `json:"user_code"`
	VerificationURL string `json:"verification_url"`
	Interval        int    `json:"interval"`
}

// RequestDeviceCode initiates the device code flow with OpenAI.
func RequestDeviceCode(ctx context.Context) (*DeviceCode, error) {
	apiURL := defaultIssuer + "/api/accounts/deviceauth/usercode"

	body, err := json.Marshal(map[string]string{
		"client_id": clientID,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, strings.NewReader(string(body)))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", userAgent)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("device code login is not enabled for this server")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("device code request failed: status %d, body %q", resp.StatusCode, string(respBody))
	}

	var dc DeviceCode
	if err := json.Unmarshal(respBody, &dc); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	dc.VerificationURL = defaultIssuer + "/codex/device"

	if dc.Interval < 5 {
		dc.Interval = 5
	}

	return &dc, nil
}

// codeSuccessResp is returned when polling succeeds.
type codeSuccessResp struct {
	AuthorizationCode string `json:"authorization_code"`
	CodeChallenge     string `json:"code_challenge"`
	CodeVerifier      string `json:"code_verifier"`
}

// PollForToken polls the OpenAI device auth endpoint until the user
// completes authorization. It then exchanges the authorization code for
// tokens via PKCE and returns an OAuth token with an API key as the access
// token.
func PollForToken(ctx context.Context, dc *DeviceCode) (*oauth.Token, error) {
	ctx, cancel := context.WithTimeout(ctx, MaxPollTimeout)
	defer cancel()

	interval := time.Duration(dc.Interval) * time.Second
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	apiURL := defaultIssuer + "/api/accounts/deviceauth/token"

	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("device auth timed out")
		case <-ticker.C:
		}

		body, err := json.Marshal(map[string]string{
			"device_auth_id": dc.DeviceAuthID,
			"user_code":      dc.UserCode,
		})
		if err != nil {
			return nil, fmt.Errorf("marshal request: %w", err)
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, strings.NewReader(string(body)))
		if err != nil {
			return nil, fmt.Errorf("create request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("User-Agent", userAgent)

		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("execute request: %w", err)
		}

		respBody, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
		resp.Body.Close()

		if err != nil {
			return nil, fmt.Errorf("read response: %w", err)
		}

		if resp.StatusCode == http.StatusOK {
			var codeResp codeSuccessResp
			if err := json.Unmarshal(respBody, &codeResp); err != nil {
				return nil, fmt.Errorf("unmarshal token response: %w", err)
			}

			return exchangeCodeForTokens(ctx, &codeResp)
		}

		if resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusNotFound {
			continue
		}

		return nil, fmt.Errorf("device auth failed: status %d, body %q", resp.StatusCode, string(respBody))
	}
}

// exchangeCodeForTokens exchanges the authorization code for tokens using
// PKCE, then performs a token exchange to get an API key.
func exchangeCodeForTokens(ctx context.Context, codeResp *codeSuccessResp) (*oauth.Token, error) {
	redirectURI := defaultIssuer + "/deviceauth/callback"

	data := url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {codeResp.AuthorizationCode},
		"redirect_uri":  {redirectURI},
		"client_id":     {clientID},
		"code_verifier": {codeResp.CodeVerifier},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, defaultIssuer+"/oauth/token", strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("create token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", userAgent)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute token request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("read token response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token exchange failed: status %d, body %q", resp.StatusCode, string(body))
	}

	var tokenResp struct {
		IDToken      string `json:"id_token"`
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
	}
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("unmarshal token response: %w", err)
	}

	apiKey, err := obtainAPIKey(ctx, tokenResp.IDToken)
	if err != nil {
		apiKey = tokenResp.AccessToken
	}

	token := &oauth.Token{
		AccessToken:  apiKey,
		RefreshToken: tokenResp.RefreshToken,
		ExpiresIn:    tokenResp.ExpiresIn,
	}
	token.SetExpiresAt()

	return token, nil
}

// obtainAPIKey exchanges the ID token for an OpenAI API key.
func obtainAPIKey(ctx context.Context, idToken string) (string, error) {
	data := url.Values{
		"grant_type":         {"urn:ietf:params:oauth:grant-type:token-exchange"},
		"client_id":          {clientID},
		"requested_token":    {"openai-api-key"},
		"subject_token":      {idToken},
		"subject_token_type": {"urn:ietf:params:oauth:token-type:id_token"},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, defaultIssuer+"/oauth/token", strings.NewReader(data.Encode()))
	if err != nil {
		return "", fmt.Errorf("create API key request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", userAgent)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("execute API key request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", fmt.Errorf("read API key response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API key exchange failed: status %d, body %q", resp.StatusCode, string(body))
	}

	var result struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("unmarshal API key response: %w", err)
	}

	return result.AccessToken, nil
}

// RefreshToken refreshes an OpenAI OAuth token using the refresh token.
func RefreshToken(ctx context.Context, refreshToken string) (*oauth.Token, error) {
	data := url.Values{
		"grant_type":    {"refresh_token"},
		"client_id":     {clientID},
		"refresh_token": {refreshToken},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, defaultIssuer+"/oauth/token", strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("create refresh request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", userAgent)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute refresh request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("read refresh response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token refresh failed: status %d, body %q", resp.StatusCode, string(body))
	}

	var tokenResp struct {
		IDToken      string `json:"id_token"`
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
	}
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("unmarshal refresh response: %w", err)
	}

	newRefreshToken := tokenResp.RefreshToken
	if newRefreshToken == "" {
		newRefreshToken = refreshToken
	}

	apiKey, err := obtainAPIKey(ctx, tokenResp.IDToken)
	if err != nil {
		apiKey = tokenResp.AccessToken
	}

	token := &oauth.Token{
		AccessToken:  apiKey,
		RefreshToken: newRefreshToken,
		ExpiresIn:    tokenResp.ExpiresIn,
	}
	token.SetExpiresAt()

	return token, nil
}

