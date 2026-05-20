package keys

import (
	"fmt"
	"strings"
	"time"

	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/cli"
	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/cli/output"
	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/admin"
	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/transport"
)

// renderClient writes a single client view in the requested output format.
// Table output omits the bcrypt hash (which is never returned by the mock
// in any case).
func renderClient(g *cli.GlobalFlags, c admin.ClientView) error {
	r, err := g.Renderer()
	if err != nil {
		return err
	}
	if r.Format != output.FormatTable {
		return r.Object(c)
	}
	rows := clientRows(c)
	return r.KV(output.KeyValueTable{Title: "client", Rows: rows})
}

// renderClientList writes a list of client views. Table output prints one
// KV block per item separated by blank lines to keep the format consistent
// with single-item output.
func renderClientList(g *cli.GlobalFlags, items []admin.ClientView) error {
	r, err := g.Renderer()
	if err != nil {
		return err
	}
	if r.Format != output.FormatTable {
		return r.Object(items)
	}
	if len(items) == 0 {
		_, _ = fmt.Fprintln(g.Stdout, "no clients")
		return nil
	}
	for i, c := range items {
		if i > 0 {
			_, _ = fmt.Fprintln(g.Stdout)
		}
		if err := r.KV(output.KeyValueTable{Title: c.ID, Rows: clientRows(c)}); err != nil {
			return err
		}
	}
	return nil
}

// renderCreateOrRotate prints a client view alongside its plaintext secret
// with a one-time-save warning per D-023. For json/yaml output it embeds
// the secret in the structured response.
func renderCreateOrRotate(g *cli.GlobalFlags, title string, c admin.ClientView, secret string) error {
	r, err := g.Renderer()
	if err != nil {
		return err
	}
	type withSecret struct {
		admin.ClientView
		ClientSecret string `json:"client_secret" yaml:"client_secret"`
	}
	if r.Format != output.FormatTable {
		return r.Object(withSecret{ClientView: c, ClientSecret: secret})
	}
	rows := clientRows(c)
	rows = append(rows, output.KeyValueRow{Key: "client_secret", Value: secret})
	if err := r.KV(output.KeyValueTable{Title: title, Rows: rows}); err != nil {
		return err
	}
	_, _ = fmt.Fprintln(g.Stderr, "save the client_secret now — it cannot be recovered later")
	return nil
}

// renderWhoAmI writes the result of `keys test` / `auth whoami` in the
// chosen format.
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

func clientRows(c admin.ClientView) []output.KeyValueRow {
	return []output.KeyValueRow{
		{Key: "id", Value: c.ID},
		{Key: "merchant_id", Value: c.MerchantID},
		{Key: "name", Value: c.Name},
		{Key: "scopes", Value: strings.Join(c.Scopes, " ")},
		{Key: "enabled", Value: c.Enabled},
		{Key: "created_at", Value: c.CreatedAt.Format(time.RFC3339)},
		{Key: "updated_at", Value: c.UpdatedAt.Format(time.RFC3339)},
	}
}
