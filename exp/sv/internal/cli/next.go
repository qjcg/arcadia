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
	All          bool     `descr:"Calculate next version for all modules"`
	Path         []string `descr:"Explicit module path(s)" optional:"true"`
	DefaultPatch bool     `descr:"Default to patch bump for non-feat/fix commits" short:"d"`
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
		if err := runNext(cmd.OutOrStdout(), root, m, allMods, p.DefaultPatch); err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "Error for module %s: %v\n", m.Name, err)
		}
	}
	return nil
}

func runNext(out io.Writer, root string, m discovery.Module, allModulesList []discovery.Module, defaultPatch bool) error {
	tag, warning, err := latestNonRetractedTag(root, m.Name)
	if err != nil {
		return err
	}

	if warning != "" {
		fmt.Fprintln(out, warning)
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

	next, err := semver.CalculateNext(tag, m.Name, commits, defaultPatch)
	if err != nil {
		return err
	}

	if next == tag && tag != "" {
		return nil // No change
	}

	fmt.Fprintln(out, next)
	return nil
}
