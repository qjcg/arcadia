package cli

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"text/template"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/qjcg/arcadia/cmd/sv/internal/discovery"
	"github.com/qjcg/arcadia/cmd/sv/internal/git"
	"github.com/qjcg/arcadia/cmd/sv/internal/semver"
	"github.com/spf13/cobra"
)

type tagData struct {
	Version string // Full version string, e.g. "v1.2.3" or "x/mod/v1.2.3"
	Module  string // Module path, e.g. "." or "x/mod"
	Major   string // Major version number
	Minor   string // Minor version number
	Patch   string // Patch version number
	Prefix  string // Module path prefix with trailing slash if any, e.g. "" or "x/mod/"
}

type nextParams struct {
	All          bool     `descr:"Calculate next version for all modules"`
	Path         []string `descr:"Explicit module path(s)" optional:"true"`
	DefaultPatch bool     `descr:"Default to patch bump for non-feat/fix commits" short:"d"`
	Tag          bool     `descr:"Create annotated git tag for the new version"`
	TagFormat    string   `descr:"Go template for the tag message (default: the new version string)" short:"f" optional:"true"`
	DryRun       bool     `descr:"Print git commands that would be executed without running them"`
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
		if err := runNext(cmd.OutOrStdout(), cmd.ErrOrStderr(), root, m, allMods, p.DefaultPatch, p.Tag, p.TagFormat, p.DryRun); err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "Error for module %s: %v\n", m.Name, err)
		}
	}
	return nil
}

func runNext(out, errOut io.Writer, root string, m discovery.Module, allModulesList []discovery.Module, defaultPatch, createTag bool, tagFormat string, dryRun bool) error {
	tag, warning, err := latestNonRetractedTag(root, m.Name)
	if err != nil {
		return err
	}

	if warning != "" && verboseFlag {
		fmt.Fprintln(errOut, warning)
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

	// Convert git.CommitInfo to semver.Commit
	semverCommits := make([]semver.Commit, len(commits))
	for i, c := range commits {
		semverCommits[i] = semver.Commit{Message: c.Message, Files: c.Files}
	}
	next, err := semver.CalculateNext(tag, m.Name, semverCommits, defaultPatch)
	if err != nil {
		return err
	}

	if next == tag && tag != "" {
		return nil // No change
	}

	fmt.Fprintln(out, next)

	if createTag {
		if err := createAnnotatedTag(root, m.Name, next, tagFormat, dryRun); err != nil {
			fmt.Fprintf(errOut, "Warning: failed to create tag for %s: %v\n", next, err)
		}
	}

	return nil
}

// createAnnotatedTag creates an annotated git tag for the given version.
// If tagFormat is empty, the message defaults to the version string.
// Otherwise, tagFormat is interpreted as a Go template with tagData available.
// If dryRun is true, the git command is printed to stdout instead of executed.
func createAnnotatedTag(root, moduleName, version, tagFormat string, dryRun bool) error {
	// Parse version components (strip module prefix if present)
	versionPart := version
	if moduleName != "." && strings.HasPrefix(version, moduleName+"/") {
		versionPart = strings.TrimPrefix(version, moduleName+"/")
	}
	// Strip the leading "v"
	ver := strings.TrimPrefix(versionPart, "v")
	parts := strings.SplitN(ver, ".", 3)
	major, minor, patch := "", "", ""
	if len(parts) > 0 {
		major = parts[0]
	}
	if len(parts) > 1 {
		minor = parts[1]
	}
	if len(parts) > 2 {
		patch = parts[2]
	}

	prefix := ""
	if moduleName != "." {
		prefix = moduleName + "/"
	}

	data := tagData{
		Version: version,
		Module:  moduleName,
		Major:   major,
		Minor:   minor,
		Patch:   patch,
		Prefix:  prefix,
	}

	message := versionPart // default: just the semver, not the full tag path
	if tagFormat != "" {
		tmpl, err := template.New("tagmsg").Parse(tagFormat)
		if err != nil {
			return fmt.Errorf("invalid --tag-format template: %w", err)
		}
		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, data); err != nil {
			return fmt.Errorf("failed to execute --tag-format template: %w", err)
		}
		message = buf.String()
	}

	if dryRun {
		fmt.Printf("git tag -a %s -m %s\n", version, message)
		return nil
	}

	return git.TagAnnotated(root, version, message)
}
