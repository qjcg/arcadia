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

type nextParams struct {
	All     bool   `descr:"Calculate next version for all modules"`
	Path    string `descr:"Explicit module path" optional:"true"`
	Verbose bool   `descr:"Show module names with tags (used with -a)"`
}

func createNextCmd() *cobra.Command {
	return boa.CmdT[nextParams]{
		Use:   "next",
		Short: "Calculate the next version",
		RunFuncE: func(p *nextParams, cmd *cobra.Command, _ []string) error {
			return runNextCmd(p, cmd)
		},
	}.ToCmd().ToCobra()
}

func runNextCmd(p *nextParams, cmd *cobra.Command) error {
	root, err := git.Root()
	if err != nil {
		return err
	}

	allMods, err := discovery.FindModules(root)
	if err != nil {
		return err
	}

	modules, err := getModules(root, p.Path, p.All)
	if err != nil {
		return err
	}

	for _, m := range modules {
		if err := runNext(cmd.OutOrStdout(), root, m, allMods, p.All, p.Verbose); err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "Error for module %s: %v\n", m.Name, err)
		}
	}
	return nil
}

func runNext(out io.Writer, root string, m discovery.Module, allModulesList []discovery.Module, allModules, verbose bool) error {
	tag, err := git.LatestTag(root, m.Name)
	if err != nil {
		return err
	}

	// Build exclude paths: all modules except the current one
	var excludePaths []string
	if m.Name == "." {
		for _, mod := range allModulesList {
			if mod.Name != "." {
				excludePaths = append(excludePaths, mod.Name)
			}
		}
	}

	commits, err := git.CommitsSince(root, tag, m.Name, excludePaths)
	if err != nil {
		return err
	}

	next, err := semver.CalculateNext(tag, m.Name, commits)
	if err != nil {
		return err
	}

	if next == tag && tag != "" {
		return nil // No change
	}

	if verbose && allModules {
		fmt.Fprintf(out, "%s -> %s\n", m.Name, next)
	} else {
		fmt.Fprintln(out, next)
	}
	return nil
}