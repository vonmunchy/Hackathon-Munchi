package cli

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/cli/output"
)

// GlobalFlags captures the values bound from the cobra root command's
// persistent flags. Subcommands receive a pointer so changes (e.g. tests
// overriding the writer) propagate.
type GlobalFlags struct {
	Output     string
	Quiet      bool
	Verbose    bool
	NoColor    bool
	ConfigPath string
	Port       int
	Stdout     io.Writer
	Stderr     io.Writer
}

// Defaults returns a GlobalFlags pointing at os.Stdout/Stderr with V1
// defaults. Used by main and by tests that want isolated buffers.
func Defaults() *GlobalFlags {
	return &GlobalFlags{
		Output: string(output.FormatTable),
		Port:   8080,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}
}

// NewRoot builds the root cobra command. Subcommands must be attached by
// the caller (see cmd/swipe/main.go); doing so here would create a
// circular import between cli/* subpackages.
func NewRoot(g *GlobalFlags, attach func(*cobra.Command, *GlobalFlags)) *cobra.Command {
	root := &cobra.Command{
		Use:           "swipe",
		Short:         "Swipe Merchant CLI — embedded mock + API client",
		Long:          longDesc,
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.SetOut(g.Stdout)
	root.SetErr(g.Stderr)

	root.PersistentFlags().StringVarP(&g.Output, "output", "o", g.Output, "output format: table | json | yaml")
	root.PersistentFlags().BoolVar(&g.Quiet, "quiet", g.Quiet, "suppress non-essential output")
	root.PersistentFlags().BoolVar(&g.Verbose, "verbose", g.Verbose, "verbose logging (includes stack traces)")
	root.PersistentFlags().BoolVar(&g.NoColor, "no-color", g.NoColor, "disable ANSI color in table output")
	root.PersistentFlags().StringVar(&g.ConfigPath, "config", g.ConfigPath, "path to swipe config file (default ~/.swipe/config.yaml)")
	root.PersistentFlags().IntVar(&g.Port, "port", g.Port, "mock server port (overrides default for this invocation)")

	if attach != nil {
		attach(root, g)
	}
	return root
}

// Renderer builds a Renderer using the global flags' chosen format and
// writer. Surfaces format errors immediately.
func (g *GlobalFlags) Renderer() (output.Renderer, error) {
	f, err := output.ParseFormat(g.Output)
	if err != nil {
		return output.Renderer{}, err
	}
	return output.New(f, g.Stdout, !g.NoColor), nil
}

// Errorf writes a structured error line to g.Stderr. Used by subcommands
// instead of fmt.Fprintf for one-line consistency.
func (g *GlobalFlags) Errorf(format string, args ...any) {
	_, _ = fmt.Fprintf(g.Stderr, "Error: "+format+"\n", args...)
}

const longDesc = `swipe runs an embedded mock of the Swipe Merchants API on your laptop and
provides a CLI to drive it. Phase 1 brings up the binary, the mock server,
the BoltDB-backed state, the embedded OpenAPI spec, and the health probes.

  swipe mock start             # boot the mock at http://localhost:8080
  swipe health alive           # probe /health/alive
  swipe spec show              # print the embedded OpenAPI spec
  swipe config path            # show every file path swipe touches`
