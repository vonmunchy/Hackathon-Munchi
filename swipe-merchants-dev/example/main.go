package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const defaultMockURL = "http://127.0.0.1:8080"

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "sample integration: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	base := os.Getenv("SWIPE_MOCK_URL")
	if base == "" {
		base = defaultMockURL
	}
	fmt.Printf("Swipe merchant sample integration\n")
	fmt.Printf("Mock base URL: %s\n\n", base)

	client := &http.Client{Timeout: 5 * time.Second}

	fmt.Println("== Phase 1: health probes ==")
	if err := probe(client, base, "/health/alive"); err != nil {
		return fmt.Errorf("alive probe: %w", err)
	}
	if err := probe(client, base, "/health/ready"); err != nil {
		return fmt.Errorf("ready probe: %w", err)
	}

	clientID := os.Getenv("SWIPE_CLIENT_ID")
	clientSecret := os.Getenv("SWIPE_CLIENT_SECRET")
	if clientID == "" || clientSecret == "" {
		fmt.Println()
		fmt.Println("== Phase 2: skipped (SWIPE_CLIENT_ID / SWIPE_CLIENT_SECRET not set) ==")
		fmt.Println("set both env vars after `swipe keys create` to exercise auth + whoami.")
		fmt.Println()
		fmt.Println("Phase 1 sample OK")
		return nil
	}

	fmt.Println()
	fmt.Println("== Phase 2: OAuth client_credentials → whoami ==")
	token, err := exchangeToken(client, base, clientID, clientSecret)
	if err != nil {
		return fmt.Errorf("token exchange: %w", err)
	}
	fmt.Printf("POST /oauth2/token -> token (expires in %ds, scopes %q)\n\n", token.ExpiresIn, token.Scope)

	who, err := whoami(client, base, token.AccessToken)
	if err != nil {
		return fmt.Errorf("whoami: %w", err)
	}
	fmt.Printf("GET /api/v1/whoami ->\n")
	fmt.Printf("  client_id:   %s\n", who.ClientID)
	fmt.Printf("  merchant_id: %s\n", who.MerchantID)
	fmt.Printf("  scopes:      %s\n\n", strings.Join(who.Scopes, " "))

	fmt.Println("== Phase 3: read endpoints (balance, bank-accounts, history) ==")
	balances, err := apiGet(client, base, "/api/v1/balance", token.AccessToken)
	if err != nil {
		return fmt.Errorf("balance: %w", err)
	}
	fmt.Printf("GET /api/v1/balance ->\n%s\n\n", prettyJSON(balances))

	accounts, err := apiGet(client, base, "/api/v1/bank-accounts", token.AccessToken)
	if err != nil {
		return fmt.Errorf("bank-accounts: %w", err)
	}
	fmt.Printf("GET /api/v1/bank-accounts ->\n%s\n\n", prettyJSON(accounts))

	history, err := apiGet(client, base, "/api/v1/history?limit=5", token.AccessToken)
	if err != nil {
		return fmt.Errorf("history: %w", err)
	}
	fmt.Printf("GET /api/v1/history?limit=5 ->\n%s\n\n", prettyJSON(history))

	fmt.Println("== Phase 4: create payment + watch SSE ==")
	payment, err := apiPost(client, base, "/api/v1/payments", token.AccessToken,
		map[string]any{"amount": 42.50, "currency": "MVR", "type": "QR", "description": "sample"})
	if err != nil {
		return fmt.Errorf("create payment: %w", err)
	}
	fmt.Printf("POST /api/v1/payments ->\n%s\n\n", prettyJSON(payment))

	var paymentBody struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(payment, &paymentBody); err != nil {
		return fmt.Errorf("parse payment id: %w", err)
	}
	fmt.Printf("watching %s via SSE\n", paymentBody.ID)
	if err := watchPayment(base, paymentBody.ID, token.AccessToken); err != nil {
		return fmt.Errorf("watch: %w", err)
	}

	fmt.Println()
	fmt.Println("== Phase 5: customer-facing pay page (mock-only) ==")
	if err := demoPayPage(client, base, token.AccessToken); err != nil {
		return fmt.Errorf("pay page demo: %w", err)
	}

	fmt.Println("Phase 5 sample OK")
	return nil
}

// demoPayPage exercises the mock's `/pay/{shortCode}` customer flow:
// create a fresh QR payment, surface the URL a real customer would
// click, drive the transition via the pay-page POST (no auth needed —
// a real customer has no credentials), then read the resulting
// transaction back through the merchant API.
//
// The mock pay page is a developer ergonomic, not a spec surface; the
// real Swipe Merchants API has no such route. Production integrations
// rely on a bank app scanning the EMVCo QR — see `swipe mock start
// --qr-format emvco` if you need the production-shape QR.
func demoPayPage(client *http.Client, base, accessToken string) error {
	payment, err := apiPost(client, base, "/api/v1/payments", accessToken,
		map[string]any{"amount": 7.50, "currency": "MVR", "type": "QR", "description": "pay-page demo"})
	if err != nil {
		return fmt.Errorf("create: %w", err)
	}
	var body struct {
		Reference  string `json:"reference"`
		PaymentURL string `json:"payment_url"`
	}
	if err := json.Unmarshal(payment, &body); err != nil {
		return fmt.Errorf("parse: %w", err)
	}
	fmt.Printf("created reference=%s\n", body.Reference)
	fmt.Printf("payment_url:   %s   (open in browser to click Pay)\n", body.PaymentURL)

	// Simulate the customer clicking Pay. The pay-page POSTs are public
	// (no Bearer token) and 303-redirect back to GET — we don't follow.
	noRedirect := &http.Client{
		Timeout: 3 * time.Second,
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, base+"/pay/"+body.Reference+"/complete", nil)
	if err != nil {
		return fmt.Errorf("build pay POST: %w", err)
	}
	resp, err := noRedirect.Do(req)
	if err != nil {
		return fmt.Errorf("pay POST: %w", err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusSeeOther {
		return fmt.Errorf("pay POST status=%d, want 303", resp.StatusCode)
	}
	fmt.Printf("POST /pay/%s/complete -> 303 (customer paid)\n", body.Reference)

	txn, err := apiGet(client, base, "/api/v1/transactions/"+body.Reference, accessToken)
	if err != nil {
		return fmt.Errorf("read transaction: %w", err)
	}
	fmt.Printf("GET /api/v1/transactions/%s ->\n%s\n", body.Reference, prettyJSON(txn))
	return nil
}

// apiPost performs a bearer-auth POST with a JSON body and returns the raw
// response. Used for Phase 4's createPayment/createPayout.
func apiPost(client *http.Client, base, path, token string, body any) ([]byte, error) {
	raw, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal: %w", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, base+path, bytes.NewReader(raw))
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	out, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read: %w", err)
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("status %d: %s", resp.StatusCode, strings.TrimSpace(string(out)))
	}
	return out, nil
}

// watchPayment opens an SSE stream and prints each status event until a
// terminal state arrives.
func watchPayment(base, id, token string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, base+"/api/v1/payments/"+id+"/stream", nil)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "text/event-stream")
	client := &http.Client{Timeout: 0}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("open stream: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("stream status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 4096), 1024*1024)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		payload := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		var ev struct {
			ID        string `json:"id"`
			Status    string `json:"status"`
			Timestamp string `json:"timestamp"`
		}
		if err := json.Unmarshal([]byte(payload), &ev); err != nil {
			continue
		}
		fmt.Printf("  %-9s %s @ %s\n", ev.Status, ev.ID, ev.Timestamp)
		if ev.Status == "COMPLETED" || ev.Status == "EXPIRED" || ev.Status == "CANCELLED" {
			return nil
		}
	}
	return scanner.Err()
}

// apiGet performs a bearer-auth GET and returns the raw response body. Used
// for the Phase 3 read endpoints where the example just wants to print the
// JSON the merchant would receive.
func apiGet(client *http.Client, base, path, token string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, base+path, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read: %w", err)
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	return body, nil
}

func probe(client *http.Client, base, path string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, base+path, nil)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("send %s: %w", path, err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}

	fmt.Printf("GET %s -> %d\n%s\n\n", path, resp.StatusCode, prettyJSON(body))

	if resp.StatusCode >= 400 {
		return fmt.Errorf("%s returned %d", path, resp.StatusCode)
	}
	return nil
}

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope"`
}

func exchangeToken(client *http.Client, base, clientID, clientSecret string) (tokenResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	form := url.Values{}
	form.Set("grant_type", "client_credentials")
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, base+"/oauth2/token", strings.NewReader(form.Encode()))
	if err != nil {
		return tokenResponse{}, fmt.Errorf("build request: %w", err)
	}
	req.SetBasicAuth(clientID, clientSecret)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return tokenResponse{}, err
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return tokenResponse{}, fmt.Errorf("read: %w", err)
	}
	if resp.StatusCode >= 400 {
		return tokenResponse{}, fmt.Errorf("status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	var out tokenResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return tokenResponse{}, fmt.Errorf("decode: %w", err)
	}
	return out, nil
}

type whoamiResponse struct {
	ClientID   string   `json:"client_id"`
	MerchantID string   `json:"merchant_id"`
	Scopes     []string `json:"scopes"`
}

func whoami(client *http.Client, base, accessToken string) (whoamiResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, base+"/api/v1/whoami", nil)
	if err != nil {
		return whoamiResponse{}, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := client.Do(req)
	if err != nil {
		return whoamiResponse{}, err
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return whoamiResponse{}, fmt.Errorf("read: %w", err)
	}
	if resp.StatusCode >= 400 {
		return whoamiResponse{}, fmt.Errorf("status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	var out whoamiResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return whoamiResponse{}, fmt.Errorf("decode: %w", err)
	}
	return out, nil
}

func prettyJSON(b []byte) string {
	var v any
	if err := json.Unmarshal(b, &v); err != nil {
		return string(b)
	}
	out, err := json.MarshalIndent(v, "  ", "  ")
	if err != nil {
		return string(b)
	}
	return "  " + string(out)
}
