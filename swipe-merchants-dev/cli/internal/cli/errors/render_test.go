package errors_test

import (
	"bytes"
	stderrors "errors"
	"fmt"
	"strings"
	"testing"

	clierrors "github.com/BML-Digital/swipe-merchants-dev/cli/internal/cli/errors"
)

func TestRender_PlainError_PrintsErrorPrefix(t *testing.T) {
	var buf bytes.Buffer
	clierrors.Render(&buf, stderrors.New("boom"), false)
	if got := buf.String(); !strings.HasPrefix(got, "Error: boom") {
		t.Errorf("got %q", got)
	}
}

func TestRender_ProblemDetails_ExtractsTypeAndDetail(t *testing.T) {
	err := stderrors.New(`balance: balance: status 403: {"type":"FORBIDDEN","detail":"Token missing required scope: wallet:balance"}`)
	var buf bytes.Buffer
	clierrors.Render(&buf, err, false)
	out := buf.String()
	if !strings.Contains(out, "Error: FORBIDDEN") {
		t.Errorf("missing type line: %q", out)
	}
	if !strings.Contains(out, "Token missing required scope") {
		t.Errorf("missing detail: %q", out)
	}
	if !strings.Contains(out, "hint:") {
		t.Errorf("missing hint: %q", out)
	}
}

func TestRender_Verbose_IncludesChain(t *testing.T) {
	inner := stderrors.New(`status 401: {"type":"UNAUTHORIZED","detail":"token expired"}`)
	wrapped := fmt.Errorf("whoami: %w", inner)
	var buf bytes.Buffer
	clierrors.Render(&buf, wrapped, true)
	out := buf.String()
	if !strings.Contains(out, "chain:") {
		t.Errorf("verbose output missing chain: %q", out)
	}
	if !strings.Contains(out, "whoami:") {
		t.Errorf("verbose output missing wrapper: %q", out)
	}
}

func TestRender_ProblemDetailsWithFieldErrors_PrintsThem(t *testing.T) {
	err := stderrors.New(`status 400: {"type":"VALIDATION_ERROR","detail":"bad","errors":[{"name":"amount","reason":"must be positive"}]}`)
	var buf bytes.Buffer
	clierrors.Render(&buf, err, false)
	out := buf.String()
	if !strings.Contains(out, "- amount: must be positive") {
		t.Errorf("missing field error: %q", out)
	}
}

func TestRender_NilError_DoesNothing(t *testing.T) {
	var buf bytes.Buffer
	clierrors.Render(&buf, nil, true)
	if buf.Len() != 0 {
		t.Errorf("expected no output, got %q", buf.String())
	}
}
