// Package auth implements the `swipe auth` CLI tree (Phase 2). It performs
// the client_credentials grant against the mock's POST /oauth2/token,
// caches the result at ~/.swipe/token.json (D-025), and provides a
// `whoami` helper that calls /api/v1/whoami with the cached bearer token.
package auth
