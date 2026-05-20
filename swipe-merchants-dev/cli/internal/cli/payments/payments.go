package payments

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/cli"
	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/cli/output"
	pathconfig "github.com/BML-Digital/swipe-merchants-dev/cli/internal/config"
	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/handlers"
	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/transport"
)

// NewCommand builds the `swipe payments` parent command.
func NewCommand(g *cli.GlobalFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "payments",
		Short: "Create, fetch, and watch payments",
	}
	cmd.AddCommand(newCreateCommand(g))
	cmd.AddCommand(newGetCommand(g))
	cmd.AddCommand(newWatchCommand(g))
	return cmd
}

func newCreateCommand(g *cli.GlobalFlags) *cobra.Command {
	var (
		amount       float64
		currency     string
		paymentType  string
		description  string
		recipientVPA string
		fromFile     string
	)
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a payment (requires per-type scope: payments:qr | payments:contact | payments:link)",
		Long: "Create a payment. Provide flags or pipe a JSON body via -f. The mock\n" +
			"requires the per-type scope matching --type (D-013).",
		RunE: func(cmd *cobra.Command, _ []string) error {
			req, err := buildPaymentRequest(amount, currency, paymentType, description, recipientVPA, fromFile)
			if err != nil {
				return err
			}
			token, err := loadToken()
			if err != nil {
				return err
			}
			tc := transport.NewClient(baseURL(g))
			resp, err := tc.CreatePayment(cmd.Context(), token.AccessToken, req)
			if err != nil {
				return fmt.Errorf("create payment: %w", err)
			}
			return renderPayment(g, resp)
		},
	}
	cmd.Flags().Float64Var(&amount, "amount", 0, "payment amount (required unless -f)")
	cmd.Flags().StringVar(&currency, "currency", "MVR", "currency code (MVR or USD)")
	cmd.Flags().StringVar(&paymentType, "type", "QR", "payment type: QR | CONTACT | LINK")
	cmd.Flags().StringVar(&description, "description", "", "human-readable description")
	cmd.Flags().StringVar(&recipientVPA, "recipient-vpa", "", "VPA of recipient (required for CONTACT)")
	cmd.Flags().StringVarP(&fromFile, "file", "f", "", "read JSON request body from this file ('-' for stdin)")
	return cmd
}

func newGetCommand(g *cli.GlobalFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "get <id>",
		Short: "Get a payment by id (requires scope: transactions:status)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			token, err := loadToken()
			if err != nil {
				return err
			}
			tc := transport.NewClient(baseURL(g))
			body, status, err := tc.GetPayment(cmd.Context(), token.AccessToken, args[0])
			if err != nil {
				return fmt.Errorf("get payment: %w", err)
			}
			if status >= 400 {
				return fmt.Errorf("get payment: status %d: %s", status, strings.TrimSpace(string(body)))
			}
			var pretty any
			if err := json.Unmarshal(body, &pretty); err == nil {
				out, _ := json.MarshalIndent(pretty, "", "  ")
				_, _ = fmt.Fprintln(g.Stdout, string(out))
				return nil
			}
			_, _ = fmt.Fprintln(g.Stdout, string(body))
			return nil
		},
	}
}

func newWatchCommand(g *cli.GlobalFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "watch <id>",
		Short: "Stream payment status transitions via SSE",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			token, err := loadToken()
			if err != nil {
				return err
			}
			tc := transport.NewClient(baseURL(g))
			start := time.Now()
			return tc.WatchPayment(cmd.Context(), token.AccessToken, args[0], func(ev transport.StreamEvent) bool {
				elapsed := time.Since(start).Truncate(100 * time.Millisecond)
				_, _ = fmt.Fprintf(g.Stdout, "%s  %s  %s\n", elapsed, ev.Status, ev.ID)
				switch ev.Status {
				case "COMPLETED", "EXPIRED", "CANCELLED":
					return false
				default:
					return true
				}
			})
		},
	}
}

func buildPaymentRequest(amount float64, currency, paymentType, description, recipientVPA, file string) (handlers.CreatePaymentRequest, error) {
	if file != "" {
		raw, err := readRequestFile(file)
		if err != nil {
			return handlers.CreatePaymentRequest{}, err
		}
		var req handlers.CreatePaymentRequest
		if err := json.Unmarshal(raw, &req); err != nil {
			return handlers.CreatePaymentRequest{}, fmt.Errorf("parse %s: %w", file, err)
		}
		return req, nil
	}
	if amount <= 0 {
		return handlers.CreatePaymentRequest{}, fmt.Errorf("--amount is required (or use -f)")
	}
	return handlers.CreatePaymentRequest{
		Amount:       amount,
		Currency:     currency,
		Type:         paymentType,
		Description:  description,
		RecipientVPA: recipientVPA,
	}, nil
}

func readRequestFile(file string) ([]byte, error) {
	if file == "-" {
		return io.ReadAll(os.Stdin)
	}
	return os.ReadFile(file) // #nosec G304 -- file comes from the user, intentional
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

func renderPayment(g *cli.GlobalFlags, p handlers.PaymentResponse) error {
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
		{Key: "currency", Value: p.Currency},
		{Key: "status", Value: p.Status},
		{Key: "short_code", Value: p.ShortCode},
		{Key: "created_at", Value: p.CreatedAt.Format(time.RFC3339)},
	}
	if p.QRData != "" {
		rows = append(rows, output.KeyValueRow{Key: "qr_data", Value: shortLabel(p.QRData)})
	}
	if p.PaymentURL != "" {
		rows = append(rows, output.KeyValueRow{Key: "payment_url", Value: p.PaymentURL})
	}
	return r.KV(output.KeyValueTable{Title: "payment", Rows: rows})
}

func shortLabel(s string) string {
	if len(s) <= 40 {
		return s
	}
	return s[:20] + "…" + s[len(s)-12:] + " (" + lengthSuffix(len(s)) + ")"
}

func lengthSuffix(n int) string {
	return fmt.Sprintf("%d chars", n)
}
