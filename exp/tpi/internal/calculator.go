package internal

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
)

// Calculator handles penalty and interest calculations
type Calculator struct {
	rateDB *RateDB
	cra    *CRARules
	rq     *RQRules
}

// CalculationInput contains all inputs for a calculation
type CalculationInput struct {
	Year                int       `json:"year"`
	Earned              Money     `json:"earned"`
	BaseDueCRA          Money     `json:"base_due_cra"`
	BaseDueRQ           Money     `json:"base_due_rqc"`
	ExpectedFilingDate  time.Time `json:"expected_filing_date"`
	ExpectedPaymentDate time.Time `json:"expected_payment_date"`
	ActualFilingDate    time.Time `json:"actual_filing_date"`
	ActualPaymentDate   time.Time `json:"actual_payment_date"`
	HadBalanceLastYear  bool      `json:"had_balance_last_year"`
}

// CalculationResult contains the breakdown of penalties and interest
type CalculationResult struct {
	Year                int       `json:"year"`
	Earned              Money     `json:"earned"`
	BaseDueCRA          Money     `json:"base_due_cra"`
	BaseDueRQ           Money     `json:"base_due_rqc"`
	PenaltiesCRA        Money     `json:"penalties_cra"`
	InterestCRA         Money     `json:"interest_cra"`
	PenaltiesRQ         Money     `json:"penalties_rqc"`
	InterestRQ          Money     `json:"interest_rqc"`
	TotalDueCRA         Money     `json:"total_due_cra"`
	TotalDueRQ          Money     `json:"total_due_rqc"`
	TotalDue            Money     `json:"total_due"`
	ExpectedFilingDate  time.Time `json:"expected_filing_date"`
	ExpectedPaymentDate time.Time `json:"expected_payment_date"`
	ActualFilingDate    time.Time `json:"actual_filing_date"`
	ActualPaymentDate   time.Time `json:"actual_payment_date"`
}

// NewCalculator creates a new calculator with the embedded rates database
func NewCalculator() (*Calculator, error) {
	rateDB, err := NewRateDB()
	if err != nil {
		return nil, err
	}
	return &Calculator{
		rateDB: rateDB,
		cra:    NewCRARules(),
		rq:     NewRQRules(),
	}, nil
}

// NewCalculatorWithDB creates a calculator with a custom rates database
func NewCalculatorWithDB(rateDB *RateDB) *Calculator {
	return &Calculator{
		rateDB: rateDB,
		cra:    NewCRARules(),
		rq:     NewRQRules(),
	}
}

// Calculate computes penalties and interest based on input
func (c *Calculator) Calculate(inp CalculationInput) (*CalculationResult, error) {
	if err := c.validateInput(inp); err != nil {
		return nil, err
	}

	result := &CalculationResult{
		Year:                inp.Year,
		Earned:              inp.Earned,
		BaseDueCRA:          inp.BaseDueCRA,
		BaseDueRQ:           inp.BaseDueRQ,
		ExpectedFilingDate:  inp.ExpectedFilingDate,
		ExpectedPaymentDate: inp.ExpectedPaymentDate,
		ActualFilingDate:    inp.ActualFilingDate,
		ActualPaymentDate:   inp.ActualPaymentDate,
	}

	// Calculate CRA penalties and interest
	if inp.BaseDueCRA.GreaterThan(decimal.Zero) {
		result.PenaltiesCRA = c.calculatePenalty(inp.BaseDueCRA, inp.ExpectedFilingDate, inp.ActualFilingDate, inp.HadBalanceLastYear, CRA)
		result.InterestCRA = c.calculateInterest(inp.BaseDueCRA.Add(result.PenaltiesCRA), inp.ExpectedPaymentDate, inp.ActualPaymentDate, CRA)
		result.TotalDueCRA = inp.BaseDueCRA.Add(result.PenaltiesCRA).Add(result.InterestCRA)
	}

	// Calculate RQ penalties and interest
	if inp.BaseDueRQ.GreaterThan(decimal.Zero) {
		result.PenaltiesRQ = c.calculatePenalty(inp.BaseDueRQ, inp.ExpectedFilingDate, inp.ActualFilingDate, inp.HadBalanceLastYear, RQ)
		result.InterestRQ = c.calculateInterest(inp.BaseDueRQ.Add(result.PenaltiesRQ), inp.ExpectedPaymentDate, inp.ActualPaymentDate, RQ)
		result.TotalDueRQ = inp.BaseDueRQ.Add(result.PenaltiesRQ).Add(result.InterestRQ)
	}

	// Calculate combined total
	result.TotalDue = result.TotalDueCRA.Add(result.TotalDueRQ)

	return result, nil
}

func (c *Calculator) validateInput(inp CalculationInput) error {
	if inp.BaseDueCRA.LessThan(decimal.Zero) {
		return fmt.Errorf("base due CRA cannot be negative")
	}
	if inp.BaseDueRQ.LessThan(decimal.Zero) {
		return fmt.Errorf("base due RQ cannot be negative")
	}
	return nil
}

func (c *Calculator) calculatePenalty(taxAmount Money, expectedDate, actualDate time.Time, hadBalanceLastYear bool, j Jurisdiction) Money {
	switch j {
	case CRA:
		return c.cra.CalculateLateFilingPenalty(taxAmount, expectedDate, actualDate, hadBalanceLastYear)
	case RQ:
		return c.rq.CalculateLateFilingPenalty(taxAmount, expectedDate, actualDate, hadBalanceLastYear)
	default:
		return decimal.Zero
	}
}

func (c *Calculator) calculateInterest(taxAmount Money, expectedDate, actualDate time.Time, j Jurisdiction) Money {
	if actualDate.Before(expectedDate) {
		return decimal.Zero
	}

	rate, err := c.rateDB.GetPrescribedRateForDate(j, actualDate)
	if err != nil {
		return decimal.Zero
	}

	switch j {
	case CRA:
		return c.cra.CalculateInterest(taxAmount, expectedDate, actualDate, rate)
	case RQ:
		return c.rq.CalculateInterest(taxAmount, expectedDate, actualDate, rate)
	default:
		return decimal.Zero
	}
}

// GetRateDB returns the underlying rate database
func (c *Calculator) GetRateDB() *RateDB {
	return c.rateDB
}
