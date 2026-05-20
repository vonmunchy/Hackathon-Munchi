package logs_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net"
	"net/http"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"

	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/cli"
	logscmd "github.com/BML-Digital/swipe-merchants-dev/cli/internal/cli/logs"
	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock"
	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/auth"
	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/store"
)

func liveMock(t *testing.T) string {
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
		t.Fatalf("key: %v", err)
	}
	srv, err := mock.New(mock.Config{
		ListenAddr: "127.0.0.1:0",
		Store:      st,
		Logger:     slog.New(slog.NewJSONHandler(io.Discard, nil)),
		SigningKey: key,
	})
	if err != nil {
		t.Fatalf("new: %v", err)
	}
	if err := srv.Bind(); err != nil {
		t.Fatalf("bind: %v", err)
	}
	addr := srv.Listener().Addr().String()
	ctx, cancel := context.WithCancel(context.Background())
	go func() { _ = srv.Serve(ctx) }()
	t.Cleanup(cancel)

	// drive a couple of requests so the log has content
	for range 3 {
		resp, err := http.Get("http://" + addr + "/health/alive")
		if err != nil {
			t.Fatalf("warmup: %v", err)
		}
		_ = resp.Body.Close()
	}

	_, port, err := net.SplitHostPort(addr)
	if err != nil {
		t.Fatalf("split: %v", err)
	}
	return port
}

func runLogs(t *testing.T, port string, args ...string) (string, error) {
	t.Helper()
	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	g := &cli.GlobalFlags{Stdout: stdout, Stderr: stderr, Output: "json"}
	root := cli.NewRoot(g, func(r *cobra.Command, gf *cli.GlobalFlags) {
		r.AddCommand(logscmd.NewCommand(gf))
	})
	root.SetArgs(append([]string{"--port", port, "logs"}, args...))
	root.SetOut(stdout)
	root.SetErr(stderr)
	err := root.ExecuteContext(context.Background())
	return stdout.String(), err
}

func TestLogs_Tail_ReturnsEntries(t *testing.T) {
	port := liveMock(t)
	stdout, err := runLogs(t, port, "tail", "--n", "10")
	if err != nil {
		t.Fatalf("tail: %v", err)
	}
	var entries []map[string]any
	if err := json.Unmarshal([]byte(stdout), &entries); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(entries) < 3 {
		t.Errorf("expected at least 3 entries, got %d", len(entries))
	}
}

func TestLogs_Show_ByID(t *testing.T) {
	port := liveMock(t)
	tailOut, err := runLogs(t, port, "tail", "--n", "5")
	if err != nil {
		t.Fatalf("tail: %v", err)
	}
	var tail []map[string]any
	if err := json.Unmarshal([]byte(tailOut), &tail); err != nil {
		t.Fatalf("unmarshal tail: %v", err)
	}
	if len(tail) == 0 {
		t.Fatal("no entries to show")
	}
	id, _ := tail[0]["id"].(string)
	out, err := runLogs(t, port, "show", id)
	if err != nil {
		t.Fatalf("show: %v", err)
	}
	if len(out) == 0 {
		t.Errorf("empty show output")
	}
}
