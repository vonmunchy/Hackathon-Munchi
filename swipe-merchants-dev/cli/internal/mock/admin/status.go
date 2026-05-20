package admin

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/store"
	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/version"
)

// StatusResponse is the shape returned by GET /_admin/status. It is the
// source of truth that the CLI's `swipe mock status` formats.
type StatusResponse struct {
	Version       version.Info `json:"version"`
	StartedAt     time.Time    `json:"started_at"`
	UptimeSeconds int64        `json:"uptime_seconds"`
	ListenAddress string       `json:"listen_address"`
	WebhookURL    string       `json:"webhook_url,omitempty"`
	WebhookSecret string       `json:"webhook_secret"`
	Counts        Counts       `json:"counts"`
}

// Counts surfaces the per-bucket cardinality the CLI uses for the status
// summary table. Phase 1 only reports merchants and bank accounts; later
// phases extend this struct as they add more state.
type Counts struct {
	Merchants    int `json:"merchants"`
	BankAccounts int `json:"bank_accounts"`
	Wallets      int `json:"wallets"`
	Clients      int `json:"clients"`
	Payments     int `json:"payments"`
	Payouts      int `json:"payouts"`
}

// StatusHandler builds GET /_admin/status responses from in-memory state
// (start time + listen address) and the BoltDB store (counts + secret).
type StatusHandler struct {
	StartedAt     time.Time
	ListenAddress string
	WebhookURL    string
	Store         *store.Store
}

// ServeHTTP implements http.Handler. It is unauthenticated; the localhost
// guard upstream is the access control.
func (h *StatusHandler) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	resp := StatusResponse{
		Version:       version.Get(),
		StartedAt:     h.StartedAt.UTC(),
		UptimeSeconds: int64(time.Since(h.StartedAt).Seconds()),
		ListenAddress: h.ListenAddress,
		WebhookURL:    h.WebhookURL,
	}
	if h.Store != nil {
		secret, _ := h.Store.GetMeta(store.MetaKeyWebhookSecret)
		resp.WebhookSecret = secret
		resp.Counts = h.computeCounts()
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *StatusHandler) computeCounts() Counts {
	c := Counts{}
	if merchants, err := h.Store.ListMerchants(); err == nil {
		c.Merchants = len(merchants)
		for _, m := range merchants {
			if accs, err := h.Store.ListBankAccounts(m.ID); err == nil {
				c.BankAccounts += len(accs)
			}
			if _, err := h.Store.GetWallet(m.ID); err == nil {
				c.Wallets++
			}
		}
	}
	return c
}
