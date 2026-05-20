// Package errors renders failing CLI errors with extra structure when
// the underlying cause carries a ProblemDetails JSON envelope. The
// goal: replace `Error: balance: balance: status 403: {"type":"...","detail":"..."}`
// with a clean, actionable two-line message.
package errors
