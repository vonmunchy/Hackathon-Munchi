package mock_test

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock"
	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/auth"
	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/store"
)

func newTestServer(t *testing.T) (*httptest.Server, *store.Store) {
	t.Helper()
	dir := t.TempDir()
	st, err := store.Open(filepath.Join(dir, "state.db"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })
	if _, err := st.SeedDefaults(); err != nil {
		t.Fatalf("seed: %v", err)
	}
	key, err := auth.LoadOrCreateSigningKey(filepath.Join(dir, "keys"))
	if err != nil {
		t.Fatalf("signing key: %v", err)
	}
	srv, err := mock.New(mock.Config{
		ListenAddr: "127.0.0.1:0",
		Store:      st,
		Logger:     slog.New(slog.NewJSONHandler(io.Discard, nil)),
		StartedAt:  time.Now().UTC(),
		SigningKey: key,
	})
	if err != nil {
		t.Fatalf("build server: %v", err)
	}
	if err := srv.Bind(); err != nil {
		t.Fatalf("bind: %v", err)
	}
	addr := srv.Listener().Addr().String()
	ctx, cancel := context.WithCancel(context.Background())
	startErr := make(chan error, 1)
	go func() { startErr <- srv.Serve(ctx) }()
	t.Cleanup(func() {
		cancel()
		select {
		case <-startErr:
		case <-time.After(2 * time.Second):
			t.Errorf("server did not stop in time")
		}
	})
	return &httptest.Server{URL: "http://" + addr}, st
}

func TestHealthAlive_Returns200_OK(t *testing.T) {
	t.Parallel()
	ts, _ := newTestServer(t)

	resp, err := http.Get(ts.URL + "/health/alive") //nolint:gosec,noctx
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}
	var body map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body["status"] != "ok" {
		t.Errorf("status field = %v, want ok", body["status"])
	}
}

func TestHealthReady_AfterSeed_Returns200_OK(t *testing.T) {
	t.Parallel()
	ts, _ := newTestServer(t)

	resp, err := http.Get(ts.URL + "/health/ready") //nolint:gosec,noctx
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}
	var body map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body["status"] != "ok" {
		t.Errorf("status field = %v, want ok", body["status"])
	}
	checks, ok := body["checks"].(map[string]any)
	if !ok {
		t.Fatalf("checks field is not an object: %T", body["checks"])
	}
	if checks["database"] == nil {
		t.Errorf("expected database check, got %v", checks)
	}
}

func TestAdminStatus_FromLoopback_Returns200(t *testing.T) {
	t.Parallel()
	ts, _ := newTestServer(t)

	resp, err := http.Get(ts.URL + "/_admin/status") //nolint:gosec,noctx
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}
	var body map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	counts, ok := body["counts"].(map[string]any)
	if !ok {
		t.Fatalf("counts is not an object: %T", body["counts"])
	}
	if counts["merchants"].(float64) != 1 {
		t.Errorf("merchants count = %v, want 1", counts["merchants"])
	}
	if counts["bank_accounts"].(float64) != 2 {
		t.Errorf("bank_accounts count = %v, want 2", counts["bank_accounts"])
	}
	if !strings.HasPrefix(body["webhook_secret"].(string), "whsec_") {
		t.Errorf("webhook_secret = %v, want whsec_ prefix", body["webhook_secret"])
	}
}
