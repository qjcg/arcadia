package internal

import (
	"fmt"
	"time"
)

// Calculator handles penalty and interest calculations
type Calculator struct {
	rateDB *RateDB
	cra    *CRARules
	rq     *RQRules
}

// CalculationInput contains all inputs for a calculation
type CalculationInput struct {
	TaxAmount        float64
	DueDate          time.Time
	PaymentDate     time.Time
	Jurisdiction    Jurisdiction
	HadBalanceLastYear bool
}

// CalculationResult contains the breakdown of penalties and interest
type CalculationResult struct {
	TaxAmount           float64   `json:"tax_amount"`
	DueDate            time.Time `json:"due_date"`
	PaymentDate        time.Time `json:"payment_date"`
	Jurisdiction       string    `json:"jurisdiction"`
	LateFilingPenalty  float64   `json:"late_filing_penalty"`
	Interest           float64   `json:"interest"`
	TotalAmount        float64   `json:"total_amount"`
	DaysOwed           int       `json:"days_owed"`
	EffectiveRate      float64   `json:"effective_rate"`
	LateFilingMonths   int       `json:"late_filing_months"`
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
		TaxAmount:     inp.TaxAmount,
		DueDate:       inp.DueDate,
		PaymentDate:   inp.PaymentDate,
		Jurisdiction:  string(inp.Jurisdiction),
		DaysOwed:      c.daysOwed(inp.DueDate, inp.PaymentDate),
	}

	// Calculate late filing penalty if had balance last year
	if inp.HadBalanceLastYear {
		result.LateFilingPenalty = c.calculatePenalty(inp)
		result.LateFilingMonths = c.getMonthsLate(inp)
		if result.LateFilingMonths < 0 {
			result.LateFilingMonths = 0
		}
		if result.LateFilingMonths > 12 {
			result.LateFilingMonths = 12
		}
	}

	// Calculate interest
	result.Interest = c.calculateInterest(inp)

	// Calculate total
	result.TotalAmount = inp.TaxAmount + result.LateFilingPenalty + result.Interest

	// Calculate effective rate for display
	if result.DaysOwed > 0 && result.TaxAmount > 0 {
		result.EffectiveRate = (result.Interest / result.TaxAmount) / float64(result.DaysOwed) * 365 * 100
	}

	return result, nil
}

func (c *Calculator) validateInput(inp CalculationInput) error {
	if inp.TaxAmount < 0 {
		return fmt.Errorf("tax amount cannot be negative")
	}
	if inp.PaymentDate.Before(inp.DueDate) {
		return fmt.Errorf("payment date cannot be before due date")
	}
	if inp.Jurisdiction != CRA && inp.Jurisdiction != RQ {
		return fmt.Errorf("invalid jurisdiction: %s (must be 'cra' or 'rq')", inp.Jurisdiction)
	}
	return nil
}

func (c *Calculator) daysOwed(dueDate, paymentDate time.Time) int {
	days := paymentDate.Sub(dueDate).Hours() / 24
	if days < 0 {
		return 0
	}
	return int(days)
}

func (c *Calculator) getMonthsLate(inp CalculationInput) int {
	switch inp.Jurisdiction {
	case CRA:
		return inp.PaymentDate.Year() - inp.DueDate.Year()*12 + int(inp.PaymentDate.Month()) - int(inp.DueDate.Month())
	case RQ:
		return inp.PaymentDate.Year() - inp.DueDate.Year()*12 + int(inp.PaymentDate.Month()) - int(inp.DueDate.Month())
	default:
		return 0
	}
}

func (c *Calculator) calculatePenalty(inp CalculationInput) float64 {
	switch inp.Jurisdiction {
	case CRA:
		return c.cra.CalculateLateFilingPenalty(inp.TaxAmount, inp.DueDate, inp.PaymentDate, inp.HadBalanceLastYear)
	case RQ:
		return c.rq.CalculateLateFilingPenalty(inp.TaxAmount, inp.DueDate, inp.PaymentDate, inp.HadBalanceLastYear)
	default:
		return 0
	}
}

func (c *Calculator) calculateInterest(inp CalculationInput) float64 {
	if inp.PaymentDate.Before(inp.DueDate) {
		return 0
	}

	// Get prescribed rate for the payment date
	rate, err := c.rateDB.GetPrescribedRateForDate(inp.Jurisdiction, inp.PaymentDate)
	if err != nil {
		// If we can't get the rate, return 0 rather than failing
		return 0
	}

	switch inp.Jurisdiction {
	case CRA:
		return c.cra.CalculateInterest(inp.TaxAmount, inp.DueDate, inp.PaymentDate, rate)
	case RQ:
		return c.rq.CalculateInterest(inp.TaxAmount, inp.DueDate, inp.PaymentDate, rate)
	default:
		return 0
	}
}

// GetRateDB returns the underlying rate database
func (c *Calculator) GetRateDB() *RateDB {
	return c.rateDB
}