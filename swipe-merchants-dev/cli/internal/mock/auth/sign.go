package auth

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jws"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

// signTokenWithKID signs the supplied jwt.Token with the issuer's signing
// key and attaches the kid header so verifiers can pick the matching JWKS
// entry. Returns the compact-serialized signed JWT.
func signTokenWithKID(tok jwt.Token, key *SigningKey) ([]byte, error) {
	headers := jws.NewHeaders()
	if err := headers.Set(jws.KeyIDKey, key.KeyID); err != nil {
		return nil, fmt.Errorf("set kid: %w", err)
	}
	signingKey, err := jwk.FromRaw(key.Private)
	if err != nil {
		return nil, fmt.Errorf("wrap signing key: %w", err)
	}
	if err := signingKey.Set(jwk.KeyIDKey, key.KeyID); err != nil {
		return nil, fmt.Errorf("set signing kid: %w", err)
	}
	signed, err := jwt.Sign(tok, jwt.WithKey(SigningAlgorithm, signingKey, jws.WithProtectedHeaders(headers)))
	if err != nil {
		return nil, fmt.Errorf("sign token: %w", err)
	}
	return signed, nil
}

// writeJSON serializes v as application/json at the given status. Failures
// are unreachable on the value-typed structs the auth package emits, so
// the encode error is intentionally not surfaced — there's nothing the
// caller can do at this point in the response cycle.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
