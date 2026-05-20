package auth

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/cli"
	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/cli/output"
	pathconfig "github.com/BML-Digital/swipe-merchants-dev/cli/internal/config"
	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/transport"
)

// NewCommand builds the `swipe auth` parent command.
func NewCommand(g *cli.GlobalFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "OAuth2 client credentials login + token cache",
		Long: "Exchange a client_id + client_secret pair for an access token,\n" +
			"cache it at ~/.swipe/token.json, and query /api/v1/whoami.",
	}
	cmd.AddCommand(newLoginCommand(g))
	cmd.AddCommand(newTokenCommand(g))
	cmd.AddCommand(newWhoamiCommand(g))
	cmd.AddCommand(newLogoutCommand(g))
	return cmd
}

func newLoginCommand(g *cli.GlobalFlags) *cobra.Command {
	var clientID, clientSecret string
	var scopes []string
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Exchange client_credentials for an access token; cache to disk",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if strings.TrimSpace(clientID) == "" || strings.TrimSpace(clientSecret) == "" {
				return fmt.Errorf("--client-id and --client-secret are required")
			}
			base := baseURL(g)
			tc := transport.NewClient(base)
			resp, err := tc.IssueToken(cmd.Context(), clientID, clientSecret, scopes)
			if err != nil {
				return fmt.Errorf("login: %w", err)
			}
			paths, err := pathsOrErr()
			if err != nil {
				return err
			}
			cached := transport.CachedToken{
				AccessToken: resp.AccessToken,
				TokenType:   resp.TokenType,
				ExpiresAt:   resp.ExpiresAt(),
				Scope:       resp.Scope,
				ClientID:    clientID,
				IssuerURL:   base,
			}
			if err := transport.SaveCachedToken(paths.TokenCache, cached); err != nil {
				return err
			}
			return renderTokenSummary(g, cached, false)
		},
	}
	cmd.Flags().StringVar(&clientID, "client-id", "", "OAuth client id (cli_*)")
	cmd.Flags().StringVar(&clientSecret, "client-secret", "", "OAuth client secret (sec_*)")
	cmd.Flags().StringSliceVar(&scopes, "scopes", nil, "comma-separated scope request (default: all client scopes)")
	return cmd
}

func newTokenCommand(g *cli.GlobalFlags) *cobra.Command {
	var show bool
	cmd := &cobra.Command{
		Use:   "token",
		Short: "Print the cached access token (masked unless --show)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			paths, err := pathsOrErr()
			if err != nil {
				return err
			}
			cached, err := transport.LoadCachedToken(paths.TokenCache)
			if err != nil {
				if errors.Is(err, os.ErrNotExist) {
					return fmt.Errorf("no cached token; run `swipe auth login` first")
				}
				return fmt.Errorf("load token cache: %w", err)
			}
			return renderTokenSummary(g, cached, show)
		},
	}
	cmd.Flags().BoolVar(&show, "show", false, "show the full access token instead of a mask")
	return cmd
}

func newWhoamiCommand(g *cli.GlobalFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "whoami",
		Short: "Call /api/v1/whoami with the cached token",
		RunE: func(cmd *cobra.Command, _ []string) error {
			paths, err := pathsOrErr()
			if err != nil {
				return err
			}
			cached, err := transport.LoadCachedToken(paths.TokenCache)
			if err != nil {
				if errors.Is(err, os.ErrNotExist) {
					return fmt.Errorf("no cached token; run `swipe auth login` first")
				}
				return fmt.Errorf("load token cache: %w", err)
			}
			if cached.IsExpired(time.Now().UTC()) {
				return fmt.Errorf("cached token expired at %s; run `swipe auth login` again", cached.ExpiresAt.Format(time.RFC3339))
			}
			tc := transport.NewClient(baseURL(g))
			who, err := tc.WhoAmI(cmd.Context(), cached.AccessToken)
			if err != nil {
				return fmt.Errorf("whoami: %w", err)
			}
			return renderWhoAmI(g, who, int(time.Until(cached.ExpiresAt).Seconds()))
		},
	}
}

func newLogoutCommand(g *cli.GlobalFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Clear the cached access token",
		RunE: func(_ *cobra.Command, _ []string) error {
			paths, err := pathsOrErr()
			if err != nil {
				return err
			}
			if err := transport.ClearCachedToken(paths.TokenCache); err != nil {
				return err
			}
			_, _ = fmt.Fprintln(g.Stdout, "cleared cached token")
			return nil
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

func pathsOrErr() (pathconfig.Paths, error) {
	p, err := pathconfig.DefaultPaths()
	if err != nil {
		return pathconfig.Paths{}, fmt.Errorf("resolve paths: %w", err)
	}
	return p, nil
}

// renderTokenSummary prints the cached token's metadata; the raw token is
// masked unless show=true. Per D-023 secrets are masked by default.
func renderTokenSummary(g *cli.GlobalFlags, t transport.CachedToken, show bool) error {
	r, err := g.Renderer()
	if err != nil {
		return err
	}
	displayed := maskToken(t.AccessToken)
	if show {
		displayed = t.AccessToken
	}
	if r.Format != output.FormatTable {
		type withMask struct {
			transport.CachedToken
			AccessToken string `json:"access_token" yaml:"access_token"`
		}
		w := withMask{CachedToken: t, AccessToken: displayed}
		w.CachedToken.AccessToken = "" // avoid duplicate field, masking is in the outer
		return r.Object(w)
	}
	rows := []output.KeyValueRow{
		{Key: "client_id", Value: t.ClientID},
		{Key: "token_type", Value: t.TokenType},
		{Key: "scope", Value: t.Scope},
		{Key: "expires_at", Value: t.ExpiresAt.Format(time.RFC3339)},
		{Key: "issuer", Value: t.IssuerURL},
		{Key: "access_token", Value: displayed},
	}
	return r.KV(output.KeyValueTable{Title: "swipe auth token", Rows: rows})
}

func renderWhoAmI(g *cli.GlobalFlags, who transport.WhoAmIResponse, expiresInSec int) error {
	r, err := g.Renderer()
	if err != nil {
		return err
	}
	if r.Format != output.FormatTable {
		out := struct {
			transport.WhoAmIResponse
			ExpiresIn int `json:"expires_in_seconds" yaml:"expires_in_seconds"`
		}{WhoAmIResponse: who, ExpiresIn: expiresInSec}
		return r.Object(out)
	}
	rows := []output.KeyValueRow{
		{Key: "client_id", Value: who.ClientID},
		{Key: "merchant_id", Value: who.MerchantID},
		{Key: "scopes", Value: strings.Join(who.Scopes, " ")},
		{Key: "expires_in_seconds", Value: expiresInSec},
	}
	return r.KV(output.KeyValueTable{Title: "whoami", Rows: rows})
}

// maskToken replaces all but the last 6 characters of the access token
// with a fixed mask, keeping a tail snippet for visual confirmation that
// the cached token matches what was just minted.
func maskToken(t string) string {
	if len(t) < 12 {
		return strings.Repeat("*", len(t))
	}
	return "****" + t[len(t)-6:]
}
