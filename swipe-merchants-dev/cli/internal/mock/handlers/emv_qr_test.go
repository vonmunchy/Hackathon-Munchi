package handlers

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image/png"
	"regexp"
	"strings"
	"testing"

	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/store"
)

func TestBuildEMVCoQR_StructureAndCRC(t *testing.T) {
	emv := buildEMVCoQR(42.50, store.CurrencyMVR, "U103MCDB", "Default Test Merchant")

	// Must start with the EMVCo MPM payload format indicator (tag 00,
	// length 02, value "01").
	if !strings.HasPrefix(emv, "000201") {
		t.Fatalf("missing payload format indicator: %q", emv)
	}

	// Must end with a tag-63 CRC TLV (4-hex value).
	crcTag := regexp.MustCompile(`6304[0-9A-F]{4}$`)
	if !crcTag.MatchString(emv) {
		t.Fatalf("CRC tag not present or malformed at tail: %q", emv)
	}

	// Recomputing the CRC over the payload (minus the last 4 hex
	// characters) must match the appended CRC.
	body := emv[:len(emv)-4]
	want := fmt.Sprintf("%04X", crc16CCITTFalse([]byte(body)))
	got := emv[len(emv)-4:]
	if got != want {
		t.Errorf("CRC mismatch: payload says %q, recomputed %q", got, want)
	}

	// Tag-53 (currency) must carry the ISO-4217 numeric for MVR.
	if !strings.Contains(emv, "5303462") {
		t.Errorf("expected MVR currency code 462 (tag 53), got: %q", emv)
	}

	// Tag-58 (country) must be MV.
	if !strings.Contains(emv, "5802MV") {
		t.Errorf("expected country MV (tag 58), got: %q", emv)
	}

	// The short code must appear inside the tag-26 merchant info
	// subfield 01.
	if !strings.Contains(emv, "U103MCDB") {
		t.Errorf("expected short code embedded in tag 26, got: %q", emv)
	}
}

func TestBuildEMVCoQR_USDCurrencyCode(t *testing.T) {
	emv := buildEMVCoQR(100, store.CurrencyUSD, "AAAA1111", "M")
	if !strings.Contains(emv, "5303840") {
		t.Errorf("expected USD currency code 840 (tag 53), got: %q", emv)
	}
}

func TestBuildPaymentArtifacts_QR_DefaultFormat_PaymentURLPopulated(t *testing.T) {
	code, qrData, payURL, err := buildPaymentArtifacts(store.PaymentTypeQR, "http://localhost:8080", QRFormatJSON, 1.00, store.CurrencyMVR, "M")
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	if payURL != "http://localhost:8080/pay/"+code {
		t.Errorf("payment_url = %q, want pay-page URL for code %q", payURL, code)
	}
	assertPNG(t, qrData)
}

func TestBuildQRJSONPayload_ContainsURLAndDetails(t *testing.T) {
	content := buildQRJSONPayload("http://localhost:8080/pay/ABC12345", "ABC12345", 42.50, store.CurrencyMVR, "Test Merchant")
	var got qrJSONPayload
	if err := json.Unmarshal([]byte(content), &got); err != nil {
		t.Fatalf("unmarshal: %v\nraw: %s", err, content)
	}
	if got.URL != "http://localhost:8080/pay/ABC12345" {
		t.Errorf("url = %q", got.URL)
	}
	if got.Reference != "ABC12345" {
		t.Errorf("reference = %q", got.Reference)
	}
	if got.Amount != 42.50 {
		t.Errorf("amount = %v", got.Amount)
	}
	if got.Currency != "MVR" {
		t.Errorf("currency = %q", got.Currency)
	}
	if got.Merchant != "Test Merchant" {
		t.Errorf("merchant = %q", got.Merchant)
	}
}

func TestBuildPaymentArtifacts_QR_EMVCoFormat_StillPopulatesPaymentURL(t *testing.T) {
	code, qrData, payURL, err := buildPaymentArtifacts(store.PaymentTypeQR, "http://localhost:8080", QRFormatEMVCo, 1.00, store.CurrencyMVR, "M")
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	if payURL != "http://localhost:8080/pay/"+code {
		t.Errorf("payment_url = %q, want pay-page URL even when QR encodes EMVCo", payURL)
	}
	assertPNG(t, qrData)
}

func TestRenderQRPNG_RoundTripsThroughPNGDecoder(t *testing.T) {
	emv := buildEMVCoQR(42.50, store.CurrencyMVR, "U103MCDB", "Default Test Merchant")
	raw, err := renderQRPNG(emv)
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	if _, err := png.Decode(bytes.NewReader(raw)); err != nil {
		t.Fatalf("rendered bytes are not a valid PNG: %v", err)
	}
}

// assertPNG decodes the base64 qr_data and verifies the bytes are a
// valid PNG of the expected size. Both QR-format branches use it.
func assertPNG(t *testing.T, qrData string) {
	t.Helper()
	raw, err := base64.StdEncoding.DecodeString(qrData)
	if err != nil {
		t.Fatalf("qr_data is not valid base64: %v", err)
	}
	pngSig := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	if len(raw) < len(pngSig) || !bytes.Equal(raw[:len(pngSig)], pngSig) {
		t.Fatalf("qr_data does not start with PNG signature: % x", raw[:min(8, len(raw))])
	}
	img, err := png.Decode(bytes.NewReader(raw))
	if err != nil {
		t.Fatalf("qr_data PNG decode: %v", err)
	}
	b := img.Bounds()
	if b.Dx() != qrPNGSize || b.Dy() != qrPNGSize {
		t.Errorf("PNG size = %dx%d, want %dx%d", b.Dx(), b.Dy(), qrPNGSize, qrPNGSize)
	}
}

func TestCRC16CCITTFalse_KnownVector(t *testing.T) {
	// Per EMVCo MPM Annex B sample: payload
	//   "00020101021229300012D156000000000510A93FO3230Q31280012D15600000001030812345678520441115802CN5914BEST TRANSPORT6007BEIJING64200002ZH0104最佳运输0202北京540523.7253031565502016233030412340603***0708A60086670902ME91320016A0112233449988770708123456786304"
	// has CRC value "A13A".
	// We construct it programmatically to keep this test stable.
	input := "00020101021229300012D156000000000510A93FO3230Q31280012D15600000001030812345678520441115802CN5914BEST TRANSPORT6007BEIJING64200002ZH0104最佳运输0202北京540523.7253031565502016233030412340603***0708A60086670902ME91320016A0112233449988770708123456786304"
	got := fmt.Sprintf("%04X", crc16CCITTFalse([]byte(input)))
	want := "A13A"
	if got != want {
		t.Errorf("CRC = %s, want %s (EMVCo sample)", got, want)
	}
}
