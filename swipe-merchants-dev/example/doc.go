// Package main is a developer-facing sample integration against the Swipe
// Merchants mock server. It exercises the API end-to-end the way a real
// merchant integration would: HTTP only, no internal imports.
//
// The example grows phase-by-phase as the mock implements more of the spec.
// Today (Phase 1) it only probes the health endpoints. Future phases will
// add: OAuth token exchange (P2), wallet balance + bank accounts (P3),
// create payment + SSE status stream (P4), webhook verification (P5),
// scenario toggling (P6), and a full happy-path demo (P7).
//
// Run it via `make sample` from the repo root, which starts the mock in
// the background, runs this binary, and tears the mock down on exit.
package main
