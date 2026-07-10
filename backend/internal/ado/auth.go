package ado

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// tokenProvider manages OAuth2 tokens for Service Principal authentication.
type tokenProvider struct {
	tenantID     string
	clientID     string
	clientSecret string

	mu          sync.RWMutex
	accessToken string
	expiresAt   time.Time
}

// azureDevOpsScope is the resource scope for Azure DevOps API tokens.
const azureDevOpsScope = "499b84ac-1321-427f-aa17-267ca6975798/.default"

// newTokenProvider creates a token provider for the given SP credentials.
func newTokenProvider(tenantID, clientID, clientSecret string) *tokenProvider {
	return &tokenProvider{
		tenantID:     tenantID,
		clientID:     clientID,
		clientSecret: clientSecret,
	}
}

// getToken returns a valid access token, refreshing if expired or about to expire.
func (tp *tokenProvider) getToken() (string, error) {
	tp.mu.RLock()
	if tp.accessToken != "" && time.Now().Before(tp.expiresAt.Add(-2*time.Minute)) {
		token := tp.accessToken
		tp.mu.RUnlock()
		return token, nil
	}
	tp.mu.RUnlock()

	return tp.refreshToken()
}

// refreshToken acquires a new token from Azure AD using client credentials flow.
func (tp *tokenProvider) refreshToken() (string, error) {
	tp.mu.Lock()
	defer tp.mu.Unlock()

	// Double-check after acquiring write lock
	if tp.accessToken != "" && time.Now().Before(tp.expiresAt.Add(-2*time.Minute)) {
		return tp.accessToken, nil
	}

	tokenURL := fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/v2.0/token", tp.tenantID)

	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("client_id", tp.clientID)
	data.Set("client_secret", tp.clientSecret)
	data.Set("scope", azureDevOpsScope)

	resp, err := http.Post(tokenURL, "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
	if err != nil {
		return "", fmt.Errorf("requesting token: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading token response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("token request failed (status %d): %s", resp.StatusCode, string(body))
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
		TokenType   string `json:"token_type"`
	}
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", fmt.Errorf("parsing token response: %w", err)
	}

	tp.accessToken = tokenResp.AccessToken
	tp.expiresAt = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	return tp.accessToken, nil
}
