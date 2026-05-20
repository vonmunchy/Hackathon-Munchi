package mock_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/spf13/cobra"

	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/cli"
	mockcmd "github.com/BML-Digital/swipe-merchants-dev/cli/internal/cli/mock"
)

// syncBuffer wraps a bytes.Buffer with a mutex so a foreground `mock start`
// (which writes its own messages plus drives a slog handler from another
// goroutine) does not race the test reading the buffer for assertions.
type syncBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (s *syncBuffer) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.buf.Write(p)
}

func (s *syncBuffer) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.buf.String()
}

// freeTCPPort returns an OS-picked free TCP port. The test then passes this
// to `swipe mock start --port <port>`. There is a small race between
// listener-close and the CLI's bind, but in practice the test is reliable
// on dev machines and CI; the real-world cost of this race is at worst a
// rare flake which `-count=1` will retry on demand.
func freeTCPPort(t *testing.T) int {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("pick port: %v", err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	_ = ln.Close()
	return port
}

// newMockRoot returns a fresh root command (with only the mock subcommand
// attached). Each test gets its own root + GlobalFlags so they are
// independent under -race. Callers should pass io.Discard or a syncBuffer
// for streams that the running mock writes to from multiple goroutines.
func newMockRoot(t *testing.T, port int, stdout, stderr io.Writer) *cobra.Command {
	t.Helper()
	g := &cli.GlobalFlags{
		Stdout: stdout,
		Stderr: stderr,
		Output: "json",
		Port:   port,
	}
	return cli.NewRoot(g, func(r *cobra.Command, gf *cli.GlobalFlags) {
		r.AddCommand(mockcmd.NewCommand(gf))
	})
}

// waitForReady polls /health/alive until ok or deadline. Used by tests
// to know when `mock start` has finished binding and seeding.
func waitForReady(t *testing.T, port int, deadline time.Duration) {
	t.Helper()
	url := fmt.Sprintf("http://127.0.0.1:%d/health/alive", port)
	end := time.Now().Add(deadline)
	for time.Now().Before(end) {
		resp, err := http.Get(url) //nolint:gosec,noctx
		if err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return
			}
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("mock did not become ready on %s within %s", url, deadline)
}

// isolatedHome makes the mock's path resolver point at a temp dir for the
// duration of this test. config.DefaultPaths uses os.UserHomeDir which
// honors $HOME on unix and %USERPROFILE% on windows.
func isolatedHome(t *testing.T) string {
	t.Helper()
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("USERPROFILE", tmp) // windows
	return tmp
}

func TestMockStart_Status_Stop_FullLifecycle(t *testing.T) {
	// Not Parallel: each test isolates HOME but shares the user's process
	// signal table; running multiple `mock start` instances simultaneously
	// works (different ports + homes) but adds little signal.
	isolatedHome(t)
	port := freeTCPPort(t)

	// --- mock start (foreground, in a goroutine) ----------------------
	startStdout := &syncBuffer{}
	startCtx, cancelStart := context.WithCancel(context.Background())
	startDone := make(chan error, 1)
	go func() {
		root := newMockRoot(t, port, startStdout, io.Discard)
		root.SetArgs([]string{"mock", "start"})
		startDone <- root.ExecuteContext(startCtx)
	}()

	waitForReady(t, port, 3*time.Second)

	// Verify the seed summary made it to stdout.
	if got := startStdout.String(); !strings.Contains(got, "mer_default") {
		t.Errorf("seed summary missing mer_default:\n%s", got)
	}
	if got := startStdout.String(); !strings.Contains(got, "10000.00") {
		t.Errorf("seed summary missing MVR balance:\n%s", got)
	}

	// --- mock status --------------------------------------------------
	statusStdout, statusStderr := &bytes.Buffer{}, &bytes.Buffer{}
	statusRoot := newMockRoot(t, port, statusStdout, statusStderr)
	statusRoot.SetArgs([]string{"mock", "status"})
	if err := statusRoot.Execute(); err != nil {
		t.Fatalf("mock status: %v\nstderr: %s", err, statusStderr.String())
	}
	var status struct {
		Running bool `json:"running"`
		PID     int  `json:"pid"`
		Detail  *struct {
			Counts struct {
				Merchants    int `json:"merchants"`
				BankAccounts int `json:"bank_accounts"`
				Wallets      int `json:"wallets"`
			} `json:"counts"`
		} `json:"detail"`
	}
	if err := json.Unmarshal(statusStdout.Bytes(), &status); err != nil {
		t.Fatalf("parse status: %v: %s", err, statusStdout.String())
	}
	if !status.Running {
		t.Errorf("status.running = false; want true")
	}
	if status.PID == 0 {
		t.Errorf("status.pid not set: %+v", status)
	}
	if status.Detail == nil {
		t.Fatalf("status.detail missing: %s", statusStdout.String())
	}
	if status.Detail.Counts.Merchants != 1 || status.Detail.Counts.BankAccounts != 2 {
		t.Errorf("counts = %+v, want merchants=1 bank_accounts=2", status.Detail.Counts)
	}

	// --- second `mock start` is rejected with "already running" -------
	dupStdout, dupStderr := &bytes.Buffer{}, &bytes.Buffer{}
	dupRoot := newMockRoot(t, port, dupStdout, dupStderr)
	dupRoot.SetArgs([]string{"mock", "start"})
	if err := dupRoot.Execute(); err == nil {
		t.Errorf("expected duplicate mock start to fail; got nil error")
	} else if !strings.Contains(err.Error(), "already running") {
		t.Errorf("err = %q, want containing 'already running'", err)
	}

	// --- mock stop ---------------------------------------------------
	stopStdout, stopStderr := &bytes.Buffer{}, &bytes.Buffer{}
	stopRoot := newMockRoot(t, port, stopStdout, stopStderr)
	stopRoot.SetArgs([]string{"mock", "stop"})
	if err := stopRoot.Execute(); err != nil {
		t.Fatalf("mock stop: %v\nstderr: %s", err, stopStderr.String())
	}
	if !strings.Contains(stopStdout.String(), "mock stopped") {
		t.Errorf("stop output: %s", stopStdout.String())
	}

	// --- the foreground start goroutine must exit cleanly ------------
	cancelStart() // belt-and-braces in case stop didn't reach it
	select {
	case err := <-startDone:
		if err != nil {
			t.Errorf("mock start returned error after stop: %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatalf("mock start goroutine did not exit within 3s")
	}

	// --- mock status after stop reports "running": false -------------
	postStopStdout := &bytes.Buffer{}
	postStopRoot := newMockRoot(t, port, postStopStdout, &bytes.Buffer{})
	postStopRoot.SetArgs([]string{"mock", "status"})
	if err := postStopRoot.Execute(); err != nil {
		t.Errorf("post-stop status: %v", err)
	}
	if !strings.Contains(postStopStdout.String(), `"running": false`) {
		t.Errorf("post-stop status didn't show running=false:\n%s", postStopStdout.String())
	}
}

func TestMockStart_StatePersistsAcrossRestart(t *testing.T) {
	isolatedHome(t)
	port := freeTCPPort(t)

	// First run — fresh seed.
	firstStdout := &syncBuffer{}
	firstCtx, firstCancel := context.WithCancel(context.Background())
	firstDone := make(chan error, 1)
	go func() {
		root := newMockRoot(t, port, firstStdout, io.Discard)
		root.SetArgs([]string{"mock", "start"})
		firstDone <- root.ExecuteContext(firstCtx)
	}()
	waitForReady(t, port, 3*time.Second)
	firstSummary := firstStdout.String()

	// Stop the first run.
	stopRoot := newMockRoot(t, port, &bytes.Buffer{}, &bytes.Buffer{})
	stopRoot.SetArgs([]string{"mock", "stop"})
	if err := stopRoot.Execute(); err != nil {
		t.Fatalf("stop: %v", err)
	}
	firstCancel()
	if err := <-firstDone; err != nil {
		t.Fatalf("first start returned: %v", err)
	}

	// Second run — same $HOME, same state file.
	secondStdout := &syncBuffer{}
	secondCtx, secondCancel := context.WithCancel(context.Background())
	secondDone := make(chan error, 1)
	go func() {
		root := newMockRoot(t, port, secondStdout, io.Discard)
		root.SetArgs([]string{"mock", "start"})
		secondDone <- root.ExecuteContext(secondCtx)
	}()
	waitForReady(t, port, 3*time.Second)
	secondSummary := secondStdout.String()

	// Cleanup before assertions so a failure still leaves the test process tidy.
	stopRoot2 := newMockRoot(t, port, &bytes.Buffer{}, &bytes.Buffer{})
	stopRoot2.SetArgs([]string{"mock", "stop"})
	_ = stopRoot2.Execute()
	secondCancel()
	<-secondDone

	if !strings.Contains(secondSummary, "state restored from disk") {
		t.Errorf("second start did not report restored state:\n%s", secondSummary)
	}
	// Both summaries should reference the same merchant id.
	if !strings.Contains(firstSummary, "mer_default") || !strings.Contains(secondSummary, "mer_default") {
		t.Errorf("merchant id missing across runs\nfirst: %s\nsecond: %s", firstSummary, secondSummary)
	}
	// Webhook secret is generated once and persisted — extract from each
	// summary and compare.
	firstSecret := extractWebhookSecret(firstSummary)
	secondSecret := extractWebhookSecret(secondSummary)
	if firstSecret == "" || firstSecret != secondSecret {
		t.Errorf("webhook secret rotated unexpectedly: %q -> %q", firstSecret, secondSecret)
	}
}

func TestMockStop_NoRunningMock_ReportsAndExitsCleanly(t *testing.T) {
	isolatedHome(t)
	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	root := newMockRoot(t, 0, stdout, stderr)
	root.SetArgs([]string{"mock", "stop"})
	if err := root.Execute(); err != nil {
		t.Fatalf("stop with no mock: %v", err)
	}
	if !strings.Contains(stdout.String(), "no mock recorded as running") {
		t.Errorf("stdout: %s", stdout.String())
	}
}

func TestMockStatus_InvalidGlobalOutput_RendererErrors(t *testing.T) {
	isolatedHome(t)
	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	g := &cli.GlobalFlags{Stdout: stdout, Stderr: stderr, Output: "garbage", Port: 0}
	root := cli.NewRoot(g, func(r *cobra.Command, gf *cli.GlobalFlags) {
		r.AddCommand(mockcmd.NewCommand(gf))
	})
	root.SetArgs([]string{"mock", "status"})
	if err := root.Execute(); err == nil {
		t.Errorf("expected renderer error from invalid global output")
	}
}

func TestMockStatus_NoRunningMock_ReportsRunningFalse(t *testing.T) {
	isolatedHome(t)
	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	root := newMockRoot(t, 0, stdout, stderr)
	root.SetArgs([]string{"mock", "status", "--output", "json"})
	if err := root.Execute(); err != nil {
		t.Fatalf("status with no mock: %v\nstderr: %s", err, stderr.String())
	}
	var got map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("decode: %v: %s", err, stdout.String())
	}
	if got["running"] != false {
		t.Errorf("running = %v, want false", got["running"])
	}
}

func TestMockStatus_TableFormat_RendersExpectedRows(t *testing.T) {
	isolatedHome(t)
	port := freeTCPPort(t)

	startStdout := &syncBuffer{}
	startCtx, cancel := context.WithCancel(context.Background())
	startDone := make(chan error, 1)
	go func() {
		root := newMockRoot(t, port, startStdout, io.Discard)
		// Pass --webhook-url so the printSeedSummary and
		// renderStatusTable webhook-url branches are also exercised.
		root.SetArgs([]string{"mock", "start", "--webhook-url", "http://example.com/hook"})
		startDone <- root.ExecuteContext(startCtx)
	}()
	waitForReady(t, port, 3*time.Second)
	defer func() {
		cancel()
		<-startDone
	}()

	if !strings.Contains(startStdout.String(), "webhook url:") {
		t.Errorf("seed summary missing webhook url line:\n%s", startStdout.String())
	}

	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	statusRoot := newMockRoot(t, port, stdout, stderr)
	statusRoot.SetArgs([]string{"mock", "status", "--output", "table"})
	if err := statusRoot.Execute(); err != nil {
		t.Fatalf("status table: %v\nstderr: %s", err, stderr.String())
	}
	out := stdout.String()
	for _, want := range []string{"running", "true", "merchants", "bank_accounts", "webhook_secret", "webhook_url"} {
		if !strings.Contains(out, want) {
			t.Errorf("table output missing %q:\n%s", want, out)
		}
	}
}

func TestMockStart_AlreadyRunningGuard_FailsCleanly(t *testing.T) {
	// Explicit small test for the duplicate-start guard, isolated from the
	// big lifecycle test so coverage tooling marks the branch reliably.
	isolatedHome(t)
	port := freeTCPPort(t)

	startStdout := &syncBuffer{}
	startCtx, cancel := context.WithCancel(context.Background())
	startDone := make(chan error, 1)
	go func() {
		root := newMockRoot(t, port, startStdout, io.Discard)
		root.SetArgs([]string{"mock", "start"})
		startDone <- root.ExecuteContext(startCtx)
	}()
	waitForReady(t, port, 3*time.Second)
	defer func() {
		cancel()
		<-startDone
	}()

	dupRoot := newMockRoot(t, port, &bytes.Buffer{}, &bytes.Buffer{})
	dupRoot.SetArgs([]string{"mock", "start"})
	err := dupRoot.Execute()
	if err == nil || !strings.Contains(err.Error(), "already running") {
		t.Errorf("expected 'already running' error; got %v", err)
	}
}

func TestMockStart_EnsureDirsFails_ReturnsError(t *testing.T) {
	// Point HOME at a regular file so that MkdirAll(~/.swipe) fails with
	// ENOTDIR. This also forces DefaultPaths to resolve to a path under
	// the file, which is what we want for the test.
	parent := t.TempDir()
	blocker := filepath.Join(parent, "home-as-file")
	if err := os.WriteFile(blocker, []byte(""), 0o600); err != nil {
		t.Fatalf("seed blocker file: %v", err)
	}
	t.Setenv("HOME", blocker)
	t.Setenv("USERPROFILE", blocker)

	root := newMockRoot(t, freeTCPPort(t), &bytes.Buffer{}, io.Discard)
	root.SetArgs([]string{"mock", "start"})
	err := root.Execute()
	if err == nil {
		t.Fatalf("expected ensure-dirs error; got nil")
	}
	if !strings.Contains(err.Error(), "ensure dirs") {
		t.Errorf("err = %q, want containing 'ensure dirs'", err)
	}
}

func TestMockStart_StateDBPathIsDirectory_OpenErrors(t *testing.T) {
	// Pre-create ~/.swipe/mock/state.db as a directory so bolt.Open fails.
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("USERPROFILE", tmp)
	mockDir := filepath.Join(tmp, ".swipe", "mock")
	if err := os.MkdirAll(filepath.Join(mockDir, "state.db"), 0o700); err != nil {
		t.Fatalf("seed state.db-as-dir: %v", err)
	}
	root := newMockRoot(t, freeTCPPort(t), &bytes.Buffer{}, io.Discard)
	root.SetArgs([]string{"mock", "start"})
	err := root.Execute()
	if err == nil {
		t.Fatalf("expected open-state error")
	}
	if !strings.Contains(err.Error(), "open state") {
		t.Errorf("err = %q, want 'open state'", err)
	}
}

func TestMockStart_PIDFilePathIsDirectory_WriteErrors(t *testing.T) {
	// Pre-create ~/.swipe/mock/mock.pid as a directory so the WriteFile
	// inside PIDFile.Write fails. The bind step succeeds first because
	// the port is free.
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("USERPROFILE", tmp)
	mockDir := filepath.Join(tmp, ".swipe", "mock")
	if err := os.MkdirAll(filepath.Join(mockDir, "mock.pid"), 0o700); err != nil {
		t.Fatalf("seed pidfile-as-dir: %v", err)
	}
	root := newMockRoot(t, freeTCPPort(t), &bytes.Buffer{}, io.Discard)
	root.SetArgs([]string{"mock", "start"})
	err := root.Execute()
	if err == nil {
		t.Fatalf("expected pidfile write error")
	}
	if !strings.Contains(err.Error(), "write pidfile") {
		t.Errorf("err = %q, want 'write pidfile'", err)
	}
}

func TestMockStart_PortAlreadyBound_BindErrorPropagates(t *testing.T) {
	isolatedHome(t)
	// runStart binds on `:port` (all interfaces). Use the same address
	// spec for the hold listener so the conflict is unambiguous on
	// every OS (macOS lets `127.0.0.1:p` and `[::]:p` coexist).
	hold, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("hold listener: %v", err)
	}
	defer func() { _ = hold.Close() }()
	port := hold.Addr().(*net.TCPAddr).Port

	// Use a deadline-bounded context so a regression here cannot hang
	// the whole test process.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	root := newMockRoot(t, port, &bytes.Buffer{}, io.Discard)
	root.SetArgs([]string{"mock", "start"})
	err = root.ExecuteContext(ctx)
	if err == nil {
		t.Fatalf("expected bind error; got nil")
	}
	if !strings.Contains(err.Error(), "bind") && !strings.Contains(err.Error(), "address already in use") {
		t.Errorf("unexpected error: %v", err)
	}
}

// extractWebhookSecret pulls the `whsec_<hex>` token out of a mock-start
// summary line that looks like `  webhook secret: whsec_<32hex>`.
func extractWebhookSecret(s string) string {
	for line := range strings.SplitSeq(s, "\n") {
		if i := strings.Index(line, "whsec_"); i >= 0 {
			tail := line[i:]
			end := strings.IndexFunc(tail, func(r rune) bool {
				return r == ' ' || r == '\t' || r == '\n' || r == '\r'
			})
			if end < 0 {
				return tail
			}
			return tail[:end]
		}
	}
	return ""
}
