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

Create a project from any built-in template:

```sh
pavona new tool -o ./my-cli
pavona new lib -o ./my-lib
pavona new site -o ./my-site
pavona new tui -o ./my-tui
pavona new app -o ./my-app
pavona new agent -o ./my-agent
pavona new pavona -o ./my-template
pavona new monorepo-go -o ./my-monorepo
```

### Create a custom template

```sh
pavona new pavona -o ./my-template -n my-template -q
pavona new ./my-template -o ./my-project -n my-project -q
```

The generated directory includes a starter `config.cue`.

### Non-interactive mode

```sh
pavona new tool -o ./my-cli -n my-cli -q
```

### List built-in templates

```sh
pavona list
pavona ls
```

### Use a custom template

```sh
pavona new /path/to/my-template -o ./project
```

---

## Commands

| Command | Description |
|---------|-------------|
| `pavona new <template>` | Create a project from a built-in or local template |
| `pavona list` | List built-in templates |
| `pavona ls` | Short alias for `list` |

The `new` command accepts `--output` (`-o`), `--name` (`-n`), and `--quiet` (`-q`).

---

## Built-in Templates

| Name    | Description                                                   |
|---------|---------------------------------------------------------------|
| `tool`  | Go CLI tool with cobra subcommands and BDD tests              |
| `lib`   | Minimal Go library module with test helpers                   |
| `site`  | Static site with Markdown or org-mode content                 |
| `tui`   | Terminal UI app using bubbletea                               |
| `app`   | Full-stack web app with templ, SQLite, HTMX, Tailwind/DaisyUI |
| `agent`  | NATS Agent Protocol service with JetStream                    |
| `pavona`      | Starter template for creating Pavona templates                |
| `monorepo-go` | Go workspace monorepo                                         |

---

## Creating Custom Templates

Every template needs a `config.cue` file at its root. Start with the built-in scaffold:

```sh
pavona new pavona -o ./my-template -n my-template -q
```

It generates a starter `config.cue` that you can extend with your own template files.

```cue
package template

name:        "my-template"
description: "A custom template"

variables: {
	// Project name
	project_name: string

	// Greeting message
	message?: string | *"Hello, World!"
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
