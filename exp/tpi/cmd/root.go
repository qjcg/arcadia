package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
)

const (
	Version = "0.1.0"
)

// RootCmd returns the root command
func RootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:     "tpi",
		Short:   "tpi calculates tax penalties and interest for CRA and Revenu Québec",
		Version: Version,
	}
	rootCmd.SetOut(os.Stdout)
	rootCmd.SetErr(os.Stderr)

	rootCmd.AddCommand(calculateCmd())
	rootCmd.AddCommand(updateCmd())
	rootCmd.AddCommand(ratesCmd())

	return rootCmd
}

// SetupCLI configures the CLI and returns the root command
func SetupCLI(out, err io.Writer) *cobra.Command {
	rootCmd := RootCmd()
	rootCmd.SetOut(out)
	rootCmd.SetErr(err)
	return rootCmd
}

// printVersion prints the version (used by cobra --version)
var _ = fmt.Printf // for cobra --version integration