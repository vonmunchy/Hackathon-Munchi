package spec_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/spec"
)

func TestLoad_EmbedsValidYAML_AndExposesVersion(t *testing.T) {
	t.Parallel()
	s, err := spec.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if s.Version() == "" {
		t.Error("version is empty")
	}
	if !strings.Contains(s.Title(), "Swipe") {
		t.Errorf("title = %q, expected to contain 'Swipe'", s.Title())
	}
}

func TestRawYAML_ContainsKnownPaths(t *testing.T) {
	t.Parallel()
	s := spec.MustLoad()
	yaml := string(s.RawYAML())
	for _, want := range []string{
		"/api/v1/balance",
		"/api/v1/payments",
		"/api/v1/payouts",
		"/api/v1/whoami",
		"/health/alive",
		"/health/ready",
	} {
		if !strings.Contains(yaml, want) {
			t.Errorf("raw spec missing path %q", want)
		}
	}
}

func TestRawJSON_ProducesParseableJSON(t *testing.T) {
	t.Parallel()
	s := spec.MustLoad()
	buf, err := s.RawJSON()
	if err != nil {
		t.Fatalf("raw json: %v", err)
	}
	var doc map[string]any
	if err := json.Unmarshal(buf, &doc); err != nil {
		t.Fatalf("decode round-tripped json: %v", err)
	}
	info, ok := doc["info"].(map[string]any)
	if !ok {
		t.Fatalf("info is not an object: %T", doc["info"])
	}
	if info["version"] != s.Version() {
		t.Errorf("info.version = %v, want %q", info["version"], s.Version())
	}
}
