// Package webhooks implements Standard Webhooks-style signing (D-018) and
// the fire-and-forget dispatcher (D-019) that POSTs `transaction.state_changed`
// events to the URL configured via `swipe mock start --webhook-url`.
package webhooks
