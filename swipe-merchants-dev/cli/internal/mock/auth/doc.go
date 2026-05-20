// Package auth implements the mock's OAuth2 issuer (Phase 2). It owns the
// RSA-2048 signing keypair (D-011), issues RS256 JWTs via the
// client_credentials grant, publishes the JWKS and RFC 8414 metadata
// documents, and provides middleware that verifies bearer tokens for
// handlers under /api/v1.
package auth
