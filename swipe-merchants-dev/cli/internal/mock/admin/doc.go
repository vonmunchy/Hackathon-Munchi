// Package admin implements the out-of-spec /_admin/* endpoints that the CLI
// uses to manage the running mock (status, keys, scenarios, reset). They
// are localhost-only and never appear in the public OpenAPI spec, so
// merchants cannot take a dependency on them. Phase 1 ships only the
// /_admin/status endpoint.
package admin
