---
name: cli-design
description: Design, review, and improve CLI interfaces following clig.dev guidelines. Use when building CLI tools, commands, flags, or help text — especially with Cobra/Viper in Go. Triggers on requests to create CLI tools, add commands, design flag interfaces, write help text, or improve CLI user experience. Make sure to load reference files from references/ when designing specific areas like output, errors, or configuration.
---

# CLI Design

Apply clig.dev principles to create command-line interfaces that are intuitive, consistent, and user-friendly.

## Core Principles

### 1. Design for Humans First
CLIs are used by humans, not just scripted. Prioritize clarity and learnability over terseness.

**Do:**
- Use clear, descriptive names (`--output-file` not `-o`)
- Provide helpful defaults that "just work"
- Make error messages actionable

**Don't:**
- Require memorization of arcane flags
- Fail silently or with cryptic codes
- Break conventions that users expect

### 2. Be Consistent
Follow established CLI conventions. Users transfer knowledge between tools.

**Patterns to follow:**
- `--help` for assistance
- `--version` for version info
- `--verbose` / `-v` for more output
- `--quiet` / `-q` for less output
- `--dry-run` to preview changes
- stdin/stdout via `-` (e.g., `cmd -`)

### 3. Make Discovery Easy
Users should find functionality without reading manuals.

**Checklist:**
- [ ] Help text is comprehensive and shows examples
- [ ] Subcommands document their purpose
- [ ] Common workflows work without flags
- [ ] Suggestions appear in error messages

### 4. Respect Configuration Precedence
Users expect this order (highest to lowest):
1. Flags (command-line arguments)
2. Environment variables
3. Project config (`.env`, local config files)
4. User config (`~/.config/app/`)
5. System config

## Design Checklist

When building or reviewing a CLI, check each area:

### Flags & Arguments
- [ ] Flags over positional arguments for optional values
- [ ] Full flag names are readable (`--max-retries` not `--mr`)
- [ ] Short flags for commonly used options (`-c`, `-f`, `-v`)
- [ ] Boolean flags are clearly false when omitted
- [ ] `-` reads from stdin / writes to stdout

### Output
- [ ] Human-readable by default
- [ ] `--json` flag for machine-readable output
- [ ] `--plain` or `--no-color` for scripting
- [ ] Progress indicators for long operations
- [ ] `--verbose` increases detail, `--quiet` reduces it

### Errors
- [ ] Errors explain what happened and why
- [ ] Errors suggest how to fix them
- [ ] Errors are human-readable, not raw stack traces
- [ ] Exit codes are documented (0 = success, non-zero = error)

### Help Text
- [ ] Shows common usage patterns
- [ ] Explains what each flag does
- [ ] Includes examples
- [ ] Links to full documentation

### Configuration
- [ ] Uses XDG paths for user config (`~/.config/app/`)
- [ ] Environment variables documented
- [ ] Config file format documented

## Reference Files

Load these for detailed guidance:

| Area           | File                          |
|----------------|-------------------------------|
| Output design  | `references/output.md`        |
| Error messages | `references/errors.md`        |
| Flag design    | `references/flags.md`         |
| Configuration  | `references/configuration.md` |
| Compatibility  | `references/compatibility.md` |

## Quick CLI Review

When asked to review a CLI, use this order:

1. **Install/run** — Does it work out of the box?
2. **`--help`** — Is help clear and complete?
3. **Common tasks** — Can you do the obvious thing without flags?
4. **Error handling** — What happens with bad input?
5. **Flags** — Do they follow conventions?
6. **Output** — Is it readable? Is `--json` available?

## For Go Projects (Cobra/Viper)

Per AGENTS.md, this codebase uses:
- `github.com/spf13/cobra` for commands
- `github.com/spf13/viper` for configuration

Ensure your CLI follows these in addition to clig.dev:
- Viper reads env vars with `automaticEnv()` and `SetEnvPrefix()`
- Cobra root command sets `SilenceUsage: true` to avoid spamming help on errors
- Use `cobra.OnInitialize()` for setup logic
