package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/store"
)

// HealthResponse mirrors the spec's HealthResponse schema.
type HealthResponse struct {
	Status string `json:"status"`
}

// ReadinessResponse mirrors the spec's ReadinessResponse schema. The
// `checks` field carries individual dependency probes; the mock currently
// has only one (the embedded BoltDB store), but the field is shaped to
// accept additional checks (redis, temporal) without breaking schema.
type ReadinessResponse struct {
	Status string                  `json:"status"`
	Checks map[string]ReadinessHit `json:"checks"`
}

// ReadinessHit mirrors the spec's HealthCheck schema.
type ReadinessHit struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

// Health bundles the dependencies the health endpoints need: a database
// readiness probe.
type Health struct {
	// Store is the BoltDB-backed state store. Its readiness is the only
	// dependency probe in Phase 1.
	Store *store.Store
}

// Alive serves GET /health/alive. It always returns 200 with `{"status":"ok"}`.
// The liveness probe never inspects dependencies (per spec).
func (h *Health) Alive(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, HealthResponse{Status: "ok"})
}

// Ready serves GET /health/ready. It probes every registered dependency
// and returns 200 with `status: ok` if all healthy, otherwise 503 with
// `status: unhealthy` and per-check details.
func (h *Health) Ready(w http.ResponseWriter, _ *http.Request) {
	resp := ReadinessResponse{
		Status: "ok",
		Checks: map[string]ReadinessHit{},
	}
	dbHit := h.probeDatabase()
	resp.Checks["database"] = dbHit
	if dbHit.Status != "ok" {
		resp.Status = "unhealthy"
	}
	status := http.StatusOK
	if resp.Status != "ok" {
		status = http.StatusServiceUnavailable
	}
	writeJSON(w, status, resp)
}

// probeDatabase performs a trivial bolt operation that verifies the
// mock-default merchant is readable. It is fast (microseconds) and
// indirectly verifies the file handle, the buckets, and the seed.
func (h *Health) probeDatabase() ReadinessHit {
	if h == nil || h.Store == nil {
		return ReadinessHit{Status: "unhealthy", Message: "store not initialized"}
	}
	if _, err := h.Store.GetMerchant(store.DefaultMerchantID); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return ReadinessHit{Status: "unhealthy", Message: "default merchant not seeded"}
		}
		return ReadinessHit{Status: "unhealthy", Message: err.Error()}
	}
	return ReadinessHit{Status: "ok"}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
