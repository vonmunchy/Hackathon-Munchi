package problemdetails

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// ContentType is the media type emitted with every ProblemDetails body.
// The spec's response declarations use application/json (the schema is
// ProblemDetails inside a generic application/json envelope), so we match
// that to stay spec-conformant for consumers that parse by media type.
// RFC 9457 recommends application/problem+json; deviation noted in the
// Phase 3 completion-notes for platform-team confirmation.
const ContentType = "application/json"

// Type is one of the catalogued error identifiers from D-016. Producers must
// pick from this set; ad-hoc strings are intentionally not permitted.
type Type string

// The full catalog from DECISIONS.md D-016. New values are added by amending
// DECISIONS.md and this list together.
const (
	TypeValidationError   Type = "VALIDATION_ERROR"
	TypeUnauthorized      Type = "UNAUTHORIZED"
	TypeForbidden         Type = "FORBIDDEN"
	TypeNotFound          Type = "NOT_FOUND"
	TypeRateLimited       Type = "RATE_LIMITED"
	TypeInsufficientFunds Type = "INSUFFICIENT_FUNDS"
	TypeInvalidRequest    Type = "INVALID_REQUEST"
	TypeInvalidGrant      Type = "INVALID_GRANT"
	TypeInvalidClient     Type = "INVALID_CLIENT"
	TypeInvalidScope      Type = "INVALID_SCOPE"
	// TypeInternalError covers 5xx server errors. Added after auditing
	// kodek which uses INTERNAL_ERROR for the same case
	// (apierror/problem.go:67-69).
	TypeInternalError Type = "INTERNAL_ERROR"
)

// FieldError matches spec/app.yaml#FieldError. It is appended to the Errors
// slice when an individual request field fails validation.
type FieldError struct {
	Name   string `json:"name"`
	Reason string `json:"reason"`
}

// ProblemDetails matches spec/app.yaml#ProblemDetails. The on-wire JSON keys
// are exactly type, detail, errors per the spec.
type ProblemDetails struct {
	Type   Type         `json:"type"`
	Detail string       `json:"detail,omitempty"`
	Errors []FieldError `json:"errors,omitempty"`
}

// New constructs a ProblemDetails with the supplied type and a fmt-formatted
// detail string. It is the most common producer in handler code.
func New(t Type, format string, args ...any) ProblemDetails {
	return ProblemDetails{Type: t, Detail: fmt.Sprintf(format, args...)}
}

// WithFieldError returns a copy with the given field error appended.
func (p ProblemDetails) WithFieldError(name, reason string) ProblemDetails {
	p.Errors = append(p.Errors, FieldError{Name: name, Reason: reason})
	return p
}

// Write serializes p to w with the given status code. Failures to encode are
// effectively impossible on a fixed-shape struct; any error from the
// underlying writer is silently dropped because nothing actionable can
// happen at this layer (the request has already failed).
func Write(w http.ResponseWriter, status int, p ProblemDetails) {
	w.Header().Set("Content-Type", ContentType)
	w.WriteHeader(status)
	// json.Marshal of a value-typed struct cannot fail.
	buf, _ := json.Marshal(p)
	_, _ = w.Write(buf)
}

// WriteNew is the one-liner combining New + Write. Most call sites use this.
func WriteNew(w http.ResponseWriter, status int, t Type, format string, args ...any) {
	Write(w, status, New(t, format, args...))
}
