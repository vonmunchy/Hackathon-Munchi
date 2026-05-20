package store_test

import (
	"path/filepath"
	"testing"

	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/store"
)

func TestSeedDefaults_FreshStore_PopulatesDefaults(t *testing.T) {
	t.Parallel()
	dbPath := filepath.Join(t.TempDir(), "state.db")
	st, err := store.Open(dbPath)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })

	summary, err := st.SeedDefaults()
	if err != nil {
		t.Fatalf("seed: %v", err)
	}

	if !summary.FreshSeed {
		t.Errorf("expected FreshSeed=true on first seed, got false")
	}
	if summary.Merchant.ID != store.DefaultMerchantID {
		t.Errorf("merchant id = %q, want %q", summary.Merchant.ID, store.DefaultMerchantID)
	}
	if got := summary.Wallet.Balances[store.CurrencyMVR].Available; got != 10000.00 {
		t.Errorf("MVR available = %f, want 10000.00", got)
	}
	if got := summary.Wallet.Balances[store.CurrencyUSD].Available; got != 650.00 {
		t.Errorf("USD available = %f, want 650.00", got)
	}
	if len(summary.BankAccounts) != 2 {
		t.Fatalf("bank accounts: got %d, want 2", len(summary.BankAccounts))
	}
	currencies := map[store.Currency]bool{}
	for _, ba := range summary.BankAccounts {
		currencies[ba.Currency] = true
		if ba.Status != store.BankAccountActive {
			t.Errorf("bank %s status = %s, want ACTIVE", ba.ID, ba.Status)
		}
	}
	if !currencies[store.CurrencyMVR] || !currencies[store.CurrencyUSD] {
		t.Errorf("expected MVR + USD bank accounts; got %v", currencies)
	}
	if summary.WebhookSecret == "" {
		t.Error("webhook secret was empty")
	}
}

func TestSeedDefaults_Idempotent_PreservesState(t *testing.T) {
	t.Parallel()
	dbPath := filepath.Join(t.TempDir(), "state.db")
	st, err := store.Open(dbPath)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })

	first, err := st.SeedDefaults()
	if err != nil {
		t.Fatalf("first seed: %v", err)
	}
	second, err := st.SeedDefaults()
	if err != nil {
		t.Fatalf("second seed: %v", err)
	}
	if second.FreshSeed {
		t.Error("expected FreshSeed=false on second seed")
	}
	if first.WebhookSecret != second.WebhookSecret {
		t.Errorf("webhook secret rotated unexpectedly: %q vs %q", first.WebhookSecret, second.WebhookSecret)
	}
	firstIDs := map[string]bool{}
	for _, ba := range first.BankAccounts {
		firstIDs[ba.ID] = true
	}
	for _, ba := range second.BankAccounts {
		if !firstIDs[ba.ID] {
			t.Errorf("bank account %q appeared in second seed but not first (IDs rotated)", ba.ID)
		}
	}
	if len(first.BankAccounts) != len(second.BankAccounts) {
		t.Errorf("bank account count changed: %d vs %d", len(first.BankAccounts), len(second.BankAccounts))
	}
}

func TestStore_PersistsAcrossOpen(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "state.db")

	st, err := store.Open(dbPath)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	first, err := st.SeedDefaults()
	if err != nil {
		t.Fatalf("seed: %v", err)
	}
	if err := st.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}

	// Reopen — analogous to `mock stop && mock start`.
	st2, err := store.Open(dbPath)
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	t.Cleanup(func() { _ = st2.Close() })

	got, err := st2.GetMerchant(store.DefaultMerchantID)
	if err != nil {
		t.Fatalf("get merchant after reopen: %v", err)
	}
	if got.ID != first.Merchant.ID {
		t.Errorf("merchant lost across reopen: %q vs %q", got.ID, first.Merchant.ID)
	}
	wallet, err := st2.GetWallet(store.DefaultMerchantID)
	if err != nil {
		t.Fatalf("get wallet: %v", err)
	}
	if wallet.Balances[store.CurrencyMVR].Available != 10000.00 {
		t.Errorf("MVR balance lost across reopen: %v", wallet.Balances[store.CurrencyMVR])
	}
	accounts, err := st2.ListBankAccounts(store.DefaultMerchantID)
	if err != nil {
		t.Fatalf("list bank accounts: %v", err)
	}
	if len(accounts) != 2 {
		t.Errorf("bank accounts lost: got %d, want 2", len(accounts))
	}
}

func TestStore_GetMerchant_Missing_ReturnsErrNotFound(t *testing.T) {
	t.Parallel()
	dbPath := filepath.Join(t.TempDir(), "state.db")
	st, err := store.Open(dbPath)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })

	_, err = st.GetMerchant("mer_missing")
	if err == nil {
		t.Fatalf("expected ErrNotFound, got nil")
	}
}

func TestStore_PutMerchant_RejectsEmptyID(t *testing.T) {
	t.Parallel()
	st, err := store.Open(filepath.Join(t.TempDir(), "state.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })

	if err := st.PutMerchant(store.Merchant{Name: "x"}); err == nil {
		t.Errorf("expected error for empty id; got nil")
	}
}

func TestStore_ListMerchants_AfterSeed_ContainsDefault(t *testing.T) {
	t.Parallel()
	st, err := store.Open(filepath.Join(t.TempDir(), "state.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })
	if _, err := st.SeedDefaults(); err != nil {
		t.Fatalf("seed: %v", err)
	}

	merchants, err := st.ListMerchants()
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(merchants) != 1 || merchants[0].ID != store.DefaultMerchantID {
		t.Errorf("merchants = %+v, want only mer_default", merchants)
	}
}

func TestStore_ListBankAccounts_FiltersByMerchant(t *testing.T) {
	t.Parallel()
	st, err := store.Open(filepath.Join(t.TempDir(), "state.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })
	if _, err := st.SeedDefaults(); err != nil {
		t.Fatalf("seed: %v", err)
	}

	got, err := st.ListBankAccounts(store.DefaultMerchantID)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("default merchant accounts = %d, want 2", len(got))
	}
	other, err := st.ListBankAccounts("mer_unknown")
	if err != nil {
		t.Fatalf("list other: %v", err)
	}
	if len(other) != 0 {
		t.Errorf("unknown merchant got %d accounts, want 0", len(other))
	}
}

func TestStore_Meta_RoundTrip(t *testing.T) {
	t.Parallel()
	dbPath := filepath.Join(t.TempDir(), "state.db")
	st, err := store.Open(dbPath)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })

	if err := st.PutMeta("foo", "bar"); err != nil {
		t.Fatalf("put: %v", err)
	}
	got, err := st.GetMeta("foo")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got != "bar" {
		t.Errorf("meta foo = %q, want %q", got, "bar")
	}
}
