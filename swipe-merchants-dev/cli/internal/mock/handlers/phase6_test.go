package handlers_test

import (
	"context"
	"io"
	"log/slog"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock"
	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/auth"
	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/handlers"
	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/store"
	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/transport"
)

// scenarioServer builds a Phase-6-aware mock + a transport.Client + a
// bearer token whose client has the listed scopes. The PaymentTTL is
// short so payment_transitions / payment_stuck_pending finish quickly
// (or stay stuck, respectively).
func scenarioServer(t *testing.T, scopes []string) (*transport.Client, *store.Store, string) {
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
		PaymentTTL: 200 * time.Millisecond,
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

	c, secret, err := st.CreateClient(store.DefaultMerchantID, "p6", scopes)
	if err != nil {
		t.Fatalf("client: %v", err)
	}
	base := "http://" + srv.Listener().Addr().String()
	tc := transport.NewClient(base)
	tok, err := tc.IssueToken(context.Background(), c.ID, secret, nil)
	if err != nil {
		t.Fatalf("token: %v", err)
	}
	return tc, st, tok.AccessToken
}

func TestScenario_Random5xx_FailsRequests(t *testing.T) {
	tc, st, token := scenarioServer(t, []string{"wallet:balance"})
	if err := st.EnableScenario("random_5xx", map[string]string{"rate": "1.0"}); err != nil {
		t.Fatalf("enable: %v", err)
	}
	_, err := tc.Balance(context.Background(), token)
	if err == nil || !strings.Contains(err.Error(), "scenario random_5xx") {
		t.Errorf("err = %v, want random_5xx", err)
	}
	// Verify the request log captured the scenario annotation.
	ac := transport.NewAdminClient(tc.BaseURL)
	entries, err := ac.TailLogs(context.Background(), 5)
	if err != nil {
		t.Fatalf("tail logs: %v", err)
	}
	var found bool
	for _, e := range entries {
		if e.Path == "/api/v1/balance" && len(e.Scenarios) > 0 {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("no scenarios annotation on /api/v1/balance log entry")
	}
}

func TestScenario_InsufficientFunds_Returns402(t *testing.T) {
	tc, st, token := scenarioServer(t, []string{"payments:qr"})
	if err := st.EnableScenario("insufficient_funds", map[string]string{"rate": "1.0"}); err != nil {
		t.Fatalf("enable: %v", err)
	}
	_, err := tc.CreatePayment(context.Background(), token, handlers.CreatePaymentRequest{
		Amount: 100, Currency: "MVR", Type: "QR",
	})
	if err == nil || !strings.Contains(err.Error(), "402") {
		t.Errorf("err = %v, want 402", err)
	}
}

func TestScenario_TokenShortTTL_OverridesIssuer(t *testing.T) {
	tc, st, _ := scenarioServer(t, []string{"wallet:balance"})
	if err := st.EnableScenario("token_short_ttl", map[string]string{"ttl": "5s"}); err != nil {
		t.Fatalf("enable: %v", err)
	}
	// Re-create a fresh client so the new TTL applies to a brand-new token.
	c, secret, _ := st.CreateClient(store.DefaultMerchantID, "p6b", []string{"wallet:balance"})
	tok, err := tc.IssueToken(context.Background(), c.ID, secret, nil)
	if err != nil {
		t.Fatalf("token: %v", err)
	}
	if tok.ExpiresIn != 5 {
		t.Errorf("expires_in = %d, want 5", tok.ExpiresIn)
	}
}

func TestScenario_ScopeDowngrade_StripsRequestedScope(t *testing.T) {
	tc, st, _ := scenarioServer(t, []string{"wallet:balance", "payments:qr"})
	if err := st.EnableScenario("scope_downgrade", map[string]string{"remove": "wallet:balance"}); err != nil {
		t.Fatalf("enable: %v", err)
	}
	c, secret, _ := st.CreateClient(store.DefaultMerchantID, "p6c", []string{"wallet:balance", "payments:qr"})
	tok, err := tc.IssueToken(context.Background(), c.ID, secret, nil)
	if err != nil {
		t.Fatalf("token: %v", err)
	}
	if strings.Contains(tok.Scope, "wallet:balance") {
		t.Errorf("scope = %q, expected wallet:balance stripped", tok.Scope)
	}
	if !strings.Contains(tok.Scope, "payments:qr") {
		t.Errorf("scope = %q, expected payments:qr retained", tok.Scope)
	}
}

func TestScenario_PaymentTransitions_TargetsEXPIRED(t *testing.T) {
	tc, st, token := scenarioServer(t, []string{"payments:qr", "transactions:status"})
	if err := st.EnableScenario("payment_transitions", map[string]string{"to": "EXPIRED"}); err != nil {
		t.Fatalf("enable: %v", err)
	}
	created, err := tc.CreatePayment(context.Background(), token, handlers.CreatePaymentRequest{
		Amount: 10, Currency: "MVR", Type: "QR",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		body, status, err := tc.GetPayment(context.Background(), token, created.ID)
		if err == nil && status == 200 && strings.Contains(string(body), `"EXPIRED"`) {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Errorf("payment did not transition to EXPIRED")
}

func TestScenario_PaymentStuckPending_StaysPending(t *testing.T) {
	tc, st, token := scenarioServer(t, []string{"payments:qr", "transactions:status"})
	if err := st.EnableScenario("payment_stuck_pending", nil); err != nil {
		t.Fatalf("enable: %v", err)
	}
	created, err := tc.CreatePayment(context.Background(), token, handlers.CreatePaymentRequest{
		Amount: 10, Currency: "MVR", Type: "QR",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	// Wait longer than the configured PaymentTTL — the payment should
	// still be PENDING because the scenario set TransitionAt far out.
	time.Sleep(600 * time.Millisecond)
	body, status, err := tc.GetPayment(context.Background(), token, created.ID)
	if err != nil || status != 200 {
		t.Fatalf("get: %v status=%d", err, status)
	}
	if !strings.Contains(string(body), `"PENDING"`) {
		t.Errorf("expected still PENDING, got %s", body)
	}
}
