package wallet_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/cli"
	walletcmd "github.com/BML-Digital/swipe-merchants-dev/cli/internal/cli/wallet"
	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock"
	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/auth"
	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/store"
	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/transport"
)

// liveMock returns a running mock, its listen port, and the store handle.
func liveMock(t *testing.T) (port string, st *store.Store) {
	t.Helper()
	dir := t.TempDir()
	s, err := store.Open(filepath.Join(dir, "state.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })
	if _, err := s.SeedDefaults(); err != nil {
		t.Fatalf("seed: %v", err)
	}
	key, err := auth.LoadOrCreateSigningKey(filepath.Join(dir, "keys"))
	if err != nil {
		t.Fatalf("key: %v", err)
	}
	srv, err := mock.New(mock.Config{
		ListenAddr: "127.0.0.1:0",
		Store:      s,
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
	_, port, err = net.SplitHostPort(srv.Listener().Addr().String())
	if err != nil {
		t.Fatalf("split: %v", err)
	}
	return port, s
}

// seedTokenCache creates a client + writes a cached token at $HOME/.swipe/token.json.
func seedTokenCache(t *testing.T, home, base string, st *store.Store, scopes []string) {
	t.Helper()
	c, secret, err := st.CreateClient(store.DefaultMerchantID, "p3", scopes)
	if err != nil {
		t.Fatalf("client: %v", err)
	}
	tc := transport.NewClient(base)
	tok, err := tc.IssueToken(context.Background(), c.ID, secret, nil)
	if err != nil {
		t.Fatalf("issue token: %v", err)
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

func runWallet(t *testing.T, home, port string, args ...string) (string, error) {
	t.Helper()
	t.Setenv("HOME", home)
	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	g := &cli.GlobalFlags{Stdout: stdout, Stderr: stderr, Output: "json"}
	root := cli.NewRoot(g, func(r *cobra.Command, gf *cli.GlobalFlags) {
		r.AddCommand(walletcmd.NewCommand(gf))
	})
	root.SetArgs(append([]string{"--port", port, "wallet"}, args...))
	root.SetOut(stdout)
	root.SetErr(stderr)
	err := root.ExecuteContext(context.Background())
	return stdout.String(), err
}

func TestWallet_Balance_HappyPath(t *testing.T) {
	port, st := liveMock(t)
	home := t.TempDir()
	seedTokenCache(t, home, "http://localhost:"+port, st, []string{"wallet:balance"})

	stdout, err := runWallet(t, home, port, "balance")
	if err != nil {
		t.Fatalf("balance: %v", err)
	}
	var body []map[string]any
	if err := json.Unmarshal([]byte(stdout), &body); err != nil {
		t.Fatalf("unmarshal: %v (out=%s)", err, stdout)
	}
	if len(body) != 2 {
		t.Errorf("got %d entries", len(body))
	}
}

func TestWallet_Balance_NoCachedToken_Errors(t *testing.T) {
	port, _ := liveMock(t)
	_, err := runWallet(t, t.TempDir(), port, "balance")
	if err == nil || !strings.Contains(err.Error(), "no cached token") {
		t.Errorf("err = %v", err)
	}
}

func TestWallet_Accounts_HappyPath(t *testing.T) {
	port, st := liveMock(t)
	home := t.TempDir()
	seedTokenCache(t, home, "http://localhost:"+port, st, []string{"wallet:balance"})

	stdout, err := runWallet(t, home, port, "accounts")
	if err != nil {
		t.Fatalf("accounts: %v", err)
	}
	var body []map[string]any
	if err := json.Unmarshal([]byte(stdout), &body); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(body) != 2 {
		t.Errorf("got %d accounts", len(body))
	}
}

func TestWallet_Balance_WrongScope_Errors(t *testing.T) {
	port, st := liveMock(t)
	home := t.TempDir()
	seedTokenCache(t, home, "http://localhost:"+port, st, []string{"transactions:status"})
	_, err := runWallet(t, home, port, "balance")
	if err == nil || !strings.Contains(err.Error(), "403") {
		t.Errorf("err = %v, want 403", err)
	}
}
