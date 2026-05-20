package scenarios

import (
	"context"
	"strconv"
	"sync"
)

// ScenarioWebhookDeliveryFail is the scenario name.
const ScenarioWebhookDeliveryFail = "webhook_delivery_fail"

// failureCounter tracks how many deliveries have already been forced to
// fail since the scenario was last enabled. Module-level so the count
// survives across requests (the dispatcher is one goroutine but the
// counter is independent of its lifecycle).
var failureCounter struct {
	mu    sync.Mutex
	count int
}

func init() {
	register(Descriptor{
		Name:        ScenarioWebhookDeliveryFail,
		Description: "Force the first N webhook deliveries to fail (logged, not retried)",
		Args: []ArgDef{
			{Name: "count", Type: ArgInt, Default: "3", Description: "number of upcoming deliveries to fail"},
		},
	})
}

// ConsumeWebhookFailure returns true when the next delivery should be
// forced to fail. Each true return increments the internal counter; once
// it reaches the configured `count`, subsequent calls return false until
// the scenario is disabled-then-reenabled (Reset).
func ConsumeWebhookFailure(ctx context.Context, eng *Engine) bool {
	if eng == nil {
		return false
	}
	args, ok := eng.IsEnabled(ScenarioWebhookDeliveryFail)
	if !ok {
		return false
	}
	limit, err := args.Int("count", 3)
	if err != nil || limit <= 0 {
		return false
	}
	failureCounter.mu.Lock()
	defer failureCounter.mu.Unlock()
	if failureCounter.count >= limit {
		return false
	}
	failureCounter.count++
	Record(ctx, ScenarioWebhookDeliveryFail, strconv.Itoa(failureCounter.count)+"/"+strconv.Itoa(limit))
	return true
}

// ResetWebhookFailureCounter clears the internal counter. Called by the
// engine when the scenario transitions from disabled to enabled.
func ResetWebhookFailureCounter() {
	failureCounter.mu.Lock()
	defer failureCounter.mu.Unlock()
	failureCounter.count = 0
}
