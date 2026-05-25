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

// setupCLI creates the root command and all subcommands.
func setupCLI(out, err io.Writer) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:     "sv",
		Short:   "sv is a semantic versioning tool for monorepos",
		Version: getVersion(),
	}
	rootCmd.SetOut(out)
	rootCmd.SetErr(err)

	rootCmd.AddCommand(createNextCmd())
	rootCmd.AddCommand(createCurrentCmd())
	rootCmd.AddCommand(createBumpCmd("major", semver.BumpMajor, "Force a major version bump"))
	rootCmd.AddCommand(createBumpCmd("minor", semver.BumpMinor, "Force a minor version bump"))
	rootCmd.AddCommand(createBumpCmd("patch", semver.BumpPatch, "Force a patch version bump"))

	return rootCmd
}

// Execute runs the sv CLI with stdout/stderr.
func Execute() {
	setupCLI(os.Stdout, os.Stderr).Execute()
}