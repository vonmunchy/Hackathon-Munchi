package scenarios

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// ScenarioLatency is the scenario name as published over /_admin/scenarios.
const ScenarioLatency = "latency_injection"

func init() {
	register(Descriptor{
		Name:        ScenarioLatency,
		Description: "Add artificial latency to responses",
		Args: []ArgDef{
			{Name: "delay", Type: ArgDuration, Default: "100ms", Description: "fixed delay added per matched request"},
			{Name: "operations", Type: ArgCSV, Description: "operation path prefixes to delay (empty = all api/v1)"},
		},
	})
}

// ApplyLatency consults the engine for the latency_injection scenario and
// blocks for the configured delay when the request path matches. Returns
// the delay actually applied (0 when nothing fired) so the caller can
// record it in the fired holder.
func ApplyLatency(ctx context.Context, eng *Engine, path string) time.Duration {
	if eng == nil {
		return 0
	}
	args, ok := eng.IsEnabled(ScenarioLatency)
	if !ok {
		return 0
	}
	if !pathMatchesArg(args, path) {
		return 0
	}
	delay, err := args.Duration("delay", 100*time.Millisecond)
	if err != nil || delay <= 0 {
		return 0
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
	case <-timer.C:
	}
	Record(ctx, ScenarioLatency, fmt.Sprintf("+%s", delay))
	return delay
}

// pathMatchesArg returns true if the supplied path is in the operations
// csv (matches by prefix) or the csv is empty (match-all).
func pathMatchesArg(args Args, path string) bool {
	prefixes := args.CSV("operations")
	if len(prefixes) == 0 {
		return strings.HasPrefix(path, "/api/v1")
	}
	for _, p := range prefixes {
		if strings.HasPrefix(path, p) {
			return true
		}
	}
	return false
}
