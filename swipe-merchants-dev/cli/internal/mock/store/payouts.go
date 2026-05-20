package store

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"sort"

	bolt "go.etcd.io/bbolt"
)

// CreatePayout persists a fresh Payout record. The caller supplies the
// monetary fields + bank_account_id; the store fills in ID + timestamps.
func (s *Store) CreatePayout(p Payout) (Payout, error) {
	if p.MerchantID == "" {
		return Payout{}, fmt.Errorf("create payout: merchant id required")
	}
	if p.Amount <= 0 {
		return Payout{}, fmt.Errorf("create payout: amount must be positive")
	}
	if p.BankAccountID == "" {
		return Payout{}, fmt.Errorf("create payout: bank account id required")
	}
	now := s.nowUTC()
	id, err := newULID(now, "pyo_", rand.Reader)
	if err != nil {
		return Payout{}, fmt.Errorf("create payout: %w", err)
	}
	p.ID = id
	p.CreatedAt = now
	p.UpdatedAt = now
	if p.Status == "" {
		p.Status = PayoutPending
	}
	err = s.db.Update(func(tx *bolt.Tx) error {
		buf, _ := json.Marshal(p)
		return tx.Bucket(bucketPayouts).Put([]byte(p.ID), buf)
	})
	if err != nil {
		return Payout{}, err
	}
	return p, nil
}

// GetPayout returns the payout with the given id or ErrNotFound.
func (s *Store) GetPayout(id string) (Payout, error) {
	var out Payout
	err := s.db.View(func(tx *bolt.Tx) error {
		buf := tx.Bucket(bucketPayouts).Get([]byte(id))
		if buf == nil {
			return ErrNotFound
		}
		if err := json.Unmarshal(buf, &out); err != nil {
			return fmt.Errorf("unmarshal payout %s: %w", id, err)
		}
		return nil
	})
	if err != nil {
		return Payout{}, err
	}
	return out, nil
}

// ListPayoutsForMerchant returns every payout owned by merchantID.
func (s *Store) ListPayoutsForMerchant(merchantID string) ([]Payout, error) {
	var out []Payout
	err := s.db.View(func(tx *bolt.Tx) error {
		return tx.Bucket(bucketPayouts).ForEach(func(_, v []byte) error {
			var p Payout
			if err := json.Unmarshal(v, &p); err != nil {
				return fmt.Errorf("unmarshal payout: %w", err)
			}
			if p.MerchantID == merchantID {
				out = append(out, p)
			}
			return nil
		})
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}
