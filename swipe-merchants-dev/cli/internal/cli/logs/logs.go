package logs

import (
	"fmt"
	"net/url"
	"time"

	"github.com/spf13/cobra"

	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/cli"
	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/cli/output"
	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/store"
	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/transport"
)

// NewCommand builds the `swipe logs` parent command.
func NewCommand(g *cli.GlobalFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logs",
		Short: "Inspect the mock's request log",
	}
	cmd.AddCommand(newTailCommand(g))
	cmd.AddCommand(newShowCommand(g))
	return cmd
}

func newTailCommand(g *cli.GlobalFlags) *cobra.Command {
	var n int
	cmd := &cobra.Command{
		Use:   "tail",
		Short: "Print the most recent request log entries",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ac := transport.NewAdminClient(baseURL(g))
			entries, err := ac.TailLogs(cmd.Context(), n)
			if err != nil {
				return fmt.Errorf("tail logs: %w", err)
			}
			return renderEntries(g, entries)
		},
	}
	cmd.Flags().IntVarP(&n, "n", "n", 20, "number of entries to show (newest first)")
	return cmd
}

func newShowCommand(g *cli.GlobalFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "show <request-id>",
		Short: "Show the full request/response pair for a request id",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ac := transport.NewAdminClient(baseURL(g))
			entry, err := ac.ShowLog(cmd.Context(), args[0])
			if err != nil {
				return fmt.Errorf("show log: %w", err)
			}
			return renderEntry(g, entry)
		},
	}
}

func baseURL(g *cli.GlobalFlags) string {
	port := g.Port
	if port == 0 {
		port = 8080
	}
	u := &url.URL{Scheme: "http", Host: fmt.Sprintf("localhost:%d", port)}
	return u.String()
}

func renderEntries(g *cli.GlobalFlags, entries []store.RequestLogEntry) error {
	r, err := g.Renderer()
	if err != nil {
		return err
	}
	if r.Format != output.FormatTable {
		return r.Object(entries)
	}
	if len(entries) == 0 {
		_, _ = fmt.Fprintln(g.Stdout, "no log entries")
		return nil
	}
	for i, e := range entries {
		if i > 0 {
			_, _ = fmt.Fprintln(g.Stdout)
		}
		if err := r.KV(output.KeyValueTable{Title: e.ID, Rows: rowsFor(e)}); err != nil {
			return err
		}
	}
	return nil
}

func renderEntry(g *cli.GlobalFlags, e store.RequestLogEntry) error {
	r, err := g.Renderer()
	if err != nil {
		return err
	}
	if r.Format != output.FormatTable {
		return r.Object(e)
	}
	return r.KV(output.KeyValueTable{Title: e.ID, Rows: rowsFor(e)})
}

func rowsFor(e store.RequestLogEntry) []output.KeyValueRow {
	return []output.KeyValueRow{
		{Key: "started_at", Value: e.StartedAt.Format(time.RFC3339)},
		{Key: "method", Value: e.Method},
		{Key: "path", Value: e.Path},
		{Key: "status", Value: e.Status},
		{Key: "duration_ms", Value: e.DurationMS},
		{Key: "client", Value: e.ClientID},
		{Key: "merchant", Value: e.MerchantID},
	}
}
