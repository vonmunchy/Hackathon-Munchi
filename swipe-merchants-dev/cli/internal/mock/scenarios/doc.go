// Package scenarios implements the V1 scenarios catalog (scaffold §7,
// D-020, D-021). Scenarios are hardcoded Go types registered at init()
// time; the engine consults the BoltDB `scenarios` bucket for enabled-
// state + args at each integration point (pre-handler middleware, token
// issuer, payment creation, webhook dispatcher, background timers).
//
// Per D-021, every request log entry records which scenarios fired and
// with what effect — without that, scenarios become spooky. The engine
// pushes named effects onto a per-request FiredHolder that the outer
// request-logger middleware drains at end-of-request.
package scenarios
