package auth

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"

	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/problemdetails"
	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/store"
)

// Middleware verifies bearer JWTs against the issuer's signing key and
// attaches a Principal to the request context. It implements every check
// listed in scaffold §5.4: signature, audience, expiry, client liveness.
type Middleware struct {
	SigningKey *SigningKey
	Store      *store.Store
	Clock      func() time.Time
}

// Wrap returns an http.Handler that runs the middleware before next. On
// any verification failure the handler responds with a ProblemDetails and
// next is not invoked.
//
// As a side effect, when the request context carries a PrincipalHolder
// (placed there by an outer middleware), the holder is populated so the
// outer middleware can observe the verified principal after handling.
func (m *Middleware) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		principal, perr := m.verify(r)
		if perr != nil {
			perr.write(w)
			return
		}
		if holder, ok := PrincipalHolderFromContext(r.Context()); ok {
			holder.Principal = principal
			holder.Set = true
		}
		next.ServeHTTP(w, r.WithContext(WithPrincipal(r.Context(), principal)))
	})
}

// verify performs the full set of bearer-token checks. The custom errType
// short-circuits the response selection in Wrap.
func (m *Middleware) verify(r *http.Request) (Principal, *authError) {
	raw, err := extractBearer(r.Header.Get("Authorization"))
	if err != nil {
		return Principal{}, newAuthErr(http.StatusUnauthorized, problemdetails.TypeUnauthorized, "%v", err)
	}
	pubKey, err := jwk.FromRaw(&m.SigningKey.Private.PublicKey)
	if err != nil {
		return Principal{}, newAuthErr(http.StatusInternalServerError, problemdetails.TypeUnauthorized, "load verification key: %v", err)
	}
	if err := pubKey.Set(jwk.KeyIDKey, m.SigningKey.KeyID); err != nil {
		return Principal{}, newAuthErr(http.StatusInternalServerError, problemdetails.TypeUnauthorized, "annotate verification key: %v", err)
	}

	tok, err := jwt.Parse([]byte(raw),
		jwt.WithKey(SigningAlgorithm, pubKey),
		jwt.WithClock(jwt.ClockFunc(m.now)),
		jwt.WithAudience(Audience),
		jwt.WithValidate(true),
	)
	if err != nil {
		return Principal{}, classifyParseError(err)
	}

	clientID, _ := tok.Get(ClaimClientID)
	cid, _ := clientID.(string)
	if cid == "" {
		// fall back to `sub` if the explicit claim is missing
		cid = tok.Subject()
	}
	if cid == "" {
		return Principal{}, newAuthErr(http.StatusUnauthorized, problemdetails.TypeUnauthorized, "token has no client identity")
	}

	c, err := m.Store.GetClient(cid)
	switch {
	case errors.Is(err, store.ErrNotFound):
		return Principal{}, newAuthErr(http.StatusUnauthorized, problemdetails.TypeUnauthorized, "client %s no longer exists", cid)
	case err != nil:
		return Principal{}, newAuthErr(http.StatusInternalServerError, problemdetails.TypeUnauthorized, "load client %s: %v", cid, err)
	}
	if !c.Enabled {
		return Principal{}, newAuthErr(http.StatusUnauthorized, problemdetails.TypeUnauthorized, "client %s is revoked", cid)
	}

	scopes := scopesFromClaims(tok)
	return PrincipalFromClient(c, scopes), nil
}

func (m *Middleware) now() time.Time {
	if m.Clock == nil {
		return time.Now().UTC()
	}
	return m.Clock().UTC()
}

// classifyParseError maps jwx parsing errors to ProblemDetails ones. The
// jwx package returns sentinel errors for the structural cases we care
// about; everything else falls back to a generic UNAUTHORIZED.
func classifyParseError(err error) *authError {
	switch {
	case errors.Is(err, jwt.ErrTokenExpired()):
		return newAuthErr(http.StatusUnauthorized, problemdetails.TypeUnauthorized, "token expired")
	case errors.Is(err, jwt.ErrInvalidAudience()):
		return newAuthErr(http.StatusUnauthorized, problemdetails.TypeUnauthorized, "token audience invalid")
	default:
		return newAuthErr(http.StatusUnauthorized, problemdetails.TypeUnauthorized, "invalid token: %v", err)
	}
}

// scopesFromClaims reads the space-separated scope claim into a slice. An
// absent or empty claim yields a nil slice.
func scopesFromClaims(tok jwt.Token) []string {
	raw, ok := tok.Get(ClaimScope)
	if !ok {
		return nil
	}
	s, ok := raw.(string)
	if !ok || s == "" {
		return nil
	}
	return strings.Fields(s)
}

// extractBearer pulls the token out of a "Bearer <jwt>" header value.
func extractBearer(header string) (string, error) {
	if header == "" {
		return "", fmt.Errorf("missing Authorization header")
	}
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
		return "", fmt.Errorf("malformed Authorization header")
	}
	return strings.TrimSpace(parts[1]), nil
}

// authError bundles the status + problemdetails fields for failed
// verifications so they can be returned as a single typed value before
// being written to the wire by the wrapper.
type authError struct {
	status int
	body   problemdetails.ProblemDetails
}

func newAuthErr(status int, t problemdetails.Type, format string, args ...any) *authError {
	return &authError{status: status, body: problemdetails.New(t, format, args...)}
}

func (e *authError) write(w http.ResponseWriter) {
	problemdetails.Write(w, e.status, e.body)
}

// RequireScope returns a middleware that fails with FORBIDDEN unless the
// authenticated principal has the named scope. Phase 3+ uses it to enforce
// per-operation coarse scopes; Phase 2's only consumer is whoami, which
// does not require a specific scope.
func RequireScope(scope string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			p, ok := PrincipalFromContext(r.Context())
			if !ok {
				problemdetails.WriteNew(w, http.StatusUnauthorized, problemdetails.TypeUnauthorized, "missing principal in request context")
				return
			}
			if !p.HasScope(scope) {
				problemdetails.WriteNew(w, http.StatusForbidden, problemdetails.TypeForbidden, "Token missing required scope: %s", scope)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RequireAnyScope returns a middleware that fails with FORBIDDEN unless
// the authenticated principal holds at least one of the listed scopes.
// Used by createPayment for the coarse pre-check (per D-013); the
// per-type fine-grained check happens in the handler.
func RequireAnyScope(scopes ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			p, ok := PrincipalFromContext(r.Context())
			if !ok {
				problemdetails.WriteNew(w, http.StatusUnauthorized, problemdetails.TypeUnauthorized, "missing principal in request context")
				return
			}
			for _, scope := range scopes {
				if p.HasScope(scope) {
					next.ServeHTTP(w, r)
					return
				}
			}
			problemdetails.WriteNew(w, http.StatusForbidden, problemdetails.TypeForbidden, "Token missing any of required scopes: %v", scopes)
		})
	}
}
