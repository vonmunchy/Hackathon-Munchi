package keys_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/cli"
	keyscmd "github.com/BML-Digital/swipe-merchants-dev/cli/internal/cli/keys"
	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock"
	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/auth"
	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/store"
)

// liveMock spins up a full mock.New server and returns its port + the
// shared store handle so a test can seed extra clients out-of-band.
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

// runKeys executes the keys cobra tree with the supplied args, returning
// stdout, stderr, and the cobra ExecuteContext error.
func runKeys(t *testing.T, port string, args ...string) (string, string, error) {
	t.Helper()
	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	g := &cli.GlobalFlags{Stdout: stdout, Stderr: stderr, Output: "json"}
	root := cli.NewRoot(g, func(r *cobra.Command, gf *cli.GlobalFlags) {
		r.AddCommand(keyscmd.NewCommand(gf))
	})
	root.SetArgs(append([]string{"--port", port, "keys"}, args...))
	root.SetOut(stdout)
	root.SetErr(stderr)
	err := root.ExecuteContext(context.Background())
	return stdout.String(), stderr.String(), err
}

// httptestPort is unused; kept here to keep the linter from complaining if
// future tests want a raw httptest server independent of the mock package.
var _ = httptest.NewServer

func TestKeys_Create_PrintsSecret(t *testing.T) {
	port, _ := liveMock(t)
	stdout, _, err := runKeys(t, port, "create", "--name", "phase2-cli", "--scopes", "wallet:balance")
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	var body map[string]any
	if err := json.Unmarshal([]byte(stdout), &body); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !strings.HasPrefix(body["id"].(string), "cli_") {
		t.Errorf("id: %v", body["id"])
	}
	if !strings.HasPrefix(body["client_secret"].(string), "sec_") {
		t.Errorf("secret: %v", body["client_secret"])
	}
}

func TestKeys_Create_RequiresName(t *testing.T) {
	port, _ := liveMock(t)
	_, _, err := runKeys(t, port, "create")
	if err == nil || !strings.Contains(err.Error(), "--name is required") {
		t.Errorf("err = %v", err)
	}
}

func TestKeys_List_ShowsCreated(t *testing.T) {
	port, st := liveMock(t)
	c, _, _ := st.CreateClient(store.DefaultMerchantID, "p", []string{"wallet:balance"})
	stdout, _, err := runKeys(t, port, "list")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if !strings.Contains(stdout, c.ID) {
		t.Errorf("list output missing %s: %s", c.ID, stdout)
	}
}

func TestKeys_Show_ReturnsClient(t *testing.T) {
	port, st := liveMock(t)
	c, _, _ := st.CreateClient(store.DefaultMerchantID, "p", nil)
	stdout, _, err := runKeys(t, port, "show", c.ID)
	if err != nil {
		t.Fatalf("show: %v", err)
	}
	if !strings.Contains(stdout, c.ID) {
		t.Errorf("show output missing id")
	}
}

func TestKeys_Rotate_ReturnsNewSecret(t *testing.T) {
	port, st := liveMock(t)
	c, originalSecret, _ := st.CreateClient(store.DefaultMerchantID, "p", nil)
	stdout, _, err := runKeys(t, port, "rotate", c.ID)
	if err != nil {
		t.Fatalf("rotate: %v", err)
	}
	var body map[string]any
	if err := json.Unmarshal([]byte(stdout), &body); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if body["client_secret"].(string) == originalSecret {
		t.Errorf("rotate returned original secret")
	}
}

func TestKeys_Revoke_DisablesClient(t *testing.T) {
	port, st := liveMock(t)
	c, _, _ := st.CreateClient(store.DefaultMerchantID, "p", nil)
	stdout, _, err := runKeys(t, port, "revoke", c.ID)
	if err != nil {
		t.Fatalf("revoke: %v", err)
	}
	if !strings.Contains(stdout, `"enabled": false`) {
		t.Errorf("revoke output should show enabled=false: %s", stdout)
	}
}

func TestKeys_Delete_RemovesClient(t *testing.T) {
	port, st := liveMock(t)
	c, _, _ := st.CreateClient(store.DefaultMerchantID, "p", nil)
	_, _, err := runKeys(t, port, "delete", c.ID)
	if err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, getErr := st.GetClient(c.ID); getErr == nil {
		t.Errorf("client still present after delete")
	}
}

func TestKeys_Test_RequiresSecret(t *testing.T) {
	port, _ := liveMock(t)
	_, _, err := runKeys(t, port, "test", "cli_xx")
	if err == nil || !strings.Contains(err.Error(), "--secret is required") {
		t.Errorf("err = %v", err)
	}
}

func TestKeys_Test_HappyPath(t *testing.T) {
	port, st := liveMock(t)
	c, secret, _ := st.CreateClient(store.DefaultMerchantID, "p", []string{"wallet:balance"})
	stdout, _, err := runKeys(t, port, "test", c.ID, "--secret", secret)
	if err != nil {
		t.Fatalf("test: %v", err)
	}
	var body map[string]any
	if err := json.Unmarshal([]byte(stdout), &body); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if body["client_id"].(string) != c.ID {
		t.Errorf("client_id = %v", body["client_id"])
	}
	if body["merchant_id"].(string) != store.DefaultMerchantID {
		t.Errorf("merchant_id = %v", body["merchant_id"])
	}
}
