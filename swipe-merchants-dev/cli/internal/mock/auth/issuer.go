package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwt"

	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/problemdetails"
	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/store"
)

// DefaultTokenTTL is the default access-token lifetime (scaffold §5.3, D-014
// open question — 1h per the DECISIONS.md fallback table for Phase 2).
const DefaultTokenTTL = time.Hour

// TokenResponse mirrors RFC 6749 §5.1: the JSON body returned by a
// successful token endpoint response. Scope is intentionally a single
// space-delimited string.
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
	Scope       string `json:"scope,omitempty"`
}

// Issuer issues RS256 access tokens via the client_credentials grant. It is
// stateless apart from the SigningKey and the Store it consults for client
// verification.
type Issuer struct {
	// IssuerURL is the value of the `iss` claim (e.g. http://localhost:8080).
	// Trailing slashes are accepted; the issuer normalizes them away.
	IssuerURL string
	// SigningKey provides the RSA private key + kid for header propagation.
	SigningKey *SigningKey
	// Store is consulted to verify client credentials at token-grant time.
	Store *store.Store
	// TTL is the token lifetime. Zero falls back to DefaultTokenTTL.
	TTL time.Duration
	// Clock is the time source used for iat/exp claims; nil means time.Now.
	Clock func() time.Time
	// TTLOverride, when non-nil, is consulted at the start of each Issue
	// call. Returning (d, true) replaces the configured TTL for this mint
	// only. Used by the token_short_ttl scenario.
	TTLOverride func(ctx context.Context) (time.Duration, bool)
	// ScopeFilter, when non-nil, is applied to the granted scope slice
	// after intersection with the client's permitted scopes. Used by the
	// scope_downgrade scenario.
	ScopeFilter func(ctx context.Context, granted []string) []string
}

// Issue mints a JWT for the supplied client with the given requested scopes,
// returning the encoded token + the chosen TTL. Requested scopes are
// intersected with the client's permitted scopes; an empty requested set
// grants all of the client's scopes (RFC 6749 §3.3). Returns ErrInvalidScope
// when the intersection rejects a requested scope, ErrUnauthorizedClient
// when the client is disabled, and other errors via wrapping.
//
// The ctx is passed to TTLOverride / ScopeFilter hooks so scenarios can
// record their effects on the request log.
func (i *Issuer) Issue(ctx context.Context, c store.Client, requested []string) (string, time.Duration, []string, error) {
	if !c.Enabled {
		return "", 0, nil, ErrUnauthorizedClient
	}
	granted, err := intersectScopes(c.Scopes, requested)
	if err != nil {
		return "", 0, nil, err
	}
	if i.ScopeFilter != nil {
		granted = i.ScopeFilter(ctx, granted)
	}

	now := i.now()
	ttl := i.ttl()
	if i.TTLOverride != nil {
		if overridden, ok := i.TTLOverride(ctx); ok && overridden > 0 {
			ttl = overridden
		}
	}
	tok, err := jwt.NewBuilder().
		Issuer(i.normalizedIssuer()).
		Subject(c.ID).
		Audience([]string{Audience}).
		IssuedAt(now).
		Expiration(now.Add(ttl)).
		Claim(ClaimClientID, c.ID).
		Claim(ClaimMerchantID, c.MerchantID).
		Claim(ClaimScope, strings.Join(granted, " ")).
		Build()
	if err != nil {
		return "", 0, nil, fmt.Errorf("build token: %w", err)
	}
	signed, err := signTokenWithKID(tok, i.SigningKey)
	if err != nil {
		return "", 0, nil, err
	}
	return string(signed), ttl, granted, nil
}

// HandleTokenEndpoint implements POST /oauth2/token (RFC 6749 §4.4 client
// credentials grant). Only application/x-www-form-urlencoded bodies are
// accepted. Errors follow RFC 6749 §5.2 (oauth-style) wrapped inside
// ProblemDetails so the wire shape matches D-015.
func (i *Issuer) HandleTokenEndpoint(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		problemdetails.WriteNew(w, http.StatusMethodNotAllowed, problemdetails.TypeInvalidRequest, "method not allowed")
		return
	}
	if err := r.ParseForm(); err != nil {
		problemdetails.WriteNew(w, http.StatusBadRequest, problemdetails.TypeInvalidRequest, "parse form: %v", err)
		return
	}

	if got := r.PostForm.Get("grant_type"); got != "client_credentials" {
		problemdetails.WriteNew(w, http.StatusBadRequest, problemdetails.TypeInvalidGrant, "unsupported grant_type %q; only client_credentials is supported", got)
		return
	}

	clientID, clientSecret, ok := clientCredentials(r)
	if !ok {
		problemdetails.WriteNew(w, http.StatusUnauthorized, problemdetails.TypeInvalidClient, "missing client_id / client_secret")
		return
	}

	c, err := i.Store.VerifyClientSecret(clientID, clientSecret)
	switch {
	case errors.Is(err, store.ErrInvalidSecret), errors.Is(err, store.ErrNotFound):
		problemdetails.WriteNew(w, http.StatusUnauthorized, problemdetails.TypeInvalidClient, "invalid client credentials")
		return
	case err != nil:
		problemdetails.WriteNew(w, http.StatusInternalServerError, problemdetails.TypeInvalidRequest, "verify client: %v", err)
		return
	}

	requested := splitScopeParam(r.PostForm.Get("scope"))
	signed, ttl, granted, err := i.Issue(r.Context(), c, requested)
	switch {
	case errors.Is(err, ErrInvalidScopeRequest):
		problemdetails.WriteNew(w, http.StatusBadRequest, problemdetails.TypeInvalidScope, "%v", err)
		return
	case errors.Is(err, ErrUnauthorizedClient):
		problemdetails.WriteNew(w, http.StatusUnauthorized, problemdetails.TypeInvalidClient, "client is disabled")
		return
	case err != nil:
		problemdetails.WriteNew(w, http.StatusInternalServerError, problemdetails.TypeInvalidRequest, "issue token: %v", err)
		return
	}

	resp := TokenResponse{
		AccessToken: signed,
		TokenType:   "Bearer",
		ExpiresIn:   int64(ttl.Seconds()),
		Scope:       strings.Join(granted, " "),
	}
	writeJSON(w, http.StatusOK, resp)
}

func (i *Issuer) now() time.Time {
	if i.Clock == nil {
		return time.Now().UTC()
	}
	return i.Clock().UTC()
}

func (i *Issuer) ttl() time.Duration {
	if i.TTL <= 0 {
		return DefaultTokenTTL
	}
	return i.TTL
}

func (i *Issuer) normalizedIssuer() string {
	return strings.TrimRight(i.IssuerURL, "/")
}

// clientCredentials extracts client_id and client_secret from either HTTP
// Basic auth (RFC 6749 §2.3.1) or POST body parameters. The Basic form is
// the canonical client_credentials transport; the body form is supported
// for ergonomics (curl one-liners, etc.).
func clientCredentials(r *http.Request) (string, string, bool) {
	if id, secret, ok := r.BasicAuth(); ok && id != "" && secret != "" {
		return id, secret, true
	}
	id := r.PostForm.Get("client_id")
	secret := r.PostForm.Get("client_secret")
	if id == "" || secret == "" {
		return "", "", false
	}
	return id, secret, true
}

func splitScopeParam(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Fields(s)
	if len(parts) == 0 {
		return nil
	}
	return parts
}
