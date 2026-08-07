package cli

import (
	"io"
	"os"
	"runtime/debug"

	"github.com/qjcg/arcadia/cmd/awesome-lint/internal/linter"
	"github.com/spf13/cobra"
)

func getVersion() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		return info.Main.Version
	}
	return ""
}

func setupCLI(out, err io.Writer) *cobra.Command {
	var filename string
	var jsonOutput bool
	var fixMode bool

	rootCmd := &cobra.Command{
		Use:   "awesome-lint [file]",
		Short: "Lint an Awesome list for compliance with awesome.re guidelines",
		Long: `Lint an Awesome list for compliance with awesome.re guidelines.

If no file is specified, readme.md in the current directory is used.
Accepts a local file path or a GitHub repository URL.`,
		Args:    cobra.MaximumNArgs(1),
		Version: getVersion(),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := filename
			if len(args) > 0 {
				path = args[0]
			}

			r := linter.New()

			var results *linter.Results
			var lintErr error
			if fixMode {
				results, lintErr = r.LintWithFix(path)
			} else {
				results, lintErr = r.Lint(path)
			}
			if lintErr != nil {
				return lintErr
			}

			if jsonOutput {
				return results.WriteJSON(cmd.OutOrStdout())
			}

			results.WritePretty(cmd.OutOrStdout())
			if results.HasErrors() {
				os.Exit(1)
			}
			return nil
		},
	}
	rootCmd.SetOut(out)
	rootCmd.SetErr(err)

	rootCmd.Flags().StringVarP(&filename, "filename", "f", "readme.md", "Path to the markdown file to lint")
	rootCmd.Flags().BoolVarP(&jsonOutput, "json", "j", false, "Output results as JSON")
	rootCmd.Flags().BoolVarP(&fixMode, "fix", "x", false, "Auto-fix fixable issues")

	return rootCmd
}

func Execute() int {
	if err := setupCLI(os.Stdout, os.Stderr).Execute(); err != nil {
		return 1
	}
	return 0
}
