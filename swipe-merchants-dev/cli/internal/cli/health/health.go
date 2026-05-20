package health

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/cli"
	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/cli/output"
	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/transport"
)

// NewCommand builds the `swipe health` parent command.
func NewCommand(g *cli.GlobalFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "health",
		Short: "Probe the running mock's health endpoints",
	}
	cmd.AddCommand(newAliveCommand(g))
	cmd.AddCommand(newReadyCommand(g))
	return cmd
}

func newAliveCommand(g *cli.GlobalFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "alive",
		Short: "Call GET /health/alive",
		RunE: func(_ *cobra.Command, _ []string) error {
			return runAlive(g)
		},
	}
}

func newReadyCommand(g *cli.GlobalFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "ready",
		Short: "Call GET /health/ready (exits non-zero if the mock is not ready)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runReady(cmd, g)
		},
	}
}

// AliveResult is the renderable result of `health alive`.
type AliveResult struct {
	Status   string `json:"status" yaml:"status"`
	Endpoint string `json:"endpoint" yaml:"endpoint"`
}

// ReadyResult is the renderable result of `health ready`. `Healthy` is a
// scalar form of the body's status (true/false) suitable for shell
// scripting; `Body` is the full parsed response.
type ReadyResult struct {
	HTTPStatus int    `json:"http_status" yaml:"http_status"`
	Healthy    bool   `json:"healthy" yaml:"healthy"`
	Endpoint   string `json:"endpoint" yaml:"endpoint"`
	Body       any    `json:"body" yaml:"body"`
}

func runAlive(g *cli.GlobalFlags) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	client := transport.NewClient(baseURL(g))
	body, err := client.HealthAlive(ctx)
	if err != nil {
		return fmt.Errorf("call /health/alive: %w", err)
	}
	status, _ := body["status"].(string)
	r, err := g.Renderer()
	if err != nil {
		return err
	}
	return r.KV(output.KeyValueTable{
		Title: "/health/alive",
		Rows: []output.KeyValueRow{
			{Key: "status", Value: status},
			{Key: "endpoint", Value: client.BaseURL + "/health/alive"},
		},
	})
}

func runReady(cmd *cobra.Command, g *cli.GlobalFlags) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	client := transport.NewClient(baseURL(g))
	status, body, err := client.HealthReady(ctx)
	if err != nil {
		return fmt.Errorf("call /health/ready: %w", err)
	}
	healthy := status >= 200 && status < 300
	if statusStr, ok := body["status"].(string); ok && statusStr != "ok" {
		healthy = false
	}
	r, err := g.Renderer()
	if err != nil {
		return err
	}
	if err := r.Object(ReadyResult{
		HTTPStatus: status,
		Healthy:    healthy,
		Endpoint:   client.BaseURL + "/health/ready",
		Body:       body,
	}); err != nil {
		return err
	}
	if !healthy {
		// Cobra would normally print usage on a returned error; we want a
		// clean non-zero exit instead. SilenceUsage is set on root, so
		// returning the sentinel is fine.
		_ = cmd
		return fmt.Errorf("not ready (http %d)", status)
	}
	return nil
}

// baseURL resolves the mock's address. Phase 1 hardcodes localhost; the
// --port flag overrides the default 8080 (DECISIONS D-001 / D-025).
func baseURL(g *cli.GlobalFlags) string {
	port := g.Port
	if port == 0 {
		port = 8080
	}
	return fmt.Sprintf("http://localhost:%d", port)
}
