package cmd

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/qjcg/arcadia/exp/tpi/internal"
	"github.com/spf13/cobra"
)

type CalculateParams struct {
	Amount       float64 `descr:"Tax amount owed"`
	DueDate      string  `descr:"Payment due date (YYYY-MM-DD)"`
	PaymentDate  string  `descr:"Actual payment date (YYYY-MM-DD)"`
	Jurisdiction string  `descr:"Tax authority: 'cra' or 'rq'"`
	Output       string  `descr:"Output format: 'text' or 'json'" default:"text"`
	HadBalance   bool    `descr:"Had a balance owing in the previous tax year"`
}

func CalculateCmd() *cobra.Command {
	return boa.CmdT[CalculateParams]{
		Use:   "calculate",
		Short: "Calculate tax penalties and interest",
		Long: `Calculate penalties and interest on income tax for the Canada Revenue Agency (CRA)
or Revenu Québec based on the prescribed interest rates.

Examples:
  tpi calculate --amount 5000 --due-date 2024-04-30 --payment-date 2024-06-15 --jurisdiction cra
  tpi calculate --amount 10000 --due-date 2024-03-31 --payment-date 2024-07-01 --jurisdiction rq --output json`,
		RunFuncE: func(p *CalculateParams, cmd *cobra.Command, _ []string) error {
			return runCalculate(p, cmd)
		},
	}.ToCmd().ToCobra()
}

func runCalculate(p *CalculateParams, cmd *cobra.Command) error {
	dueDate, err := time.Parse("2006-01-02", p.DueDate)
	if err != nil {
		return fmt.Errorf("invalid due-date format: %w", err)
	}

	paymentDate, err := time.Parse("2006-01-02", p.PaymentDate)
	if err != nil {
		return fmt.Errorf("invalid payment-date format: %w", err)
	}

	jur := internal.Jurisdiction(p.Jurisdiction)
	if jur != internal.CRA && jur != internal.RQ {
		return fmt.Errorf("invalid jurisdiction: %s (must be 'cra' or 'rq')", p.Jurisdiction)
	}

	calc, err := internal.NewCalculator()
	if err != nil {
		return fmt.Errorf("failed to initialize calculator: %w", err)
	}

	result, err := calc.Calculate(internal.CalculationInput{
		TaxAmount:          p.Amount,
		DueDate:            dueDate,
		PaymentDate:        paymentDate,
		Jurisdiction:       jur,
		HadBalanceLastYear: p.HadBalance,
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

func outputJSON(cmd *cobra.Command, result *internal.CalculationResult) error {
	enc := json.NewEncoder(cmd.OutOrStdout())
	enc.SetIndent("", "  ")
	return enc.Encode(result)
}

func outputText(cmd *cobra.Command, result *internal.CalculationResult, calc *internal.Calculator) error {
	rateDB := calc.GetRateDB()
	meta := rateDB.GetMeta()

	w := cmd.OutOrStdout()

	fmt.Fprintln(w, "═══════════════════════════════════════════════════════════════")
	fmt.Fprintln(w, "              TAX PENALTIES & INTEREST CALCULATION")
	fmt.Fprintln(w, "═══════════════════════════════════════════════════════════════")
	fmt.Fprintln(w, "")
	fmt.Fprintf(w, "  Tax Authority:     %s\n", result.Jurisdiction)
	fmt.Fprintf(w, "  Tax Amount Owed:   $%.2f\n", result.TaxAmount)
	fmt.Fprintf(w, "  Due Date:          %s\n", result.DueDate.Format("January 2, 2006"))
	fmt.Fprintf(w, "  Payment Date:      %s\n", result.PaymentDate.Format("January 2, 2006"))
	fmt.Fprintf(w, "  Days Owed:         %d days\n", result.DaysOwed)
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "───────────────────────────────────────────────────────────────")
	fmt.Fprintln(w, "CHARGES")
	fmt.Fprintln(w, "───────────────────────────────────────────────────────────────")

	if result.LateFilingMonths > 0 {
		fmt.Fprintf(w, "  Late Filing Penalty:  $%.2f (%d months)\n",
			result.LateFilingPenalty, result.LateFilingMonths)
	} else {
		fmt.Fprintf(w, "  Late Filing Penalty:  $%.2f\n", result.LateFilingPenalty)
	}

	fmt.Fprintf(w, "  Interest:             $%.2f\n", result.Interest)
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "───────────────────────────────────────────────────────────────")
	fmt.Fprintf(w, "  TOTAL AMOUNT OWED:    $%.2f\n", result.TotalAmount)
	fmt.Fprintln(w, "═══════════════════════════════════════════════════════════════")
	fmt.Fprintln(w, "")
	fmt.Fprintf(w, "  Database Version:    %d\n", meta.Version)
	fmt.Fprintf(w, "  Database Updated:    %s\n", meta.Updated)

	return nil
}
