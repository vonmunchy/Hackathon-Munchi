package scenarios

import (
	"context"
	"strings"
	"time"
)

// ScenarioPaymentTransitions is the scenario name.
const ScenarioPaymentTransitions = "payment_transitions"

func init() {
	register(Descriptor{
		Name:        ScenarioPaymentTransitions,
		Description: "Override the auto-transition target status of newly-created payments",
		Args: []ArgDef{
			{Name: "to", Type: ArgString, Default: "EXPIRED", Description: "terminal status to transition to (COMPLETED|EXPIRED|CANCELLED)"},
			{Name: "after", Type: ArgDuration, Default: "", Description: "delay before transition; empty = same as the mock default TTL"},
		},
	})
}

// PaymentTransitionOverride returns the (target status, optional delay
// override) when the scenario is enabled. The defaultStatus is returned
// when disabled.
func PaymentTransitionOverride(ctx context.Context, eng *Engine, defaultStatus string) (status string, delay time.Duration, fired bool) {
	if eng == nil {
		return defaultStatus, 0, false
	}
	args, ok := eng.IsEnabled(ScenarioPaymentTransitions)
	if !ok {
		return defaultStatus, 0, false
	}
	to := strings.ToUpper(args.StringOr("to", "EXPIRED"))
	switch to {
	case "COMPLETED", "EXPIRED", "CANCELLED":
	default:
		return defaultStatus, 0, false
	}
	delay, _ = args.Duration("after", 0)
	Record(ctx, ScenarioPaymentTransitions, to)
	return to, delay, true
}
