package webhooks

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/cli"
	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/webhooks"
	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/transport"
)

// NewCommand builds the `swipe webhooks` parent command.
func NewCommand(g *cli.GlobalFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "webhooks",
		Short: "Webhook tooling: listen + verify + secret",
	}
	cmd.AddCommand(newListenCommand(g))
	cmd.AddCommand(newVerifyCommand(g))
	cmd.AddCommand(newSecretCommand(g))
	return cmd
}

func newListenCommand(g *cli.GlobalFlags) *cobra.Command {
	var (
		listenAddr string
		forwardTo  string
		secret     string
		skipVerify bool
		printBody  bool
	)
	cmd := &cobra.Command{
		Use:   "listen",
		Short: "Receive webhooks locally, verify signatures, optionally forward",
		Long: "Spin up a small HTTP receiver that prints incoming webhook deliveries\n" +
			"with their signature-verification status. Use --forward-to to relay\n" +
			"each verified delivery to a downstream endpoint (typically your real\n" +
			"webhook handler running locally).",
		RunE: func(cmd *cobra.Command, _ []string) error {
			s := strings.TrimSpace(secret)
			if s == "" {
				resolved, err := readMockSecret(cmd.Context(), g)
				if err != nil {
					return err
				}
				s = resolved
			}
			cfg := listenConfig{
				addr:       listenAddr,
				forwardTo:  forwardTo,
				secret:     s,
				skipVerify: skipVerify,
				printBody:  printBody,
				out:        g.Stdout,
				errOut:     g.Stderr,
			}
			return runListen(cmd.Context(), cfg)
		},
	}
	cmd.Flags().StringVar(&listenAddr, "listen", ":9000", "local address to listen on")
	cmd.Flags().StringVar(&forwardTo, "forward-to", "", "downstream URL to forward verified deliveries to")
	cmd.Flags().StringVar(&secret, "secret", "", "webhook secret (default: read from the running mock)")
	cmd.Flags().BoolVar(&skipVerify, "skip-verify", false, "accept deliveries without verifying the signature")
	cmd.Flags().BoolVar(&printBody, "print-body", true, "pretty-print the JSON envelope")
	return cmd
}

func newVerifyCommand(g *cli.GlobalFlags) *cobra.Command {
	var (
		file      string
		idHeader  string
		tsHeader  string
		sigHeader string
		secret    string
	)
	cmd := &cobra.Command{
		Use:   "verify",
		Short: "Offline verify a webhook payload + signature",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if strings.TrimSpace(secret) == "" {
				return fmt.Errorf("--secret is required")
			}
			if strings.TrimSpace(idHeader) == "" || strings.TrimSpace(tsHeader) == "" || strings.TrimSpace(sigHeader) == "" {
				return fmt.Errorf("--id, --timestamp, and --signature are required")
			}
			body, err := readPayload(cmd.InOrStdin(), file)
			if err != nil {
				return err
			}
			err = webhooks.Verify(secret, idHeader, tsHeader, sigHeader, body, time.Now().UTC())
			if err != nil {
				_, _ = fmt.Fprintf(g.Stdout, "invalid: %v\n", err)
				return err
			}
			_, _ = fmt.Fprintln(g.Stdout, "valid")
			return nil
		},
	}
	cmd.Flags().StringVarP(&file, "file", "f", "-", "path to the JSON payload (default stdin)")
	cmd.Flags().StringVar(&idHeader, "id", "", "value of the webhook-id header")
	cmd.Flags().StringVar(&tsHeader, "timestamp", "", "value of the webhook-timestamp header (unix seconds)")
	cmd.Flags().StringVar(&sigHeader, "signature", "", "value of the webhook-signature header")
	cmd.Flags().StringVar(&secret, "secret", "", "webhook signing secret")
	return cmd
}

func newSecretCommand(g *cli.GlobalFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "secret",
		Short: "Print the running mock's webhook secret",
		RunE: func(cmd *cobra.Command, _ []string) error {
			secret, err := readMockSecret(cmd.Context(), g)
			if err != nil {
				return err
			}
			_, _ = fmt.Fprintln(g.Stdout, secret)
			return nil
		},
	}
}

// listenConfig bundles the parsed flags for the listen command.
type listenConfig struct {
	addr       string
	forwardTo  string
	secret     string
	skipVerify bool
	printBody  bool
	out        io.Writer
	errOut     io.Writer
}

func runListen(ctx context.Context, cfg listenConfig) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", handleListen(cfg))
	server := &http.Server{
		Addr:              cfg.addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
	ln, err := net.Listen("tcp", cfg.addr)
	if err != nil {
		return fmt.Errorf("listen on %s: %w", cfg.addr, err)
	}
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()
	_, _ = fmt.Fprintf(cfg.out, "webhooks listen on http://%s (waiting for deliveries — Ctrl+C to stop)\n", ln.Addr().String())
	if cfg.forwardTo != "" {
		_, _ = fmt.Fprintf(cfg.out, "forwarding verified deliveries to %s\n", cfg.forwardTo)
	}
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(signalCh)
	go func() {
		<-signalCh
		_ = server.Shutdown(context.Background())
	}()
	if err := server.Serve(ln); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("serve: %w", err)
	}
	return nil
}

func handleListen(cfg listenConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "read body", http.StatusBadRequest)
			return
		}
		defer func() { _ = r.Body.Close() }()
		id, ts, sig := webhooks.ExtractHeaders(r.Header)
		status := "verified"
		var verifyErr error
		if !cfg.skipVerify {
			verifyErr = webhooks.Verify(cfg.secret, id, ts, sig, body, time.Now().UTC())
			if verifyErr != nil {
				status = "invalid: " + verifyErr.Error()
			}
		} else {
			status = "skipped"
		}
		_, _ = fmt.Fprintf(cfg.out, "→ %s %s  webhook-id=%s  signature=%s\n", r.Method, r.URL.Path, id, status)
		if cfg.printBody {
			_, _ = fmt.Fprintln(cfg.out, prettyJSON(body))
		}
		if cfg.forwardTo != "" && (cfg.skipVerify || verifyErr == nil) {
			if err := forwardDelivery(r, body, cfg.forwardTo); err != nil {
				_, _ = fmt.Fprintf(cfg.errOut, "forward: %v\n", err)
			}
		}
		if verifyErr != nil {
			http.Error(w, verifyErr.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func forwardDelivery(orig *http.Request, body []byte, target string) error {
	req, err := http.NewRequestWithContext(context.Background(), orig.Method, target, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("build forward request: %w", err)
	}
	// Replay all of the original's headers so the downstream sees the
	// same signature material it would have if the mock had hit it
	// directly.
	for k, v := range orig.Header {
		req.Header[k] = v
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("forward POST: %w", err)
	}
	_ = resp.Body.Close()
	return nil
}

// readMockSecret fetches the running mock's webhook secret via /_admin/status.
func readMockSecret(ctx context.Context, g *cli.GlobalFlags) (string, error) {
	ac := transport.NewAdminClient(baseURL(g))
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	status, err := ac.Status(ctx)
	if err != nil {
		return "", fmt.Errorf("read mock secret: %w (is the mock running?)", err)
	}
	if status.WebhookSecret == "" {
		return "", fmt.Errorf("mock returned no webhook secret")
	}
	return status.WebhookSecret, nil
}

func baseURL(g *cli.GlobalFlags) string {
	port := g.Port
	if port == 0 {
		port = 8080
	}
	u := &url.URL{Scheme: "http", Host: fmt.Sprintf("localhost:%d", port)}
	return u.String()
}

func readPayload(stdin io.Reader, file string) ([]byte, error) {
	if file == "-" {
		return io.ReadAll(stdin)
	}
	return os.ReadFile(file) // #nosec G304 -- caller-supplied path
}

func prettyJSON(b []byte) string {
	var v any
	if err := json.Unmarshal(b, &v); err != nil {
		return string(b)
	}
	out, _ := json.MarshalIndent(v, "  ", "  ")
	return "  " + string(out)
}
