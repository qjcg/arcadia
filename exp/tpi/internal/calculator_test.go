package internal

import (
	"testing"
	"time"
)

func TestCRARulesCalculateLateFilingPenalty(t *testing.T) {
	cra := NewCRARules()

	tests := []struct {
		name          string
		taxAmount     float64
		dueDate       time.Time
		filingDate    time.Time
		hadBalance    bool
		expectedPenalty float64
	}{
		{
			name:           "no balance last year",
			taxAmount:      5000,
			dueDate:        time.Date(2024, 4, 30, 0, 0, 0, 0, time.UTC),
			filingDate:     time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC),
			hadBalance:     false,
			expectedPenalty: 0,
		},
		{
			name:           "on time (still pays base 5% penalty if had balance)",
			taxAmount:      5000,
			dueDate:        time.Date(2024, 4, 30, 0, 0, 0, 0, time.UTC),
			filingDate:     time.Date(2024, 4, 30, 0, 0, 0, 0, time.UTC),
			hadBalance:     true,
			expectedPenalty: 250, // 5% base penalty always applies if hadBalanceLastYear
		},
		{
			name:           "1 month late",
			taxAmount:      5000,
			dueDate:        time.Date(2024, 4, 30, 0, 0, 0, 0, time.UTC),
			filingDate:     time.Date(2024, 5, 30, 0, 0, 0, 0, time.UTC),
			hadBalance:     true,
			expectedPenalty: 300, // 5% + 1% = 6% of 5000
		},
		{
			name:           "12 months late",
			taxAmount:      5000,
			dueDate:        time.Date(2024, 4, 30, 0, 0, 0, 0, time.UTC),
			filingDate:     time.Date(2025, 4, 30, 0, 0, 0, 0, time.UTC),
			hadBalance:     true,
			expectedPenalty: 850, // 5% + 12% = 17% of 5000
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			penalty := cra.CalculateLateFilingPenalty(tt.taxAmount, tt.dueDate, tt.filingDate, tt.hadBalance)
			if penalty != tt.expectedPenalty {
				t.Errorf("expected penalty %.2f, got %.2f", tt.expectedPenalty, penalty)
			}
		})
	}
}

func TestCRARulesMonthsLate(t *testing.T) {
	cra := NewCRARules()

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
			expected:   2,
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
			months := cra.MonthsLate(tt.dueDate, tt.filingDate)
			if months != tt.expected {
				t.Errorf("expected %d months, got %d", tt.expected, months)
			}
		})
	}
}

func TestCRARulesCalculateInterest(t *testing.T) {
	cra := NewCRARules()

	// Test daily compounding: $1000 at 4% for 1 year should be about $40.81
	// (1000 * (1.04 - 1) = 40, but with daily compounding it's slightly more)
	amount := 1000.0
	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)
	dailyRate := 4.0

	interest := cra.CalculateInterest(amount, startDate, endDate, dailyRate)

	// Should be approximately $40.81 (more than simple 4% due to daily compounding)
	if interest < 40 || interest > 42 {
		t.Errorf("expected interest around 40-42, got %.2f", interest)
	}
}

func TestRateDBGetPrescribedRate(t *testing.T) {
	rateDB, err := NewRateDB()
	if err != nil {
		t.Fatalf("failed to create rate DB: %v", err)
	}

	tests := []struct {
		name       string
		jurisdiction Jurisdiction
		quarter    string
		expected   float64
		expectErr  bool
	}{
		{
			name:       "CRA 2024-Q1",
			jurisdiction: CRA,
			quarter:    "2024-Q1",
			expected:   4.0,
			expectErr:  false,
		},
		{
			name:       "RQ 2024-Q1",
			jurisdiction: RQ,
			quarter:    "2024-Q1",
			expected:   4.0,
			expectErr:  false,
		},
		{
			name:       "invalid quarter",
			jurisdiction: CRA,
			quarter:    "invalid",
			expected:   0,
			expectErr:  true,
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
	rateDB, err := NewRateDB()
	if err != nil {
		t.Fatalf("failed to create rate DB: %v", err)
	}

	tests := []struct {
		name        string
		jurisdiction Jurisdiction
		date        time.Time
		expected    float64
	}{
		{
			name:        "CRA January 2024",
			jurisdiction: CRA,
			date:        time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			expected:    4.0,
		},
		{
			name:        "CRA April 2024",
			jurisdiction: CRA,
			date:        time.Date(2024, 4, 15, 0, 0, 0, 0, time.UTC),
			expected:    4.0,
		},
		{
			name:        "RQ March 2024",
			jurisdiction: RQ,
			date:        time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC),
			expected:    4.0,
		},
		{
			name:        "RQ June 2024",
			jurisdiction: RQ,
			date:        time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC),
			expected:    4.0,
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
		name       string
		input      string
		startYear  int
		endYear    int
		expectErr  bool
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
			rng, err := ParseRange(tt.input)
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
		name         string
		input        CalculationInput
		expectErr    bool
		minTotal     float64
	}{
		{
			name: "basic calculation CRA",
			input: CalculationInput{
				TaxAmount:        5000,
				DueDate:          time.Date(2024, 4, 30, 0, 0, 0, 0, time.UTC),
				PaymentDate:      time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC),
				Jurisdiction:     CRA,
				HadBalanceLastYear: true,
			},
			expectErr: false,
			minTotal:   5375, // 5000 + 350 (5% + 1%) + interest
		},
		{
			name: "on time payment CRA",
			input: CalculationInput{
				TaxAmount:           5000,
				DueDate:             time.Date(2024, 4, 30, 0, 0, 0, 0, time.UTC),
				PaymentDate:         time.Date(2024, 4, 30, 0, 0, 0, 0, time.UTC),
				Jurisdiction:        CRA,
				HadBalanceLastYear:  true,
			},
			expectErr: false,
			minTotal:   5250, // 5000 + 250 (5% base penalty), no interest
		},
		{
			name: "negative amount",
			input: CalculationInput{
				TaxAmount:        -100,
				DueDate:          time.Date(2024, 4, 30, 0, 0, 0, 0, time.UTC),
				PaymentDate:      time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC),
				Jurisdiction:     CRA,
				HadBalanceLastYear: true,
			},
			expectErr: true,
		},
		{
			name: "payment before due",
			input: CalculationInput{
				TaxAmount:        5000,
				DueDate:          time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC),
				PaymentDate:      time.Date(2024, 4, 30, 0, 0, 0, 0, time.UTC),
				Jurisdiction:     CRA,
				HadBalanceLastYear: true,
			},
			expectErr: true,
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
			if result.TotalAmount < tt.minTotal {
				t.Errorf("expected total at least %.2f, got %.2f", tt.minTotal, result.TotalAmount)
			}
		})
	}
}

func TestCalculatorInvalidJurisdiction(t *testing.T) {
	calc, err := NewCalculator()
	if err != nil {
		t.Fatalf("failed to create calculator: %v", err)
	}

	input := CalculationInput{
		TaxAmount:    5000,
		DueDate:      time.Date(2024, 4, 30, 0, 0, 0, 0, time.UTC),
		PaymentDate:  time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC),
		Jurisdiction: "invalid",
	}

	_, err = calc.Calculate(input)
	if err == nil {
		t.Error("expected error for invalid jurisdiction, got nil")
	}
}