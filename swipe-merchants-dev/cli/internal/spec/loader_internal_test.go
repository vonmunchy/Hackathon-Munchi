package spec

import (
	"strings"
	"testing"
)

// White-box tests for the loader's error and helper paths.

func TestNormalizeYAMLMaps_HandlesNumericKeys(t *testing.T) {
	t.Parallel()
	in := map[any]any{
		1:     "one",
		"two": 2,
		"nested": map[any]any{
			3: "three",
		},
	}
	got, ok := normalizeYAMLMaps(in).(map[string]any)
	if !ok {
		t.Fatalf("not coerced to map[string]any: %T", got)
	}
	if got["1"] != "one" {
		t.Errorf("1 -> %v, want one", got["1"])
	}
	nested, ok := got["nested"].(map[string]any)
	if !ok {
		t.Fatalf("nested not coerced: %T", got["nested"])
	}
	if nested["3"] != "three" {
		t.Errorf("nested[3] = %v", nested["3"])
	}
}

func TestNormalizeYAMLMaps_HandlesSlices(t *testing.T) {
	t.Parallel()
	in := []any{
		map[any]any{1: "one"},
		"plain",
	}
	got := normalizeYAMLMaps(in).([]any)
	first := got[0].(map[string]any)
	if first["1"] != "one" {
		t.Errorf("first[1] = %v", first["1"])
	}
	if got[1] != "plain" {
		t.Errorf("got[1] = %v", got[1])
	}
}

func TestNormalizeYAMLMaps_PrimitivePassthrough(t *testing.T) {
	t.Parallel()
	for _, in := range []any{42, "hello", 3.14, true} {
		if got := normalizeYAMLMaps(in); got != in {
			t.Errorf("primitive passthrough: %v -> %v", in, got)
		}
	}
}

// loadFromBytes is a test-only mirror of Load that swaps the package's
// rawYAML for the duration of a test. Tests using it are serialized
// (no t.Parallel) because they mutate package-level state.
func loadFromBytes(t *testing.T, src []byte) (*Spec, error) {
	t.Helper()
	saved := rawYAML
	rawYAML = src
	t.Cleanup(func() { rawYAML = saved })
	return Load()
}

// All Load-error tests below mutate the package-level `rawYAML` and so
// must run serially. They share state with each other and with
// TestMustLoad_BadEmbed_Panics; do not add t.Parallel here.

func TestLoad_EmptyBytes_Errors(t *testing.T) {
	_, err := loadFromBytes(t, []byte{})
	if err == nil || !strings.Contains(err.Error(), "embedded spec is empty") {
		t.Errorf("err = %v, want 'embedded spec is empty'", err)
	}
}

func TestLoad_MalformedYAML_Errors(t *testing.T) {
	// Unclosed quote — guaranteed YAML parse error.
	_, err := loadFromBytes(t, []byte("foo: 'unclosed\n"))
	if err == nil || !strings.Contains(err.Error(), "parse YAML") {
		t.Errorf("err = %v, want 'parse YAML'", err)
	}
}

func TestLoad_MissingOpenAPIField_Errors(t *testing.T) {
	_, err := loadFromBytes(t, []byte("info:\n  title: foo\n  version: 1.0\n"))
	if err == nil || !strings.Contains(err.Error(), "missing `openapi`") {
		t.Errorf("err = %v, want 'missing openapi'", err)
	}
}

func TestLoad_MissingInfoVersion_Errors(t *testing.T) {
	_, err := loadFromBytes(t, []byte("openapi: 3.0.0\ninfo:\n  title: foo\n"))
	if err == nil || !strings.Contains(err.Error(), "missing `info.version`") {
		t.Errorf("err = %v, want 'missing info.version'", err)
	}
}

func TestRawJSON_BadYAML_ReturnsParseError(t *testing.T) {
	t.Parallel()
	// A list at the root cannot unmarshal into map[string]any.
	s := &Spec{raw: []byte("- item1\n- item2\n")}
	if _, err := s.RawJSON(); err == nil {
		t.Errorf("expected parse error; got nil")
	}
}

func TestMustLoad_BadEmbed_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic; got none")
		}
	}()
	saved := rawYAML
	rawYAML = nil
	defer func() { rawYAML = saved }()
	_ = MustLoad()
}
