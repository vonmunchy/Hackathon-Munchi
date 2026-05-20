package scenarios

import (
	"context"
	"sync"

	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/store"
)

// Engine glues the BoltDB scenarios bucket to the per-integration-point
// callers (pre-handler middleware, issuer hooks, payment-create hooks,
// webhook dispatcher hook). The Engine has no goroutine of its own — it
// just reads enabled state synchronously when consulted.
type Engine struct {
	Store *store.Store
}

// New builds an Engine backed by st.
func New(st *store.Store) *Engine {
	return &Engine{Store: st}
}

// IsEnabled returns (args, true) when scenario `name` is enabled.
func (e *Engine) IsEnabled(name string) (Args, bool) {
	if e == nil || e.Store == nil {
		return nil, false
	}
	state, ok, err := e.Store.GetScenario(name)
	if err != nil || !ok || !state.Enabled {
		return nil, false
	}
	return Args(state.Args), true
}

// ListEnabled returns the names of currently-enabled scenarios in stable
// sorted order. Useful for non-handler callers that want a quick scan.
func (e *Engine) ListEnabled() []string {
	states, err := e.Store.ListScenarioStates()
	if err != nil {
		return nil
	}
	out := make([]string, 0, len(states))
	for _, s := range states {
		if s.Enabled {
			out = append(out, s.Name)
		}
	}
	return out
}

// FiredEntry is one annotation appended to a request's fired-scenarios
// holder. Effect is a short, human-readable description of what the
// scenario did (e.g. "503", "+200ms", "rate_limit").
type FiredEntry struct {
	Name   string `json:"name"`
	Effect string `json:"effect,omitempty"`
}

// FiredHolder is the per-request scratch space scenarios append to when
// they fire. The outer request-logger middleware reads the entries and
// stores them on the RequestLogEntry (D-021).
type FiredHolder struct {
	mu      sync.Mutex
	entries []FiredEntry
}

// Append records that name fired with the given effect. Safe for
// concurrent use; scenarios may run in different goroutines for SSE etc.
func (h *FiredHolder) Append(name, effect string) {
	if h == nil {
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	h.entries = append(h.entries, FiredEntry{Name: name, Effect: effect})
}

// Entries returns a copy of the recorded entries.
func (h *FiredHolder) Entries() []FiredEntry {
	if h == nil {
		return nil
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	out := make([]FiredEntry, len(h.entries))
	copy(out, h.entries)
	return out
}

// holderKey is the context key for FiredHolder. Unexported type ensures
// no foreign package can collide.
type holderKey struct{}

// WithFiredHolder attaches an empty holder to ctx so inner middleware /
// handlers can append.
func WithFiredHolder(ctx context.Context, h *FiredHolder) context.Context {
	return context.WithValue(ctx, holderKey{}, h)
}

// FiredHolderFromContext returns the attached holder, if any.
func FiredHolderFromContext(ctx context.Context) (*FiredHolder, bool) {
	h, ok := ctx.Value(holderKey{}).(*FiredHolder)
	return h, ok
}

// Record is the one-call shortcut for scenario code: pull the holder
// from ctx (if present) and append (name, effect).
func Record(ctx context.Context, name, effect string) {
	if h, ok := FiredHolderFromContext(ctx); ok {
		h.Append(name, effect)
	}
}
