package mock

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

// White-box tests for lifecycle.go and the isLoopback helper.

func TestPIDFile_Write_AddrFileWriteFails(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	pidPath := filepath.Join(dir, "mock.pid")
	addrPath := filepath.Join(dir, "mock.addr")
	// Make addr path a directory so os.WriteFile cannot create the file
	// at that path while pidfile creation succeeds.
	if err := os.Mkdir(addrPath, 0o700); err != nil {
		t.Fatalf("mkdir addr-as-dir: %v", err)
	}
	p := PIDFile{PIDPath: pidPath, AddrPath: addrPath}
	if err := p.Write("addr"); err == nil {
		t.Errorf("expected addr write error; got nil")
	}
}

func TestPIDFile_Write_PIDFileWriteFails(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	pidPath := filepath.Join(dir, "mock.pid")
	addrPath := filepath.Join(dir, "mock.addr")
	// Make pidfile path a directory so os.WriteFile fails first.
	if err := os.Mkdir(pidPath, 0o700); err != nil {
		t.Fatalf("mkdir pid-as-dir: %v", err)
	}
	p := PIDFile{PIDPath: pidPath, AddrPath: addrPath}
	if err := p.Write("addr"); err == nil {
		t.Errorf("expected pidfile write error; got nil")
	}
}

func TestPIDFile_Remove_DirInsteadOfFile_Errors(t *testing.T) {
	t.Parallel()
	// os.Remove on a non-empty directory returns an error which Remove()
	// propagates rather than swallowing.
	dir := t.TempDir()
	pidPath := filepath.Join(dir, "as-dir")
	if err := os.Mkdir(pidPath, 0o700); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	// Put a file inside so the dir is non-empty -> os.Remove fails.
	if err := os.WriteFile(filepath.Join(pidPath, "child"), []byte("x"), 0o600); err != nil {
		t.Fatalf("seed child: %v", err)
	}
	p := PIDFile{PIDPath: pidPath, AddrPath: filepath.Join(dir, "missing")}
	if err := p.Remove(); err == nil {
		t.Errorf("expected error removing non-empty directory")
	}
}

func TestPIDFile_Write_MkdirParent_Errors(t *testing.T) {
	t.Parallel()
	// Create a *file* and try to root the pidfile under it (treating it as
	// a directory). MkdirAll will fail with ENOTDIR.
	root := filepath.Join(t.TempDir(), "not-a-dir")
	if err := os.WriteFile(root, []byte(""), 0o600); err != nil {
		t.Fatalf("seed file: %v", err)
	}
	p := PIDFile{
		PIDPath:  filepath.Join(root, "child", "mock.pid"),
		AddrPath: filepath.Join(root, "child", "mock.addr"),
	}
	if err := p.Write("addr"); err == nil {
		t.Errorf("expected MkdirAll error; got nil")
	}
}

func TestPIDFile_Read_PIDFileMalformed_Errors(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	pidPath := filepath.Join(dir, "mock.pid")
	addrPath := filepath.Join(dir, "mock.addr")
	if err := os.WriteFile(pidPath, []byte("not-a-number"), 0o600); err != nil {
		t.Fatalf("seed: %v", err)
	}
	if err := os.WriteFile(addrPath, []byte("addr"), 0o600); err != nil {
		t.Fatalf("seed: %v", err)
	}
	p := PIDFile{PIDPath: pidPath, AddrPath: addrPath}
	if _, _, err := p.Read(); err == nil {
		t.Errorf("expected parse error")
	}
}

func TestPIDFile_Read_AddrMissing_ReturnsPidAndIsNotExist(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	pidPath := filepath.Join(dir, "mock.pid")
	addrPath := filepath.Join(dir, "mock.addr")
	if err := os.WriteFile(pidPath, []byte("12345"), 0o600); err != nil {
		t.Fatalf("seed: %v", err)
	}
	p := PIDFile{PIDPath: pidPath, AddrPath: addrPath}
	pid, addr, err := p.Read()
	if !errors.Is(err, os.ErrNotExist) {
		t.Errorf("err = %v, want os.ErrNotExist", err)
	}
	if pid != 12345 {
		t.Errorf("pid = %d, want 12345", pid)
	}
	if addr != "" {
		t.Errorf("addr = %q, want empty", addr)
	}
}

func TestIsLoopback_ParsesVariants(t *testing.T) {
	t.Parallel()
	cases := map[string]bool{
		"127.0.0.1:8080":                     true,
		"127.1.2.3:0":                        true,
		"[::1]:8080":                         true,
		"localhost:1":                        true,
		"LocalHost:1":                        true, // case-insensitive
		"10.0.0.1:80":                        false,
		"[2001:db8::1]:80":                   false,
		"":                                   false,
		"definitely-not-an-ip-or-hostname:1": false,
	}
	for remote, want := range cases {
		req := httptest.NewRequest(http.MethodGet, "http://x/", nil)
		req.RemoteAddr = remote
		if got := isLoopback(req); got != want {
			t.Errorf("isLoopback(%q) = %v, want %v", remote, got, want)
		}
	}
}
