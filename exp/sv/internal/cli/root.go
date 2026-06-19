package cli

import (
	"io"
	"os"
	"runtime/debug"

	"github.com/qjcg/arcadia/exp/sv/internal/semver"
	"github.com/spf13/cobra"
)

func getVersion() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		return info.Main.Version
	}
	return ""
}

// verboseFlag is set via the --verbose persistent flag on the root command.
var verboseFlag bool

// setupCLI creates the root command and all subcommands.
func setupCLI(out, err io.Writer) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:     "sv",
		Short:   "sv is a semantic versioning tool for monorepos",
		Version: getVersion(),
	}
	rootCmd.SetOut(out)
	rootCmd.SetErr(err)

	rootCmd.PersistentFlags().BoolVar(&verboseFlag, "verbose", false, "Enable verbose output (e.g., retraction warnings)")

	rootCmd.AddCommand(createNextCmd())
	rootCmd.AddCommand(createCurrentCmd())
	rootCmd.AddCommand(createBumpCmd("major", semver.BumpMajor, "Force a major version bump"))
	rootCmd.AddCommand(createBumpCmd("minor", semver.BumpMinor, "Force a minor version bump"))
	rootCmd.AddCommand(createBumpCmd("patch", semver.BumpPatch, "Force a patch version bump"))
	rootCmd.AddCommand(createChangelogCmd())

	return rootCmd
}

// Execute runs the sv CLI with stdout/stderr and returns the exit code.
func Execute() int {
	if err := setupCLI(os.Stdout, os.Stderr).Execute(); err != nil {
		return 1
	}
	return 0
}
