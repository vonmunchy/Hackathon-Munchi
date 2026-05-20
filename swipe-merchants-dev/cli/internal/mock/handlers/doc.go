// Package handlers wires public OpenAPI operations to chi http.Handlers.
// In Phase 1 it covers only the health probes; later phases attach the
// merchant resource handlers (payments, payouts, history, balance, etc.).
package handlers
