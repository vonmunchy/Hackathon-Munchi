package events

import (
	"sync"
	"time"
)

// Event is the on-bus shape of a state-change notification. The Phase 4
// flow only publishes payment transitions; later phases (payouts,
// webhooks) reuse the same envelope by varying Status / Resource.
type Event struct {
	Resource   string    `json:"resource"`           // e.g. "payment", "payout"
	ResourceID string    `json:"resource_id"`        // e.g. pay_<ulid>
	Status     string    `json:"status"`             // PENDING | COMPLETED | EXPIRED | CANCELLED
	Previous   string    `json:"previous,omitempty"` // prior status, when known
	OccurredAt time.Time `json:"occurred_at"`
	MerchantID string    `json:"merchant_id,omitempty"`
}

// TopicAll is the well-known topic name for "every event". Subscribers use
// it when they want to receive notifications for every resource.
const TopicAll = "*"

// Bus is a single-process pub/sub bus. It is safe for concurrent use.
type Bus struct {
	mu          sync.RWMutex
	subscribers map[string][]*subscription
	closed      bool
}

// subscription wraps a single subscriber's delivery channel + topic.
type subscription struct {
	topic string
	ch    chan Event
}

// New builds an empty Bus.
func New() *Bus {
	return &Bus{subscribers: make(map[string][]*subscription)}
}

// Subscribe registers a subscriber for the given topic. The returned
// channel receives events; the returned cancel func unregisters and
// closes the channel. The channel has a small buffer; slow subscribers
// drop the oldest events to keep publishers from blocking.
//
// Topic "*" (TopicAll) receives every published event.
func (b *Bus) Subscribe(topic string) (<-chan Event, func()) {
	sub := &subscription{topic: topic, ch: make(chan Event, 16)}
	b.mu.Lock()
	if b.closed {
		b.mu.Unlock()
		close(sub.ch)
		return sub.ch, func() {}
	}
	b.subscribers[topic] = append(b.subscribers[topic], sub)
	b.mu.Unlock()
	return sub.ch, func() { b.unsubscribe(sub) }
}

// Publish delivers e to every subscriber of e.ResourceID's topic plus
// every TopicAll subscriber. Delivery is best-effort: subscribers whose
// channels are full have the oldest event dropped to make room.
func (b *Bus) Publish(e Event) {
	if e.OccurredAt.IsZero() {
		e.OccurredAt = time.Now().UTC()
	}
	b.mu.RLock()
	defer b.mu.RUnlock()
	if b.closed {
		return
	}
	deliver(b.subscribers[e.ResourceID], e)
	deliver(b.subscribers[TopicAll], e)
}

// Close marks the bus as closed and drops every subscription. Existing
// subscriber channels are closed so range-receive loops exit cleanly.
func (b *Bus) Close() {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.closed {
		return
	}
	b.closed = true
	for topic, subs := range b.subscribers {
		for _, s := range subs {
			close(s.ch)
		}
		delete(b.subscribers, topic)
	}
}

func (b *Bus) unsubscribe(sub *subscription) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.closed {
		return
	}
	list := b.subscribers[sub.topic]
	for i, s := range list {
		if s == sub {
			b.subscribers[sub.topic] = append(list[:i], list[i+1:]...)
			close(s.ch)
			return
		}
	}
}

// deliver runs while b.mu is held in read mode. Drops the oldest message
// when a subscriber's channel is full so a slow consumer cannot block the
// publisher (and through it, the calling HTTP handler).
func deliver(subs []*subscription, e Event) {
	for _, s := range subs {
		select {
		case s.ch <- e:
		default:
			// Channel full — drop the oldest queued message and try again.
			select {
			case <-s.ch:
			default:
			}
			select {
			case s.ch <- e:
			default:
				// still full (rare: subscriber gone away mid-receive); skip.
			}
		}
	}
}
