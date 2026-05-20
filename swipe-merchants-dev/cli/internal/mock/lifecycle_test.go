package mock_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock"
)

func newPIDFile(t *testing.T) mock.PIDFile {
	t.Helper()
	dir := t.TempDir()
	return mock.NewPIDFile(filepath.Join(dir, "mock.pid"), filepath.Join(dir, "mock.addr"))
}

func TestPIDFile_WriteRead_RoundTrip(t *testing.T) {
	t.Parallel()
	p := newPIDFile(t)
	if err := p.Write("127.0.0.1:8080"); err != nil {
		t.Fatalf("write: %v", err)
	}
	pid, addr, err := p.Read()
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if pid != os.Getpid() {
		t.Errorf("pid = %d, want %d", pid, os.Getpid())
	}
	if addr != "127.0.0.1:8080" {
		t.Errorf("addr = %q, want 127.0.0.1:8080", addr)
	}
}

func TestPIDFile_Read_MissingFile_ReturnsErrNotExist(t *testing.T) {
	t.Parallel()
	p := newPIDFile(t)
	_, _, err := p.Read()
	if err == nil || !errors.Is(err, os.ErrNotExist) {
		t.Errorf("err = %v, want os.ErrNotExist", err)
	}
}

func TestPIDFile_IsAlive_CurrentProcess_ReturnsTrue(t *testing.T) {
	t.Parallel()
	p := newPIDFile(t)
	if err := p.Write("addr"); err != nil {
		t.Fatalf("write: %v", err)
	}
	alive, pid, addr := p.IsAlive()
	if !alive {
		t.Errorf("expected alive=true for self pid")
	}
	if pid != os.Getpid() || addr != "addr" {
		t.Errorf("pid/addr mismatch: got %d/%q, want %d/%q", pid, addr, os.Getpid(), "addr")
	}
}

func TestPIDFile_IsAlive_DeadPID_ReturnsFalse(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	pidPath := filepath.Join(dir, "mock.pid")
	addrPath := filepath.Join(dir, "mock.addr")
	// PID 1 is init/launchd; cannot signal it from an unprivileged process.
	// PID 999999 is almost certainly free. We use the latter because Signal
	// 0 returns nil for any process the caller is privileged to signal.
	if err := os.WriteFile(pidPath, []byte("999999"), 0o600); err != nil {
		t.Fatalf("write pid: %v", err)
	}
	if err := os.WriteFile(addrPath, []byte("addr"), 0o600); err != nil {
		t.Fatalf("write addr: %v", err)
	}
	p := mock.NewPIDFile(pidPath, addrPath)
	alive, _, _ := p.IsAlive()
	if alive {
		t.Errorf("expected alive=false for nonexistent PID")
	}
}

func TestPIDFile_SignalStop_NoFile_ReturnsErrNotExist(t *testing.T) {
	t.Parallel()
	p := newPIDFile(t)
	err := p.SignalStop()
	if !errors.Is(err, os.ErrNotExist) {
		t.Errorf("err = %v, want os.ErrNotExist", err)
	}
}

func TestPIDFile_Remove_Idempotent(t *testing.T) {
	t.Parallel()
	p := newPIDFile(t)
	if err := p.Write("addr"); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := p.Remove(); err != nil {
		t.Errorf("first remove: %v", err)
	}
	if err := p.Remove(); err != nil {
		t.Errorf("second remove (should be a no-op): %v", err)
	}
}

func TestPIDFile_WaitUntilStopped_DeadPID_ReturnsImmediately(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	pidPath := filepath.Join(dir, "mock.pid")
	addrPath := filepath.Join(dir, "mock.addr")
	_ = os.WriteFile(pidPath, []byte("999999"), 0o600)
	_ = os.WriteFile(addrPath, []byte("addr"), 0o600)
	p := mock.NewPIDFile(pidPath, addrPath)

	start := time.Now()
	if err := p.WaitUntilStopped(2 * time.Second); err != nil {
		t.Errorf("wait: %v", err)
	}
	if elapsed := time.Since(start); elapsed > 500*time.Millisecond {
		t.Errorf("wait took %s, want <500ms (PID was already dead)", elapsed)
	}
}

func TestPIDFile_WaitUntilStopped_LiveProcess_TimesOut(t *testing.T) {
	t.Parallel()
	p := newPIDFile(t)
	if err := p.Write("addr"); err != nil {
		t.Fatalf("write: %v", err)
	}
	err := p.WaitUntilStopped(150 * time.Millisecond)
	if err == nil {
		t.Fatalf("expected timeout error")
	}
	if !strings.Contains(err.Error(), "did not stop") {
		t.Errorf("err = %q, want containing 'did not stop'", err)
	}
}
