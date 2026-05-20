package transport

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/handlers"
)

// doAuthed runs a request with a bearer token attached. It returns the
// raw status code, body bytes, and any transport error. Callers decide
// what to do with non-2xx responses since the same code path is reused
// by smoke tests that want to assert on 401/403.
func (c *Client) doAuthed(ctx context.Context, method, path, token string, query map[string]string) (int, []byte, error) {
	u := c.BaseURL + path
	if len(query) > 0 {
		parts := make([]string, 0, len(query))
		for k, v := range query {
			parts = append(parts, k+"="+v)
		}
		u += "?" + strings.Join(parts, "&")
	}
	req, err := http.NewRequestWithContext(ctx, method, u, nil)
	if err != nil {
		return 0, nil, fmt.Errorf("build %s %s: %w", method, path, err)
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	req.Header.Set("Accept", "application/json")
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return 0, nil, fmt.Errorf("%s %s: %w", method, path, err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, nil, fmt.Errorf("%s %s: read body: %w", method, path, err)
	}
	return resp.StatusCode, body, nil
}

// Balance calls GET /api/v1/balance with the supplied bearer token.
func (c *Client) Balance(ctx context.Context, token string) ([]handlers.BalanceResponse, error) {
	status, body, err := c.doAuthed(ctx, http.MethodGet, "/api/v1/balance", token, nil)
	if err != nil {
		return nil, err
	}
	if status >= 400 {
		return nil, fmt.Errorf("balance: status %d: %s", status, strings.TrimSpace(string(body)))
	}
	var out []handlers.BalanceResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("balance: decode: %w", err)
	}
	return out, nil
}

// BankAccounts calls GET /api/v1/bank-accounts with the supplied bearer
// token.
func (c *Client) BankAccounts(ctx context.Context, token string) ([]handlers.BankAccountResponse, error) {
	status, body, err := c.doAuthed(ctx, http.MethodGet, "/api/v1/bank-accounts", token, nil)
	if err != nil {
		return nil, err
	}
	if status >= 400 {
		return nil, fmt.Errorf("bank-accounts: status %d: %s", status, strings.TrimSpace(string(body)))
	}
	var out []handlers.BankAccountResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("bank-accounts: decode: %w", err)
	}
	return out, nil
}

// History calls GET /api/v1/history with the supplied bearer token and
// pagination parameters. Pass limit=0 to omit (server uses its default).
func (c *Client) History(ctx context.Context, token string, limit, offset int) (handlers.HistoryResponse, error) {
	q := map[string]string{}
	if limit > 0 {
		q["limit"] = strconv.Itoa(limit)
	}
	if offset > 0 {
		q["offset"] = strconv.Itoa(offset)
	}
	status, body, err := c.doAuthed(ctx, http.MethodGet, "/api/v1/history", token, q)
	if err != nil {
		return handlers.HistoryResponse{}, err
	}
	if status >= 400 {
		return handlers.HistoryResponse{}, fmt.Errorf("history: status %d: %s", status, strings.TrimSpace(string(body)))
	}
	var out handlers.HistoryResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return handlers.HistoryResponse{}, fmt.Errorf("history: decode: %w", err)
	}
	return out, nil
}

// GetTransaction calls GET /api/v1/transactions/{reference} (spec v1.1.0;
// operationId getTransactionStatus). Returns the TransactionItem or a
// wrapped status error.
func (c *Client) GetTransaction(ctx context.Context, token, reference string) (handlers.TransactionItem, error) {
	status, body, err := c.doAuthed(ctx, http.MethodGet, "/api/v1/transactions/"+reference, token, nil)
	if err != nil {
		return handlers.TransactionItem{}, err
	}
	if status >= 400 {
		return handlers.TransactionItem{}, fmt.Errorf("get transaction: status %d: %s", status, strings.TrimSpace(string(body)))
	}
	var out handlers.TransactionItem
	if err := json.Unmarshal(body, &out); err != nil {
		return handlers.TransactionItem{}, fmt.Errorf("get transaction: decode: %w", err)
	}
	return out, nil
}

// GetPayment calls GET /api/v1/payments/{id} with the supplied bearer
// token. Phase 3 returns NOT_FOUND for every id until Phase 4 lands the
// payment store.
func (c *Client) GetPayment(ctx context.Context, token, id string) ([]byte, int, error) {
	status, body, err := c.doAuthed(ctx, http.MethodGet, "/api/v1/payments/"+id, token, nil)
	if err != nil {
		return nil, 0, err
	}
	return body, status, nil
}
