package transport

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client is the HTTP transport for the public OpenAPI surface (everything
// under /api/v1, /health, /oauth2, /.well-known). It is unauthenticated by
// default; later phases add a TokenProvider injection point.
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
}

// NewClient builds a Client targeting baseURL (e.g. http://localhost:8080).
// The default http.Client uses a 30s timeout, sufficient for every Phase 1
// operation; long-poll endpoints (SSE) override per-call.
func NewClient(baseURL string) *Client {
	return &Client{
		BaseURL: strings.TrimRight(baseURL, "/"),
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// HealthAlive calls GET /health/alive and returns the parsed body.
func (c *Client) HealthAlive(ctx context.Context) (map[string]any, error) {
	var out map[string]any
	if err := c.do(ctx, http.MethodGet, "/health/alive", nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// HealthReady calls GET /health/ready. The mock returns 503 with a body
// when unhealthy; the body is parsed regardless and returned alongside the
// status code so callers can render either outcome.
func (c *Client) HealthReady(ctx context.Context) (int, map[string]any, error) {
	req, err := c.newRequest(ctx, http.MethodGet, "/health/ready", nil)
	if err != nil {
		return 0, nil, err
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return 0, nil, fmt.Errorf("health ready: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, nil, fmt.Errorf("health ready: read body: %w", err)
	}
	var parsed map[string]any
	if len(body) > 0 {
		if err := json.Unmarshal(body, &parsed); err != nil {
			return resp.StatusCode, nil, fmt.Errorf("health ready: parse body: %w", err)
		}
	}
	return resp.StatusCode, parsed, nil
}

func (c *Client) newRequest(ctx context.Context, method, path string, body any) (*http.Request, error) {
	url := c.BaseURL + path
	var rdr io.Reader
	if body != nil {
		buf, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal request body: %w", err)
		}
		rdr = bytes.NewReader(buf)
	}
	req, err := http.NewRequestWithContext(ctx, method, url, rdr)
	if err != nil {
		return nil, fmt.Errorf("new request %s %s: %w", method, url, err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")
	return req, nil
}

func (c *Client) do(ctx context.Context, method, path string, body any, out any) error {
	req, err := c.newRequest(ctx, method, path, body)
	if err != nil {
		return err
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("%s %s: %w", method, path, err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode >= 400 {
		buf, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("%s %s: status %d: %s", method, path, resp.StatusCode, strings.TrimSpace(string(buf)))
	}
	if out == nil || resp.StatusCode == http.StatusNoContent {
		return nil
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("%s %s: decode response: %w", method, path, err)
	}
	return nil
}
