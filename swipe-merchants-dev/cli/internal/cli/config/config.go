package config

import (
	"github.com/spf13/cobra"

	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/cli"
	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/cli/output"
	pathconfig "github.com/BML-Digital/swipe-merchants-dev/cli/internal/config"
)

// NewCommand builds the `swipe config` parent command.
func NewCommand(g *cli.GlobalFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Inspect swipe configuration and file paths",
	}
	cmd.AddCommand(newPathCommand(g))
	return cmd
}

func newPathCommand(g *cli.GlobalFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "path",
		Short: "Print all swipe-managed file system paths",
		RunE: func(_ *cobra.Command, _ []string) error {
			return runPath(g)
		},
	}
}

func runPath(g *cli.GlobalFlags) error {
	paths, err := pathconfig.DefaultPaths()
	if err != nil {
		return err
	}
	r, err := g.Renderer()
	if err != nil {
		return err
	}
	return r.KV(output.KeyValueTable{
		Title: "swipe paths",
		Rows: []output.KeyValueRow{
			{Key: "root", Value: paths.Root},
			{Key: "config_file", Value: paths.ConfigFile},
			{Key: "mock_dir", Value: paths.MockDir},
			{Key: "mock_state_db", Value: paths.MockStateDB},
			{Key: "mock_keys_dir", Value: paths.MockKeysDir},
			{Key: "mock_pidfile", Value: paths.MockPIDFile},
			{Key: "mock_address_file", Value: paths.MockSocketFile},
			{Key: "token_cache", Value: paths.TokenCache},
		},
	})
}
