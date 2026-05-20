package scenarios

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"
)

// ScenarioRateLimitBurst is the scenario name.
const ScenarioRateLimitBurst = "rate_limit_burst"

// Default args per scaffold §13 fallback table (100 RPS placeholder).
const (
	defaultRateLimit  = 100
	defaultRateWindow = time.Second
)

func init() {
	register(Descriptor{
		Name:        ScenarioRateLimitBurst,
		Description: "Reject requests with 429 after N within a window per client",
		Args: []ArgDef{
			{Name: "limit", Type: ArgInt, Default: strconv.Itoa(defaultRateLimit), Description: "max requests per window per client"},
			{Name: "window", Type: ArgDuration, Default: "1s", Description: "rolling window size"},
			{Name: "retry_after", Type: ArgDuration, Default: "1s", Description: "Retry-After header value"},
			{Name: "operations", Type: ArgCSV, Description: "operation path prefixes to enforce (empty = all api/v1)"},
		},
	})
}

// rateLimitTracker holds per-client request timestamps inside a rolling
// window. Reset when the scenario is disabled-then-reenabled.
var rateLimitTracker struct {
	mu      sync.Mutex
	clients map[string][]time.Time
}

// MaybeRateLimit returns (retryAfter, true) when the scenario fires for
// the supplied client + path. The caller should respond with 429 +
// Retry-After header.
func MaybeRateLimit(ctx context.Context, eng *Engine, clientID, path string) (time.Duration, bool) {
	if eng == nil {
		return 0, false
	}
	args, ok := eng.IsEnabled(ScenarioRateLimitBurst)
	if !ok {
		return 0, false
	}
	if !pathMatchesArg(args, path) {
		return 0, false
	}
	limit, err := args.Int("limit", defaultRateLimit)
	if err != nil || limit <= 0 {
		return 0, false
	}
	window, err := args.Duration("window", defaultRateWindow)
	if err != nil || window <= 0 {
		return 0, false
	}
	retryAfter, err := args.Duration("retry_after", time.Second)
	if err != nil || retryAfter <= 0 {
		retryAfter = time.Second
	}
	key := clientID
	if key == "" {
		key = "anonymous"
	}

	rateLimitTracker.mu.Lock()
	if rateLimitTracker.clients == nil {
		rateLimitTracker.clients = make(map[string][]time.Time)
	}
	now := time.Now()
	cutoff := now.Add(-window)
	prev := rateLimitTracker.clients[key]
	kept := prev[:0]
	for _, t := range prev {
		if t.After(cutoff) {
			kept = append(kept, t)
		}
	}
	if len(kept) >= limit {
		rateLimitTracker.clients[key] = kept
		rateLimitTracker.mu.Unlock()
		Record(ctx, ScenarioRateLimitBurst, fmt.Sprintf("429:%s", retryAfter))
		return retryAfter, true
	}
	kept = append(kept, now)
	rateLimitTracker.clients[key] = kept
	rateLimitTracker.mu.Unlock()
	return 0, false
}
