package main

import (
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/qjcg/arcadia/exp/skillo/internal/extract"
	"github.com/qjcg/arcadia/exp/skillo/internal/manifest"
	"github.com/qjcg/arcadia/exp/skillo/internal/skilldirs"
	"github.com/qjcg/arcadia/exp/skillo/internal/types"
)

var version = "0.0.1"

// readSkillsFrom returns a set of skill names from a directory.
func readSkillsFrom(dir string) (map[string]bool, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]bool), nil
		}
		return nil, err
	}
	skills := make(map[string]bool)
	for _, e := range entries {
		if e.IsDir() {
			skills[e.Name()] = true
		}
	}
	return skills, nil
}

// buildManifestVersionMap loads the manifest and returns a map of skill name -> version.
func buildManifestVersionMap(m *manifest.Manifest) map[string]string {
	skillVersion := make(map[string]string)
	if m == nil {
		return skillVersion
	}
	for modPath, skills := range m.ModuleSkills {
		ver := m.ModuleVersions[modPath]
		if ver == "" {
			ver = "latest"
		}
		for _, s := range skills {
			skillVersion[s] = ver
		}
	}
	return skillVersion
}

// buildModuleMap returns a map of skill name -> module path.
func buildModuleMap(m *manifest.Manifest) map[string]string {
	skillModule := make(map[string]string)
	if m == nil {
		return skillModule
	}
	for modPath, skills := range m.ModuleSkills {
		for _, s := range skills {
			skillModule[s] = modPath
		}
	}
	return skillModule
}

// readSkillDescriptions reads descriptions from SKILL.md files in a skills directory.
func readSkillDescriptions(dir string) map[string]string {
	descs := make(map[string]string)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return descs
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		skillMDPath := filepath.Join(dir, e.Name(), "SKILL.md")
		data, err := os.ReadFile(skillMDPath)
		if err != nil {
			continue
		}
		content := string(data)
		lines := strings.Split(content, "\n")
		if len(lines) < 2 || lines[0] != "---" {
			continue
		}
		var frontmatter []string
		for i := 1; i < len(lines); i++ {
			if lines[i] == "---" {
				break
			}
			frontmatter = append(frontmatter, lines[i])
		}
		fmStr := strings.Join(frontmatter, "\n")
		var skill types.Skill
		if err := yaml.Unmarshal([]byte(fmStr), &skill); err != nil {
			continue
		}
		if skill.Description != "" {
			descs[e.Name()] = skill.Description
		}
	}
	return descs
}

func newRootCmd() *cobra.Command {
	var skillsDir string
	var modulesDir string
	var userFlag bool

	home, _ := os.UserHomeDir()
	cwd, _ := os.Getwd()
	defaultModules := filepath.Join(home, ".skillo")

	rootCmd := &cobra.Command{
		Use:     "skillo",
		Short:   "Agent skills manager",
		Version: version,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Only compute default if --skills-dir wasn't explicitly set
			if !cmd.Flags().Changed("skills-dir") {
				skillsDir = skilldirs.DefaultDir(home, cwd)
			}
			return nil
		},
	}

	// Compute user-level dir
	userSkills := skilldirs.UserDir(home)

	rootCmd.PersistentFlags().StringVar(&modulesDir, "modules-dir", defaultModules, "Go modules dir")

	// Register the skills-dir flag but compute default dynamically
	rootCmd.PersistentFlags().StringVar(&skillsDir, "skills-dir", "", "Skills directory (default: auto-detect: project .agents/skills/ in a git repo, else ~/.agents/skills/)")
	rootCmd.MarkPersistentFlagFilename("skills-dir")

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
		Long:  "Install a skill from a Git repository. Version can be specified as @latest, @v1.2.3, @main, etc. If no version is specified, @latest is used by default. Use --user to install to ~/.agents/skills/ instead of the project dir.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			targetDir := skillsDir
			if userFlag {
				targetDir = userSkills
			}

			if err := os.MkdirAll(modulesDir, 0o755); err != nil {
				return fmt.Errorf("mkdir %s: %w", modulesDir, err)
			}
			if err := os.MkdirAll(targetDir, 0o755); err != nil {
				return fmt.Errorf("mkdir %s: %w", targetDir, err)
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
			return extract.ExtractSkills(moduleDir, targetDir, modulesDir, module, version)
		},
	}
	getCmd.Flags().BoolVarP(&userFlag, "user", "u", false, "Install to ~/.agents/skills/ instead of the primary skills dir")

	listCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List installed skills",
		RunE: func(cmd *cobra.Command, args []string) error {
			showUser, _ := cmd.Flags().GetBool("user")
			showProject, _ := cmd.Flags().GetBool("project")
			showOutdated, _ := cmd.Flags().GetBool("outdated")
			format, _ := cmd.Flags().GetString("format")

			// Default: show primary + secondary (user-only)
			showPrimary := !showUser || showProject
			showSecondary := !showProject || showUser

			// If user didn't pass any filter, also show secondary
			if !showUser && !showProject {
				showSecondary = true
			}

			m, _ := manifest.Load(modulesDir)
			skillVersion := buildManifestVersionMap(m)
			skillModule := buildModuleMap(m)

			// Read primary (project or user-default) skills
			primarySkills := make(map[string]bool)
			if showPrimary {
				var err error
				primarySkills, err = readSkillsFrom(skillsDir)
				if err != nil {
					return fmt.Errorf("read skills dir: %w", err)
				}
			}

			// Read secondary (user-level) skills
			secondarySkills := make(map[string]bool)
			hasSeparateSecondary := skillsDir != userSkills
			if showSecondary && hasSeparateSecondary {
				var err error
				secondarySkills, err = readSkillsFrom(userSkills)
				if err != nil {
					return fmt.Errorf("read user skills dir: %w", err)
				}
			}

			// Read descriptions from SKILL.md
			descriptions := readSkillDescriptions(skillsDir)
			if hasSeparateSecondary {
				userDescs := readSkillDescriptions(userSkills)
				maps.Copy(descriptions, userDescs)
			}

			// Collect all skill entries
			type skillEntry struct {
				Name        string `json:"name"`
				Version     string `json:"version,omitempty"`
				Module      string `json:"module,omitempty"`
				Description string `json:"description,omitempty"`
				Location    string `json:"location,omitempty"`
				Outdated    bool   `json:"outdated"`
			}

			var entries []skillEntry
			for name := range primarySkills {
				entry := skillEntry{
					Name:        name,
					Version:     skillVersion[name],
					Module:      skillModule[name],
					Description: descriptions[name],
				}
				entries = append(entries, entry)
			}
			for name := range secondarySkills {
				if primarySkills[name] {
					continue // already shown
				}
				entry := skillEntry{
					Name:        name,
					Version:     skillVersion[name],
					Module:      skillModule[name],
					Description: descriptions[name],
					Location:    "user",
				}
				entries = append(entries, entry)
			}

			// Check for outdated skills via Go proxy
			if showOutdated || cmd.Flags().Changed("outdated") {
				latestVersion := make(map[string]string)
				out, err := exec.Command("go", "list", "-u", "-m", "-f", "{{.Path}} {{.Version}}{{if .Update}} {{.Update.Version}}{{end}}", "all").CombinedOutput()
				if err == nil {
					for line := range strings.SplitSeq(string(out), "\n") {
						parts := strings.Fields(line)
						if len(parts) >= 3 {
							latestVersion[parts[0]] = parts[2]
						}
					}
					for i, e := range entries {
						if mod := e.Module; mod != "" {
							if latest, ok := latestVersion[mod]; ok {
								if latest != e.Version {
									entries[i].Outdated = true
								}
							}
						}
					}
				}
			}

			// If --outdated flag is set, filter to only show outdated
			if showOutdated {
				var filtered []skillEntry
				for _, e := range entries {
					if e.Outdated {
						filtered = append(filtered, e)
					}
				}
				entries = filtered
			}

			// Output
			switch format {
			case "json":
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				if len(entries) == 0 {
					fmt.Println("[]")
					return nil
				}
				return enc.Encode(entries)
			default:
				// Table output
				for _, e := range entries {
					var buf strings.Builder
					buf.WriteString(e.Name)
					if e.Version != "" {
						buf.WriteString(fmt.Sprintf("  %s", e.Version))
					}
					if e.Outdated {
						buf.WriteString(" (outdated)")
					}
					if e.Location == "user" {
						buf.WriteString(" (user)")
					}
					if e.Module != "" {
						buf.WriteString(fmt.Sprintf("  [%s]", e.Module))
					}
					if e.Description != "" {
						buf.WriteString(fmt.Sprintf("  # %s", e.Description))
					}
					fmt.Println(buf.String())
				}
				return nil
			}
		},
	}
	listCmd.Flags().BoolP("user", "u", false, "Show only user-level skills")
	listCmd.Flags().BoolP("project", "p", false, "Show only project-level skills")
	listCmd.Flags().Bool("outdated", false, "Show only skills with available upgrades")
	listCmd.Flags().String("format", "table", "Output format (table, json)")

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
			listCmd := exec.Command("go", "list", "-m", "-f", "{{if not .Main}}{{.Path}} {{.Dir}} {{.Version}}{{end}}", "all")
			listCmd.Dir = modulesDir
			listOut, err := listCmd.CombinedOutput()
			if err != nil {
				return fmt.Errorf("go list modules: %v\n%s", err, listOut)
			}
			lines := strings.SplitSeq(strings.TrimSpace(string(listOut)), "\n")
			for line := range lines {
				if line == "" {
					continue
				}
				parts := strings.Fields(line)
				if len(parts) >= 2 {
					moduleDir := parts[1]
					version := "latest"
					if len(parts) >= 3 {
						version = parts[2]
					}
					if err := extract.ExtractSkills(moduleDir, skillsDir, modulesDir, parts[0], version); err != nil {
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
		Long:  "Remove a skill module from the workspace. By default removes from the primary skills dir. Use --user to remove from ~/.agents/skills/.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := os.MkdirAll(modulesDir, 0o755); err != nil {
				return fmt.Errorf("mkdir %s: %w", modulesDir, err)
			}
			module := args[0]

			// Remove from manifest to get associated skill names
			m, _ := manifest.Load(modulesDir)
			skillNames := m.RemoveModule(module)

			// Determine which dir to clean up
			targetDir := skillsDir
			if userFlag {
				targetDir = userSkills
			}

			// Remove skill directories
			for _, name := range skillNames {
				skillPath := filepath.Join(targetDir, name)
				if err := os.RemoveAll(skillPath); err != nil {
					fmt.Printf("Warning: failed to remove %s: %v\n", skillPath, err)
				}
			}

			// Update manifest
			if err := m.Save(modulesDir); err != nil {
				return fmt.Errorf("save manifest: %w", err)
			}

			// Remove from go module
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
	removeCmd.Flags().BoolVarP(&userFlag, "user", "u", false, "Remove from ~/.agents/skills/ instead of the primary skills dir")

	searchCmd := &cobra.Command{
		Use:   "search [query]",
		Short: "Search for skills",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			query := args[0]
			fmt.Printf("Searching for '%s'...\n", query)

			// Search primary dir
			primaryFound := make(map[string]bool)
			entries, err := os.ReadDir(skillsDir)
			if err == nil {
				for _, entry := range entries {
					if entry.IsDir() && strings.Contains(entry.Name(), query) {
						fmt.Println(entry.Name())
						primaryFound[entry.Name()] = true
					}
				}
			}

			// Search secondary (user) dir if different from primary
			hasSeparateSecondary := skillsDir != userSkills
			if hasSeparateSecondary {
				entries, err := os.ReadDir(userSkills)
				if err == nil {
					for _, entry := range entries {
						if entry.IsDir() && strings.Contains(entry.Name(), query) && !primaryFound[entry.Name()] {
							fmt.Printf("%s (user)\n", entry.Name())
						}
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
			return extract.ExtractSkills(args[0], tempDir, "", "", "")
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
