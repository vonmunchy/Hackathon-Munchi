package auth_test

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
	authcmd "github.com/BML-Digital/swipe-merchants-dev/cli/internal/cli/auth"
	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock"
	mockauth "github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/auth"
	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/store"
)

// liveMock boots a full mock server with seeded defaults and returns its
// listen port plus the live store handle.
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
	key, err := mockauth.LoadOrCreateSigningKey(filepath.Join(dir, "keys"))
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

// runAuth executes the auth subcommand tree with the given args. It also
// redirects HOME to a temp dir so the token cache is isolated per test.
func runAuth(t *testing.T, port string, args ...string) (string, string, error) {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	g := &cli.GlobalFlags{Stdout: stdout, Stderr: stderr, Output: "json"}
	root := cli.NewRoot(g, func(r *cobra.Command, gf *cli.GlobalFlags) {
		r.AddCommand(authcmd.NewCommand(gf))
	})
	root.SetArgs(append([]string{"--port", port, "auth"}, args...))
	root.SetOut(stdout)
	root.SetErr(stderr)
	err := root.ExecuteContext(context.Background())
	return stdout.String(), stderr.String(), err
}

// runAuthWithHome reuses an explicit HOME so a multi-step test (login →
// whoami → logout) can share a token cache across invocations.
func runAuthWithHome(t *testing.T, home, port string, args ...string) (string, string, error) {
	t.Helper()
	t.Setenv("HOME", home)
	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	g := &cli.GlobalFlags{Stdout: stdout, Stderr: stderr, Output: "json"}
	root := cli.NewRoot(g, func(r *cobra.Command, gf *cli.GlobalFlags) {
		r.AddCommand(authcmd.NewCommand(gf))
	})
	root.SetArgs(append([]string{"--port", port, "auth"}, args...))
	root.SetOut(stdout)
	root.SetErr(stderr)
	err := root.ExecuteContext(context.Background())
	return stdout.String(), stderr.String(), err
}

func TestAuth_Login_RequiresCredentials(t *testing.T) {
	port, _ := liveMock(t)
	_, _, err := runAuth(t, port, "login")
	if err == nil || !strings.Contains(err.Error(), "client-id and --client-secret are required") {
		t.Errorf("err = %v", err)
	}
}

func TestAuth_Login_WhoAmI_Logout_FullFlow(t *testing.T) {
	port, st := liveMock(t)
	c, secret, _ := st.CreateClient(store.DefaultMerchantID, "p", []string{"wallet:balance"})
	home := t.TempDir()

	_, _, err := runAuthWithHome(t, home, port, "login", "--client-id", c.ID, "--client-secret", secret)
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	// token file should now exist
	tokenPath := filepath.Join(home, ".swipe", "token.json")
	if _, statErr := stat(tokenPath); statErr != nil {
		t.Fatalf("token cache absent: %v", statErr)
	}

	stdout, _, err := runAuthWithHome(t, home, port, "whoami")
	if err != nil {
		t.Fatalf("whoami: %v", err)
	}
	var body map[string]any
	if err := json.Unmarshal([]byte(stdout), &body); err != nil {
		t.Fatalf("whoami unmarshal: %v", err)
	}
	if body["client_id"].(string) != c.ID {
		t.Errorf("client_id = %v", body["client_id"])
	}

	_, _, err = runAuthWithHome(t, home, port, "logout")
	if err != nil {
		t.Fatalf("logout: %v", err)
	}
	if _, statErr := stat(tokenPath); statErr == nil {
		t.Errorf("token cache should be removed after logout")
	}
}

func TestAuth_Whoami_NoCachedToken_Errors(t *testing.T) {
	port, _ := liveMock(t)
	_, _, err := runAuth(t, port, "whoami")
	if err == nil || !strings.Contains(err.Error(), "no cached token") {
		t.Errorf("err = %v", err)
	}
}

func TestAuth_Token_RequiresCachedToken(t *testing.T) {
	port, _ := liveMock(t)
	_, _, err := runAuth(t, port, "token")
	if err == nil || !strings.Contains(err.Error(), "no cached token") {
		t.Errorf("err = %v", err)
	}
}

func TestAuth_Token_MasksByDefault_AndShowsWithFlag(t *testing.T) {
	port, st := liveMock(t)
	c, secret, _ := st.CreateClient(store.DefaultMerchantID, "p", []string{"wallet:balance"})
	home := t.TempDir()

	if _, _, err := runAuthWithHome(t, home, port, "login", "--client-id", c.ID, "--client-secret", secret); err != nil {
		t.Fatalf("login: %v", err)
	}

	masked, _, err := runAuthWithHome(t, home, port, "token")
	if err != nil {
		t.Fatalf("token: %v", err)
	}
	if !strings.Contains(masked, "****") {
		t.Errorf("masked token should contain '****': %s", masked)
	}

	shown, _, err := runAuthWithHome(t, home, port, "token", "--show")
	if err != nil {
		t.Fatalf("token --show: %v", err)
	}
	if strings.Contains(shown, "****") {
		// the shown version should contain the actual JWT (long), not the mask
		t.Errorf("--show output should not contain '****': %s", shown)
	}
}

// stat is a tiny os.Stat wrapper to keep imports tight in this test file.
func stat(path string) (any, error) {
	return statFunc(path)
}
