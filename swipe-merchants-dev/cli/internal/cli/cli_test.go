package cli_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/cli"
	configcmd "github.com/BML-Digital/swipe-merchants-dev/cli/internal/cli/config"
	speccmd "github.com/BML-Digital/swipe-merchants-dev/cli/internal/cli/spec"
	versioncmd "github.com/BML-Digital/swipe-merchants-dev/cli/internal/cli/version"
)

// runCLI builds the root command tree (excluding `mock`/`health`, which
// would need an external server to be meaningful) and runs it with the
// given args, returning stdout, stderr, and the run error.
func runCLI(t *testing.T, args ...string) (string, string, error) {
	t.Helper()
	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	g := &cli.GlobalFlags{Stdout: stdout, Stderr: stderr, Output: "table", Port: 8080}
	root := cli.NewRoot(g, func(r *cobra.Command, gf *cli.GlobalFlags) {
		r.AddCommand(versioncmd.NewCommand(gf))
		r.AddCommand(configcmd.NewCommand(gf))
		r.AddCommand(speccmd.NewCommand(gf))
	})
	root.SetArgs(args)
	root.SetOut(stdout)
	root.SetErr(stderr)
	err := root.Execute()
	return stdout.String(), stderr.String(), err
}

func TestVersion_JSON_ContainsAllExpectedFields(t *testing.T) {
	t.Parallel()
	stdout, _, err := runCLI(t, "version", "--output", "json")
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	var got map[string]any
	if err := json.Unmarshal([]byte(stdout), &got); err != nil {
		t.Fatalf("parse: %v: %s", err, stdout)
	}
	for _, key := range []string{"version", "commit", "build_date", "spec_version"} {
		if _, ok := got[key]; !ok {
			t.Errorf("expected %q in version output: %v", key, got)
		}
	}
	if got["spec_version"] == "0.0.0" {
		t.Errorf("spec_version was %q; expected to be read from embedded spec, not the ldflag default", got["spec_version"])
	}
}

func TestVersion_YAML_ProducesParseableYAML(t *testing.T) {
	t.Parallel()
	stdout, _, err := runCLI(t, "version", "--output", "yaml")
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	var got map[string]any
	if err := yaml.Unmarshal([]byte(stdout), &got); err != nil {
		t.Fatalf("parse: %v: %s", err, stdout)
	}
	if got["spec_version"] == nil {
		t.Errorf("spec_version missing: %v", got)
	}
}

func TestVersion_Table_PrintsKeyValueRows(t *testing.T) {
	t.Parallel()
	stdout, _, err := runCLI(t, "version")
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	for _, want := range []string{"version", "commit", "build_date", "spec_version"} {
		if !strings.Contains(stdout, want) {
			t.Errorf("table output missing %q:\n%s", want, stdout)
		}
	}
}

func TestSpecVersion_JSON_ReportsEmbeddedVersion(t *testing.T) {
	t.Parallel()
	stdout, _, err := runCLI(t, "spec", "version", "--output", "json")
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	var got map[string]any
	if err := json.Unmarshal([]byte(stdout), &got); err != nil {
		t.Fatalf("parse: %v: %s", err, stdout)
	}
	if got["spec_version"] == nil || got["spec_version"] == "" {
		t.Errorf("spec_version is empty: %v", got)
	}
	if got["spec_title"] == nil || !strings.Contains(got["spec_title"].(string), "Swipe") {
		t.Errorf("spec_title = %v, expected to mention 'Swipe'", got["spec_title"])
	}
}

func TestSpecShow_DefaultsToYAML(t *testing.T) {
	t.Parallel()
	stdout, _, err := runCLI(t, "spec", "show")
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if !strings.Contains(stdout, "openapi:") {
		t.Errorf("YAML output missing 'openapi:' header:\n%s", stdout[:min(200, len(stdout))])
	}
	if !strings.Contains(stdout, "/api/v1/payments") {
		t.Errorf("YAML output missing /api/v1/payments path")
	}
}

func TestSpecShow_FormatJSON_ProducesParseableJSON(t *testing.T) {
	t.Parallel()
	stdout, _, err := runCLI(t, "spec", "show", "--format", "json")
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	var got map[string]any
	if err := json.Unmarshal([]byte(stdout), &got); err != nil {
		t.Fatalf("parse: %v", err)
	}
	if got["openapi"] == nil {
		t.Errorf("missing openapi field: %v", got)
	}
}

func TestSpecShow_FormatTable_ReturnsError(t *testing.T) {
	t.Parallel()
	_, _, err := runCLI(t, "spec", "show", "--format", "table")
	if err == nil {
		t.Errorf("expected error for --format table; got nil")
	}
}

func TestConfigPath_JSON_ContainsAllPathKeys(t *testing.T) {
	t.Parallel()
	stdout, _, err := runCLI(t, "config", "path", "--output", "json")
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	var got map[string]any
	if err := json.Unmarshal([]byte(stdout), &got); err != nil {
		t.Fatalf("parse: %v: %s", err, stdout)
	}
	for _, key := range []string{
		"root", "config_file", "mock_dir", "mock_state_db",
		"mock_keys_dir", "mock_pidfile", "mock_address_file", "token_cache",
	} {
		if _, ok := got[key]; !ok {
			t.Errorf("expected %q in config path output: %v", key, got)
		}
	}
}

func TestParseFormat_InvalidValue_ReturnsError(t *testing.T) {
	t.Parallel()
	_, _, err := runCLI(t, "version", "--output", "xml")
	if err == nil {
		t.Errorf("expected error for invalid --output value; got nil")
	}
}
