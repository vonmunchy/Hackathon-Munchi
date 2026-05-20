package problemdetails_test

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/problemdetails"
)

func TestWriteNew_SetsHeadersAndBody(t *testing.T) {
	rec := httptest.NewRecorder()
	problemdetails.WriteNew(rec, 401, problemdetails.TypeUnauthorized, "token %s", "expired")

	if got := rec.Code; got != 401 {
		t.Fatalf("status = %d, want 401", got)
	}
	if got := rec.Header().Get("Content-Type"); got != problemdetails.ContentType {
		t.Fatalf("content-type = %q, want %q", got, problemdetails.ContentType)
	}
	var body problemdetails.ProblemDetails
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("body unmarshal: %v", err)
	}
	if body.Type != problemdetails.TypeUnauthorized {
		t.Errorf("type = %q, want UNAUTHORIZED", body.Type)
	}
	if body.Detail != "token expired" {
		t.Errorf("detail = %q", body.Detail)
	}
}

func TestWithFieldError_Appends(t *testing.T) {
	p := problemdetails.New(problemdetails.TypeValidationError, "bad request").
		WithFieldError("amount", "must be positive").
		WithFieldError("currency", "unsupported")
	if len(p.Errors) != 2 {
		t.Fatalf("errors len = %d, want 2", len(p.Errors))
	}
	if p.Errors[0].Name != "amount" || p.Errors[1].Name != "currency" {
		t.Errorf("errors order wrong: %+v", p.Errors)
	}
}
