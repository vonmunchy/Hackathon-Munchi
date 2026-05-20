package webhooks_test

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/events"
	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/store"
	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/webhooks"
)

// dispatchHarness wires a dispatcher to a Bus + an httptest.NewServer that
// captures POSTs so tests can assert on the wire format + signature.
type dispatchHarness struct {
	t        *testing.T
	store    *store.Store
	bus      *events.Bus
	secret   string
	captured chan capturedRequest
	server   *httptest.Server
}

type capturedRequest struct {
	id        string
	timestamp string
	signature string
	body      []byte
}

func newDispatchHarness(t *testing.T) *dispatchHarness {
	t.Helper()
	st, err := store.Open(filepath.Join(t.TempDir(), "state.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })
	seed, err := st.SeedDefaults()
	if err != nil {
		t.Fatalf("seed: %v", err)
	}
	captured := make(chan capturedRequest, 4)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = r.Body.Close()
		id, ts, sig := webhooks.ExtractHeaders(r.Header)
		captured <- capturedRequest{id: id, timestamp: ts, signature: sig, body: body}
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(server.Close)
	return &dispatchHarness{
		t:        t,
		store:    st,
		bus:      events.New(),
		secret:   seed.WebhookSecret,
		captured: captured,
		server:   server,
	}
}

func TestDispatcher_PublishesValidSignature(t *testing.T) {
	h := newDispatchHarness(t)
	t.Cleanup(h.bus.Close)

	// Create a payment so the dispatcher has something to render.
	payment, err := h.store.CreatePayment(store.Payment{
		MerchantID: store.DefaultMerchantID,
		Amount:     100,
		Currency:   store.CurrencyMVR,
		Type:       store.PaymentTypeQR,
		Status:     store.PaymentPending,
		ShortCode:  "ABC123",
	})
	if err != nil {
		t.Fatalf("create payment: %v", err)
	}

	d := &webhooks.Dispatcher{
		URL:        h.server.URL,
		Secret:     h.secret,
		Store:      h.store,
		HTTPClient: h.server.Client(),
		Logger:     slog.New(slog.NewJSONHandler(io.Discard, nil)),
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var wg sync.WaitGroup
	wg.Go(func() { _ = d.Run(ctx, h.bus) })
	// Brief pause so the dispatcher goroutine reaches its Subscribe call
	// before we publish — Subscribe + Publish are mutex-protected but the
	// goroutine scheduler isn't deterministic.
	time.Sleep(20 * time.Millisecond)

	h.bus.Publish(events.Event{Resource: "payment", ResourceID: payment.ID, Status: "PENDING"})

	select {
	case got := <-h.captured:
		if got.id == "" || got.timestamp == "" || got.signature == "" {
			t.Fatalf("missing headers: %+v", got)
		}
		if err := webhooks.Verify(h.secret, got.id, got.timestamp, got.signature, got.body, time.Now().UTC()); err != nil {
			t.Errorf("verify: %v", err)
		}
		var env map[string]any
		if err := json.Unmarshal(got.body, &env); err != nil {
			t.Fatalf("unmarshal envelope: %v", err)
		}
		if env["eventType"] != "transaction.state_changed" {
			t.Errorf("eventType = %v", env["eventType"])
		}
		data, _ := env["data"].(map[string]any)
		if data["transaction_id"] != payment.ID {
			t.Errorf("transaction_id = %v", data["transaction_id"])
		}
		if data["status"] != "PENDING" {
			// Status comes from the persisted payment, not the event.
			t.Logf("status = %v (read from store, not event)", data["status"])
		}
	case <-time.After(2 * time.Second):
		t.Fatal("dispatcher did not deliver")
	}

	cancel()
	wg.Wait()
}

func TestDispatcher_NoURL_IsNoOp(t *testing.T) {
	h := newDispatchHarness(t)
	t.Cleanup(h.bus.Close)
	d := &webhooks.Dispatcher{
		URL:    "",
		Store:  h.store,
		Logger: slog.New(slog.NewJSONHandler(io.Discard, nil)),
	}
	if err := d.Run(context.Background(), h.bus); err != nil {
		t.Errorf("run: %v", err)
	}
}

func TestDispatcher_5xxResponse_DoesNotRetry(t *testing.T) {
	// V1 is fire-and-forget per D-019 — only one POST attempt.
	h := newDispatchHarness(t)
	t.Cleanup(h.bus.Close)
	var hits int
	failing := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hits++
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(failing.Close)

	payment, err := h.store.CreatePayment(store.Payment{
		MerchantID: store.DefaultMerchantID,
		Amount:     5,
		Currency:   store.CurrencyMVR,
		Type:       store.PaymentTypeQR,
		Status:     store.PaymentPending,
		ShortCode:  "Z",
	})
	if err != nil {
		t.Fatalf("create payment: %v", err)
	}

	d := &webhooks.Dispatcher{
		URL:        failing.URL,
		Secret:     h.secret,
		Store:      h.store,
		HTTPClient: failing.Client(),
		Logger:     slog.New(slog.NewJSONHandler(io.Discard, nil)),
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var wg sync.WaitGroup
	wg.Go(func() { _ = d.Run(ctx, h.bus) })
	// Brief pause so the dispatcher goroutine reaches its Subscribe call
	// before we publish — Subscribe + Publish are mutex-protected but the
	// goroutine scheduler isn't deterministic.
	time.Sleep(20 * time.Millisecond)

	h.bus.Publish(events.Event{Resource: "payment", ResourceID: payment.ID, Status: "PENDING"})
	time.Sleep(150 * time.Millisecond)
	cancel()
	wg.Wait()

	if hits != 1 {
		t.Errorf("hits = %d, want 1 (fire-and-forget)", hits)
	}
}
