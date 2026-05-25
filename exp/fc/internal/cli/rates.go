package cli

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/qjcg/arcadia/exp/fc/internal"
	"github.com/spf13/cobra"
)

type RatesParams struct {
	Jurisdiction string `descr:"Show rates for 'cra', 'rq', or 'both'" default:"both"`
	Year         int    `descr:"Show rates for a specific year" optional:"true"`
}

func ratesCmd() *cobra.Command {
	return boa.CmdT[RatesParams]{
		Use:   "rates",
		Short: "Display current interest rates",
		Long: `Display the embedded interest rates database for CRA and/or Revenu Québec.

Examples:
  fc tax penalties-and-interest rates
  fc tax pi rates --jurisdiction cra
  fc tax penalties-and-interest rates --year 2024 --jurisdiction rq`,
		RunFuncE: func(p *RatesParams, cmd *cobra.Command, _ []string) error {
			return runRates(cmd, p.Jurisdiction, p.Year)
		},
	}.ToCmd().ToCobra()
}

func runRates(cmd *cobra.Command, jurisdiction string, filterYear int) error {
	rateDB, err := internal.NewRateDB()
	if err != nil {
		return fmt.Errorf("failed to load rates database: %w", err)
	}

	meta := rateDB.GetMeta()
	w := cmd.OutOrStdout()

	fmt.Fprintln(w, "═══════════════════════════════════════════════════════════════")
	fmt.Fprintln(w, "                 INTEREST RATES DATABASE")
	fmt.Fprintln(w, "═══════════════════════════════════════════════════════════════")
	fmt.Fprintf(w, "  Version:    %d\n", meta.Version)
	fmt.Fprintf(w, "  Updated:    %s\n", meta.Updated)
	fmt.Fprintln(w, "  Sources:")
	for line := range strings.SplitSeq(meta.Source, "\n") {
		if line != "" {
			fmt.Fprintf(w, "    - %s\n", line)
		}
	}
	fmt.Fprintln(w, "═══════════════════════════════════════════════════════════════")

	showCRA := jurisdiction == "both" || jurisdiction == "cra"
	showRQ := jurisdiction == "both" || jurisdiction == "rq"

	if showCRA {
		printJurisdictionRates(w, "CRA (Canada Revenue Agency)", rateDB, internal.CRA, filterYear)
	}

	if showRQ {
		printJurisdictionRates(w, "RQ (Revenu Québec)", rateDB, internal.RQ, filterYear)
	}

	return nil
}

func printJurisdictionRates(w io.Writer, name string, rateDB *internal.RateDB, j internal.Jurisdiction, filterYear int) {
	fmt.Fprintln(w, "")
	fmt.Fprintf(w, "  ── %s\n", name)
	fmt.Fprintln(w, "")

	quarters := rateDB.GetAllQuarters(j)

	if filterYear > 0 {
		var filtered []string
		for _, q := range quarters {
			var year int
			fmt.Sscanf(q, "%d", &year)
			if year == filterYear {
				filtered = append(filtered, q)
			}
		}
		quarters = filtered
	}

	if len(quarters) == 0 {
		fmt.Fprintf(w, "    No rates available for the specified criteria.\n")
		return
	}

	// Group by year
	byYear := make(map[int][]string)
	for _, q := range quarters {
		var year int
		fmt.Sscanf(q, "%d", &year)
		byYear[year] = append(byYear[year], q)
	}

	// Sort years
	var years []int
	for year := range byYear {
		years = append(years, year)
	}
	sort.Ints(years)

	for _, year := range years {
		fmt.Fprintf(w, "    %d:\n", year)
		// Sort quarters within year
		sort.Strings(byYear[year])
		for _, q := range byYear[year] {
			rate, _ := rateDB.GetPrescribedRate(j, q)
			fmt.Fprintf(w, "      %s  %.2f%%\n", q, rate)
		}
	}
}
