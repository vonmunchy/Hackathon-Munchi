package handlers_test

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"testing"

	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock"
	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/auth"
	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/handlers"
	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/store"
	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/transport"
)

// phase3Server boots a fresh mock with a seeded client whose scope set is
// `scopes`. It returns a transport.Client + the bearer token + the helper
// HTTP client (for raw probes).
func phase3Server(t *testing.T, scopes []string) (*transport.Client, string) {
	t.Helper()
	dir := t.TempDir()
	st, err := store.Open(filepath.Join(dir, "state.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })
	if _, err := st.SeedDefaults(); err != nil {
		t.Fatalf("seed: %v", err)
	}
	key, err := auth.LoadOrCreateSigningKey(filepath.Join(dir, "keys"))
	if err != nil {
		t.Fatalf("key: %v", err)
	}
	srv, err := mock.New(mock.Config{
		ListenAddr: "127.0.0.1:0",
		Store:      st,
		Logger:     slog.New(slog.NewJSONHandler(io.Discard, nil)),
		SigningKey: key,
	})
	if err != nil {
		t.Fatalf("new: %v", err)
	}
	if err := srv.Bind(); err != nil {
		t.Fatalf("bind: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	go func() { _ = srv.Serve(ctx) }()
	t.Cleanup(cancel)

	c, secret, err := st.CreateClient(store.DefaultMerchantID, "t", scopes)
	if err != nil {
		t.Fatalf("client: %v", err)
	}
	base := "http://" + srv.Listener().Addr().String()
	tc := transport.NewClient(base)
	tok, err := tc.IssueToken(context.Background(), c.ID, secret, nil)
	if err != nil {
		t.Fatalf("token: %v", err)
	}
	return tc, tok.AccessToken
}

func TestBalance_HappyPath(t *testing.T) {
	tc, token := phase3Server(t, []string{"wallet:balance"})
	balances, err := tc.Balance(context.Background(), token)
	if err != nil {
		t.Fatalf("balance: %v", err)
	}
	if len(balances) != 2 {
		t.Fatalf("got %d balances, want 2", len(balances))
	}
	// The seed inserts MVR (10000) and USD (650). Iteration order is sorted.
	if balances[0].Currency != "MVR" || balances[0].AvailableBalance != 10000 {
		t.Errorf("MVR: %+v", balances[0])
	}
	if balances[1].Currency != "USD" || balances[1].AvailableBalance != 650 {
		t.Errorf("USD: %+v", balances[1])
	}
}

func TestBalance_MissingScope_Returns403(t *testing.T) {
	tc, token := phase3Server(t, []string{"transactions:status"})
	if _, err := tc.Balance(context.Background(), token); err == nil || !strings.Contains(err.Error(), "403") {
		t.Errorf("err = %v, want 403", err)
	}
}

func TestBalance_NoToken_Returns401(t *testing.T) {
	tc, _ := phase3Server(t, nil)
	if _, err := tc.Balance(context.Background(), ""); err == nil || !strings.Contains(err.Error(), "401") {
		t.Errorf("err = %v, want 401", err)
	}
}

func TestBankAccounts_HappyPath(t *testing.T) {
	tc, token := phase3Server(t, []string{"wallet:balance"})
	accounts, err := tc.BankAccounts(context.Background(), token)
	if err != nil {
		t.Fatalf("bank-accounts: %v", err)
	}
	if len(accounts) != 2 {
		t.Fatalf("got %d accounts, want 2", len(accounts))
	}
	if accounts[0].Status != "ACTIVE" {
		t.Errorf("first account status: %q", accounts[0].Status)
	}
}

func TestHistory_DefaultPagination_ReturnsEmpty(t *testing.T) {
	tc, token := phase3Server(t, []string{"transactions:history"})
	resp, err := tc.History(context.Background(), token, 0, 0)
	if err != nil {
		t.Fatalf("history: %v", err)
	}
	if len(resp.Transactions) != 0 || resp.Total != 0 {
		t.Errorf("phase 3 history should be empty: %+v", resp)
	}
}

func TestHistory_BadLimit_Returns400(t *testing.T) {
	tc, token := phase3Server(t, []string{"transactions:history"})
	u, _ := url.Parse(tc.BaseURL + "/api/v1/history?limit=abc")
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, u.String(), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := tc.HTTPClient.Do(req)
	if err != nil {
		t.Fatalf("do: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != 400 {
		t.Errorf("status = %d, want 400", resp.StatusCode)
	}
}

func TestGetPayment_AnyID_Returns404(t *testing.T) {
	tc, token := phase3Server(t, []string{"transactions:status"})
	body, status, err := tc.GetPayment(context.Background(), token, "pay_doesnotexist")
	if err != nil {
		t.Fatalf("get payment: %v", err)
	}
	if status != 404 {
		t.Errorf("status = %d, want 404", status)
	}
	var pd struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(body, &pd); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if pd.Type != "NOT_FOUND" {
		t.Errorf("type = %q, want NOT_FOUND", pd.Type)
	}
}

func TestRequestLog_PersistsForApiCalls(t *testing.T) {
	tc, token := phase3Server(t, []string{"wallet:balance"})
	if _, err := tc.Balance(context.Background(), token); err != nil {
		t.Fatalf("balance: %v", err)
	}
	ac := transport.NewAdminClient(tc.BaseURL)
	entries, err := ac.TailLogs(context.Background(), 10)
	if err != nil {
		t.Fatalf("tail logs: %v", err)
	}
	var found bool
	for _, e := range entries {
		if e.Method == "GET" && e.Path == "/api/v1/balance" {
			found = true
			if e.ClientID == "" {
				t.Errorf("ClientID empty on balance entry: %+v", e)
			}
		}
	}
	if !found {
		t.Errorf("balance request not in log: %+v", entries)
	}

	if len(entries) > 0 {
		shown, err := ac.ShowLog(context.Background(), entries[0].ID)
		if err != nil {
			t.Errorf("show log: %v", err)
		}
		if shown.ID != entries[0].ID {
			t.Errorf("show mismatch: %s vs %s", shown.ID, entries[0].ID)
		}
	}
}

// silence unused-import check for handlers in tests that pass through
// transport instead of importing the handler types directly.
var _ = handlers.BalanceResponse{}
