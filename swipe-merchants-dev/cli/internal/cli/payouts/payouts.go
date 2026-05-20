package payouts

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
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

// NewCommand builds the `swipe payouts` parent command.
func NewCommand(g *cli.GlobalFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "payouts",
		Short: "Initiate payouts to a linked bank account",
	}
	cmd.AddCommand(newCreateCommand(g))
	return cmd
}

func newCreateCommand(g *cli.GlobalFlags) *cobra.Command {
	var (
		amount        float64
		bankAccountID string
		fromFile      string
	)
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a payout (requires scope: wallet:withdraw)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			req, err := buildPayoutRequest(amount, bankAccountID, fromFile)
			if err != nil {
				return err
			}
			token, err := loadToken()
			if err != nil {
				return err
			}
			tc := transport.NewClient(baseURL(g))
			resp, err := tc.CreatePayout(cmd.Context(), token.AccessToken, req)
			if err != nil {
				return fmt.Errorf("create payout: %w", err)
			}
			return renderPayout(g, resp)
		},
	}
	cmd.Flags().Float64Var(&amount, "amount", 0, "payout amount (required unless -f)")
	cmd.Flags().StringVar(&bankAccountID, "bank-account-id", "", "destination bank account id (required unless -f)")
	cmd.Flags().StringVarP(&fromFile, "file", "f", "", "read JSON request body from this file ('-' for stdin)")
	return cmd
}

func buildPayoutRequest(amount float64, bankAccountID, file string) (handlers.CreatePayoutRequest, error) {
	if file != "" {
		raw, err := readRequestFile(file)
		if err != nil {
			return handlers.CreatePayoutRequest{}, err
		}
		var req handlers.CreatePayoutRequest
		if err := json.Unmarshal(raw, &req); err != nil {
			return handlers.CreatePayoutRequest{}, fmt.Errorf("parse %s: %w", file, err)
		}
		return req, nil
	}
	if amount <= 0 {
		return handlers.CreatePayoutRequest{}, fmt.Errorf("--amount is required (or use -f)")
	}
	if bankAccountID == "" {
		return handlers.CreatePayoutRequest{}, fmt.Errorf("--bank-account-id is required (or use -f)")
	}
	return handlers.CreatePayoutRequest{Amount: amount, BankAccountID: bankAccountID}, nil
}

func readRequestFile(file string) ([]byte, error) {
	if file == "-" {
		return io.ReadAll(os.Stdin)
	}
	return os.ReadFile(file) // #nosec G304 -- file from user
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

func renderPayout(g *cli.GlobalFlags, p handlers.PayoutResponse) error {
	r, err := g.Renderer()
	if err != nil {
		return err
	}
	if r.Format != output.FormatTable {
		return r.Object(p)
	}
	rows := []output.KeyValueRow{
		{Key: "id", Value: p.ID},
		{Key: "amount", Value: fmt.Sprintf("%.2f", p.Amount)},
		{Key: "status", Value: p.Status},
		{Key: "created_at", Value: p.CreatedAt.Format(time.RFC3339)},
	}
	if p.Reference != "" {
		rows = append(rows, output.KeyValueRow{Key: "reference", Value: p.Reference})
	}
	return r.KV(output.KeyValueTable{Title: "payout", Rows: rows})
}
