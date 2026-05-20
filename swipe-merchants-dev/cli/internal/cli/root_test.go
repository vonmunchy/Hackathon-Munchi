package cli_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/cli"
)

func TestDefaults_ReturnsTableFormatAndPort8080(t *testing.T) {
	t.Parallel()
	g := cli.Defaults()
	if g.Output != "table" {
		t.Errorf("Output = %q, want table", g.Output)
	}
	if g.Port != 8080 {
		t.Errorf("Port = %d, want 8080", g.Port)
	}
	if g.Stdout == nil || g.Stderr == nil {
		t.Errorf("Stdout/Stderr not set: %+v", g)
	}
}

func TestErrorf_WritesErrorPrefix(t *testing.T) {
	t.Parallel()
	stderr := &bytes.Buffer{}
	g := &cli.GlobalFlags{Stderr: stderr}
	g.Errorf("something %s", "broke")
	got := stderr.String()
	if !strings.HasPrefix(got, "Error: ") {
		t.Errorf("missing Error: prefix: %q", got)
	}
	if !strings.Contains(got, "something broke") {
		t.Errorf("missing message: %q", got)
	}
}

func TestRenderer_InvalidFormat_Errors(t *testing.T) {
	t.Parallel()
	g := &cli.GlobalFlags{Output: "garbage", Stdout: &bytes.Buffer{}, Stderr: &bytes.Buffer{}}
	if _, err := g.Renderer(); err == nil {
		t.Errorf("expected error for invalid format")
	}
}
