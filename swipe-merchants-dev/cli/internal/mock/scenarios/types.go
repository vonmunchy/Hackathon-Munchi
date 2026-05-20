package scenarios

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

// ArgType enumerates the supported argument types in an ArgDef. Each is
// parsed by a corresponding helper on Args.
type ArgType string

// Supported arg types.
const (
	ArgString   ArgType = "string"
	ArgInt      ArgType = "int"
	ArgFloat    ArgType = "float"
	ArgDuration ArgType = "duration"
	ArgCSV      ArgType = "csv"
)

// ArgDef describes one configurable knob on a scenario.
type ArgDef struct {
	Name        string  `json:"name"`
	Type        ArgType `json:"type"`
	Default     string  `json:"default,omitempty"`
	Description string  `json:"description,omitempty"`
}

// Descriptor is the static metadata for a scenario. Each scenario file
// registers its Descriptor in init().
type Descriptor struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Args        []ArgDef `json:"args,omitempty"`
}

// Args wraps the parsed key/value pairs supplied when enabling a
// scenario. Stored verbatim in BoltDB and re-loaded at runtime.
type Args map[string]string

// String returns a stable "k=v,k=v" rendering, sorted by key.
func (a Args) String() string {
	if len(a) == 0 {
		return ""
	}
	keys := make([]string, 0, len(a))
	for k := range a {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, k+"="+a[k])
	}
	return strings.Join(parts, ",")
}

// Duration returns the named arg parsed as a time.Duration, falling back
// to fallback when the key is unset or empty.
func (a Args) Duration(key string, fallback time.Duration) (time.Duration, error) {
	raw, ok := a[key]
	if !ok || raw == "" {
		return fallback, nil
	}
	d, err := time.ParseDuration(raw)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", key, err)
	}
	return d, nil
}

// Int returns the named arg parsed as int.
func (a Args) Int(key string, fallback int) (int, error) {
	raw, ok := a[key]
	if !ok || raw == "" {
		return fallback, nil
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", key, err)
	}
	return n, nil
}

// Float returns the named arg parsed as float64.
func (a Args) Float(key string, fallback float64) (float64, error) {
	raw, ok := a[key]
	if !ok || raw == "" {
		return fallback, nil
	}
	f, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", key, err)
	}
	return f, nil
}

// CSV returns the named arg split on commas with whitespace trimmed.
// Empty entries are dropped. Missing keys yield a nil slice (not error).
func (a Args) CSV(key string) []string {
	raw, ok := a[key]
	if !ok || raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

// StringOr returns the named arg, falling back to fallback when absent.
func (a Args) StringOr(key, fallback string) string {
	raw, ok := a[key]
	if !ok || raw == "" {
		return fallback
	}
	return raw
}
