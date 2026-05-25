package cra

import (
	"fmt"
	"time"

	"github.com/qjcg/arcadia/exp/fc/internal/money"
	"github.com/shopspring/decimal"
)

// CRARules implements CRA-specific penalty and interest calculations
type CRARules struct{}

// NewCRARules returns a new CRA rules instance
func NewCRARules() *CRARules {
	return &CRARules{}
}

// CalculateLateFilingPenalty calculates the late filing penalty
// CRA charges 5% of the balance owing, plus an additional 1% per month of late filing
// up to a maximum of 12 months (12%) if the taxpayer had a balance owing in the previous year
func (c *CRARules) CalculateLateFilingPenalty(taxAmount money.Money, dueDate, filingDate time.Time, hadBalanceLastYear bool) money.Money {
	fivePercent := decimal.NewFromFloat(0.05)
	onePercent := decimal.NewFromFloat(0.01)

	basePenalty := taxAmount.Mul(fivePercent)

	monthsLate := max(min(c.MonthsLate(dueDate, filingDate), 12), 0)

	additionalPenalty := taxAmount.Mul(onePercent).Mul(decimal.NewFromInt(int64(monthsLate)))
	return basePenalty.Add(additionalPenalty)
}

// MonthsLate calculates the number of complete months late
// A month is considered "complete" when the filing date passes the same day
// in the following month. For example, April 30 to June 30 = 2 complete months.
func (c *CRARules) MonthsLate(dueDate, filingDate time.Time) int {
	if filingDate.Before(dueDate) || filingDate.Equal(dueDate) {
		return 0
	}

	years := filingDate.Year() - dueDate.Year()
	months := int(filingDate.Month()) - int(dueDate.Month())
	completeMonths := years*12 + months

	// If filing day is before due day, that month isn't complete yet
	if filingDate.Day() < dueDate.Day() {
		completeMonths--
	}

	if completeMonths < 0 {
		return 0
	}
	return completeMonths
}

// CalculateInterest calculates daily compounded interest on amount owed
// CRA compounds interest daily at the prescribed rate
func (c *CRARules) CalculateInterest(amount money.Money, startDate, endDate time.Time, dailyRate float64) money.Money {
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
	factor := decimal.NewFromInt(1)
	dailyRateDecimal := rate.Div(dailyDivisor)
	for range days {
		factor = factor.Mul(decimal.NewFromInt(1).Add(dailyRateDecimal))
	}

	interest := amount.Mul(factor.Sub(one))
	return interest
}

// GetPrescribedRateSource returns the CRA official source URL
func (c *CRARules) GetPrescribedRateSource() string {
	return "https://www.canada.ca/en/revenue-agency/services/tax/prescribed-interest-rates.html"
}

// GetLateFilingPenaltyInfo returns human-readable penalty info
func (c *CRARules) GetLateFilingPenaltyInfo() string {
	return "5% of balance owing + 1% per month late (max 12 months = 17% total)"
}

// ValidateDate checks if date is valid for CRA calculations
func (c *CRARules) ValidateDate(t time.Time) error {
	if t.Before(time.Date(1996, 1, 1, 0, 0, 0, 0, time.UTC)) {
		return fmt.Errorf("CRA rates not available before 1996")
	}
	return nil
}