package transactions

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/cli"
	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/cli/output"
	pathconfig "github.com/BML-Digital/swipe-merchants-dev/cli/internal/config"
	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/handlers"
	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/transport"
)

// NewCommand builds the `swipe transactions` parent command.
func NewCommand(g *cli.GlobalFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "transactions",
		Short: "List transactions (payments, payouts, adjustments)",
	}
	cmd.AddCommand(newHistoryCommand(g))
	cmd.AddCommand(newGetCommand(g))
	return cmd
}

func newGetCommand(g *cli.GlobalFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "get <reference>",
		Short: "Look up a transaction by reference (requires scope: transactions:status)",
		Long: "Look up a transaction by its merchant-facing reference. For\n" +
			"payments this is the value returned as `reference` on the\n" +
			"createPayment response (the short code). For payouts it is the\n" +
			"payout id.",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			token, err := loadToken()
			if err != nil {
				return err
			}
			tc := transport.NewClient(baseURL(g))
			txn, err := tc.GetTransaction(cmd.Context(), token.AccessToken, args[0])
			if err != nil {
				return fmt.Errorf("get transaction: %w", err)
			}
			return renderTransactionItem(g, txn)
		},
	}
}

func newHistoryCommand(g *cli.GlobalFlags) *cobra.Command {
	var limit, offset int
	cmd := &cobra.Command{
		Use:   "history",
		Short: "Get transaction history (requires scope: transactions:history)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			token, err := loadToken()
			if err != nil {
				return err
			}
			tc := transport.NewClient(baseURL(g))
			resp, err := tc.History(cmd.Context(), token.AccessToken, limit, offset)
			if err != nil {
				return fmt.Errorf("history: %w", err)
			}
			return renderHistory(g, resp)
		},
	}
	cmd.Flags().IntVar(&limit, "limit", 0, "max number of transactions (default 20)")
	cmd.Flags().IntVar(&offset, "offset", 0, "pagination offset (default 0)")
	return cmd
}

func loadToken() (transport.CachedToken, error) {
	paths, err := pathconfig.DefaultPaths()
	if err != nil {
		return transport.CachedToken{}, fmt.Errorf("resolve paths: %w", err)
	}
	tok, err := transport.LoadCachedToken(paths.TokenCache)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return transport.CachedToken{}, fmt.Errorf("no cached token; run `swipe auth login` first")
		}
		return transport.CachedToken{}, fmt.Errorf("load token cache: %w", err)
	}
	if tok.IsExpired(time.Now().UTC()) {
		return transport.CachedToken{}, fmt.Errorf("cached token expired at %s; run `swipe auth login` again", tok.ExpiresAt.Format(time.RFC3339))
	}
	return tok, nil
}

func baseURL(g *cli.GlobalFlags) string {
	port := g.Port
	if port == 0 {
		port = 8080
	}
	u := &url.URL{Scheme: "http", Host: fmt.Sprintf("localhost:%d", port)}
	return u.String()
}

func renderHistory(g *cli.GlobalFlags, resp handlers.HistoryResponse) error {
	r, err := g.Renderer()
	if err != nil {
		return err
	}
	if r.Format != output.FormatTable {
		return r.Object(resp)
	}
	if len(resp.Transactions) == 0 {
		_, _ = fmt.Fprintf(g.Stdout, "no transactions (total %d)\n", resp.Total)
		return nil
	}
	for i, txn := range resp.Transactions {
		if i > 0 {
			_, _ = fmt.Fprintln(g.Stdout)
		}
		rows := []output.KeyValueRow{
			{Key: "id", Value: txn.ID},
			{Key: "reference", Value: txn.Reference},
			{Key: "amount", Value: fmt.Sprintf("%.2f", txn.Amount)},
			{Key: "currency", Value: txn.Currency},
			{Key: "type", Value: txn.Type},
			{Key: "status", Value: txn.Status},
			{Key: "created_at", Value: txn.CreatedAt},
		}
		if err := r.KV(output.KeyValueTable{Title: txn.ID, Rows: rows}); err != nil {
			return err
		}
	}
	_, _ = fmt.Fprintf(g.Stdout, "\ntotal: %d\n", resp.Total)
	return nil
}

func renderTransactionItem(g *cli.GlobalFlags, t handlers.TransactionItem) error {
	r, err := g.Renderer()
	if err != nil {
		return err
	}
	if r.Format != output.FormatTable {
		return r.Object(t)
	}
	rows := []output.KeyValueRow{
		{Key: "id", Value: t.ID},
		{Key: "reference", Value: t.Reference},
		{Key: "type", Value: t.Type},
		{Key: "status", Value: t.Status},
		{Key: "amount", Value: fmt.Sprintf("%.2f", t.Amount)},
		{Key: "currency", Value: t.Currency},
		{Key: "gross_amount", Value: fmt.Sprintf("%.2f", t.GrossAmount)},
		{Key: "fee_amount", Value: fmt.Sprintf("%.2f", t.FeeAmount)},
		{Key: "net_amount", Value: fmt.Sprintf("%.2f", t.NetAmount)},
		{Key: "original_amount", Value: fmt.Sprintf("%.2f", t.OriginalAmount)},
		{Key: "created_at", Value: t.CreatedAt},
	}
	if t.Description != "" {
		rows = append(rows, output.KeyValueRow{Key: "description", Value: t.Description})
	}
	return r.KV(output.KeyValueTable{Title: "transaction", Rows: rows})
}
