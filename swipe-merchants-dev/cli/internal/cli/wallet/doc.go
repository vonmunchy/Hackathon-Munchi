// Package wallet implements the `swipe wallet` CLI subtree (Phase 3). It
// hits the authenticated read endpoints `/api/v1/balance` and
// `/api/v1/bank-accounts` using the cached bearer token from
// `~/.swipe/token.json` (per D-025).
package wallet
