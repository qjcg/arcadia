package cli

import (
	"fmt"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/qjcg/arcadia/cmd/ap/internal/validate"
	"github.com/spf13/cobra"
)

type validateParams struct {
	Paths []string `descr:"Plugin package root(s) or bare skill tree(s) to validate" positional:"true"`
}

func createValidateCmd() *cobra.Command {
	return boa.CmdT[validateParams]{
		Use:   "validate",
		Short: "Validate plugin packages or bare skill trees",
		Long: "Validate against Agent Plugins v1.0.0 and the Agent Skills spec.\n\n" +
			"Each path may be a plugin package root (contains plugin.json), a bare skill " +
			"tree (children contain SKILL.md), or a directory of plugin package roots. " +
			"Violations are listed one per line; any violation fails the run.",
		RunFuncE: func(p *validateParams, cmd *cobra.Command, _ []string) error {
			return runValidateCmd(p, cmd)
		},
	}.ToCmd().ToCobra()
}

func runValidateCmd(p *validateParams, cmd *cobra.Command) error {
	failed := false
	for _, path := range p.Paths {
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
}
