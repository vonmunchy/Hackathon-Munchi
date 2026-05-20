package scenarios

import (
	"context"
	"slices"
	"strings"
)

// ScenarioScopeDowngrade is the scenario name.
const ScenarioScopeDowngrade = "scope_downgrade"

func init() {
	register(Descriptor{
		Name:        ScenarioScopeDowngrade,
		Description: "Strip listed scopes from newly-issued tokens even if the client holds them",
		Args: []ArgDef{
			{Name: "remove", Type: ArgCSV, Description: "scopes to drop from the granted set"},
		},
	})
}

// FilterScopes returns a possibly-modified copy of the granted scope
// slice when the scenario is enabled. The original is returned when
// disabled. The fired holder is updated whenever a scope is actually
// stripped (a no-op enable shouldn't pollute the log).
func FilterScopes(ctx context.Context, eng *Engine, granted []string) []string {
	if eng == nil {
		return granted
	}
	args, ok := eng.IsEnabled(ScenarioScopeDowngrade)
	if !ok {
		return granted
	}
	remove := args.CSV("remove")
	if len(remove) == 0 {
		return granted
	}
	out := make([]string, 0, len(granted))
	dropped := make([]string, 0, len(granted))
	for _, s := range granted {
		if slices.Contains(remove, s) {
			dropped = append(dropped, s)
			continue
		}
		out = append(out, s)
	}
	if len(dropped) > 0 {
		Record(ctx, ScenarioScopeDowngrade, "-"+strings.Join(dropped, ","))
	}
	return out
}
