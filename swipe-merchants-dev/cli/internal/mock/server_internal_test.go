package mock

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// White-box tests for unexported helpers in the mock package. Public
// behavior is exercised through the integration tests in server_test.go.

func TestLocalhostOnly_AllowsLoopbackRemote(t *testing.T) {
	t.Parallel()
	called := false
	h := localhostOnly(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	cases := []string{"127.0.0.1:1234", "[::1]:1234", "localhost:1234"}
	for _, addr := range cases {
		called = false
		req := httptest.NewRequest(http.MethodGet, "http://x/_admin/status", nil)
		req.RemoteAddr = addr
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if !called {
			t.Errorf("loopback %q: handler was not called; status=%d body=%s", addr, rec.Code, rec.Body.String())
		}
	}
}

func TestLocalhostOnly_RejectsNonLoopbackRemote(t *testing.T) {
	t.Parallel()
	h := localhostOnly(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		t.Fatalf("handler should not have been called")
	}))

	cases := []string{"10.0.0.5:1234", "192.168.1.10:1234", "8.8.8.8:1234", "[2001:db8::1]:1234"}
	for _, addr := range cases {
		req := httptest.NewRequest(http.MethodGet, "http://x/_admin/status", nil)
		req.RemoteAddr = addr
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusForbidden {
			t.Errorf("remote %q: status = %d, want 403", addr, rec.Code)
		}
	}
}

func TestLocalhostOnly_RejectsMissingRemote(t *testing.T) {
	t.Parallel()
	h := localhostOnly(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		t.Fatalf("handler should not have been called")
	}))
	req := httptest.NewRequest(http.MethodGet, "http://x/_admin/status", nil)
	req.RemoteAddr = "" // empty / unknown
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Errorf("status = %d, want 403", rec.Code)
	}
}
