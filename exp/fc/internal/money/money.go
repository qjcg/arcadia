package money

import (
	"github.com/shopspring/decimal"
)

// Money represents monetary values with arbitrary precision.
// Use decimal.Decimal to avoid floating-point precision errors in financial calculations.
type Money = decimal.Decimal

// NewMoney creates a Money from a float64 value.
// Note: Using float64 input may introduce imprecision; prefer string or int64 (cents) input.
func NewMoney(f float64) Money {
	return decimal.NewFromFloat(f)
}

// NewMoneyFromInt creates a Money from an integer (e.g., cents).
func NewMoneyFromInt(i int64) Money {
	return decimal.NewFromInt(i)
}

// NewMoneyFromString creates a Money from a string representation.
func NewMoneyFromString(s string) (Money, error) {
	return decimal.NewFromString(s)
}