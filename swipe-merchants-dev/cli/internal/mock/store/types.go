package store

import "time"

// Merchant represents a merchant tenant inside the mock. The default seed
// creates one (mer_default). Phase 2's `keys create` will associate new
// OAuth clients with merchants.
type Merchant struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

// Wallet holds per-currency available and pending balances for a merchant.
// One wallet exists per merchant; multi-currency support is built in.
type Wallet struct {
	MerchantID string                       `json:"merchant_id"`
	ID         string                       `json:"id"`
	Balances   map[Currency]CurrencyBalance `json:"balances"`
	UpdatedAt  time.Time                    `json:"updated_at"`
}

// Currency identifies a wallet currency. The mock seeds MVR and USD; the
// spec enumerates these as the only supported currencies for V1.
type Currency string

// Supported currencies, per spec/app.yaml CreatePaymentRequest.currency enum.
const (
	CurrencyMVR Currency = "MVR"
	CurrencyUSD Currency = "USD"
)

// CurrencyBalance is the available + pending pair for a single currency.
type CurrencyBalance struct {
	Available float64 `json:"available_balance"`
	Pending   float64 `json:"pending_balance"`
}

// BankAccountStatus mirrors the spec's BankAccountResponse.status enum.
type BankAccountStatus string

// Bank account status values per spec/app.yaml.
const (
	BankAccountActive    BankAccountStatus = "ACTIVE"
	BankAccountInactive  BankAccountStatus = "INACTIVE"
	BankAccountSuspended BankAccountStatus = "SUSPENDED"
	BankAccountClosed    BankAccountStatus = "CLOSED"
)

// BankAccount represents a merchant's linked bank account, the destination
// for `createPayout` requests in later phases.
type BankAccount struct {
	ID                string            `json:"id"`
	MerchantID        string            `json:"merchant_id"`
	AccountNumber     string            `json:"account_number"`
	AccountHolderName string            `json:"account_holder_name"`
	Currency          Currency          `json:"currency"`
	Status            BankAccountStatus `json:"status"`
	CreatedAt         time.Time         `json:"created_at"`
}

// PaymentType enumerates the spec's CreatePaymentRequest.type values.
type PaymentType string

// Payment types.
const (
	PaymentTypeQR      PaymentType = "QR"
	PaymentTypeContact PaymentType = "CONTACT"
	PaymentTypeLink    PaymentType = "LINK"
)

// PaymentStatus enumerates the spec's PaymentResponse.status values.
type PaymentStatus string

// Payment status values.
const (
	PaymentPending   PaymentStatus = "PENDING"
	PaymentCompleted PaymentStatus = "COMPLETED"
	PaymentExpired   PaymentStatus = "EXPIRED"
	PaymentCancelled PaymentStatus = "CANCELLED"
)

// Payment is the persisted shape of a created payment.
type Payment struct {
	ID           string        `json:"id"`
	MerchantID   string        `json:"merchant_id"`
	Amount       float64       `json:"amount"`
	Currency     Currency      `json:"currency"`
	Type         PaymentType   `json:"type"`
	Status       PaymentStatus `json:"status"`
	ShortCode    string        `json:"short_code"`
	QRData       string        `json:"qr_data,omitempty"`
	PaymentURL   string        `json:"payment_url,omitempty"`
	RecipientVPA string        `json:"recipient_vpa,omitempty"`
	Description  string        `json:"description,omitempty"`
	CreatedAt    time.Time     `json:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at"`
	// TransitionAt is the scheduled time for the default
	// PENDING -> COMPLETED transition. Zero means no auto-transition.
	TransitionAt time.Time `json:"transition_at,omitempty"`
	// TransitionTo is the target status when TransitionAt is reached.
	// Defaults to COMPLETED unless a scenario overrides at create time.
	TransitionTo PaymentStatus `json:"transition_to,omitempty"`
}

// PayoutStatus enumerates the payout lifecycle states. The spec's
// PayoutResponse.status is open-ended so we mirror the payment enum.
type PayoutStatus string

// Payout status values.
const (
	PayoutPending   PayoutStatus = "PENDING"
	PayoutCompleted PayoutStatus = "COMPLETED"
	PayoutFailed    PayoutStatus = "FAILED"
)

// Payout is the persisted shape of an initiated withdrawal.
type Payout struct {
	ID            string       `json:"id"`
	MerchantID    string       `json:"merchant_id"`
	Amount        float64      `json:"amount"`
	Currency      Currency     `json:"currency"`
	BankAccountID string       `json:"bank_account_id"`
	Status        PayoutStatus `json:"status"`
	Reference     string       `json:"reference,omitempty"`
	CreatedAt     time.Time    `json:"created_at"`
	UpdatedAt     time.Time    `json:"updated_at"`
}

// Transaction is the unified ledger row that aggregates payments, payouts,
// and (later) adjustments behind `GET /api/v1/history`. The fields mirror
// the spec's TransactionItem.
//
// The four monetary fields mirror the upstream spec: GrossAmount is the
// pre-fee amount, FeeAmount is the platform's deduction, NetAmount is
// what reaches the merchant, OriginalAmount is the source currency
// amount before FX. The mock has no fee model and no FX, so all four
// default to `Amount` except FeeAmount which is 0.
type Transaction struct {
	ID             string    `json:"id"`
	Reference      string    `json:"reference,omitempty"`
	Amount         float64   `json:"amount"`
	GrossAmount    float64   `json:"gross_amount,omitempty"`
	FeeAmount      float64   `json:"fee_amount,omitempty"`
	NetAmount      float64   `json:"net_amount,omitempty"`
	OriginalAmount float64   `json:"original_amount,omitempty"`
	Currency       Currency  `json:"currency"`
	Type           string    `json:"type"`   // "PAYMENT", "PAYOUT", future "ADJUSTMENT"
	Status         string    `json:"status"` // mirrors source status
	Description    string    `json:"description,omitempty"`
	MerchantID     string    `json:"merchant_id"`
	SourceID       string    `json:"source_id"` // pay_<ulid> or pyo_<ulid>
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
