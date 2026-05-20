package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"slices"

	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/auth"
	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/problemdetails"
	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/store"
)

// BalanceResponse mirrors spec/app.yaml#BalanceResponse.
type BalanceResponse struct {
	AvailableBalance float64 `json:"available_balance"`
	PendingBalance   float64 `json:"pending_balance"`
	Currency         string  `json:"currency"`
}

// Balance serves GET /api/v1/balance. Per spec it returns an array, one
// entry per supported currency. The principal must have wallet:balance.
type Balance struct {
	Store *store.Store
}

// ServeHTTP implements http.Handler.
func (h *Balance) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		problemdetails.WriteNew(w, http.StatusUnauthorized, problemdetails.TypeUnauthorized, "missing principal")
		return
	}
	wallet, err := h.Store.GetWallet(p.MerchantID)
	switch {
	case errors.Is(err, store.ErrNotFound):
		problemdetails.WriteNew(w, http.StatusNotFound, problemdetails.TypeNotFound, "no wallet for merchant %s", p.MerchantID)
		return
	case err != nil:
		problemdetails.WriteNew(w, http.StatusInternalServerError, problemdetails.TypeInvalidRequest, "load wallet: %v", err)
		return
	}

	currencies := make([]store.Currency, 0, len(wallet.Balances))
	for c := range wallet.Balances {
		currencies = append(currencies, c)
	}
	slices.Sort(currencies)

	out := make([]BalanceResponse, 0, len(currencies))
	for _, c := range currencies {
		bal := wallet.Balances[c]
		out = append(out, BalanceResponse{
			AvailableBalance: bal.Available,
			PendingBalance:   bal.Pending,
			Currency:         string(c),
		})
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(out)
}
