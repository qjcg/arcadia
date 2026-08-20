package cli

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/qjcg/arcadia/cmd/sv/internal/changelog"
	"github.com/qjcg/arcadia/cmd/sv/internal/git"
	"github.com/qjcg/arcadia/cmd/sv/internal/semver"
	"github.com/spf13/cobra"
)

type changelogParams struct {
	All       bool     `descr:"Generate changelog for all modules"`
	Path      []string `descr:"Explicit module path(s)" optional:"true"`
	Exclude   []string `descr:"Module path(s) to exclude (exact path or directory subtree prefix; repeatable or comma-separated)" optional:"true"`
	From      string   `descr:"Start version (inclusive), or a date (year like 2025, duration like 8w, or ISO date like 2024-01-15)" optional:"true"`
	To        string   `descr:"End version (inclusive)" optional:"true"`
	Dir       string   `descr:"Directory to write individual changelog entry files" short:"d" optional:"true"`
	URLPrefix string   `descr:"URL prefix for linking commit hashes in changelog items" short:"u" optional:"true"`
	Write     bool     `descr:"Write CHANGELOG.md into each module directory" short:"w"`
	Release   bool     `descr:"Emit the pending (not-yet-tagged) version as a dated entry instead of [unreleased]"`
	Date      string   `descr:"Date for the pending release entry (default: today)" optional:"true"`
	Tag       bool     `descr:"Create annotated tags for pending releases; implies --release; requires --write and/or --dir"`
	CommitMsg string   `descr:"Commit message for the changelog commit (default: 'docs: update changelogs for released versions')" optional:"true"`
	TagFormat string   `descr:"Go template for the tag message (default: the new version string)" optional:"true"`
	Ref       string   `descr:"Only run --tag when HEAD is at this ref (default: current HEAD)" optional:"true"`
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

	// --tag implies --release
	release := p.Release || p.Tag

	// --tag requires --write and/or --dir (there must be changelog files to commit)
	if p.Tag && !p.Write && p.Dir == "" {
		err := fmt.Errorf("--tag requires --write and/or --dir (there is nothing changelog-generated to commit and tag against)")
		fmt.Fprintln(cmd.ErrOrStderr(), err)
		return err
	}

	// --ref guard: only run --tag when HEAD is at the requested ref
	if p.Tag {
		if err := checkRef(root, p.Ref); err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
	}

	// Resolve the pending release date (default: today)
	date := p.Date
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}

	modules, err := getModules(root, p.Path, p.Exclude, p.All)
	if err != nil {
		return err
	}

	// Accumulate per-module changelogs for combined output
	moduleChangelogs := make(map[string]*changelog.Changelog)

	// Track written changelog paths and pending releases for --tag
	var writtenPaths []string
	var pendingReleases []*semver.Pending

	for _, m := range modules {
		var cl *changelog.Changelog
		var pending *semver.Pending

		if release {
			cl, pending, err = changelog.GenerateRelease(root, m.Name, date)
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "Error for module %s: %v\n", m.Name, err)
				continue
			}
			if cl == nil {
				// No pending release for this module; fall back to normal generation.
				cl, err = changelog.Generate(root, m.Name, p.From, p.To)
				if err != nil {
					fmt.Fprintf(cmd.ErrOrStderr(), "Error for module %s: %v\n", m.Name, err)
					continue
				}
			}
		} else {
			cl, err = changelog.Generate(root, m.Name, p.From, p.To)
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "Error for module %s: %v\n", m.Name, err)
				continue
			}
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

		if p.Dir != "" {
			paths, werr := changelog.WriteEntryDir(p.Dir, cl)
			if werr != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "Error writing entries for module %s: %v\n", m.Name, werr)
			}
			writtenPaths = append(writtenPaths, paths...)
		}

		// Write CHANGELOG.md into the module's directory if --write is set
		if p.Write {
			if len(cl.Entries) > 0 {
				path, werr := changelog.WriteChangelogFile(m.Path, cl)
				if werr != nil {
					fmt.Fprintf(cmd.ErrOrStderr(), "Error writing CHANGELOG.md for module %s: %v\n", m.Name, werr)
				}
				writtenPaths = append(writtenPaths, path)
			}
		} else if p.Dir == "" {
			moduleChangelogs[m.Name] = cl
		}

		if pending != nil {
			pendingReleases = append(pendingReleases, pending)
		}
	}

	// Output combined changelog
	if p.Dir == "" && len(moduleChangelogs) > 0 {
		if len(moduleChangelogs) == 1 {
			for _, cl := range moduleChangelogs {
				output := changelog.FormatChangelog(cl)
				fmt.Fprint(cmd.OutOrStdout(), output)
			}
		} else {
			output := changelog.FormatMultiModuleChangelog(moduleChangelogs)
			fmt.Fprint(cmd.OutOrStdout(), output)
		}
	}

	// --tag: stage only changelog paths, commit, then create tags at HEAD
	if p.Tag {
		if err := tagAndCommit(root, writtenPaths, pendingReleases, p.CommitMsg, p.TagFormat, cmd); err != nil {
			return err
		}
	}

	return nil
}

// checkRef verifies that HEAD is at the requested ref. When ref is empty, it
// defaults to the current HEAD (always matches). Returns an error on mismatch.
func checkRef(root, ref string) error {
	if ref == "" {
		return nil // default: current HEAD, always matches
	}
	head, err := git.ResolveRef(root, "HEAD")
	if err != nil {
		return err
	}
	target, err := git.ResolveRef(root, ref)
	if err != nil {
		return fmt.Errorf("cannot resolve --ref %q: %w", ref, err)
	}
	if head != target {
		return fmt.Errorf("--tag can only run at ref %q (HEAD is at %s)", ref, head[:7])
	}
	return nil
}

// tagAndCommit stages only the changelog paths that changed, commits them with
// the given message, then creates annotated tags for each pending release at
// the new HEAD (so each tag points at a commit that contains its own entry).
func tagAndCommit(root string, writtenPaths []string, pendingReleases []*semver.Pending, commitMsg, tagFormat string, cmd *cobra.Command) error {
	// Filter written paths to those inside the repo tree (--dir may point outside).
	var stagePaths []string
	for _, p := range writtenPaths {
		rel, err := filepath.Rel(root, p)
		if err == nil && !strings.HasPrefix(rel, "..") {
			stagePaths = append(stagePaths, p)
		}
	}

	// Stage and commit only if there are actual changes.
	if len(stagePaths) > 0 {
		changed, err := git.HasChanges(root, stagePaths)
		if err != nil {
			return err
		}
		if changed {
			if err := git.Stage(root, stagePaths); err != nil {
				return err
			}
			msg := commitMsg
			if msg == "" {
				msg = "docs: update changelogs for released versions"
			}
			if err := git.CommitAnnotated(root, msg); err != nil {
				return err
			}
		}
	}

	// Create tags at HEAD (now the changelog commit).
	for _, pr := range pendingReleases {
		if err := createAnnotatedTag(root, pr.Module, pr.Version, tagFormat, false); err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "Warning: failed to create tag for %s: %v\n", pr.Version, err)
		}
	}
	return nil
}
