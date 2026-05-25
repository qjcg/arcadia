package calculator

import (
	"testing"
	"time"

	"github.com/qjcg/arcadia/exp/fc/internal/money"
	"github.com/qjcg/arcadia/exp/fc/internal/rates"
	"github.com/qjcg/arcadia/exp/fc/internal/rules"
	"github.com/qjcg/arcadia/exp/fc/internal/rules/cra"
	"github.com/qjcg/arcadia/exp/fc/internal/rules/rq"
)

func TestCRARulesCalculateLateFilingPenalty(t *testing.T) {
	craRules := cra.NewCRARules()

	tests := []struct {
		name            string
		taxAmount       money.Money
		dueDate         time.Time
		filingDate      time.Time
		hadBalance      bool
		expectedPenalty money.Money
	}{
		{
			name:            "no balance last year: 5% base + monthly always applies",
			taxAmount:       money.NewMoney(5000),
			dueDate:         time.Date(2024, 4, 30, 0, 0, 0, 0, time.UTC),
			filingDate:      time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC),
			hadBalance:      false,
			expectedPenalty: money.NewMoney(300), // 5% + 1% = 6% of 5000; penalty always applies
		},
		{
			name:            "on time (still pays base 5% penalty if had balance)",
			taxAmount:       money.NewMoney(5000),
			dueDate:         time.Date(2024, 4, 30, 0, 0, 0, 0, time.UTC),
			filingDate:      time.Date(2024, 4, 30, 0, 0, 0, 0, time.UTC),
			hadBalance:      true,
			expectedPenalty: money.NewMoney(250), // 5% base penalty always applies if hadBalanceLastYear
		},
		{
			name:            "1 month late",
			taxAmount:       money.NewMoney(5000),
			dueDate:         time.Date(2024, 4, 30, 0, 0, 0, 0, time.UTC),
			filingDate:      time.Date(2024, 5, 30, 0, 0, 0, 0, time.UTC),
			hadBalance:      true,
			expectedPenalty: money.NewMoney(300), // 5% + 1% = 6% of 5000
		},
		{
			name:            "12 months late",
			taxAmount:       money.NewMoney(5000),
			dueDate:         time.Date(2024, 4, 30, 0, 0, 0, 0, time.UTC),
			filingDate:      time.Date(2025, 4, 30, 0, 0, 0, 0, time.UTC),
			hadBalance:      true,
			expectedPenalty: money.NewMoney(850), // 5% + 12% = 17% of 5000
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			penalty := craRules.CalculateLateFilingPenalty(tt.taxAmount, tt.dueDate, tt.filingDate, tt.hadBalance)
			if !penalty.Equal(tt.expectedPenalty) {
				t.Errorf("expected penalty %s, got %s", tt.expectedPenalty, penalty)
			}
		})
	}
}

func TestCRARulesMonthsLate(t *testing.T) {
	craRules := cra.NewCRARules()

	tests := []struct {
		name       string
		dueDate    time.Time
		filingDate time.Time
		expected   int
	}{
		{
			name:       "on time",
			dueDate:    time.Date(2024, 4, 30, 0, 0, 0, 0, time.UTC),
			filingDate: time.Date(2024, 4, 30, 0, 0, 0, 0, time.UTC),
			expected:   0,
		},
		{
			name:       "early",
			dueDate:    time.Date(2024, 4, 30, 0, 0, 0, 0, time.UTC),
			filingDate: time.Date(2024, 4, 20, 0, 0, 0, 0, time.UTC),
			expected:   0,
		},
		{
			name:       "1 month late",
			dueDate:    time.Date(2024, 4, 30, 0, 0, 0, 0, time.UTC),
			filingDate: time.Date(2024, 5, 30, 0, 0, 0, 0, time.UTC),
			expected:   1,
		},
		{
			name:       "1 month 1 day late",
			dueDate:    time.Date(2024, 4, 30, 0, 0, 0, 0, time.UTC),
			filingDate: time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC),
			expected:   1, // June 1 < June 30, so not a complete month
		},
		{
			name:       "cross year",
			dueDate:    time.Date(2024, 4, 30, 0, 0, 0, 0, time.UTC),
			filingDate: time.Date(2024, 6, 30, 0, 0, 0, 0, time.UTC),
			expected:   2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			months := craRules.MonthsLate(tt.dueDate, tt.filingDate)
			if months != tt.expected {
				t.Errorf("expected %d months, got %d", tt.expected, months)
			}
		})
	}
}

func TestCRARulesCalculateInterest(t *testing.T) {
	craRules := cra.NewCRARules()

	// Test daily compounding: $1000 at 4% for 1 year should be about $40.81
	// (1000 * (1.04 - 1) = 40, but with daily compounding it's slightly more)
	amount := money.NewMoney(1000)
	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)
	dailyRate := 4.0

	interest := craRules.CalculateInterest(amount, startDate, endDate, dailyRate)

	// Should be approximately $40.81 (more than simple 4% due to daily compounding)
	if interest.LessThan(money.NewMoney(40)) || interest.GreaterThan(money.NewMoney(42)) {
		t.Errorf("expected interest around 40-42, got %s", interest)
	}
}

func TestRateDBGetPrescribedRate(t *testing.T) {
	rateDB, err := rates.NewRateDB()
	if err != nil {
		t.Fatalf("failed to create rate DB: %v", err)
	}

	tests := []struct {
		name         string
		jurisdiction rules.Jurisdiction
		quarter      string
		expected     float64
		expectErr    bool
	}{
		{
			name:         "CRA 2024-Q1",
			jurisdiction: rules.CRA,
			quarter:      "2024-Q1",
			expected:     4.0,
			expectErr:    false,
		},
		{
			name:         "RQ 2024-Q1",
			jurisdiction: rules.RQ,
			quarter:      "2024-Q1",
			expected:     4.0,
			expectErr:    false,
		},
		{
			name:         "invalid quarter - uses fallback to most recent",
			jurisdiction: rules.CRA,
			quarter:      "invalid",
			expected:     4.0, // fallback to most recent rate
			expectErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rate, err := rateDB.GetPrescribedRate(tt.jurisdiction, tt.quarter)
			if tt.expectErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if rate != tt.expected {
				t.Errorf("expected rate %.2f, got %.2f", tt.expected, rate)
			}
		})
	}
}

func TestRateDBGetPrescribedRateForDate(t *testing.T) {
	rateDB, err := rates.NewRateDB()
	if err != nil {
		t.Fatalf("failed to create rate DB: %v", err)
	}

	tests := []struct {
		name         string
		jurisdiction rules.Jurisdiction
		date         time.Time
		expected     float64
	}{
		{
			name:         "CRA January 2024",
			jurisdiction: rules.CRA,
			date:         time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			expected:     4.0,
		},
		{
			name:         "CRA April 2024",
			jurisdiction: rules.CRA,
			date:         time.Date(2024, 4, 15, 0, 0, 0, 0, time.UTC),
			expected:     4.0,
		},
		{
			name:         "RQ March 2024",
			jurisdiction: rules.RQ,
			date:         time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC),
			expected:     4.0,
		},
		{
			name:         "RQ June 2024",
			jurisdiction: rules.RQ,
			date:         time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC),
			expected:     4.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rate, err := rateDB.GetPrescribedRateForDate(tt.jurisdiction, tt.date)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if rate != tt.expected {
				t.Errorf("expected rate %.2f, got %.2f", tt.expected, rate)
			}
		})
	}
}

func TestParseRange(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		startYear int
		endYear   int
		expectErr bool
	}{
		{
			name:      "valid range",
			input:     "2015-2025",
			startYear: 2015,
			endYear:   2025,
			expectErr: false,
		},
		{
			name:      "single year",
			input:     "2024-2024",
			startYear: 2024,
			endYear:   2024,
			expectErr: false,
		},
		{
			name:      "invalid format",
			input:     "invalid",
			startYear: 0,
			endYear:   0,
			expectErr: true,
		},
		{
			name:      "start after end",
			input:     "2025-2015",
			startYear: 0,
			endYear:   0,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rng, err := rules.ParseRange(tt.input)
			if tt.expectErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if rng.StartYear != tt.startYear || rng.EndYear != tt.endYear {
				t.Errorf("expected %d-%d, got %d-%d", tt.startYear, tt.endYear, rng.StartYear, rng.EndYear)
			}
		})
	}
}

func TestCalculatorCalculate(t *testing.T) {
	calc, err := NewCalculator()
	if err != nil {
		t.Fatalf("failed to create calculator: %v", err)
	}

	tests := []struct {
		name      string
		input     CalculationInput
		expectErr bool
		minTotal  money.Money
	}{
		{
			name: "basic calculation CRA",
			input: CalculationInput{
				Year:                2024,
				BaseDueCRA:          money.NewMoney(5000),
				ExpectedFilingDate:  time.Date(2024, 4, 30, 0, 0, 0, 0, time.UTC),
				ExpectedPaymentDate: time.Date(2024, 4, 30, 0, 0, 0, 0, time.UTC),
				ActualFilingDate:    time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC),
				ActualPaymentDate:   time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC),
				HadBalanceLastYear:  true,
			},
			expectErr: false,
			minTotal:  money.NewMoney(5325), // 5000 + 300 (5% + 1%) + ~27 interest (early payment, partial period)
		},
		{
			name: "on time payment CRA",
			input: CalculationInput{
				Year:                2024,
				BaseDueCRA:          money.NewMoney(5000),
				ExpectedFilingDate:  time.Date(2024, 4, 30, 0, 0, 0, 0, time.UTC),
				ExpectedPaymentDate: time.Date(2024, 4, 30, 0, 0, 0, 0, time.UTC),
				ActualFilingDate:    time.Date(2024, 4, 30, 0, 0, 0, 0, time.UTC),
				ActualPaymentDate:   time.Date(2024, 4, 30, 0, 0, 0, 0, time.UTC),
				HadBalanceLastYear:  true,
			},
			expectErr: false,
			minTotal:  money.NewMoney(5250), // 5000 + 250 (5% base penalty), no interest
		},
		{
			name: "negative amount",
			input: CalculationInput{
				Year:                2024,
				BaseDueCRA:          money.NewMoney(-100),
				ExpectedFilingDate:  time.Date(2024, 4, 30, 0, 0, 0, 0, time.UTC),
				ExpectedPaymentDate: time.Date(2024, 4, 30, 0, 0, 0, 0, time.UTC),
				ActualFilingDate:    time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC),
				ActualPaymentDate:   time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC),
				HadBalanceLastYear:  true,
			},
			expectErr: true,
		},
		{
			name: "early payment is valid (no error)",
			input: CalculationInput{
				Year:                2024,
				BaseDueCRA:          money.NewMoney(5000),
				ExpectedFilingDate:  time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC),
				ExpectedPaymentDate: time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC),
				ActualFilingDate:    time.Date(2024, 4, 30, 0, 0, 0, 0, time.UTC),
				ActualPaymentDate:   time.Date(2024, 4, 30, 0, 0, 0, 0, time.UTC),
				HadBalanceLastYear:  true,
			},
			expectErr: false,
			minTotal:  money.NewMoney(5250), // 5000 + 250 (5% base penalty), no interest (early)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := calc.Calculate(tt.input)
			if tt.expectErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if result.TotalDue.LessThan(tt.minTotal) {
				t.Errorf("expected total at least %s, got %s", tt.minTotal, result.TotalDue)
			}
		})
	}
}

func TestCalculatorNegativeBaseDue(t *testing.T) {
	calc, err := NewCalculator()
	if err != nil {
		t.Fatalf("failed to create calculator: %v", err)
	}

	input := CalculationInput{
		Year:                2024,
		BaseDueCRA:          money.NewMoney(-100),
		ExpectedFilingDate:  time.Date(2024, 4, 30, 0, 0, 0, 0, time.UTC),
		ExpectedPaymentDate: time.Date(2024, 4, 30, 0, 0, 0, 0, time.UTC),
		ActualFilingDate:    time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC),
		ActualPaymentDate:   time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC),
		HadBalanceLastYear:  true,
	}

	_, err = calc.Calculate(input)
	if err == nil {
		t.Errorf("expected error for negative base due, got nil")
	}
}

func TestRQRulesCalculateLateFilingPenalty(t *testing.T) {
	rqRules := rq.NewRQRules()

	tests := []struct {
		name            string
		taxAmount       money.Money
		dueDate         time.Time
		filingDate      time.Time
		hadBalance      bool
		expectedPenalty money.Money
	}{
		{
			name:            "no balance last year: no penalty",
			taxAmount:       money.NewMoney(5000),
			dueDate:         time.Date(2024, 4, 30, 0, 0, 0, 0, time.UTC),
			filingDate:      time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC),
			hadBalance:      false,
			expectedPenalty: money.NewMoney(0),
		},
		{
			name:            "had balance: 5% base + 1% monthly",
			taxAmount:       money.NewMoney(5000),
			dueDate:         time.Date(2024, 4, 30, 0, 0, 0, 0, time.UTC),
			filingDate:      time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC),
			hadBalance:      true,
			expectedPenalty: money.NewMoney(350), // 5% + 2% = 7% of 5000
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			penalty := rqRules.CalculateLateFilingPenalty(tt.taxAmount, tt.dueDate, tt.filingDate, tt.hadBalance)
			if !penalty.Equal(tt.expectedPenalty) {
				t.Errorf("expected penalty %s, got %s", tt.expectedPenalty, penalty)
			}
		})
	}
}

func TestRQRulesCalculateInterest(t *testing.T) {
	rqRules := rq.NewRQRules()

	amount := money.NewMoney(1000)
	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)
	dailyRate := 4.0

	interest := rqRules.CalculateInterest(amount, startDate, endDate, dailyRate)

	if interest.LessThan(money.NewMoney(40)) || interest.GreaterThan(money.NewMoney(42)) {
		t.Errorf("expected interest around 40-42, got %s", interest)
	}
}
