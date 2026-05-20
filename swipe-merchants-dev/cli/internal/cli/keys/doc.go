// Package keys implements the `swipe keys` CLI tree. The subcommands hit
// the mock's /_admin/keys endpoints (D-009) to create, list, inspect,
// update, rotate, revoke, and delete OAuth clients.
//
// Per D-023 + D-010, client secrets are only ever printed on `keys create`
// and `keys rotate` responses, with a one-time-save warning attached.
package keys
