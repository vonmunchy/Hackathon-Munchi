package store_test

import (
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/store"
)

func newStore(t *testing.T) *store.Store {
	t.Helper()
	st, err := store.Open(filepath.Join(t.TempDir(), "state.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })
	return st
}

func TestAppendRequestLog_AssignsIDAndPersists(t *testing.T) {
	st := newStore(t)
	entry, err := st.AppendRequestLog(store.RequestLogEntry{
		Method: "GET", Path: "/health/alive", Status: 200, DurationMS: 4,
	})
	if err != nil {
		t.Fatalf("append: %v", err)
	}
	if entry.ID == "" {
		t.Fatal("ID not assigned")
	}
	got, err := st.GetRequestLog(entry.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Status != 200 || got.Method != "GET" {
		t.Errorf("got: %+v", got)
	}
}

func TestGetRequestLog_Missing_ReturnsNotFound(t *testing.T) {
	st := newStore(t)
	_, err := st.GetRequestLog("req_missing")
	if !errors.Is(err, store.ErrNotFound) {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}

func TestTailRequestLog_ReturnsNewestFirst(t *testing.T) {
	st := newStore(t)
	for i := range 5 {
		if _, err := st.AppendRequestLog(store.RequestLogEntry{Method: "GET", Path: "/p", Status: 200}); err != nil {
			t.Fatalf("append %d: %v", i, err)
		}
		time.Sleep(2 * time.Millisecond) // keep ULID timestamps distinct
	}
	tail, err := st.TailRequestLog(3)
	if err != nil {
		t.Fatalf("tail: %v", err)
	}
	if len(tail) != 3 {
		t.Fatalf("len = %d, want 3", len(tail))
	}
	for i := 1; i < len(tail); i++ {
		if tail[i-1].ID < tail[i].ID {
			t.Errorf("not newest-first: %s before %s", tail[i-1].ID, tail[i].ID)
		}
	}
}

func TestAppendRequestLog_EvictsOldestPastMax(t *testing.T) {
	// We can't easily set Max from tests; the constant is 1000 which is
	// expensive but tractable. Reduce to a quick check that the count never
	// exceeds Max + 1.
	st := newStore(t)
	for range store.MaxRequestLogEntries + 10 {
		if _, err := st.AppendRequestLog(store.RequestLogEntry{Method: "GET", Path: "/p"}); err != nil {
			t.Fatalf("append: %v", err)
		}
	}
	count, err := st.CountRequestLog()
	if err != nil {
		t.Fatalf("count: %v", err)
	}
	if count > store.MaxRequestLogEntries {
		t.Errorf("count %d exceeds max %d", count, store.MaxRequestLogEntries)
	}
}
