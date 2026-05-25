package cli

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/qjcg/arcadia/exp/fc/internal/calculator"
	"github.com/shopspring/decimal"
	"github.com/spf13/cobra"
)

type PicParams struct {
	Year                int    `descr:"Tax year (e.g., 2024)"`
	Earned              string `descr:"Income earned during the year" optional:"true"`
	BaseDueCRA          string `descr:"Base tax amount owed to CRA" default:"0"`
	BaseDueRQ           string `descr:"Base tax amount owed to Revenu Québec" default:"0"`
	ExpectedFilingDate  string `descr:"Expected filing date (YYYY-MM-DD)"`
	ExpectedPaymentDate string `descr:"Expected payment date (YYYY-MM-DD)"`
	ActualFilingDate    string `descr:"Actual filing date (YYYY-MM-DD)"`
	ActualPaymentDate   string `descr:"Actual payment date (YYYY-MM-DD)"`
	HadBalance          bool   `descr:"Had a balance owing in the previous tax year" optional:"true"`
	Output              string `descr:"Output format: 'text' or 'json'" default:"text"`
}

func picCmd() *cobra.Command {
	return boa.CmdT[PicParams]{
		Use:     "penalties-and-interest",
		Aliases: []string{"pi"},
		Short: "Calculate tax penalties and interest",
		Long: `Calculate penalties and interest on income tax for CRA and Revenu Québec
based on the prescribed interest rates.

Examples:
  fc tax penalties-and-interest --year 2024 --base-due-cra 5000 --expected-filing-date 2024-04-30 --actual-filing-date 2024-06-15 --expected-payment-date 2024-04-30 --actual-payment-date 2024-06-15
  fc tax pi --year 2024 --base-due-cra 5000 --base-due-rq 3000 --expected-payment-date 2024-04-30 --actual-payment-date 2024-07-01 --output json`,
		RunFuncE: func(p *PicParams, cmd *cobra.Command, _ []string) error {
			return runPic(p, cmd)
		},
	}.ToCmd().ToCobra()
}

func runPic(p *PicParams, cmd *cobra.Command) error {
	earned, _ := decimal.NewFromString(p.Earned)
	baseDueCRA, _ := decimal.NewFromString(p.BaseDueCRA)
	baseDueRQ, _ := decimal.NewFromString(p.BaseDueRQ)

	expectedFiling, err := time.Parse("2006-01-02", p.ExpectedFilingDate)
	if err != nil {
		return fmt.Errorf("invalid expected-filing-date format: %w", err)
	}

	expectedPayment, err := time.Parse("2006-01-02", p.ExpectedPaymentDate)
	if err != nil {
		return fmt.Errorf("invalid expected-payment-date format: %w", err)
	}

	actualFiling, err := time.Parse("2006-01-02", p.ActualFilingDate)
	if err != nil {
		return fmt.Errorf("invalid actual-filing-date format: %w", err)
	}

	actualPayment, err := time.Parse("2006-01-02", p.ActualPaymentDate)
	if err != nil {
		return fmt.Errorf("invalid actual-payment-date format: %w", err)
	}

	calc, err := calculator.NewCalculator()
	if err != nil {
		return fmt.Errorf("failed to initialize calculator: %w", err)
	}

	result, err := calc.Calculate(calculator.CalculationInput{
		Year:                p.Year,
		Earned:              earned,
		BaseDueCRA:          baseDueCRA,
		BaseDueRQ:           baseDueRQ,
		ExpectedFilingDate:  expectedFiling,
		ExpectedPaymentDate: expectedPayment,
		ActualFilingDate:    actualFiling,
		ActualPaymentDate:   actualPayment,
		HadBalanceLastYear:  p.HadBalance,
	})
	if err != nil {
		return fmt.Errorf("calculation failed: %w", err)
	}

	switch p.Output {
	case "json":
		return outputJSON(cmd, result)
	case "text":
		return outputText(cmd, result, calc)
	default:
		return fmt.Errorf("invalid output format: %s (must be 'text' or 'json')", p.Output)
	}
}

func outputJSON(cmd *cobra.Command, result *calculator.CalculationResult) error {
	enc := json.NewEncoder(cmd.OutOrStdout())
	enc.SetIndent("", "  ")
	return enc.Encode(result)
}

func outputText(cmd *cobra.Command, result *calculator.CalculationResult, calc *calculator.Calculator) error {
	rateDB := calc.GetRateDB()
	meta := rateDB.GetMeta()

	w := cmd.OutOrStdout()

	fmt.Fprintln(w, "═══════════════════════════════════════════════════════════════")
	fmt.Fprintln(w, "              TAX PENALTIES & INTEREST CALCULATION")
	fmt.Fprintln(w, "═══════════════════════════════════════════════════════════════")
	fmt.Fprintln(w, "")
	fmt.Fprintf(w, "  Year:                    %d\n", result.Year)
	if result.Earned.GreaterThan(decimal.Zero) {
		fmt.Fprintf(w, "  Income Earned:           $%.2f\n", result.Earned.InexactFloat64())
	}
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "───────────────────────────────────────────────────────────────")
	fmt.Fprintln(w, "CRA (Canada Revenue Agency)")
	fmt.Fprintln(w, "───────────────────────────────────────────────────────────────")
	fmt.Fprintf(w, "  Base Tax Due:            $%.2f\n", result.BaseDueCRA.InexactFloat64())
	fmt.Fprintf(w, "  Penalties:               $%.2f\n", result.PenaltiesCRA.InexactFloat64())
	fmt.Fprintf(w, "  Interest:                $%.2f\n", result.InterestCRA.InexactFloat64())
	fmt.Fprintf(w, "  Total Due CRA:           $%.2f\n", result.TotalDueCRA.InexactFloat64())
	fmt.Fprintln(w, "")

	if result.BaseDueRQ.GreaterThan(decimal.Zero) {
		fmt.Fprintln(w, "───────────────────────────────────────────────────────────────")
		fmt.Fprintln(w, "RQ (Revenu Québec)")
		fmt.Fprintln(w, "───────────────────────────────────────────────────────────────")
		fmt.Fprintf(w, "  Base Tax Due:            $%.2f\n", result.BaseDueRQ.InexactFloat64())
		fmt.Fprintf(w, "  Penalties:               $%.2f\n", result.PenaltiesRQ.InexactFloat64())
		fmt.Fprintf(w, "  Interest:                $%.2f\n", result.InterestRQ.InexactFloat64())
		fmt.Fprintf(w, "  Total Due RQ:             $%.2f\n", result.TotalDueRQ.InexactFloat64())
		fmt.Fprintln(w, "")
	}

	fmt.Fprintln(w, "═══════════════════════════════════════════════════════════════")
	fmt.Fprintf(w, "  TOTAL AMOUNT OWED:       $%.2f\n", result.TotalDue.InexactFloat64())
	fmt.Fprintln(w, "═══════════════════════════════════════════════════════════════")
	fmt.Fprintln(w, "")
	fmt.Fprintf(w, "  Expected Filing Date:    %s\n", result.ExpectedFilingDate.Format("January 2, 2006"))
	fmt.Fprintf(w, "  Actual Filing Date:      %s\n", result.ActualFilingDate.Format("January 2, 2006"))
	fmt.Fprintf(w, "  Expected Payment Date:   %s\n", result.ExpectedPaymentDate.Format("January 2, 2006"))
	fmt.Fprintf(w, "  Actual Payment Date:     %s\n", result.ActualPaymentDate.Format("January 2, 2006"))
	fmt.Fprintln(w, "")
	fmt.Fprintf(w, "  Database Version:         %d\n", meta.Version)
	fmt.Fprintf(w, "  Database Updated:        %s\n", meta.Updated)

	return nil
}
