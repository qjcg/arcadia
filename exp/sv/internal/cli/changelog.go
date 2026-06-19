package cli

import (
	"fmt"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/qjcg/arcadia/exp/sv/internal/changelog"
	"github.com/qjcg/arcadia/exp/sv/internal/git"
	"github.com/spf13/cobra"
)

type changelogParams struct {
	All       bool     `descr:"Generate changelog for all modules"`
	Path      []string `descr:"Explicit module path(s)" optional:"true"`
	From      string   `descr:"Start version (inclusive)" optional:"true"`
	To        string   `descr:"End version (inclusive)" optional:"true"`
	Since     string   `descr:"Changelog since a date (year like 2025, duration like 8w, or ISO date)" optional:"true"`
	Dir       string   `descr:"Directory to write individual changelog entry files" short:"d" optional:"true"`
	URLPrefix string   `descr:"URL prefix for linking commit hashes in changelog items" short:"u" optional:"true"`
}

func createChangelogCmd() *cobra.Command {
	return boa.CmdT[changelogParams]{
		Use:   "changelog",
		Short: "Generate a changelog from conventional commits",
		RunFuncE: func(p *changelogParams, cmd *cobra.Command, _ []string) error {
			return runChangelogCmd(p, cmd)
		},
	}.ToCmd().ToCobra()
}

func runChangelogCmd(p *changelogParams, cmd *cobra.Command) error {
	// Set URL prefix for formatted items
	changelog.URLPrefix = p.URLPrefix

	root, err := git.Root()
	if err != nil {
		return err
	}

	modules, err := getModules(root, p.Path, p.All)
	if err != nil {
		return err
	}

	// If --since is specified, use date-based generation
	if p.Since != "" {
		sinceDate, err := changelog.ParseSince(p.Since)
		if err != nil {
			return err
		}

		for _, m := range modules {
			cl, err := changelog.GenerateSinceDate(root, m.Name, sinceDate)
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "Error for module %s: %v\n", m.Name, err)
				continue
			}

			output := changelog.FormatChangelog(cl)
			fmt.Fprint(cmd.OutOrStdout(), output)

			if p.Dir != "" {
				if err := changelog.WriteEntryDir(p.Dir, cl); err != nil {
					fmt.Fprintf(cmd.ErrOrStderr(), "Error writing entries for module %s: %v\n", m.Name, err)
				}
			}
		}
		return nil
	}

	// Use tag-based generation
	for _, m := range modules {
		cl, err := changelog.GenerateFromGit(root, m.Name, p.From, p.To)
		if err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "Error for module %s: %v\n", m.Name, err)
			continue
		}

		// Load overview files if --dir was specified
		if p.Dir != "" {
			overviews, loadErr := changelog.LoadOverviewFiles(p.Dir)
			if loadErr == nil && len(overviews) > 0 {
				for i := range cl.Entries {
					if ov, ok := overviews[cl.Entries[i].Version]; ok {
						cl.Entries[i].Overview = ov
					}
				}
			}
		}

		output := changelog.FormatChangelog(cl)
		fmt.Fprint(cmd.OutOrStdout(), output)

		if p.Dir != "" {
			if err := changelog.WriteEntryDir(p.Dir, cl); err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "Error writing entries for module %s: %v\n", m.Name, err)
			}
		}
	}

	return nil
}
