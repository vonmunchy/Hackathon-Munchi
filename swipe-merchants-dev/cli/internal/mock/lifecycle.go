package mock

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// PIDFile records the running mock-server PID and bound listen address. It
// is used by `swipe mock stop` and `swipe mock status` to find the running
// process without scanning ports. The two pieces of state live in two
// files so each is editable/readable in isolation.
type PIDFile struct {
	PIDPath  string
	AddrPath string
}

// NewPIDFile builds a PIDFile rooted at the directory containing pidPath.
func NewPIDFile(pidPath, addrPath string) PIDFile {
	return PIDFile{PIDPath: pidPath, AddrPath: addrPath}
}

// Write writes the current process's PID and the bound listen address.
// The directory is created with 0o700 if it does not yet exist.
func (p PIDFile) Write(addr string) error {
	if err := os.MkdirAll(filepath.Dir(p.PIDPath), 0o700); err != nil {
		return fmt.Errorf("mkdir for pidfile: %w", err)
	}
	if err := os.WriteFile(p.PIDPath, []byte(strconv.Itoa(os.Getpid())), 0o600); err != nil {
		return fmt.Errorf("write pidfile: %w", err)
	}
	if err := os.WriteFile(p.AddrPath, []byte(addr), 0o600); err != nil {
		return fmt.Errorf("write addrfile: %w", err)
	}
	return nil
}

// Remove deletes both files. Missing files are not an error.
func (p PIDFile) Remove() error {
	for _, f := range []string{p.PIDPath, p.AddrPath} {
		if err := os.Remove(f); err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("remove %s: %w", f, err)
		}
	}
	return nil
}

// Read returns the recorded PID and address. The error is wrapped from
// os.ReadFile, so callers can use errors.Is(err, os.ErrNotExist) to
// detect "no mock recorded".
func (p PIDFile) Read() (int, string, error) {
	pidBytes, err := os.ReadFile(p.PIDPath)
	if err != nil {
		return 0, "", err
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(pidBytes)))
	if err != nil {
		return 0, "", fmt.Errorf("parse pidfile %s: %w", p.PIDPath, err)
	}
	addrBytes, err := os.ReadFile(p.AddrPath)
	if err != nil {
		return pid, "", err
	}
	return pid, strings.TrimSpace(string(addrBytes)), nil
}

// IsAlive returns true if the recorded PID matches a running process.
// Uses signal 0 (no-op) to probe.
func (p PIDFile) IsAlive() (bool, int, string) {
	pid, addr, err := p.Read()
	if err != nil {
		return false, 0, ""
	}
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false, pid, addr
	}
	if err := proc.Signal(syscall.Signal(0)); err != nil {
		return false, pid, addr
	}
	return true, pid, addr
}

// SignalStop sends SIGTERM to the recorded PID. Returns os.ErrNotExist if
// no live mock was recorded.
func (p PIDFile) SignalStop() error {
	alive, pid, _ := p.IsAlive()
	if !alive {
		return os.ErrNotExist
	}
	proc, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("find process %d: %w", pid, err)
	}
	if err := proc.Signal(syscall.SIGTERM); err != nil {
		return fmt.Errorf("signal %d: %w", pid, err)
	}
	return nil
}

// WaitUntilStopped polls IsAlive every 50ms until it returns false or the
// timeout elapses. Returns nil on stop, error on timeout.
func (p PIDFile) WaitUntilStopped(timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		alive, _, _ := p.IsAlive()
		if !alive {
			return nil
		}
		time.Sleep(50 * time.Millisecond)
	}
	return fmt.Errorf("mock did not stop within %s", timeout)
}
