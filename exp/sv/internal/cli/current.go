package cli

import (
	"fmt"
	"io"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/qjcg/arcadia/exp/sv/internal/discovery"
	"github.com/qjcg/arcadia/exp/sv/internal/git"
	"github.com/spf13/cobra"
)

type currentParams struct {
	All  bool     `descr:"Show current version for all modules"`
	Path []string `descr:"Explicit module path(s)" optional:"true"`
}

func createCurrentCmd() *cobra.Command {
	return boa.CmdT[currentParams]{
		Use:   "current",
		Short: "Show the current version",
		RunFuncE: func(p *currentParams, cmd *cobra.Command, _ []string) error {
			return runCurrentCmd(p, cmd)
		},
	}.ToCmd().ToCobra()
}

func runCurrentCmd(p *currentParams, cmd *cobra.Command) error {
	root, err := git.Root()
	if err != nil {
		return err
	}

	modules, err := getModules(root, p.Path, p.All)
	if err != nil {
		return err
	}

	for _, m := range modules {
		if err := runCurrent(cmd.OutOrStdout(), root, m); err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "Error for module %s: %v\n", m.Name, err)
		}
	}
	return nil
}

func runCurrent(out io.Writer, root string, m discovery.Module) error {
	tag, warning, err := latestNonRetractedTag(root, m.Name)
	if err != nil {
		return err
	}

	if tag == "" {
		return nil // Skip modules with no tags
	}

	if warning != "" {
		fmt.Fprintln(out, warning)
	}
	fmt.Fprintln(out, tag)
	return nil
}
