package spec

import (
	"bytes"
	"encoding/json"
	"fmt"

	"gopkg.in/yaml.v3"
)

// Spec is a parsed view of the embedded OpenAPI spec exposed for use by both
// the mock server and the CLI's `spec` subcommands. Phase 1 surfaces only the
// raw bytes and a few `info`-level fields; later phases attach a kin-openapi
// document for routing and validation.
type Spec struct {
	raw     []byte
	version string
	title   string
}

// Load parses the embedded spec and returns a Spec ready for read-only use.
// It returns an error if the embed is missing or malformed.
func Load() (*Spec, error) {
	if len(rawYAML) == 0 {
		return nil, fmt.Errorf("load spec: embedded spec is empty (run `make sync-spec` before building)")
	}
	var head struct {
		OpenAPI string `yaml:"openapi"`
		Info    struct {
			Title   string `yaml:"title"`
			Version string `yaml:"version"`
		} `yaml:"info"`
	}
	if err := yaml.Unmarshal(rawYAML, &head); err != nil {
		return nil, fmt.Errorf("load spec: parse YAML: %w", err)
	}
	if head.OpenAPI == "" {
		return nil, fmt.Errorf("load spec: missing `openapi` field at document root")
	}
	if head.Info.Version == "" {
		return nil, fmt.Errorf("load spec: missing `info.version`")
	}
	return &Spec{
		raw:     rawYAML,
		version: head.Info.Version,
		title:   head.Info.Title,
	}, nil
}

// MustLoad is Load that panics on error. Reserved for build-time-validated
// callers (e.g. CLI command initialization where a malformed embed is a bug).
func MustLoad() *Spec {
	s, err := Load()
	if err != nil {
		panic(err)
	}
	return s
}

// Version returns the `info.version` field from the embedded spec.
func (s *Spec) Version() string { return s.version }

// Title returns the `info.title` field from the embedded spec.
func (s *Spec) Title() string { return s.title }

// RawYAML returns a copy of the embedded YAML bytes.
func (s *Spec) RawYAML() []byte {
	out := make([]byte, len(s.raw))
	copy(out, s.raw)
	return out
}

// RawJSON returns the spec re-serialized as canonical JSON (indented).
func (s *Spec) RawJSON() ([]byte, error) {
	var doc map[string]any
	if err := yaml.Unmarshal(s.raw, &doc); err != nil {
		return nil, fmt.Errorf("spec to json: parse YAML: %w", err)
	}
	normalized := normalizeYAMLMaps(doc).(map[string]any)
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetIndent("", "  ")
	if err := enc.Encode(normalized); err != nil {
		return nil, fmt.Errorf("spec to json: marshal: %w", err)
	}
	return buf.Bytes(), nil
}

// normalizeYAMLMaps walks a value tree produced by yaml.v3 and converts any
// `map[any]any` (which encoding/json refuses) into
// `map[string]any` with string-coerced keys. yaml.v3 already returns
// string-keyed maps for top-level objects but nested numeric keys can slip
// through; this defends against that.
func normalizeYAMLMaps(in any) any {
	switch v := in.(type) {
	case map[any]any:
		out := make(map[string]any, len(v))
		for k, vv := range v {
			out[fmt.Sprint(k)] = normalizeYAMLMaps(vv)
		}
		return out
	case map[string]any:
		for k, vv := range v {
			v[k] = normalizeYAMLMaps(vv)
		}
		return v
	case []any:
		for i, vv := range v {
			v[i] = normalizeYAMLMaps(vv)
		}
		return v
	default:
		return v
	}
}
