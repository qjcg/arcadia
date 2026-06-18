# Changelog Feature Design

## 1. Motivation

`sv` already calculates next versions from conventional commits and manages path-based tags for monorepo modules. Adding a `sv changelog` command closes the loop: a team can run `sv next` to see what version is coming, then `sv changelog` to generate a human-readable release notes document — without leaving the CLI or reaching for a separate tool.

Output follows [keepachangelog.com v1.1.0](https://keepachangelog.com/en/1.1.0/): reverse chronological releases, categorised entries under Added / Changed / Deprecated / Removed / Fixed / Security.

---

## 2. User Stories

### U1 — Per-module changelog (default)
> As a developer inside `x/slidesdeck/`, I run `sv changelog` and get a changelog for that module, showing all releases from the first tag to `HEAD`.

### U2 — Unreleased changes only
> I run `sv changelog --unreleased` to see what will go into the *next* release, useful for PR review or release prep.

### U3 — All modules
> As a monorepo maintainer, I run `sv changelog --all` to get changelogs for every module, separated by module heading.

### U4 — Version range
> I run `sv changelog --since v1.0.0 --to v1.5.0` to see only changes between two specific tags.

### U5 — Write to file
> I run `sv changelog --output CHANGELOG.md` to write directly into my module's standard changelog file.

---

## 3. CLI Interface

### `sv changelog`

```
Generate a keepachangelog-style changelog for the current module

Usage:
  sv changelog [flags]

Flags:
      --all              Generate changelog for all modules
  -p, --path strings     Explicit module path(s)
      --since string     Start tag (default: first tag for the module)
      --to string        End tag (default: HEAD)
      --unreleased       Only show changes since the latest tag
  -o, --output string    Write to file instead of stdout (use "-" for explicit stdout)
      --no-version-links Skip linking to tag refs in git
  -h, --help             help for changelog
```

### Integration with existing flags

| Flag | Behaviour | Inherited from |
|---|---|---|
| `--all` | All discovered modules | `next`, `current`, `bump` |
| `--path` / `-p` | Specific module path(s) | `next`, `current`, `bump` |
| `--verbose` | Print warnings (e.g. skipped retracted tags) | `root` persistent flag |

---

## 4. Architecture

### 4.1 New internal packages

```
internal/
  changelog/
    changelog.go    — Core generation logic
    format.go       — Markdown formatting
    entry.go        — Entry, Section, Release types
    conventional.go — Commit message parsing (type, scope, breaking)
```

### 4.2 Types

```go
// ———————————————— entry.go ————————————————

// Conventional commit breakdown.
type CommitType int

const (
    CommitUnknown CommitType = iota
    CommitFeat
    CommitFix
    CommitDeprecated
    CommitRemoved
    CommitSecurity
    CommitPerf
    CommitRefactor
    CommitStyle
    CommitTest
    CommitChore
    CommitCI
    CommitBuild
    CommitDocs
)

// A single changelog entry derived from a commit.
type Entry struct {
    Hash     string     // abbreviated commit hash (7 chars)
    Message  string     // commit subject (after stripping type/scope)
    Type     CommitType
    Scope    string     // optional scope from conventional commit, e.g. "api"
    Breaking bool       // marked with `!` or `BREAKING CHANGE:` trailer
    Time     time.Time  // author time
}

// A section within a release (e.g. "Added", "Fixed").
type Section struct {
    Title   string  // "Added", "Changed", "Deprecated", "Removed", "Fixed", "Security"
    Icon    string  // e.g. "✨ Added" — see §5.4 for customisation
    Entries []Entry
}

// A single release (version).
type Release struct {
    Version  string    // e.g. "v1.2.3" or "x/mod/v1.2.3"
    Tag      string    // full git tag name
    Date     string    // date of the tag commit, or today for unreleased
    IsRecent bool      // true if this is the latest (HEAD) unreleased block
    Sections []Section
}

// ———————————————— changelog.go ————————————————

// Changelog holds the full changelog for a single module.
type Changelog struct {
    ModulePath string    // "." or "x/mod"
    ModuleName string    // module path from go.mod, for header display
    Releases   []Release
}

// Options configure generation.
type Options struct {
    Since       string   // start tag (empty = first)
    To          string   // end tag (empty = HEAD)
    Unreleased  bool     // only show unreleased (commits since latest tag)
    NoTagLinks  bool     // skip linking to tag refs
}
```

### 4.3 Conventional commit parser (`conventional.go`)

The parser handles the [Conventional Commits 1.0.0](https://www.conventionalcommits.org/en/v1.0.0/) format:

```
type(scope)!: description

BREAKING CHANGE: <body>
```

Parsing rules:

| Pattern                         | Type    | Scope    | Breaking                 |
|---------------------------------|---------|----------|--------------------------|
| `feat: add widget`              | Feat    | ""       | false                    |
| `feat(api): add endpoint`       | Feat    | "api"    | false                    |
| `feat!: breaking thing`         | Feat    | ""       | true                     |
| `feat(api)!: breaking`          | Feat    | "api"    | true                     |
| `fix: crash`                    | Fix     | ""       | false                    |
| `fix(parser)!: BREAKING CHANGE` | Fix     | "parser" | true                     |
| `BREAKING CHANGE: big rewrite`  | Unknown | ""       | true (any type)          |
| `docs: update readme`           | Docs    | ""       | false                    |
| Custom (no recognised type)     | Unknown | ""       | false — falls to Changed |

The parser reads the first line of the commit message only (for the subject). Breaking changes detected via `!` before `:` OR `BREAKING CHANGE:` / `BREAKING-CHANGE:` in the body.

### 4.4 Category mapping

| Parsed Type                                         | Section    | Condition                                   |
|-----------------------------------------------------|------------|---------------------------------------------|
| Feat                                                | Added      | non-breaking                                |
| Feat                                                | Changed    | breaking                                    |
| Fix                                                 | Fixed      | non-breaking                                |
| Fix                                                 | Changed    | breaking                                    |
| Deprecated                                          | Deprecated | any                                         |
| Removed                                             | Removed    | any                                         |
| Security                                            | Security   | any                                         |
| Perf, Refactor, Style, Test, Chore, CI, Build, Docs | Changed    | any                                         |
| Unknown                                             | Changed    | any                                         |
| *any*                                               | Changed    | if Breaking=true (overrides normal section) |

### 4.5 Entry deduplication

When `--all` is used, the same commit may appear in multiple modules' changelogs if it touches multiple modules. This is correct behaviour — each module's changelog reflects the changes scoped to that module, exactly as `sv next` does today.

---

## 5. Output Format

### 5.1 Full module changelog (`sv changelog`)

```markdown
# Changelog — x/slidesdeck

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- GraphQL subscription support by @alice ([abc1234](https://...))
- Rate limiting for public endpoints ([def5678](https://...))

### Fixed

- Panic on nil pointer in session handler ([ghi9012](https://...))

## [v2.3.1] - 2025-06-10

### Changed

- **BREAKING**: Renamed `CreateUser` to `RegisterUser` ([jkl3456](https://...))

### Deprecated

- Legacy OAuth flow will be removed in v3 ([mno7890](https://...))
```

### 5.2 Unreleased only (`sv changelog --unreleased`)

```markdown
## [Unreleased]

### Added

- GraphQL subscription support

### Fixed

- Panic on nil pointer in session handler
```

Only the `## [Unreleased]` block, no header boilerplate, no tag links — optimal for PR templates or CI previews.

### 5.3 File mode (`sv changelog -o CHANGELOG.md`)

When writing to an **existing** file:

1. **If the file already contains a `## [Unreleased]` block** — the command replaces that block's content in-place, preserving all prior release entries. This is the key workflow for CI: each merge updates the unreleased section, and before a release tag you manually move entries from `[Unreleased]` to the new version heading.

2. **Otherwise** — prepend the full output to the file. This requires the file to start with a blank line or comment that `sv` can anchor to. If the file doesn't exist, write fresh.

> **Note**: Full in-place mutation of the Unreleased block is a v2 enhancement. v1 starts with stdout-only output and `--output` writes a fresh file only.

### 5.4 Entry formatting

- Each entry starts with a capital letter (the commit message is normalised).
- Breaking entries are prefixed with **`BREAKING`** in bold.
- Hash is included as a short 7-char ref when available.
- When `--no-version-links` is not set and git remote exists, hashes link to the remote commit URL.

---

## 6. Data Flow

```
git tags --list <pattern>
        │
        ▼
  Sort tags (version descending) ───► pair consecutive tags
        │                                   │
        ▼                                   ▼
  For each pair (tagA, tagB):        git log --format="%H%n%aI%n%s%n%b"
  get commits between them                │
        │                                 ▼
        ▼                          Parse each commit:
  Categorise entries              conventional commit → Entry
        │
        ▼
  Group entries → Sections → Release
        │
        ▼
  Template → markdown string ──► stdout / file
```

### Detailed step-by-step for a single module

1. **Get all tags**: `git tag --list <module-prefix>/v* --sort=-v:refname`
2. **Apply filtering**:
   - If `--since` provided, skip tags older than that.
   - If `--to` provided, stop at that tag.
   - If `--unreleased`, only consider the latest tag → HEAD.
3. **Pair tags**: Walk tags in ascending order, producing pairs `(prev, next)`. The first pair starts from the first tag (no prev). If there's no tag yet, use `HEAD` → first commit as a single unreleased block.
4. **Collect commits** between each pair: `git log <prev>..<next>` (or `git log <first-tag>` for the first release, or `git log <latest>..HEAD` for unreleased).
5. **Filter by module path**: Use the existing `CommitsSince` with `excludePaths` for the root module, exactly as `sv next` does.
6. **Parse each commit** into an `Entry`.
7. **Group entries** into sections by type.
8. **Sort sections** in standard keepachangelog order: Added → Changed → Deprecated → Removed → Fixed → Security.
9. **Sort entries** within each section chronologically (oldest first per release block).
10. **Render** via `text/template`.

### Remote URL detection for hash links

1. Run `git remote get-url origin`
2. Parse common host patterns:
   - `github.com/owner/repo` → `https://github.com/owner/repo/commit/<hash>`
   - `gitlab.com/owner/repo` → `https://gitlab.com/owner/repo/-/commit/<hash>`
   - `bitbucket.org/owner/repo` → `https://bitbucket.org/owner/repo/commits/<hash>`
3. If unknown or `--no-version-links`, omit the link.

---

## 7. Key Design Decisions

### D1 — Why a new `changelog` subcommand instead of output flag on `next`?

`sv next` is a single-line output tool used in scripting (`sv next | while read tag`). Adding multi-line markdown there would break pipelines. A dedicated `changelog` command keeps concerns cleanly separated.

### D2 — Why not integrate with GitHub/GitLab releases API?

The tool should work offline and with any git host. Remote links are a nicety, not a requirement. The output can be pasted into a release UI by the user.

### D3 — Why replace the Unreleased block in-place for `-o` mode?

If `sv changelog -o CHANGELOG.md` always prepended a full file, prior release notes would accumulate duplicates every run. In-place replacement of only the `## [Unreleased]` section lets CI safely re-generate the unreleased block on every merge without duplicating history. The user manages the final move to a versioned heading as part of their release process.

### D4 — Tag links vs no links

Tag links are opt-out (`--no-version-links`) so the common case (running locally, sharing a markdown snippet) includes navigable links. CI that generates changelogs for internal consumption can disable them.

---

## 8. Test Plan

### 8.1 Unit tests (`internal/changelog/`)

| Test | What it covers |
|---|---|
| `TestParseFeat` | `feat: x` → Feat, "", false |
| `TestParseBreaking` | `feat!: x` → Feat, "", true |
| `TestParseScope` | `feat(api): x` → Feat, "api", false |
| `TestParseBreakingScope` | `feat(api)!: x` → Feat, "api", true |
| `TestParseBodyBreaking` | body containing `BREAKING CHANGE: x` → Breaking=true |
| `TestParseUnknown` | `random text` → Unknown, Changed section |
| `TestMapFeatToAdded` | Feat → "Added" |
| `TestMapFixToFixed` | Fix → "Fixed" |
| `TestMapBreakingFeatToChanged` | Feat+Breaking → "Changed" (with BREAKING prefix) |
| `TestCategoriseEntries` | Mix of types → correct sections |
| `TestReleaseSort` | Releases sorted descending version |
| `TestSectionSort` | Sections in canonical order |
| `TestRenderEmpty` | No entries → minimal valid markdown |
| `TestRenderSingleRelease` | One release with one entry per section → valid output |
| `TestRenderMultiRelease` | Multiple releases |
| `TestRenderUnreleasedOnly` | `--unreleased` flag output |
| `TestDedupEntries` | Deduplication within a section |

### 8.2 Integration tests (`testdata/*.txtar`)

| Test file | Scenario |
|---|---|
| `changelog_no_tags.txtar` | Module with no tags → empty output |
| `changelog_simple.txtar` | Single module, feat + fix commits → two releases, Added + Fixed sections |
| `changelog_breaking.txtar` | Breaking commit → Changed section with BREAKING prefix |
| `changelog_unreleased.txtar` | `--unreleased` flag → only latest block |
| `changelog_all.txtar` | `--all` → multiple modules |
| `changelog_range.txtar` | `--since v1.0.0 --to v2.0.0` → specific version range |
| `changelog_modules.txtar` | Module with path prefix (x/mod) → correct headings + tags |

### 8.3 Existing patterns to follow

Tests use `testscript` (txtar files with `exec`/`stdout` directives). The test file `main_test.go` registers the `sv` command. All new test files go in `testdata/`. Example pattern:

```
# testdata/changelog_simple.txtar

exec git init
exec git config user.email test@example.com
exec git config user.name "Test User"
exec git commit --allow-empty -m "initial"
exec git tag v1.0.0
exec git commit --allow-empty -m "feat: add widget"
exec git commit --allow-empty -m "fix: fix crash"
exec git tag v1.1.0

exec sv changelog
stdout '## \[v1\.1\.0\]'
stdout '### Added'
stdout '### Fixed'
! stdout 'BREAKING'
```

---

## 9. Implementation Order

1. **`internal/changelog/conventional.go`** — Commit message parser
2. **`internal/changelog/entry.go`** — Types (Entry, Section, Release)
3. **`internal/changelog/changelog.go`** — Core generation logic (collect tags, pair, fetch commits, categorise)
4. **`internal/changelog/format.go`** — Markdown rendering via `text/template`
5. **`internal/cli/changelog.go`** — New `changelog` subcommand, wired into `root.go`
6. **Integration tests** — txtar test scripts
7. **Documentation** — Update README with `changelog` section

---

## 10. Open Questions (for future iterations)

1. **Should `sv changelog` respect retracted tags?** — Yes, same as `sv next`. The `latestNonRetractedTag` helper already exists. Retracted versions should be skipped in the tag pairing, with a note in the output if a version was retracted.

2. **Should entries include the author?** — Keep a Changelog doesn't mandate it, but teams may want `@mention` attribution. Making this opt-in via `--with-author` is a natural v2 enhancement.

3. **Should `sv changelog` update CHANGELOG.md in-place (replacing Unreleased block)?** — Documented in §5.3 as aspirational. v1 writes to stdout or overwrites/creates the output file. The in-place replace is a v2 enhancement.

4. **Cross-module commit deduplication in `--all` output?** — As argued in §4.5, the same commit may legitimately belong to multiple modules' changelogs if it touches both. No dedup for v1.
