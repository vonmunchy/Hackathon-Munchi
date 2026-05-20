package store

import (
	"encoding/json"
	"fmt"

	bolt "go.etcd.io/bbolt"
)

// PutMerchant inserts or replaces a merchant.
func (s *Store) PutMerchant(m Merchant) error {
	if m.ID == "" {
		return fmt.Errorf("put merchant: empty id")
	}
	return s.db.Update(func(tx *bolt.Tx) error {
		buf, _ := json.Marshal(m) // plain struct — Marshal cannot fail
		return tx.Bucket(bucketMerchants).Put([]byte(m.ID), buf)
	})
}

// GetMerchant returns the merchant with the given id, or ErrNotFound.
func (s *Store) GetMerchant(id string) (Merchant, error) {
	var m Merchant
	err := s.db.View(func(tx *bolt.Tx) error {
		buf := tx.Bucket(bucketMerchants).Get([]byte(id))
		if buf == nil {
			return ErrNotFound
		}
		if err := json.Unmarshal(buf, &m); err != nil {
			return fmt.Errorf("unmarshal merchant %s: %w", id, err)
		}
		return nil
	})
	if err != nil {
		return Merchant{}, err
	}
	return m, nil
}

// ListMerchants returns every merchant ordered by ULID (insertion-ordered).
func (s *Store) ListMerchants() ([]Merchant, error) {
	var out []Merchant
	err := s.db.View(func(tx *bolt.Tx) error {
		return tx.Bucket(bucketMerchants).ForEach(func(_, v []byte) error {
			var m Merchant
			if err := json.Unmarshal(v, &m); err != nil {
				return fmt.Errorf("unmarshal merchant: %w", err)
			}
			out = append(out, m)
			return nil
		})
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}
