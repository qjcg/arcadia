package internal

import (
	"fmt"
	"time"

	"github.com/qjcg/arcadia/exp/tpi/rates"
)

// Jurisdiction represents CRA or Revenu Québec
type Jurisdiction string

const (
	CRA Jurisdiction = "cra"
	RQ  Jurisdiction = "rq"
)

// RateDB wraps the parsed rates database
type RateDB struct {
	db *rates.RatesDB
}

// NewRateDB parses and returns a new RateDB from embedded YAML
func NewRateDB() (*RateDB, error) {
	db, err := rates.Load(rates.EmbeddedRates)
	if err != nil {
		return nil, fmt.Errorf("failed to parse rates database: %w", err)
	}
	return &RateDB{db: db}, nil
}

// LoadRateDB loads rates from custom YAML bytes (for update command)
func LoadRateDB(data []byte) (*RateDB, error) {
	db, err := rates.Load(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse rates database: %w", err)
	}
	return &RateDB{db: db}, nil
}

// GetPrescribedRate returns the prescribed interest rate for a jurisdiction and quarter
func (r *RateDB) GetPrescribedRate(j Jurisdiction, quarter string) (float64, error) {
	ratesSlice := r.getRates(j).PrescribedRate
	for _, qr := range ratesSlice {
		if qr.Quarter == quarter {
			return qr.Rate, nil
		}
	}
	return 0, fmt.Errorf("no prescribed rate found for %s in quarter %s", j, quarter)
}

// GetLateFilingPenaltyRate returns the late filing penalty rate for a jurisdiction and year
func (r *RateDB) GetLateFilingPenaltyRate(j Jurisdiction, year int) (float64, error) {
	ratesSlice := r.getRates(j).LateFilingPenalty
	for _, yr := range ratesSlice {
		if yr.Year == year {
			return yr.Rate, nil
		}
	}
	return 0, fmt.Errorf("no late filing penalty rate found for %s in year %d", j, year)
}

// GetPrescribedRateForDate returns the prescribed rate for a jurisdiction on a given date
func (r *RateDB) GetPrescribedRateForDate(j Jurisdiction, t time.Time) (float64, error) {
	quarter := rates.GetQuarterForDate(t)
	return r.GetPrescribedRate(j, quarter)
}

func (r *RateDB) getRates(j Jurisdiction) *rates.JurisdictionRates {
	switch j {
	case CRA:
		return &r.db.CRA
	case RQ:
		return &r.db.RQ
	default:
		return nil
	}
}

// GetAllQuarters returns all quarters in the database for a jurisdiction
func (r *RateDB) GetAllQuarters(j Jurisdiction) []string {
	ratesSlice := r.getRates(j).PrescribedRate
	quarters := make([]string, len(ratesSlice))
	for i, qr := range ratesSlice {
		quarters[i] = qr.Quarter
	}
	return quarters
}

// GetMeta returns the rates metadata
func (r *RateDB) GetMeta() rates.RatesMeta {
	return r.db.Meta
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
