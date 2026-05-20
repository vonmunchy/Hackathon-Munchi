package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
)

// SigningAlgorithm is the JWT signature algorithm used for all mock-issued
// tokens, fixed at RS256 by D-011.
const SigningAlgorithm = jwa.RS256

// SigningKeyBits is the RSA modulus size for the signing key, fixed at 2048
// by D-011.
const SigningKeyBits = 2048

// SigningKeyFilename is the on-disk filename of the persisted signing key
// (PKCS#1 PEM) inside the configured keys directory.
const SigningKeyFilename = "signing-key.pem"

// SigningKey bundles the loaded RSA key with its JWK representation,
// pre-built for /.well-known/jwks.json responses. The KeyID is a stable
// thumbprint of the public key so callers can correlate tokens with the
// publishing JWKS entry.
type SigningKey struct {
	Private *rsa.PrivateKey
	Public  jwk.Key
	KeyID   string
}

// LoadOrCreateSigningKey returns the persisted key from
// <keysDir>/signing-key.pem, generating and writing a fresh 2048-bit RSA
// keypair if the file does not exist. The keysDir is created with 0o700
// permissions if missing. Both the private file and parent dir are owned
// by the running user and not world-readable.
func LoadOrCreateSigningKey(keysDir string) (*SigningKey, error) {
	if err := os.MkdirAll(keysDir, 0o700); err != nil {
		return nil, fmt.Errorf("ensure keys dir %s: %w", keysDir, err)
	}
	path := filepath.Join(keysDir, SigningKeyFilename)
	priv, err := loadRSAFromPEMFile(path)
	switch {
	case err == nil:
		return buildSigningKey(priv)
	case errors.Is(err, os.ErrNotExist):
		// fall through to generation
	default:
		return nil, fmt.Errorf("load signing key %s: %w", path, err)
	}

	priv, err = rsa.GenerateKey(rand.Reader, SigningKeyBits)
	if err != nil {
		return nil, fmt.Errorf("generate signing key: %w", err)
	}
	if err := writeRSAToPEMFile(path, priv); err != nil {
		return nil, fmt.Errorf("persist signing key %s: %w", path, err)
	}
	return buildSigningKey(priv)
}

// buildSigningKey wraps an RSA private key into a SigningKey, attaching the
// derived JWK and its thumbprint as the key id.
func buildSigningKey(priv *rsa.PrivateKey) (*SigningKey, error) {
	pubKey, err := jwk.FromRaw(&priv.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("build jwk: %w", err)
	}
	thumb, err := pubKey.Thumbprint(thumbprintAlgo)
	if err != nil {
		return nil, fmt.Errorf("jwk thumbprint: %w", err)
	}
	kid := encodeKeyID(thumb)
	_ = pubKey.Set(jwk.KeyIDKey, kid)
	_ = pubKey.Set(jwk.AlgorithmKey, SigningAlgorithm)
	_ = pubKey.Set(jwk.KeyUsageKey, "sig")
	return &SigningKey{
		Private: priv,
		Public:  pubKey,
		KeyID:   kid,
	}, nil
}

func loadRSAFromPEMFile(path string) (*rsa.PrivateKey, error) {
	buf, err := os.ReadFile(path) // #nosec G304 -- path is built from a configured keysDir, not user input
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(buf)
	if block == nil {
		return nil, fmt.Errorf("decode pem: no block")
	}
	switch block.Type {
	case "RSA PRIVATE KEY":
		return x509.ParsePKCS1PrivateKey(block.Bytes)
	case "PRIVATE KEY":
		key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("parse pkcs8: %w", err)
		}
		rsaKey, ok := key.(*rsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("pkcs8 key is not RSA")
		}
		return rsaKey, nil
	default:
		return nil, fmt.Errorf("unexpected pem type %q", block.Type)
	}
}

func writeRSAToPEMFile(path string, priv *rsa.PrivateKey) error {
	der := x509.MarshalPKCS1PrivateKey(priv)
	buf := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: der})
	return os.WriteFile(path, buf, 0o600)
}
