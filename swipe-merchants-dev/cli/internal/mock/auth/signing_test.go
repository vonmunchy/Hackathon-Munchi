package auth_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/auth"
)

func TestLoadOrCreateSigningKey_CreatesOnFirstCall(t *testing.T) {
	dir := t.TempDir()
	keysDir := filepath.Join(dir, "keys")
	k, err := auth.LoadOrCreateSigningKey(keysDir)
	if err != nil {
		t.Fatalf("first load: %v", err)
	}
	if k.Private == nil || k.Public == nil {
		t.Fatalf("missing key material: %+v", k)
	}
	if k.KeyID == "" {
		t.Fatalf("empty kid")
	}
	info, err := os.Stat(filepath.Join(keysDir, auth.SigningKeyFilename))
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Errorf("file mode = %v, want 0o600", info.Mode().Perm())
	}
}

func TestLoadOrCreateSigningKey_ReusesPersistedKey(t *testing.T) {
	dir := t.TempDir()
	keysDir := filepath.Join(dir, "keys")
	first, err := auth.LoadOrCreateSigningKey(keysDir)
	if err != nil {
		t.Fatalf("first: %v", err)
	}
	second, err := auth.LoadOrCreateSigningKey(keysDir)
	if err != nil {
		t.Fatalf("second: %v", err)
	}
	if first.KeyID != second.KeyID {
		t.Errorf("kid changed across loads: %q vs %q", first.KeyID, second.KeyID)
	}
}

func TestLoadOrCreateSigningKey_RejectsCorruptPEM(t *testing.T) {
	dir := t.TempDir()
	keysDir := filepath.Join(dir, "keys")
	if err := os.MkdirAll(keysDir, 0o700); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(keysDir, auth.SigningKeyFilename), []byte("garbage"), 0o600); err != nil {
		t.Fatalf("write garbage: %v", err)
	}
	if _, err := auth.LoadOrCreateSigningKey(keysDir); err == nil {
		t.Errorf("expected error on garbage pem")
	}
}
