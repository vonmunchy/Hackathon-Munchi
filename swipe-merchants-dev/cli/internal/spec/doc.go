// Package spec embeds the canonical OpenAPI spec (spec/app.yaml at the repo
// root, copied here at build time via `make sync-spec`) and exposes a parsed
// view of it for both the mock server and CLI `spec` subcommands.
package spec
