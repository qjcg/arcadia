# Pavona — Template Engine for Developers

> *Pavona is a **cookiecutter-inspired** template engine for Go. Point it at a
> template directory (local or built-in), answer a few questions, and get a
> fully hydrated project in seconds.*

Named after leaf coral of the *Pavona* genus: layered, branching, and
symbiotic. A single template branches into many possible outputs.

---

## Philosophy

**"Bring your own template, or use one of ours."**

Pavona does one thing and does it well — hydrate templates. It is not a
framework, not a scaffolding tool, not a site builder. It is a template
engine that:

1. Reads a `config.cue` file to discover what variables a template needs.
2. Prompts the user for those variables interactively (with sensible defaults).
3. Renders every `.tmpl` file through Go's `text/template`, stripping the `.tmpl` extension.
4. Copies all other files as-is.
5. Renders directory names through the same template engine.

---

## CLI Interface

```
pavona -t <template> [-o <output-dir>]
```

### Flags

| Flag         | Short | Description                                                               |
|--------------|-------|---------------------------------------------------------------------------|
| `--template` | `-t`  | Template source: a named built-in or a path to a local template directory |
| `--output`   | `-o`  | Output directory (default: current directory)                             |
| `--name`     | `-n`  | Project name (skips the first prompt if provided)                         |
| `--quiet`    | `-q`  | Non-interactive mode — use defaults for all variables                     |
| `--list`     | `-l`  | List available built-in templates and exit                                |

### Examples

```sh
pavona -t tool -o ./my-cli-tool              # built-in "tool" template
pavona -t /path/to/custom-template            # custom template on disk
pavona -t tool -o ./my-cli --name my-cli -q   # non-interactive
pavona --list                                 # show built-in templates
```

### Exit Codes

| Code | Meaning                                                        |
|------|----------------------------------------------------------------|
| 0    | Success                                                        |
| 1    | Template not found, config.cue parse error, or hydration error |
| 2    | Output directory already exists and is not empty               |

---

## Template Directory Structure

```
my-template/
├── config.cue              # REQUIRED: defines variables + metadata
├── {{.name}}/              # directory name uses template syntax
│   ├── main.go.tmpl        # .tmpl file → rendered, extension stripped
│   ├── go.mod.tmpl
│   ├── README.md           # non-.tmpl file → copied verbatim
│   └── internal/
│       └── handler.go.tmpl
├── static/                 # static dir copied as-is
│   └── logo.svg
└── .gitignore.tmpl
```

When hydrated with `name = "my-cli"`:

```
my-cli/
├── main.go
├── go.mod
├── README.md
├── internal/
│   └── handler.go
├── static/
│   └── logo.svg
└── .gitignore
```

### Rules

1. **`config.cue`** is required. It defines the template's metadata and
   variables using CUE type syntax.
2. **`.tmpl` files** are rendered through `text/template` and written without
   the `.tmpl` suffix.
3. **Non-`.tmpl` files** are copied byte-for-byte.
4. **Directory names** containing `{{...}}` are rendered through the same
   template engine.
5. **Hidden files and directories** (starting with `.`) are included and
   processed normally.
6. **`config.cue` itself** is never written to the output.

---

## `config.cue` Format

The CUE configuration file lives at the root of every template directory.
It defines template metadata and all variables needed for hydration.

Variables are defined using **CUE type syntax**, not special struct fields:

| Concept       | CUE Syntax                                      | Meaning                     |
|---------------|-------------------------------------------------|-----------------------------|
| Freeform      | `name: string`                                  | Required text input         |
| Optional      | `name?: string \| *"default"`                   | Optional, with default      |
| Choice        | `name?: "a" \| *"b" \| "c"`                     | Choice list, default "b"    |
| Required      | `name: "a" \| "b"`                              | Choice list, must pick      |
| Prompt text   | `// Doc comment before the field`               | Used as the prompt          |
| Field marker  | `?` on field name                               | Optional field in the struct|

### Key principles

- **No `prompt`, `default`, `required`, `choices`, or `help` fields.**
  These are fully expressed through CUE's type system and doc comments.
- A **doc comment** (`// ...`) above a variable field serves as the prompt
  text displayed to the user.
- The **`?` marker** on a field name makes the variable optional. Without it,
  the user must provide a value (or accept the default).
- A **default value** is indicated with `*value` inside a disjunction.
- **Choices** are a disjunction of concrete string literals
  (e.g. `"a" | "b" | "c"`). If the disjunction includes the `string` type,
  it becomes freeform instead.

### Built-in Template Schema

```cue
package template

name:        string
description: string

variables: { ... }
```

### Example: CLI Tool Template

```cue
package template

name:        "tool"
description: "A Go CLI tool with cobra subcommands and BDD tests"

variables: {
	// Project name (e.g., gh-deploy)
	project_name: string

	// Short description
	description?: string | *"A CLI tool built with Pavona"

	// Initial version
	version?: string | *"0.1.0"
}
```

This defines three variables:
- `project_name` — freeform, required (no `?`, no default)
- `description` — freeform, optional (`?`), defaults to `"A CLI tool built with Pavona"`
- `version` — freeform, optional (`?`), defaults to `"0.1.0"`

### Example: Static Site Template

```cue
package template

name:        "site"
description: "A static site with Markdown or org-mode content and a custom theme"

variables: {
	// Site name (e.g., My Blog)
	site_name: string

	// Author name
	author?: string | *""

	// Content format
	format?: *"markdown" | "org"
}
```

Here `format` is a **choice** — the user selects from `"markdown"` or
`"org"`, with `"markdown"` as the default. `author` is an optional freeform field with
an empty default (user can skip it).

### How CUE Types Map to Prompts

| CUE Definition                              | Prompt Type  | Required | Default  |
|---------------------------------------------|--------------|----------|----------|
| `name: string`                              | text input   | yes      | none     |
| `name?: string \| *"hello"`                 | text input   | no       | `"hello"`|
| `name?: "x" \| *"y" \| "z"`                | list select  | no       | `"y"`    |
| `name: "x" \| "y"`                         | list select  | yes      | none     |

---

## Built-in Templates

Pavona ships with a set of built-in templates embedded at build time via
`//go:embed`. They live under `internal/scaffold/templates/` and are
compiled into the binary.

| Name    | Description                                                 |
|---------|-------------------------------------------------------------|
| `tool`  | Go CLI tool with cobra-based subcommands and BDD tests      |
| `lib`   | Minimal Go library module with test helpers                 |
| `site`  | Static site with Markdown/org-mode content and custom theme |
| `tui`   | Terminal UI app using bubbletea                             |
| `app`   | Full-stack web app with templ, SQLite, HTMX, and Tailwind   |
| `agent` | NATS Agent Protocol service                                 |

Each built-in template follows the same `config.cue` + `*.tmpl` structure.

---

## Template Resolution Order

When the user passes `-t <source>`, pavona resolves the template in this order:

1. **Built-in match** — if `<source>` matches a built-in template name, use it.
2. **Exact path** — if `<source>` is a directory containing `config.cue`, use it.
3. **XDG data path** — check `$XDG_DATA_HOME/pavona/templates/<source>/` (or
   `~/.local/share/pavona/templates/<source>/`).
4. **Error** — if none of the above match, exit with code 1.

---

## Interactive Prompt Flow

Pavona uses **bubbletea v2** (via `charm.land/bubbletea/v2`) with the
`charm.land/bubbles/v2` component library for interactive prompts:

1. Read `config.cue` and extract `variables` using the CUE Go API.
2. For each variable (alphabetically sorted):
   - **Freeform** (`string` type) — show a `textinput.Model` with the
     default value pre-filled.
   - **Choice** (disjunction of string literals) — show a `list.Model`
     with the default item pre-selected.
3. User navigates with arrow keys (list) or types text (freeform) and
   presses Enter to confirm.
4. Pressing Ctrl+C at any point cancels the wizard.
5. Assemble the variable map.
6. Walk the template directory, render every `.tmpl` file, copy every other
   file, render directory names.

### Prompt Components

| Component      | Package                      | Used For           |
|----------------|------------------------------|--------------------|
| `textinput`    | `charm.land/bubbles/textinput` | Freeform variables |
| `list`         | `charm.land/bubbles/list`      | Choice variables   |
| `lipgloss`     | `charm.land/lipgloss/v2`       | Styling            |

---

## Architecture

```
pavona/
├── main.go                      # CLI entry point
├── internal/
│   ├── cli/
│   │   └── template.go          # -t flag parsing, template resolution
│   └── scaffold/
│       ├── scaffold.go          # Template engine: walk, render, copy
│       ├── config.go            # config.cue parsing (cuelang SDK)
│       ├── prompt.go            # Interactive prompts via bubbletea
│       └── templates/           # Built-in templates (embedded)
│           ├── tool/
│           │   ├── config.cue
│           │   ├── {{.project_name}}/
│           │   │   ├── main.go.tmpl
│           │   │   ├── go.mod.tmpl
│           │   │   └── ...
│           │   └── ...
│           ├── lib/
│           ├── site/
│           ├── tui/
│           ├── app/
│           └── agent/
```

### Package Responsibilities

| Package                | Responsibility                                    |
|------------------------|---------------------------------------------------|
| `main`                 | Parse flags, invoke `scaffold.Hydrate()`          |
| `cli/template.go`      | CLI parsing, `-t`/`-o`/`-q` flags                 |
| `scaffold/scaffold.go` | Walk template dir, render files, manage output    |
| `scaffold/config.go`   | Load and validate `config.cue`, extract variables |
| `scaffold/prompt.go`   | Interactive bubbletea prompts for each variable   |

### Template Resolution (`scaffold/scaffold.go`)

```go
// Resolve finds a template source: built-in, path, or XDG data.
func Resolve(name string) (string, error)

// Hydrate renders a template directory into an output directory.
func Hydrate(templateDir, outputDir string, vars map[string]string) error

// ListBuiltin returns names and descriptions of all built-in templates.
func ListBuiltin() []TemplateInfo
```

### Config Parsing (`scaffold/config.go`)

```go
// VariableType indicates how a variable is presented to the user.
type VariableType int
const (
    TypeFreeform VariableType = iota // free-form text input
    TypeChoice                       // select from a list of options
)

// Config is the parsed config.cue for a template.
type Config struct {
    Name        string
    Description string
    Variables   []Variable
}

// Variable defines a single template variable, inferred from CUE types.
type Variable struct {
    Name     string
    Prompt   string      // From CUE doc comments
    Type     VariableType
    Default  string      // From CUE *default syntax
    Required bool        // From ? field marker
    Choices  []string    // From CUE disjunction literals
}

// ParseConfig reads and validates config.cue from a directory.
func ParseConfig(dir string) (*Config, error)
```

### Prompt Flow

```go
// PromptForVariables asks the user for each variable interactively
// using bubbletea (textinput for freeform, list for choices).
// In quiet mode (-q), returns defaults for everything.
func PromptForVariables(vars []Variable, quiet bool) map[string]string
```

---

## Hydration Algorithm

```
function Hydrate(templateDir, outputDir, vars):
  config = ParseConfig(templateDir)

  for each entry in Walk(templateDir):
    relative = entry.path relative to templateDir

    if entry.name == "config.cue":
      continue                                    // skip config file

    // Render directory names
    renderedPath = render(relative, vars)
    renderedPath = strings.TrimSuffix(renderedPath, ".tmpl")

    if entry is directory:
      createDir(filepath.Join(outputDir, renderedPath))

    if entry is file:
      sourcePath = entry.path
      destPath = filepath.Join(outputDir, renderedPath)

      if strings.HasSuffix(entry.name, ".tmpl"):
        content = readFile(sourcePath)
        rendered = render(content, vars)          // text/template
        writeFile(trimSuffix(destPath, ".tmpl"), rendered)
      else:
        copyFile(sourcePath, destPath)            // byte-for-byte
```

---

## Error Handling

| Scenario                        | Behavior                                              |
|---------------------------------|-------------------------------------------------------|
| `-t` not specified              | Print usage and exit 1                                |
| Template not found              | Print "template not found: <name>" and list built-ins |
| `config.cue` parse error        | Print CUE error and exit 1                            |
| Output dir exists and non-empty | Print error and exit 2 (unless `--force`)             |
| Template execution error        | Print file + template error and exit 1                |
| User cancels (Ctrl+C)           | Print "Cancelled." and exit 1                         |

---

## Testing Strategy

| Layer        | What                | How                                               |
|--------------|---------------------|---------------------------------------------------|
| Unit         | Config parsing      | Parse sample `config.cue` files, verify variables |
| Unit         | Template rendering  | Render a `.tmpl` file, verify output              |
| Unit         | File walking        | Walk a mock template dir, verify callbacks        |
| Integration  | `pavona -t tool`    | Run the CLI, verify output structure              |
| Integration  | `pavona -t /custom` | Run with an external template                     |
| Golden files | Full hydration      | Compare hydrated output against golden files      |

---

## Future Possibilities

- **Template repositories**: `pavona -t github.com/user/repo`
- **Template chains**: `pavona -t base,plugin` — layer templates
- **Post-hydration hooks**: scripts in `config.cue` that run after hydration
- **Template authoring**: `pavona init-template` to create a template skeleton
