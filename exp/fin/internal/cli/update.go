package cli

import (
	"fmt"
	"os"
	"time"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/qjcg/arcadia/exp/fin/internal/fetcher"
	"github.com/qjcg/arcadia/exp/fin/internal/rates"
	"github.com/qjcg/arcadia/exp/fin/internal/rules"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type UpdateParams struct {
	Range  string `descr:"Year range (e.g., 2015-2025)"`
	Source string `descr:"Source: 'cra', 'rq', or 'both'"`
	Output string `descr:"Output file path" optional:"true"`
	DryRun bool   `descr:"Show what would be updated without making changes" optional:"true"`
}

func updateCmd() *cobra.Command {
	return boa.CmdT[UpdateParams]{
		Use:   "update",
		Short: "Update the interest rates database",
		Long: `Fetch and update the interest rates database from official CRA and
Revenu Québec sources.

The update command fetches prescribed interest rates for a specified year
range and either displays them (--dry-run) or updates the rates.yaml file.

Examples:
  fin tax penalties-and-interest update --range 2020-2025 --source cra
  fin tax pi update --range 2020-2025 --source both --output rates.yaml
  fin tax penalties-and-interest update --range 2024-2024 --source cra --dry-run`,
		RunFuncE: func(p *UpdateParams, cmd *cobra.Command, _ []string) error {
			return runUpdate(p, cmd)
		},
	}.ToCmd().ToCobra()
}

func runUpdate(p *UpdateParams, cmd *cobra.Command) error {
	rng, err := rules.ParseRange(p.Range)
	if err != nil {
		return err
	}

	source := rules.Jurisdiction(p.Source)
	if source != rules.CRA && source != rules.RQ && p.Source != "both" {
		return fmt.Errorf("invalid source: %s (must be 'cra', 'rq', or 'both')", p.Source)
	}

	f := fetcher.NewFetcher()

	var fetchedRates []fetcher.FetchedRate

	if p.Source == "both" {
		craRates, err := f.FetchRates(rules.CRA, rng.StartYear, rng.EndYear)
		if err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "Warning: failed to fetch CRA rates: %v\n", err)
		} else {
			fetchedRates = append(fetchedRates, craRates...)
		}

		rqRates, err := f.FetchRates(rules.RQ, rng.StartYear, rng.EndYear)
		if err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "Warning: failed to fetch RQ rates: %v\n", err)
		} else {
			fetchedRates = append(fetchedRates, rqRates...)
		}
	} else {
		fetchedRates, err = f.FetchRates(source, rng.StartYear, rng.EndYear)
		if err != nil {
			return fmt.Errorf("failed to fetch rates: %w", err)
		}
	}

	if p.DryRun {
		return outputDryRun(cmd, fetchedRates)
	}

	// Validate fetched rates
	if err := f.ValidateRates(fetchedRates); err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "Warning: %v\n", err)
		fmt.Fprintf(cmd.ErrOrStderr(), "Attempting to continue with fetched data...\n")
	}

	// Load current rates
	rateDB, err := rates.Load(rates.EmbeddedRates)
	if err != nil {
		return fmt.Errorf("failed to load current rates: %w", err)
	}

	// Update rates
	updated := updateRatesInDB(rateDB, fetchedRates)

	// Serialize back to YAML
	output, err := yaml.Marshal(updated)
	if err != nil {
		return fmt.Errorf("failed to serialize updated rates: %w", err)
	}

	// Determine output path
	outputPath := p.Output
	if outputPath == "" {
		outputPath = "rates/rates.yaml"
	}

	// Write updated rates
	if err := os.WriteFile(outputPath, output, 0o644); err != nil {
		return fmt.Errorf("failed to write updated rates: %w", err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Successfully updated %d rates in %s\n", len(fetchedRates), outputPath)
	return nil
}

func outputDryRun(cmd *cobra.Command, fetchedRates []fetcher.FetchedRate) error {
	w := cmd.OutOrStdout()
	fmt.Fprintln(w, "═══════════════════════════════════════════════════════════════")
	fmt.Fprintln(w, "                    DRY RUN - NO CHANGES MADE")
	fmt.Fprintln(w, "═══════════════════════════════════════════════════════════════")
	fmt.Fprintln(w, "")

	if len(fetchedRates) == 0 {
		fmt.Fprintln(w, "  No rates would be fetched for the specified range.")
		fmt.Fprintln(w, "")
		fmt.Fprintf(w, "  Note: Web scraping may require manual verification.\n")
		fmt.Fprintf(w, "  Visit the official sources to verify rates.\n")
	} else {
		fmt.Fprintf(w, "  Would update %d rates:\n\n", len(fetchedRates))
		for _, r := range fetchedRates {
			fmt.Fprintf(w, "    %s %-6s  %.2f%%\n", r.Source, r.Quarter, r.Rate)
		}
	}

	fmt.Fprintln(w, "")
	return nil
}

func updateRatesInDB(db *rates.RatesDB, newRates []fetcher.FetchedRate) *rates.RatesDB {
	for _, r := range newRates {
		switch r.Source {
		case "CRA":
			db.CRA.PrescribedRate = upsertQuarterRate(db.CRA.PrescribedRate, r.Quarter, r.Rate)
		case "RQ":
			db.RQ.PrescribedRate = upsertQuarterRate(db.RQ.PrescribedRate, r.Quarter, r.Rate)
		}
	}

	// Update metadata
	db.Meta.Version++
	db.Meta.Updated = time.Now().Format("2006-01-02")

	return db
}

func upsertQuarterRate(list []rates.QuarterRate, quarter string, rate float64) []rates.QuarterRate {
	for i, qr := range list {
		if qr.Quarter == quarter {
			list[i].Rate = rate
			return list
		}
	}

	// Insert in sorted order
	newRate := rates.QuarterRate{Quarter: quarter, Rate: rate}
	list = append(list, newRate)

	// Simple bubble sort to maintain order
	for i := 0; i < len(list)-1; i++ {
		for j := 0; j < len(list)-i-1; j++ {
			if list[j].Quarter > list[j+1].Quarter {
				list[j], list[j+1] = list[j+1], list[j]
			}
		}
	}

	return list
}
