package mock

import (
	"bytes"
	"strings"
	"testing"

	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/cli"
	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/store"
)

func TestNormalizeBaseURL_AllVariants(t *testing.T) {
	t.Parallel()
	cases := map[string]string{
		"127.0.0.1:8080": "http://127.0.0.1:8080",
		"localhost:8080": "http://localhost:8080",
		":8080":          "http://localhost:8080",
		"0.0.0.0:8080":   "http://localhost:8080",
		"[::]:8080":      "http://localhost:8080",
		"example.com:80": "http://example.com:80",
	}
	for in, want := range cases {
		if got := normalizeBaseURL(in); got != want {
			t.Errorf("normalizeBaseURL(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestResolvePort_FlagOverride_Wins(t *testing.T) {
	t.Parallel()
	g := &cli.GlobalFlags{Port: 9999}
	if got := resolvePort(g); got != 9999 {
		t.Errorf("resolvePort = %d, want 9999", got)
	}
}

func TestResolvePort_Zero_FallsBackTo8080(t *testing.T) {
	t.Parallel()
	g := &cli.GlobalFlags{Port: 0}
	if got := resolvePort(g); got != 8080 {
		t.Errorf("resolvePort = %d, want 8080", got)
	}
}

func TestPrintSeedSummary_StripsZerosBindAddr(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	printSeedSummary(&buf, "0.0.0.0:8080", "", store.SeedSummary{
		Merchant:      store.Merchant{ID: "mer_x"},
		Wallet:        store.Wallet{Balances: map[store.Currency]store.CurrencyBalance{store.CurrencyMVR: {Available: 1.5}}},
		BankAccounts:  []store.BankAccount{{ID: "bnk_x"}},
		WebhookSecret: "whsec_x",
		FreshSeed:     true,
	})
	out := buf.String()
	if !strings.Contains(out, "http://localhost:8080") {
		t.Errorf("0.0.0.0 not normalized to localhost:\n%s", out)
	}
	if strings.Contains(out, "0.0.0.0") {
		t.Errorf("output still contains 0.0.0.0:\n%s", out)
	}
}
