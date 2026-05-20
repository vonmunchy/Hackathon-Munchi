package scenarios

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"strconv"
)

// ScenarioRandom5xx is the scenario name.
const ScenarioRandom5xx = "random_5xx"

func init() {
	register(Descriptor{
		Name:        ScenarioRandom5xx,
		Description: "Return a 5xx response on a fraction of requests",
		Args: []ArgDef{
			{Name: "rate", Type: ArgFloat, Default: "0.5", Description: "fraction in [0,1] of requests to fail"},
			{Name: "codes", Type: ArgCSV, Default: "500,502,503", Description: "candidate status codes"},
			{Name: "operations", Type: ArgCSV, Description: "operation path prefixes to fail (empty = all api/v1)"},
		},
	})
}

// MaybeRandom5xx returns (status, true) when the scenario fires for this
// request. Caller should write a ProblemDetails response with the given
// status and bypass further processing.
func MaybeRandom5xx(ctx context.Context, eng *Engine, path string) (int, bool) {
	if eng == nil {
		return 0, false
	}
	args, ok := eng.IsEnabled(ScenarioRandom5xx)
	if !ok {
		return 0, false
	}
	if !pathMatchesArg(args, path) {
		return 0, false
	}
	rate, err := args.Float("rate", 0.5)
	if err != nil || rate <= 0 {
		return 0, false
	}
	if rate > 1 {
		rate = 1
	}
	if randomFloat() >= rate {
		return 0, false
	}
	status := pickStatus(args)
	Record(ctx, ScenarioRandom5xx, strconv.Itoa(status))
	return status, true
}

// randomFloat returns a uniform float in [0,1). Uses crypto/rand because
// scaffold §10 forbids math/rand in production paths (timing is sensitive
// enough that a deterministic test seed would defeat the scenario).
func randomFloat() float64 {
	var buf [8]byte
	_, _ = rand.Read(buf[:])
	u := binary.BigEndian.Uint64(buf[:]) & ((1 << 53) - 1)
	return float64(u) / float64(1<<53)
}

// pickStatus selects a status code from the comma-separated `codes` arg.
// Falls back to 500 when no valid value is configured.
func pickStatus(args Args) int {
	codes := args.CSV("codes")
	if len(codes) == 0 {
		codes = []string{"500", "502", "503"}
	}
	idx := randomIndex(len(codes))
	n, err := strconv.Atoi(codes[idx])
	if err != nil || n < 500 || n > 599 {
		return 500
	}
	return n
}

// randomIndex returns a uniform integer in [0, n). n must be >= 1.
func randomIndex(n int) int {
	if n <= 1 {
		return 0
	}
	var buf [8]byte
	_, _ = rand.Read(buf[:])
	u := binary.BigEndian.Uint64(buf[:])
	return int(u % uint64(n)) //nolint:gosec // intentional non-crypto bounded conversion
}
