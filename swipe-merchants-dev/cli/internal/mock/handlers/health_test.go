package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/handlers"
	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/store"
)

func newStore(t *testing.T, seed bool) *store.Store {
	t.Helper()
	st, err := store.Open(filepath.Join(t.TempDir(), "state.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })
	if seed {
		if _, err := st.SeedDefaults(); err != nil {
			t.Fatalf("seed: %v", err)
		}
	}
	return st
}

func TestAlive_AlwaysReturnsOK(t *testing.T) {
	t.Parallel()
	h := &handlers.Health{Store: newStore(t, false)}
	rec := httptest.NewRecorder()
	h.Alive(rec, httptest.NewRequest(http.MethodGet, "/health/alive", nil))
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
	var body map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body["status"] != "ok" {
		t.Errorf("status field = %v, want ok", body["status"])
	}
}

func TestReady_Seeded_Returns200(t *testing.T) {
	t.Parallel()
	h := &handlers.Health{Store: newStore(t, true)}
	rec := httptest.NewRecorder()
	h.Ready(rec, httptest.NewRequest(http.MethodGet, "/health/ready", nil))
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
	var body map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body["status"] != "ok" {
		t.Errorf("status = %v, want ok", body["status"])
	}
}

func TestReady_UnseededStore_Returns503(t *testing.T) {
	t.Parallel()
	h := &handlers.Health{Store: newStore(t, false)}
	rec := httptest.NewRecorder()
	h.Ready(rec, httptest.NewRequest(http.MethodGet, "/health/ready", nil))
	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("status = %d, want 503", rec.Code)
	}
	var body map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body["status"] != "unhealthy" {
		t.Errorf("status = %v, want unhealthy", body["status"])
	}
	checks, ok := body["checks"].(map[string]any)
	if !ok {
		t.Fatalf("checks is not an object: %T", body["checks"])
	}
	db, ok := checks["database"].(map[string]any)
	if !ok {
		t.Fatalf("database check missing or wrong shape: %v", checks["database"])
	}
	if db["status"] != "unhealthy" {
		t.Errorf("database status = %v, want unhealthy", db["status"])
	}
	if db["message"] == nil || db["message"] == "" {
		t.Errorf("expected non-empty message; got %v", db["message"])
	}
}

func TestReady_NilStore_Returns503(t *testing.T) {
	t.Parallel()
	h := &handlers.Health{Store: nil}
	rec := httptest.NewRecorder()
	h.Ready(rec, httptest.NewRequest(http.MethodGet, "/health/ready", nil))
	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("status = %d, want 503", rec.Code)
	}
}

func TestReady_ClosedStore_Returns503WithDriverError(t *testing.T) {
	t.Parallel()
	st := newStore(t, true)
	if err := st.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}
	// At this point store calls error with an open-DB-ish message that
	// is neither ErrNotFound nor nil — exercising the "other err" branch
	// of probeDatabase.
	h := &handlers.Health{Store: st}
	rec := httptest.NewRecorder()
	h.Ready(rec, httptest.NewRequest(http.MethodGet, "/health/ready", nil))
	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("status = %d, want 503", rec.Code)
	}
	var body map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	checks := body["checks"].(map[string]any)
	db := checks["database"].(map[string]any)
	msg, _ := db["message"].(string)
	if msg == "" || msg == "default merchant not seeded" {
		t.Errorf("message = %q, want a driver error", msg)
	}
}
