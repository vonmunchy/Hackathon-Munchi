package transport_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/transport"
)

func TestHealthAlive_HTTPErrorStatus_BubblesError(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, `{"type":"INTERNAL"}`, http.StatusInternalServerError)
	}))
	t.Cleanup(srv.Close)

	c := transport.NewClient(srv.URL)
	_, err := c.HealthAlive(context.Background())
	if err == nil {
		t.Fatalf("expected error from 500 response; got nil")
	}
	if !strings.Contains(err.Error(), "500") {
		t.Errorf("err = %q, want containing 500", err)
	}
}

func TestHealthAlive_BadJSON_ReturnsDecodeError(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("this is not json"))
	}))
	t.Cleanup(srv.Close)

	c := transport.NewClient(srv.URL)
	_, err := c.HealthAlive(context.Background())
	if err == nil {
		t.Fatalf("expected decode error; got nil")
	}
}

func TestHealthReady_BadJSON_ReturnsParseError(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("{not json"))
	}))
	t.Cleanup(srv.Close)

	c := transport.NewClient(srv.URL)
	_, _, err := c.HealthReady(context.Background())
	if err == nil {
		t.Fatalf("expected parse error; got nil")
	}
}

func TestHealthAlive_TransportFailure_ReturnsError(t *testing.T) {
	t.Parallel()
	c := transport.NewClient("http://127.0.0.1:1") // port 1 reliably refuses
	_, err := c.HealthAlive(context.Background())
	if err == nil {
		t.Fatalf("expected transport error")
	}
}

func TestHealthReady_TransportFailure_ReturnsError(t *testing.T) {
	t.Parallel()
	c := transport.NewClient("http://127.0.0.1:1")
	_, _, err := c.HealthReady(context.Background())
	if err == nil {
		t.Fatalf("expected transport error")
	}
}

func TestHealthAlive_MalformedBaseURL_ReturnsRequestError(t *testing.T) {
	t.Parallel()
	c := transport.NewClient("http://[::1") // unbalanced brackets — http.NewRequest rejects
	_, err := c.HealthAlive(context.Background())
	if err == nil {
		t.Fatalf("expected new-request error")
	}
}

func TestHealthAlive_NilContext_ReturnsRequestError(t *testing.T) {
	t.Parallel()
	c := transport.NewClient("http://127.0.0.1:1")
	//nolint:staticcheck // intentionally nil to drive the error branch
	_, err := c.HealthAlive(nil)
	if err == nil {
		t.Fatalf("expected error from nil context")
	}
}
