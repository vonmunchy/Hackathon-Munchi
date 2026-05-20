package events_test

import (
	"sync"
	"testing"
	"time"

	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/events"
)

func TestBus_SubscribeAndPublish_DeliversToPerTopicAndAll(t *testing.T) {
	bus := events.New()
	t.Cleanup(bus.Close)

	scoped, cancelScoped := bus.Subscribe("pay_x")
	all, cancelAll := bus.Subscribe(events.TopicAll)
	t.Cleanup(cancelScoped)
	t.Cleanup(cancelAll)

	bus.Publish(events.Event{Resource: "payment", ResourceID: "pay_x", Status: "PENDING"})

	select {
	case e := <-scoped:
		if e.Status != "PENDING" {
			t.Errorf("scoped status = %q", e.Status)
		}
	case <-time.After(time.Second):
		t.Fatal("scoped did not receive event")
	}

	select {
	case e := <-all:
		if e.ResourceID != "pay_x" {
			t.Errorf("all resource = %q", e.ResourceID)
		}
	case <-time.After(time.Second):
		t.Fatal("all did not receive event")
	}
}

func TestBus_Unsubscribe_StopsDeliveryAndClosesChannel(t *testing.T) {
	bus := events.New()
	t.Cleanup(bus.Close)

	ch, cancel := bus.Subscribe("pay_y")
	cancel()

	bus.Publish(events.Event{ResourceID: "pay_y", Status: "PENDING"})
	select {
	case _, open := <-ch:
		if open {
			t.Errorf("expected channel closed after cancel")
		}
	case <-time.After(time.Second):
		t.Fatal("read after cancel blocked")
	}
}

func TestBus_Close_ClosesAllSubscribers(t *testing.T) {
	bus := events.New()
	a, _ := bus.Subscribe("pay_a")
	b, _ := bus.Subscribe(events.TopicAll)
	bus.Close()

	for i, ch := range []<-chan events.Event{a, b} {
		select {
		case _, open := <-ch:
			if open {
				t.Errorf("subscriber %d: expected closed channel", i)
			}
		case <-time.After(time.Second):
			t.Errorf("subscriber %d: read blocked after close", i)
		}
	}
}

func TestBus_ConcurrentPublishSubscribe_DoesNotPanic(t *testing.T) {
	bus := events.New()
	t.Cleanup(bus.Close)

	var wg sync.WaitGroup
	for range 5 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range 50 {
				bus.Publish(events.Event{ResourceID: "pay_z", Status: "PENDING", Previous: ""})
				_ = i
			}
		}()
	}
	wg.Wait()
}
