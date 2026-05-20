package transport

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

// White-box tests for the unexported helpers in client.go.

func TestNewRequest_NonMarshalableBody_Errors(t *testing.T) {
	t.Parallel()
	c := NewClient("http://127.0.0.1:1")
	// `chan` cannot be JSON-marshaled — hits newRequest's marshal-error branch.
	_, err := c.newRequest(context.Background(), http.MethodPost, "/x", make(chan int))
	if err == nil {
		t.Fatalf("expected marshal error; got nil")
	}
}

func TestNewRequest_MalformedURL_Errors(t *testing.T) {
	t.Parallel()
	c := NewClient("http://[::1") // unbalanced bracket
	_, err := c.newRequest(context.Background(), http.MethodGet, "/x", nil)
	if err == nil {
		t.Fatalf("expected new-request error; got nil")
	}
}

func TestDo_NilOut_ReturnsNilWithoutDecoding(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("ignored body"))
	}))
	t.Cleanup(srv.Close)

	c := NewClient(srv.URL)
	if err := c.do(context.Background(), http.MethodGet, "/x", nil, nil); err != nil {
		t.Errorf("do with nil out: %v", err)
	}
}

func TestDo_204NoContent_ReturnsNilEvenWithOutPtr(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(srv.Close)

	c := NewClient(srv.URL)
	var out map[string]any
	if err := c.do(context.Background(), http.MethodGet, "/x", nil, &out); err != nil {
		t.Errorf("do with 204: %v", err)
	}
}

func TestNewRequest_WithBody_SetsContentType(t *testing.T) {
	t.Parallel()
	c := NewClient("http://127.0.0.1:1")
	req, err := c.newRequest(context.Background(), http.MethodPost, "/x", map[string]string{"k": "v"})
	if err != nil {
		t.Fatalf("newRequest: %v", err)
	}
	if got := req.Header.Get("Content-Type"); got != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", got)
	}
}
