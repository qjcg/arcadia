package rq

import (
	"fmt"
	"time"

	"github.com/qjcg/arcadia/exp/fin/internal/money"
	"github.com/shopspring/decimal"
)

// RQRules implements Revenu Québec-specific penalty and interest calculations
type RQRules struct{}

// NewRQRules returns a new Revenu Québec rules instance
func NewRQRules() *RQRules {
	return &RQRules{}
}

// CalculateLateFilingPenalty calculates the late filing penalty for RQ
// Revenu Québec charges a penalty similar to CRA but with Quebec-specific rules
func (r *RQRules) CalculateLateFilingPenalty(taxAmount money.Money, dueDate, filingDate time.Time, hadBalanceLastYear bool) money.Money {
	if !hadBalanceLastYear {
		return decimal.Zero
	}

	fivePercent := decimal.NewFromFloat(0.05)
	onePercent := decimal.NewFromFloat(0.01)

	basePenalty := taxAmount.Mul(fivePercent)

	monthsLate := max(min(r.MonthsLate(dueDate, filingDate), 12), 0)

	additionalPenalty := taxAmount.Mul(onePercent).Mul(decimal.NewFromInt(int64(monthsLate)))
	return basePenalty.Add(additionalPenalty)
}

// MonthsLate calculates the number of months late (rounded down)
func (r *RQRules) MonthsLate(dueDate, filingDate time.Time) int {
	if filingDate.Before(dueDate) || filingDate.Equal(dueDate) {
		return 0
	}

	years := filingDate.Year() - dueDate.Year()
	months := int(filingDate.Month()) - int(dueDate.Month())
	return years*12 + months
}

// CalculateInterest calculates daily compounded interest on amount owed
// Revenu Québec compounds interest daily at the prescribed rate
func (r *RQRules) CalculateInterest(amount money.Money, startDate, endDate time.Time, dailyRate float64) money.Money {
	if amount.LessThanOrEqual(decimal.Zero) || endDate.Before(startDate) {
		return decimal.Zero
	}

	days := int(endDate.Sub(startDate).Hours()/24) + 1
	if days < 0 {
		return decimal.Zero
	}

	// Daily rate = annual rate / 365 (or 366 for leap years)
	// Interest is compounded daily: principal * ((1 + rate/365)^days - 1)
	rate := decimal.NewFromFloat(dailyRate / 100)
	one := decimal.NewFromInt(1)
	dailyDivisor := decimal.NewFromInt(365)

	// Calculate (1 + rate/365)^days using decimal arithmetic
	factor := one
	dailyRateDecimal := rate.Div(dailyDivisor)
	for range days {
		factor = factor.Mul(one.Add(dailyRateDecimal))
	}

	interest := amount.Mul(factor.Sub(one))
	return interest
}

// GetPrescribedRateSource returns the RQ official source URL
func (r *RQRules) GetPrescribedRateSource() string {
	return "https://www.revenuquebec.ca/en/one-mission-concrete-actions/ensuring-tax-compliance/penalties-and-interest/interest-rates-on-debts/"
}

// GetLateFilingPenaltyInfo returns human-readable penalty info
func (r *RQRules) GetLateFilingPenaltyInfo() string {
	return "5% of balance owing + 1% per month late (max 12 months = 17% total)"
}

// ValidateDate checks if date is valid for RQ calculations
func (r *RQRules) ValidateDate(t time.Time) error {
	if t.Before(time.Date(1996, 1, 1, 0, 0, 0, 0, time.UTC)) {
		return fmt.Errorf("Revenu Québec rates not available before 1996")
	}
	return nil
}
