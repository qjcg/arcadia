package calculator

import (
	"fmt"
	"time"

	"github.com/qjcg/arcadia/exp/fin/internal/money"
	"github.com/qjcg/arcadia/exp/fin/internal/rates"
	"github.com/qjcg/arcadia/exp/fin/internal/rules"
	"github.com/qjcg/arcadia/exp/fin/internal/rules/cra"
	"github.com/qjcg/arcadia/exp/fin/internal/rules/rq"
	"github.com/shopspring/decimal"
)

// Calculator handles penalty and interest calculations
type Calculator struct {
	rateDB  *rates.RateDB
	craRule rules.Interface
	rqRule  rules.Interface
}

// CalculationInput contains all inputs for a calculation
type CalculationInput struct {
	Year                int         `json:"year"`
	Earned              money.Money `json:"earned"`
	BaseDueCRA          money.Money `json:"base_due_cra"`
	BaseDueRQ           money.Money `json:"base_due_rqc"`
	ExpectedFilingDate  time.Time   `json:"expected_filing_date"`
	ExpectedPaymentDate time.Time   `json:"expected_payment_date"`
	ActualFilingDate    time.Time   `json:"actual_filing_date"`
	ActualPaymentDate   time.Time   `json:"actual_payment_date"`
	HadBalanceLastYear  bool        `json:"had_balance_last_year"`
}

// CalculationResult contains the breakdown of penalties and interest
type CalculationResult struct {
	Year                int         `json:"year"`
	Earned              money.Money `json:"earned"`
	BaseDueCRA          money.Money `json:"base_due_cra"`
	BaseDueRQ           money.Money `json:"base_due_rqc"`
	PenaltiesCRA        money.Money `json:"penalties_cra"`
	InterestCRA         money.Money `json:"interest_cra"`
	PenaltiesRQ         money.Money `json:"penalties_rqc"`
	InterestRQ          money.Money `json:"interest_rqc"`
	TotalDueCRA         money.Money `json:"total_due_cra"`
	TotalDueRQ          money.Money `json:"total_due_rqc"`
	TotalDue            money.Money `json:"total_due"`
	ExpectedFilingDate  time.Time   `json:"expected_filing_date"`
	ExpectedPaymentDate time.Time   `json:"expected_payment_date"`
	ActualFilingDate    time.Time   `json:"actual_filing_date"`
	ActualPaymentDate   time.Time   `json:"actual_payment_date"`
}

// NewCalculator creates a new calculator with the embedded rates database
func NewCalculator() (*Calculator, error) {
	rateDB, err := rates.NewRateDB()
	if err != nil {
		return nil, err
	}
	return &Calculator{
		rateDB:  rateDB,
		craRule: cra.NewCRARules(),
		rqRule:  rq.NewRQRules(),
	}, nil
}

// NewCalculatorWithDB creates a calculator with a custom rates database
func NewCalculatorWithDB(rateDB *rates.RateDB) *Calculator {
	return &Calculator{
		rateDB:  rateDB,
		craRule: cra.NewCRARules(),
		rqRule:  rq.NewRQRules(),
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
		result.PenaltiesCRA = c.craRule.CalculateLateFilingPenalty(inp.BaseDueCRA, inp.ExpectedFilingDate, inp.ActualFilingDate, inp.HadBalanceLastYear)
		result.InterestCRA = c.calculateInterest(inp.BaseDueCRA.Add(result.PenaltiesCRA), inp.ExpectedPaymentDate, inp.ActualPaymentDate, c.craRule, rules.CRA)
		result.TotalDueCRA = inp.BaseDueCRA.Add(result.PenaltiesCRA).Add(result.InterestCRA)
	}

	// Calculate RQ penalties and interest
	if inp.BaseDueRQ.GreaterThan(decimal.Zero) {
		result.PenaltiesRQ = c.rqRule.CalculateLateFilingPenalty(inp.BaseDueRQ, inp.ExpectedFilingDate, inp.ActualFilingDate, inp.HadBalanceLastYear)
		result.InterestRQ = c.calculateInterest(inp.BaseDueRQ.Add(result.PenaltiesRQ), inp.ExpectedPaymentDate, inp.ActualPaymentDate, c.rqRule, rules.RQ)
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

func (c *Calculator) calculateInterest(taxAmount money.Money, expectedDate, actualDate time.Time, rule rules.Interface, j rules.Jurisdiction) money.Money {
	if actualDate.Before(expectedDate) {
		return decimal.Zero
	}

	rate, err := c.rateDB.GetPrescribedRateForDate(j, actualDate)
	if err != nil {
		return decimal.Zero
	}

	return rule.CalculateInterest(taxAmount, expectedDate, actualDate, rate)
}

// GetRateDB returns the underlying rate database
func (c *Calculator) GetRateDB() *rates.RateDB {
	return c.rateDB
}
