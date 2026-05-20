package validator_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/validator"
	specpkg "github.com/BML-Digital/swipe-merchants-dev/cli/internal/spec"
)

func TestNew_LoadsEmbeddedSpec(t *testing.T) {
	loaded, err := specpkg.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if _, err := validator.New(loaded.RawYAML()); err != nil {
		t.Fatalf("new validator: %v", err)
	}
}

func TestMiddleware_OutOfSpecPath_PassesThrough(t *testing.T) {
	v := buildValidator(t)
	handler := v.Middleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/_admin/status", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("admin path should pass through, got %d", rec.Code)
	}
}

func TestMiddleware_HealthAlive_Passes(t *testing.T) {
	v := buildValidator(t)
	handler := v.Middleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodGet, "/health/alive", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("alive: status = %d", rec.Code)
	}
}

func TestMiddleware_PaymentMalformedBody_Returns400(t *testing.T) {
	v := buildValidator(t)
	handler := v.Middleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		t.Error("handler should not be reached on validation failure")
	}))
	req := httptest.NewRequest(http.MethodPost, "/api/v1/payments", bytes.NewReader([]byte(`{"amount":"not-a-number"}`)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
}

func buildValidator(t *testing.T) *validator.Validator {
	t.Helper()
	loaded, err := specpkg.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	v, err := validator.New(loaded.RawYAML())
	if err != nil {
		t.Fatalf("new: %v", err)
	}
	return v
}
