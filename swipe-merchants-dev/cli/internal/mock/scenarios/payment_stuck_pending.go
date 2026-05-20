package scenarios

import (
	"context"
	"time"
)

// ScenarioPaymentStuckPending is the scenario name.
const ScenarioPaymentStuckPending = "payment_stuck_pending"

// forever is the sentinel TransitionAt for payments that never auto-
// complete under this scenario. 100 years out is functionally forever for
// a local dev process while staying within time.Duration range.
var forever = 100 * 365 * 24 * time.Hour

func init() {
	register(Descriptor{
		Name:        ScenarioPaymentStuckPending,
		Description: "Newly-created payments stay PENDING indefinitely (or for the configured duration)",
		Args: []ArgDef{
			{Name: "duration", Type: ArgDuration, Default: "", Description: "if set, payments transition normally but only after this much extra time; empty = never"},
		},
	})
}

// PaymentStuckOverride returns the TransitionAt to use for a newly-
// created payment (overriding the default TTL). The bool reports whether
// the scenario fired at all.
func PaymentStuckOverride(ctx context.Context, eng *Engine, defaultAt time.Time) (time.Time, bool) {
	if eng == nil {
		return defaultAt, false
	}
	args, ok := eng.IsEnabled(ScenarioPaymentStuckPending)
	if !ok {
		return defaultAt, false
	}
	dur, err := args.Duration("duration", 0)
	if err != nil {
		return defaultAt, false
	}
	var target time.Time
	effect := "forever"
	if dur > 0 {
		target = time.Now().UTC().Add(dur)
		effect = dur.String()
	} else {
		target = time.Now().UTC().Add(forever)
	}
	Record(ctx, ScenarioPaymentStuckPending, effect)
	return target, true
}
