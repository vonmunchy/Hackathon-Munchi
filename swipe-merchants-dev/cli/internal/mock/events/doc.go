// Package events provides the in-process event bus that connects the state
// store to the SSE streamer and (Phase 5+) the webhook dispatcher. It is a
// single-process pub/sub with per-payment topics and a global "all" topic.
package events
