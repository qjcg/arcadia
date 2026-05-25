package cli

import (
	"fmt"
	"io"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/qjcg/arcadia/exp/sv/internal/discovery"
	"github.com/qjcg/arcadia/exp/sv/internal/git"
	"github.com/qjcg/arcadia/exp/sv/internal/semver"
	"github.com/spf13/cobra"
)

type bumpParams struct {
	All  bool     `descr:"Bump all modules"`
	Path []string `descr:"Explicit module path(s)" optional:"true"`
}

func createBumpCmd(name string, bump semver.Bump, doc string) *cobra.Command {
	return boa.CmdT[bumpParams]{
		Use:   name,
		Short: doc,
		RunFuncE: func(p *bumpParams, cmd *cobra.Command, _ []string) error {
			return runBumpCmd(p, bump, cmd)
		},
	}.ToCmd().ToCobra()
}

func runBumpCmd(p *bumpParams, bump semver.Bump, cmd *cobra.Command) error {
	root, err := git.Root()
	if err != nil {
		return err
	}

	modules, err := getModules(root, p.Path, p.All)
	if err != nil {
		return err
	}

	for _, m := range modules {
		if err := runBump(cmd.OutOrStdout(), root, m, bump); err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "Error for module %s: %v\n", m.Name, err)
		}
	}
	return nil
}

func runBump(out io.Writer, root string, m discovery.Module, bump semver.Bump) error {
	tag, err := git.LatestTag(root, m.Name)
	if err != nil {
		return err
	}

	next, err := semver.Increment(tag, m.Name, bump)
	if err != nil {
		return err
	}

	fmt.Fprintln(out, next)
	return nil
}
