package store

import (
	"encoding/json"
	"fmt"
	"sort"

	bolt "go.etcd.io/bbolt"
)

// ScenarioState captures the enabled state + args for one scenario. Stored
// in the BoltDB `scenarios` bucket keyed by the scenario name.
type ScenarioState struct {
	Name    string            `json:"name"`
	Enabled bool              `json:"enabled"`
	Args    map[string]string `json:"args,omitempty"`
}

// EnableScenario marks the named scenario as enabled with the supplied
// args. Existing args for the same name are replaced.
func (s *Store) EnableScenario(name string, args map[string]string) error {
	if name == "" {
		return fmt.Errorf("enable scenario: name required")
	}
	return s.db.Update(func(tx *bolt.Tx) error {
		buf, _ := json.Marshal(ScenarioState{Name: name, Enabled: true, Args: args})
		return tx.Bucket(bucketScenarios).Put([]byte(name), buf)
	})
}

// DisableScenario removes the scenario record entirely.
func (s *Store) DisableScenario(name string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(bucketScenarios).Delete([]byte(name))
	})
}

// GetScenario returns the persisted state for name; ok=false when absent.
func (s *Store) GetScenario(name string) (ScenarioState, bool, error) {
	var out ScenarioState
	var ok bool
	err := s.db.View(func(tx *bolt.Tx) error {
		buf := tx.Bucket(bucketScenarios).Get([]byte(name))
		if buf == nil {
			return nil
		}
		if err := json.Unmarshal(buf, &out); err != nil {
			return fmt.Errorf("unmarshal scenario %s: %w", name, err)
		}
		ok = true
		return nil
	})
	if err != nil {
		return ScenarioState{}, false, err
	}
	return out, ok, nil
}

// ListScenarioStates returns every persisted scenario record ordered by
// name. Empty bucket yields a nil slice (not error).
func (s *Store) ListScenarioStates() ([]ScenarioState, error) {
	var out []ScenarioState
	err := s.db.View(func(tx *bolt.Tx) error {
		return tx.Bucket(bucketScenarios).ForEach(func(_, v []byte) error {
			var st ScenarioState
			if err := json.Unmarshal(v, &st); err != nil {
				return fmt.Errorf("unmarshal scenario: %w", err)
			}
			out = append(out, st)
			return nil
		})
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}
