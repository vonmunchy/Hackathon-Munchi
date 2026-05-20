package handlers_test

import (
	"context"
	"io"
	"log/slog"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock"
	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/auth"
	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/handlers"
	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/store"
	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/transport"
)

// phase4Server boots a fresh mock with a short PaymentTTL so TTL
// transitions land within seconds rather than the production 60s.
func phase4Server(t *testing.T, scopes []string, ttl time.Duration) (*transport.Client, string) {
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
		PaymentTTL: ttl,
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

	c, secret, err := st.CreateClient(store.DefaultMerchantID, "p4", scopes)
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

func TestCreatePayment_HappyPath_ReturnsPending(t *testing.T) {
	tc, token := phase4Server(t, []string{"payments:qr"}, time.Minute)
	got, err := tc.CreatePayment(context.Background(), token, handlers.CreatePaymentRequest{
		Amount: 100, Currency: "MVR", Type: "QR", Description: "test",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if !strings.HasPrefix(got.ID, "pay_") {
		t.Errorf("id = %q, want pay_ prefix", got.ID)
	}
	if got.Status != "PENDING" {
		t.Errorf("status = %q, want PENDING", got.Status)
	}
	if got.ShortCode == "" {
		t.Errorf("short_code empty")
	}
	if got.QRData == "" {
		t.Errorf("qr_data empty for QR")
	}
}

func TestCreatePayment_WrongTypeScope_Returns403(t *testing.T) {
	// Client has payments:qr but tries to create CONTACT.
	tc, token := phase4Server(t, []string{"payments:qr"}, time.Minute)
	_, err := tc.CreatePayment(context.Background(), token, handlers.CreatePaymentRequest{
		Amount: 100, Currency: "MVR", Type: "CONTACT", RecipientVPA: "v@p",
	})
	if err == nil || !strings.Contains(err.Error(), "403") {
		t.Errorf("err = %v, want 403", err)
	}
}

func TestCreatePayment_NoPaymentScope_Returns403(t *testing.T) {
	tc, token := phase4Server(t, []string{"wallet:balance"}, time.Minute)
	_, err := tc.CreatePayment(context.Background(), token, handlers.CreatePaymentRequest{
		Amount: 100, Currency: "MVR", Type: "QR",
	})
	if err == nil || !strings.Contains(err.Error(), "403") {
		t.Errorf("err = %v, want 403", err)
	}
}

func TestCreatePayment_ContactWithoutVPA_Returns400(t *testing.T) {
	tc, token := phase4Server(t, []string{"payments:contact"}, time.Minute)
	_, err := tc.CreatePayment(context.Background(), token, handlers.CreatePaymentRequest{
		Amount: 100, Currency: "MVR", Type: "CONTACT",
	})
	if err == nil || !strings.Contains(err.Error(), "400") {
		t.Errorf("err = %v, want 400", err)
	}
}

func TestCreatePayment_LinkType_PopulatesPaymentURL(t *testing.T) {
	tc, token := phase4Server(t, []string{"payments:link"}, time.Minute)
	got, err := tc.CreatePayment(context.Background(), token, handlers.CreatePaymentRequest{
		Amount: 50, Currency: "USD", Type: "LINK",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if got.PaymentURL == "" {
		t.Errorf("payment_url empty for LINK")
	}
	if got.QRData != "" {
		t.Errorf("qr_data set for LINK: %q", got.QRData)
	}
}

func TestPaymentTTL_TransitionsToCompleted(t *testing.T) {
	tc, token := phase4Server(t,
		[]string{"payments:qr", "transactions:status"},
		200*time.Millisecond,
	)
	created, err := tc.CreatePayment(context.Background(), token, handlers.CreatePaymentRequest{
		Amount: 10, Currency: "MVR", Type: "QR",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	// Wait for TTL + worker tick to land.
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		body, status, err := tc.GetPayment(context.Background(), token, created.ID)
		if err == nil && status == 200 && strings.Contains(string(body), `"COMPLETED"`) {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Errorf("payment did not transition to COMPLETED within deadline")
}

func TestStreamPayment_DeliversTerminalTransition(t *testing.T) {
	tc, token := phase4Server(t,
		[]string{"payments:qr", "transactions:status"},
		300*time.Millisecond,
	)
	created, err := tc.CreatePayment(context.Background(), token, handlers.CreatePaymentRequest{
		Amount: 5, Currency: "MVR", Type: "QR",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var mu sync.Mutex
	var statuses []string
	err = tc.WatchPayment(ctx, token, created.ID, func(ev transport.StreamEvent) bool {
		mu.Lock()
		statuses = append(statuses, ev.Status)
		mu.Unlock()
		return ev.Status == "PENDING"
	})
	if err != nil {
		t.Fatalf("watch: %v", err)
	}
	mu.Lock()
	defer mu.Unlock()
	if len(statuses) < 2 {
		t.Fatalf("expected >=2 statuses, got %v", statuses)
	}
	if statuses[0] != "PENDING" {
		t.Errorf("first status = %q, want PENDING", statuses[0])
	}
	if statuses[len(statuses)-1] != "COMPLETED" {
		t.Errorf("last status = %q, want COMPLETED", statuses[len(statuses)-1])
	}
}

func TestCreatePayout_HappyPath(t *testing.T) {
	tc, token := phase4Server(t, []string{"wallet:withdraw", "wallet:balance"}, time.Minute)
	// pick a bank account first
	accounts, err := tc.BankAccounts(context.Background(), token)
	if err != nil {
		t.Fatalf("bank accounts: %v", err)
	}
	if len(accounts) == 0 {
		t.Fatal("no seeded bank accounts")
	}
	got, err := tc.CreatePayout(context.Background(), token, handlers.CreatePayoutRequest{
		Amount: 25, BankAccountID: accounts[0].ID,
	})
	if err != nil {
		t.Fatalf("create payout: %v", err)
	}
	if !strings.HasPrefix(got.ID, "pyo_") {
		t.Errorf("id = %q, want pyo_ prefix", got.ID)
	}
	if got.Status != "PENDING" {
		t.Errorf("status = %q", got.Status)
	}
}

func TestCreatePayout_InsufficientFunds_Returns402(t *testing.T) {
	tc, token := phase4Server(t, []string{"wallet:withdraw", "wallet:balance"}, time.Minute)
	accounts, err := tc.BankAccounts(context.Background(), token)
	if err != nil {
		t.Fatalf("bank accounts: %v", err)
	}
	// Pick the USD account (650.00 available) and try to withdraw way more.
	var usdID string
	for _, a := range accounts {
		if a.Currency == "USD" {
			usdID = a.ID
			break
		}
	}
	if usdID == "" {
		t.Fatal("no USD account")
	}
	_, err = tc.CreatePayout(context.Background(), token, handlers.CreatePayoutRequest{
		Amount: 1_000_000, BankAccountID: usdID,
	})
	if err == nil || !strings.Contains(err.Error(), "402") {
		t.Errorf("err = %v, want 402", err)
	}
}

func TestHistory_PendingPayment_ExcludedUntilCompleted(t *testing.T) {
	// History only carries paid payment requests (D-handlers convention).
	// Right after createPayment the payment is PENDING and must not show.
	// After the TTL worker transitions it to COMPLETED a transaction is
	// projected and the row appears.
	tc, token := phase4Server(t,
		[]string{"payments:qr", "transactions:history"},
		50*time.Millisecond,
	)
	_, err := tc.CreatePayment(context.Background(), token, handlers.CreatePaymentRequest{
		Amount: 7, Currency: "MVR", Type: "QR",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	resp, err := tc.History(context.Background(), token, 0, 0)
	if err != nil {
		t.Fatalf("history (pending): %v", err)
	}
	if resp.Total != 0 {
		t.Errorf("total = %d, want 0 while payment is PENDING; transactions=%+v", resp.Total, resp.Transactions)
	}

	got := pollHistory(t, tc, token, 1, 3*time.Second)
	if len(got.Transactions) != 1 || got.Transactions[0].Type != "PAYMENT" {
		t.Errorf("after TTL transactions = %+v", got.Transactions)
	}
	if got.Transactions[0].Status != "COMPLETED" {
		t.Errorf("status = %q, want COMPLETED", got.Transactions[0].Status)
	}
}

// pollHistory polls /api/v1/history until total >= want or deadline expires.
// Used by tests that exercise the TTL-driven transaction projection.
func pollHistory(t *testing.T, tc *transport.Client, token string, want int, timeout time.Duration) handlers.HistoryResponse {
	t.Helper()
	deadline := time.Now().Add(timeout)
	var last handlers.HistoryResponse
	for time.Now().Before(deadline) {
		resp, err := tc.History(context.Background(), token, 0, 0)
		if err != nil {
			t.Fatalf("history poll: %v", err)
		}
		last = resp
		if resp.Total >= want {
			return resp
		}
		time.Sleep(75 * time.Millisecond)
	}
	t.Fatalf("history did not reach total>=%d within %v; last total=%d", want, timeout, last.Total)
	return last
}
