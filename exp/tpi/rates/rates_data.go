package rates

import (
	_ "embed"
	"time"

	"gopkg.in/yaml.v3"
)

//go:embed rates.yaml
var EmbeddedRates []byte

// QuarterRate represents a quarterly prescribed interest rate
type QuarterRate struct {
	Quarter string  `yaml:"quarter"`
	Rate    float64 `yaml:"rate"`
}

// YearRate represents an annual penalty rate
type YearRate struct {
	Year int     `yaml:"year"`
	Rate float64 `yaml:"rate"`
}

// JurisdictionRates holds rates for a jurisdiction
type JurisdictionRates struct {
	PrescribedRate    []QuarterRate `yaml:"prescribed_rate"`
	LateFilingPenalty []YearRate    `yaml:"late_filing_penalty"`
}

// RatesDB is the root structure of the rates YAML
type RatesDB struct {
	Meta RatesMeta          `yaml:"meta"`
	CRA  JurisdictionRates `yaml:"cra"`
	RQ   JurisdictionRates `yaml:"rq"`
}

// RatesMeta contains metadata about the rates database
type RatesMeta struct {
	Version int    `yaml:"version"`
	Updated string `yaml:"updated"`
	Source  string `yaml:"source"`
}

// Load parses and returns a RatesDB from bytes
func Load(data []byte) (*RatesDB, error) {
	var db RatesDB
	if err := yaml.Unmarshal(data, &db); err != nil {
		return nil, err
	}
	return &db, nil
}

// GetQuarterForDate returns the quarter string for a given date
func GetQuarterForDate(t time.Time) string {
	year := t.Year()
	month := t.Month()
	quarter := int((month-1)/3) + 1
	return FormatQuarter(year, quarter)
}

// FormatQuarter creates "YYYY-QN" string
func FormatQuarter(year, quarter int) string {
	return time.Date(year, time.Month((quarter-1)*3+1), 1, 0, 0, 0, 0, time.UTC).Format("2006-Q1")
}