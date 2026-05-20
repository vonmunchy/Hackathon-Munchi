package completion

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/cli"
)

// NewCommand builds `swipe completion <shell>`. The actual completion
// scripts come from cobra's built-in generators; this command is a thin
// wrapper that routes by shell name.
func NewCommand(g *cli.GlobalFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "completion <shell>",
		Short: "Generate shell completion script",
		Long: `Generate the autocompletion script for the specified shell.

Install:
  bash:        source <(swipe completion bash)
  zsh:         swipe completion zsh > "$fpath[1]/_swipe"
  fish:        swipe completion fish > ~/.config/fish/completions/swipe.fish
  powershell:  swipe completion powershell > swipe.ps1`,
		Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			root := cmd.Root()
			switch args[0] {
			case "bash":
				return root.GenBashCompletionV2(g.Stdout, true)
			case "zsh":
				return root.GenZshCompletion(g.Stdout)
			case "fish":
				return root.GenFishCompletion(g.Stdout, true)
			case "powershell":
				return root.GenPowerShellCompletionWithDesc(g.Stdout)
			default:
				return fmt.Errorf("unsupported shell %q", args[0])
			}
		},
	}
	return cmd
}
