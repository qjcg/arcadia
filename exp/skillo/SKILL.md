---
name: skillo
description: Go CLI for installing agent skills from Git repos using Go modules for versioning.
---

# Skillo CLI

## When to use
Manage agent skills (SKILL.md-based) with `go get`-style versioning.

## Commands
- `init`: Setup workspace
- `get <repo[@v]>`: Install/extract
- `search <term>`: GitHub + local
- `list [--outdated]`
- `update`
- `remove <name>`
- `validate <dir>`

Skills in `~/.config/agents/skills/<name>/`.
