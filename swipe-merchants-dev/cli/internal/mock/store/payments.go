package store

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"sort"

	bolt "go.etcd.io/bbolt"
)

// CreatePayment persists a fresh Payment record, generating its ULID and
// timestamps server-side. The caller supplies the wire fields (amount,
// currency, type, etc.) plus any short-code/QR/URL that depends on type.
//
// The returned Payment has ID set; the input p is otherwise echoed back
// with CreatedAt/UpdatedAt filled in.
func (s *Store) CreatePayment(p Payment) (Payment, error) {
	if p.MerchantID == "" {
		return Payment{}, fmt.Errorf("create payment: merchant id required")
	}
	if p.Amount <= 0 {
		return Payment{}, fmt.Errorf("create payment: amount must be positive")
	}
	now := s.nowUTC()
	id, err := newULID(now, "pay_", rand.Reader)
	if err != nil {
		return Payment{}, fmt.Errorf("create payment: %w", err)
	}
	p.ID = id
	p.CreatedAt = now
	p.UpdatedAt = now
	if p.Status == "" {
		p.Status = PaymentPending
	}
	if p.TransitionTo == "" {
		p.TransitionTo = PaymentCompleted
	}
	if err := s.putPayment(p); err != nil {
		return Payment{}, err
	}
	return p, nil
}

// GetPayment returns the payment with the given id, or ErrNotFound.
func (s *Store) GetPayment(id string) (Payment, error) {
	var out Payment
	err := s.db.View(func(tx *bolt.Tx) error {
		buf := tx.Bucket(bucketPayments).Get([]byte(id))
		if buf == nil {
			return ErrNotFound
		}
		if err := json.Unmarshal(buf, &out); err != nil {
			return fmt.Errorf("unmarshal payment %s: %w", id, err)
		}
		return nil
	})
	if err != nil {
		return Payment{}, err
	}
	return out, nil
}

// FindPaymentByShortCode looks up a payment by its merchant-facing
// short_code (a.k.a. reference). Short codes are unique per active
// payment but scoped to "the most recent match wins" if a developer
// ever resets state and gets a duplicate — the customer-facing pay
// page uses this to surface "your payment is here".
//
// Returns ErrNotFound when no payment matches.
func (s *Store) FindPaymentByShortCode(shortCode string) (Payment, error) {
	var out Payment
	var found bool
	err := s.db.View(func(tx *bolt.Tx) error {
		return tx.Bucket(bucketPayments).ForEach(func(_, v []byte) error {
			var p Payment
			if err := json.Unmarshal(v, &p); err != nil {
				return fmt.Errorf("unmarshal payment: %w", err)
			}
			if p.ShortCode == shortCode {
				// Newest-first via ID ordering: ULIDs are k-sortable so
				// the larger ID always wins. Keeps stale records from
				// resurfacing if state is partially cleared.
				if !found || p.ID > out.ID {
					out = p
					found = true
				}
			}
			return nil
		})
	})
	if err != nil {
		return Payment{}, err
	}
	if !found {
		return Payment{}, ErrNotFound
	}
	return out, nil
}

// UpdatePayment applies the supplied mutator inside a single write
// transaction. Returns ErrNotFound when the payment is absent.
func (s *Store) UpdatePayment(id string, mut func(Payment) Payment) (Payment, error) {
	var updated Payment
	err := s.db.Update(func(tx *bolt.Tx) error {
		buf := tx.Bucket(bucketPayments).Get([]byte(id))
		if buf == nil {
			return ErrNotFound
		}
		var p Payment
		if err := json.Unmarshal(buf, &p); err != nil {
			return fmt.Errorf("unmarshal payment %s: %w", id, err)
		}
		p = mut(p)
		p.UpdatedAt = s.nowUTC()
		newBuf, _ := json.Marshal(p)
		if err := tx.Bucket(bucketPayments).Put([]byte(p.ID), newBuf); err != nil {
			return fmt.Errorf("put payment %s: %w", id, err)
		}
		updated = p
		return nil
	})
	if err != nil {
		return Payment{}, err
	}
	return updated, nil
}

// ListPaymentsForMerchant returns every payment owned by merchantID,
// ordered by ULID (insertion order).
func (s *Store) ListPaymentsForMerchant(merchantID string) ([]Payment, error) {
	out, err := s.listPayments()
	if err != nil {
		return nil, err
	}
	filtered := out[:0]
	for _, p := range out {
		if p.MerchantID == merchantID {
			filtered = append(filtered, p)
		}
	}
	return filtered, nil
}

// ListPendingPayments returns every payment in PENDING status. Used by
// the TTL worker that drives the default PENDING -> COMPLETED transition.
func (s *Store) ListPendingPayments() ([]Payment, error) {
	all, err := s.listPayments()
	if err != nil {
		return nil, err
	}
	out := all[:0]
	for _, p := range all {
		if p.Status == PaymentPending {
			out = append(out, p)
		}
	}
	return out, nil
}

func (s *Store) listPayments() ([]Payment, error) {
	var out []Payment
	err := s.db.View(func(tx *bolt.Tx) error {
		return tx.Bucket(bucketPayments).ForEach(func(_, v []byte) error {
			var p Payment
			if err := json.Unmarshal(v, &p); err != nil {
				return fmt.Errorf("unmarshal payment: %w", err)
			}
			out = append(out, p)
			return nil
		})
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

func (s *Store) putPayment(p Payment) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		buf, _ := json.Marshal(p)
		return tx.Bucket(bucketPayments).Put([]byte(p.ID), buf)
	})
}
