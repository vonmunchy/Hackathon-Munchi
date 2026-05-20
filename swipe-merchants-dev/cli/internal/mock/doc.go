// Package mock implements the embedded mock of the Swipe Merchants API. It
// owns the HTTP server lifecycle, the BoltDB-backed state, and (in later
// phases) OAuth issuance, request validation, scenarios, and webhook
// delivery. Phase 1 covers only health endpoints, the /_admin/status
// endpoint, the lifecycle (start/stop/status), and seed data.
package mock
