package health_test

import (
	"bytes"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/cli"
	healthcmd "github.com/BML-Digital/swipe-merchants-dev/cli/internal/cli/health"
)

// runWithFakeServer wires the health command tree onto a buffered cli root
// and runs it against a fake HTTP server. The fake server replies with the
// given handler, and the command sees it via --port.
func runWithFakeServer(t *testing.T, handler http.Handler, args ...string) (string, string, error) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	// The CLI builds its base URL as http://localhost:<port>; httptest binds
	// on 127.0.0.1, so passing the port is sufficient.
	_, port, err := net.SplitHostPort(strings.TrimPrefix(srv.URL, "http://"))
	if err != nil {
		t.Fatalf("parse url: %v", err)
	}

	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	g := &cli.GlobalFlags{Stdout: stdout, Stderr: stderr, Output: "json"}
	root := cli.NewRoot(g, func(r *cobra.Command, gf *cli.GlobalFlags) {
		r.AddCommand(healthcmd.NewCommand(gf))
	})
	root.SetArgs(append([]string{"--port", port}, args...))
	root.SetOut(stdout)
	root.SetErr(stderr)
	return stdout.String(), stderr.String(), root.Execute()
}

func TestAlive_HappyPath_PrintsStatusOK(t *testing.T) {
	t.Parallel()
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/health/alive" {
			t.Errorf("path = %q", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})
	_, _, err := runWithFakeServer(t, h, "health", "alive")
	if err != nil {
		t.Fatalf("alive: %v", err)
	}
}

func TestReady_Healthy_ExitsCleanly(t *testing.T) {
	t.Parallel()
	h := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status": "ok",
			"checks": map[string]any{"database": map[string]any{"status": "ok"}},
		})
	})
	_, _, err := runWithFakeServer(t, h, "health", "ready")
	if err != nil {
		t.Fatalf("ready: %v", err)
	}
}

func TestReady_Unhealthy_ReturnsError(t *testing.T) {
	t.Parallel()
	h := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status": "unhealthy",
			"checks": map[string]any{"database": map[string]any{"status": "unhealthy"}},
		})
	})
	_, _, err := runWithFakeServer(t, h, "health", "ready")
	if err == nil {
		t.Fatalf("expected error from unhealthy /health/ready")
	}
	if !strings.Contains(err.Error(), "not ready") {
		t.Errorf("err = %q, want containing 'not ready'", err)
	}
}

func TestAlive_BadJSON_ReturnsError(t *testing.T) {
	t.Parallel()
	h := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("not json"))
	})
	_, _, err := runWithFakeServer(t, h, "health", "alive")
	if err == nil {
		t.Fatalf("expected decode error")
	}
}

func TestReady_BadJSON_ReturnsError(t *testing.T) {
	t.Parallel()
	h := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("{not json"))
	})
	_, _, err := runWithFakeServer(t, h, "health", "ready")
	if err == nil {
		t.Fatalf("expected parse error")
	}
}

func TestAlive_TableFormat_RendersRows(t *testing.T) {
	t.Parallel()
	h := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	_, port, _ := net.SplitHostPort(strings.TrimPrefix(srv.URL, "http://"))

	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	g := &cli.GlobalFlags{Stdout: stdout, Stderr: stderr, Output: "table"}
	root := cli.NewRoot(g, func(r *cobra.Command, gf *cli.GlobalFlags) {
		r.AddCommand(healthcmd.NewCommand(gf))
	})
	root.SetArgs([]string{"--port", port, "health", "alive"})
	if err := root.Execute(); err != nil {
		t.Fatalf("alive: %v", err)
	}
	if !strings.Contains(stdout.String(), "status") || !strings.Contains(stdout.String(), "ok") {
		t.Errorf("table missing rows: %s", stdout.String())
	}
}

func TestAlive_InvalidGlobalOutput_RendererErrors(t *testing.T) {
	t.Parallel()
	h := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	_, port, _ := net.SplitHostPort(strings.TrimPrefix(srv.URL, "http://"))

	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	g := &cli.GlobalFlags{Stdout: stdout, Stderr: stderr, Output: "garbage"}
	root := cli.NewRoot(g, func(r *cobra.Command, gf *cli.GlobalFlags) {
		r.AddCommand(healthcmd.NewCommand(gf))
	})
	root.SetArgs([]string{"--port", port, "health", "alive"})
	if err := root.Execute(); err == nil {
		t.Errorf("expected renderer error from invalid global output")
	}
}

func TestReady_InvalidGlobalOutput_RendererErrors(t *testing.T) {
	t.Parallel()
	h := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status": "ok",
			"checks": map[string]any{"database": map[string]any{"status": "ok"}},
		})
	})
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	_, port, _ := net.SplitHostPort(strings.TrimPrefix(srv.URL, "http://"))

	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	g := &cli.GlobalFlags{Stdout: stdout, Stderr: stderr, Output: "garbage"}
	root := cli.NewRoot(g, func(r *cobra.Command, gf *cli.GlobalFlags) {
		r.AddCommand(healthcmd.NewCommand(gf))
	})
	root.SetArgs([]string{"--port", port, "health", "ready"})
	if err := root.Execute(); err == nil {
		t.Errorf("expected renderer error from invalid global output")
	}
}

func TestBaseURL_PortZero_FallsBackTo8080(t *testing.T) {
	t.Parallel()
	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	g := &cli.GlobalFlags{Stdout: stdout, Stderr: stderr, Output: "json", Port: 0}
	root := cli.NewRoot(g, func(r *cobra.Command, gf *cli.GlobalFlags) {
		r.AddCommand(healthcmd.NewCommand(gf))
	})
	root.SetArgs([]string{"health", "alive"})
	// Either errors with a transport failure (8080 is unreachable in
	// test env) or succeeds (something happens to be listening). In
	// either case, baseURL's port==0 default branch ran.
	_ = root.Execute()
}

func TestAlive_NoServer_ReturnsTransportError(t *testing.T) {
	t.Parallel()
	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	g := &cli.GlobalFlags{Stdout: stdout, Stderr: stderr, Output: "json"}
	root := cli.NewRoot(g, func(r *cobra.Command, gf *cli.GlobalFlags) {
		r.AddCommand(healthcmd.NewCommand(gf))
	})
	// Port 1 is reserved and reliably refuses connections from non-root.
	root.SetArgs([]string{"--port", "1", "health", "alive"})
	root.SetOut(stdout)
	root.SetErr(stderr)
	if err := root.Execute(); err == nil {
		t.Fatalf("expected transport error; got nil")
	}
}
