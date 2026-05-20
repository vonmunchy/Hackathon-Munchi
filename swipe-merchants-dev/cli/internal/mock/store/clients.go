package store

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	bolt "go.etcd.io/bbolt"
	"golang.org/x/crypto/bcrypt"
)

// Client is an OAuth2 client owned by a merchant. Secrets are stored as
// bcrypt hashes per D-012; the plaintext only exists at creation/rotation
// time, returned once to the caller and then discarded.
type Client struct {
	ID         string    `json:"id"`
	MerchantID string    `json:"merchant_id"`
	Name       string    `json:"name"`
	SecretHash string    `json:"secret_hash"`
	Scopes     []string  `json:"scopes"`
	Enabled    bool      `json:"enabled"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// ClientSecretBcryptCost is the bcrypt work factor for hashing client
// secrets per D-012.
const ClientSecretBcryptCost = 10

// ErrInvalidSecret is returned by VerifyClientSecret when the plaintext
// does not match the stored hash.
var ErrInvalidSecret = fmt.Errorf("store: client secret mismatch")

// CreateClient persists a new client with a freshly-generated ULID, hashes
// the secret with bcrypt, and returns both the persisted record and the
// plaintext secret. The plaintext is the only copy ever produced — the
// store retains only the hash.
//
// Scopes are deduplicated and sorted alphabetically for stable output.
func (s *Store) CreateClient(merchantID, name string, scopes []string) (Client, string, error) {
	if merchantID == "" {
		return Client{}, "", fmt.Errorf("create client: merchant id required")
	}
	if name == "" {
		return Client{}, "", fmt.Errorf("create client: name required")
	}
	if _, err := s.GetMerchant(merchantID); err != nil {
		return Client{}, "", fmt.Errorf("create client: %w", err)
	}

	now := s.nowUTC()
	id, err := newULID(now, "cli_", rand.Reader)
	if err != nil {
		return Client{}, "", fmt.Errorf("create client: %w", err)
	}
	secret, err := generateClientSecret()
	if err != nil {
		return Client{}, "", fmt.Errorf("create client: %w", err)
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(secret), ClientSecretBcryptCost)
	if err != nil {
		return Client{}, "", fmt.Errorf("create client: hash secret: %w", err)
	}

	c := Client{
		ID:         id,
		MerchantID: merchantID,
		Name:       name,
		SecretHash: string(hash),
		Scopes:     normalizeScopes(scopes),
		Enabled:    true,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if err := s.putClient(c); err != nil {
		return Client{}, "", err
	}
	return c, secret, nil
}

// GetClient returns the client with the given id, or ErrNotFound.
func (s *Store) GetClient(id string) (Client, error) {
	var c Client
	err := s.db.View(func(tx *bolt.Tx) error {
		buf := tx.Bucket(bucketClients).Get([]byte(id))
		if buf == nil {
			return ErrNotFound
		}
		if err := json.Unmarshal(buf, &c); err != nil {
			return fmt.Errorf("unmarshal client %s: %w", id, err)
		}
		return nil
	})
	if err != nil {
		return Client{}, err
	}
	return c, nil
}

// ListClients returns every client ordered by id (which is ULID-sortable).
// Bolt's ForEach already iterates in key order, so no extra sort is needed,
// but we keep an explicit Sort to stay tolerant of future migrations.
func (s *Store) ListClients() ([]Client, error) {
	var out []Client
	err := s.db.View(func(tx *bolt.Tx) error {
		return tx.Bucket(bucketClients).ForEach(func(_, v []byte) error {
			var c Client
			if err := json.Unmarshal(v, &c); err != nil {
				return fmt.Errorf("unmarshal client: %w", err)
			}
			out = append(out, c)
			return nil
		})
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

// UpdateClient applies the supplied mutator inside a single write transaction.
// The mutator is given the current client and must return the updated copy.
// Returns ErrNotFound if the client does not exist.
func (s *Store) UpdateClient(id string, mut func(Client) Client) (Client, error) {
	var updated Client
	err := s.db.Update(func(tx *bolt.Tx) error {
		buf := tx.Bucket(bucketClients).Get([]byte(id))
		if buf == nil {
			return ErrNotFound
		}
		var c Client
		if err := json.Unmarshal(buf, &c); err != nil {
			return fmt.Errorf("unmarshal client %s: %w", id, err)
		}
		c = mut(c)
		c.UpdatedAt = s.nowUTC()
		c.Scopes = normalizeScopes(c.Scopes)
		newBuf, _ := json.Marshal(c) // plain struct — cannot fail
		if err := tx.Bucket(bucketClients).Put([]byte(c.ID), newBuf); err != nil {
			return fmt.Errorf("put client %s: %w", id, err)
		}
		updated = c
		return nil
	})
	if err != nil {
		return Client{}, err
	}
	return updated, nil
}

// RotateClientSecret generates a new secret for the client, replaces the
// stored hash, and returns the plaintext. The plaintext is returned only
// once — there is no way to recover it later.
func (s *Store) RotateClientSecret(id string) (Client, string, error) {
	secret, err := generateClientSecret()
	if err != nil {
		return Client{}, "", fmt.Errorf("rotate client secret: %w", err)
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(secret), ClientSecretBcryptCost)
	if err != nil {
		return Client{}, "", fmt.Errorf("rotate client secret: hash: %w", err)
	}
	updated, err := s.UpdateClient(id, func(c Client) Client {
		c.SecretHash = string(hash)
		return c
	})
	if err != nil {
		return Client{}, "", err
	}
	return updated, secret, nil
}

// DeleteClient hard-removes the client. Returns nil if the client was not
// present; deletion is idempotent.
func (s *Store) DeleteClient(id string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(bucketClients).Delete([]byte(id))
	})
}

// VerifyClientSecret compares the plaintext secret against the bcrypt hash
// for the named client. It returns ErrInvalidSecret on a mismatch, ErrNotFound
// when the client does not exist, and other errors from the underlying store.
// A revoked (Enabled=false) client always fails verification.
func (s *Store) VerifyClientSecret(id, plaintext string) (Client, error) {
	c, err := s.GetClient(id)
	if err != nil {
		return Client{}, err
	}
	if !c.Enabled {
		return Client{}, ErrInvalidSecret
	}
	if err := bcrypt.CompareHashAndPassword([]byte(c.SecretHash), []byte(plaintext)); err != nil {
		return Client{}, ErrInvalidSecret
	}
	return c, nil
}

func (s *Store) putClient(c Client) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		buf, _ := json.Marshal(c) // plain struct — cannot fail
		return tx.Bucket(bucketClients).Put([]byte(c.ID), buf)
	})
}

// normalizeScopes returns a copy of scopes with leading/trailing whitespace
// trimmed, empty entries removed, duplicates collapsed, and the result
// sorted. Stability matters because scopes appear in JWT claims and CLI
// table output.
func normalizeScopes(scopes []string) []string {
	if len(scopes) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(scopes))
	out := make([]string, 0, len(scopes))
	for _, s := range scopes {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	sort.Strings(out)
	return out
}

// generateClientSecret returns an opaque secret suitable for OAuth client
// credentials: 32 hex characters, prefixed with `sec_` for easy logging
// recognition. rand.Read on a healthy OS never fails; an error is
// nonetheless surfaced because secret entropy is security-critical and the
// caller deserves to know.
func generateClientSecret() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("read entropy: %w", err)
	}
	return "sec_" + hex.EncodeToString(buf), nil
}
