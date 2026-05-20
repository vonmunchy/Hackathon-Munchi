package handlers_test

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/handlers"
)

// payPageGET fetches /pay/{code} using a no-redirect client so 303s
// from action handlers can be observed directly. Returns the response
// for the caller to inspect; the body is consumed before return so the
// caller doesn't need to worry about it.
func payPageGET(t *testing.T, baseURL, code string) *http.Response {
	t.Helper()
	client := &http.Client{
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}
	resp, err := client.Get(baseURL + "/pay/" + code)
	if err != nil {
		t.Fatalf("get pay page: %v", err)
	}
	t.Cleanup(func() { _ = resp.Body.Close() })
	return resp
}

func payPagePOST(t *testing.T, baseURL, code, action string) *http.Response {
	t.Helper()
	client := &http.Client{
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}
	resp, err := client.PostForm(baseURL+"/pay/"+code+"/"+action, url.Values{})
	if err != nil {
		t.Fatalf("post pay action: %v", err)
	}
	t.Cleanup(func() { _ = resp.Body.Close() })
	return resp
}

func TestPayPage_GET_PendingPayment_RendersForm(t *testing.T) {
	tc, token := phase4Server(t, []string{"payments:qr"}, time.Minute)
	pay, err := tc.CreatePayment(context.Background(), token, handlers.CreatePaymentRequest{
		Amount: 9.99, Currency: "MVR", Type: "QR",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	baseURL := strings.TrimSuffix(tc.BaseURL, "/")
	resp := payPageGET(t, baseURL, pay.Reference)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); !strings.HasPrefix(ct, "text/html") {
		t.Errorf("content-type = %q, want text/html", ct)
	}
}

func TestPayPage_GET_UnknownShortCode_Returns404(t *testing.T) {
	tc, _ := phase4Server(t, []string{"payments:qr"}, time.Minute)
	baseURL := strings.TrimSuffix(tc.BaseURL, "/")
	resp := payPageGET(t, baseURL, "NOSUCH00")
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("status = %d, want 404", resp.StatusCode)
	}
}

func TestPayPage_Complete_TransitionsAndRedirects(t *testing.T) {
	// Long TTL so the TTL worker doesn't race the manual transition.
	tc, token := phase4Server(t, []string{"payments:qr", "transactions:status"}, time.Minute)
	pay, err := tc.CreatePayment(context.Background(), token, handlers.CreatePaymentRequest{
		Amount: 10, Currency: "MVR", Type: "QR",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	baseURL := strings.TrimSuffix(tc.BaseURL, "/")
	resp := payPagePOST(t, baseURL, pay.Reference, "complete")
	if resp.StatusCode != http.StatusSeeOther {
		t.Fatalf("status = %d, want 303", resp.StatusCode)
	}
	if loc := resp.Header.Get("Location"); !strings.HasSuffix(loc, "/pay/"+pay.Reference) {
		t.Errorf("Location = %q, want suffix /pay/%s", loc, pay.Reference)
	}
	// The transition should also make the transaction lookup succeed.
	got, err := tc.GetTransaction(context.Background(), token, pay.Reference)
	if err != nil {
		t.Fatalf("get transaction after complete: %v", err)
	}
	if got.Status != "COMPLETED" {
		t.Errorf("status = %q, want COMPLETED", got.Status)
	}
}

func TestPayPage_Cancel_DoesNotSurfaceInHistory(t *testing.T) {
	tc, token := phase4Server(t, []string{"payments:qr", "transactions:history"}, time.Minute)
	pay, err := tc.CreatePayment(context.Background(), token, handlers.CreatePaymentRequest{
		Amount: 5, Currency: "MVR", Type: "QR",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	baseURL := strings.TrimSuffix(tc.BaseURL, "/")
	resp := payPagePOST(t, baseURL, pay.Reference, "cancel")
	if resp.StatusCode != http.StatusSeeOther {
		t.Fatalf("cancel status = %d, want 303", resp.StatusCode)
	}
	// CANCELLED payments must stay hidden from /history (filter-on-read).
	hist, err := tc.History(context.Background(), token, 0, 0)
	if err != nil {
		t.Fatalf("history: %v", err)
	}
	if hist.Total != 0 {
		t.Errorf("history total = %d, want 0; cancelled payment leaked", hist.Total)
	}
}

func TestPayPage_ActionOnTerminalPayment_Returns409(t *testing.T) {
	tc, token := phase4Server(t, []string{"payments:qr"}, time.Minute)
	pay, err := tc.CreatePayment(context.Background(), token, handlers.CreatePaymentRequest{
		Amount: 5, Currency: "MVR", Type: "QR",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	baseURL := strings.TrimSuffix(tc.BaseURL, "/")
	if r := payPagePOST(t, baseURL, pay.Reference, "complete"); r.StatusCode != http.StatusSeeOther {
		t.Fatalf("first complete = %d", r.StatusCode)
	}
	r := payPagePOST(t, baseURL, pay.Reference, "complete")
	if r.StatusCode != http.StatusConflict {
		t.Errorf("second complete = %d, want 409", r.StatusCode)
	}
}
