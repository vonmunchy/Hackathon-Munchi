package spec

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/cli"
	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/cli/output"
	specpkg "github.com/BML-Digital/swipe-merchants-dev/cli/internal/spec"
	versionpkg "github.com/BML-Digital/swipe-merchants-dev/cli/internal/version"
)

// NewCommand builds the `swipe spec` parent command.
func NewCommand(g *cli.GlobalFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "spec",
		Short: "Inspect the OpenAPI spec embedded in this binary",
	}
	cmd.AddCommand(newShowCommand(g))
	cmd.AddCommand(newVersionCommand(g))
	cmd.AddCommand(newValidateCommand(g))
	return cmd
}

func newShowCommand(g *cli.GlobalFlags) *cobra.Command {
	var format string
	cmd := &cobra.Command{
		Use:   "show",
		Short: "Print the embedded OpenAPI spec",
		Long:  "Print the embedded OpenAPI spec as YAML (default) or JSON. Use --format to override; the global --output flag also influences output where applicable.",
		RunE: func(_ *cobra.Command, _ []string) error {
			return runShow(g, format)
		},
	}
	cmd.Flags().StringVarP(&format, "format", "f", "", "spec serialization format: yaml | json (defaults to --output, then yaml)")
	return cmd
}

func newVersionCommand(g *cli.GlobalFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the embedded spec version, binary version, and git sha",
		RunE: func(_ *cobra.Command, _ []string) error {
			return runVersion(g)
		},
	}
}

func runShow(g *cli.GlobalFlags, format string) error {
	s, err := specpkg.Load()
	if err != nil {
		return fmt.Errorf("load spec: %w", err)
	}
	target, err := resolveSpecFormat(format, g.Output)
	if err != nil {
		return err
	}
	r, err := g.Renderer()
	if err != nil {
		return err
	}
	switch target {
	case output.FormatJSON:
		buf, err := s.RawJSON()
		if err != nil {
			return err
		}
		return r.Raw(buf)
	default:
		return r.Raw(s.RawYAML())
	}
}

// resolveSpecFormat picks the serialization format for `spec show`. The
// command-local --format wins; otherwise --output is honored when it is
// json or yaml. The default is yaml (the spec's native form).
func resolveSpecFormat(local, global string) (output.Format, error) {
	if local != "" {
		f, err := output.ParseFormat(local)
		if err != nil {
			return "", fmt.Errorf("invalid --format: %w", err)
		}
		if f == output.FormatTable {
			return "", fmt.Errorf("--format table is not meaningful for `spec show`")
		}
		return f, nil
	}
	switch global {
	case string(output.FormatJSON):
		return output.FormatJSON, nil
	case string(output.FormatYAML), string(output.FormatTable), "":
		return output.FormatYAML, nil
	default:
		return output.FormatYAML, nil
	}
}

// VersionInfo bundles the identifiers that `swipe spec version` prints:
// binary version + commit + build date alongside the embedded spec
// version + title.
type VersionInfo struct {
	BinaryVersion string `json:"binary_version" yaml:"binary_version"`
	BinaryCommit  string `json:"binary_commit" yaml:"binary_commit"`
	BuildDate     string `json:"build_date" yaml:"build_date"`
	SpecVersion   string `json:"spec_version" yaml:"spec_version"`
	SpecTitle     string `json:"spec_title" yaml:"spec_title"`
}

func runVersion(g *cli.GlobalFlags) error {
	s, err := specpkg.Load()
	if err != nil {
		return fmt.Errorf("load spec: %w", err)
	}
	info := versionpkg.Get()
	out := VersionInfo{
		BinaryVersion: info.Version,
		BinaryCommit:  info.Commit,
		BuildDate:     info.BuildDate,
		SpecVersion:   s.Version(),
		SpecTitle:     s.Title(),
	}
	r, err := g.Renderer()
	if err != nil {
		return err
	}
	return r.Object(out)
}
