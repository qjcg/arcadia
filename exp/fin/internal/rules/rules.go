package rules

import (
	"fmt"
	"time"

	"github.com/qjcg/arcadia/exp/fin/internal/money"
)

// Jurisdiction represents CRA or Revenu Québec
type Jurisdiction string

const (
	CRA Jurisdiction = "cra"
	RQ  Jurisdiction = "rq"
)

// Jurisdiction defines the calculation rules for a tax jurisdiction.
type Interface interface {
	CalculateLateFilingPenalty(taxAmount money.Money, dueDate, filingDate time.Time, hadBalanceLastYear bool) money.Money
	MonthsLate(dueDate, filingDate time.Time) int
	CalculateInterest(amount money.Money, startDate, endDate time.Time, dailyRate float64) money.Money
	GetPrescribedRateSource() string
	GetLateFilingPenaltyInfo() string
	ValidateDate(t time.Time) error
}

// Range represents a time range for updates
type Range struct {
	StartYear int
	EndYear   int
}

// ParseRange parses "YYYY-YYYY" format
func ParseRange(s string) (*Range, error) {
	var start, end int
	_, err := fmt.Sscanf(s, "%d-%d", &start, &end)
	if err != nil {
		return nil, fmt.Errorf("invalid range format %q (expected YYYY-YYYY): %w", s, err)
	}
	if start > end {
		return nil, fmt.Errorf("start year %d is after end year %d", start, end)
	}
	return &Range{StartYear: start, EndYear: end}, nil
}
