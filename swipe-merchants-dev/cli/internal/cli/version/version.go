package version

import (
	"github.com/spf13/cobra"

	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/cli"
	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/spec"
	versionpkg "github.com/BML-Digital/swipe-merchants-dev/cli/internal/version"
)

// NewCommand builds the `swipe version` cobra command.
func NewCommand(g *cli.GlobalFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print binary version, commit, build date, and embedded spec version",
		RunE: func(_ *cobra.Command, _ []string) error {
			return run(g)
		},
	}
}

func run(g *cli.GlobalFlags) error {
	r, err := g.Renderer()
	if err != nil {
		return err
	}
	info := versionpkg.Get()
	if s, err := spec.Load(); err == nil {
		// Authoritative version is in the embedded spec, not the ldflag.
		info.SpecVersion = s.Version()
	}
	return r.Object(info)
}
