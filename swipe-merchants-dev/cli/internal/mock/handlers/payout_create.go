package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/auth"
	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/events"
	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/problemdetails"
	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/store"
)

// CreatePayoutRequest mirrors spec/app.yaml#CreatePayoutRequest.
type CreatePayoutRequest struct {
	Amount        float64 `json:"amount"`
	BankAccountID string  `json:"bank_account_id"`
}

// PayoutResponse mirrors spec/app.yaml#PayoutResponse with the extra
// fields the mock emits to keep ledger correlation simple.
type PayoutResponse struct {
	ID        string    `json:"id"`
	Amount    float64   `json:"amount"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	Reference string    `json:"reference,omitempty"`
}

// CreatePayout serves POST /api/v1/payouts. The principal must have
// wallet:withdraw (enforced by the coarse scope middleware) and the
// supplied bank account must belong to the same merchant. Successful
// requests debit the wallet for the requested currency and persist
// Payout + Transaction records.
type CreatePayout struct {
	Store *store.Store
	Bus   *events.Bus
}

// ServeHTTP implements http.Handler.
func (h *CreatePayout) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		problemdetails.WriteNew(w, http.StatusUnauthorized, problemdetails.TypeUnauthorized, "missing principal")
		return
	}
	var req CreatePayoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		problemdetails.WriteNew(w, http.StatusBadRequest, problemdetails.TypeValidationError, "decode body: %v", err)
		return
	}
	pd := problemdetails.New(problemdetails.TypeValidationError, "request validation failed")
	bad := false
	if req.Amount <= 0 {
		pd = pd.WithFieldError("amount", "must be greater than 0")
		bad = true
	}
	if strings.TrimSpace(req.BankAccountID) == "" {
		pd = pd.WithFieldError("bank_account_id", "required")
		bad = true
	}
	if bad {
		problemdetails.Write(w, http.StatusBadRequest, pd)
		return
	}

	bank, err := h.Store.GetBankAccount(req.BankAccountID)
	switch {
	case errors.Is(err, store.ErrNotFound):
		problemdetails.WriteNew(w, http.StatusBadRequest, problemdetails.TypeValidationError, "bank account %s not found", req.BankAccountID)
		return
	case err != nil:
		problemdetails.WriteNew(w, http.StatusInternalServerError, problemdetails.TypeValidationError, "load bank account: %v", err)
		return
	}
	if bank.MerchantID != p.MerchantID {
		problemdetails.WriteNew(w, http.StatusForbidden, problemdetails.TypeForbidden, "bank account %s does not belong to your merchant", req.BankAccountID)
		return
	}
	if bank.Status != store.BankAccountActive {
		problemdetails.WriteNew(w, http.StatusBadRequest, problemdetails.TypeValidationError, "bank account %s is %s", bank.ID, bank.Status)
		return
	}

	if err := h.Store.DebitWallet(p.MerchantID, bank.Currency, req.Amount); err != nil {
		switch {
		case errors.Is(err, store.ErrInsufficientFunds):
			problemdetails.WriteNew(w, http.StatusPaymentRequired, problemdetails.TypeInsufficientFunds, "insufficient %s balance", bank.Currency)
			return
		case errors.Is(err, store.ErrNotFound):
			problemdetails.WriteNew(w, http.StatusBadRequest, problemdetails.TypeValidationError, "no wallet for merchant %s", p.MerchantID)
			return
		default:
			problemdetails.WriteNew(w, http.StatusInternalServerError, problemdetails.TypeValidationError, "debit wallet: %v", err)
			return
		}
	}

	payout, err := h.Store.CreatePayout(store.Payout{
		MerchantID:    p.MerchantID,
		Amount:        req.Amount,
		Currency:      bank.Currency,
		BankAccountID: bank.ID,
		Status:        store.PayoutPending,
	})
	if err != nil {
		problemdetails.WriteNew(w, http.StatusInternalServerError, problemdetails.TypeValidationError, "create payout: %v", err)
		return
	}
	if _, err := h.Store.CreateTransaction(store.Transaction{
		MerchantID:     payout.MerchantID,
		Reference:      payout.ID,
		Amount:         payout.Amount,
		GrossAmount:    payout.Amount,
		FeeAmount:      0,
		NetAmount:      payout.Amount,
		OriginalAmount: payout.Amount,
		Currency:       payout.Currency,
		Type:           "PAYOUT",
		Status:         string(payout.Status),
		SourceID:       payout.ID,
	}); err != nil {
		problemdetails.WriteNew(w, http.StatusInternalServerError, problemdetails.TypeValidationError, "create transaction: %v", err)
		return
	}
	if h.Bus != nil {
		h.Bus.Publish(events.Event{
			Resource:   "payout",
			ResourceID: payout.ID,
			Status:     string(payout.Status),
			MerchantID: payout.MerchantID,
			OccurredAt: payout.CreatedAt,
		})
	}

	// Reference is the payout id — that's what we use as the transaction
	// reference (see CreateTransaction above with Reference: payout.ID),
	// so `getTransactionStatus` can find this payout by the same value.
	writeJSON(w, http.StatusCreated, PayoutResponse{
		ID:        payout.ID,
		Amount:    payout.Amount,
		Status:    string(payout.Status),
		CreatedAt: payout.CreatedAt,
		Reference: payout.ID,
	})
}
