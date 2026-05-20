package handlers

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"time"

	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/auth"
	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/events"
	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/problemdetails"
	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/store"
)

// CreatePaymentRequest mirrors spec/app.yaml#CreatePaymentRequest.
type CreatePaymentRequest struct {
	Amount       float64 `json:"amount"`
	Currency     string  `json:"currency"`
	Type         string  `json:"type"`
	Description  string  `json:"description,omitempty"`
	RecipientVPA string  `json:"recipient_vpa,omitempty"`
}

// PaymentResponse mirrors spec/app.yaml#PaymentResponse.
type PaymentResponse struct {
	ID         string    `json:"id"`
	Amount     float64   `json:"amount"`
	Currency   string    `json:"currency"`
	Status     string    `json:"status"`
	Reference  string    `json:"reference,omitempty"`
	ShortCode  string    `json:"short_code,omitempty"`
	QRData     string    `json:"qr_data,omitempty"`
	PaymentURL string    `json:"payment_url,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

// PaymentTransitionTTL is the default time a PENDING payment stays open
// before auto-transitioning to COMPLETED. Per scaffold §13 + DECISIONS.md
// open-questions table, 60s is the placeholder default pending platform
// confirmation. The `--payment-ttl` flag on `swipe mock start` would
// override (not yet exposed).
const PaymentTransitionTTL = 60 * time.Second

// CreatePayment serves POST /api/v1/payments. It enforces D-013's
// per-type scope check on top of the coarse auth pre-check, validates
// the request body, persists Payment + Transaction, and publishes a
// PENDING event so SSE subscribers see the new payment.
type CreatePayment struct {
	Store *store.Store
	Bus   *events.Bus
	// TTL overrides PaymentTransitionTTL when non-zero (test hook).
	TTL time.Duration
	// IssuerURL is used to build payment_url for type=LINK responses.
	IssuerURL string
	// QRFormat selects URL vs EMVCo content for QR-type payments.
	// Empty falls back to DefaultQRFormat.
	QRFormat QRFormat
	// InsufficientFunds, when non-nil and returning true, short-circuits
	// the request with 402 INSUFFICIENT_FUNDS. Used by the
	// insufficient_funds scenario.
	InsufficientFunds func(ctx context.Context) bool
	// TransitionAtOverride, when non-nil, can override the TransitionAt
	// stamped on a newly-created PENDING payment. Used by
	// payment_stuck_pending.
	TransitionAtOverride func(ctx context.Context, defaultAt time.Time) (time.Time, bool)
	// TransitionToOverride, when non-nil, can override the target status
	// of the auto-transition. Used by payment_transitions.
	TransitionToOverride func(ctx context.Context, defaultStatus string) (status string, delay time.Duration, fired bool)
}

// ServeHTTP implements http.Handler.
func (h *CreatePayment) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		problemdetails.WriteNew(w, http.StatusUnauthorized, problemdetails.TypeUnauthorized, "missing principal")
		return
	}
	var req CreatePaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		problemdetails.WriteNew(w, http.StatusBadRequest, problemdetails.TypeValidationError, "decode body: %v", err)
		return
	}
	if pd, badReq := validatePaymentRequest(req); badReq {
		problemdetails.Write(w, http.StatusBadRequest, pd)
		return
	}
	requiredScope, ok := scopeForType(req.Type)
	if !ok {
		problemdetails.WriteNew(w, http.StatusBadRequest, problemdetails.TypeValidationError, "unsupported payment type %q", req.Type)
		return
	}
	if !p.HasScope(requiredScope) {
		problemdetails.WriteNew(w, http.StatusForbidden, problemdetails.TypeForbidden, "Token missing required scope: %s", requiredScope)
		return
	}
	if h.InsufficientFunds != nil && h.InsufficientFunds(r.Context()) {
		problemdetails.WriteNew(w, http.StatusPaymentRequired, problemdetails.TypeInsufficientFunds, "insufficient funds (scenario)")
		return
	}

	ttl := h.TTL
	if ttl <= 0 {
		ttl = PaymentTransitionTTL
	}

	merchantName := merchantNameFor(h.Store, p.MerchantID)
	qrFormat := h.QRFormat
	if qrFormat == "" {
		qrFormat = DefaultQRFormat
	}
	shortCode, qrData, paymentURL, err := buildPaymentArtifacts(
		store.PaymentType(req.Type),
		h.IssuerURL,
		qrFormat,
		req.Amount,
		store.Currency(req.Currency),
		merchantName,
	)
	if err != nil {
		problemdetails.WriteNew(w, http.StatusInternalServerError, problemdetails.TypeValidationError, "build artifacts: %v", err)
		return
	}

	now := time.Now().UTC()
	transitionAt := now.Add(ttl)
	transitionTo := store.PaymentCompleted
	if h.TransitionToOverride != nil {
		if target, delay, fired := h.TransitionToOverride(r.Context(), string(transitionTo)); fired {
			transitionTo = store.PaymentStatus(target)
			if delay > 0 {
				transitionAt = now.Add(delay)
			}
		}
	}
	if h.TransitionAtOverride != nil {
		if overridden, fired := h.TransitionAtOverride(r.Context(), transitionAt); fired {
			transitionAt = overridden
		}
	}
	payment := store.Payment{
		MerchantID:   p.MerchantID,
		Amount:       req.Amount,
		Currency:     store.Currency(req.Currency),
		Type:         store.PaymentType(req.Type),
		Status:       store.PaymentPending,
		ShortCode:    shortCode,
		QRData:       qrData,
		PaymentURL:   paymentURL,
		RecipientVPA: req.RecipientVPA,
		Description:  req.Description,
		TransitionAt: transitionAt,
		TransitionTo: transitionTo,
	}
	persisted, err := h.Store.CreatePayment(payment)
	if err != nil {
		problemdetails.WriteNew(w, http.StatusInternalServerError, problemdetails.TypeValidationError, "create payment: %v", err)
		return
	}
	// A Transaction row is persisted immediately, mirroring the payment's
	// PENDING status. It is intentionally hidden from `GET /api/v1/history`
	// and `GET /api/v1/transactions/{reference}` until the payment moves
	// to COMPLETED — see ListTransactionsForMerchant + FindTransactionByReference,
	// which filter to settled rows. This keeps the transaction id stable
	// across the payment lifecycle without exposing intent-stage rows.
	if _, err := h.Store.CreateTransaction(store.Transaction{
		MerchantID:     persisted.MerchantID,
		Reference:      persisted.ShortCode,
		Amount:         persisted.Amount,
		GrossAmount:    persisted.Amount,
		FeeAmount:      0,
		NetAmount:      persisted.Amount,
		OriginalAmount: persisted.Amount,
		Currency:       persisted.Currency,
		Type:           "PAYMENT",
		Status:         string(persisted.Status),
		Description:    persisted.Description,
		SourceID:       persisted.ID,
	}); err != nil {
		problemdetails.WriteNew(w, http.StatusInternalServerError, problemdetails.TypeValidationError, "create transaction: %v", err)
		return
	}
	if h.Bus != nil {
		h.Bus.Publish(events.Event{
			Resource:   "payment",
			ResourceID: persisted.ID,
			Status:     string(persisted.Status),
			MerchantID: persisted.MerchantID,
			OccurredAt: persisted.CreatedAt,
		})
	}

	writeJSON(w, http.StatusCreated, PaymentResponse{
		ID:         persisted.ID,
		Amount:     persisted.Amount,
		Currency:   string(persisted.Currency),
		Status:     string(persisted.Status),
		Reference:  persisted.ShortCode,
		ShortCode:  persisted.ShortCode,
		QRData:     persisted.QRData,
		PaymentURL: persisted.PaymentURL,
		CreatedAt:  persisted.CreatedAt,
	})
}

// validatePaymentRequest enforces spec-level constraints not caught by
// the kin-openapi middleware (currency + type enums; CONTACT requires
// recipient_vpa).
func validatePaymentRequest(req CreatePaymentRequest) (problemdetails.ProblemDetails, bool) {
	pd := problemdetails.New(problemdetails.TypeValidationError, "request validation failed")
	bad := false
	if req.Amount <= 0 {
		pd = pd.WithFieldError("amount", "must be greater than 0")
		bad = true
	}
	if req.Currency == "" {
		req.Currency = "MVR"
	}
	switch store.Currency(req.Currency) {
	case store.CurrencyMVR, store.CurrencyUSD:
	default:
		pd = pd.WithFieldError("currency", "must be MVR or USD")
		bad = true
	}
	if req.Type == "" {
		req.Type = string(store.PaymentTypeQR)
	}
	switch store.PaymentType(req.Type) {
	case store.PaymentTypeQR, store.PaymentTypeContact, store.PaymentTypeLink:
	default:
		pd = pd.WithFieldError("type", "must be QR, CONTACT, or LINK")
		bad = true
	}
	if store.PaymentType(req.Type) == store.PaymentTypeContact && strings.TrimSpace(req.RecipientVPA) == "" {
		pd = pd.WithFieldError("recipient_vpa", "required for type=CONTACT")
		bad = true
	}
	return pd, bad
}

// merchantNameFor looks up the merchant's name for embedding into the
// EMV QR's tag-59 field. Falls back to a generic label if the merchant
// is somehow missing — the QR is still structurally valid in that case.
func merchantNameFor(st *store.Store, merchantID string) string {
	m, err := st.GetMerchant(merchantID)
	if err != nil || m.Name == "" {
		return "Merchant"
	}
	return m.Name
}

func scopeForType(t string) (string, bool) {
	switch store.PaymentType(t) {
	case store.PaymentTypeQR:
		return "payments:qr", true
	case store.PaymentTypeContact:
		return "payments:contact", true
	case store.PaymentTypeLink:
		return "payments:link", true
	default:
		return "", false
	}
}

// buildPaymentArtifacts returns the short_code + (qr_data, payment_url)
// triple depending on payment type.
//
// For QR payments, payment_url is always populated with the mock pay
// page URL — a developer-side fallback so the dev can always click
// through to the pay page even if their scanner can't read the QR. The
// QR's encoded content is either that same URL (qrFormat=URL, default)
// or an EMVCo MPM string with a valid CRC-16/CCITT-FALSE checksum
// (qrFormat=EMVCo, matches production behaviour).
func buildPaymentArtifacts(typ store.PaymentType, issuer string, qrFormat QRFormat, amount float64, currency store.Currency, merchantName string) (string, string, string, error) {
	code, err := randomShortCode(8)
	if err != nil {
		return "", "", "", err
	}
	payURL := buildPayURL(issuer, code)
	switch typ {
	case store.PaymentTypeQR:
		var qrContent string
		switch qrFormat {
		case QRFormatEMVCo:
			qrContent = buildEMVCoQR(amount, currency, code, merchantName)
		default: // JSON is the default; treat unknown values as JSON.
			qrContent = buildQRJSONPayload(payURL, code, amount, currency, merchantName)
		}
		png, err := renderQRPNG(qrContent)
		if err != nil {
			return "", "", "", err
		}
		return code, base64.StdEncoding.EncodeToString(png), payURL, nil
	case store.PaymentTypeLink:
		return code, "", payURL, nil
	default: // CONTACT
		return code, "", "", nil
	}
}

// buildPayURL returns the URL of the mock pay page for the given short
// code, derived from the mock's issuer URL. Mirrors the URL embedded
// inside the JSON QR so the two surfaces stay in lockstep.
func buildPayURL(issuer, shortCode string) string {
	base := strings.TrimRight(issuer, "/")
	if base == "" {
		base = "http://localhost:8080"
	}
	return base + "/pay/" + shortCode
}

// qrJSONPayload is what the JSON-format QR encodes. It carries the
// mock pay-page URL plus the payment essentials a custom scanner needs
// to display "Pay 42.50 MVR to <merchant>" without an HTTP roundtrip.
// Keep the field set small — large payloads bloat the QR's data
// density and reduce scanning reliability.
type qrJSONPayload struct {
	URL       string  `json:"url"`
	Reference string  `json:"reference"`
	Amount    float64 `json:"amount"`
	Currency  string  `json:"currency"`
	Merchant  string  `json:"merchant"`
}

// buildQRJSONPayload renders the JSON content the mock pay QR encodes
// when --qr-format=json (the default). json.Marshal of this fixed
// struct cannot fail in practice — every field is a string or float —
// so the function does not surface an error.
func buildQRJSONPayload(payURL, shortCode string, amount float64, currency store.Currency, merchantName string) string {
	b, _ := json.Marshal(qrJSONPayload{
		URL:       payURL,
		Reference: shortCode,
		Amount:    amount,
		Currency:  string(currency),
		Merchant:  merchantName,
	})
	return string(b)
}

// randomShortCode returns an uppercase alphanumeric short code of length n.
func randomShortCode(n int) (string, error) {
	const alphabet = "0123456789ABCDEFGHJKLMNPQRSTUVWXYZ" // Crockford-ish; no I/O
	out := make([]byte, n)
	limit := big.NewInt(int64(len(alphabet)))
	for i := range out {
		idx, err := rand.Int(rand.Reader, limit)
		if err != nil {
			return "", fmt.Errorf("randomShortCode: %w", err)
		}
		out[i] = alphabet[idx.Int64()]
	}
	return string(out), nil
}
