package store

import (
	"encoding/json"
	"errors"
	"fmt"

	bolt "go.etcd.io/bbolt"
)

// ErrInsufficientFunds signals a withdrawal exceeding the available
// balance for the requested currency. Mapped to the spec's
// INSUFFICIENT_FUNDS problem type.
var ErrInsufficientFunds = errors.New("store: insufficient funds")

// DebitWallet atomically subtracts amount from the merchant's available
// balance in the given currency. Returns ErrInsufficientFunds when the
// balance is too low, ErrNotFound when the wallet does not exist.
func (s *Store) DebitWallet(merchantID string, currency Currency, amount float64) error {
	if amount <= 0 {
		return fmt.Errorf("debit wallet: amount must be positive")
	}
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketWallets)
		buf := b.Get([]byte(merchantID))
		if buf == nil {
			return ErrNotFound
		}
		var w Wallet
		if err := json.Unmarshal(buf, &w); err != nil {
			return fmt.Errorf("unmarshal wallet %s: %w", merchantID, err)
		}
		bal, ok := w.Balances[currency]
		if !ok {
			return fmt.Errorf("wallet %s: no balance entry for %s", merchantID, currency)
		}
		if bal.Available < amount {
			return ErrInsufficientFunds
		}
		bal.Available -= amount
		w.Balances[currency] = bal
		w.UpdatedAt = s.nowUTC()
		next, _ := json.Marshal(w)
		return b.Put([]byte(merchantID), next)
	})
}
