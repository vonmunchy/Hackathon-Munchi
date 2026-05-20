package transport

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/handlers"
)

// CreatePayment calls POST /api/v1/payments with the supplied request and
// bearer token. Returns the parsed PaymentResponse or a wrapped error
// that preserves the HTTP status for sad-path tests.
func (c *Client) CreatePayment(ctx context.Context, token string, req handlers.CreatePaymentRequest) (handlers.PaymentResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return handlers.PaymentResponse{}, fmt.Errorf("marshal request: %w", err)
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/api/v1/payments", bytes.NewReader(body))
	if err != nil {
		return handlers.PaymentResponse{}, fmt.Errorf("build request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+token)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return handlers.PaymentResponse{}, fmt.Errorf("create payment: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	rawBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return handlers.PaymentResponse{}, fmt.Errorf("create payment: read body: %w", err)
	}
	if resp.StatusCode >= 400 {
		return handlers.PaymentResponse{}, fmt.Errorf("create payment: status %d: %s", resp.StatusCode, strings.TrimSpace(string(rawBody)))
	}
	var out handlers.PaymentResponse
	if err := json.Unmarshal(rawBody, &out); err != nil {
		return handlers.PaymentResponse{}, fmt.Errorf("create payment: decode: %w", err)
	}
	return out, nil
}

// CreatePayout calls POST /api/v1/payouts with the supplied request and
// bearer token.
func (c *Client) CreatePayout(ctx context.Context, token string, req handlers.CreatePayoutRequest) (handlers.PayoutResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return handlers.PayoutResponse{}, fmt.Errorf("marshal request: %w", err)
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/api/v1/payouts", bytes.NewReader(body))
	if err != nil {
		return handlers.PayoutResponse{}, fmt.Errorf("build request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+token)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return handlers.PayoutResponse{}, fmt.Errorf("create payout: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	rawBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return handlers.PayoutResponse{}, fmt.Errorf("create payout: read body: %w", err)
	}
	if resp.StatusCode >= 400 {
		return handlers.PayoutResponse{}, fmt.Errorf("create payout: status %d: %s", resp.StatusCode, strings.TrimSpace(string(rawBody)))
	}
	var out handlers.PayoutResponse
	if err := json.Unmarshal(rawBody, &out); err != nil {
		return handlers.PayoutResponse{}, fmt.Errorf("create payout: decode: %w", err)
	}
	return out, nil
}

// StreamEvent is one event delivered by WatchPayment.
type StreamEvent struct {
	ID        string `json:"id"`
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
}

// WatchPayment opens an SSE stream on
// GET /api/v1/payments/{id}/stream, parses one event per yielded record,
// and pushes results to fn until ctx is cancelled or fn returns false.
// The returned error is non-nil for transport/parse failures; reaching a
// terminal state (after fn observes it) is a clean nil return.
//
// fn is invoked for every status event the server emits. Returning false
// signals "stop watching" — the SSE connection is closed and WatchPayment
// returns nil. Returning true keeps the stream open.
func (c *Client) WatchPayment(ctx context.Context, token, paymentID string, fn func(StreamEvent) bool) error {
	url := c.BaseURL + "/api/v1/payments/" + paymentID + "/stream"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("build watch request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "text/event-stream")

	// The default Client has a 30s timeout which would cap watch sessions.
	// Build a per-call client without a timeout; context is the cancel signal.
	client := *c.HTTPClient
	client.Timeout = 0

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("open stream: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("stream: status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 4096), 1024*1024)
	var dataBuf strings.Builder
	for scanner.Scan() {
		line := scanner.Text()
		switch {
		case line == "":
			if dataBuf.Len() == 0 {
				continue
			}
			var ev StreamEvent
			if err := json.Unmarshal([]byte(dataBuf.String()), &ev); err == nil {
				if !fn(ev) {
					return nil
				}
			}
			dataBuf.Reset()
		case strings.HasPrefix(line, "data:"):
			payload := strings.TrimPrefix(line, "data:")
			payload = strings.TrimPrefix(payload, " ")
			dataBuf.WriteString(payload)
		}
		// Lines starting with ':' (heartbeat comments) or "event:" (event
		// type) are not part of the data accumulation and are intentionally
		// dropped on the floor.
	}
	if err := scanner.Err(); err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return nil
		}
		return fmt.Errorf("read stream: %w", err)
	}
	return nil
}
