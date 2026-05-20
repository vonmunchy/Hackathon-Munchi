package transport_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/transport"
)

func TestHealthAlive_ReturnsParsedBody(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/health/alive" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	t.Cleanup(srv.Close)

	c := transport.NewClient(srv.URL)
	out, err := c.HealthAlive(context.Background())
	if err != nil {
		t.Fatalf("alive: %v", err)
	}
	if out["status"] != "ok" {
		t.Errorf("status = %v, want ok", out["status"])
	}
}

func TestHealthReady_ReturnsStatusCodeAndBody_OnUnhealthy(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status": "unhealthy",
			"checks": map[string]any{"database": map[string]any{"status": "unhealthy"}},
		})
	}))
	t.Cleanup(srv.Close)

	c := transport.NewClient(srv.URL)
	code, body, err := c.HealthReady(context.Background())
	if err != nil {
		t.Fatalf("ready: %v", err)
	}
	if code != http.StatusServiceUnavailable {
		t.Errorf("code = %d, want 503", code)
	}
	if body["status"] != "unhealthy" {
		t.Errorf("status = %v, want unhealthy", body["status"])
	}
}
