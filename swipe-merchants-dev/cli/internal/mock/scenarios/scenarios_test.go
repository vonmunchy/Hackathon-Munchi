package scenarios_test

import (
	"context"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/scenarios"
	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/store"
)

func TestRegistry_HasAllTenScenarios(t *testing.T) {
	want := []string{
		"client_revoked_after",
		"insufficient_funds",
		"latency_injection",
		"payment_stuck_pending",
		"payment_transitions",
		"random_5xx",
		"rate_limit_burst",
		"scope_downgrade",
		"token_short_ttl",
		"webhook_delivery_fail",
	}
	got := scenarios.Names()
	if !slices.Equal(got, want) {
		t.Errorf("Names() = %v\nwant %v", got, want)
	}
}

func TestArgs_DurationFallback(t *testing.T) {
	a := scenarios.Args{}
	d, err := a.Duration("missing", 100*time.Millisecond)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if d != 100*time.Millisecond {
		t.Errorf("d = %v", d)
	}
}

func TestArgs_FloatParse(t *testing.T) {
	a := scenarios.Args{"rate": "0.25"}
	f, err := a.Float("rate", 1.0)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if f != 0.25 {
		t.Errorf("f = %v", f)
	}
}

func TestArgs_CSV(t *testing.T) {
	a := scenarios.Args{"codes": "500, 502, 503"}
	got := a.CSV("codes")
	want := []string{"500", "502", "503"}
	if !slices.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestEngine_IsEnabled_RoundTripsThroughStore(t *testing.T) {
	st, err := store.Open(filepath.Join(t.TempDir(), "state.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })
	eng := scenarios.New(st)

	if _, ok := eng.IsEnabled("random_5xx"); ok {
		t.Errorf("should be disabled before enable")
	}
	if err := st.EnableScenario("random_5xx", map[string]string{"rate": "0.5"}); err != nil {
		t.Fatalf("enable: %v", err)
	}
	args, ok := eng.IsEnabled("random_5xx")
	if !ok {
		t.Fatalf("should be enabled")
	}
	if args["rate"] != "0.5" {
		t.Errorf("args = %v", args)
	}
}

func TestFiredHolder_AppendAndRead(t *testing.T) {
	h := &scenarios.FiredHolder{}
	ctx := scenarios.WithFiredHolder(context.Background(), h)
	scenarios.Record(ctx, "latency_injection", "+50ms")
	scenarios.Record(ctx, "random_5xx", "503")
	entries := h.Entries()
	if len(entries) != 2 {
		t.Fatalf("entries = %v", entries)
	}
	if entries[0].Name != "latency_injection" || entries[1].Name != "random_5xx" {
		t.Errorf("order wrong: %+v", entries)
	}
}

func TestRecord_NoHolderInContext_NoPanic(t *testing.T) {
	// Should not panic when no holder is attached.
	scenarios.Record(context.Background(), "random_5xx", "503")
}
