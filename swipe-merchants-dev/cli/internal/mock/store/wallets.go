package store

import (
	"encoding/json"
	"fmt"

	bolt "go.etcd.io/bbolt"
)

// PutWallet inserts or replaces a wallet, keyed by its merchant id.
func (s *Store) PutWallet(w Wallet) error {
	if w.MerchantID == "" {
		return fmt.Errorf("put wallet: empty merchant id")
	}
	return s.db.Update(func(tx *bolt.Tx) error {
		buf, _ := json.Marshal(w) // plain struct — Marshal cannot fail
		return tx.Bucket(bucketWallets).Put([]byte(w.MerchantID), buf)
	})
}

// GetWallet returns the wallet for a merchant, or ErrNotFound.
func (s *Store) GetWallet(merchantID string) (Wallet, error) {
	var w Wallet
	err := s.db.View(func(tx *bolt.Tx) error {
		buf := tx.Bucket(bucketWallets).Get([]byte(merchantID))
		if buf == nil {
			return ErrNotFound
		}
		if err := json.Unmarshal(buf, &w); err != nil {
			return fmt.Errorf("unmarshal wallet for %s: %w", merchantID, err)
		}
		return nil
	})
	if err != nil {
		return Wallet{}, err
	}
	return w, nil
}
