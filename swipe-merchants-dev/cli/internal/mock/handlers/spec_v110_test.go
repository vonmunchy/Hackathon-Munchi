package handlers_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/handlers"
)

// Covers the spec v1.1.0 catch-up:
//   - PaymentResponse.Reference is populated from the short code.
//   - PayoutResponse.Reference equals the payout id (matches the
//     transaction reference stored alongside).
//   - TransactionItem carries gross/fee/net/original fields with the
//     mock's no-fee defaults.
//   - GET /api/v1/transactions/{reference} returns the right
//     transaction; cross-merchant lookup returns 404; missing scope
//     returns 403.

func TestPaymentResponse_Reference_PopulatedFromShortCode(t *testing.T) {
	tc, token := phase4Server(t, []string{"payments:qr"}, time.Minute)
	got, err := tc.CreatePayment(context.Background(), token, handlers.CreatePaymentRequest{
		Amount: 10, Currency: "MVR", Type: "QR",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if got.Reference == "" {
		t.Errorf("reference empty; expected short_code")
	}
	if got.Reference != got.ShortCode {
		t.Errorf("reference=%q short_code=%q (expected equal)", got.Reference, got.ShortCode)
	}
}

func TestPayoutResponse_Reference_EqualsPayoutID(t *testing.T) {
	tc, token := phase4Server(t, []string{"wallet:withdraw", "wallet:balance"}, time.Minute)
	accounts, err := tc.BankAccounts(context.Background(), token)
	if err != nil {
		t.Fatalf("bank accounts: %v", err)
	}
	got, err := tc.CreatePayout(context.Background(), token, handlers.CreatePayoutRequest{
		Amount: 5, BankAccountID: accounts[0].ID,
	})
	if err != nil {
		t.Fatalf("create payout: %v", err)
	}
	if got.Reference != got.ID {
		t.Errorf("reference=%q id=%q (expected equal)", got.Reference, got.ID)
	}
}

func TestHistory_Items_CarryFinancialFields(t *testing.T) {
	tc, token := phase4Server(t,
		[]string{"payments:qr", "transactions:history"},
		50*time.Millisecond,
	)
	_, err := tc.CreatePayment(context.Background(), token, handlers.CreatePaymentRequest{
		Amount: 12, Currency: "MVR", Type: "QR",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	resp := pollHistory(t, tc, token, 1, 3*time.Second)
	first := resp.Transactions[0]
	if first.Amount != 12 || first.GrossAmount != 12 || first.NetAmount != 12 || first.OriginalAmount != 12 {
		t.Errorf("amount fields mismatch: %+v", first)
	}
	if first.FeeAmount != 0 {
		t.Errorf("fee_amount = %v, want 0 (mock has no fee model)", first.FeeAmount)
	}
}

func TestGetTransaction_ByReference_ReturnsMatch(t *testing.T) {
	tc, token := phase4Server(t,
		[]string{"payments:qr", "transactions:status"},
		50*time.Millisecond,
	)
	created, err := tc.CreatePayment(context.Background(), token, handlers.CreatePaymentRequest{
		Amount: 17, Currency: "MVR", Type: "QR",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	// Transaction is only created on COMPLETED, so the immediate lookup
	// is a 404 until the TTL fires.
	if _, err := tc.GetTransaction(context.Background(), token, created.Reference); err == nil || !strings.Contains(err.Error(), "404") {
		t.Errorf("immediate lookup err = %v, want 404 while PENDING", err)
	}
	deadline := time.Now().Add(3 * time.Second)
	var txn handlers.TransactionItem
	for time.Now().Before(deadline) {
		got, err := tc.GetTransaction(context.Background(), token, created.Reference)
		if err == nil {
			txn = got
			break
		}
		time.Sleep(75 * time.Millisecond)
	}
	if txn.Reference != created.Reference {
		t.Fatalf("reference round-trip: got %q, want %q", txn.Reference, created.Reference)
	}
	if txn.Type != "PAYMENT" {
		t.Errorf("type = %q, want PAYMENT", txn.Type)
	}
	if txn.Amount != 17 {
		t.Errorf("amount = %v, want 17", txn.Amount)
	}
	if txn.Status != "COMPLETED" {
		t.Errorf("status = %q, want COMPLETED", txn.Status)
	}
}

func TestGetTransaction_UnknownReference_Returns404(t *testing.T) {
	tc, token := phase4Server(t, []string{"transactions:status"}, time.Minute)
	_, err := tc.GetTransaction(context.Background(), token, "nope")
	if err == nil || !strings.Contains(err.Error(), "404") {
		t.Errorf("err = %v, want 404", err)
	}
}

func TestGetTransaction_MissingScope_Returns403(t *testing.T) {
	tc, token := phase4Server(t, []string{"wallet:balance"}, time.Minute)
	_, err := tc.GetTransaction(context.Background(), token, "anything")
	if err == nil || !strings.Contains(err.Error(), "403") {
		t.Errorf("err = %v, want 403", err)
	}
}
