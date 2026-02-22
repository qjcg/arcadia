package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/qjcg/arcadia/exp/sv/internal/discovery"
	"github.com/qjcg/arcadia/exp/sv/internal/git"
	"github.com/qjcg/arcadia/exp/sv/internal/semver"
	"github.com/spf13/cobra"
)

var (
	modulePath string
	allModules bool
	verbose    bool
)

func main() {
	setupCLI(os.Stdout, os.Stderr).Execute()
}

func setupCLI(out, err io.Writer) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "sv",
		Short: "sv is a semantic versioning tool for monorepos",
	}
	rootCmd.SetOut(out)
	rootCmd.SetErr(err)

	nextCmd := &cobra.Command{
		Use:   "next",
		Short: "Calculate the next version",
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := git.Root()
			if err != nil {
				return err
			}

			var modules []discovery.Module
			if allModules {
				var err error
				modules, err = discovery.FindModules(root)
				if err != nil {
					return err
				}
			} else {
				var m discovery.Module
				var err error
				if modulePath != "" {
					absPath, err := filepath.Abs(modulePath)
					if err != nil {
						return err
					}
					rel, err := filepath.Rel(root, absPath)
					if err != nil {
						return err
					}
					m = discovery.Module{Name: filepath.ToSlash(rel), Path: absPath}
				} else {
					cwd, _ := os.Getwd()
					m, err = discovery.GetCurrentModule(root, cwd)
					if err != nil {
						return err
					}
				}
				modules = []discovery.Module{m}
			}

			for _, m := range modules {
				if err := runNext(cmd.OutOrStdout(), root, m, verbose); err != nil {
					fmt.Fprintf(cmd.ErrOrStderr(), "Error for module %s: %v\n", m.Name, err)
				}
			}
			return nil
		},
	}

	currentCmd := &cobra.Command{
		Use:   "current",
		Short: "Show the current version",
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := git.Root()
			if err != nil {
				return err
			}

			var modules []discovery.Module
			if allModules {
				var err error
				modules, err = discovery.FindModules(root)
				if err != nil {
					return err
				}
			} else {
				var m discovery.Module
				var err error
				if modulePath != "" {
					absPath, err := filepath.Abs(modulePath)
					if err != nil {
						return err
					}
					rel, err := filepath.Rel(root, absPath)
					if err != nil {
						return err
					}
					m = discovery.Module{Name: filepath.ToSlash(rel), Path: absPath}
				} else {
					cwd, _ := os.Getwd()
					m, err = discovery.GetCurrentModule(root, cwd)
					if err != nil {
						return err
					}
				}
				modules = []discovery.Module{m}
			}

			for _, m := range modules {
				if err := runCurrent(cmd.OutOrStdout(), root, m, verbose); err != nil {
					fmt.Fprintf(cmd.ErrOrStderr(), "Error for module %s: %v\n", m.Name, err)
				}
			}
			return nil
		},
	}

	nextCmd.Flags().BoolVarP(&allModules, "all", "a", false, "Calculate next version for all modules")
	nextCmd.Flags().StringVarP(&modulePath, "path", "p", "", "Explicit module path")
	nextCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show module names with tags (used with -a)")

	currentCmd.Flags().BoolVarP(&allModules, "all", "a", false, "Show current version for all modules")
	currentCmd.Flags().StringVarP(&modulePath, "path", "p", "", "Explicit module path")
	currentCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show module names with tags (used with -a)")

	rootCmd.AddCommand(nextCmd, currentCmd,
		createBumpCmd("major", semver.BumpMajor, "Force a major version bump"),
		createBumpCmd("minor", semver.BumpMinor, "Force a minor version bump"),
		createBumpCmd("patch", semver.BumpPatch, "Force a patch version bump"),
	)
	return rootCmd
}

func createBumpCmd(name string, bump semver.Bump, doc string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   name,
		Short: doc,
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := git.Root()
			if err != nil {
				return err
			}

			var modules []discovery.Module
			if allModules {
				var err error
				modules, err = discovery.FindModules(root)
				if err != nil {
					return err
				}
			} else {
				var m discovery.Module
				var err error
				if modulePath != "" {
					absPath, err := filepath.Abs(modulePath)
					if err != nil {
						return err
					}
					rel, err := filepath.Rel(root, absPath)
					if err != nil {
						return err
					}
					m = discovery.Module{Name: filepath.ToSlash(rel), Path: absPath}
				} else {
					cwd, _ := os.Getwd()
					m, err = discovery.GetCurrentModule(root, cwd)
					if err != nil {
						return err
					}
				}
				modules = []discovery.Module{m}
			}

			for _, m := range modules {
				if err := runBump(cmd.OutOrStdout(), root, m, bump, verbose); err != nil {
					fmt.Fprintf(cmd.ErrOrStderr(), "Error for module %s: %v\n", m.Name, err)
				}
			}
			return nil
		},
	}
	cmd.Flags().BoolVarP(&allModules, "all", "a", false, "Bump all modules")
	cmd.Flags().StringVarP(&modulePath, "path", "p", "", "Explicit module path")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show module names with tags (used with -a)")
	return cmd
}

func runBump(out io.Writer, root string, m discovery.Module, bump semver.Bump, verbose bool) error {
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

func runCurrent(out io.Writer, root string, m discovery.Module, verbose bool) error {
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

func runNext(out io.Writer, root string, m discovery.Module, verbose bool) error {
	tag, err := git.LatestTag(root, m.Name)
	if err != nil {
		return err
	}

	commits, err := git.CommitsSince(root, tag, m.Name)
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
