package webhooks

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Standard Webhooks HTTP headers (D-018). Values are unique per delivery
// attempt; the signature header may carry multiple comma-separated
// versions per RFC, though the mock only emits v1.
const (
	HeaderID        = "webhook-id"
	HeaderTimestamp = "webhook-timestamp"
	HeaderSignature = "webhook-signature"
)

// SigningVersion is the prefix used in the webhook-signature header
// (`v1,<base64>`). Only v1 is implemented.
const SigningVersion = "v1"

// MaxClockSkew is the tolerated drift between the receiver's clock and
// the timestamp header, in either direction. Picked to be loose enough
// for laptops with mild NTP drift but tight enough to refuse very old
// payloads.
const MaxClockSkew = 5 * time.Minute

// ErrInvalidSignature signals a verification failure. Returned by Verify
// when no provided signature matches the expected value.
var ErrInvalidSignature = errors.New("webhooks: invalid signature")

// ErrClockSkew signals a timestamp outside MaxClockSkew. Returned by
// Verify before the HMAC comparison so callers can give a clear error.
var ErrClockSkew = errors.New("webhooks: timestamp outside tolerance")

// Sign computes the v1 signature for an event payload. id is the event's
// webhook-id, ts is its webhook-timestamp value (Unix seconds), and body
// is the JSON envelope verbatim. Returns the header value formatted as
// "v1,<base64>".
func Sign(secret, id, ts string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(stripSecretPrefix(secret)))
	mac.Write([]byte(id))
	mac.Write([]byte("."))
	mac.Write([]byte(ts))
	mac.Write([]byte("."))
	mac.Write(body)
	return SigningVersion + "," + base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

// Verify recomputes the signature for the (id, ts, body) tuple and
// returns nil if any of the comma-separated values in sigHeader matches.
// The timestamp is checked against now within MaxClockSkew before HMAC
// comparison.
func Verify(secret, id, tsHeader, sigHeader string, body []byte, now time.Time) error {
	if id == "" || tsHeader == "" || sigHeader == "" {
		return fmt.Errorf("webhooks: missing required header")
	}
	tsSeconds, err := strconv.ParseInt(strings.TrimSpace(tsHeader), 10, 64)
	if err != nil {
		return fmt.Errorf("webhooks: parse timestamp: %w", err)
	}
	ts := time.Unix(tsSeconds, 0)
	if delta := absDuration(now.Sub(ts)); delta > MaxClockSkew {
		return fmt.Errorf("%w: %s", ErrClockSkew, delta)
	}
	expected := Sign(secret, id, tsHeader, body)
	expectedSig := strings.TrimPrefix(expected, SigningVersion+",")
	for _, candidate := range strings.Split(sigHeader, ",") {
		candidate = strings.TrimSpace(candidate)
		// Tolerate either "v1,<sig>" candidates or bare "<sig>" tokens.
		candidate = strings.TrimPrefix(candidate, SigningVersion+",")
		if hmac.Equal([]byte(candidate), []byte(expectedSig)) {
			return nil
		}
	}
	return ErrInvalidSignature
}

// ExtractHeaders pulls the three Standard Webhooks headers from h. Each
// returned value is empty if the header is absent — the caller decides
// whether that is a fatal validation error.
func ExtractHeaders(h http.Header) (id, ts, signature string) {
	return h.Get(HeaderID), h.Get(HeaderTimestamp), h.Get(HeaderSignature)
}

// stripSecretPrefix removes the `whsec_` ULID-ish prefix the mock seeds
// onto its webhook secret (scaffold §5.7 example: `whsec_<32hex>`). The
// signing material is the hex tail. If the prefix is absent the entire
// string is used.
func stripSecretPrefix(secret string) string {
	return strings.TrimPrefix(secret, "whsec_")
}

// absDuration returns the absolute value of d.
func absDuration(d time.Duration) time.Duration {
	if d < 0 {
		return -d
	}
	return d
}
