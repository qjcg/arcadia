package main

import (
	"fmt"
	"io"
	"os"
	"runtime/debug"

	"github.com/qjcg/arcadia/exp/fin/internal/cli"
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
		Use:     "fin",
		Short:   "fin is a financial calculator",
		Version: getVersion(),
	}
	rootCmd.SetOut(out)
	rootCmd.SetErr(err)

	rootCmd.AddCommand(cli.TaxCmd())

	return rootCmd
}
