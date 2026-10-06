package cli

import (
	"io"
	"os"
	"runtime/debug"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/spf13/cobra"
)

func getVersion() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		return info.Main.Version
	}
	return ""
}

// setupCLI creates the root command and all subcommands.
func setupCLI(out, err io.Writer) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:     "ap",
		Short:   "ap generates and validates Agent Plugin packages",
		Version: getVersion(),
	}
	rootCmd.SetOut(out)
	rootCmd.SetErr(err)

	rootCmd.AddCommand(createSyncCmd())
	rootCmd.AddCommand(createValidateCmd())

	return rootCmd
}

// Execute runs the ap CLI with stdout/stderr and returns the exit code.
// boa.Execute is required for commands from boa's ToCobra(): plain
// cmd.Execute() would silently swallow errors (boa sets SilenceErrors).
func Execute() int {
	if err := boa.Execute(setupCLI(os.Stdout, os.Stderr)); err != nil {
		return 1
	}
	return 0
}
