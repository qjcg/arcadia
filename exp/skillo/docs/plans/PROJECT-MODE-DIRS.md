# Implementation Plan: Project-Mode Skill Directories

## 1. Goal

skillo manages agent skills in two scopes:

- **Project scope** (`<project-root>/.skillo/`) — skills pinned to a specific
  project. This directory is meant to be committed so a team shares the same
  skill versions and selections.
- **User scope** (`~/.config/skillo/`) — skills the user wants available
  globally across all projects.

Both directories are self-contained: each has its own Go module workspace
for pinning versions, and a `selections.json` that records which modules
(and optionally which subset of their skills) are registered in that scope.

The `sync` command materializes `.agents/skills/` from `.skillo/` — so a
fresh clone can be fully provisioned without committing the extracted skill
files.

## 2. Current State

| Concept                      | Path                         | Role                                                            |
|------------------------------|------------------------------|-----------------------------------------------------------------|
| Go module workspace (single) | `~/.skillo/`                 | `go.mod`, `go.sum`, downloaded modules, `.skillo-manifest.json` |
| Skills output (project)      | `<git-root>/.agents/skills/` | Extracted skill files (auto-detected)                           |
| Skills output (user)         | `~/.agents/skills/`          | Extracted skill files                                           |
| Project skillo dir           | **does not exist**           | —                                                               |
| User config dir              | **does not exist**           | —                                                               |
| `sync` command               | **does not exist**           | —                                                               |

## 3. Directory Layout

### User skillo dir (`~/.config/skillo/`)

```
~/.config/skillo/
├── config.json              # Optional: override skills_dir (§4.1)
├── go.mod                   # Go module workspace for user-installed skills
├── go.sum
└── selections.json          # Modules + skill selections (§4.2)
```

### Project skillo dir (`<project-root>/.skillo/`)

```
<project-root>/.skillo/
├── config.json              # Optional: override skills_dir
├── go.mod                   # Go module workspace for project-pinned skills
├── go.sum
└── selections.json          # Modules + skill selections (§4.2)
```

### Skills extraction destinations (unchanged)

- Project: `<git-root>/.agents/skills/<skill-name>/`
- User: `~/.agents/skills/<skill-name>/`

The agent (Crush) reads `.agents/skills/`. The `.skillo/` directory is the
*specification*; `.agents/skills/` is the *materialization*.

## 4. Data Formats

### 4.1 `config.json` (optional override)

```json
{
  "skills_dir": ".agents/skills"
}
```

Relative paths are resolved against the parent of `.skillo/` (e.g., for
project `<project-root>/.skillo/`, value `.agents/skills` →
`<project-root>/.agents/skills/`). Absent → default `.agents/skills`.

Only needed when the extraction target is something other than the default.

### 4.2 `selections.json` (the one intentional state file)

```json
{
  "github.com/user/agent-browser": ["agent-browser"],
  "github.com/org/skill-collection": ["cli-design", "tester"]
}
```

Keys = Go module paths. Values = which skills from that module to extract:
- **non-empty array** → only those skills
- **empty array** (`[]`) → all skills from that module

Module absent from selections → not managed in this scope.

This is the *only* state that cannot be derived. Everything else fans out
from it:

| Fact                           | Source                                          |
|--------------------------------|-------------------------------------------------|
| Module version                 | `go list -m -f {{.Version}} <module>`           |
| Module dir on disk             | `go list -m -f {{.Dir}} <module>`               |
| Available skills in a module   | Walk module dir for SKILL.md subdirs            |
| Whether a skill is extracted   | `os.Stat(<skills-dir>/<name>/SKILL.md)`         |
| Whether a skill has an upgrade | `go list -u -m -f {{.Update.Version}} <module>` |

No `manifest.json`, no `versions.json`, no `.skillo-manifest.json`. All of
those were either derivable caches or redundant with `go.mod`/`go.sum`.

## 5. Scope Resolution

When resolving which skills are visible from `list`:

1. Start with **user scope** (`~/.config/skillo/`):
   - Modules in `selections.json` → resolved via Go module cache
   - Skills present on disk but not in selections → shown with `(orphaned)` indicator
   - Labeled `(user)` in merged output

2. Overlay **project scope** (`<project-root>/.skillo/`) if in a repo:
   - Same module in both scopes: project version wins, no `(user)` label
   - Skill in project scope only: shown as primary, no label
   - Skill in user scope only: shown with `(user)` label
   - Skill absent from both scopes but present on disk → shown as orphaned

3. `--scope user` or `--scope project` filters to a single scope's view.

## 6. Module Path Normalization

All commands that accept a module path (`add`, `remove` when targeting a
module, `sync` — indirectly) normalize the input into a full Go module path
before use.

### Input formats

| Format         | Example                        | Resolution                                                |
|----------------|--------------------------------|-----------------------------------------------------------|
| Go import path | `github.com/user/repo@v1.2.3`  | Passed through as-is. `@latest` assumed if no `@` suffix. |
| HTTPS URL      | `https://github.com/user/repo` | Strip `https://` prefix → `github.com/user/repo`          |
| Short form     | `user/repo` or `org/repo`      | Prepend `github.com/` → `github.com/user/repo`            |

### Normalization rules

```
https://github.com/user/repo      →  github.com/user/repo
https://github.com/user/repo@v1   →  github.com/user/repo@v1
user/repo                          →  github.com/user/repo
user/repo@v1.2.3                   →  github.com/user/repo@v1.2.3
github.com/user/repo               →  github.com/user/repo (passthrough)
github.com/user/repo@v1.2.3        →  github.com/user/repo@v1.2.3 (passthrough)
```

### Heuristic

1. If the argument starts with `https://` or `http://` — strip the scheme.
   The remainder is treated as a Go import path.
2. If the argument contains a `.` — assume it's already a full Go import path.
   Pass through unchanged.
3. Otherwise (short form with no dot, e.g. `user/repo`) — prepend
   `github.com/`.
4. Extract and validate version suffix (`@<spec>`). If absent, append `@latest`.

### `normalizeModule` helper

```go
func normalizeModule(input string) (module, version string, err error)
```

Pure function, no I/O. Used by `add`, `remove` (when arg looks like module),
and anywhere else a module path is accepted.

### Commands that use normalization

| Command           | When normalization applies                                                                                                             |
|-------------------|----------------------------------------------------------------------------------------------------------------------------------------|
| `add <repo>`      | Always — normalizes the repo argument before `go get`                                                                                  |
| `remove <module>` | When the argument looks like a module path (contains `.` or `/`). If it looks like a skill name, it's treated as a skill name instead. |

## 7. Commands

### `init [--project]`

```
skillo init              # Create ~/.config/skillo/ with go.mod + empty selections.json
skillo init --project    # Create <project-root>/.skillo/ with go.mod + empty selections.json
```

Migration: if legacy `~/.skillo/` exists and `~/.config/skillo/` doesn't,
rename `~/.skillo/` → `~/.config/skillo/`. If the legacy dir contains
`.skillo-manifest.json`, convert its entries into `selections.json` format.

### `add <repo>[@version] [--skill name...] [--user | --project]`

Register a module and select which of its skills to install.

```
skillo add user/agent-browser@v1.2.3
skillo add github.com/org/skill-collection@v0.5.0 --skill cli-design --skill tester
skillo add https://github.com/org/skill-collection --skill cli-design,tester
skillo add org/skill-collection --user
```

1. Resolve target scope (auto-detect, `--user`, or `--project`)
2. Normalize the module path via `normalizeModule` — strips `https://`,
   prepends `github.com/` for short form, defaults to `@latest`
3. `go get <module>@<version>` in that scope's `.skillo/` dir
4. Find the module dir: `go list -m -f {{.Dir}} <module>`
5. Walk the module dir, find every directory containing SKILL.md, collect skill names
6. If `--skill` was provided:
   - Validate each named skill actually exists in the module (error if not)
   - Only extract those skills
   - Write only those skills to selections.json
7. If no `--skill`: extract all skills, write all to selections.json
8. Append/update `selections.json`: `"<module>": [list of extracted skill names]`

The `--skill` flag accepts multiple values (repeated flag or comma-separated):

```
--skill foo --skill bar
--skill foo,bar
```

### `remove <name|module> [--user | --project]` / `rm <name|module>`

Remove by skill name or by entire module.

```
skillo remove tester                           # Remove single skill
skillo remove github.com/org/skill-collection  # Remove entire module
skillo remove org/skill-collection             # Short form — normalized first
skillo rm tester                               # Alias
```

**Auto-detection**: if the argument contains `.` or `/`, treat as a module
path (normalize it first via `normalizeModule`, then look up in selections).
Otherwise, treat as a skill name (scan selections to find owning module).

#### Single skill (no dots or slashes):

1. Scan `selections.json` to find which module owns `<name>`
2. Remove `<name>` from that module's array
3. Delete skill dir from skills dir
4. If the module's array is now empty:
   - `go mod edit -droprequire <module>`
   - `go mod tidy`
   - Remove the module entry from `selections.json`

#### Entire module (contains `.` or `/`):

1. Normalize the input via `normalizeModule` (strips scheme, prepends
   `github.com/` for short form, drops version suffix)
2. Look up the normalized module path in `selections.json`
3. Remove the module entry from `selections.json`
4. Delete all skill dirs associated with that module
5. `go mod edit -droprequire <module>`
6. `go mod tidy`

### `sync [--user | --project]`

Idempotent: materialize `.agents/skills/` from `.skillo/` state.

1. Resolve scope
2. `go mod download` — ensure all modules in `go.mod` are cached
3. For each module in `selections.json`:
   - `go list -m -f {{.Dir}} <module>` to find local path
   - Walk the module dir, find all SKILL.md subdirs
   - Filter by the skill list in selections.json (if non-empty)
   - Extract matching skills to scope's skills dir
4. Remove any skill dirs from skills dir whose owning module is no longer
   in `selections.json` (cleanup stale extractions)

This is the command you run after `git clone` to provision the environment.
`.skillo/` is committed (selections.json + go.mod/go.sum); `.agents/skills/`
is gitignored and reproduced by `sync`.

### `list [--scope user | project] [--filter name]`

```
skillo list                         # Merged view of both scopes
skillo list --scope user            # User scope only
skillo list --scope project         # Project scope only
```

For each scope: read `selections.json`, resolve module dirs and versions from
the Go module cache, cross-reference with skills dir on disk.

Output format:

```
agent-browser          v1.2.3         github.com/user/agent-browser
cli-design             v0.5.0         github.com/org/skill-collection
tester                 v0.5.0         github.com/org/skill-collection   (user)
old-skill              —              —                                 (orphaned)
```

- Primary (project-scoped) skills have no label
- User-only skills labeled `(user)`
- Skills on disk but not in any selections → `(orphaned)`
- Skills in selections but missing from disk → `(stale — run sync)`
- Outdated modules → `(update available: v1.3.0)`

### `update`

- `go get -u ./...` in each scope's `.skillo/`
- Then `sync` each scope to re-extract with latest versions

## 8. Packages

### `internal/normalize/` — Module path normalization

```go
// ModulePath converts user input into a full Go module path and version.
//   "user/repo"           → ("github.com/user/repo", "latest")
//   "user/repo@v1.2.3"    → ("github.com/user/repo", "v1.2.3")
//   "https://github.com/user/repo" → ("github.com/user/repo", "latest")
//   "github.com/user/repo" → ("github.com/user/repo", "latest")
func ModulePath(input string) (module, version string, err error)

// LooksLikeModulePath returns true if the string contains '.' or '/'.
// Used by remove to distinguish skill names from module paths.
func LooksLikeModulePath(s string) bool
```

### `internal/selections/` — The one state file

```go
// Selections maps module path → list of skill names.
// nil/empty slice means "all skills from this module".
// Non-empty slice means "only these skills from this module".
type Selections map[string][]string
```

| Export         | Signature                                          | Purpose                                |
|----------------|----------------------------------------------------|----------------------------------------|
| `Load`         | `(dir string) (Selections, error)`                 | Reads `<dir>/selections.json`          |
| `Save`         | `(dir string, s Selections) error`                 | Writes `<dir>/selections.json`         |
| `Init`         | `(dir string) error`                               | Creates empty `{}` selections.json     |
| `AddModule`    | `(dir, module string, skills []string) error`      | Register module + skills               |
| `RemoveSkill`  | `(dir, name string) error`                         | Remove skill from its module's array   |
| `RemoveModule` | `(dir, module string) error`                       | Remove entire module entry             |
| `FindModule`   | `(s Selections, skillName string) (string, error)` | Scan map for which module owns a skill |
| `ModuleSkills` | `(s Selections, module string) []string`           | Get skill list for a module            |

### `internal/skilldirs/` — Updated

```go
type Sources struct {
    Primary          string  // primary skills dir (project or user)
    Secondary        string  // user skills dir
    HomeDir          string
    UserSkilloDir    string  // ~/.config/skillo/
    ProjectSkilloDir string  // <git-root>/.skillo/ (empty if not in repo)
}
```

| Export                                 | Purpose                                                    |
|----------------------------------------|------------------------------------------------------------|
| `UserSkilloDir(home string) string`    | Returns `~/.config/skillo/`                                |
| `ProjectSkilloDir(root string) string` | Returns `<root>/.skillo/`                                  |
| `Detect(home, cwd string) *Sources`    | Populates all fields including new ones                    |
| `SkilloDirs(s *Sources) []string`      | Returns `[ProjectSkilloDir, UserSkilloDir]` ignoring empty |

### `cmd/skillo/` — Changes

**New commands**: `add`, `sync`, `remove`/`rm`

**Updated commands**: `init`, `list`, `update`

**Removed**: `get` (replaced entirely by `add`), old `remove <module>` (replaced
by unified `remove`), `--modules-dir` flag, `.skillo-manifest.json` dependency.

**Key helpers**:

```go
func resolveScope(cmd *cobra.Command, sources *skilldirs.Sources) (skilloDir, skillsDir string, err error)
func syncScope(skilloDir string) error
func parseSkillFlag(cmd *cobra.Command) []string
```

## 9. Migration

### `~/.skillo/` → `~/.config/skillo/`

On first `init` or any command when `~/.skillo/` exists but
`~/.config/skillo/` doesn't:

1. Rename `~/.skillo/` → `~/.config/skillo/`
2. If `~/.config/skillo/.skillo-manifest.json` exists:
   - Read it
   - Convert `module_skills` entries into `selections.json` format
   - Write `selections.json`
   - Remove `.skillo-manifest.json`
3. Create empty `selections.json` if neither manifest nor selections exist
4. Create default `config.json`

### Old commands removed

- `get` — gone. Use `add`.
- Old `remove <module>` — gone. Unified `remove` handles both skill names
  and module paths (auto-detected + normalized).

## 10. Testing Plan

### Testscript (`cmd/skillo/testdata/`)

| File                           | What it tests                                                                             |
|--------------------------------|-------------------------------------------------------------------------------------------|
| `init_user.txtar`              | `init` creates `~/.config/skillo/` with go.mod, empty selections.json                     |
| `init_project.txtar`           | `init --project` creates `<root>/.skillo/` in a git repo                                  |
| `add.txtar`                    | `add repo@v` → go get, scan SKILL.md, extract all skills, write selections                |
| `add_skill_flag_repeat.txtar`  | `add repo@v --skill foo --skill bar` extracts only foo and bar                            |
| `add_skill_flag_csv.txtar`     | `add repo@v --skill foo,bar` same result comma-separated                                  |
| `add_short_form.txtar`         | `add user/repo` auto-prepends `github.com/`                                               |
| `add_https_url.txtar`          | `add https://github.com/user/repo` strips scheme                                          |
| `add_skill_flag_invalid.txtar` | `add repo@v --skill nonexistent` errors                                                   |
| `add_user.txtar`               | `add --user` targets user scope explicitly                                                |
| `remove_skill.txtar`           | `remove name` → find module, droprequire, delete, update selections                       |
| `remove_module_go_path.txtar`  | `remove github.com/org/skill-collection` removes entire module                            |
| `remove_module_short.txtar`    | `remove org/skill-collection` normalizes short form first                                 |
| `remove_module_https.txtar`    | `remove https://github.com/org/skill-collection` normalizes URL first                     |
| `rm_alias.txtar`               | `rm name` works as alias for `remove`                                                     |
| `sync.txtar`                   | `sync` downloads modules, extracts skills, prunes stale                                   |
| `sync_fresh_clone.txtar`       | Clone scenario: committed selections.json + go.mod, empty cache, `sync` provisions skills |
| `list_scope.txtar`             | `list --scope user` / `list --scope project`                                              |
| `list_merge.txtar`             | Default `list` merges scopes, shows (user) and (orphaned) labels                          |
| `list_indicators.txtar`        | Shows (stale), (orphaned), (update available) indicators                                  |
| `migrate_legacy.txtar`         | Legacy `~/.skillo/` with manifest → migrated to `~/.config/skillo/` on init               |

### Feature tests (`features/*.feature`)

- `add.feature` — add command scenarios, `--skill` flag variants, short form, HTTPS URL
- `remove.feature` — remove by skill name, remove by module path (full, short, HTTPS), rm alias
- `sync.feature` — sync command scenarios
- `list.feature` — add `--scope` scenarios, orphaned/stale indicators
- `init.feature` — add `--project` scenario, migration scenario

### Unit tests

- `internal/normalize/normalize_test.go` — all input format permutations,
  edge cases (no version, `@latest`, `@v1.2.3`, scheme stripping, short form),
  `LooksLikeModulePath` true/false cases
- `internal/selections/selections_test.go` — load/save, add/remove module,
  find module by skill name, empty states
- `internal/skilldirs/skilldirs_test.go` — path resolution with new fields,
  git root detection, fallbacks, SkilloDirs ordering

## 11. `.gitignore` conventions

Recommended patterns (documented in README, not enforced by skillo):

```
# Project skillo dir is committed
.skillo/

# Extracted skill output is derived
.agents/skills/
```

## 12. Implementation Order

| #  | Area                   | Description                                                                    | Est. |
|----|------------------------|--------------------------------------------------------------------------------|------|
| 1  | `internal/normalize/`  | New pkg: `ModulePath()` + `LooksLikeModulePath()`                              | 10m  |
| 2  | `internal/selections/` | New pkg: `Selections` type, load/save, CRUD operations                         | 20m  |
| 3  | `internal/skilldirs/`  | Update: add new path fields to `Sources`, update `Detect()`                    | 10m  |
| 4  | `cmd/skillo/`          | Update `init`: user vs `--project`, legacy migration                           | 15m  |
| 5  | `cmd/skillo/`          | New `add` command: normalize + go get + extract + `--skill` + selections       | 25m  |
| 6  | `cmd/skillo/`          | New `remove`/`rm`: auto-detect skill vs module, normalize module path, cleanup | 20m  |
| 7  | `cmd/skillo/`          | New `sync` command: download + iterate + extract + prune                       | 20m  |
| 8  | `cmd/skillo/`          | Update `list`: selections, versions, filesystem, `--scope` flag                | 15m  |
| 9  | `cmd/skillo/`          | Update `update`: scoped go get -u + sync                                       | 10m  |
| 10 | `cmd/skillo/`          | Remove `get`, old `remove <module>`, old state file deps                       | 10m  |
| 11 | Tests                  | testscript + feature + unit tests                                              | 40m  |
| 12 | docs                   | Update README, SKILL.md, PRD.md                                                | 10m  |

**Total estimate**: ~205 min (agent time)

