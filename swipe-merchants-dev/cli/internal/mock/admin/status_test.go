package admin_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/admin"
	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/store"
)

func TestStatusHandler_ReportsCountsFromSeededStore(t *testing.T) {
	t.Parallel()
	st, err := store.Open(filepath.Join(t.TempDir(), "state.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })
	if _, err := st.SeedDefaults(); err != nil {
		t.Fatalf("seed: %v", err)
	}

	h := &admin.StatusHandler{
		StartedAt:     time.Now().Add(-90 * time.Second),
		ListenAddress: "127.0.0.1:9999",
		WebhookURL:    "http://hook",
		Store:         st,
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/_admin/status", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var body admin.StatusResponse
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.ListenAddress != "127.0.0.1:9999" {
		t.Errorf("listen_address = %q", body.ListenAddress)
	}
	if body.WebhookURL != "http://hook" {
		t.Errorf("webhook_url = %q", body.WebhookURL)
	}
	if body.UptimeSeconds < 80 {
		t.Errorf("uptime_seconds = %d, want at least 80", body.UptimeSeconds)
	}
	if body.Counts.Merchants != 1 {
		t.Errorf("counts.merchants = %d, want 1", body.Counts.Merchants)
	}
	if body.Counts.Wallets != 1 {
		t.Errorf("counts.wallets = %d, want 1", body.Counts.Wallets)
	}
	if body.Counts.BankAccounts != 2 {
		t.Errorf("counts.bank_accounts = %d, want 2", body.Counts.BankAccounts)
	}
	if body.WebhookSecret == "" {
		t.Error("webhook_secret is empty")
	}
}

func TestStatusHandler_NilStore_DoesNotCrash(t *testing.T) {
	t.Parallel()
	h := &admin.StatusHandler{
		StartedAt:     time.Now(),
		ListenAddress: "127.0.0.1:0",
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/_admin/status", nil))
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
	var body admin.StatusResponse
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.WebhookSecret != "" {
		t.Errorf("webhook_secret = %q with nil store, want empty", body.WebhookSecret)
	}
	if body.Counts.Merchants != 0 {
		t.Errorf("counts.merchants = %d with nil store, want 0", body.Counts.Merchants)
	}
}
