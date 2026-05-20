package handlers

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/auth"
	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/problemdetails"
	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/store"
)

// PaymentGet serves GET /api/v1/payments/{paymentId}. Looks up the payment
// in the store and verifies merchant ownership before returning.
type PaymentGet struct {
	Store *store.Store
}

// ServeHTTP implements http.Handler.
func (h *PaymentGet) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
		problemdetails.WriteNew(w, http.StatusNotFound, problemdetails.TypeNotFound, "payment %s not found", id)
		return
	}
	writeJSON(w, http.StatusOK, PaymentResponse{
		ID:         payment.ID,
		Amount:     payment.Amount,
		Currency:   string(payment.Currency),
		Status:     string(payment.Status),
		Reference:  payment.ShortCode,
		ShortCode:  payment.ShortCode,
		QRData:     payment.QRData,
		PaymentURL: payment.PaymentURL,
		CreatedAt:  payment.CreatedAt,
	})
}
