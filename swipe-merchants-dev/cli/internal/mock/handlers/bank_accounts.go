package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/auth"
	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/problemdetails"
	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/store"
)

// BankAccountResponse mirrors spec/app.yaml#BankAccountResponse.
type BankAccountResponse struct {
	ID                string `json:"id"`
	AccountNumber     string `json:"account_number"`
	AccountHolderName string `json:"account_holder_name"`
	Currency          string `json:"currency"`
	Status            string `json:"status"`
}

// BankAccounts serves GET /api/v1/bank-accounts. The principal must have
// wallet:balance per the spec's security stanza. The response is the set
// of bank accounts linked to the principal's merchant.
type BankAccounts struct {
	Store *store.Store
}

// ServeHTTP implements http.Handler.
func (h *BankAccounts) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		problemdetails.WriteNew(w, http.StatusUnauthorized, problemdetails.TypeUnauthorized, "missing principal")
		return
	}
	accounts, err := h.Store.ListBankAccounts(p.MerchantID)
	if err != nil {
		problemdetails.WriteNew(w, http.StatusInternalServerError, problemdetails.TypeInvalidRequest, "list bank accounts: %v", err)
		return
	}
	out := make([]BankAccountResponse, 0, len(accounts))
	for _, a := range accounts {
		out = append(out, BankAccountResponse{
			ID:                a.ID,
			AccountNumber:     a.AccountNumber,
			AccountHolderName: a.AccountHolderName,
			Currency:          string(a.Currency),
			Status:            string(a.Status),
		})
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(out)
}
