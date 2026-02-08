package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/qjcg/arcadia/x/sv/internal/discovery"
	"github.com/qjcg/arcadia/x/sv/internal/git"
	"github.com/qjcg/arcadia/x/sv/internal/semver"
	"github.com/spf13/cobra"
)

var (
	modulePath string
	allModules bool
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

			if allModules {
				modules, err := discovery.FindModules(root)
				if err != nil {
					return err
				}
				for _, m := range modules {
					if err := runNext(cmd.OutOrStdout(), root, m); err != nil {
						fmt.Fprintf(cmd.ErrOrStderr(), "Error for module %s: %v\n", m.Name, err)
					}
				}
				return nil
			}

			var m discovery.Module
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

			return runNext(cmd.OutOrStdout(), root, m)
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

			cwd, _ := os.Getwd()
			m, err := discovery.GetCurrentModule(root, cwd)
			if err != nil {
				return err
			}

			tag, err := git.LatestTag(root, m.Name)
			if err != nil {
				return err
			}
			if tag == "" {
				fmt.Fprintln(cmd.OutOrStdout(), "no tags found")
			} else {
				fmt.Fprintln(cmd.OutOrStdout(), tag)
			}
			return nil
		},
	}

	nextCmd.Flags().BoolVarP(&allModules, "all", "a", false, "Calculate next version for all modules")
	nextCmd.Flags().StringVarP(&modulePath, "path", "p", "", "Explicit module path")

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

			if allModules {
				modules, err := discovery.FindModules(root)
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

			var m discovery.Module
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

			return runBump(cmd.OutOrStdout(), root, m, bump)
		},
	}
	cmd.Flags().BoolVarP(&allModules, "all", "a", false, "Bump all modules")
	cmd.Flags().StringVarP(&modulePath, "path", "p", "", "Explicit module path")
	return cmd
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

	if m.Name == "." {
		fmt.Fprintf(out, ". -> %s\n", next)
	} else {
		fmt.Fprintf(out, "%s -> %s\n", m.Name, next)
	}
	return nil
}

func runNext(out io.Writer, root string, m discovery.Module) error {
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

	if m.Name == "." {
		fmt.Fprintf(out, ". -> %s\n", next)
	} else {
		fmt.Fprintf(out, "%s -> %s\n", m.Name, next)
	}
	return nil
}
