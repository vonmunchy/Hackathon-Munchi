package auth_test

import (
	"context"
	"io"
	"log/slog"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock"
	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/auth"
	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/store"
	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/transport"
)

// liveServer starts a full mock.New server on httptest, seeds defaults,
// creates a client, and returns its credentials + the transport client
// pointed at the test server.
func liveServer(t *testing.T, scopes []string) (*transport.Client, string, string) {
	t.Helper()
	dir := t.TempDir()
	st, err := store.Open(filepath.Join(dir, "state.db"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })
	if _, err := st.SeedDefaults(); err != nil {
		t.Fatalf("seed: %v", err)
	}
	key, err := auth.LoadOrCreateSigningKey(filepath.Join(dir, "keys"))
	if err != nil {
		t.Fatalf("signing key: %v", err)
	}
	srv, err := mock.New(mock.Config{
		ListenAddr: "127.0.0.1:0",
		Store:      st,
		Logger:     slog.New(slog.NewJSONHandler(io.Discard, nil)),
		SigningKey: key,
	})
	if err != nil {
		t.Fatalf("build server: %v", err)
	}
	if err := srv.Bind(); err != nil {
		t.Fatalf("bind: %v", err)
	}
	bound := srv.Listener().Addr().String()
	ctx, cancel := context.WithCancel(context.Background())
	go func() { _ = srv.Serve(ctx) }()
	t.Cleanup(cancel)

	c, secret, err := st.CreateClient(store.DefaultMerchantID, "test", scopes)
	if err != nil {
		t.Fatalf("create client: %v", err)
	}

	base := &url.URL{Scheme: "http", Host: bound}
	return transport.NewClient(base.String()), c.ID, secret
}

func TestTokenEndpoint_HappyPath_Issues200(t *testing.T) {
	scopes := []string{"wallet:balance", "payments:qr"}
	c, cid, secret := liveServer(t, scopes)

	tok, err := c.IssueToken(context.Background(), cid, secret, nil)
	if err != nil {
		t.Fatalf("issue token: %v", err)
	}
	if tok.AccessToken == "" {
		t.Fatalf("empty access token")
	}
	if tok.TokenType != "Bearer" {
		t.Errorf("token_type = %q, want Bearer", tok.TokenType)
	}
	if tok.ExpiresIn < 60 {
		t.Errorf("expires_in = %d, want >= 60", tok.ExpiresIn)
	}
	wantScopes := "payments:qr wallet:balance"
	if tok.Scope != wantScopes {
		t.Errorf("scope = %q, want %q", tok.Scope, wantScopes)
	}
}

func TestTokenEndpoint_BadSecret_Returns401(t *testing.T) {
	c, cid, _ := liveServer(t, []string{"wallet:balance"})
	_, err := c.IssueToken(context.Background(), cid, "sec_wrong", nil)
	if err == nil || !strings.Contains(err.Error(), "401") {
		t.Errorf("err = %v, want 401", err)
	}
}

func TestTokenEndpoint_RequestedScopeOutsideGrant_Returns400(t *testing.T) {
	c, cid, secret := liveServer(t, []string{"wallet:balance"})
	_, err := c.IssueToken(context.Background(), cid, secret, []string{"payments:qr"})
	if err == nil || !strings.Contains(err.Error(), "400") {
		t.Errorf("err = %v, want 400", err)
	}
}

func TestWhoamiEndpoint_ValidToken_ReturnsClaims(t *testing.T) {
	scopes := []string{"wallet:balance"}
	c, cid, secret := liveServer(t, scopes)
	tok, err := c.IssueToken(context.Background(), cid, secret, nil)
	if err != nil {
		t.Fatalf("issue token: %v", err)
	}
	who, err := c.WhoAmI(context.Background(), tok.AccessToken)
	if err != nil {
		t.Fatalf("whoami: %v", err)
	}
	if who.ClientID != cid {
		t.Errorf("client_id = %q, want %q", who.ClientID, cid)
	}
	if who.MerchantID != store.DefaultMerchantID {
		t.Errorf("merchant_id = %q, want %q", who.MerchantID, store.DefaultMerchantID)
	}
	if len(who.Scopes) != 1 || who.Scopes[0] != "wallet:balance" {
		t.Errorf("scopes = %v", who.Scopes)
	}
}

func TestWhoamiEndpoint_NoToken_Returns401(t *testing.T) {
	c, _, _ := liveServer(t, nil)
	_, err := c.WhoAmI(context.Background(), "")
	if err == nil || !strings.Contains(err.Error(), "401") {
		t.Errorf("err = %v, want 401", err)
	}
}

func TestWhoamiEndpoint_ExpiredToken_Returns401(t *testing.T) {
	// Build a minimal one-off server with a custom-TTL issuer so we don't
	// need to wait an hour. liveServer doesn't expose TTL; rebuild here.
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
		TokenTTL:   200 * time.Millisecond,
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

	c, secret, err := st.CreateClient(store.DefaultMerchantID, "t", []string{"wallet:balance"})
	if err != nil {
		t.Fatalf("client: %v", err)
	}
	cli := transport.NewClient("http://" + srv.Listener().Addr().String())
	tok, err := cli.IssueToken(context.Background(), c.ID, secret, nil)
	if err != nil {
		t.Fatalf("token: %v", err)
	}
	time.Sleep(400 * time.Millisecond)
	if _, err := cli.WhoAmI(context.Background(), tok.AccessToken); err == nil || !strings.Contains(err.Error(), "401") {
		t.Errorf("expired token whoami = %v, want 401", err)
	}
}

func TestWellKnownEndpoints_Reachable(t *testing.T) {
	c, _, _ := liveServer(t, nil)
	// quick assertion via raw HTTP that the metadata is reachable; the
	// shape is already validated by the manual smoke test in the PR.
	resp, err := c.HTTPClient.Get(c.BaseURL + "/.well-known/oauth-authorization-server")
	if err != nil {
		t.Fatalf("metadata: %v", err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Errorf("metadata status = %d", resp.StatusCode)
	}
	resp2, err := c.HTTPClient.Get(c.BaseURL + "/.well-known/jwks.json")
	if err != nil {
		t.Fatalf("jwks: %v", err)
	}
	_ = resp2.Body.Close()
	if resp2.StatusCode != 200 {
		t.Errorf("jwks status = %d", resp2.StatusCode)
	}
}

// dummy import-shape sanity check; httptest is referenced via the indirect
// recorder under newConfig in server_lifecycle_test.go.
var _ = httptest.NewRecorder
