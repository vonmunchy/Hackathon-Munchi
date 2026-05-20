package webhooks_test

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/webhooks"
)

func TestSign_Verify_RoundTrip(t *testing.T) {
	secret := "whsec_deadbeef"
	id := "msg_01ABC"
	ts := "1700000000"
	body := []byte(`{"eventType":"transaction.state_changed"}`)

	sig := webhooks.Sign(secret, id, ts, body)
	if err := webhooks.Verify(secret, id, ts, sig, body, time.Unix(1700000010, 0)); err != nil {
		t.Fatalf("verify: %v", err)
	}
}

func TestVerify_BadSignature_Errors(t *testing.T) {
	body := []byte(`{}`)
	err := webhooks.Verify("whsec_abc", "msg_x", "1700000000", "v1,deadbeef", body, time.Unix(1700000000, 0))
	if !errors.Is(err, webhooks.ErrInvalidSignature) {
		t.Errorf("err = %v, want ErrInvalidSignature", err)
	}
}

func TestVerify_TamperedBody_Errors(t *testing.T) {
	secret := "whsec_secret"
	id := "msg_a"
	ts := "1700000000"
	body := []byte(`{"original":true}`)
	sig := webhooks.Sign(secret, id, ts, body)
	tampered := []byte(`{"original":false}`)
	if err := webhooks.Verify(secret, id, ts, sig, tampered, time.Unix(1700000000, 0)); !errors.Is(err, webhooks.ErrInvalidSignature) {
		t.Errorf("err = %v, want ErrInvalidSignature on tampered body", err)
	}
}

func TestVerify_ClockSkew_TooOld_Errors(t *testing.T) {
	secret := "whsec_x"
	id := "msg_a"
	ts := "1700000000"
	body := []byte(`{}`)
	sig := webhooks.Sign(secret, id, ts, body)
	tooLate := time.Unix(1700000000+int64((webhooks.MaxClockSkew+time.Minute)/time.Second), 0)
	if err := webhooks.Verify(secret, id, ts, sig, body, tooLate); !errors.Is(err, webhooks.ErrClockSkew) {
		t.Errorf("err = %v, want ErrClockSkew", err)
	}
}

func TestVerify_MultiVersionHeader_AcceptsAnyMatch(t *testing.T) {
	secret := "whsec_zzz"
	id := "msg_m"
	ts := "1700000000"
	body := []byte(`{}`)
	good := webhooks.Sign(secret, id, ts, body)
	// Mix one bogus v1 entry with the real one — should still verify.
	header := "v1,not-a-real-sig," + good
	if err := webhooks.Verify(secret, id, ts, header, body, time.Unix(1700000000, 0)); err != nil {
		t.Errorf("verify: %v", err)
	}
}

func TestExtractHeaders_ReadsAll(t *testing.T) {
	h := http.Header{}
	h.Set(webhooks.HeaderID, "msg_y")
	h.Set(webhooks.HeaderTimestamp, "1700000001")
	h.Set(webhooks.HeaderSignature, "v1,xxx")
	id, ts, sig := webhooks.ExtractHeaders(h)
	if id != "msg_y" || ts != "1700000001" || sig != "v1,xxx" {
		t.Errorf("got id=%q ts=%q sig=%q", id, ts, sig)
	}
}
