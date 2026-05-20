package scenarios

import (
	"fmt"
	"sort"
	"sync"
)

// registry holds every Descriptor registered via init(). Registrations
// happen at package init time and are read concurrently afterward; a
// mutex protects the rare case where tests want to register dynamically.
var registryState struct {
	mu      sync.RWMutex
	entries map[string]Descriptor
}

// register installs d in the registry. Called from each scenario's init().
// Duplicate names panic — they signal a copy/paste bug in scenario files.
func register(d Descriptor) {
	registryState.mu.Lock()
	defer registryState.mu.Unlock()
	if registryState.entries == nil {
		registryState.entries = make(map[string]Descriptor)
	}
	if _, exists := registryState.entries[d.Name]; exists {
		panic(fmt.Sprintf("scenarios: duplicate registration for %q", d.Name))
	}
	registryState.entries[d.Name] = d
}

// List returns every registered Descriptor sorted by name. Stable order
// matters for `scenarios list` output + tests.
func List() []Descriptor {
	registryState.mu.RLock()
	defer registryState.mu.RUnlock()
	out := make([]Descriptor, 0, len(registryState.entries))
	for _, d := range registryState.entries {
		out = append(out, d)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// Get returns the descriptor for name and whether it exists.
func Get(name string) (Descriptor, bool) {
	registryState.mu.RLock()
	defer registryState.mu.RUnlock()
	d, ok := registryState.entries[name]
	return d, ok
}

// Names returns every registered scenario name in stable order.
func Names() []string {
	descriptors := List()
	out := make([]string, 0, len(descriptors))
	for _, d := range descriptors {
		out = append(out, d.Name)
	}
	return out
}
