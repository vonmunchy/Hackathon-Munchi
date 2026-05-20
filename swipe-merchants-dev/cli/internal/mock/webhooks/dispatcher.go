package webhooks

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/oklog/ulid/v2"

	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/events"
	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/store"
)

// EventType is the literal value of the `eventType` field in the spec's
// WebhookEvent envelope. Phase 4 only emits transaction.state_changed.
const EventType = "transaction.state_changed"

// dispatchTimeout caps the per-request POST. Fire-and-forget per D-019;
// we just want to fail fast on a hung receiver.
const dispatchTimeout = 10 * time.Second

// Dispatcher consumes events from the bus and POSTs WebhookEvent payloads
// to the configured URL. V1 is fire-and-forget per D-019.
type Dispatcher struct {
	// URL is the target endpoint. Empty disables dispatch.
	URL string
	// Secret is the merchant's webhook signing secret.
	Secret string
	// Store provides the merchant/wallet context for transactions referenced
	// by ID in the event.
	Store *store.Store
	// HTTPClient overrides the default *http.Client (tests inject one
	// pointing at httptest.NewServer).
	HTTPClient *http.Client
	// Logger receives delivery summaries.
	Logger *slog.Logger
	// ShouldFail, when non-nil and returning true, makes the next
	// delivery short-circuit with a synthetic error (no actual POST).
	// Used by the webhook_delivery_fail scenario.
	ShouldFail func(ctx context.Context) bool
}

// Run subscribes to the bus and dispatches webhook events until ctx is
// cancelled. The returned error is always nil — V1 delivery is
// fire-and-forget and the worker only stops on context cancellation.
func (d *Dispatcher) Run(ctx context.Context, bus *events.Bus) error {
	if d.URL == "" {
		return nil
	}
	ch, cancel := bus.Subscribe(events.TopicAll)
	defer cancel()
	client := d.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: dispatchTimeout}
	}
	for {
		select {
		case <-ctx.Done():
			return nil
		case ev, open := <-ch:
			if !open {
				return nil
			}
			if ev.Resource != "payment" {
				// Phase 5 only dispatches transaction.state_changed events.
				// Payout transitions are out of scope until the spec adds them.
				continue
			}
			if err := d.deliver(ctx, client, ev); err != nil {
				d.Logger.Warn("webhook delivery failed",
					slog.String("resource_id", ev.ResourceID),
					slog.String("err", err.Error()))
			}
		}
	}
}

// deliver builds, signs, and POSTs one webhook envelope. Errors are
// returned to Run for logging; nothing else acts on them (D-019).
func (d *Dispatcher) deliver(ctx context.Context, client *http.Client, ev events.Event) error {
	if d.ShouldFail != nil && d.ShouldFail(ctx) {
		return fmt.Errorf("scenario webhook_delivery_fail: synthetic failure")
	}
	body, err := d.payloadFor(ev)
	if err != nil {
		return fmt.Errorf("payload: %w", err)
	}
	id := newMessageID()
	ts := strconv.FormatInt(time.Now().UTC().Unix(), 10)
	sig := Sign(d.Secret, id, ts, body)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, d.URL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Webhook-Event-Type", EventType)
	req.Header.Set(HeaderID, id)
	req.Header.Set(HeaderTimestamp, ts)
	req.Header.Set(HeaderSignature, sig)

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("send: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	d.Logger.Info("webhook delivered",
		slog.String("id", id),
		slog.String("resource_id", ev.ResourceID),
		slog.String("status", ev.Status),
		slog.Int("http_status", resp.StatusCode))

	if resp.StatusCode >= 400 {
		return fmt.Errorf("non-2xx: %d", resp.StatusCode)
	}
	return nil
}

// payloadFor constructs the WebhookEvent envelope referenced in the spec.
// Spec required fields on WebhookData: transaction_id, transaction_code,
// wallet_id, status, amount, currency. We fill those plus a few optional
// fields the platform exposes.
func (d *Dispatcher) payloadFor(ev events.Event) ([]byte, error) {
	p, err := d.Store.GetPayment(ev.ResourceID)
	if err != nil {
		return nil, fmt.Errorf("load payment %s: %w", ev.ResourceID, err)
	}
	now := time.Now().UTC().Format(time.RFC3339)
	data := map[string]any{
		"transaction_id":   p.ID,
		"transaction_code": p.ShortCode,
		"wallet_id":        "wal_" + p.MerchantID,
		"status":           string(p.Status),
		"amount":           p.Amount,
		"currency":         string(p.Currency),
		"type":             "P2M",
		"entry_type":       "CREDIT",
		"created_at":       p.CreatedAt.UTC().Format(time.RFC3339),
		"updated_at":       p.UpdatedAt.UTC().Format(time.RFC3339),
		"post_date":        now,
	}
	if p.RecipientVPA != "" {
		data["recipient_vpa"] = p.RecipientVPA
	}
	envelope := map[string]any{
		"eventType": EventType,
		"data":      data,
	}
	return json.Marshal(envelope)
}

// newMessageID returns a fresh msg_<ulid> identifier per scaffold §5.7.
func newMessageID() string {
	id, err := ulid.New(ulid.Timestamp(time.Now()), rand.Reader)
	if err != nil {
		return "msg_unknown"
	}
	return "msg_" + id.String()
}
