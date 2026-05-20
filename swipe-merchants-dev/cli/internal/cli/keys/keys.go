package keys

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/cli"
	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/admin"
	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/transport"
)

// NewCommand builds the `swipe keys` parent command and attaches every
// subcommand. The parent itself does nothing without a subcommand.
func NewCommand(g *cli.GlobalFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "keys",
		Short: "Manage OAuth clients in the mock state store",
		Long: "Create, list, inspect, update, rotate, revoke, and delete OAuth\n" +
			"clients used by `swipe auth login`. Secrets are returned exactly\n" +
			"once on create or rotate — save them right away.",
	}
	cmd.AddCommand(newCreateCommand(g))
	cmd.AddCommand(newListCommand(g))
	cmd.AddCommand(newShowCommand(g))
	cmd.AddCommand(newUpdateCommand(g))
	cmd.AddCommand(newRotateCommand(g))
	cmd.AddCommand(newRevokeCommand(g))
	cmd.AddCommand(newDeleteCommand(g))
	cmd.AddCommand(newTestCommand(g))
	return cmd
}

func newCreateCommand(g *cli.GlobalFlags) *cobra.Command {
	var name, merchantID string
	var scopes []string
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new OAuth client",
		Long: "Create a new OAuth client. The plaintext client secret is\n" +
			"printed once on this response — save it now, it cannot be\n" +
			"recovered later.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if strings.TrimSpace(name) == "" {
				return fmt.Errorf("--name is required")
			}
			req := admin.CreateRequest{Name: name, MerchantID: merchantID, Scopes: scopes}
			ac, err := newAdminClient(g)
			if err != nil {
				return err
			}
			resp, err := ac.CreateClient(cmd.Context(), req)
			if err != nil {
				return fmt.Errorf("create client: %w", err)
			}
			return renderCreateOrRotate(g, "keys create", resp.Client, resp.Secret)
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "human-readable name for the client (required)")
	cmd.Flags().StringVar(&merchantID, "merchant", "", "merchant id the client belongs to (default mer_default)")
	cmd.Flags().StringSliceVar(&scopes, "scopes", nil, "comma-separated OAuth scopes (e.g. wallet:balance,payments:qr)")
	return cmd
}

func newListCommand(g *cli.GlobalFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List OAuth clients",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ac, err := newAdminClient(g)
			if err != nil {
				return err
			}
			items, err := ac.ListClients(cmd.Context())
			if err != nil {
				return fmt.Errorf("list clients: %w", err)
			}
			return renderClientList(g, items)
		},
	}
}

func newShowCommand(g *cli.GlobalFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "show <id>",
		Short: "Show a single OAuth client",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ac, err := newAdminClient(g)
			if err != nil {
				return err
			}
			c, err := ac.GetClient(cmd.Context(), args[0])
			if err != nil {
				return fmt.Errorf("show client: %w", err)
			}
			return renderClient(g, c)
		},
	}
}

func newUpdateCommand(g *cli.GlobalFlags) *cobra.Command {
	var name string
	var scopes []string
	var enabled bool
	var setName, setScopes, setEnabled bool
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update an OAuth client's name, scopes, or enabled flag",
		Args:  cobra.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			setName = cmd.Flags().Changed("name")
			setScopes = cmd.Flags().Changed("scopes")
			setEnabled = cmd.Flags().Changed("enabled")
			if !setName && !setScopes && !setEnabled {
				return fmt.Errorf("at least one of --name, --scopes, --enabled must be set")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ac, err := newAdminClient(g)
			if err != nil {
				return err
			}
			req := admin.UpdateRequest{}
			if setName {
				req.Name = &name
			}
			if setScopes {
				req.Scopes = &scopes
			}
			if setEnabled {
				req.Enabled = &enabled
			}
			c, err := ac.UpdateClient(cmd.Context(), args[0], req)
			if err != nil {
				return fmt.Errorf("update client: %w", err)
			}
			return renderClient(g, c)
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "new name")
	cmd.Flags().StringSliceVar(&scopes, "scopes", nil, "replacement scope set (comma-separated)")
	cmd.Flags().BoolVar(&enabled, "enabled", true, "enable (true) or disable (false) the client")
	return cmd
}

func newRotateCommand(g *cli.GlobalFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "rotate <id>",
		Short: "Rotate the client secret",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ac, err := newAdminClient(g)
			if err != nil {
				return err
			}
			resp, err := ac.RotateClientSecret(cmd.Context(), args[0])
			if err != nil {
				return fmt.Errorf("rotate client: %w", err)
			}
			return renderCreateOrRotate(g, "keys rotate", resp.Client, resp.Secret)
		},
	}
}

func newRevokeCommand(g *cli.GlobalFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "revoke <id>",
		Short: "Revoke (disable) the client without deleting it",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ac, err := newAdminClient(g)
			if err != nil {
				return err
			}
			c, err := ac.RevokeClient(cmd.Context(), args[0])
			if err != nil {
				return fmt.Errorf("revoke client: %w", err)
			}
			return renderClient(g, c)
		},
	}
}

func newDeleteCommand(g *cli.GlobalFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "delete <id>",
		Short: "Permanently delete the client",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ac, err := newAdminClient(g)
			if err != nil {
				return err
			}
			if err := ac.DeleteClient(cmd.Context(), args[0]); err != nil {
				return fmt.Errorf("delete client: %w", err)
			}
			_, _ = fmt.Fprintf(g.Stdout, "deleted %s\n", args[0])
			return nil
		},
	}
}

// newTestCommand requires the client secret as a flag because the admin
// list/show endpoints intentionally never expose it. The command exchanges
// the credentials for a token and prints the claims via /api/v1/whoami.
func newTestCommand(g *cli.GlobalFlags) *cobra.Command {
	var clientSecret string
	cmd := &cobra.Command{
		Use:   "test <id>",
		Short: "Exchange this client's credentials for a token and call /whoami",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if strings.TrimSpace(clientSecret) == "" {
				return fmt.Errorf("--secret is required (the plaintext returned by keys create or keys rotate)")
			}
			base := baseURL(g)
			c := transport.NewClient(base)
			token, err := c.IssueToken(cmd.Context(), args[0], clientSecret, nil)
			if err != nil {
				return fmt.Errorf("issue token: %w", err)
			}
			who, err := c.WhoAmI(cmd.Context(), token.AccessToken)
			if err != nil {
				return fmt.Errorf("whoami: %w", err)
			}
			return renderWhoAmI(g, who, int(token.ExpiresAt().Sub(time.Now().UTC()).Seconds()))
		},
	}
	cmd.Flags().StringVar(&clientSecret, "secret", "", "the plaintext client secret returned by keys create / rotate")
	return cmd
}

// newAdminClient builds the localhost-only admin transport client from the
// current global flags. It errors out cleanly when the user passes a port
// the mock is not running on.
func newAdminClient(g *cli.GlobalFlags) (*transport.AdminClient, error) {
	base := baseURL(g)
	return transport.NewAdminClient(base), nil
}

// baseURL is "http://localhost:<port>" where port follows the same rules as
// the mock CLI: --port flag override, falling back to D-025's 8080.
func baseURL(g *cli.GlobalFlags) string {
	port := g.Port
	if port == 0 {
		port = 8080
	}
	u := &url.URL{Scheme: "http", Host: fmt.Sprintf("localhost:%d", port)}
	return u.String()
}
