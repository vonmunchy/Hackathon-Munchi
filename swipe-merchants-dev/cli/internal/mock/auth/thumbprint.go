package auth

import (
	"crypto"
	"encoding/base64"
)

// thumbprintAlgo is the hash used for JWK thumbprints (RFC 7638). SHA-256 is
// the conventional choice and gives a stable, short key id.
const thumbprintAlgo = crypto.SHA256

// encodeKeyID renders a raw thumbprint as URL-safe base64 without padding,
// the canonical form for `kid` values in JWT headers and JWKS responses.
func encodeKeyID(raw []byte) string {
	return base64.RawURLEncoding.EncodeToString(raw)
}
