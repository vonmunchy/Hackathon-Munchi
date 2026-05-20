package mock_test

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock"
	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/auth"
	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/store"
)

func newConfig(t *testing.T) mock.Config {
	t.Helper()
	dir := t.TempDir()
	st, err := store.Open(filepath.Join(dir, "state.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })
	if _, err := st.SeedDefaults(); err != nil {
		t.Fatalf("seed: %v", err)
	}
	key, err := auth.LoadOrCreateSigningKey(filepath.Join(dir, "keys"))
	if err != nil {
		t.Fatalf("signing key: %v", err)
	}
	return mock.Config{
		ListenAddr: "127.0.0.1:0",
		Store:      st,
		Logger:     slog.New(slog.NewJSONHandler(io.Discard, nil)),
		SigningKey: key,
	}
}

func TestNew_NilStore_Errors(t *testing.T) {
	t.Parallel()
	_, err := mock.New(mock.Config{
		Logger: slog.New(slog.NewJSONHandler(io.Discard, nil)),
	})
	if err == nil {
		t.Fatalf("expected error for nil store")
	}
	if !strings.Contains(err.Error(), "store is required") {
		t.Errorf("err = %q", err)
	}
}

func TestNew_NilLogger_Errors(t *testing.T) {
	t.Parallel()
	st, err := store.Open(filepath.Join(t.TempDir(), "state.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })

	_, err = mock.New(mock.Config{Store: st})
	if err == nil {
		t.Fatalf("expected error for nil logger")
	}
	if !strings.Contains(err.Error(), "logger is required") {
		t.Errorf("err = %q", err)
	}
}

func TestNew_DefaultStartedAt_FillsIn(t *testing.T) {
	t.Parallel()
	cfg := newConfig(t)
	cfg.StartedAt = time.Time{} // zero
	srv, err := mock.New(cfg)
	if err != nil {
		t.Fatalf("new: %v", err)
	}
	// Bind synchronously, never enter Serve, so no goroutine reads/writes
	// the unexported listener concurrently with the test.
	if err := srv.Bind(); err != nil {
		t.Fatalf("bind: %v", err)
	}
	if err := srv.Shutdown(context.Background()); err != nil {
		t.Errorf("shutdown: %v", err)
	}
}

func TestBind_TwiceOnSameServer_Errors(t *testing.T) {
	t.Parallel()
	srv, err := mock.New(newConfig(t))
	if err != nil {
		t.Fatalf("new: %v", err)
	}
	if err := srv.Bind(); err != nil {
		t.Fatalf("first bind: %v", err)
	}
	t.Cleanup(func() { _ = srv.Shutdown(context.Background()) })

	if err := srv.Bind(); err == nil || !strings.Contains(err.Error(), "already bound") {
		t.Errorf("second bind: err = %v, want containing 'already bound'", err)
	}
}

func TestServe_WithoutBind_Errors(t *testing.T) {
	t.Parallel()
	srv, err := mock.New(newConfig(t))
	if err != nil {
		t.Fatalf("new: %v", err)
	}
	err = srv.Serve(context.Background())
	if err == nil || !strings.Contains(err.Error(), "before Bind") {
		t.Errorf("err = %v, want containing 'before Bind'", err)
	}
}

func TestStart_ConvenienceWrapper_BindsAndServesUntilCancelled(t *testing.T) {
	t.Parallel()
	// Pick a free port up-front so the test can probe HTTP without ever
	// touching srv.Listener() concurrently with Start's bind.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("pick port: %v", err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	_ = ln.Close()

	cfg := newConfig(t)
	cfg.ListenAddr = fmt.Sprintf("127.0.0.1:%d", port)
	srv, err := mock.New(cfg)
	if err != nil {
		t.Fatalf("new: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- srv.Start(ctx) }()

	// Probe the bound port via HTTP — confirms Start completed Bind and
	// entered Serve without racing on the unexported listener field.
	url := fmt.Sprintf("http://127.0.0.1:%d/health/alive", port)
	deadline := time.Now().Add(2 * time.Second)
	ready := false
	for time.Now().Before(deadline) {
		resp, err := http.Get(url) //nolint:gosec,noctx
		if err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				ready = true
				break
			}
		}
		time.Sleep(20 * time.Millisecond)
	}
	if !ready {
		cancel()
		t.Fatalf("server not ready on %s within 2s", url)
	}
	cancel()

	select {
	case err := <-done:
		if err != nil {
			t.Errorf("start returned: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("Start did not return after ctx cancel")
	}
}

func TestBind_EmptyListenAddr_DefaultsTo8080(t *testing.T) {
	t.Parallel()
	// Hold :8080 to guarantee the bind fails (we don't want to leave a
	// real listener on the canonical port). The error message contains
	// the resolved address; verifying it's :8080 proves the empty-string
	// branch was taken.
	hold, err := net.Listen("tcp", ":8080")
	if err != nil {
		t.Skip("port 8080 already in use; default-branch test skipped")
	}
	defer func() { _ = hold.Close() }()

	cfg := newConfig(t)
	cfg.ListenAddr = ""
	srv, err := mock.New(cfg)
	if err != nil {
		t.Fatalf("new: %v", err)
	}
	bindErr := srv.Bind()
	if bindErr == nil {
		t.Errorf("expected bind to fail with :8080 held")
	} else if !strings.Contains(bindErr.Error(), ":8080") {
		t.Errorf("err = %q, want containing :8080 (default branch)", bindErr)
	}
}

func TestServe_ListenerClosedExternally_ReturnsAcceptError(t *testing.T) {
	t.Parallel()
	srv, err := mock.New(newConfig(t))
	if err != nil {
		t.Fatalf("new: %v", err)
	}
	if err := srv.Bind(); err != nil {
		t.Fatalf("bind: %v", err)
	}
	// Close the listener so Serve's inner Accept fails immediately.
	_ = srv.Listener().Close()

	err = srv.Serve(context.Background())
	if err == nil {
		t.Errorf("expected accept error from Serve; got nil")
	}
}

func TestStart_BindFails_ReturnsBindError(t *testing.T) {
	t.Parallel()
	// Hold a port so Start's Bind cannot succeed.
	hold, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("hold: %v", err)
	}
	defer func() { _ = hold.Close() }()
	port := hold.Addr().(*net.TCPAddr).Port

	cfg := newConfig(t)
	cfg.ListenAddr = fmt.Sprintf(":%d", port)
	srv, err := mock.New(cfg)
	if err != nil {
		t.Fatalf("new: %v", err)
	}
	if err := srv.Start(context.Background()); err == nil {
		t.Errorf("expected bind error from Start")
	}
}

func TestShutdown_OnUnstartedServer_NoOps(t *testing.T) {
	t.Parallel()
	// A Server that never had New called returns nil from Shutdown.
	srv := &mock.Server{}
	if err := srv.Shutdown(context.Background()); err != nil {
		t.Errorf("shutdown of zero Server: %v", err)
	}
}
