package store

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"time"

	bolt "go.etcd.io/bbolt"
)

// MaxRequestLogEntries is the soft cap on the request_log bucket. Older
// entries are evicted in FIFO order as new ones are appended, so the file
// stays bounded across long-lived mock sessions.
const MaxRequestLogEntries = 1000

// RequestLogEntry is the persisted shape of a single request/response pair
// captured by the mock middleware. The fields mirror the standard log keys
// from CLAUDE.md (req, method, path, status, duration_ms, client, merchant)
// plus a scenarios slot reserved for Phase 6.
type RequestLogEntry struct {
	ID           string    `json:"id"`
	StartedAt    time.Time `json:"started_at"`
	Method       string    `json:"method"`
	Path         string    `json:"path"`
	Status       int       `json:"status"`
	DurationMS   int64     `json:"duration_ms"`
	ClientID     string    `json:"client_id,omitempty"`
	MerchantID   string    `json:"merchant_id,omitempty"`
	Scenarios    []string  `json:"scenarios,omitempty"`
	RemoteAddr   string    `json:"remote_addr,omitempty"`
	UserAgent    string    `json:"user_agent,omitempty"`
	RequestBody  string    `json:"request_body,omitempty"`
	ResponseBody string    `json:"response_body,omitempty"`
}

// AppendRequestLog stores a new entry. If the bucket already holds
// MaxRequestLogEntries items the oldest (lowest ULID) entries are evicted
// in FIFO order to keep the file bounded.
//
// The entry's ID is overwritten with a fresh req_ ULID derived from
// StartedAt to keep keys sortable in arrival order.
func (s *Store) AppendRequestLog(entry RequestLogEntry) (RequestLogEntry, error) {
	if entry.StartedAt.IsZero() {
		entry.StartedAt = s.nowUTC()
	}
	id, err := newULID(entry.StartedAt, "req_", rand.Reader)
	if err != nil {
		return RequestLogEntry{}, fmt.Errorf("append request log: %w", err)
	}
	entry.ID = id

	err = s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketRequestLog)
		buf, _ := json.Marshal(entry) // plain struct — cannot fail
		if err := b.Put([]byte(entry.ID), buf); err != nil {
			return fmt.Errorf("put: %w", err)
		}
		return evictOldestRequestLog(b, MaxRequestLogEntries)
	})
	if err != nil {
		return RequestLogEntry{}, err
	}
	return entry, nil
}

// GetRequestLog returns the entry with the given id or ErrNotFound.
func (s *Store) GetRequestLog(id string) (RequestLogEntry, error) {
	var out RequestLogEntry
	err := s.db.View(func(tx *bolt.Tx) error {
		buf := tx.Bucket(bucketRequestLog).Get([]byte(id))
		if buf == nil {
			return ErrNotFound
		}
		if err := json.Unmarshal(buf, &out); err != nil {
			return fmt.Errorf("unmarshal log %s: %w", id, err)
		}
		return nil
	})
	if err != nil {
		return RequestLogEntry{}, err
	}
	return out, nil
}

// TailRequestLog returns up to limit most-recent entries (newest first).
// Pass 0 to use a sane default (50).
func (s *Store) TailRequestLog(limit int) ([]RequestLogEntry, error) {
	if limit <= 0 {
		limit = 50
	}
	out := make([]RequestLogEntry, 0, limit)
	err := s.db.View(func(tx *bolt.Tx) error {
		cur := tx.Bucket(bucketRequestLog).Cursor()
		// Reverse iteration: walk from the end backwards.
		for k, v := cur.Last(); k != nil && len(out) < limit; k, v = cur.Prev() {
			var entry RequestLogEntry
			if err := json.Unmarshal(v, &entry); err != nil {
				return fmt.Errorf("unmarshal log %s: %w", string(k), err)
			}
			out = append(out, entry)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

// CountRequestLog returns the current number of entries in the bucket.
func (s *Store) CountRequestLog() (int, error) {
	var n int
	err := s.db.View(func(tx *bolt.Tx) error {
		n = tx.Bucket(bucketRequestLog).Stats().KeyN
		return nil
	})
	return n, err
}

// evictOldestRequestLog trims b down to at most max entries by deleting the
// numerically smallest ULID keys. Runs inside the same writable tx as the
// append it follows so eviction is atomic with insertion.
//
// The count is computed by walking the cursor rather than via Stats().KeyN
// because bbolt's bucket stats are not guaranteed to reflect in-progress
// changes within the same transaction (the page rebalancing happens on
// commit). Deletion happens in a second pass because modifying the bucket
// while a cursor is iterating invalidates that cursor's position.
func evictOldestRequestLog(b *bolt.Bucket, limit int) error {
	if limit <= 0 {
		return nil
	}
	keys := make([][]byte, 0)
	cur := b.Cursor()
	for k, _ := cur.First(); k != nil; k, _ = cur.Next() {
		keyCopy := make([]byte, len(k))
		copy(keyCopy, k)
		keys = append(keys, keyCopy)
	}
	if len(keys) <= limit {
		return nil
	}
	excess := len(keys) - limit
	for i := range excess {
		if err := b.Delete(keys[i]); err != nil {
			return fmt.Errorf("evict %s: %w", string(keys[i]), err)
		}
	}
	return nil
}
