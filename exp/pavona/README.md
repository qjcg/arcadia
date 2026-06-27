# Pavona

*Pavona is a Go scaffolding tool and framework for building CLI tools,
libraries, static sites, TUIs, web apps, and agents — with an emphasis
on developer happiness, minimal lock-in, and boring-is-beautiful
design.*

Named after leaf coral of the *Pavona* genus: layered, branching, and
symbiotic. You start with a single `pavona new ...` and own every file
it produces.

## Installation

```sh
go install github.com/qjcg/arcadia/exp/pavona@latest
```


## Quick Start

Scaffold any project type in seconds:

```sh
pavona new tool gh-deploy        # CLI tool with subcommands
pavona new lib go-csvstream      # Go library with test helpers
pavona new site blog             # Static site (markdown or org-mode)
pavona new tui chatmonitor       # Terminal UI (bubbletea)
pavona new app urlshort --demo   # Full-stack web app (templ + HTMX + SQLite)
pavona new agent triagebot       # NATS Agent Protocol service
```

Every scaffold compiles on first try and includes a BDD test suite.

## Commands

| Command  | Description                                               |
|----------|-----------------------------------------------------------|
| `new`    | Scaffold a new project (tool, lib, site, tui, app, agent) |
| `add`    | Add features to an existing project                       |
| `remove` | Remove features from an existing project                  |
| `build`  | Build a static site from content to HTML                  |
| `serve`  | Start the dev server with live reload                     |

## Site Builder

The `site` type generates a static site with a filesystem-as-URL structure:

```
blog/
├── content/
│   ├── index.md              →  /
│   ├── about.md              →  /about.html
│   └── services/
│       ├── index.md          →  /services/
│       ├── consulting.md     →  /services/consulting.html
│       └── support.md        →  /services/support.html
├── theme/default.templ       # DaisyUI 5, dark mode, full-text search
├── build.go                  # Self-contained — works without the CLI
├── package.json              # Tailwind v4 + DaisyUI 5
└── go.mod
```

Supports Markdown (default) and org-mode (via `--format org`). Themes are built with `templ` and fully customizable.

## Full-Stack App Scaffold

The `app` type with `--demo` flag produces a working URL shortener with:

- Go backend with `net/http`, SQLite, and embedded migrations
- `templ` templates and HTMX for dynamic interactions
- Alpine.js for client-side state (filterable/sortable table)
- QR code generation
- Tailwind CSS v4 + DaisyUI 5
- Mobile-responsive layout with hamburger menu

## Design Principles

| #  | Principle                  | What it means                                                               |
|----|----------------------------|-----------------------------------------------------------------------------|
| 1  | **Developer happiness**    | Scaffold compiles on first try. Tests pass with no setup.                   |
| 2  | **Surfaceless**            | No `init()` magic. No hidden goroutines. Everything is explicit.            |
| 3  | **Peel away**              | Start with Pavona, replace any component with your own.                     |
| 4  | **Compile-time > runtime** | If `go build` can catch it, it should.                                      |
| 5  | **Locality**               | Related code lives together. No folder religion. No import cycles.          |
| 6  | **Testable by default**    | Every component produces BDD step helpers and clean interfaces.             |
| 7  | **5-minute onboarding**    | The scaffold *is* the documentation.                                        |
| 8  | **No lock-in exit**        | Drop `pavona` from `go.mod` and the project still compiles.                 |
| 9  | **Symmetry**               | Every `pavona add` has a corresponding `pavona remove`.                     |
| 10 | **Boring is beautiful**    | `net/http`, `slog`, `database/sql`, `text/template`. Reads like the stdlib. |

## Development

```sh
task build       # Build the CLI
task test        # Run all tests (BDD scenarios)
task install     # Install to $GOBIN
task dev         # Watch + test with watchexec
```

## License

GPL-3.0-only
