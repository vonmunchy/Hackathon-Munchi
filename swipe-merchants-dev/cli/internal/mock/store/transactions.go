package store

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"sort"

	bolt "go.etcd.io/bbolt"
)

// CreateTransaction persists a fresh Transaction record (the history-side
// projection of a payment, payout, or adjustment). The caller supplies
// every field except ID/CreatedAt/UpdatedAt.
func (s *Store) CreateTransaction(t Transaction) (Transaction, error) {
	if t.MerchantID == "" {
		return Transaction{}, fmt.Errorf("create transaction: merchant id required")
	}
	now := s.nowUTC()
	id, err := newULID(now, "txn_", rand.Reader)
	if err != nil {
		return Transaction{}, fmt.Errorf("create transaction: %w", err)
	}
	t.ID = id
	t.CreatedAt = now
	t.UpdatedAt = now
	err = s.db.Update(func(tx *bolt.Tx) error {
		buf, _ := json.Marshal(t)
		return tx.Bucket(bucketTransactions).Put([]byte(t.ID), buf)
	})
	if err != nil {
		return Transaction{}, err
	}
	return t, nil
}

// UpdateTransactionStatus updates the status (and UpdatedAt) of the
// transaction with the given id. Returns ErrNotFound when missing. Other
// fields are left untouched.
func (s *Store) UpdateTransactionStatus(id, status string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		buf := tx.Bucket(bucketTransactions).Get([]byte(id))
		if buf == nil {
			return ErrNotFound
		}
		var t Transaction
		if err := json.Unmarshal(buf, &t); err != nil {
			return fmt.Errorf("unmarshal transaction %s: %w", id, err)
		}
		t.Status = status
		t.UpdatedAt = s.nowUTC()
		next, _ := json.Marshal(t)
		return tx.Bucket(bucketTransactions).Put([]byte(t.ID), next)
	})
}

// ListTransactionsForMerchant returns transactions owned by merchantID
// newest-first, paginated by limit/offset.
//
// Only rows that represent a settled movement of money — payment-type
// rows with Status=COMPLETED, plus payouts (which are settled on
// creation, see payout_create.go) — are returned. Payment-intent rows
// in PENDING/EXPIRED/CANCELLED/FAILED states stay in the store (so the
// transaction id is stable across the lifecycle) but are hidden from
// the public API surface.
func (s *Store) ListTransactionsForMerchant(merchantID string, limit, offset int) ([]Transaction, int, error) {
	var all []Transaction
	err := s.db.View(func(tx *bolt.Tx) error {
		return tx.Bucket(bucketTransactions).ForEach(func(_, v []byte) error {
			var t Transaction
			if err := json.Unmarshal(v, &t); err != nil {
				return fmt.Errorf("unmarshal transaction: %w", err)
			}
			if t.MerchantID != merchantID {
				return nil
			}
			if !isVisibleTransaction(t) {
				return nil
			}
			all = append(all, t)
			return nil
		})
	})
	if err != nil {
		return nil, 0, err
	}
	sort.Slice(all, func(i, j int) bool { return all[i].ID > all[j].ID }) // newest first
	total := len(all)
	if offset >= total {
		return []Transaction{}, total, nil
	}
	if limit <= 0 {
		limit = total - offset
	}
	end := offset + limit
	if end > total {
		end = total
	}
	return all[offset:end], total, nil
}

// FindTransactionByReference looks up a transaction by its `Reference`
// field (the merchant-facing transaction code or payout id, as set by
// the createPayment / createPayout handlers). Used by `GET
// /api/v1/transactions/{reference}` (operationId getTransactionStatus).
//
// Returns ErrNotFound when no transaction matches in `merchantID`'s set.
// merchantID gates the lookup so one merchant cannot inspect another's
// transactions via this endpoint.
func (s *Store) FindTransactionByReference(merchantID, reference string) (Transaction, error) {
	var out Transaction
	var found bool
	err := s.db.View(func(tx *bolt.Tx) error {
		return tx.Bucket(bucketTransactions).ForEach(func(_, v []byte) error {
			var t Transaction
			if err := json.Unmarshal(v, &t); err != nil {
				return fmt.Errorf("unmarshal transaction: %w", err)
			}
			if t.MerchantID != merchantID || t.Reference != reference {
				return nil
			}
			if !isVisibleTransaction(t) {
				return nil
			}
			out = t
			found = true
			return nil
		})
	})
	if err != nil {
		return Transaction{}, err
	}
	if !found {
		return Transaction{}, ErrNotFound
	}
	return out, nil
}

// isVisibleTransaction is the read-side filter that hides
// payment-intent rows in non-settled states. Non-PAYMENT rows (today:
// payouts) are always visible because they represent an immediate debit
// on the wallet. PAYMENT rows are visible only when COMPLETED. The
// filter lives next to the read paths so adding a new transaction type
// or a new payment status only requires one place to update.
func isVisibleTransaction(t Transaction) bool {
	if t.Type == "PAYMENT" {
		return t.Status == string(PaymentCompleted)
	}
	return true
}

// FindTransactionBySource returns the transaction record whose source_id
// matches the supplied pay_/pyo_ id, plus ok=false if none exists.
func (s *Store) FindTransactionBySource(sourceID string) (Transaction, bool, error) {
	var out Transaction
	var ok bool
	err := s.db.View(func(tx *bolt.Tx) error {
		return tx.Bucket(bucketTransactions).ForEach(func(_, v []byte) error {
			var t Transaction
			if err := json.Unmarshal(v, &t); err != nil {
				return fmt.Errorf("unmarshal transaction: %w", err)
			}
			if t.SourceID == sourceID {
				out = t
				ok = true
			}
			return nil
		})
	})
	if err != nil {
		return Transaction{}, false, err
	}
	return out, ok, nil
}
