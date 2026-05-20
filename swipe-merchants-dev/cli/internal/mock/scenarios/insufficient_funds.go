package scenarios

import "context"

// ScenarioInsufficientFunds is the scenario name.
const ScenarioInsufficientFunds = "insufficient_funds"

func init() {
	register(Descriptor{
		Name:        ScenarioInsufficientFunds,
		Description: "Reject createPayment with INSUFFICIENT_FUNDS regardless of wallet balance",
		Args: []ArgDef{
			{Name: "rate", Type: ArgFloat, Default: "1.0", Description: "fraction in [0,1] of createPayment calls to reject"},
		},
	})
}

// MaybeInsufficientFunds returns true when the scenario fires for this
// payment-create request. Caller responds with 402 INSUFFICIENT_FUNDS.
func MaybeInsufficientFunds(ctx context.Context, eng *Engine) bool {
	if eng == nil {
		return false
	}
	args, ok := eng.IsEnabled(ScenarioInsufficientFunds)
	if !ok {
		return false
	}
	rate, err := args.Float("rate", 1.0)
	if err != nil || rate <= 0 {
		return false
	}
	if rate < 1 && randomFloat() >= rate {
		return false
	}
	Record(ctx, ScenarioInsufficientFunds, "402")
	return true
}
