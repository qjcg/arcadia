---
name: skillo
description: Go module-based agent skills manager with project and user scopes.
---

# Skillo CLI

## When to use
Manage agent skills (SKILL.md-based) with `go get`-style versioning, across
both user-global and project-scoped directories.

## Directory layout

**User scope** (`~/.config/skillo/`):
```
~/.config/skillo/
├── config.json           # Optional: skills_dir override
├── go.mod / go.sum       # Go module workspace
└── selections.json       # Module + skill selections
```

**Project scope** (`<project-root>/.skillo/`):
```
<project-root>/.skillo/
├── config.json           # Optional: skills_dir override
├── go.mod / go.sum       # Go module workspace
└── selections.json       # Module + skill selections
```

**Skills are extracted to** `.agents/skills/` (project) or `~/.agents/skills/` (user).

## Commands

- `init [--project]` — Initialize user or project skillo workspace
- `add <repo>[@v] [--skill name...] [--user | --project]` — Register module + install skills
- `remove <name|module>` / `rm` — Remove skill or entire module
- `sync [--user | --project]` — Materialize skills from selections.json
- `list [--scope user|project] [--outdated] [--format json]` — List installed skills
- `update` — Update all modules and re-extract

## Module path formats

All commands accept three input formats:
- Full Go import path: `github.com/user/repo`
- Short form: `user/repo` (`github.com/` is prepended)
- HTTPS URL: `https://github.com/user/repo`
- Version via `@` suffix: `repo@v1.2.3` (default: `@latest`)
