package mock

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/cli"
	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/cli/output"
	pathconfig "github.com/BML-Digital/swipe-merchants-dev/cli/internal/config"
	mockpkg "github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock"
	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/admin"
	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/auth"
	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/handlers"
	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/store"
	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/validator"
	specpkg "github.com/BML-Digital/swipe-merchants-dev/cli/internal/spec"
	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/transport"
)

// NewCommand builds the `swipe mock` parent command.
func NewCommand(g *cli.GlobalFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mock",
		Short: "Manage the embedded mock server",
	}
	cmd.AddCommand(newStartCommand(g))
	cmd.AddCommand(newStopCommand(g))
	cmd.AddCommand(newStatusCommand(g))
	cmd.AddCommand(newScenariosCommand(g))
	return cmd
}

func newStartCommand(g *cli.GlobalFlags) *cobra.Command {
	var (
		webhookURL string
		paymentTTL time.Duration
		qrFormat   string
	)
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Boot the embedded mock server (foreground)",
		Long: "Boot the embedded mock server. Foreground only: the command blocks\n" +
			"and prints the seed summary; use Ctrl+C or `swipe mock stop` from\n" +
			"another terminal to stop it.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			parsed, err := parseQRFormat(qrFormat)
			if err != nil {
				return err
			}
			return runStart(cmd.Context(), g, webhookURL, paymentTTL, parsed)
		},
	}
	cmd.Flags().StringVar(&webhookURL, "webhook-url", "", "URL the mock should POST webhook deliveries to (Phase 5+)")
	cmd.Flags().DurationVar(&paymentTTL, "payment-ttl", 0, "override default PENDING->COMPLETED transition (e.g. 2s); 0 = scaffold default 60s")
	cmd.Flags().StringVar(&qrFormat, "qr-format", "json", "QR-type payments encode `json` (pay-page URL + payment details, default) or `emvco` (production-shape EMVCo MPM)")
	return cmd
}

// parseQRFormat validates the --qr-format value. Unknown values fail
// early at flag-parse time rather than silently defaulting at request
// time, so a typo doesn't ship a misconfigured mock.
func parseQRFormat(s string) (handlers.QRFormat, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "", "json":
		return handlers.QRFormatJSON, nil
	case "emvco":
		return handlers.QRFormatEMVCo, nil
	default:
		return "", fmt.Errorf("--qr-format: %q is not one of {json, emvco}", s)
	}
}

func newStopCommand(g *cli.GlobalFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "stop",
		Short: "Stop the running mock server (graceful SIGTERM)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runStop(cmd.Context(), g)
		},
	}
}

func newStatusCommand(g *cli.GlobalFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Print the running mock server's status",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runStatus(cmd.Context(), g)
		},
	}
}

func runStart(ctx context.Context, g *cli.GlobalFlags, webhookURL string, paymentTTL time.Duration, qrFormat handlers.QRFormat) error {
	paths, err := pathconfig.DefaultPaths()
	if err != nil {
		return fmt.Errorf("resolve paths: %w", err)
	}
	if err := paths.EnsureDirs(); err != nil {
		return fmt.Errorf("ensure dirs: %w", err)
	}

	pidFile := mockpkg.NewPIDFile(paths.MockPIDFile, paths.MockSocketFile)
	if alive, pid, addr := pidFile.IsAlive(); alive {
		return fmt.Errorf("mock already running (pid %d, address %s); run `swipe mock stop` first", pid, addr)
	}

	st, err := store.Open(paths.MockStateDB)
	if err != nil {
		return fmt.Errorf("open state: %w", err)
	}
	defer func() { _ = st.Close() }()

	summary, err := st.SeedDefaults()
	if err != nil {
		return fmt.Errorf("seed defaults: %w", err)
	}

	signingKey, err := auth.LoadOrCreateSigningKey(paths.MockKeysDir)
	if err != nil {
		return fmt.Errorf("load signing key: %w", err)
	}

	loadedSpec, err := specpkg.Load()
	if err != nil {
		return fmt.Errorf("load embedded spec: %w", err)
	}
	specValidator, err := validator.New(loadedSpec.RawYAML())
	if err != nil {
		return fmt.Errorf("build spec validator: %w", err)
	}

	logger := slog.New(slog.NewJSONHandler(g.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))
	listenAddr := fmt.Sprintf(":%d", resolvePort(g))
	server, err := mockpkg.New(mockpkg.Config{
		ListenAddr:    listenAddr,
		WebhookURL:    webhookURL,
		Store:         st,
		Logger:        logger,
		StartedAt:     time.Now().UTC(),
		SigningKey:    signingKey,
		SpecValidator: specValidator,
		PaymentTTL:    paymentTTL,
		QRFormat:      qrFormat,
	})
	if err != nil {
		return fmt.Errorf("build server: %w", err)
	}

	if err := server.Bind(); err != nil {
		return fmt.Errorf("bind: %w", err)
	}
	bound := server.Listener().Addr().String()
	if err := pidFile.Write(bound); err != nil {
		return fmt.Errorf("write pidfile: %w", err)
	}
	defer func() { _ = pidFile.Remove() }()

	printSeedSummary(g.Stdout, bound, webhookURL, summary)

	srvCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	startErrCh := make(chan error, 1)
	go func() { startErrCh <- server.Serve(srvCtx) }()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	select {
	case sig := <-sigCh:
		_, _ = fmt.Fprintf(g.Stderr, "\nreceived %s, shutting down\n", sig)
		cancel()
	case err := <-startErrCh:
		return err
	}

	if err := <-startErrCh; err != nil {
		return err
	}
	return nil
}

// resolvePort returns the listen port for the mock. Honors --port; falls
// back to the documented default of 8080 (D-025).
func resolvePort(g *cli.GlobalFlags) int {
	if g.Port > 0 {
		return g.Port
	}
	return 8080
}

func printSeedSummary(w io.Writer, addr, webhookURL string, s store.SeedSummary) {
	displayAddr := addr
	if strings.HasPrefix(displayAddr, ":") || strings.HasPrefix(displayAddr, "0.0.0.0:") {
		displayAddr = "localhost" + strings.TrimPrefix(displayAddr, "0.0.0.0")
	}
	mvr := s.Wallet.Balances[store.CurrencyMVR]
	usd := s.Wallet.Balances[store.CurrencyUSD]
	header := "mock running"
	if !s.FreshSeed {
		header = "mock running (state restored from disk)"
	}
	_, _ = fmt.Fprintf(w, "%s on http://%s\n", header, displayAddr)
	_, _ = fmt.Fprintf(w, "  merchant:       %s\n", s.Merchant.ID)
	_, _ = fmt.Fprintf(w, "  wallet (MVR):   %10.2f available\n", mvr.Available)
	_, _ = fmt.Fprintf(w, "  wallet (USD):   %10.2f available\n", usd.Available)
	_, _ = fmt.Fprintf(w, "  bank accounts:  %d active\n", len(s.BankAccounts))
	_, _ = fmt.Fprintf(w, "  webhook secret: %s\n", s.WebhookSecret)
	_, _ = fmt.Fprintf(w, "  oauth issuer:   http://%s/  (Phase 2+)\n", displayAddr)
	if webhookURL != "" {
		_, _ = fmt.Fprintf(w, "  webhook url:    %s  (Phase 5+)\n", webhookURL)
	}
	_, _ = fmt.Fprintln(w)
	_, _ = fmt.Fprintln(w, "create your first OAuth client:  (available in Phase 2)")
	_, _ = fmt.Fprintln(w, "  swipe keys create --name \"My Test\" --scopes wallet:balance,payments:qr")
}

func runStop(_ context.Context, g *cli.GlobalFlags) error {
	paths, err := pathconfig.DefaultPaths()
	if err != nil {
		return fmt.Errorf("resolve paths: %w", err)
	}
	pidFile := mockpkg.NewPIDFile(paths.MockPIDFile, paths.MockSocketFile)
	if err := pidFile.SignalStop(); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			_, _ = fmt.Fprintln(g.Stdout, "no mock recorded as running")
			return nil
		}
		return fmt.Errorf("stop mock: %w", err)
	}
	if err := pidFile.WaitUntilStopped(5 * time.Second); err != nil {
		return fmt.Errorf("await mock stop: %w", err)
	}
	_ = pidFile.Remove()
	_, _ = fmt.Fprintln(g.Stdout, "mock stopped")
	return nil
}

// StatusResult is the renderable wrapper around admin.StatusResponse used
// by `swipe mock status`. It includes the on-disk pidfile state for cases
// where the recorded mock is not actually responding.
type StatusResult struct {
	Running       bool                  `json:"running" yaml:"running"`
	PID           int                   `json:"pid,omitempty" yaml:"pid,omitempty"`
	ListenAddress string                `json:"listen_address,omitempty" yaml:"listen_address,omitempty"`
	Detail        *admin.StatusResponse `json:"detail,omitempty" yaml:"detail,omitempty"`
}

func runStatus(ctx context.Context, g *cli.GlobalFlags) error {
	paths, err := pathconfig.DefaultPaths()
	if err != nil {
		return fmt.Errorf("resolve paths: %w", err)
	}
	pidFile := mockpkg.NewPIDFile(paths.MockPIDFile, paths.MockSocketFile)
	alive, pid, addr := pidFile.IsAlive()
	res := StatusResult{Running: alive, PID: pid, ListenAddress: addr}

	r, err := g.Renderer()
	if err != nil {
		return err
	}

	if alive && addr != "" {
		base := normalizeBaseURL(addr)
		client := transport.NewAdminClient(base)
		statusCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
		defer cancel()
		detail, err := client.Status(statusCtx)
		if err == nil {
			res.Detail = &detail
		}
	}

	if r.Format == output.FormatTable {
		return renderStatusTable(g, res)
	}
	return r.Object(res)
}

func normalizeBaseURL(addr string) string {
	if rest, ok := strings.CutPrefix(addr, "[::]"); ok {
		return "http://localhost" + rest
	}
	if strings.HasPrefix(addr, ":") {
		return "http://localhost" + addr
	}
	if rest, ok := strings.CutPrefix(addr, "0.0.0.0"); ok {
		return "http://localhost" + rest
	}
	return "http://" + addr
}

func renderStatusTable(g *cli.GlobalFlags, res StatusResult) error {
	r, err := g.Renderer()
	if err != nil {
		return err
	}
	rows := []output.KeyValueRow{
		{Key: "running", Value: res.Running},
	}
	if res.Running {
		rows = append(rows,
			output.KeyValueRow{Key: "pid", Value: res.PID},
			output.KeyValueRow{Key: "listen_address", Value: res.ListenAddress},
		)
	}
	if res.Detail != nil {
		rows = append(rows,
			output.KeyValueRow{Key: "binary_version", Value: res.Detail.Version.Version},
			output.KeyValueRow{Key: "spec_version", Value: res.Detail.Version.SpecVersion},
			output.KeyValueRow{Key: "started_at", Value: res.Detail.StartedAt.Format(time.RFC3339)},
			output.KeyValueRow{Key: "uptime_seconds", Value: res.Detail.UptimeSeconds},
			output.KeyValueRow{Key: "merchants", Value: res.Detail.Counts.Merchants},
			output.KeyValueRow{Key: "wallets", Value: res.Detail.Counts.Wallets},
			output.KeyValueRow{Key: "bank_accounts", Value: res.Detail.Counts.BankAccounts},
			output.KeyValueRow{Key: "webhook_secret", Value: res.Detail.WebhookSecret},
		)
		if res.Detail.WebhookURL != "" {
			rows = append(rows, output.KeyValueRow{Key: "webhook_url", Value: res.Detail.WebhookURL})
		}
	}
	return r.KV(output.KeyValueTable{Title: "swipe mock status", Rows: rows})
}
