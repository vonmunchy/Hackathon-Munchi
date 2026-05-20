package output

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"gopkg.in/yaml.v3"
)

// Format is the output format selected by --output.
type Format string

// Supported formats.
const (
	FormatTable Format = "table"
	FormatJSON  Format = "json"
	FormatYAML  Format = "yaml"
)

// ParseFormat coerces a string from --output to a Format. Unknown values
// produce a clear error rather than silently falling back.
func ParseFormat(s string) (Format, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "", "table":
		return FormatTable, nil
	case "json":
		return FormatJSON, nil
	case "yaml", "yml":
		return FormatYAML, nil
	default:
		return "", fmt.Errorf("invalid --output value %q (want one of: table, json, yaml)", s)
	}
}

// Renderer renders structured values to an io.Writer in a chosen Format.
// Color is a hint that the receiving terminal supports ANSI escapes; it is
// honored only by the table renderer (JSON/YAML are always plain).
type Renderer struct {
	Format Format
	Writer io.Writer
	Color  bool
}

// New builds a Renderer. The writer is typically os.Stdout.
func New(format Format, w io.Writer, color bool) Renderer {
	return Renderer{Format: format, Writer: w, Color: color}
}

// Object renders v as a single object. Each format chooses its own
// representation:
//   - table: a two-column key/value table built from JSON-tagged fields
//   - json:  json.Marshal with two-space indent
//   - yaml:  yaml.v3 marshal
func (r Renderer) Object(v any) error {
	switch r.Format {
	case FormatJSON:
		return r.writeJSON(v)
	case FormatYAML:
		return r.writeYAML(v)
	default:
		return r.writeKeyValueTable(v)
	}
}

// KeyValueTable renders an ordered list of key/value rows. Useful when the
// caller wants table output not derived from a struct (e.g. paths).
type KeyValueTable struct {
	Title string
	Rows  []KeyValueRow
}

// KeyValueRow is a single row in a KeyValueTable.
type KeyValueRow struct {
	Key   string
	Value any
}

// KV renders a KeyValueTable with format-appropriate output. Formats
// other than table get a flat object representation (key→value map).
func (r Renderer) KV(kv KeyValueTable) error {
	switch r.Format {
	case FormatJSON, FormatYAML:
		obj := make(map[string]any, len(kv.Rows))
		for _, row := range kv.Rows {
			obj[row.Key] = row.Value
		}
		if r.Format == FormatJSON {
			return r.writeJSON(obj)
		}
		return r.writeYAML(obj)
	default:
		t := table.NewWriter()
		t.SetOutputMirror(r.Writer)
		if kv.Title != "" {
			t.SetTitle(kv.Title)
		}
		t.AppendHeader(table.Row{"key", "value"})
		for _, row := range kv.Rows {
			t.AppendRow(table.Row{row.Key, row.Value})
		}
		t.SetStyle(tableStyle(r.Color))
		t.Render()
		return nil
	}
}

// Raw writes pre-formatted bytes followed by a newline if the buffer does
// not already end with one. Used by `swipe spec show`.
func (r Renderer) Raw(b []byte) error {
	if _, err := r.Writer.Write(b); err != nil {
		return fmt.Errorf("write raw: %w", err)
	}
	if len(b) == 0 || b[len(b)-1] != '\n' {
		_, _ = r.Writer.Write([]byte("\n"))
	}
	return nil
}

func (r Renderer) writeJSON(v any) error {
	enc := json.NewEncoder(r.Writer)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		return fmt.Errorf("encode json: %w", err)
	}
	return nil
}

func (r Renderer) writeYAML(v any) error {
	enc := yaml.NewEncoder(r.Writer)
	enc.SetIndent(2)
	defer func() { _ = enc.Close() }()
	if err := enc.Encode(v); err != nil {
		return fmt.Errorf("encode yaml: %w", err)
	}
	return nil
}

// writeKeyValueTable produces a 2-column table from struct fields using
// JSON tags as keys. For non-struct values it falls back to YAML output
// (better than a single-cell table for slice-of-struct).
func (r Renderer) writeKeyValueTable(v any) error {
	buf, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("marshal for table: %w", err)
	}
	var asMap map[string]any
	if err := json.Unmarshal(buf, &asMap); err != nil {
		return r.writeYAML(v)
	}
	t := table.NewWriter()
	t.SetOutputMirror(r.Writer)
	t.AppendHeader(table.Row{"key", "value"})
	for _, k := range sortedKeys(asMap) {
		t.AppendRow(table.Row{k, fmt.Sprintf("%v", asMap[k])})
	}
	t.SetStyle(tableStyle(r.Color))
	t.Render()
	return nil
}

func sortedKeys(m map[string]any) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	// Insertion-stable sort: lexicographic. Good enough for status / config.
	for i := 1; i < len(out); i++ {
		for j := i; j > 0 && out[j-1] > out[j]; j-- {
			out[j-1], out[j] = out[j], out[j-1]
		}
	}
	return out
}

func tableStyle(color bool) table.Style {
	st := table.StyleLight
	if !color {
		st.Color = table.ColorOptions{}
	}
	return st
}
