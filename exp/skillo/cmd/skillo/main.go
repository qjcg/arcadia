package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/qjcg/arcadia/exp/skillo/internal/extract"
	"github.com/qjcg/arcadia/exp/skillo/internal/normalize"
	"github.com/qjcg/arcadia/exp/skillo/internal/selections"
	"github.com/qjcg/arcadia/exp/skillo/internal/skilldirs"
	"github.com/qjcg/arcadia/exp/skillo/internal/types"
)

var version = "0.2.0"

func newRootCmd() *cobra.Command {
	var (
		skillsDir   string
		userFlag    bool
		projectFlag bool
	)

	home, _ := os.UserHomeDir()
	cwd, _ := os.Getwd()

	rootCmd := &cobra.Command{
		Use:     "skillo",
		Short:   "Agent skills manager",
		Version: version,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if !cmd.Flags().Changed("skills-dir") {
				sources := skilldirs.Detect(home, cwd)
				skillsDir = sources.Primary
			}
			return nil
		},
	}

	rootCmd.PersistentFlags().StringVar(&skillsDir, "skills-dir", "", "Skills directory (default: auto-detect)")
	rootCmd.PersistentFlags().BoolVarP(&userFlag, "user", "u", false, "Target user scope")
	rootCmd.PersistentFlags().BoolVarP(&projectFlag, "project", "p", false, "Target project scope")

	// resolveScope returns the skillo dir and skills dir for the target scope.
	resolveScope := func() (skilloDir, skillsDirOut string) {
		sources := skilldirs.Detect(home, cwd)
		if userFlag {
			return sources.UserSkilloDir, skilldirs.UserSkillsDir(home)
		}
		if projectFlag && sources.ProjectSkilloDir != "" {
			return sources.ProjectSkilloDir, sources.Primary
		}
		if sources.ProjectSkilloDir != "" {
			return sources.ProjectSkilloDir, sources.Primary
		}
		return sources.UserSkilloDir, skilldirs.UserSkillsDir(home)
	}

	// ─── init ────────────────────────────────────────────────────────────────
	initCmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a skillo workspace",
		RunE: func(cmd *cobra.Command, args []string) error {
			isProject, _ := cmd.Flags().GetBool("project")

			var targetDir string
			if isProject {
				sources := skilldirs.Detect(home, cwd)
				if sources.ProjectSkilloDir == "" {
					return fmt.Errorf("not inside a git repository; use --project from within a repo")
				}
				targetDir = sources.ProjectSkilloDir
			} else {
				legacy := filepath.Join(home, ".skillo")
				userSkillo := skilldirs.UserSkilloDir(home)
				if _, err := os.Stat(legacy); err == nil {
					if _, err := os.Stat(userSkillo); os.IsNotExist(err) {
						fmt.Println("Migrating legacy ~/.skillo/ →", userSkillo)
						if err := os.Rename(legacy, userSkillo); err != nil {
							return fmt.Errorf("migrate legacy dir: %w", err)
						}
						oldManifest := filepath.Join(userSkillo, ".skillo-manifest.json")
						if data, err := os.ReadFile(oldManifest); err == nil {
							var lm struct {
								ModuleSkills map[string][]string `json:"module_skills"`
							}
							if json.Unmarshal(data, &lm) == nil && lm.ModuleSkills != nil {
								sel := selections.ConvertLegacyManifest(lm.ModuleSkills)
								if err := selections.Save(userSkillo, sel); err != nil {
									return fmt.Errorf("convert legacy manifest: %w", err)
								}
								fmt.Println("Converted legacy .skillo-manifest.json to selections.json")
							}
							os.Remove(oldManifest)
						}
					}
				}
				targetDir = userSkillo
			}

			if err := os.MkdirAll(targetDir, 0o755); err != nil {
				return fmt.Errorf("mkdir %s: %w", targetDir, err)
			}

			goCmd := exec.Command("go", "mod", "init", "skillo.local/skills")
			goCmd.Dir = targetDir
			out, err := goCmd.CombinedOutput()
			if err != nil {
				if !strings.Contains(string(out), "already exists") {
					return fmt.Errorf("go mod init: %v\n%s", err, out)
				}
				fmt.Println("workspace already initialized at", targetDir)
			} else {
				fmt.Println("initialized at", targetDir)
			}

			if err := selections.Init(targetDir); err != nil {
				return fmt.Errorf("init selections: %w", err)
			}
			return nil
		},
	}
	initCmd.Flags().Bool("project", false, "Initialize project .skillo/ instead of user ~/.config/skillo/")

	// ─── add ─────────────────────────────────────────────────────────────────
	addCmd := &cobra.Command{
		Use:   "add <repo>[@version]",
		Short: "Register a module and install its skills",
		Long: `Register a Go module and install its skills. The repo can be specified as:
  - Full Go import path: github.com/user/repo
  - Short form: user/repo (github.com/ is prepended)
  - HTTPS URL: https://github.com/user/repo

Use --skill to install only specific skills from the module.
Use --user or --project to target a specific scope.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			skillList, _ := cmd.Flags().GetStringSlice("skill")

			skilloDir, targetSkillsDir := resolveScope()
			if err := os.MkdirAll(skilloDir, 0o755); err != nil {
				return fmt.Errorf("mkdir %s: %w", skilloDir, err)
			}
			if err := os.MkdirAll(targetSkillsDir, 0o755); err != nil {
				return fmt.Errorf("mkdir %s: %w", targetSkillsDir, err)
			}

			module, version, err := normalize.ModulePath(args[0])
			if err != nil {
				return err
			}

			goGet := exec.Command("go", "get", module+"@"+version)
			goGet.Dir = skilloDir
			out, err := goGet.CombinedOutput()
			if err != nil {
				return fmt.Errorf("go get %s@%s: %v\n%s", module, version, err, out)
			}
			fmt.Printf("Installed %s@%s\n", module, version)

			// Ensure module is fully resolved before listing its directory.
			// go get does not always make the module visible to go list -m
			// when fetched over the network.
			goDown := exec.Command("go", "mod", "download")
			goDown.Dir = skilloDir
			goDown.Run()

			goList := exec.Command("go", "list", "-m", "-f", "{{.Dir}}", module)
			goList.Dir = skilloDir
			listOut, err := goList.CombinedOutput()
			if err != nil {
				return fmt.Errorf("go list: %v\n%s", err, listOut)
			}
			moduleDir := strings.TrimSpace(string(listOut))

			var extracted []string
			if len(skillList) > 0 {
				available, err := extract.ListAvailableSkills(moduleDir)
				if err != nil {
					return fmt.Errorf("scan module: %w", err)
				}
				availSet := make(map[string]bool, len(available))
				for _, a := range available {
					availSet[a] = true
				}
				for _, s := range skillList {
					if !availSet[s] {
						return fmt.Errorf("skill %q not found in module %s", s, module)
					}
				}
				extracted, err = extract.ExtractFiltered(moduleDir, targetSkillsDir, skillList)
				if err != nil {
					return fmt.Errorf("extract skills: %w", err)
				}
			} else {
				extracted, err = extract.ExtractAll(moduleDir, targetSkillsDir)
				if err != nil {
					return fmt.Errorf("extract skills: %w", err)
				}
			}

			if err := selections.AddModule(skilloDir, module, extracted); err != nil {
				return fmt.Errorf("update selections: %w", err)
			}
			return nil
		},
	}
	addCmd.Flags().StringSlice("skill", nil, "Specific skill(s) to install (repeat or comma-separated)")

	// ─── remove / rm ─────────────────────────────────────────────────────────
	removeCmd := &cobra.Command{
		Use:     "remove <name|module>",
		Aliases: []string{"rm"},
		Short:   "Remove a skill or entire module",
		Long: `Remove by skill name or by entire module.
  - remove tester          removes a single skill
  - remove github.com/org/repo  removes all skills from that module
  - remove org/repo             short form (normalized)
  - remove https://github.com/org/repo  HTTPS URL`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			arg := args[0]
			skilloDir, targetSkillsDir := resolveScope()

			if normalize.LooksLikeModulePath(arg) {
				module, _, err := normalize.ModulePath(arg)
				if err != nil {
					return err
				}
				sel, err := selections.Load(skilloDir)
				if err != nil {
					return err
				}
				skillNames := selections.ModuleSkills(sel, module)

				if err := selections.RemoveModule(skilloDir, module); err != nil {
					return err
				}
				for _, name := range skillNames {
					os.RemoveAll(filepath.Join(targetSkillsDir, name))
				}
				if len(skillNames) == 0 {
					entries, _ := os.ReadDir(targetSkillsDir)
					for _, e := range entries {
						if e.IsDir() {
							os.RemoveAll(filepath.Join(targetSkillsDir, e.Name()))
						}
					}
				}
				goEdit := exec.Command("go", "mod", "edit", "-droprequire", module)
				goEdit.Dir = skilloDir
				goEdit.CombinedOutput()
				goTidy := exec.Command("go", "mod", "tidy")
				goTidy.Dir = skilloDir
				goTidy.Run()
				fmt.Printf("Removed module %s\n", module)
			} else {
				sel, err := selections.Load(skilloDir)
				if err != nil {
					return err
				}
				module := selections.FindModule(sel, arg)
				if module == "" {
					return fmt.Errorf("skill %q not found in selections", arg)
				}
				if err := selections.RemoveSkill(skilloDir, arg); err != nil {
					return err
				}
				os.RemoveAll(filepath.Join(targetSkillsDir, arg))

				sel, _ = selections.Load(skilloDir)
				if _, exists := sel[module]; !exists {
					goEdit := exec.Command("go", "mod", "edit", "-droprequire", module)
					goEdit.Dir = skilloDir
					goEdit.CombinedOutput()
					goTidy := exec.Command("go", "mod", "tidy")
					goTidy.Dir = skilloDir
					goTidy.Run()
				}
				fmt.Printf("Removed skill %s\n", arg)
			}
			return nil
		},
	}

	// ─── sync ─────────────────────────────────────────────────────────────────
	syncCmd := &cobra.Command{
		Use:   "sync",
		Short: "Materialize skills from selections.json",
		Long: `Reads selections.json and go.mod, downloads modules, and extracts
skills to the skills directory. Removes stale skills whose modules
are no longer in selections. Idempotent — safe to run repeatedly.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			skilloDir, targetSkillsDir := resolveScope()
			return syncScope(skilloDir, targetSkillsDir)
		},
	}

	// ─── list ─────────────────────────────────────────────────────────────────
	listCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List installed skills",
		RunE: func(cmd *cobra.Command, args []string) error {
			scope, _ := cmd.Flags().GetString("scope")
			showOutdated, _ := cmd.Flags().GetBool("outdated")
			format, _ := cmd.Flags().GetString("format")

			sources := skilldirs.Detect(home, cwd)

			type skillEntry struct {
				Name        string `json:"name"`
				Version     string `json:"version,omitempty"`
				Module      string `json:"module,omitempty"`
				Description string `json:"description,omitempty"`
				Location    string `json:"location,omitempty"`
				Outdated    bool   `json:"outdated"`
				Stale       bool   `json:"stale,omitempty"`
				Orphaned    bool   `json:"orphaned,omitempty"`
			}

			// readScopeSkills reads entries from a single scope.
			readScopeSkills := func(skilloDir, skillsDir, location string) []skillEntry {
				sel, _ := selections.Load(skilloDir)

				// Read existing skills on disk
				diskSkills := make(map[string]bool)
				fEntries, err := os.ReadDir(skillsDir)
				if err == nil {
					for _, e := range fEntries {
						if e.IsDir() {
							diskSkills[e.Name()] = true
						}
					}
				}

				// Build version map from go list
				skillVersion := make(map[string]string)
				skillModule := make(map[string]string)
				for module := range sel {
					goList := exec.Command("go", "list", "-m", "-f", "{{.Version}}", module)
					goList.Dir = skilloDir
					if verOut, err := goList.CombinedOutput(); err == nil {
						ver := strings.TrimSpace(string(verOut))
						for _, name := range sel[module] {
							skillVersion[name] = ver
							skillModule[name] = module
						}
					}
				}

				descriptions := readSkillDescriptions(skillsDir)
				var result []skillEntry

				for module, names := range sel {
					if len(names) > 0 {
						for _, name := range names {
							entry := skillEntry{
								Name:        name,
								Version:     skillVersion[name],
								Module:      skillModule[name],
								Description: descriptions[name],
								Location:    location,
							}
							if !diskSkills[name] {
								entry.Stale = true
							} else {
								delete(diskSkills, name)
							}
							result = append(result, entry)
						}
					} else {
						// Empty array = all skills from this module
						goList := exec.Command("go", "list", "-m", "-f", "{{.Dir}}", module)
						goList.Dir = skilloDir
						if dirOut, err := goList.CombinedOutput(); err == nil {
							modDir := strings.TrimSpace(string(dirOut))
							if available, err := extract.ListAvailableSkills(modDir); err == nil {
								for _, name := range available {
									entry := skillEntry{
										Name:        name,
										Version:     skillVersion[name],
										Module:      module,
										Description: descriptions[name],
										Location:    location,
									}
									if !diskSkills[name] {
										entry.Stale = true
									} else {
										delete(diskSkills, name)
									}
									result = append(result, entry)
								}
							}
						}
					}
				}
				// Orphaned
				for name := range diskSkills {
					result = append(result, skillEntry{
						Name:     name,
						Location: location,
						Orphaned: true,
					})
				}
				return result
			}

			showUser := scope == "user"
			showProject := scope == "project"
			showBoth := scope == "" || scope == "auto"

			var allEntries []skillEntry

			if showUser || showBoth {
				userSkillo := skilldirs.UserSkilloDir(home)
				userSkillsOut := skilldirs.UserSkillsDir(home)
				userEntries := readScopeSkills(userSkillo, userSkillsOut, "user")
				allEntries = append(allEntries, userEntries...)
			}

			if (showProject || showBoth) && sources.ProjectSkilloDir != "" {
				projEntries := readScopeSkills(sources.ProjectSkilloDir, sources.Primary, "project")
				if showBoth {
					projNames := make(map[string]bool)
					for _, e := range projEntries {
						projNames[e.Name] = true
					}
					var filtered []skillEntry
					for _, e := range allEntries {
						if !projNames[e.Name] {
							filtered = append(filtered, e)
						}
					}
					allEntries = filtered
				}
				allEntries = append(allEntries, projEntries...)
			}

			// Outdated check
			if showOutdated || cmd.Flags().Changed("outdated") {
				latestVersion := make(map[string]string)
				for _, scopeDir := range skilldirs.SkilloDirs(sources) {
					goList := exec.Command("go", "list", "-u", "-m", "-f", "{{.Path}} {{.Version}}{{if .Update}} {{.Update.Version}}{{end}}", "all")
					goList.Dir = scopeDir
					out, err := goList.CombinedOutput()
					if err == nil {
						for line := range strings.SplitSeq(string(out), "\n") {
							parts := strings.Fields(line)
							if len(parts) >= 3 {
								latestVersion[parts[0]] = parts[2]
							}
						}
					}
				}
				for i, e := range allEntries {
					if mod := e.Module; mod != "" {
						if latest, ok := latestVersion[mod]; ok && latest != e.Version {
							allEntries[i].Outdated = true
						}
					}
				}
				if showOutdated {
					var filtered []skillEntry
					for _, e := range allEntries {
						if e.Outdated {
							filtered = append(filtered, e)
						}
					}
					allEntries = filtered
				}
			}

			switch format {
			case "json":
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				if len(allEntries) == 0 {
					fmt.Println("[]")
					return nil
				}
				return enc.Encode(allEntries)
			default:
				for _, e := range allEntries {
					var buf strings.Builder
					buf.WriteString(e.Name)
					if e.Version != "" {
						buf.WriteString(fmt.Sprintf("  %s", e.Version))
					}
					if e.Outdated {
						buf.WriteString(" (outdated)")
					}
					if e.Orphaned {
						buf.WriteString(" (orphaned)")
					}
					if e.Stale {
						buf.WriteString(" (stale — run sync)")
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
	listCmd.Flags().String("scope", "", "Scope to show: user, project, or auto (default)")
	listCmd.Flags().Bool("outdated", false, "Show only skills with available upgrades")
	listCmd.Flags().String("format", "table", "Output format (table, json)")

	// ─── update ───────────────────────────────────────────────────────────────
	updateCmd := &cobra.Command{
		Use:   "update",
		Short: "Update all modules and re-extract skills",
		RunE: func(cmd *cobra.Command, args []string) error {
			sources := skilldirs.Detect(home, cwd)
			for _, skilloDir := range skilldirs.SkilloDirs(sources) {
				var outSkillsDir string
				if skilloDir == sources.ProjectSkilloDir {
					outSkillsDir = sources.Primary
				} else {
					outSkillsDir = skilldirs.UserSkillsDir(home)
				}
				if err := os.MkdirAll(skilloDir, 0o755); err != nil {
					return fmt.Errorf("mkdir %s: %w", skilloDir, err)
				}
				goGet := exec.Command("go", "get", "-u", "./...")
				goGet.Dir = skilloDir
				out, err := goGet.CombinedOutput()
				if err != nil {
					fmt.Printf("Warning: go get -u in %s: %v\n%s", skilloDir, err, out)
					continue
				}
				fmt.Println("Updated modules in", skilloDir)
				if err := syncScope(skilloDir, outSkillsDir); err != nil {
					fmt.Printf("Warning: sync in %s: %v\n", skilloDir, err)
				}
			}
			return nil
		},
	}

	rootCmd.AddCommand(initCmd, addCmd, removeCmd, syncCmd, listCmd, updateCmd)
	return rootCmd
}

func main() {
	if err := newRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}

func syncScope(skilloDir, skillsDir string) error {
	if err := os.MkdirAll(skilloDir, 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", skilloDir, err)
	}
	if err := os.MkdirAll(skillsDir, 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", skillsDir, err)
	}

	goMod := exec.Command("go", "mod", "download")
	goMod.Dir = skilloDir
	if out, err := goMod.CombinedOutput(); err != nil {
		return fmt.Errorf("go mod download: %v\n%s", err, out)
	}

	sel, err := selections.Load(skilloDir)
	if err != nil {
		return err
	}

	expected := make(map[string]bool)

	for module, names := range sel {
		goList := exec.Command("go", "list", "-m", "-f", "{{.Dir}}", module)
		goList.Dir = skilloDir
		listOut, err := goList.CombinedOutput()
		if err != nil {
			fmt.Printf("Warning: go list %s: %v\n", module, err)
			continue
		}
		moduleDir := strings.TrimSpace(string(listOut))

		var extracted []string
		if len(names) > 0 {
			extracted, err = extract.ExtractFiltered(moduleDir, skillsDir, names)
		} else {
			extracted, err = extract.ExtractAll(moduleDir, skillsDir)
		}
		if err != nil {
			fmt.Printf("Warning: extract from %s: %v\n", module, err)
		}
		for _, name := range extracted {
			expected[name] = true
		}
	}

	fEntries, err := os.ReadDir(skillsDir)
	if err == nil {
		for _, e := range fEntries {
			if e.IsDir() && !expected[e.Name()] {
				os.RemoveAll(filepath.Join(skillsDir, e.Name()))
				fmt.Printf("Removed stale skill: %s\n", e.Name())
			}
		}
	}

	fmt.Println("Sync complete")
	return nil
}

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
		if len(content) < 3 || content[:3] != "---" {
			continue
		}
		rest := content[3:]
		end := strings.Index(rest, "\n---")
		if end < 0 {
			continue
		}
		var skill types.Skill
		if err := yaml.Unmarshal([]byte(rest[:end]), &skill); err != nil {
			continue
		}
		if skill.Description != "" {
			descs[e.Name()] = skill.Description
		}
	}
	return descs
}
