package auth

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/lestrrat-go/jwx/v2/jwk"
)

// JWKSHandler serves GET /.well-known/jwks.json. The published key set
// contains the single mock RSA public key plus its alg + kid metadata.
type JWKSHandler struct {
	SigningKey *SigningKey
}

// ServeHTTP implements http.Handler.
func (h *JWKSHandler) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	set := jwk.NewSet()
	// adding a typed jwk.Key to a fresh set cannot fail.
	_ = set.AddKey(h.SigningKey.Public)
	w.Header().Set("Content-Type", "application/json")
	// json.Marshal of a jwk.Set yields the canonical "keys" envelope.
	buf, _ := json.Marshal(set)
	_, _ = w.Write(buf)
}

// MetadataHandler serves GET /.well-known/oauth-authorization-server (RFC
// 8414). The handler renders the document fresh per request so the issuer
// can adapt to the bound address of the listener.
type MetadataHandler struct {
	IssuerURL       string
	ScopesSupported []string
}

// AuthorizationServerMetadata is the subset of RFC 8414 fields the mock
// publishes. Additional fields are allowed by the RFC; we publish what is
// required by D-026 + what a typical OAuth client expects.
type AuthorizationServerMetadata struct {
	Issuer                            string   `json:"issuer"`
	TokenEndpoint                     string   `json:"token_endpoint"`
	JWKSURI                           string   `json:"jwks_uri"`
	GrantTypesSupported               []string `json:"grant_types_supported"`
	TokenEndpointAuthMethodsSupported []string `json:"token_endpoint_auth_methods_supported"`
	ResponseTypesSupported            []string `json:"response_types_supported"`
	ScopesSupported                   []string `json:"scopes_supported,omitempty"`
}

// ServeHTTP implements http.Handler.
func (h *MetadataHandler) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	issuer := strings.TrimRight(h.IssuerURL, "/")
	doc := AuthorizationServerMetadata{
		Issuer:        issuer,
		TokenEndpoint: issuer + "/oauth2/token",
		JWKSURI:       issuer + "/.well-known/jwks.json",
		GrantTypesSupported: []string{
			"client_credentials",
		},
		TokenEndpointAuthMethodsSupported: []string{
			"client_secret_basic",
			"client_secret_post",
		},
		ResponseTypesSupported: []string{"token"},
		ScopesSupported:        h.ScopesSupported,
	}
	writeJSON(w, http.StatusOK, doc)
}

// AllSpecScopes returns the canonical scope list declared in the spec's
// oauth2 security schemes. The order matches the documentation order so
// metadata output is stable.
func AllSpecScopes() []string {
	return []string{
		"payments:link",
		"payments:contact",
		"payments:qr",
		"transactions:history",
		"transactions:status",
		"wallet:balance",
		"wallet:withdraw",
	}
}
