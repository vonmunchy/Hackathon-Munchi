// Command swipe is the entry point for the Swipe Merchant CLI binary. It
// wires the cobra root command from internal/cli to its Phase 1 subtree
// (version, config, spec, health, mock).
package main

import (
	"context"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/cli"
	authcmd "github.com/BML-Digital/swipe-merchants-dev/cli/internal/cli/auth"
	completioncmd "github.com/BML-Digital/swipe-merchants-dev/cli/internal/cli/completion"
	configcmd "github.com/BML-Digital/swipe-merchants-dev/cli/internal/cli/config"
	clierrors "github.com/BML-Digital/swipe-merchants-dev/cli/internal/cli/errors"
	healthcmd "github.com/BML-Digital/swipe-merchants-dev/cli/internal/cli/health"
	keyscmd "github.com/BML-Digital/swipe-merchants-dev/cli/internal/cli/keys"
	logscmd "github.com/BML-Digital/swipe-merchants-dev/cli/internal/cli/logs"
	mockcmd "github.com/BML-Digital/swipe-merchants-dev/cli/internal/cli/mock"
	paymentscmd "github.com/BML-Digital/swipe-merchants-dev/cli/internal/cli/payments"
	payoutscmd "github.com/BML-Digital/swipe-merchants-dev/cli/internal/cli/payouts"
	speccmd "github.com/BML-Digital/swipe-merchants-dev/cli/internal/cli/spec"
	transactionscmd "github.com/BML-Digital/swipe-merchants-dev/cli/internal/cli/transactions"
	versioncmd "github.com/BML-Digital/swipe-merchants-dev/cli/internal/cli/version"
	walletcmd "github.com/BML-Digital/swipe-merchants-dev/cli/internal/cli/wallet"
	webhookscmd "github.com/BML-Digital/swipe-merchants-dev/cli/internal/cli/webhooks"
)

func main() {
	os.Exit(Run(os.Args[1:], os.Stdout, os.Stderr))
}

// Run executes the swipe CLI with the given args and writers, returning a
// process exit code (0 success, 1 error). It is the testable seam for the
// otherwise os.Exit-coupled main.
func Run(args []string, stdout, stderr io.Writer) int {
	g := cli.Defaults()
	g.Stdout = stdout
	g.Stderr = stderr
	root := cli.NewRoot(g, attachSubcommands)
	root.SetArgs(args)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := root.ExecuteContext(ctx); err != nil {
		clierrors.Render(g.Stderr, err, g.Verbose)
		return 1
	}
	return 0
}

func attachSubcommands(root *cobra.Command, g *cli.GlobalFlags) {
	root.AddCommand(versioncmd.NewCommand(g))
	root.AddCommand(configcmd.NewCommand(g))
	root.AddCommand(speccmd.NewCommand(g))
	root.AddCommand(healthcmd.NewCommand(g))
	root.AddCommand(mockcmd.NewCommand(g))
	root.AddCommand(keyscmd.NewCommand(g))
	root.AddCommand(authcmd.NewCommand(g))
	root.AddCommand(walletcmd.NewCommand(g))
	root.AddCommand(transactionscmd.NewCommand(g))
	root.AddCommand(paymentscmd.NewCommand(g))
	root.AddCommand(payoutscmd.NewCommand(g))
	root.AddCommand(webhookscmd.NewCommand(g))
	root.AddCommand(logscmd.NewCommand(g))
	root.AddCommand(completioncmd.NewCommand(g))
}
