package webhooks_test

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"

	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/cli"
	webhookscmd "github.com/BML-Digital/swipe-merchants-dev/cli/internal/cli/webhooks"
	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock"
	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/auth"
	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/store"
	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/webhooks"
)

func runWebhooks(t *testing.T, port string, stdin string, args ...string) (string, string, error) {
	t.Helper()
	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	g := &cli.GlobalFlags{Stdout: stdout, Stderr: stderr, Output: "json"}
	root := cli.NewRoot(g, func(r *cobra.Command, gf *cli.GlobalFlags) {
		r.AddCommand(webhookscmd.NewCommand(gf))
	})
	if port != "" {
		root.SetArgs(append([]string{"--port", port, "webhooks"}, args...))
	} else {
		root.SetArgs(append([]string{"webhooks"}, args...))
	}
	root.SetOut(stdout)
	root.SetErr(stderr)
	if stdin != "" {
		root.SetIn(strings.NewReader(stdin))
	}
	err := root.ExecuteContext(context.Background())
	return stdout.String(), stderr.String(), err
}

func liveMock(t *testing.T) (string, string) {
	t.Helper()
	dir := t.TempDir()
	st, err := store.Open(filepath.Join(dir, "state.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })
	seed, err := st.SeedDefaults()
	if err != nil {
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
	ctx, cancel := context.WithCancel(context.Background())
	go func() { _ = srv.Serve(ctx) }()
	t.Cleanup(cancel)
	_, port, err := net.SplitHostPort(srv.Listener().Addr().String())
	if err != nil {
		t.Fatalf("split: %v", err)
	}
	return port, seed.WebhookSecret
}

func TestWebhooks_Secret_ReadsFromMock(t *testing.T) {
	port, secret := liveMock(t)
	stdout, _, err := runWebhooks(t, port, "", "secret")
	if err != nil {
		t.Fatalf("secret: %v", err)
	}
	got := strings.TrimSpace(stdout)
	if got != secret {
		t.Errorf("secret = %q, want %q", got, secret)
	}
}

func TestWebhooks_Verify_GoodSignature_Succeeds(t *testing.T) {
	secret := "whsec_test"
	id := "msg_a"
	ts := strconv.FormatInt(time.Now().UTC().Unix(), 10)
	body := []byte(`{"eventType":"transaction.state_changed"}`)
	sig := webhooks.Sign(secret, id, ts, body)

	stdout, _, err := runWebhooks(t, "", string(body),
		"verify", "--id", id, "--timestamp", ts, "--signature", sig, "--secret", secret)
	if err != nil {
		t.Fatalf("verify: %v\nout=%s", err, stdout)
	}
	if !strings.Contains(stdout, "valid") {
		t.Errorf("stdout = %q", stdout)
	}
}

func TestWebhooks_Verify_BadSignature_Fails(t *testing.T) {
	secret := "whsec_test"
	id := "msg_b"
	ts := strconv.FormatInt(time.Now().UTC().Unix(), 10)
	body := []byte(`{}`)
	_, _, err := runWebhooks(t, "", string(body),
		"verify", "--id", id, "--timestamp", ts, "--signature", "v1,bogus", "--secret", secret)
	if err == nil {
		t.Errorf("expected error on bad signature")
	}
}

func TestWebhooks_Verify_RequiresSecret(t *testing.T) {
	_, _, err := runWebhooks(t, "", "{}",
		"verify", "--id", "x", "--timestamp", "1", "--signature", "v1,x")
	if err == nil || !strings.Contains(err.Error(), "--secret is required") {
		t.Errorf("err = %v", err)
	}
}
