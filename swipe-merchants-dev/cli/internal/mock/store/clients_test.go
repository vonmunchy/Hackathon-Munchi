package store_test

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/store"
)

func openSeededStore(t *testing.T) *store.Store {
	t.Helper()
	st, err := store.Open(filepath.Join(t.TempDir(), "state.db"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })
	if _, err := st.SeedDefaults(); err != nil {
		t.Fatalf("seed: %v", err)
	}
	return st
}

func TestCreateClient_PersistsAndReturnsPlaintextOnce(t *testing.T) {
	st := openSeededStore(t)
	c, secret, err := st.CreateClient(store.DefaultMerchantID, "My Test", []string{"wallet:balance", "payments:qr"})
	if err != nil {
		t.Fatalf("create client: %v", err)
	}
	if !strings.HasPrefix(c.ID, "cli_") {
		t.Errorf("client id = %q, want cli_ prefix", c.ID)
	}
	if !strings.HasPrefix(secret, "sec_") {
		t.Errorf("secret = %q, want sec_ prefix", secret)
	}
	if c.SecretHash == "" || c.SecretHash == secret {
		t.Errorf("secret hash empty or equals plaintext: hash=%q", c.SecretHash)
	}
	if !c.Enabled {
		t.Errorf("new client should be enabled")
	}
	want := []string{"payments:qr", "wallet:balance"}
	if got := c.Scopes; !equalStrings(got, want) {
		t.Errorf("scopes = %v, want %v", got, want)
	}

	got, err := st.GetClient(c.ID)
	if err != nil {
		t.Fatalf("get client: %v", err)
	}
	if got.ID != c.ID || got.Name != "My Test" {
		t.Errorf("get mismatch: %+v", got)
	}
}

func TestCreateClient_RejectsUnknownMerchant(t *testing.T) {
	st := openSeededStore(t)
	_, _, err := st.CreateClient("mer_ghost", "x", nil)
	if !errors.Is(err, store.ErrNotFound) {
		t.Errorf("err = %v, want ErrNotFound wrapped", err)
	}
}

func TestVerifyClientSecret_Matches(t *testing.T) {
	st := openSeededStore(t)
	c, secret, _ := st.CreateClient(store.DefaultMerchantID, "t", []string{"wallet:balance"})
	verified, err := st.VerifyClientSecret(c.ID, secret)
	if err != nil {
		t.Fatalf("verify ok case: %v", err)
	}
	if verified.ID != c.ID {
		t.Errorf("verified id mismatch")
	}
}

func TestVerifyClientSecret_RejectsWrong(t *testing.T) {
	st := openSeededStore(t)
	c, _, _ := st.CreateClient(store.DefaultMerchantID, "t", nil)
	_, err := st.VerifyClientSecret(c.ID, "sec_wrong")
	if !errors.Is(err, store.ErrInvalidSecret) {
		t.Errorf("err = %v, want ErrInvalidSecret", err)
	}
}

func TestVerifyClientSecret_RejectsRevoked(t *testing.T) {
	st := openSeededStore(t)
	c, secret, _ := st.CreateClient(store.DefaultMerchantID, "t", nil)
	if _, err := st.UpdateClient(c.ID, func(cl store.Client) store.Client {
		cl.Enabled = false
		return cl
	}); err != nil {
		t.Fatalf("update: %v", err)
	}
	if _, err := st.VerifyClientSecret(c.ID, secret); !errors.Is(err, store.ErrInvalidSecret) {
		t.Errorf("revoked client should fail verify, got %v", err)
	}
}

func TestRotateClientSecret_ReplacesHash(t *testing.T) {
	st := openSeededStore(t)
	c, originalSecret, _ := st.CreateClient(store.DefaultMerchantID, "t", nil)
	updated, newSecret, err := st.RotateClientSecret(c.ID)
	if err != nil {
		t.Fatalf("rotate: %v", err)
	}
	if newSecret == originalSecret {
		t.Errorf("rotation returned the same secret")
	}
	if updated.SecretHash == c.SecretHash {
		t.Errorf("rotation did not change hash")
	}
	if _, err := st.VerifyClientSecret(c.ID, originalSecret); !errors.Is(err, store.ErrInvalidSecret) {
		t.Errorf("old secret should no longer verify, got %v", err)
	}
	if _, err := st.VerifyClientSecret(c.ID, newSecret); err != nil {
		t.Errorf("new secret should verify, got %v", err)
	}
}

func TestListClients_SortedByID(t *testing.T) {
	st := openSeededStore(t)
	for i := 0; i < 3; i++ {
		if _, _, err := st.CreateClient(store.DefaultMerchantID, "c", nil); err != nil {
			t.Fatalf("create %d: %v", i, err)
		}
		time.Sleep(2 * time.Millisecond) // keep ULID timestamps distinct
	}
	clients, err := st.ListClients()
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(clients) != 3 {
		t.Fatalf("count = %d, want 3", len(clients))
	}
	for i := 1; i < len(clients); i++ {
		if clients[i-1].ID >= clients[i].ID {
			t.Errorf("list out of order: %s >= %s", clients[i-1].ID, clients[i].ID)
		}
	}
}

func TestDeleteClient_IsIdempotent(t *testing.T) {
	st := openSeededStore(t)
	c, _, _ := st.CreateClient(store.DefaultMerchantID, "t", nil)
	if err := st.DeleteClient(c.ID); err != nil {
		t.Fatalf("first delete: %v", err)
	}
	if err := st.DeleteClient(c.ID); err != nil {
		t.Errorf("second delete should be no-op, got %v", err)
	}
	if _, err := st.GetClient(c.ID); !errors.Is(err, store.ErrNotFound) {
		t.Errorf("get after delete = %v, want ErrNotFound", err)
	}
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
