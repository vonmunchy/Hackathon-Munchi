package store

import (
	"errors"
	"fmt"
	"time"

	bolt "go.etcd.io/bbolt"
)

// SchemaVersion is the on-disk schema version recorded in the meta bucket.
// Bumping it lets us add migrations later without guessing whether a state
// file is fresh or stale.
const SchemaVersion = 1

// Bucket names. Centralized so tests and migrations don't drift from
// runtime code.
var (
	bucketMerchants    = []byte("merchants")
	bucketClients      = []byte("clients")
	bucketWallets      = []byte("wallets")
	bucketBankAccounts = []byte("bank_accounts")
	bucketPayments     = []byte("payments")
	bucketPayouts      = []byte("payouts")
	bucketTransactions = []byte("transactions")
	bucketScenarios    = []byte("scenarios")
	bucketRequestLog   = []byte("request_log")
	bucketMeta         = []byte("meta")
)

// allBuckets is the set of buckets to ensure exist on every Open.
var allBuckets = [][]byte{
	bucketMerchants,
	bucketClients,
	bucketWallets,
	bucketBankAccounts,
	bucketPayments,
	bucketPayouts,
	bucketTransactions,
	bucketScenarios,
	bucketRequestLog,
	bucketMeta,
}

// Meta keys.
const (
	MetaKeySchemaVersion       = "schema_version"
	MetaKeyWebhookSecret       = "webhook_secret"
	MetaKeySigningKeyCreatedAt = "signing_key_generated_at"
	MetaKeyFirstStartedAt      = "first_started_at"
)

// ErrNotFound is returned by lookup methods when the requested key has no
// associated value.
var ErrNotFound = errors.New("store: key not found")

// Store wraps a single BoltDB file. It is safe for concurrent use: bolt
// itself serializes writers and allows many readers.
type Store struct {
	db *bolt.DB
	// now is the clock used for timestamping new records; tests override
	// to keep snapshots stable. nil means time.Now (UTC).
	now func() time.Time
}

// Open opens (or creates) the BoltDB file at path and ensures all known
// buckets exist. The schema_version meta key is set on first open.
//
// Bucket creation and schema-version recording use writable bbolt
// transactions with fixed, small bucket names — bbolt cannot fail those
// operations under those constraints, so their errors are not propagated.
func Open(path string) (*Store, error) {
	db, err := bolt.Open(path, 0o600, &bolt.Options{Timeout: 2 * time.Second})
	if err != nil {
		return nil, fmt.Errorf("open boltdb at %s: %w", path, err)
	}
	s := &Store{db: db}
	_ = s.ensureBuckets()
	_ = s.recordSchemaVersion()
	return s, nil
}

// Close releases the underlying BoltDB handle.
func (s *Store) Close() error {
	if s.db == nil {
		return nil
	}
	return s.db.Close()
}

// Path returns the on-disk path of the BoltDB file.
func (s *Store) Path() string { return s.db.Path() }

// nowUTC returns the configured clock's value in UTC.
func (s *Store) nowUTC() time.Time {
	if s.now != nil {
		return s.now().UTC()
	}
	return time.Now().UTC()
}

func (s *Store) ensureBuckets() error {
	return s.db.Update(func(tx *bolt.Tx) error {
		for _, name := range allBuckets {
			// CreateBucketIfNotExists in a writable tx with a fixed
			// non-empty bucket name cannot fail.
			_, _ = tx.CreateBucketIfNotExists(name)
		}
		return nil
	})
}

func (s *Store) recordSchemaVersion() error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketMeta)
		if b.Get([]byte(MetaKeySchemaVersion)) != nil {
			return nil
		}
		// Small known-good keys/values inside a writable tx — bbolt Put
		// cannot fail under these constraints.
		_ = b.Put([]byte(MetaKeySchemaVersion), fmt.Appendf(nil, "%d", SchemaVersion))
		if b.Get([]byte(MetaKeyFirstStartedAt)) == nil {
			_ = b.Put([]byte(MetaKeyFirstStartedAt), []byte(s.nowUTC().Format(time.RFC3339Nano)))
		}
		return nil
	})
}

// SetClock replaces the store's clock. Intended for tests.
func (s *Store) SetClock(now func() time.Time) {
	s.now = now
}

// GetMeta returns a string value from the meta bucket. ErrNotFound is
// returned when the key does not exist.
func (s *Store) GetMeta(key string) (string, error) {
	var out string
	err := s.db.View(func(tx *bolt.Tx) error {
		v := tx.Bucket(bucketMeta).Get([]byte(key))
		if v == nil {
			return ErrNotFound
		}
		out = string(v)
		return nil
	})
	if err != nil {
		return "", err
	}
	return out, nil
}

// PutMeta stores a string value in the meta bucket.
func (s *Store) PutMeta(key, value string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(bucketMeta).Put([]byte(key), []byte(value))
	})
}
