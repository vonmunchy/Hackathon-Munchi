// Package transport is the only place that knows the swipe binary's HTTP
// base URL. CLI resource commands call methods on Client; admin commands
// call methods on AdminClient. By keeping the URL out of the resource
// commands V2 can swap the mock for the real Swipe API without touching
// any command file (D-003).
package transport
