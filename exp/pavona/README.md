# Pavona — Template Engine for Developers

> *Pavona is a **cookiecutter-inspired** template engine for Go. Point it at a
> template directory (or use a built-in), answer a few questions, and get a
> fully hydrated project in seconds.*

Named after leaf coral of the *Pavona* genus: layered, branching, and
symbiotic. A single template branches into many possible outputs.

---

## Installation

```sh
go install github.com/qjcg/arcadia/exp/pavona@latest
```

## Quick Start

Hydrate any built-in template in seconds:

```sh
pavona -t tool -o ./my-cli         # Go CLI tool with cobra subcommands
pavona -t lib -o ./my-lib          # Go library with test helpers
pavona -t site -o ./my-site        # Static site (markdown or org-mode)
pavona -t tui -o ./my-tui          # Terminal UI (bubbletea)
pavona -t app -o ./my-app          # Full-stack web app (templ + HTMX + SQLite)
pavona -t agent -o ./my-agent      # NATS Agent Protocol service
```

### Non-interactive mode

```sh
pavona -t tool -o ./my-cli -n my-cli -q
```

### List built-in templates

```sh
pavona -l
```

### Use a custom template

```sh
pavona -t /path/to/my-template -o ./project
```

---

## Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--template` | `-t` | Template source: built-in name or path to a template directory |
| `--output` | `-o` | Output directory (default: derived from project name) |
| `--name` | `-n` | Project name (pre-fills the name prompt) |
| `--quiet` | `-q` | Non-interactive mode — use defaults for all variables |
| `--list` | `-l` | List available built-in templates |

---

## Built-in Templates

| Name | Description |
|------|-------------|
| `tool` | Go CLI tool with cobra subcommands and BDD tests |
| `lib` | Minimal Go library module with test helpers |
| `site` | Static site with Markdown or org-mode content |
| `tui` | Terminal UI app using bubbletea |
| `app` | Full-stack web app with templ, SQLite, HTMX, Tailwind/DaisyUI |
| `agent` | NATS Agent Protocol service with JetStream |

---

## Creating Custom Templates

Every template needs a `config.cue` file at its root:

```cue
package template

name:        "my-template"
description: "A custom template"

variables: {
	project_name: {
		prompt:    "Project name"
		default:   ""
		required:  true
	}
	message: {
		prompt:    "Greeting message"
		default:   "Hello, World!"
		required:  false
	}
}
```

Template files use Go's `text/template` syntax with the variables as the data context:

```go
// {{.project_name}}/main.go.tmpl
package main

import "fmt"

func main() {
	fmt.Println("{{.message}}")
}
```

### Template file rules

- Files ending in `.tmpl` are rendered through Go's `text/template` and written
  without the `.tmpl` suffix.
- Files without `.tmpl` are copied byte-for-byte.
- Directory names containing `{{...}}` are rendered as templates.
- `config.cue` is consumed by Pavona and never written to the output.

---

## Template Resolution Order

1. **Built-in** — if the name matches a built-in template, use it.
2. **Exact path** — if the argument is a directory with `config.cue`, use it.
3. **XDG data** — check `$XDG_DATA_HOME/pavona/templates/<name>/`.
4. **Error** — if none match, exit with code 1.

---

## Design

See [docs/design.md](docs/design.md) for the full architecture document.

---

## Development

```sh
task build       # Build the CLI
task test        # Run all tests
task install     # Install to $GOBIN
task dev         # Watch + test with watchexec
```

## License

GPL-3.0-only
