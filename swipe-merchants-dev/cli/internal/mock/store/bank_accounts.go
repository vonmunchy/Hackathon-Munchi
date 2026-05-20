package store

import (
	"encoding/json"
	"fmt"

	bolt "go.etcd.io/bbolt"
)

// PutBankAccount inserts or replaces a bank account.
func (s *Store) PutBankAccount(b BankAccount) error {
	if b.ID == "" {
		return fmt.Errorf("put bank account: empty id")
	}
	if b.MerchantID == "" {
		return fmt.Errorf("put bank account %s: empty merchant id", b.ID)
	}
	return s.db.Update(func(tx *bolt.Tx) error {
		buf, _ := json.Marshal(b) // plain struct — Marshal cannot fail
		return tx.Bucket(bucketBankAccounts).Put([]byte(b.ID), buf)
	})
}

// GetBankAccount returns the bank account with the given id or ErrNotFound.
func (s *Store) GetBankAccount(id string) (BankAccount, error) {
	var out BankAccount
	err := s.db.View(func(tx *bolt.Tx) error {
		buf := tx.Bucket(bucketBankAccounts).Get([]byte(id))
		if buf == nil {
			return ErrNotFound
		}
		if err := json.Unmarshal(buf, &out); err != nil {
			return fmt.Errorf("unmarshal bank account %s: %w", id, err)
		}
		return nil
	})
	if err != nil {
		return BankAccount{}, err
	}
	return out, nil
}

// ListBankAccounts returns every bank account belonging to a merchant.
func (s *Store) ListBankAccounts(merchantID string) ([]BankAccount, error) {
	var out []BankAccount
	err := s.db.View(func(tx *bolt.Tx) error {
		return tx.Bucket(bucketBankAccounts).ForEach(func(_, v []byte) error {
			var b BankAccount
			if err := json.Unmarshal(v, &b); err != nil {
				return fmt.Errorf("unmarshal bank account: %w", err)
			}
			if b.MerchantID == merchantID {
				out = append(out, b)
			}
			return nil
		})
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}
