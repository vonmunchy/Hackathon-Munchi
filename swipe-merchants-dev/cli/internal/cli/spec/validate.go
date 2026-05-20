package spec

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers/gorillamux"
	"github.com/spf13/cobra"

	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/cli"
	specpkg "github.com/BML-Digital/swipe-merchants-dev/cli/internal/spec"
)

// newValidateCommand wires `swipe spec validate`. Offline validation:
// users hand a payload + operationId and we check it against the
// embedded spec without going over the wire.
func newValidateCommand(g *cli.GlobalFlags) *cobra.Command {
	var (
		operationID string
		payloadPath string
	)
	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate a JSON payload against an operationId in the embedded spec",
		Long: `Validate a request body offline against the embedded OpenAPI spec.

Pipe a JSON payload on stdin (default) or pass --file. The --operation
flag is the operationId from the spec (e.g. createPayment).

  cat new-payment.json | swipe spec validate --operation createPayment
  swipe spec validate --operation createPayment --file pay.json`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if strings.TrimSpace(operationID) == "" {
				return fmt.Errorf("--operation is required")
			}
			body, err := readBody(cmd.InOrStdin(), payloadPath)
			if err != nil {
				return err
			}
			return runValidate(cmd.Context(), g, operationID, body)
		},
	}
	cmd.Flags().StringVar(&operationID, "operation", "", "operationId from the spec (required)")
	cmd.Flags().StringVarP(&payloadPath, "file", "f", "-", "path to the JSON payload ('-' for stdin)")
	return cmd
}

func runValidate(ctx context.Context, g *cli.GlobalFlags, operationID string, body []byte) error {
	loaded, err := specpkg.Load()
	if err != nil {
		return fmt.Errorf("load embedded spec: %w", err)
	}
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData(loaded.RawYAML())
	if err != nil {
		return fmt.Errorf("parse spec: %w", err)
	}
	doc.Servers = openapi3.Servers{{URL: "/"}}
	if err := doc.Validate(loader.Context); err != nil {
		return fmt.Errorf("spec validation: %w", err)
	}
	method, path, op := findOperation(doc, operationID)
	if op == nil {
		return fmt.Errorf("operation %q not found in spec", operationID)
	}
	if op.RequestBody == nil || op.RequestBody.Value == nil {
		return fmt.Errorf("operation %q has no request body to validate against", operationID)
	}
	router, err := gorillamux.NewRouter(doc)
	if err != nil {
		return fmt.Errorf("build router: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, method, path, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	route, pathParams, err := router.FindRoute(req)
	if err != nil {
		return fmt.Errorf("route operation: %w", err)
	}
	input := &openapi3filter.RequestValidationInput{
		Request:    req,
		PathParams: pathParams,
		Route:      route,
		Options: &openapi3filter.Options{
			AuthenticationFunc: openapi3filter.NoopAuthenticationFunc,
		},
	}
	if err := openapi3filter.ValidateRequest(ctx, input); err != nil {
		_, _ = fmt.Fprintf(g.Stdout, "invalid:\n  %v\n", err)
		return err
	}
	_, _ = fmt.Fprintln(g.Stdout, "valid")
	return nil
}

// findOperation walks the spec for the named operationId, returning the
// (method, path, op) triple or (..., nil) when not found.
func findOperation(doc *openapi3.T, operationID string) (string, string, *openapi3.Operation) {
	for path, item := range doc.Paths.Map() {
		for method, op := range item.Operations() {
			if op.OperationID == operationID {
				return method, path, op
			}
		}
	}
	return "", "", nil
}

func readBody(stdin io.Reader, file string) ([]byte, error) {
	if file == "-" {
		return io.ReadAll(stdin)
	}
	return os.ReadFile(file) // #nosec G304 -- user-supplied path
}
