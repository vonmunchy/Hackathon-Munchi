package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestRun_Version_ReturnsZero(t *testing.T) {
	t.Parallel()
	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	code := Run([]string{"version", "--output", "json"}, stdout, stderr)
	if code != 0 {
		t.Errorf("code = %d, want 0; stderr=%s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "spec_version") {
		t.Errorf("stdout missing spec_version: %s", stdout.String())
	}
}

func TestRun_UnknownCommand_ReturnsOne(t *testing.T) {
	t.Parallel()
	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	code := Run([]string{"this-command-does-not-exist"}, stdout, stderr)
	if code != 1 {
		t.Errorf("code = %d, want 1", code)
	}
	if !strings.Contains(stderr.String(), "Error:") {
		t.Errorf("stderr missing Error prefix: %s", stderr.String())
	}
}

func TestRun_HelpFlag_ReturnsZero(t *testing.T) {
	t.Parallel()
	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	code := Run([]string{"--help"}, stdout, stderr)
	if code != 0 {
		t.Errorf("code = %d, want 0; stderr=%s", code, stderr.String())
	}
}

func TestAttachSubcommands_RegistersAllPhase1Commands(t *testing.T) {
	t.Parallel()
	stdout, _ := &bytes.Buffer{}, &bytes.Buffer{}
	code := Run([]string{"--help"}, stdout, &bytes.Buffer{})
	if code != 0 {
		t.Fatal("help failed")
	}
	out := stdout.String()
	for _, name := range []string{"version", "config", "spec", "health", "mock"} {
		if !strings.Contains(out, name) {
			t.Errorf("--help output missing subcommand %q:\n%s", name, out)
		}
	}
}
