package main

import (
	"fmt"
	"io"
	"os"
	"runtime/debug"

	"github.com/qjcg/arcadia/exp/tpi/internal/cli"
	"github.com/spf13/cobra"
)

func main() {
	if err := setupCLI(os.Stdout, os.Stderr).Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func getVersion() string {
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" {
		return info.Main.Version
	}
	return "dev"
}

func setupCLI(out, err io.Writer) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:     "tpi",
		Short:   "tpi calculates tax penalties and interest for CRA and Revenu Québec",
		Version: getVersion(),
	}
	rootCmd.SetOut(out)
	rootCmd.SetErr(err)

	rootCmd.AddCommand(cli.CalculateCmd())
	rootCmd.AddCommand(cli.UpdateCmd())
	rootCmd.AddCommand(cli.RatesCmd())

	return rootCmd
}
