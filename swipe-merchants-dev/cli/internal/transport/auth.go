package transport

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// TokenResponse is the parsed body of POST /oauth2/token. ReceivedAt is
// stamped client-side at receipt time so the cache layer can compute
// expiry without re-parsing the JWT.
type TokenResponse struct {
	AccessToken string    `json:"access_token"`
	TokenType   string    `json:"token_type"`
	ExpiresIn   int64     `json:"expires_in"`
	Scope       string    `json:"scope,omitempty"`
	ReceivedAt  time.Time `json:"received_at"`
}

// ExpiresAt returns the wall-clock instant the access token expires. It is
// derived from ReceivedAt + ExpiresIn; the iat claim inside the JWT may
// differ by network latency.
func (t TokenResponse) ExpiresAt() time.Time {
	return t.ReceivedAt.Add(time.Duration(t.ExpiresIn) * time.Second)
}

// IssueToken performs the client_credentials grant against
// POST /oauth2/token. Scopes may be empty to request all of the client's
// permitted scopes (per RFC 6749 §3.3).
func (c *Client) IssueToken(ctx context.Context, clientID, clientSecret string, scopes []string) (TokenResponse, error) {
	form := url.Values{}
	form.Set("grant_type", "client_credentials")
	if len(scopes) > 0 {
		form.Set("scope", strings.Join(scopes, " "))
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/oauth2/token", strings.NewReader(form.Encode()))
	if err != nil {
		return TokenResponse{}, fmt.Errorf("build token request: %w", err)
	}
	req.SetBasicAuth(clientID, clientSecret)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return TokenResponse{}, fmt.Errorf("issue token: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return TokenResponse{}, fmt.Errorf("issue token: read body: %w", err)
	}
	if resp.StatusCode >= 400 {
		return TokenResponse{}, fmt.Errorf("issue token: status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	var out TokenResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return TokenResponse{}, fmt.Errorf("issue token: decode: %w", err)
	}
	out.ReceivedAt = time.Now().UTC()
	return out, nil
}

// WhoAmIResponse mirrors spec/app.yaml#WhoAmIResponse.
type WhoAmIResponse struct {
	ClientID   string   `json:"client_id"`
	MerchantID string   `json:"merchant_id"`
	Scopes     []string `json:"scopes,omitempty"`
}

// WhoAmI calls GET /api/v1/whoami with the supplied bearer token.
func (c *Client) WhoAmI(ctx context.Context, accessToken string) (WhoAmIResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.BaseURL+"/api/v1/whoami", nil)
	if err != nil {
		return WhoAmIResponse{}, fmt.Errorf("build whoami request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return WhoAmIResponse{}, fmt.Errorf("whoami: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return WhoAmIResponse{}, fmt.Errorf("whoami: read body: %w", err)
	}
	if resp.StatusCode >= 400 {
		return WhoAmIResponse{}, fmt.Errorf("whoami: status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	var out WhoAmIResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return WhoAmIResponse{}, fmt.Errorf("whoami: decode: %w", err)
	}
	return out, nil
}
