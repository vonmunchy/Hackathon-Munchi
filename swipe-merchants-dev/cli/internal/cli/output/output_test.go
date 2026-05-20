package output_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/cli/output"
)

type sample struct {
	Status   string `json:"status" yaml:"status"`
	Endpoint string `json:"endpoint" yaml:"endpoint"`
}

func TestRenderer_JSON_ProducesParseableJSON(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	r := output.New(output.FormatJSON, &buf, false)
	if err := r.Object(sample{Status: "ok", Endpoint: "http://x"}); err != nil {
		t.Fatalf("render: %v", err)
	}
	var got sample
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("parse: %v: %s", err, buf.String())
	}
	if got.Status != "ok" {
		t.Errorf("got %+v", got)
	}
}

func TestRenderer_YAML_ProducesParseableYAML(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	r := output.New(output.FormatYAML, &buf, false)
	if err := r.Object(sample{Status: "ok", Endpoint: "http://x"}); err != nil {
		t.Fatalf("render: %v", err)
	}
	var got sample
	if err := yaml.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("parse: %v: %s", err, buf.String())
	}
	if got.Status != "ok" {
		t.Errorf("got %+v", got)
	}
}

func TestRenderer_Table_IncludesKeyAndValue(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	r := output.New(output.FormatTable, &buf, false)
	err := r.KV(output.KeyValueTable{
		Title: "swipe paths",
		Rows: []output.KeyValueRow{
			{Key: "config_file", Value: "/tmp/x.yaml"},
			{Key: "mock_state_db", Value: "/tmp/state.db"},
		},
	})
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	out := buf.String()
	for _, want := range []string{"config_file", "/tmp/x.yaml", "mock_state_db", "/tmp/state.db"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q:\n%s", want, out)
		}
	}
}

func TestRenderer_Object_TableFormat_NonStructFallsBackToYAML(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	r := output.New(output.FormatTable, &buf, false)
	// A slice can be marshaled to JSON but not to map[string]any —
	// writeKeyValueTable falls back to YAML for that case.
	if err := r.Object([]string{"a", "b", "c"}); err != nil {
		t.Fatalf("render: %v", err)
	}
	out := buf.String()
	for _, want := range []string{"- a", "- b", "- c"} {
		if !strings.Contains(out, want) {
			t.Errorf("yaml fallback missing %q:\n%s", want, out)
		}
	}
}

func TestRenderer_Object_TableFormat_FlattensStruct(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	r := output.New(output.FormatTable, &buf, false)
	if err := r.Object(sample{Status: "ok", Endpoint: "http://x"}); err != nil {
		t.Fatalf("render: %v", err)
	}
	out := buf.String()
	for _, want := range []string{"status", "ok", "endpoint", "http://x"} {
		if !strings.Contains(out, want) {
			t.Errorf("table output missing %q:\n%s", want, out)
		}
	}
}

func TestRenderer_Raw_AppendsTrailingNewline(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	r := output.New(output.FormatYAML, &buf, false)
	if err := r.Raw([]byte("hello")); err != nil {
		t.Fatalf("raw: %v", err)
	}
	if got := buf.String(); got != "hello\n" {
		t.Errorf("got %q, want %q", got, "hello\n")
	}
	buf.Reset()
	if err := r.Raw([]byte("hello\n")); err != nil {
		t.Fatalf("raw: %v", err)
	}
	if got := buf.String(); got != "hello\n" {
		t.Errorf("got %q, want %q (no double newline)", got, "hello\n")
	}
}

func TestParseFormat_AcceptsKnownValues(t *testing.T) {
	t.Parallel()
	cases := map[string]output.Format{
		"":      output.FormatTable,
		"table": output.FormatTable,
		"json":  output.FormatJSON,
		"yaml":  output.FormatYAML,
		"YAML":  output.FormatYAML,
		"yml":   output.FormatYAML,
	}
	for in, want := range cases {
		got, err := output.ParseFormat(in)
		if err != nil {
			t.Errorf("ParseFormat(%q): unexpected error %v", in, err)
			continue
		}
		if got != want {
			t.Errorf("ParseFormat(%q) = %q, want %q", in, got, want)
		}
	}
	if _, err := output.ParseFormat("xml"); err == nil {
		t.Errorf("ParseFormat(xml): expected error, got nil")
	}
}
