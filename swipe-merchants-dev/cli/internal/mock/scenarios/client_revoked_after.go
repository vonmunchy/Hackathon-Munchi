package scenarios

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/store"
)

// ScenarioClientRevokedAfter is the scenario name.
const ScenarioClientRevokedAfter = "client_revoked_after"

func init() {
	register(Descriptor{
		Name:        ScenarioClientRevokedAfter,
		Description: "Revoke a specific client after the configured delay",
		Args: []ArgDef{
			{Name: "client", Type: ArgString, Description: "client_id to revoke"},
			{Name: "after", Type: ArgDuration, Default: "30s", Description: "delay from enable-time before revocation"},
		},
	})
}

// activeRevokes tracks scheduled revocations so re-enabling the scenario
// cancels any pending one rather than stacking them.
var activeRevokes struct {
	mu      sync.Mutex
	cancels map[string]context.CancelFunc
}

// ScheduleClientRevoke starts (or restarts) the revocation timer for the
// scenario's configured client. The timer fires once and disables the
// client by setting Enabled=false on its store record. Safe to call
// repeatedly: subsequent calls cancel any prior pending timer.
//
// Run inside its own goroutine — returns immediately.
func ScheduleClientRevoke(eng *Engine, st *store.Store, logger *slog.Logger) {
	if eng == nil || st == nil {
		return
	}
	args, ok := eng.IsEnabled(ScenarioClientRevokedAfter)
	if !ok {
		return
	}
	clientID := args.StringOr("client", "")
	if clientID == "" {
		return
	}
	delay, err := args.Duration("after", 30*time.Second)
	if err != nil || delay <= 0 {
		return
	}

	activeRevokes.mu.Lock()
	if activeRevokes.cancels == nil {
		activeRevokes.cancels = make(map[string]context.CancelFunc)
	}
	if prev, exists := activeRevokes.cancels[clientID]; exists {
		prev()
	}
	ctx, cancel := context.WithCancel(context.Background())
	activeRevokes.cancels[clientID] = cancel
	activeRevokes.mu.Unlock()

	go func() {
		timer := time.NewTimer(delay)
		defer timer.Stop()
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
		}
		_, err := st.UpdateClient(clientID, func(c store.Client) store.Client {
			c.Enabled = false
			return c
		})
		if err != nil {
			logger.Warn("client_revoked_after: update client", slog.String("id", clientID), slog.String("err", err.Error()))
			return
		}
		logger.Info("client_revoked_after: revoked", slog.String("id", clientID))
		activeRevokes.mu.Lock()
		delete(activeRevokes.cancels, clientID)
		activeRevokes.mu.Unlock()
	}()
}

// CancelClientRevokes cancels every pending revocation. Called when the
// scenario is disabled so re-enabling later starts fresh.
func CancelClientRevokes() {
	activeRevokes.mu.Lock()
	defer activeRevokes.mu.Unlock()
	for _, cancel := range activeRevokes.cancels {
		cancel()
	}
	activeRevokes.cancels = nil
}
