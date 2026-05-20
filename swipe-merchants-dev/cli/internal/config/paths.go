package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// Paths bundles every file-system location the swipe binary touches at
// runtime. They are derived once at start-up from the user's home directory
// and an optional override root.
//
// Per DECISIONS.md D-025:
//   - mock listen port:    8080 (handled at the CLI/server layer, not here)
//   - config file:         ~/.swipe/config.yaml
//   - mock state:          ~/.swipe/mock/state.db
//   - mock signing keys:   ~/.swipe/mock/keys/
//   - cached token:        ~/.swipe/token.json
type Paths struct {
	// Root is the swipe-managed directory (default ~/.swipe).
	Root string
	// ConfigFile is the swipe CLI configuration YAML.
	ConfigFile string
	// MockDir holds all mock-server runtime state.
	MockDir string
	// MockStateDB is the BoltDB file backing the mock.
	MockStateDB string
	// MockKeysDir holds the RSA signing keypair (Phase 2+).
	MockKeysDir string
	// MockPIDFile records the running mock-server PID for `mock stop`.
	MockPIDFile string
	// MockSocketFile records the listener address for the running mock so
	// other CLI invocations (mock stop / mock status) can locate it without
	// guessing a port.
	MockSocketFile string
	// TokenCache is the cached OAuth access token (Phase 2+).
	TokenCache string
}

// DefaultPaths returns the canonical Paths under the user's home directory.
// If $HOME cannot be resolved it returns an error rather than silently
// falling back, since persistence is a contract not a nice-to-have.
func DefaultPaths() (Paths, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return Paths{}, fmt.Errorf("resolve home directory: %w", err)
	}
	return PathsFromRoot(filepath.Join(home, ".swipe")), nil
}

// PathsFromRoot builds Paths under the given swipe root directory. Tests use
// this to point at a temp directory; production should use DefaultPaths.
func PathsFromRoot(root string) Paths {
	mock := filepath.Join(root, "mock")
	return Paths{
		Root:           root,
		ConfigFile:     filepath.Join(root, "config.yaml"),
		MockDir:        mock,
		MockStateDB:    filepath.Join(mock, "state.db"),
		MockKeysDir:    filepath.Join(mock, "keys"),
		MockPIDFile:    filepath.Join(mock, "mock.pid"),
		MockSocketFile: filepath.Join(mock, "mock.addr"),
		TokenCache:     filepath.Join(root, "token.json"),
	}
}

// EnsureDirs creates Root, MockDir, and MockKeysDir with 0o700 permissions.
// Existing directories are left as-is (no chmod). Safe to call repeatedly.
func (p Paths) EnsureDirs() error {
	for _, dir := range []string{p.Root, p.MockDir, p.MockKeysDir} {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			return fmt.Errorf("create %s: %w", dir, err)
		}
	}
	return nil
}
