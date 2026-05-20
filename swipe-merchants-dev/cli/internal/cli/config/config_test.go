package config_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/cli"
	configcmd "github.com/BML-Digital/swipe-merchants-dev/cli/internal/cli/config"
)

func runConfig(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	g := &cli.GlobalFlags{Stdout: stdout, Stderr: stderr, Output: "table"}
	root := cli.NewRoot(g, func(r *cobra.Command, gf *cli.GlobalFlags) {
		r.AddCommand(configcmd.NewCommand(gf))
	})
	root.SetArgs(args)
	err := root.Execute()
	return stdout.String(), stderr.String(), err
}

func TestPath_Table_ContainsAllRows(t *testing.T) {
	t.Parallel()
	stdout, _, err := runConfig(t, "config", "path")
	if err != nil {
		t.Fatalf("path: %v", err)
	}
	for _, want := range []string{"root", "config_file", "mock_state_db", "token_cache"} {
		if !strings.Contains(stdout, want) {
			t.Errorf("missing %q:\n%s", want, stdout)
		}
	}
}

func TestPath_InvalidOutputFormat_Errors(t *testing.T) {
	t.Parallel()
	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	g := &cli.GlobalFlags{Stdout: stdout, Stderr: stderr, Output: "garbage"}
	root := cli.NewRoot(g, func(r *cobra.Command, gf *cli.GlobalFlags) {
		r.AddCommand(configcmd.NewCommand(gf))
	})
	root.SetArgs([]string{"config", "path"})
	if err := root.Execute(); err == nil {
		t.Errorf("expected error for invalid format")
	}
}
