package transport_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/admin"
	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/transport"
)

func TestAdminClient_Status_DecodesFullPayload(t *testing.T) {
	t.Parallel()
	want := admin.StatusResponse{
		StartedAt:     time.Date(2026, 5, 10, 12, 0, 0, 0, time.UTC),
		UptimeSeconds: 42,
		ListenAddress: "127.0.0.1:8080",
		WebhookURL:    "http://hook",
		WebhookSecret: "whsec_abc",
		Counts:        admin.Counts{Merchants: 1, BankAccounts: 2, Wallets: 1},
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/_admin/status" {
			t.Errorf("path = %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(want)
	}))
	t.Cleanup(srv.Close)

	c := transport.NewAdminClient(srv.URL)
	got, err := c.Status(context.Background())
	if err != nil {
		t.Fatalf("status: %v", err)
	}
	if got.ListenAddress != want.ListenAddress {
		t.Errorf("listen = %q, want %q", got.ListenAddress, want.ListenAddress)
	}
	if got.UptimeSeconds != want.UptimeSeconds {
		t.Errorf("uptime = %d, want %d", got.UptimeSeconds, want.UptimeSeconds)
	}
	if got.Counts.Merchants != 1 || got.Counts.BankAccounts != 2 {
		t.Errorf("counts = %+v", got.Counts)
	}
	if got.WebhookSecret != "whsec_abc" {
		t.Errorf("secret = %q", got.WebhookSecret)
	}
}

func TestAdminClient_Status_BubblesUpHTTPError(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, `{"type":"FORBIDDEN","detail":"_admin is localhost-only"}`, http.StatusForbidden)
	}))
	t.Cleanup(srv.Close)

	c := transport.NewAdminClient(srv.URL)
	_, err := c.Status(context.Background())
	if err == nil {
		t.Fatalf("expected error on 403; got nil")
	}
}
