package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/auth"
	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/problemdetails"
)

// WhoAmIResponse mirrors spec/app.yaml#WhoAmIResponse.
type WhoAmIResponse struct {
	ClientID   string   `json:"client_id"`
	MerchantID string   `json:"merchant_id"`
	Scopes     []string `json:"scopes,omitempty"`
}

// WhoAmI implements GET /api/v1/whoami. It expects the auth middleware to
// have populated a Principal on the request context; without one it
// returns an UNAUTHORIZED ProblemDetails (the only path that can land here
// without a principal is a misconfigured route).
type WhoAmI struct{}

// ServeHTTP implements http.Handler.
func (h *WhoAmI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		problemdetails.WriteNew(w, http.StatusUnauthorized, problemdetails.TypeUnauthorized, "Missing or invalid authentication token")
		return
	}
	resp := WhoAmIResponse{
		ClientID:   p.ClientID,
		MerchantID: p.MerchantID,
		Scopes:     append([]string(nil), p.Scopes...),
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}
