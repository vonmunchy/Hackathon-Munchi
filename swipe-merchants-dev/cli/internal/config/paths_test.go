package config_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/config"
)

func TestPathsFromRoot_DerivesEverythingFromRoot(t *testing.T) {
	t.Parallel()
	root := "/tmp/swipe-test"
	p := config.PathsFromRoot(root)

	wants := map[string]string{
		"config":  filepath.Join(root, "config.yaml"),
		"db":      filepath.Join(root, "mock", "state.db"),
		"keys":    filepath.Join(root, "mock", "keys"),
		"pidfile": filepath.Join(root, "mock", "mock.pid"),
		"addr":    filepath.Join(root, "mock", "mock.addr"),
		"token":   filepath.Join(root, "token.json"),
	}
	if p.ConfigFile != wants["config"] {
		t.Errorf("ConfigFile = %s, want %s", p.ConfigFile, wants["config"])
	}
	if p.MockStateDB != wants["db"] {
		t.Errorf("MockStateDB = %s, want %s", p.MockStateDB, wants["db"])
	}
	if p.MockKeysDir != wants["keys"] {
		t.Errorf("MockKeysDir = %s, want %s", p.MockKeysDir, wants["keys"])
	}
	if p.MockPIDFile != wants["pidfile"] {
		t.Errorf("MockPIDFile = %s, want %s", p.MockPIDFile, wants["pidfile"])
	}
	if p.MockSocketFile != wants["addr"] {
		t.Errorf("MockSocketFile = %s, want %s", p.MockSocketFile, wants["addr"])
	}
	if p.TokenCache != wants["token"] {
		t.Errorf("TokenCache = %s, want %s", p.TokenCache, wants["token"])
	}
}

func TestEnsureDirs_CreatesEveryRequiredDirectory(t *testing.T) {
	t.Parallel()
	root := filepath.Join(t.TempDir(), "swipe")
	p := config.PathsFromRoot(root)
	if err := p.EnsureDirs(); err != nil {
		t.Fatalf("ensure: %v", err)
	}
	for _, dir := range []string{p.Root, p.MockDir, p.MockKeysDir} {
		if !strings.HasPrefix(dir, root) {
			t.Errorf("dir %s is not rooted under %s", dir, root)
		}
	}
}

func TestEnsureDirs_NonDirectoryParent_Errors(t *testing.T) {
	t.Parallel()
	// Make a file at the path the root wants to occupy. MkdirAll fails
	// because there's a non-directory in the way.
	parent := t.TempDir()
	rootPath := filepath.Join(parent, "swipe")
	if err := os.WriteFile(rootPath, []byte(""), 0o600); err != nil {
		t.Fatalf("seed file: %v", err)
	}
	p := config.PathsFromRoot(rootPath)
	if err := p.EnsureDirs(); err == nil {
		t.Errorf("expected error when parent is a file; got nil")
	}
}

func TestDefaultPaths_HoneorsHOME(t *testing.T) {
	// Mutates HOME env so cannot run in parallel.
	t.Setenv("HOME", "/some/test/home")
	t.Setenv("USERPROFILE", "/some/test/home")
	p, err := config.DefaultPaths()
	if err != nil {
		t.Fatalf("default paths: %v", err)
	}
	if p.Root != "/some/test/home/.swipe" {
		t.Errorf("root = %q, want %q", p.Root, "/some/test/home/.swipe")
	}
}

func TestDefaultPaths_NoHome_ReturnsError(t *testing.T) {
	// os.UserHomeDir consults HOME on unix and USERPROFILE/HOMEDRIVE+HOMEPATH
	// on windows. Unset all of them to force the failure path.
	t.Setenv("HOME", "")
	t.Setenv("USERPROFILE", "")
	t.Setenv("HOMEDRIVE", "")
	t.Setenv("HOMEPATH", "")
	if _, err := config.DefaultPaths(); err == nil {
		t.Errorf("expected error when HOME is unset")
	}
}
