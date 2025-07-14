package tonamount

import (
	"encoding/json"
	"fmt"
	"math/big"
	"strings"

	"github.com/shopspring/decimal"
)

const (
	tonPricePrecision = 2
	tonNanoScale      = 9
)

//nolint:gochecknoglobals // required for ton amount
var nanoMultiplier = decimal.New(1, tonNanoScale) // 1e9

// TonAmount stores.
type TonAmount struct {
	d decimal.Decimal
}

// NewTonAmountFromString creates a TonAmount from a string, rounds it to tonPricePrecision and checks for negativity.
func NewTonAmountFromString(s string) (*TonAmount, error) {
	d, err := decimal.NewFromString(s)
	if err != nil {
		return nil, fmt.Errorf("invalid decimal string %q: %w", s, err)
	}
	d = d.Round(tonPricePrecision)
	if d.IsNegative() {
		return nil, ErrTonAmountNegative
	}
	return &TonAmount{d: d}, nil
}

// NewTonAmountFromFloat64 creates a TonAmount from a float64, rounds it to tonPricePrecision and checks for negativity.
func NewTonAmountFromFloat64(f float64) (*TonAmount, error) {
	d := decimal.NewFromFloat(f)
	d = d.Round(tonPricePrecision)
	if d.IsNegative() {
		return nil, ErrTonAmountNegative
	}
	return &TonAmount{d: d}, nil
}

// NewTonAmountFromNano creates a TonAmount from nano units (1 TON = 1e9 nano).
func NewTonAmountFromNano(nano uint64) (*TonAmount, error) {
	bi := new(big.Int).SetUint64(nano)
	d := decimal.NewFromBigInt(bi, -tonNanoScale).Round(tonPricePrecision)
	return &TonAmount{d: d}, nil
}

// Decimal returns the internal decimal.Decimal (or zero if t == nil).
func (t *TonAmount) Decimal() decimal.Decimal {
	if t == nil {
		return decimal.Zero
	}
	return t.d
}

// String returns a human-readable representation without trailing zeros and dots.
// Examples: "1.20" → "1.2", "2.00" → "2", "1.234" → "1.23".
func (t *TonAmount) String() string {
	d := t.Decimal()
	s := d.StringFixed(tonPricePrecision)
	s = strings.TrimRight(s, "0")
	s = strings.TrimRight(s, ".")
	return s
}

// MarshalJSON returns a JSON representation of the TonAmount.
func (t *TonAmount) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.String())
}

// UnmarshalJSON can read TON from JSON number or string.
func (t *TonAmount) UnmarshalJSON(data []byte) error {
	str := strings.Trim(string(data), `"`)
	d, err := decimal.NewFromString(str)
	if err != nil {
		return err
	}
	d = d.Round(tonPricePrecision)
	if d.IsNegative() {
		return ErrTonAmountNegative
	}
	t.d = d
	return nil
}

// Add returns the sum of t + o.
func (t *TonAmount) Add(o *TonAmount) *TonAmount {
	return &TonAmount{d: t.Decimal().Add(o.Decimal())}
}

// Sub returns the difference of t - o.
func (t *TonAmount) Sub(o *TonAmount) *TonAmount {
	return &TonAmount{d: t.Decimal().Sub(o.Decimal())}
}

// IsZero checks if the value is zero.
func (t *TonAmount) IsZero() bool {
	return t.Decimal().IsZero()
}

// Negate returns the negative of the TonAmount.
func (t *TonAmount) Negate() *TonAmount {
	return &TonAmount{d: t.Decimal().Neg()}
}

// ToNano converts the TonAmount to nano units (1 TON = 1e9 nano).
func (t *TonAmount) ToNano() (uint64, error) {
	nanos := t.d.Mul(nanoMultiplier).Round(0)
	bi := nanos.BigInt()
	if bi.Sign() < 0 {
		return 0, fmt.Errorf("negative TonAmount: %s", t.d)
	}
	if !bi.IsUint64() {
		return 0, fmt.Errorf("overflow converting %s to nano", t.d)
	}
	return bi.Uint64(), nil
}
