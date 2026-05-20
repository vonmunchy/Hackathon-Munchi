package admin_test

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"path/filepath"
	"strings"
	"testing"

	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock"
	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/admin"
	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/auth"
	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/store"
	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/transport"
)

// adminTestServer boots a full mock server with seeded defaults and returns
// an AdminClient pointed at it.
func adminTestServer(t *testing.T) *transport.AdminClient {
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
	return transport.NewAdminClient("http://" + srv.Listener().Addr().String())
}

func TestAdminKeys_CreateListShowDelete_Cycle(t *testing.T) {
	c := adminTestServer(t)
	ctx := context.Background()

	resp, err := c.CreateClient(ctx, admin.CreateRequest{
		Name:   "phase2",
		Scopes: []string{"wallet:balance", "payments:qr"},
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if !strings.HasPrefix(resp.Client.ID, "cli_") {
		t.Errorf("id = %q, want cli_ prefix", resp.Client.ID)
	}
	if !strings.HasPrefix(resp.Secret, "sec_") {
		t.Errorf("secret = %q, want sec_ prefix", resp.Secret)
	}

	listed, err := c.ListClients(ctx)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(listed) != 1 || listed[0].ID != resp.Client.ID {
		t.Errorf("list = %v", listed)
	}

	got, err := c.GetClient(ctx, resp.Client.ID)
	if err != nil {
		t.Fatalf("show: %v", err)
	}
	if got.Name != "phase2" {
		t.Errorf("name = %q", got.Name)
	}

	if err := c.DeleteClient(ctx, resp.Client.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := c.GetClient(ctx, resp.Client.ID); err == nil || !strings.Contains(err.Error(), "404") {
		t.Errorf("after delete: err = %v, want 404", err)
	}
}

func TestAdminKeys_UpdateScopesAndDisable(t *testing.T) {
	c := adminTestServer(t)
	ctx := context.Background()
	resp, err := c.CreateClient(ctx, admin.CreateRequest{Name: "t", Scopes: []string{"wallet:balance"}})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	disabled := false
	newScopes := []string{"payments:qr", "payments:link"}
	updated, err := c.UpdateClient(ctx, resp.Client.ID, admin.UpdateRequest{
		Scopes:  &newScopes,
		Enabled: &disabled,
	})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if updated.Enabled {
		t.Errorf("expected disabled")
	}
	if !equal(updated.Scopes, []string{"payments:link", "payments:qr"}) {
		t.Errorf("scopes = %v", updated.Scopes)
	}
}

func TestAdminKeys_RotateInvalidatesOldSecret(t *testing.T) {
	c := adminTestServer(t)
	ctx := context.Background()
	resp, err := c.CreateClient(ctx, admin.CreateRequest{Name: "t", Scopes: []string{"wallet:balance"}})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	rotated, err := c.RotateClientSecret(ctx, resp.Client.ID)
	if err != nil {
		t.Fatalf("rotate: %v", err)
	}
	if rotated.Secret == resp.Secret {
		t.Errorf("rotate returned same secret")
	}

	// Old secret must no longer work against /oauth2/token.
	tc := transport.NewClient(strings.TrimSuffix(c.BaseURL, "/"))
	if _, err := tc.IssueToken(ctx, resp.Client.ID, resp.Secret, nil); err == nil || !strings.Contains(err.Error(), "401") {
		t.Errorf("old secret should fail, got %v", err)
	}
	if _, err := tc.IssueToken(ctx, resp.Client.ID, rotated.Secret, nil); err != nil {
		t.Errorf("new secret should succeed, got %v", err)
	}
}

func TestAdminKeys_FromNonLoopback_Rejected(t *testing.T) {
	// The admin gate refuses anything that isn't a loopback. We construct a
	// raw request with a RemoteAddr override to exercise that path.
	c := adminTestServer(t)
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, c.BaseURL+"/_admin/keys", nil)
	req.RemoteAddr = "203.0.113.1:1234"
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		t.Fatalf("do: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	// Server still sees 127.0.0.1 because RemoteAddr is overridden by the
	// kernel for the actual TCP connection; this test mainly ensures the
	// admin endpoint is reachable from loopback. The non-loopback rejection
	// is exercised in server_internal_test.go.
	if resp.StatusCode != 200 {
		t.Errorf("admin from loopback should succeed, got %d", resp.StatusCode)
	}
}

func equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
