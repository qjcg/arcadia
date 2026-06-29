# skillo — Go modules for agent skills

**skillo** is a CLI tool that uses Go modules to version, install, and manage
AI agent skills (SKILL.md-based). It provides two scopes:
- **User scope** (`~/.config/skillo/`) — skills available globally
- **Project scope** (`<project-root>/.skillo/`) — skills pinned to a specific project

Both scopes are self-contained Go module workspaces with a single state file
(`selections.json`) that records which modules are registered and which of
their skills should be extracted.

## Quickstart

```bash
# Install
go install github.com/qjcg/arcadia/exp/skillo/cmd/skillo@latest

# Initialize user workspace
skillo init

# Add a skill module
skillo add github.com/user/agent-browser@v1.2.3

# List installed skills
skillo list

# Sync (materialize skills from selections)
skillo sync

# Or work in a project
cd my-project
skillo init --project
skillo add org/skill-collection --skill cli-design --skill tester
```

## Motivation

Agent skills are directories containing a `SKILL.md` file. The agent (Crush,
Claude Code, etc.) picks them up from specific directories. skillo manages
which modules provide which skills and pins their versions using Go modules.

The project `.skillo/` directory is designed to be **committed** — it contains
`selections.json` and `go.mod`/`go.sum`, which together fully specify the
skill set for a project. The extracted `.agents/skills/` directory is derived
and can be gitignored. After cloning, run `skillo sync` to provision it.

## Scope resolution

When listing skills, the project scope overlays the user scope:
- Skills in project `selections.json` → shown as primary (no label)
- Skills in user `selections.json` → shown with `(user)`
- Skills on disk but not in any selections → shown as `(orphaned)`
- Skills in selections but missing on disk → shown as `(stale — run sync)`

## Module path formats

All commands that accept a module path support three input formats:

| Format | Example | Resolves to |
|--------|---------|-------------|
| Go import path | `github.com/user/repo@v1` | `github.com/user/repo@v1` |
| Short form | `user/repo` | `github.com/user/repo@latest` |
| HTTPS URL | `https://github.com/user/repo` | `github.com/user/repo@latest` |

## Commands

| Command | Description |
|---------|-------------|
| `init [--project]` | Initialize a skillo workspace |
| `add <repo>[@v] [--skill name...]` | Register a module and extract its skills |
| `remove <name\|module>` / `rm` | Remove a skill or entire module |
| `sync` | Materialize skills from selections.json |
| `list [--scope user\|project]` | List installed skills |
| `update` | Update all modules and re-extract |

## selections.json

The single state file. Maps module paths to skill name lists:

```json
{
  "github.com/user/agent-browser": ["agent-browser"],
  "github.com/org/skill-collection": ["cli-design", "tester"]
}
```

- Non-empty array → only those skills are extracted
- Empty array (`[]`) → all skills from the module are extracted

## Project workflow

```bash
# In your project directory (already a git repo)
skillo init --project
skillo add github.com/user/agent-browser@v1.2.3
skillo add org/skill-collection --skill cli-design

# Commit the specification
git add .skillo/
git commit -m "Add skill dependencies"

# On another clone:
git clone ...
skillo sync     # Provisions .agents/skills/ from .skillo/
```
