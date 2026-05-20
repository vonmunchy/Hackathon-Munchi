package auth

import (
	"errors"
	"fmt"
	"sort"
)

// ErrInvalidScopeRequest is returned when a token request asks for a scope
// the client is not permitted to use.
var ErrInvalidScopeRequest = errors.New("auth: requested scope outside client grant")

// ErrUnauthorizedClient is returned when the resolved client is disabled
// (revoked). Distinct from ErrInvalidScopeRequest so callers can map to the
// right HTTP / OAuth error type.
var ErrUnauthorizedClient = errors.New("auth: client is disabled")

// intersectScopes returns the granted scope set for a token, given the
// client's permitted scopes and the (possibly empty) requested set. If
// requested is empty, the client receives all of its permitted scopes
// (RFC 6749 §3.3). Otherwise every requested scope must be in the permitted
// set or the call returns ErrInvalidScopeRequest.
//
// The returned slice is sorted alphabetically so token claim ordering is
// deterministic for tests and logs.
func intersectScopes(permitted, requested []string) ([]string, error) {
	if len(requested) == 0 {
		out := append([]string(nil), permitted...)
		sort.Strings(out)
		return out, nil
	}
	allowed := make(map[string]struct{}, len(permitted))
	for _, p := range permitted {
		allowed[p] = struct{}{}
	}
	out := make([]string, 0, len(requested))
	seen := make(map[string]struct{}, len(requested))
	for _, s := range requested {
		if _, ok := allowed[s]; !ok {
			return nil, fmt.Errorf("%w: %q", ErrInvalidScopeRequest, s)
		}
		if _, dup := seen[s]; dup {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	sort.Strings(out)
	return out, nil
}
