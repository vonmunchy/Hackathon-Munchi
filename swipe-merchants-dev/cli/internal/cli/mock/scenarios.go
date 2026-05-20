package mock

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/cli"
	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/cli/output"
	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/admin"
	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/transport"
)

// newScenariosCommand builds the `swipe mock scenarios` subtree. It is
// attached by NewCommand above so the surface lives under `swipe mock`.
func newScenariosCommand(g *cli.GlobalFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "scenarios",
		Short: "Enable, disable, and inspect mock scenarios",
	}
	cmd.AddCommand(newScenariosListCommand(g))
	cmd.AddCommand(newScenariosShowCommand(g))
	cmd.AddCommand(newScenariosEnableCommand(g))
	cmd.AddCommand(newScenariosDisableCommand(g))
	return cmd
}

func newScenariosListCommand(g *cli.GlobalFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List every scenario in the catalog with enabled state",
		RunE: func(cmd *cobra.Command, _ []string) error {
			var out []admin.ScenarioView
			if err := adminGET(cmd.Context(), g, "/_admin/scenarios", &out); err != nil {
				return fmt.Errorf("list scenarios: %w", err)
			}
			return renderScenarios(g, out)
		},
	}
}

func newScenariosShowCommand(g *cli.GlobalFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "show <name>",
		Short: "Show one scenario with current args",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var out admin.ScenarioView
			if err := adminGET(cmd.Context(), g, "/_admin/scenarios/"+args[0], &out); err != nil {
				return fmt.Errorf("show scenario: %w", err)
			}
			return renderScenario(g, out)
		},
	}
}

func newScenariosEnableCommand(g *cli.GlobalFlags) *cobra.Command {
	var argPairs []string
	cmd := &cobra.Command{
		Use:   "enable <name>",
		Short: "Enable a scenario; pass --args key=value (repeatable)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			parsed, err := parseArgPairs(argPairs)
			if err != nil {
				return err
			}
			body := map[string]any{"args": parsed}
			var state map[string]any
			if err := adminPOST(cmd.Context(), g, "/_admin/scenarios/"+args[0]+"/enable", body, &state); err != nil {
				return fmt.Errorf("enable scenario: %w", err)
			}
			out, _ := json.MarshalIndent(state, "", "  ")
			_, _ = fmt.Fprintln(g.Stdout, string(out))
			return nil
		},
	}
	cmd.Flags().StringSliceVar(&argPairs, "args", nil, "scenario args as key=value pairs (e.g. --args rate=0.5,codes=503)")
	return cmd
}

func newScenariosDisableCommand(g *cli.GlobalFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "disable <name>",
		Short: "Disable a scenario",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := adminPOST(cmd.Context(), g, "/_admin/scenarios/"+args[0]+"/disable", nil, nil); err != nil {
				return fmt.Errorf("disable scenario: %w", err)
			}
			_, _ = fmt.Fprintf(g.Stdout, "disabled %s\n", args[0])
			return nil
		},
	}
}

// parseArgPairs converts "key=value" strings into a map. Duplicate keys
// keep the last one. Empty values are allowed (some scenarios accept "").
func parseArgPairs(pairs []string) (map[string]string, error) {
	if len(pairs) == 0 {
		return nil, nil
	}
	out := make(map[string]string, len(pairs))
	for _, raw := range pairs {
		// support both --args k=v --args k=v and --args k=v,k=v.
		for _, p := range strings.Split(raw, ",") {
			p = strings.TrimSpace(p)
			if p == "" {
				continue
			}
			idx := strings.IndexByte(p, '=')
			if idx <= 0 {
				return nil, fmt.Errorf("malformed --args value %q (expected key=value)", p)
			}
			out[p[:idx]] = p[idx+1:]
		}
	}
	return out, nil
}

func adminGET(ctx context.Context, g *cli.GlobalFlags, path string, out any) error {
	base := scenarioBaseURL(g)
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, base+path, nil)
	if err != nil {
		return err
	}
	resp, err := transport.NewAdminClient(base).HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	if out == nil {
		return nil
	}
	return json.Unmarshal(body, out)
}

func adminPOST(ctx context.Context, g *cli.GlobalFlags, path string, body any, out any) error {
	base := scenarioBaseURL(g)
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	var rdr io.Reader
	if body != nil {
		buf, err := json.Marshal(body)
		if err != nil {
			return err
		}
		rdr = strings.NewReader(string(buf))
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, base+path, rdr)
	if err != nil {
		return err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := transport.NewAdminClient(base).HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("status %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}
	if out == nil {
		return nil
	}
	if len(respBody) == 0 {
		return nil
	}
	return json.Unmarshal(respBody, out)
}

func scenarioBaseURL(g *cli.GlobalFlags) string {
	port := g.Port
	if port == 0 {
		port = 8080
	}
	return fmt.Sprintf("http://localhost:%d", port)
}

func renderScenarios(g *cli.GlobalFlags, views []admin.ScenarioView) error {
	r, err := g.Renderer()
	if err != nil {
		return err
	}
	if r.Format != output.FormatTable {
		return r.Object(views)
	}
	if len(views) == 0 {
		_, _ = fmt.Fprintln(g.Stdout, "no scenarios registered")
		return nil
	}
	for i, v := range views {
		if i > 0 {
			_, _ = fmt.Fprintln(g.Stdout)
		}
		rows := []output.KeyValueRow{
			{Key: "name", Value: v.Name},
			{Key: "description", Value: v.Description},
			{Key: "enabled", Value: v.Enabled},
		}
		if len(v.CurrentArgs) > 0 {
			rows = append(rows, output.KeyValueRow{Key: "current_args", Value: renderArgs(v.CurrentArgs)})
		}
		if err := r.KV(output.KeyValueTable{Title: v.Name, Rows: rows}); err != nil {
			return err
		}
	}
	return nil
}

func renderScenario(g *cli.GlobalFlags, view admin.ScenarioView) error {
	return renderScenarios(g, []admin.ScenarioView{view})
}

func renderArgs(args map[string]string) string {
	if len(args) == 0 {
		return ""
	}
	parts := make([]string, 0, len(args))
	for k, v := range args {
		parts = append(parts, k+"="+v)
	}
	return strings.Join(parts, ",")
}
