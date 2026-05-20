package auth

import (
	"context"
	"slices"

	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/store"
)

// contextKey is an unexported type used as the key for the principal value
// stored on request contexts. The unexported type ensures no other package
// can collide with it.
type contextKey struct{}

// holderKey is the context key for an out-of-band Principal holder. It is
// used by middleware that wants to read the verified principal AFTER
// inner middleware has populated it (e.g. the request logger), since
// contexts are immutable in Go and a child context with the principal is
// only visible inside the inner stack frame.
type holderKey struct{}

// PrincipalHolder is a tiny pointer-bearing struct passed through the
// request context so middleware running before the auth pass can still
// observe the verified principal after the handler returns.
type PrincipalHolder struct {
	Principal Principal
	Set       bool
}

// WithPrincipalHolder attaches an empty holder to ctx. The auth middleware
// fills it in as a side effect of verification.
func WithPrincipalHolder(ctx context.Context, h *PrincipalHolder) context.Context {
	return context.WithValue(ctx, holderKey{}, h)
}

// PrincipalHolderFromContext returns the holder, if any.
func PrincipalHolderFromContext(ctx context.Context) (*PrincipalHolder, bool) {
	h, ok := ctx.Value(holderKey{}).(*PrincipalHolder)
	return h, ok
}

// Principal carries the verified identity of an authenticated request:
// the client it represents, the merchant the client belongs to, and the
// scopes granted to its token.
type Principal struct {
	ClientID   string
	MerchantID string
	Scopes     []string
}

// HasScope reports whether the principal's token includes the named scope.
func (p Principal) HasScope(s string) bool {
	return slices.Contains(p.Scopes, s)
}

// WithPrincipal returns a new context carrying the given principal.
func WithPrincipal(ctx context.Context, p Principal) context.Context {
	return context.WithValue(ctx, contextKey{}, p)
}

// PrincipalFromContext returns the principal attached to ctx and whether
// one was present. Handlers behind the auth middleware should always find
// a principal; helper code that may run outside the middleware (e.g.
// admin endpoints) should branch on ok.
func PrincipalFromContext(ctx context.Context) (Principal, bool) {
	p, ok := ctx.Value(contextKey{}).(Principal)
	return p, ok
}

// PrincipalFromClient is a convenience for assembling a Principal directly
// from a store.Client, used by middleware after token verification.
func PrincipalFromClient(c store.Client, scopes []string) Principal {
	return Principal{
		ClientID:   c.ID,
		MerchantID: c.MerchantID,
		Scopes:     append([]string(nil), scopes...),
	}
}
