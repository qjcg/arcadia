package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/qjcg/arcadia/exp/skillo/internal/extract"
)

var version = "0.0.1"

func newRootCmd() *cobra.Command {
	var skillsDir string
	var modulesDir string

	rootCmd := &cobra.Command{
		Use:     "skillo",
		Short:   "Agent skills manager",
		Version: version,
	}

	home, _ := os.UserHomeDir()
	defaultModules := filepath.Join(home, ".skillo")
	defaultSkills := filepath.Join(home, ".config", "agents", "skills")

	rootCmd.PersistentFlags().StringVar(&modulesDir, "modules-dir", defaultModules, "Go modules dir")
	rootCmd.PersistentFlags().StringVar(&skillsDir, "skills-dir", defaultSkills, "Skills directory")
	viper.BindPFlags(rootCmd.PersistentFlags())

	initCmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize skillo workspace",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := os.MkdirAll(modulesDir, 0o755); err != nil {
				return fmt.Errorf("mkdir %s: %w", modulesDir, err)
			}
			goCmd := exec.Command("go", "mod", "init", "skillo.local/skills")
			goCmd.Dir = modulesDir
			out, err := goCmd.CombinedOutput()
			if err != nil {
				if strings.Contains(string(out), "already exists") {
					fmt.Println("workspace already initialized at " + modulesDir)
					return nil
				}
				return fmt.Errorf("go mod init: %v\n%s", err, out)
			}
			fmt.Println("initialized at " + modulesDir)
			return nil
		},
	}

	getCmd := &cobra.Command{
		Use:   "get [repo@version]",
		Short: "Install skill from Git repo",
		Long:  "Install a skill from a Git repository. Version can be specified as @latest, @v1.2.3, @main, etc. If no version is specified, @latest is used by default.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := os.MkdirAll(modulesDir, 0o755); err != nil {
				return fmt.Errorf("mkdir %s: %w", modulesDir, err)
			}
			if err := os.MkdirAll(skillsDir, 0o755); err != nil {
				return fmt.Errorf("mkdir %s: %w", skillsDir, err)
			}
			moduleSpec := args[0]

			// Parse module and version
			module, version := moduleSpec, "latest"
			if at := strings.LastIndex(moduleSpec, "@"); at > 0 && at < len(moduleSpec)-1 {
				module = moduleSpec[:at]
				version = moduleSpec[at+1:]
			}

			goCmd := exec.Command("go", "get", module+"@"+version)
			goCmd.Dir = modulesDir
			out, err := goCmd.CombinedOutput()
			if err != nil {
				return fmt.Errorf("go get %s@%s: %v\n%s", module, version, err, out)
			}
			fmt.Printf("Installed %s@%s\n", module, version)
			listCmd := exec.Command("go", "list", "-m", "-f", "{{.Dir}}", module)
			listCmd.Dir = modulesDir
			listOut, err := listCmd.CombinedOutput()
			if err != nil {
				return fmt.Errorf("go list: %v\n%s", err, listOut)
			}
			moduleDir := strings.TrimSpace(string(listOut))
			return extract.ExtractSkills(moduleDir, skillsDir, modulesDir, module)
		},
	}

	listCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List installed skills",
		RunE: func(cmd *cobra.Command, args []string) error {
			entries, err := os.ReadDir(skillsDir)
			if err != nil {
				if os.IsNotExist(err) {
					return nil
				}
				return fmt.Errorf("read skills dir: %w", err)
			}
			for _, entry := range entries {
				if entry.IsDir() {
					fmt.Println(entry.Name())
				}
			}
			return nil
		},
	}

	updateCmd := &cobra.Command{
		Use:   "update",
		Short: "Update all skills",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := os.MkdirAll(modulesDir, 0o755); err != nil {
				return fmt.Errorf("mkdir %s: %w", modulesDir, err)
			}
			if err := os.MkdirAll(skillsDir, 0o755); err != nil {
				return fmt.Errorf("mkdir %s: %w", skillsDir, err)
			}
			goCmd := exec.Command("go", "get", "-u", "./...")
			goCmd.Dir = modulesDir
			out, err := goCmd.CombinedOutput()
			if err != nil {
				return fmt.Errorf("go get -u: %v\n%s", err, out)
			}
			fmt.Println("Updated all modules. Re-extracting skills...")
			listCmd := exec.Command("go", "list", "-m", "-f", "{{if not .Main}}{{.Path}} {{.Dir}}{{end}}", "all")
			listCmd.Dir = modulesDir
			listOut, err := listCmd.CombinedOutput()
			if err != nil {
				return fmt.Errorf("go list modules: %v\n%s", err, listOut)
			}
			lines := strings.Split(strings.TrimSpace(string(listOut)), "\n")
			for _, line := range lines {
				if line == "" {
					continue
				}
				parts := strings.Fields(line)
				if len(parts) == 2 {
					moduleDir := parts[1]
					if err := extract.ExtractSkills(moduleDir, skillsDir, modulesDir, parts[0]); err != nil {
						fmt.Printf("Warning: failed to extract from %s: %v\n", parts[0], err)
					}
				}
			}
			return nil
		},
	}

	removeCmd := &cobra.Command{
		Use:   "remove [module]",
		Short: "Remove a skill module",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := os.MkdirAll(modulesDir, 0o755); err != nil {
				return fmt.Errorf("mkdir %s: %w", modulesDir, err)
			}
			module := args[0]
			goCmd := exec.Command("go", "mod", "edit", "-droprequire", module)
			goCmd.Dir = modulesDir
			if out, err := goCmd.CombinedOutput(); err != nil {
				return fmt.Errorf("go mod edit: %v\n%s", err, out)
			}
			tidyCmd := exec.Command("go", "mod", "tidy")
			tidyCmd.Dir = modulesDir
			tidyCmd.Run()
			fmt.Printf("Removed %s from workspace\n", module)
			return nil
		},
	}

	searchCmd := &cobra.Command{
		Use:   "search [query]",
		Short: "Search for skills",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := os.MkdirAll(skillsDir, 0o755); err != nil {
				return fmt.Errorf("mkdir %s: %w", skillsDir, err)
			}
			query := args[0]
			fmt.Printf("Searching for '%s'...\n", query)
			entries, err := os.ReadDir(skillsDir)
			if err == nil {
				for _, entry := range entries {
					if entry.IsDir() && strings.Contains(entry.Name(), query) {
						fmt.Println(entry.Name())
					}
				}
			}
			fmt.Println("GitHub search integration coming soon.")
			return nil
		},
	}

	validateCmd := &cobra.Command{
		Use:   "validate [dir]",
		Short: "Validate SKILL.md in a directory",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			tempDir, err := os.MkdirTemp("", "skillo-validate-*")
			if err != nil {
				return err
			}
			defer os.RemoveAll(tempDir)
			return extract.ExtractSkills(args[0], tempDir, "", "")
		},
	}

	rootCmd.AddCommand(initCmd, getCmd, listCmd, updateCmd, removeCmd, validateCmd, searchCmd)
	return rootCmd
}

func main() {
	if err := newRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}
