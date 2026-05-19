package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime/debug"

	"github.com/GiGurra/boa/pkg/boa"
	"github.com/qjcg/arcadia/exp/sv/internal/discovery"
	"github.com/qjcg/arcadia/exp/sv/internal/git"
	"github.com/qjcg/arcadia/exp/sv/internal/semver"
	"github.com/spf13/cobra"
)

type NextParams struct {
	All     bool   `descr:"Calculate next version for all modules"`
	Path    string `descr:"Explicit module path" optional:"true"`
	Verbose bool   `descr:"Show module names with tags (used with -a)"`
}

type CurrentParams struct {
	All     bool   `descr:"Show current version for all modules"`
	Path    string `descr:"Explicit module path" optional:"true"`
	Verbose bool   `descr:"Show module names with tags (used with -a)"`
}

type BumpParams struct {
	All     bool   `descr:"Bump all modules"`
	Path    string `descr:"Explicit module path" optional:"true"`
	Verbose bool   `descr:"Show module names with tags (used with -a)"`
}

func main() {
	setupCLI(os.Stdout, os.Stderr).Execute()
}

func getVersion() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		return info.Main.Version
	}
	return ""
}

func setupCLI(out, err io.Writer) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:     "sv",
		Short:   "sv is a semantic versioning tool for monorepos",
		Version: getVersion(),
	}
	rootCmd.SetOut(out)
	rootCmd.SetErr(err)

	rootCmd.AddCommand(nextCmd())
	rootCmd.AddCommand(currentCmd())
	rootCmd.AddCommand(createBumpCmd("major", semver.BumpMajor, "Force a major version bump"))
	rootCmd.AddCommand(createBumpCmd("minor", semver.BumpMinor, "Force a minor version bump"))
	rootCmd.AddCommand(createBumpCmd("patch", semver.BumpPatch, "Force a patch version bump"))

	return rootCmd
}

func nextCmd() *cobra.Command {
	return boa.CmdT[NextParams]{
		Use:   "next",
		Short: "Calculate the next version",
		RunFuncE: func(p *NextParams, cmd *cobra.Command, _ []string) error {
			return runNextCmd(p, cmd)
		},
	}.ToCmd().ToCobra()
}

func currentCmd() *cobra.Command {
	return boa.CmdT[CurrentParams]{
		Use:   "current",
		Short: "Show the current version",
		RunFuncE: func(p *CurrentParams, cmd *cobra.Command, _ []string) error {
			return runCurrentCmd(p, cmd)
		},
	}.ToCmd().ToCobra()
}

func createBumpCmd(name string, bump semver.Bump, doc string) *cobra.Command {
	return boa.CmdT[BumpParams]{
		Use:   name,
		Short: doc,
		RunFuncE: func(p *BumpParams, cmd *cobra.Command, _ []string) error {
			return runBumpCmd(p, bump, cmd)
		},
	}.ToCmd().ToCobra()
}

func getModules(root, modulePath string, allModules bool) ([]discovery.Module, error) {
	if allModules {
		return discovery.FindModules(root)
	}

	var m discovery.Module
	var err error
	if modulePath != "" {
		absPath, err := filepath.Abs(modulePath)
		if err != nil {
			return nil, err
		}
		rel, err := filepath.Rel(root, absPath)
		if err != nil {
			return nil, err
		}
		m = discovery.Module{Name: filepath.ToSlash(rel), Path: absPath}
	} else {
		cwd, _ := os.Getwd()
		m, err = discovery.GetCurrentModule(root, cwd)
		if err != nil {
			return nil, err
		}
	}
	return []discovery.Module{m}, nil
}

func runNextCmd(p *NextParams, cmd *cobra.Command) error {
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

func runCurrentCmd(p *CurrentParams, cmd *cobra.Command) error {
	root, err := git.Root()
	if err != nil {
		return err
	}

	modules, err := getModules(root, p.Path, p.All)
	if err != nil {
		return err
	}

	for _, m := range modules {
		if err := runCurrent(cmd.OutOrStdout(), root, m, p.All, p.Verbose); err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "Error for module %s: %v\n", m.Name, err)
		}
	}
	return nil
}

func runBumpCmd(p *BumpParams, bump semver.Bump, cmd *cobra.Command) error {
	root, err := git.Root()
	if err != nil {
		return err
	}

	modules, err := getModules(root, p.Path, p.All)
	if err != nil {
		return err
	}

	for _, m := range modules {
		if err := runBump(cmd.OutOrStdout(), root, m, bump, p.All, p.Verbose); err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "Error for module %s: %v\n", m.Name, err)
		}
	}
	return nil
}

func runBump(out io.Writer, root string, m discovery.Module, bump semver.Bump, allModules, verbose bool) error {
	tag, err := git.LatestTag(root, m.Name)
	if err != nil {
		return err
	}

	next, err := semver.Increment(tag, m.Name, bump)
	if err != nil {
		return err
	}

	if verbose && allModules {
		fmt.Fprintf(out, "%s -> %s\n", m.Name, next)
	} else {
		fmt.Fprintln(out, next)
	}
	return nil
}

func runCurrent(out io.Writer, root string, m discovery.Module, allModules, verbose bool) error {
	tag, err := git.LatestTag(root, m.Name)
	if err != nil {
		return err
	}

	if tag == "" {
		return nil // Skip modules with no tags
	}

	if verbose && allModules {
		fmt.Fprintf(out, "%s -> %s\n", m.Name, tag)
	} else {
		fmt.Fprintln(out, tag)
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
