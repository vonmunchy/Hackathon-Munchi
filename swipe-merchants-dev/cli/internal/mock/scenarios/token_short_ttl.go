package scenarios

import (
	"context"
	"time"
)

// ScenarioTokenShortTTL is the scenario name.
const ScenarioTokenShortTTL = "token_short_ttl" //nolint:gosec // scenario name string, not a credential

func init() {
	register(Descriptor{
		Name:        ScenarioTokenShortTTL,
		Description: "Issue OAuth tokens with a short TTL instead of the default",
		Args: []ArgDef{
			{Name: "ttl", Type: ArgDuration, Default: "10s", Description: "lifetime of newly-issued tokens"},
		},
	})
}

// TokenTTLOverride returns (ttl, true) when token_short_ttl is enabled.
// The auth issuer consults this before each token mint.
func TokenTTLOverride(ctx context.Context, eng *Engine) (time.Duration, bool) {
	if eng == nil {
		return 0, false
	}
	args, ok := eng.IsEnabled(ScenarioTokenShortTTL)
	if !ok {
		return 0, false
	}
	ttl, err := args.Duration("ttl", 10*time.Second)
	if err != nil || ttl <= 0 {
		return 0, false
	}
	Record(ctx, ScenarioTokenShortTTL, ttl.String())
	return ttl, true
}
