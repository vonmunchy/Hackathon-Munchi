package handlers

import (
	"fmt"
	"strconv"

	"github.com/skip2/go-qrcode"

	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/store"
)

// QRFormat selects what payload is encoded inside the rendered PNG QR
// for `type=QR` payments. The wire shape stays the same — `qr_data` is
// always a base64-encoded PNG — only the content the QR points at
// differs.
type QRFormat string

const (
	// QRFormatJSON renders a QR whose payload is a small JSON object
	// carrying the mock pay-page URL plus the payment essentials
	// (reference, amount, currency, merchant name). Custom scanners
	// can show "Pay 42.50 MVR to <merchant>" without fetching the URL,
	// and fall back to opening `url` for the full mock pay page. This
	// is the developer-friendly default — opt out via --qr-format
	// emvco when you need production-shape EMVCo.
	QRFormatJSON QRFormat = "json"
	// QRFormatEMVCo renders a QR that encodes an EMVCo Merchant
	// Presented Mode (MPM) string with a valid CRC-16/CCITT-FALSE
	// checksum. This matches the production API's wire shape — opt in
	// via `swipe mock start --qr-format emvco` when testing bank-app
	// scanning paths.
	QRFormatEMVCo QRFormat = "emvco"
)

// DefaultQRFormat is the mock's default. The real Swipe Merchants API
// ships EMVCo; the mock chooses JSON for dev usability and lists it on
// the parity whitelist as a deliberate divergence.
const DefaultQRFormat = QRFormatJSON

// qrPNGSize is the rendered PNG width/height in pixels for the QR
// returned in PaymentResponse.qr_data. 256 is small enough to keep the
// base64 payload compact (~3 KB) but large enough that consumer apps
// can display it without upscaling artefacts. Error correction level
// is Medium — matches the EMVCo MPM recommendation for merchant-
// presented codes and is generous enough for short URLs too.
const qrPNGSize = 256

// renderQRPNG renders the given string content into a PNG QR image.
// The content can be either the EMV MPM payload or a URL; the QR
// encoding doesn't care, the consumer scanner does.
func renderQRPNG(content string) ([]byte, error) {
	png, err := qrcode.Encode(content, qrcode.Medium, qrPNGSize)
	if err != nil {
		return nil, fmt.Errorf("render qr png: %w", err)
	}
	return png, nil
}

// EMVCo Merchant Presented Mode (MPM) QR Code Specification fields the
// mock emits. The mock generates a structurally-valid EMVCo MPM string
// with a real CRC16-CCITT-FALSE checksum so EMV parsers accept it. The
// merchant account info field is a placeholder GUID — scanning with the
// real Swipe app won't process it (the prod backend doesn't know this
// payment), but anything that validates EMV structure before processing
// will pass cleanly.
const (
	emvIDPayloadFormat            = "00" // value: "01"
	emvIDPointOfInitiation        = "01" // value: "11" static / "12" dynamic
	emvIDMerchantAccountInfoStart = "26" // 26..51 — first available domestic slot
	emvIDMerchantCategoryCode     = "52" // value: 4-digit MCC ("0000" = unspecified)
	emvIDTransactionCurrency      = "53" // value: 3-digit ISO 4217 numeric ("462" = MVR, "840" = USD)
	emvIDTransactionAmount        = "54" // value: amount as string, max 13 digits incl. decimal point
	emvIDCountryCode              = "58" // value: 2-char ISO 3166-1 alpha-2 ("MV")
	emvIDMerchantName             = "59" // value: human-readable merchant name
	emvIDMerchantCity             = "60" // value: human-readable city
	emvIDCRC                      = "63" // value: 4-hex CRC-16/CCITT-FALSE over preceding TLVs + "6304"
)

// emvCurrencyCode maps the mock's supported currencies to their ISO 4217
// numeric codes per the EMVCo MPM spec field 53.
func emvCurrencyCode(c store.Currency) string {
	switch c {
	case store.CurrencyMVR:
		return "462"
	case store.CurrencyUSD:
		return "840"
	default:
		return "999" // ISO 4217: "no currency"
	}
}

// buildEMVCoQR builds an EMVCo MPM payload for a payment. The merchant
// info under tag 26 is a domestic-acquirer placeholder containing a
// reverse-DNS-style globally-unique identifier and the payment's short
// code so a scanning system can correlate the QR back to the payment
// without sniffing the database.
//
// Returns the raw EMVCo string (not base64-encoded — caller decides on
// transport encoding).
func buildEMVCoQR(amount float64, currency store.Currency, shortCode, merchantName string) string {
	var sb byteBuilder

	// 00: Payload Format Indicator — fixed "01" per EMVCo MPM.
	sb.tlv(emvIDPayloadFormat, "01")
	// 01: Point of Initiation Method — "12" = dynamic (each payment is one-time).
	sb.tlv(emvIDPointOfInitiation, "12")
	// 26: Merchant Account Information — sub-TLVs:
	//     00: Globally Unique Identifier (reverse-DNS for the mock).
	//     01: merchant-side reference (the payment short code).
	merchantInfo := tlv("00", "mv.swipe.mock") + tlv("01", shortCode)
	sb.tlv(emvIDMerchantAccountInfoStart, merchantInfo)
	// 52: Merchant Category Code — "0000" unspecified.
	sb.tlv(emvIDMerchantCategoryCode, "0000")
	// 53: Transaction Currency.
	sb.tlv(emvIDTransactionCurrency, emvCurrencyCode(currency))
	// 54: Transaction Amount. Two-decimal fixed-point, no trailing zero
	// trimming since EMVCo treats the field as a string.
	sb.tlv(emvIDTransactionAmount, strconv.FormatFloat(amount, 'f', 2, 64))
	// 58: Country Code.
	sb.tlv(emvIDCountryCode, "MV")
	// 59: Merchant Name. Trimmed to EMVCo max length (25) to be safe.
	sb.tlv(emvIDMerchantName, truncate(merchantName, 25))
	// 60: Merchant City — mock seed lives in Malé.
	sb.tlv(emvIDMerchantCity, "MALE")
	// 63: CRC — appended last. The CRC is computed over the entire
	// preceding payload INCLUDING the "6304" prefix (the tag + length of
	// the CRC field itself), per EMVCo Annex B.
	body := sb.String()
	crcInput := body + emvIDCRC + "04"
	crc := crc16CCITTFalse([]byte(crcInput))
	sb.tlv(emvIDCRC, fmt.Sprintf("%04X", crc))
	return sb.String()
}

// tlv formats a single EMV TLV: 2-digit tag + 2-digit decimal length +
// value. Length is the byte length of the value as a base-10 string.
func tlv(id, value string) string {
	return id + fmt.Sprintf("%02d", len(value)) + value
}

// byteBuilder is a thin wrapper around strings.Builder that exposes a
// tlv-append shortcut.
type byteBuilder struct {
	b []byte
}

func (s *byteBuilder) tlv(id, value string) {
	s.b = append(s.b, tlv(id, value)...)
}

func (s *byteBuilder) String() string {
	return string(s.b)
}

// crc16CCITTFalse computes the CRC-16/CCITT-FALSE checksum (poly 0x1021,
// init 0xFFFF, no input/output reflection, no XOR-out) over data. This
// is the variant mandated by EMVCo MPM Annex B for the QR Code
// Specification's CRC field (tag 63).
func crc16CCITTFalse(data []byte) uint16 {
	crc := uint16(0xFFFF)
	for _, b := range data {
		crc ^= uint16(b) << 8
		for i := 0; i < 8; i++ {
			if crc&0x8000 != 0 {
				crc = (crc << 1) ^ 0x1021
			} else {
				crc <<= 1
			}
		}
	}
	return crc
}

// truncate clips s to at most n runes, returning the prefix.
func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
