// Command ap generates and validates Agent Plugin packages under agents/plugins.
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/qjcg/arcadia/cmd/ap/internal/spec"
	"github.com/qjcg/arcadia/cmd/ap/internal/sync"
	"github.com/qjcg/arcadia/cmd/ap/internal/validate"
)

func main() {
	os.Exit(execute())
}

func execute() int {
	var (
		specsDir string
		outDir   string
	)

	root := &cobra.Command{
		Use:   "ap",
		Short: "Generate and validate Agent Plugin packages",
	}

	syncCmd := &cobra.Command{
		Use:   "sync",
		Short: "Regenerate plugin packages from their sync specs",
		RunE: func(cmd *cobra.Command, args []string) error {
			specs, err := spec.Load(specsDir)
			if err != nil {
				return err
			}
			if len(specs) == 0 {
				return fmt.Errorf("no sync specs found in %s", specsDir)
			}
			for _, s := range specs {
				fmt.Fprintf(cmd.OutOrStdout(), "syncing %s from %s\n", s.Name, sync.String(s))
			}
			roots, err := sync.Run(specs, ".", outDir)
			if err != nil {
				return err
			}
			for _, root := range roots {
				fmt.Fprintf(cmd.OutOrStdout(), "wrote %s\n", root)
			}
			return nil
		},
	}
	syncCmd.Flags().StringVar(&specsDir, "specs", "agents/plugins/specs", "sync specs directory")
	syncCmd.Flags().StringVar(&outDir, "out", "agents/plugins", "output directory for generated packages")

	validateCmd := &cobra.Command{
		Use:   "validate PATH...",
		Short: "Validate plugin packages or bare skill trees",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("at least one path is required")
			}
			failed := false
			for _, path := range args {
				errs, err := validate.Validate(path)
				if err != nil {
					return err
				}
				if len(errs) == 0 {
					fmt.Fprintf(cmd.OutOrStdout(), "ok      %s\n", path)
					continue
				}
				failed = true
				for _, e := range errs {
					fmt.Fprintf(cmd.OutOrStdout(), "invalid %v\n", e)
				}
			}
			if failed {
				return fmt.Errorf("validation failed")
			}
			return nil
		},
	}

	root.AddCommand(syncCmd, validateCmd)
	if err := root.Execute(); err != nil {
		return 1
	}
	return 0
}
