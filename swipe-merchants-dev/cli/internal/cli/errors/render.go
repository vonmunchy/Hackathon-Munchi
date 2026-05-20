package errors

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"
)

// problemPattern matches the JSON envelope embedded in transport-level
// errors (e.g. `status 403: {"type":"FORBIDDEN","detail":"..."}`).
var problemPattern = regexp.MustCompile(`\{\s*"type"\s*:\s*"[^"]+".*\}`)

// problem mirrors the on-wire ProblemDetails shape. Only the fields the
// CLI cares about are deserialized; extra fields are ignored.
type problem struct {
	Type   string `json:"type"`
	Detail string `json:"detail"`
	Errors []struct {
		Name   string `json:"name"`
		Reason string `json:"reason"`
	} `json:"errors,omitempty"`
}

// Render writes err to w with as much structure as it can extract:
//   - When err's chain contains a JSON ProblemDetails body, the top line
//     becomes "Error: <type>: <detail>" and field errors print indented.
//   - When --verbose, the underlying error chain is appended verbatim.
//   - Otherwise, the plain "Error: <message>" form is preserved.
func Render(w io.Writer, err error, verbose bool) {
	if err == nil {
		return
	}
	msg := err.Error()
	if p, ok := extractProblem(msg); ok {
		_, _ = fmt.Fprintf(w, "Error: %s\n  %s\n", p.Type, p.Detail)
		for _, fe := range p.Errors {
			_, _ = fmt.Fprintf(w, "  - %s: %s\n", fe.Name, fe.Reason)
		}
		if verbose {
			renderChain(w, err)
		}
		hint := hintFor(p.Type)
		if hint != "" {
			_, _ = fmt.Fprintln(w, "  hint: "+hint)
		}
		return
	}
	_, _ = fmt.Fprintf(w, "Error: %s\n", msg)
	if verbose {
		renderChain(w, err)
	}
}

// extractProblem hunts for a JSON ProblemDetails substring in msg. The
// transport returns errors like
//
//	"balance: status 403: {\"type\":\"FORBIDDEN\",\"detail\":\"...\"}"
//
// — we pull the JSON out and parse it.
func extractProblem(msg string) (problem, bool) {
	match := problemPattern.FindString(msg)
	if match == "" {
		return problem{}, false
	}
	var p problem
	if err := json.Unmarshal([]byte(match), &p); err != nil {
		return problem{}, false
	}
	if p.Type == "" {
		return problem{}, false
	}
	return p, true
}

// renderChain prints the unwrap-chain in indented form below the main
// error line. Used under --verbose.
func renderChain(w io.Writer, err error) {
	_, _ = fmt.Fprintln(w, "  chain:")
	for cur := err; cur != nil; cur = errors.Unwrap(cur) {
		_, _ = fmt.Fprintf(w, "    %s\n", strings.TrimSpace(cur.Error()))
	}
}

// hintFor returns an actionable suggestion for common error types. The
// strings are deliberately short — they nudge the user toward a fix
// without claiming exhaustive coverage.
func hintFor(t string) string {
	switch t {
	case "UNAUTHORIZED":
		return "run `swipe auth login` or `swipe auth whoami` to inspect the cached token"
	case "FORBIDDEN":
		return "update the client's scopes via `swipe keys update --scopes ...`"
	case "INVALID_CLIENT":
		return "the client id or secret is wrong; create a new pair with `swipe keys create`"
	case "INVALID_SCOPE":
		return "request only scopes the client was issued (see `swipe keys show <id>`)"
	case "RATE_LIMITED":
		return "wait the Retry-After interval or disable the rate_limit_burst scenario"
	case "INSUFFICIENT_FUNDS":
		return "check `swipe wallet balance` or disable the insufficient_funds scenario"
	case "NOT_FOUND":
		return ""
	case "VALIDATION_ERROR":
		return "validate locally with `swipe spec validate --operation <opId>`"
	default:
		return ""
	}
}
