package internal

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Fetcher handles fetching interest rates from official sources
type Fetcher struct {
	client *http.Client
}

// NewFetcher creates a new rate fetcher
func NewFetcher() *Fetcher {
	return &Fetcher{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// FetchedRate represents a single rate fetched from a source
type FetchedRate struct {
	Quarter string
	Rate    float64
	Source  string
}

// FetchRates fetches rates from the specified source for a given year range
func (f *Fetcher) FetchRates(source Jurisdiction, startYear, endYear int) ([]FetchedRate, error) {
	switch source {
	case CRA:
		return f.fetchCRARates(startYear, endYear)
	case RQ:
		return f.fetchRQRates(startYear, endYear)
	default:
		return nil, fmt.Errorf("unsupported source: %s", source)
	}
}

// fetchCRARates fetches prescribed interest rates from CRA
func (f *Fetcher) fetchCRARates(startYear, endYear int) ([]FetchedRate, error) {
	craURL := "https://www.canada.ca/en/revenue-agency/services/tax/prescribed-interest-rates.html"

	resp, err := f.client.Get(craURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch CRA rates: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("CRA returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read CRA response: %w", err)
	}

	return f.parseCRARates(string(body), startYear, endYear)
}

// parseCRARates extracts rates from CRA HTML content
func (f *Fetcher) parseCRARates(html string, startYear, endYear int) ([]FetchedRate, error) {
	// CRA publishes rates in a table format on their website
	// This parser looks for patterns like "Q1 2024" and associated rates
	var rates []FetchedRate

	// Simple parsing - look for quarter and rate patterns
	// Format typically: "January 1 to March 31, 2024" or "Q1 2024"
	lines := strings.Split(html, "\n")

	for i, line := range lines {
		// Look for table rows with quarter data
		if strings.Contains(line, "20") && strings.Contains(line, "%") {
			rate := f.extractCRARateFromLine(line)
			quarter := f.extractCRAQuarterFromLines(lines, i)
			if rate > 0 && quarter != "" {
				rates = append(rates, FetchedRate{
					Quarter: quarter,
					Rate:    rate,
					Source:  "CRA",
				})
			}
		}
	}

	// Filter to requested year range
	filtered := filterRatesByYear(rates, startYear, endYear)
	return filtered, nil
}

// extractCRARateFromLine extracts numeric rate from a line
func (f *Fetcher) extractCRARateFromLine(line string) float64 {
	// Look for percentage values in the text
	start := strings.Index(line, "%")
	if start == -1 {
		return 0
	}

	// Back up to find the number
	end := start
	start--
	for start > 0 && (line[start] == '.' || (line[start] >= '0' && line[start] <= '9')) {
		start--
	}
	start++

	numStr := line[start:end]
	var rate float64
	fmt.Sscanf(numStr, "%f", &rate)
	return rate
}

// extractCRAQuarterFromLines extracts quarter from surrounding lines
func (f *Fetcher) extractCRAQuarterFromLines(lines []string, idx int) string {
	// Search backward for quarter indicators
	for i := idx - 5; i <= idx+5; i++ {
		if i < 0 || i >= len(lines) {
			continue
		}
		line := strings.ToLower(lines[i])
		if strings.Contains(line, "q1") || strings.Contains(line, "quarter 1") ||
			strings.Contains(line, "january") || strings.Contains(line, "march") {
			return f.parseCRADateContext(lines[i], lines[idx])
		}
	}
	return ""
}

// parseCRADateContext determines quarter from date context
func (f *Fetcher) parseCRADateContext(dateLine, rateLine string) string {
	lower := strings.ToLower(dateLine)

	year := 0
	for _, token := range strings.Fields(lower) {
		if len(token) == 4 {
			for _, c := range token {
				if c < '0' || c > '9' {
					return ""
				}
			}
			fmt.Sscanf(token, "%d", &year)
		}
	}

	quarter := 0
	if strings.Contains(lower, "q1") || strings.Contains(lower, "january") || strings.Contains(lower, "march") {
		quarter = 1
	} else if strings.Contains(lower, "q2") || strings.Contains(lower, "april") || strings.Contains(lower, "june") {
		quarter = 2
	} else if strings.Contains(lower, "q3") || strings.Contains(lower, "july") || strings.Contains(lower, "september") {
		quarter = 3
	} else if strings.Contains(lower, "q4") || strings.Contains(lower, "october") || strings.Contains(lower, "december") {
		quarter = 4
	}

	if year > 0 && quarter > 0 {
		return fmt.Sprintf("%d-Q%d", year, quarter)
	}
	return ""
}

// fetchRQRates fetches prescribed interest rates from Revenu Québec
func (f *Fetcher) fetchRQRates(startYear, endYear int) ([]FetchedRate, error) {
	rqURL := "https://www.revenuquebec.ca/en/one-mission-concrete-actions/ensuring-tax-compliance/penalties-and-interest/interest-rates-on-debts/"

	resp, err := f.client.Get(rqURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch RQ rates: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("RQ returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read RQ response: %w", err)
	}

	return f.parseRQRates(string(body), startYear, endYear)
}

// parseRQRates extracts rates from RQ HTML content
func (f *Fetcher) parseRQRates(html string, startYear, endYear int) ([]FetchedRate, error) {
	var rates []FetchedRate

	lines := strings.Split(html, "\n")

	for i, line := range lines {
		if strings.Contains(line, "20") && strings.Contains(line, "%") {
			rate := f.extractCRARateFromLine(line) // RQ uses similar format
			quarter := f.extractRQQuarterFromLines(lines, i)
			if rate > 0 && quarter != "" {
				rates = append(rates, FetchedRate{
					Quarter: quarter,
					Rate:    rate,
					Source:  "RQ",
				})
			}
		}
	}

	filtered := filterRatesByYear(rates, startYear, endYear)
	return filtered, nil
}

// extractRQQuarterFromLines extracts quarter from RQ page
func (f *Fetcher) extractRQQuarterFromLines(lines []string, idx int) string {
	for i := idx - 5; i <= idx+5; i++ {
		if i < 0 || i >= len(lines) {
			continue
		}
		line := strings.ToLower(lines[i])
		if strings.Contains(line, "t1") || strings.Contains(line, "janvier") ||
			strings.Contains(line, "1er") || strings.Contains(line, "trimestre") {
			return f.parseRQDateContext(lines[i], lines[idx])
		}
	}
	return ""
}

// parseRQDateContext determines quarter from RQ French date context
func (f *Fetcher) parseRQDateContext(dateLine, rateLine string) string {
	lower := strings.ToLower(dateLine)

	year := 0
	for _, token := range strings.Fields(lower) {
		if len(token) == 4 {
			for _, c := range token {
				if c < '0' || c > '9' {
					return ""
				}
			}
			fmt.Sscanf(token, "%d", &year)
		}
	}

	quarter := 0
	if strings.Contains(lower, "t1") || strings.Contains(lower, "janvier") ||
		strings.Contains(lower, "1er") || strings.Contains(lower, "mars") {
		quarter = 1
	} else if strings.Contains(lower, "t2") || strings.Contains(lower, "avril") ||
		strings.Contains(lower, "juin") {
		quarter = 2
	} else if strings.Contains(lower, "t3") || strings.Contains(lower, "juillet") ||
		strings.Contains(lower, "septembre") {
		quarter = 3
	} else if strings.Contains(lower, "t4") || strings.Contains(lower, "octobre") ||
		strings.Contains(lower, "decembre") {
		quarter = 4
	}

	if year > 0 && quarter > 0 {
		return fmt.Sprintf("%d-Q%d", year, quarter)
	}
	return ""
}

func filterRatesByYear(rates []FetchedRate, startYear, endYear int) []FetchedRate {
	var filtered []FetchedRate
	for _, r := range rates {
		year := 0
		fmt.Sscanf(r.Quarter, "%d", &year)
		if year >= startYear && year <= endYear {
			filtered = append(filtered, r)
		}
	}
	return filtered
}

// FetchRateRange fetches rates for a specific range string (e.g., "2015-2025")
func (f *Fetcher) FetchRateRange(source Jurisdiction, rangeStr string) ([]FetchedRate, error) {
	rng, err := ParseRange(rangeStr)
	if err != nil {
		return nil, err
	}
	return f.FetchRates(source, rng.StartYear, rng.EndYear)
}

// ValidateRates checks fetched rates against expected patterns
func (f *Fetcher) ValidateRates(rates []FetchedRate) error {
	if len(rates) == 0 {
		return fmt.Errorf("no rates fetched")
	}

	for _, r := range rates {
		if r.Rate <= 0 || r.Rate > 30 {
			return fmt.Errorf("rate %.2f for %s seems invalid", r.Rate, r.Quarter)
		}
	}
	return nil
}

// BuildUpdateURL constructs a URL for manual rate lookup
func (f *Fetcher) BuildUpdateURL(source Jurisdiction) string {
	switch source {
	case CRA:
		return "https://www.canada.ca/en/revenue-agency/services/tax/prescribed-interest-rates.html"
	case RQ:
		base, _ := url.Parse("https://www.revenuquebec.ca/en/one-mission-concrete-actions/ensuring-tax-compliance/penalties-and-interest/interest-rates-on-debts/")
		return base.String()
	default:
		return ""
	}
}
