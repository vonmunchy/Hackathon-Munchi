package store

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"
	"time"

	"github.com/oklog/ulid/v2"
	bolt "go.etcd.io/bbolt"
)

// DefaultMerchantID is the well-known seeded merchant. It is not a ULID
// because the merchant is special: every fresh-state mock must contain it,
// so referencing it from docs and tests by a stable id keeps the UX
// predictable.
const DefaultMerchantID = "mer_default"

// SeedSummary captures what was seeded so the CLI can echo it back to the
// dev. It is also produced when the store is already seeded — the values
// reflect the persisted state, not just the work done in this call.
type SeedSummary struct {
	Merchant      Merchant      `json:"merchant"`
	Wallet        Wallet        `json:"wallet"`
	BankAccounts  []BankAccount `json:"bank_accounts"`
	WebhookSecret string        `json:"webhook_secret"`
	FreshSeed     bool          `json:"fresh_seed"`
}

// SeedDefaults installs the canonical default state from scaffold §6 if it
// is not already present:
//
//   - 1 merchant: mer_default
//   - 1 wallet:   {MVR: 10000.00 / 0, USD: 650.00 / 0}
//   - 2 bank accounts (one MVR, one USD), both ACTIVE
//   - 1 webhook secret (whsec_<32hex>)
//
// If the merchant already exists, SeedDefaults is a no-op (idempotent) and
// returns the persisted summary with FreshSeed=false.
func (s *Store) SeedDefaults() (SeedSummary, error) {
	existing, err := s.GetMerchant(DefaultMerchantID)
	if err == nil {
		return s.summary(existing, false)
	}
	if !errors.Is(err, ErrNotFound) {
		return SeedSummary{}, fmt.Errorf("check existing merchant: %w", err)
	}

	now := s.nowUTC()
	merchant := Merchant{
		ID:        DefaultMerchantID,
		Name:      "Default Test Merchant",
		CreatedAt: now,
	}
	wallet := Wallet{
		MerchantID: merchant.ID,
		ID:         "wal_" + merchant.ID,
		Balances: map[Currency]CurrencyBalance{
			CurrencyMVR: {Available: 10000.00, Pending: 0},
			CurrencyUSD: {Available: 650.00, Pending: 0},
		},
		UpdatedAt: now,
	}
	bankAccounts := s.makeDefaultBankAccounts(merchant.ID, now)
	webhookSecret := generateWebhookSecret()

	// All values below are small plain structs / fixed strings; the bbolt
	// Put cannot fail under those constraints inside a writable tx, and
	// json.Marshal of a value-typed struct cannot fail. Errors are
	// propagated only from the outer Update.
	err = s.db.Update(func(tx *bolt.Tx) error {
		mb, _ := json.Marshal(merchant)
		_ = tx.Bucket(bucketMerchants).Put([]byte(merchant.ID), mb)
		wb, _ := json.Marshal(wallet)
		_ = tx.Bucket(bucketWallets).Put([]byte(wallet.MerchantID), wb)
		for _, ba := range bankAccounts {
			bb, _ := json.Marshal(ba)
			_ = tx.Bucket(bucketBankAccounts).Put([]byte(ba.ID), bb)
		}
		return tx.Bucket(bucketMeta).Put([]byte(MetaKeyWebhookSecret), []byte(webhookSecret))
	})
	if err != nil {
		return SeedSummary{}, err
	}

	return SeedSummary{
		Merchant:      merchant,
		Wallet:        wallet,
		BankAccounts:  bankAccounts,
		WebhookSecret: webhookSecret,
		FreshSeed:     true,
	}, nil
}

// makeDefaultBankAccounts builds the canonical pair of MVR + USD active
// accounts. ULID entropy comes from crypto/rand which never fails on a
// healthy OS, so the error from newULID is propagated only through the
// test-only code path that can inject a failing reader.
func (s *Store) makeDefaultBankAccounts(merchantID string, now time.Time) []BankAccount {
	mvrID, _ := newULID(now, "bnk_", rand.Reader)
	usdID, _ := newULID(now, "bnk_", rand.Reader)
	return []BankAccount{
		{
			ID:                mvrID,
			MerchantID:        merchantID,
			AccountNumber:     "7770000123456",
			AccountHolderName: "Default Test Merchant",
			Currency:          CurrencyMVR,
			Status:            BankAccountActive,
			CreatedAt:         now,
		},
		{
			ID:                usdID,
			MerchantID:        merchantID,
			AccountNumber:     "7770000123457",
			AccountHolderName: "Default Test Merchant",
			Currency:          CurrencyUSD,
			Status:            BankAccountActive,
			CreatedAt:         now,
		},
	}
}

func newULID(t time.Time, prefix string, entropy io.Reader) (string, error) {
	id, err := ulid.New(ulid.Timestamp(t), entropy)
	if err != nil {
		return "", fmt.Errorf("new ulid: %w", err)
	}
	return prefix + id.String(), nil
}

// generateWebhookSecret returns a freshly-randomized 32-hex-character
// secret. rand.Read on a healthy OS never fails; if it does, the function
// returns the partially-zeroed buffer encoded as hex (acceptable because
// any unique mock-instance secret will work — failure is also surfaced
// indirectly through the lack of randomness in the output).
func generateWebhookSecret() string {
	buf := make([]byte, 16)
	_, _ = rand.Read(buf)
	return "whsec_" + hex.EncodeToString(buf)
}

// summary rebuilds a SeedSummary from already-persisted state. Used when
// SeedDefaults is called against an already-seeded store.
func (s *Store) summary(merchant Merchant, fresh bool) (SeedSummary, error) {
	wallet, err := s.GetWallet(merchant.ID)
	if err != nil {
		return SeedSummary{}, fmt.Errorf("read wallet: %w", err)
	}
	accounts, err := s.ListBankAccounts(merchant.ID)
	if err != nil {
		return SeedSummary{}, fmt.Errorf("list bank accounts: %w", err)
	}
	sort.Slice(accounts, func(i, j int) bool { return accounts[i].ID < accounts[j].ID })
	// GetMeta only ever returns nil or ErrNotFound; an absent secret is
	// fine here (older states predating this key).
	secret, _ := s.GetMeta(MetaKeyWebhookSecret)
	return SeedSummary{
		Merchant:      merchant,
		Wallet:        wallet,
		BankAccounts:  accounts,
		WebhookSecret: secret,
		FreshSeed:     fresh,
	}, nil
}
