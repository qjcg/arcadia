package internal

import (
	"fmt"
	"time"
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
func (c *CRARules) CalculateLateFilingPenalty(taxAmount float64, dueDate, filingDate time.Time, hadBalanceLastYear bool) float64 {
	if !hadBalanceLastYear {
		return 0
	}

	basePenalty := 0.05 * taxAmount

	monthsLate := max(min(c.MonthsLate(dueDate, filingDate), 12), 0)

	additionalPenalty := 0.01 * float64(monthsLate) * taxAmount
	return basePenalty + additionalPenalty
}

// MonthsLate calculates the number of months late (rounded down)
func (c *CRARules) MonthsLate(dueDate, filingDate time.Time) int {
	if filingDate.Before(dueDate) || filingDate.Equal(dueDate) {
		return 0
	}

	years := filingDate.Year() - dueDate.Year()
	months := int(filingDate.Month()) - int(dueDate.Month())
	return years*12 + months
}

// CalculateInterest calculates daily compounded interest on amount owed
// CRA compounds interest daily at the prescribed rate
func (c *CRARules) CalculateInterest(amount float64, startDate, endDate time.Time, dailyRate float64) float64 {
	if amount <= 0 || endDate.Before(startDate) {
		return 0
	}

	days := int(endDate.Sub(startDate).Hours()/24) + 1
	if days < 0 {
		return 0
	}

	// Daily rate = annual rate / 365 (or 366 for leap years)
	// Interest is compounded daily: principal * ((1 + rate/365)^days - 1)
	rate := dailyRate / 100
	factor := 1.0
	for range days {
		factor *= 1.0 + rate/365.0
	}

	interest := amount * (factor - 1.0)
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
