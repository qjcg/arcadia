package internal

import (
	"fmt"
	"time"
)

// RQRules implements Revenu Québec-specific penalty and interest calculations
type RQRules struct{}

// NewRQRules returns a new Revenu Québec rules instance
func NewRQRules() *RQRules {
	return &RQRules{}
}

// CalculateLateFilingPenalty calculates the late filing penalty for RQ
// Revenu Québec charges a penalty similar to CRA but with Quebec-specific rules
func (r *RQRules) CalculateLateFilingPenalty(taxAmount float64, dueDate, filingDate time.Time, hadBalanceLastYear bool) float64 {
	if !hadBalanceLastYear {
		return 0
	}

	basePenalty := 0.05 * taxAmount

	monthsLate := r.MonthsLate(dueDate, filingDate)
	if monthsLate > 12 {
		monthsLate = 12
	}
	if monthsLate < 0 {
		monthsLate = 0
	}

	additionalPenalty := 0.01 * float64(monthsLate) * taxAmount
	return basePenalty + additionalPenalty
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
func (r *RQRules) CalculateInterest(amount float64, startDate, endDate time.Time, dailyRate float64) float64 {
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
	for i := 0; i < days; i++ {
		factor *= 1.0 + rate/365.0
	}

	interest := amount * (factor - 1.0)
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