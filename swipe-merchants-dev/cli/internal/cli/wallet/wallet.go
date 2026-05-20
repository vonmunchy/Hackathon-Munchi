package wallet

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

// NewCommand builds the `swipe wallet` parent command.
func NewCommand(g *cli.GlobalFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "wallet",
		Short: "Read wallet balance and linked bank accounts",
	}
	cmd.AddCommand(newBalanceCommand(g))
	cmd.AddCommand(newAccountsCommand(g))
	return cmd
}

func newBalanceCommand(g *cli.GlobalFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "balance",
		Short: "Get wallet balance (requires scope: wallet:balance)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			token, err := loadToken()
			if err != nil {
				return err
			}
			tc := transport.NewClient(baseURL(g))
			balances, err := tc.Balance(cmd.Context(), token.AccessToken)
			if err != nil {
				return fmt.Errorf("balance: %w", err)
			}
			return renderBalances(g, balances)
		},
	}
}

func newAccountsCommand(g *cli.GlobalFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "accounts",
		Short: "List linked bank accounts (requires scope: wallet:balance)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			token, err := loadToken()
			if err != nil {
				return err
			}
			tc := transport.NewClient(baseURL(g))
			accounts, err := tc.BankAccounts(cmd.Context(), token.AccessToken)
			if err != nil {
				return fmt.Errorf("bank-accounts: %w", err)
			}
			return renderAccounts(g, accounts)
		},
	}
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

func renderBalances(g *cli.GlobalFlags, balances []handlers.BalanceResponse) error {
	r, err := g.Renderer()
	if err != nil {
		return err
	}
	if r.Format != output.FormatTable {
		return r.Object(balances)
	}
	if len(balances) == 0 {
		_, _ = fmt.Fprintln(g.Stdout, "no balances")
		return nil
	}
	for i, b := range balances {
		if i > 0 {
			_, _ = fmt.Fprintln(g.Stdout)
		}
		rows := []output.KeyValueRow{
			{Key: "currency", Value: b.Currency},
			{Key: "available_balance", Value: fmt.Sprintf("%.2f", b.AvailableBalance)},
			{Key: "pending_balance", Value: fmt.Sprintf("%.2f", b.PendingBalance)},
		}
		if err := r.KV(output.KeyValueTable{Title: b.Currency + " wallet", Rows: rows}); err != nil {
			return err
		}
	}
	return nil
}

func renderAccounts(g *cli.GlobalFlags, accounts []handlers.BankAccountResponse) error {
	r, err := g.Renderer()
	if err != nil {
		return err
	}
	if r.Format != output.FormatTable {
		return r.Object(accounts)
	}
	if len(accounts) == 0 {
		_, _ = fmt.Fprintln(g.Stdout, "no bank accounts")
		return nil
	}
	for i, a := range accounts {
		if i > 0 {
			_, _ = fmt.Fprintln(g.Stdout)
		}
		rows := []output.KeyValueRow{
			{Key: "id", Value: a.ID},
			{Key: "account_number", Value: a.AccountNumber},
			{Key: "account_holder_name", Value: a.AccountHolderName},
			{Key: "currency", Value: a.Currency},
			{Key: "status", Value: a.Status},
		}
		if err := r.KV(output.KeyValueTable{Title: a.ID, Rows: rows}); err != nil {
			return err
		}
	}
	return nil
}
