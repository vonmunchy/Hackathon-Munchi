package payments_test

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/cli"
	paycmd "github.com/BML-Digital/swipe-merchants-dev/cli/internal/cli/payments"
	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock"
	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/auth"
	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/store"
	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/transport"
)

func liveMock(t *testing.T) (string, *store.Store) {
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
	ctx, cancel := context.WithCancel(context.Background())
	go func() { _ = srv.Serve(ctx) }()
	t.Cleanup(cancel)
	_, port, err := net.SplitHostPort(srv.Listener().Addr().String())
	if err != nil {
		t.Fatalf("split: %v", err)
	}
	return port, st
}

func seedTokenCache(t *testing.T, home, base string, st *store.Store, scopes []string) {
	t.Helper()
	c, secret, err := st.CreateClient(store.DefaultMerchantID, "p3", scopes)
	if err != nil {
		t.Fatalf("client: %v", err)
	}
	tc := transport.NewClient(base)
	tok, err := tc.IssueToken(context.Background(), c.ID, secret, nil)
	if err != nil {
		t.Fatalf("token: %v", err)
	}
	cached := transport.CachedToken{
		AccessToken: tok.AccessToken,
		TokenType:   tok.TokenType,
		ExpiresAt:   tok.ExpiresAt(),
		Scope:       tok.Scope,
		ClientID:    c.ID,
		IssuerURL:   base,
	}
	if err := transport.SaveCachedToken(filepath.Join(home, ".swipe", "token.json"), cached); err != nil {
		t.Fatalf("save: %v", err)
	}
}

func run(t *testing.T, home, port string, args ...string) (string, error) {
	t.Helper()
	t.Setenv("HOME", home)
	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	g := &cli.GlobalFlags{Stdout: stdout, Stderr: stderr, Output: "json"}
	root := cli.NewRoot(g, func(r *cobra.Command, gf *cli.GlobalFlags) {
		r.AddCommand(paycmd.NewCommand(gf))
	})
	root.SetArgs(append([]string{"--port", port, "payments"}, args...))
	root.SetOut(stdout)
	root.SetErr(stderr)
	err := root.ExecuteContext(context.Background())
	return stdout.String(), err
}

func TestPayments_Get_UnknownID_Returns404Error(t *testing.T) {
	port, st := liveMock(t)
	home := t.TempDir()
	seedTokenCache(t, home, "http://localhost:"+port, st, []string{"transactions:status"})
	_, err := run(t, home, port, "get", "pay_doesnotexist")
	if err == nil || !strings.Contains(err.Error(), "404") {
		t.Errorf("err = %v, want 404", err)
	}
}

func TestPayments_Get_WrongScope_Returns403Error(t *testing.T) {
	port, st := liveMock(t)
	home := t.TempDir()
	seedTokenCache(t, home, "http://localhost:"+port, st, []string{"wallet:balance"})
	_, err := run(t, home, port, "get", "pay_x")
	if err == nil || !strings.Contains(err.Error(), "403") {
		t.Errorf("err = %v, want 403", err)
	}
}
