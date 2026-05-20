package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/auth"
	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/problemdetails"
	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/store"
)

// HistoryResponse mirrors spec/app.yaml#HistoryResponse.
type HistoryResponse struct {
	Transactions []TransactionItem `json:"transactions"`
	Total        int               `json:"total"`
}

// TransactionItem mirrors spec/app.yaml#TransactionItem. The four
// monetary fields beyond `amount` were added in spec v1.1.0 (see
// docs/.scratched/notes/spec-audit-2026-05-18.md). The mock has no
// fee or FX model, so gross/net/original default to `amount` and
// fee defaults to 0.
type TransactionItem struct {
	ID             string  `json:"id"`
	Reference      string  `json:"reference,omitempty"`
	Amount         float64 `json:"amount"`
	GrossAmount    float64 `json:"gross_amount,omitempty"`
	FeeAmount      float64 `json:"fee_amount,omitempty"`
	NetAmount      float64 `json:"net_amount,omitempty"`
	OriginalAmount float64 `json:"original_amount,omitempty"`
	Currency       string  `json:"currency"`
	Type           string  `json:"type,omitempty"`
	Status         string  `json:"status"`
	Description    string  `json:"description,omitempty"`
	CreatedAt      string  `json:"created_at,omitempty"`
}

// HistoryDefaultLimit is the default page size when limit is omitted, per
// the OpenAPI spec parameter default.
const HistoryDefaultLimit = 20

// History serves GET /api/v1/history. Returns transactions for the
// authenticated merchant ordered newest-first with limit/offset paging.
type History struct {
	Store *store.Store
}

// ServeHTTP implements http.Handler.
func (h *History) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		problemdetails.WriteNew(w, http.StatusUnauthorized, problemdetails.TypeUnauthorized, "missing principal")
		return
	}
	pg, err := parsePagination(r)
	if err != nil {
		problemdetails.WriteNew(w, http.StatusBadRequest, problemdetails.TypeValidationError, "%v", err)
		return
	}
	txns, total, err := h.Store.ListTransactionsForMerchant(p.MerchantID, pg.limit, pg.offset)
	if err != nil {
		problemdetails.WriteNew(w, http.StatusInternalServerError, problemdetails.TypeValidationError, "list transactions: %v", err)
		return
	}
	items := make([]TransactionItem, 0, len(txns))
	for _, t := range txns {
		items = append(items, TransactionItem{
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
	out := HistoryResponse{Transactions: items, Total: total}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(out)
}

// pagination captures parsed limit/offset values for the history endpoint.
type pagination struct {
	limit  int
	offset int
}

func parsePagination(r *http.Request) (pagination, error) {
	q := r.URL.Query()
	p := pagination{limit: HistoryDefaultLimit, offset: 0}
	if raw := q.Get("limit"); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n < 0 {
			return pagination{}, &paginationError{field: "limit", reason: "must be a non-negative integer"}
		}
		p.limit = n
	}
	if raw := q.Get("offset"); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n < 0 {
			return pagination{}, &paginationError{field: "offset", reason: "must be a non-negative integer"}
		}
		p.offset = n
	}
	return p, nil
}

type paginationError struct {
	field, reason string
}

func (e *paginationError) Error() string {
	return e.field + ": " + e.reason
}
