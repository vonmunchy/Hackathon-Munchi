package store

import (
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	bolt "go.etcd.io/bbolt"
)

// White-box tests covering accessors, error paths, and clock injection
// that the black-box tests in store_test.go cannot reach.

func TestStore_Path_ReturnsOpenedPath(t *testing.T) {
	t.Parallel()
	dbPath := filepath.Join(t.TempDir(), "state.db")
	st, err := Open(dbPath)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })
	if got := st.Path(); got != dbPath {
		t.Errorf("Path() = %q, want %q", got, dbPath)
	}
}

func TestStore_Close_NilDB_NoOps(t *testing.T) {
	t.Parallel()
	s := &Store{}
	if err := s.Close(); err != nil {
		t.Errorf("close on zero-value Store: %v", err)
	}
}

func TestStore_Close_Idempotent(t *testing.T) {
	t.Parallel()
	dbPath := filepath.Join(t.TempDir(), "state.db")
	st, err := Open(dbPath)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	if err := st.Close(); err != nil {
		t.Errorf("first close: %v", err)
	}
	// Second close on a closed bbolt DB returns an error; we just want
	// the call to not panic, which is the point of this test.
	_ = st.Close()
}

func TestStore_SetClock_OverridesNowUTC(t *testing.T) {
	t.Parallel()
	dbPath := filepath.Join(t.TempDir(), "state.db")
	st, err := Open(dbPath)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })

	fixed := time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC)
	st.SetClock(func() time.Time { return fixed })
	got := st.nowUTC()
	if !got.Equal(fixed) {
		t.Errorf("nowUTC = %v, want %v", got, fixed)
	}
}

func TestOpen_InvalidPath_ReturnsError(t *testing.T) {
	t.Parallel()
	// A directory cannot be opened as a BoltDB file.
	dir := t.TempDir()
	_, err := Open(dir)
	if err == nil {
		t.Fatalf("expected error opening a directory as bolt; got nil")
	}
	if !strings.Contains(err.Error(), "open boltdb") {
		t.Errorf("err = %q, want contains 'open boltdb'", err)
	}
}

func TestRecordSchemaVersion_AlreadyRecorded_NoOp(t *testing.T) {
	t.Parallel()
	dbPath := filepath.Join(t.TempDir(), "state.db")
	st, err := Open(dbPath)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	first, _ := st.GetMeta(MetaKeyFirstStartedAt)
	// Calling recordSchemaVersion again should NOT overwrite the first-started-at.
	if err := st.recordSchemaVersion(); err != nil {
		t.Fatalf("record: %v", err)
	}
	second, _ := st.GetMeta(MetaKeyFirstStartedAt)
	if first != second {
		t.Errorf("first_started_at changed: %q -> %q", first, second)
	}
	_ = st.Close()
}

func TestPutWallet_EmptyMerchantID_Errors(t *testing.T) {
	t.Parallel()
	st, err := Open(filepath.Join(t.TempDir(), "state.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })
	if err := st.PutWallet(Wallet{}); err == nil {
		t.Errorf("expected error for empty merchant id")
	}
}

func TestPutBankAccount_EmptyMerchantID_Errors(t *testing.T) {
	t.Parallel()
	st, err := Open(filepath.Join(t.TempDir(), "state.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })
	if err := st.PutBankAccount(BankAccount{ID: "bnk_x"}); err == nil {
		t.Errorf("expected error for empty merchant id")
	}
}

func TestPutBankAccount_EmptyID_Errors(t *testing.T) {
	t.Parallel()
	st, err := Open(filepath.Join(t.TempDir(), "state.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })
	if err := st.PutBankAccount(BankAccount{}); err == nil {
		t.Errorf("expected error for empty id")
	}
}

func TestGetMeta_MissingKey_ReturnsErrNotFound(t *testing.T) {
	t.Parallel()
	st, err := Open(filepath.Join(t.TempDir(), "state.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })
	if _, err := st.GetMeta("does_not_exist"); err == nil {
		t.Errorf("expected ErrNotFound for missing key")
	}
}

func TestPutMerchant_HappyPath_RoundTrips(t *testing.T) {
	t.Parallel()
	st, err := Open(filepath.Join(t.TempDir(), "state.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })

	want := Merchant{ID: "mer_test_1", Name: "Test", CreatedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}
	if err := st.PutMerchant(want); err != nil {
		t.Fatalf("put: %v", err)
	}
	got, err := st.GetMerchant(want.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.ID != want.ID || got.Name != want.Name {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func TestPutWallet_HappyPath_RoundTrips(t *testing.T) {
	t.Parallel()
	st, err := Open(filepath.Join(t.TempDir(), "state.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })

	want := Wallet{
		MerchantID: "mer_x",
		ID:         "wal_x",
		Balances:   map[Currency]CurrencyBalance{CurrencyMVR: {Available: 5}},
	}
	if err := st.PutWallet(want); err != nil {
		t.Fatalf("put: %v", err)
	}
	got, err := st.GetWallet(want.MerchantID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.ID != want.ID || got.Balances[CurrencyMVR].Available != 5 {
		t.Errorf("round-trip mismatch: got %+v", got)
	}
}

func TestPutBankAccount_HappyPath_RoundTrips(t *testing.T) {
	t.Parallel()
	st, err := Open(filepath.Join(t.TempDir(), "state.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })

	want := BankAccount{
		ID:                "bnk_test",
		MerchantID:        "mer_x",
		AccountNumber:     "777",
		AccountHolderName: "x",
		Currency:          CurrencyMVR,
		Status:            BankAccountActive,
	}
	if err := st.PutBankAccount(want); err != nil {
		t.Fatalf("put: %v", err)
	}
	got, err := st.ListBankAccounts("mer_x")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(got) != 1 || got[0].ID != want.ID {
		t.Errorf("round-trip mismatch: got %+v", got)
	}
}

// putRawValue writes raw bytes into the named bucket — used to inject
// corrupt JSON to drive the Unmarshal error branches in store accessors.
func (s *Store) putRawValue(bucket, key string, value []byte) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket([]byte(bucket)).Put([]byte(key), value)
	})
}

func TestGetMerchant_CorruptValue_ReturnsUnmarshalError(t *testing.T) {
	t.Parallel()
	st, err := Open(filepath.Join(t.TempDir(), "state.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })
	if err := st.putRawValue("merchants", "mer_corrupt", []byte("{not json")); err != nil {
		t.Fatalf("seed corrupt: %v", err)
	}
	if _, err := st.GetMerchant("mer_corrupt"); err == nil {
		t.Errorf("expected unmarshal error")
	}
}

func TestGetWallet_CorruptValue_ReturnsUnmarshalError(t *testing.T) {
	t.Parallel()
	st, err := Open(filepath.Join(t.TempDir(), "state.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })
	if err := st.putRawValue("wallets", "mer_corrupt", []byte("{not json")); err != nil {
		t.Fatalf("seed corrupt: %v", err)
	}
	if _, err := st.GetWallet("mer_corrupt"); err == nil {
		t.Errorf("expected unmarshal error")
	}
}

func TestListMerchants_CorruptValue_ReturnsUnmarshalError(t *testing.T) {
	t.Parallel()
	st, err := Open(filepath.Join(t.TempDir(), "state.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })
	if err := st.putRawValue("merchants", "mer_corrupt", []byte("{not json")); err != nil {
		t.Fatalf("seed corrupt: %v", err)
	}
	if _, err := st.ListMerchants(); err == nil {
		t.Errorf("expected unmarshal error")
	}
}

func TestListBankAccounts_CorruptValue_ReturnsUnmarshalError(t *testing.T) {
	t.Parallel()
	st, err := Open(filepath.Join(t.TempDir(), "state.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })
	if err := st.putRawValue("bank_accounts", "bnk_corrupt", []byte("{not json")); err != nil {
		t.Fatalf("seed corrupt: %v", err)
	}
	if _, err := st.ListBankAccounts("any"); err == nil {
		t.Errorf("expected unmarshal error")
	}
}

func TestSeedDefaults_CorruptDefaultMerchant_ReturnsError(t *testing.T) {
	t.Parallel()
	st, err := Open(filepath.Join(t.TempDir(), "state.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })
	if err := st.putRawValue("merchants", DefaultMerchantID, []byte("{not json")); err != nil {
		t.Fatalf("seed corrupt: %v", err)
	}
	if _, err := st.SeedDefaults(); err == nil {
		t.Errorf("expected error from corrupt default merchant")
	}
}

func TestSummary_CorruptBankAccount_ReturnsError(t *testing.T) {
	t.Parallel()
	st, err := Open(filepath.Join(t.TempDir(), "state.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })

	// Need: a valid merchant + valid wallet so the first two reads inside
	// summary succeed, then corrupt bank account so ListBankAccounts errors.
	if _, err := st.SeedDefaults(); err != nil {
		t.Fatalf("seed: %v", err)
	}
	if err := st.putRawValue("bank_accounts", "bnk_corrupt", []byte("{not json")); err != nil {
		t.Fatalf("inject corrupt bank: %v", err)
	}
	m, _ := st.GetMerchant(DefaultMerchantID)
	if _, err := st.summary(m, false); err == nil {
		t.Errorf("expected error from corrupt bank account")
	}
}

func TestSummary_MissingWallet_ReturnsError(t *testing.T) {
	t.Parallel()
	st, err := Open(filepath.Join(t.TempDir(), "state.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })

	// summary expects the wallet to exist; pass a merchant whose wallet
	// has not been written.
	m := Merchant{ID: "mer_no_wallet", Name: "x"}
	if _, err := st.summary(m, false); err == nil {
		t.Errorf("expected error from missing wallet")
	}
}

func TestGetWallet_Missing_ReturnsErrNotFound(t *testing.T) {
	t.Parallel()
	st, err := Open(filepath.Join(t.TempDir(), "state.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })
	if _, err := st.GetWallet("mer_missing"); err == nil {
		t.Errorf("expected ErrNotFound")
	}
}

func TestGenerateWebhookSecret_HasExpectedShape(t *testing.T) {
	t.Parallel()
	got := generateWebhookSecret()
	if !strings.HasPrefix(got, "whsec_") {
		t.Errorf("secret = %q, want whsec_ prefix", got)
	}
	if len(got) != len("whsec_")+32 {
		t.Errorf("secret length = %d, want %d", len(got), len("whsec_")+32)
	}
	// Two consecutive calls should produce different values.
	if got == generateWebhookSecret() {
		t.Errorf("two calls produced identical secret %q", got)
	}
}

func TestNewULID_DifferentEntropy_ProducesDifferentIDs(t *testing.T) {
	t.Parallel()
	now := time.Now().UTC()
	a, err := newULID(now, "x_", randReader{})
	if err != nil {
		t.Fatalf("a: %v", err)
	}
	b, err := newULID(now, "x_", randReader{})
	if err != nil {
		t.Fatalf("b: %v", err)
	}
	if a == b {
		t.Errorf("two ULIDs collide: %q", a)
	}
	if !strings.HasPrefix(a, "x_") {
		t.Errorf("prefix not honored: %q", a)
	}
}

// randReader is a tiny io.Reader that produces distinct bytes per call,
// sufficient for ULID entropy in tests. A process-wide monotonic counter is
// mixed into the seed so two reads in the same nanosecond still differ.
type randReader struct{}

// readerSeq is a process-wide counter that keeps each randReader.Read call
// distinct without requiring callers to manage state.
var readerSeq atomic.Uint64

func (randReader) Read(p []byte) (int, error) {
	seq := readerSeq.Add(1)
	seed := uint64(time.Now().UnixNano()) ^ seq //nolint:gosec // intentional non-crypto
	for i := range p {
		seed = seed*6364136223846793005 + 1442695040888963407
		p[i] = byte(seed >> 32)
	}
	return len(p), nil
}

// failingReader always errors. Drives the entropy-failure branch in newULID.
type failingReader struct{}

func (failingReader) Read(_ []byte) (int, error) {
	return 0, errFailingReader
}

var errFailingReader = newSentinelErr("synthetic reader failure")

func newSentinelErr(msg string) error {
	return &sentinelErr{msg: msg}
}

type sentinelErr struct{ msg string }

func (e *sentinelErr) Error() string { return e.msg }

func TestNewULID_ReaderFails_ReturnsError(t *testing.T) {
	t.Parallel()
	if _, err := newULID(time.Now().UTC(), "x_", failingReader{}); err == nil {
		t.Errorf("expected error from failing entropy reader")
	}
}
