package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/auth"
	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/events"
	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/problemdetails"
	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/store"
)

// streamHeartbeatInterval is the cadence of SSE keep-alive comments. 15s
// matches the scaffold §5.6 guidance and keeps middleboxes from closing
// idle streams.
const streamHeartbeatInterval = 15 * time.Second

// StreamPayment serves GET /api/v1/payments/{paymentId}/stream. It is an
// SSE handler: subscribes to the event bus topic for the payment id,
// emits one event per transition, sends a `: heartbeat` comment every
// 15s, and closes the stream on terminal state.
type StreamPayment struct {
	Store *store.Store
	Bus   *events.Bus
}

// streamEvent is the JSON payload of one SSE event.
type streamEvent struct {
	ID        string    `json:"id"`
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}

// ServeHTTP implements http.Handler.
func (h *StreamPayment) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		problemdetails.WriteNew(w, http.StatusUnauthorized, problemdetails.TypeUnauthorized, "missing principal")
		return
	}
	id := chi.URLParam(r, "paymentId")
	payment, err := h.Store.GetPayment(id)
	switch {
	case errors.Is(err, store.ErrNotFound):
		problemdetails.WriteNew(w, http.StatusNotFound, problemdetails.TypeNotFound, "payment %s not found", id)
		return
	case err != nil:
		problemdetails.WriteNew(w, http.StatusInternalServerError, problemdetails.TypeValidationError, "load payment: %v", err)
		return
	}
	if payment.MerchantID != p.MerchantID {
		problemdetails.WriteNew(w, http.StatusForbidden, problemdetails.TypeForbidden, "payment %s does not belong to your merchant", id)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	ch, cancel := h.Bus.Subscribe(payment.ID)
	defer cancel()

	// Emit the current state immediately so clients that subscribe AFTER
	// a transition still see the latest status before the stream blocks.
	writeStreamEvent(w, flusher, streamEvent{
		ID:        payment.ID,
		Status:    string(payment.Status),
		Timestamp: payment.UpdatedAt,
	})
	if isTerminalStatus(payment.Status) {
		return
	}

	heartbeat := time.NewTicker(streamHeartbeatInterval)
	defer heartbeat.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case <-heartbeat.C:
			_, _ = fmt.Fprintf(w, ": heartbeat\n\n")
			flusher.Flush()
		case event, open := <-ch:
			if !open {
				return
			}
			writeStreamEvent(w, flusher, streamEvent{
				ID:        event.ResourceID,
				Status:    event.Status,
				Timestamp: event.OccurredAt,
			})
			if isTerminalStatus(store.PaymentStatus(event.Status)) {
				return
			}
		}
	}
}

func isTerminalStatus(s store.PaymentStatus) bool {
	switch s {
	case store.PaymentCompleted, store.PaymentExpired, store.PaymentCancelled:
		return true
	default:
		return false
	}
}

func writeStreamEvent(w http.ResponseWriter, flusher http.Flusher, e streamEvent) {
	body, _ := json.Marshal(e)
	_, _ = fmt.Fprintf(w, "event: status\ndata: %s\n\n", body)
	flusher.Flush()
}
