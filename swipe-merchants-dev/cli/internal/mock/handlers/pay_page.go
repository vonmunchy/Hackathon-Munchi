// Pay-page handlers serve the mock's customer-facing payment URL.
// `GET /pay/{shortCode}` renders an HTML page; the POST siblings
// transition the payment lifecycle through the same code path the TTL
// worker uses, so a developer can click through complete/cancel/expire
// and observe webhooks + history populating in real time. Routes are
// public — a real customer hitting a payment link has no credentials.

package handlers

import (
	"context"
	_ "embed"
	"errors"
	"html/template"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/problemdetails"
	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/store"
)

//go:embed pay_page.html
var payPageTemplateSource string

// payPageTemplate is parsed once at process start. A parse failure is a
// programming error, so we panic on init — the template is checked in
// alongside the Go code and ships in the binary.
var payPageTemplate = template.Must(template.New("pay_page").Parse(payPageTemplateSource))

// TransitionFunc moves a payment to the target status, mirroring the
// status onto the existing transaction row and publishing an event on
// the bus. The signature matches the TTL worker's transition path so
// both invokers share one implementation. Implementations live in the
// mock package; this package only consumes them.
type TransitionFunc func(ctx context.Context, paymentID string, target store.PaymentStatus) error

// PayPage serves the mock pay page and its action POSTs.
type PayPage struct {
	Store      *store.Store
	Transition TransitionFunc
	// Logger captures partial-render failures and other non-fatal
	// surprises during page rendering. Required.
	Logger *slog.Logger
}

// Mount registers the four pay-page routes on r.
func (h *PayPage) Mount(r chi.Router) {
	r.Get("/pay/{shortCode}", h.render)
	r.Post("/pay/{shortCode}/complete", h.actionHandler(store.PaymentCompleted))
	r.Post("/pay/{shortCode}/cancel", h.actionHandler(store.PaymentCancelled))
	r.Post("/pay/{shortCode}/expire", h.actionHandler(store.PaymentExpired))
}

// payPageView is what the HTML template renders against.
type payPageView struct {
	MerchantName    string
	AmountFormatted string
	Currency        string
	Reference       string
	Description     string
	QRImage         string // already base64 PNG; consumed via <img data:image/png;base64,...>
	Status          string
	StatusClass     string
	IsPending       bool
}

func (h *PayPage) render(w http.ResponseWriter, r *http.Request) {
	shortCode := chi.URLParam(r, "shortCode")
	payment, err := h.Store.FindPaymentByShortCode(shortCode)
	if err != nil {
		writePayPageError(w, err, "payment not found")
		return
	}
	merchantName := merchantNameFor(h.Store, payment.MerchantID)
	view := payPageView{
		MerchantName:    merchantName,
		AmountFormatted: strconv.FormatFloat(payment.Amount, 'f', 2, 64),
		Currency:        string(payment.Currency),
		Reference:       payment.ShortCode,
		Description:     payment.Description,
		Status:          string(payment.Status),
		StatusClass:     strings.ToLower(string(payment.Status)),
		IsPending:       payment.Status == store.PaymentPending,
	}
	if payment.Type == store.PaymentTypeQR && payment.QRData != "" {
		view.QRImage = payment.QRData
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := payPageTemplate.Execute(w, view); err != nil {
		// Headers are already on the wire; nothing actionable beyond
		// surfacing it so a regression is visible in logs.
		h.Logger.Warn("pay page: template execute",
			slog.String("short_code", shortCode),
			slog.String("err", err.Error()),
		)
	}
}

func (h *PayPage) actionHandler(target store.PaymentStatus) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		shortCode := chi.URLParam(r, "shortCode")
		payment, err := h.Store.FindPaymentByShortCode(shortCode)
		if err != nil {
			writePayPageError(w, err, "payment not found")
			return
		}
		if payment.Status != store.PaymentPending {
			problemdetails.WriteNew(w, http.StatusConflict, problemdetails.TypeInvalidRequest,
				"payment %s is already %s; cannot transition to %s",
				payment.ShortCode, payment.Status, target)
			return
		}
		if h.Transition == nil {
			problemdetails.WriteNew(w, http.StatusInternalServerError, problemdetails.TypeInternalError,
				"pay page: transition handler not wired")
			return
		}
		if err := h.Transition(r.Context(), payment.ID, target); err != nil {
			problemdetails.WriteNew(w, http.StatusInternalServerError, problemdetails.TypeInternalError,
				"transition: %v", err)
			return
		}
		// 303 redirects POST → GET so refreshing doesn't re-post the
		// action — and the GET shows the new terminal state.
		http.Redirect(w, r, "/pay/"+payment.ShortCode, http.StatusSeeOther)
	}
}

// writePayPageError emits a 404 ProblemDetails for unknown short codes
// and a generic 500 otherwise. Internal payment ULIDs never leak — the
// detail string carries only short codes.
func writePayPageError(w http.ResponseWriter, err error, detail string) {
	if errors.Is(err, store.ErrNotFound) {
		problemdetails.WriteNew(w, http.StatusNotFound, problemdetails.TypeNotFound, "%s", detail)
		return
	}
	problemdetails.WriteNew(w, http.StatusInternalServerError, problemdetails.TypeInternalError, "lookup: %v", err)
}
