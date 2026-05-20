package handlers

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/auth"
	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/problemdetails"
	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/store"
)

// TransactionGet serves GET /api/v1/transactions/{reference}
// (operationId getTransactionStatus). The reference is the
// merchant-facing transaction code returned as `reference` on payment
// and payout responses. Lookup is scoped to the principal's merchant.
type TransactionGet struct {
	Store *store.Store
}

// ServeHTTP implements http.Handler.
func (h *TransactionGet) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		problemdetails.WriteNew(w, http.StatusUnauthorized, problemdetails.TypeUnauthorized, "missing principal")
		return
	}
	reference := chi.URLParam(r, "reference")
	t, err := h.Store.FindTransactionByReference(p.MerchantID, reference)
	switch {
	case errors.Is(err, store.ErrNotFound):
		problemdetails.WriteNew(w, http.StatusNotFound, problemdetails.TypeNotFound, "transaction %s not found", reference)
		return
	case err != nil:
		problemdetails.WriteNew(w, http.StatusInternalServerError, problemdetails.TypeValidationError, "load transaction: %v", err)
		return
	}
	writeJSON(w, http.StatusOK, TransactionItem{
		ID:             t.ID,
		Reference:      t.Reference,
		Amount:         t.Amount,
		GrossAmount:    t.GrossAmount,
		FeeAmount:      t.FeeAmount,
		NetAmount:      t.NetAmount,
		OriginalAmount: t.OriginalAmount,
		Currency:       string(t.Currency),
		Type:           t.Type,
		Status:         t.Status,
		Description:    t.Description,
		CreatedAt:      t.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}
