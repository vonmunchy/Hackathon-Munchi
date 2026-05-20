package spec

import (
	"bytes"
	"testing"

	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/cli"
	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/cli/output"
)

func TestResolveSpecFormat_LocalWins(t *testing.T) {
	t.Parallel()
	got, err := resolveSpecFormat("json", "yaml")
	if err != nil || got != output.FormatJSON {
		t.Errorf("local json + global yaml = %v, %v; want json", got, err)
	}
}

func TestResolveSpecFormat_LocalTable_Errors(t *testing.T) {
	t.Parallel()
	if _, err := resolveSpecFormat("table", "yaml"); err == nil {
		t.Errorf("expected error for --format table")
	}
}

func TestResolveSpecFormat_LocalInvalid_Errors(t *testing.T) {
	t.Parallel()
	if _, err := resolveSpecFormat("xml", ""); err == nil {
		t.Errorf("expected error for invalid --format")
	}
}

func TestResolveSpecFormat_GlobalJSON_Honored(t *testing.T) {
	t.Parallel()
	got, _ := resolveSpecFormat("", "json")
	if got != output.FormatJSON {
		t.Errorf("got %v, want json", got)
	}
}

func TestResolveSpecFormat_GlobalYAML_Honored(t *testing.T) {
	t.Parallel()
	got, _ := resolveSpecFormat("", "yaml")
	if got != output.FormatYAML {
		t.Errorf("got %v, want yaml", got)
	}
}

func TestResolveSpecFormat_GlobalTable_FallsBackToYAML(t *testing.T) {
	t.Parallel()
	got, _ := resolveSpecFormat("", "table")
	if got != output.FormatYAML {
		t.Errorf("got %v, want yaml fallback", got)
	}
}

func TestResolveSpecFormat_GlobalEmpty_FallsBackToYAML(t *testing.T) {
	t.Parallel()
	got, _ := resolveSpecFormat("", "")
	if got != output.FormatYAML {
		t.Errorf("got %v, want yaml fallback", got)
	}
}

func TestResolveSpecFormat_GlobalUnknown_FallsBackToYAML(t *testing.T) {
	t.Parallel()
	got, _ := resolveSpecFormat("", "garbage")
	if got != output.FormatYAML {
		t.Errorf("got %v, want yaml fallback", got)
	}
}

func TestRunShow_InvalidGlobalOutput_RendererErrors(t *testing.T) {
	t.Parallel()
	g := &cli.GlobalFlags{
		Stdout: &bytes.Buffer{},
		Stderr: &bytes.Buffer{},
		Output: "garbage",
	}
	if err := runShow(g, ""); err == nil {
		t.Errorf("expected renderer error from invalid global output")
	}
}

func TestRunVersion_InvalidGlobalOutput_RendererErrors(t *testing.T) {
	t.Parallel()
	g := &cli.GlobalFlags{
		Stdout: &bytes.Buffer{},
		Stderr: &bytes.Buffer{},
		Output: "garbage",
	}
	if err := runVersion(g); err == nil {
		t.Errorf("expected renderer error from invalid global output")
	}
}
