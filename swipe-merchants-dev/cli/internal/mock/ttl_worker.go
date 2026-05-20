package mock

import (
	"context"
	"log/slog"
	"time"

	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/store"
)

// ttlWorkerInterval is how often the payment TTL scan runs. The mock
// transitions PENDING payments past their TransitionAt; the resolution
// is coarse on purpose (developers don't need sub-second precision).
const ttlWorkerInterval = 250 * time.Millisecond

// runPaymentTTLWorker scans the payments bucket on a tick and transitions
// PENDING payments whose TransitionAt is in the past via the Server's
// shared transitionPayment method (same code path the customer-facing
// pay page uses).
//
// The worker exits cleanly when ctx is cancelled.
func (s *Server) runPaymentTTLWorker(ctx context.Context) {
	ticker := time.NewTicker(ttlWorkerInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.ttlTickOnce()
		}
	}
}

// ttlTickOnce performs a single scan + transition pass. Extracted so
// tests can drive the work without running the goroutine.
func (s *Server) ttlTickOnce() {
	pending, err := s.cfg.Store.ListPendingPayments()
	if err != nil {
		s.logger.Warn("ttl: list pending", slog.String("err", err.Error()))
		return
	}
	now := time.Now().UTC()
	for _, p := range pending {
		if p.TransitionAt.IsZero() || !now.After(p.TransitionAt) {
			continue
		}
		target := p.TransitionTo
		if target == "" {
			target = store.PaymentCompleted
		}
		if err := s.transitionPayment(p.ID, target); err != nil {
			s.logger.Warn("ttl: transition", slog.String("err", err.Error()), slog.String("id", p.ID))
			continue
		}
		s.logger.Info("payment transitioned",
			slog.String("id", p.ID),
			slog.String("from", string(store.PaymentPending)),
			slog.String("to", string(target)))
	}
}
