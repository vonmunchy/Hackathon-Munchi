package transport

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// CachedToken is the on-disk shape of ~/.swipe/token.json (D-025).
type CachedToken struct {
	AccessToken string    `json:"access_token"`
	TokenType   string    `json:"token_type"`
	ExpiresAt   time.Time `json:"expires_at"`
	Scope       string    `json:"scope,omitempty"`
	ClientID    string    `json:"client_id"`
	IssuerURL   string    `json:"issuer_url,omitempty"`
}

// IsExpired reports whether the token's lifetime is up. A 30s safety margin
// keeps the CLI from sending a token that is about to expire mid-request.
func (t CachedToken) IsExpired(now time.Time) bool {
	if t.ExpiresAt.IsZero() {
		return false
	}
	return !now.Add(30 * time.Second).Before(t.ExpiresAt)
}

// LoadCachedToken reads the token cache at path. Returns os.ErrNotExist
// when the file is absent.
func LoadCachedToken(path string) (CachedToken, error) {
	buf, err := os.ReadFile(path) // #nosec G304 -- path comes from config.DefaultPaths()
	if err != nil {
		return CachedToken{}, err
	}
	var out CachedToken
	if err := json.Unmarshal(buf, &out); err != nil {
		return CachedToken{}, fmt.Errorf("decode token cache %s: %w", path, err)
	}
	return out, nil
}

// SaveCachedToken writes t to path with 0o600 permissions, creating the
// parent directory if needed.
func SaveCachedToken(path string, t CachedToken) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("ensure parent dir: %w", err)
	}
	buf, err := json.MarshalIndent(t, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal cached token: %w", err)
	}
	if err := os.WriteFile(path, buf, 0o600); err != nil {
		return fmt.Errorf("write token cache %s: %w", path, err)
	}
	return nil
}

// ClearCachedToken removes the token cache file. Absence is not an error.
func ClearCachedToken(path string) error {
	err := os.Remove(path)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove token cache: %w", err)
	}
	return nil
}
