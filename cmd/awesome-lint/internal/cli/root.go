package cli

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"

	"github.com/qjcg/arcadia/cmd/awesome-lint/internal/linter"
	"github.com/spf13/cobra"
)

// resolveDefaultFile finds a case-insensitive match for the given path in its
// directory. This lets a default like README.md match README.MD, Readme.md,
// etc. on case-sensitive filesystems.
func resolveDefaultFile(path string) string {
	dir := filepath.Dir(path)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return path
	}
	want := strings.ToLower(filepath.Base(path))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if strings.ToLower(entry.Name()) == want {
			return filepath.Join(dir, entry.Name())
		}
	}
	return path
}

// writeJSON writes lint results as JSON. A single file is encoded as one
// object; multiple files are encoded as an array to keep output unambiguous.
func writeJSON(w io.Writer, results []*linter.Results) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if len(results) == 1 {
		return enc.Encode(results[0])
	}
	return enc.Encode(results)
}

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
		Use:   "awesome-lint [file...]",
		Short: "Lint an Awesome list for compliance with awesome.re guidelines",
		Long: `Lint an Awesome list for compliance with awesome.re guidelines.

If no file is specified, README.md in the current directory is used.
Accepts one or more local file paths or GitHub repository URLs.`,
		Args:    cobra.ArbitraryArgs,
		Version: getVersion(),
		RunE: func(cmd *cobra.Command, args []string) error {
			paths := args
			if len(paths) == 0 {
				paths = []string{resolveDefaultFile(filename)}
			}

			r := linter.New()

			var results []*linter.Results
			for _, path := range paths {
				var res *linter.Results
				var lintErr error
				if fixMode {
					res, lintErr = r.LintWithFix(path)
				} else {
					res, lintErr = r.Lint(path)
				}
				if lintErr != nil {
					return lintErr
				}
				results = append(results, res)
			}

			if jsonOutput {
				return writeJSON(cmd.OutOrStdout(), results)
			}

			hasErrors := false
			for _, res := range results {
				res.WritePretty(cmd.OutOrStdout())
				hasErrors = hasErrors || res.HasErrors()
			}
			if hasErrors {
				os.Exit(1)
			}
			return nil
		},
	}
	rootCmd.SetOut(out)
	rootCmd.SetErr(err)

	rootCmd.Flags().StringVarP(&filename, "filename", "f", "README.md", "Path to the markdown file to lint")
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
