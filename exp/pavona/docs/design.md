# Pavona

> A Go framework that grows with you — from CLI tools to full-stack web apps to agents.

Named after leaf coral of the *Pavona* genus: layered, branching, and symbiotic. You start with a single `main.go` and extend outward in predictable patterns.

## Philosophy

**"Pavona gets out of your way — and stays out."**

CLI tools, libraries, static sites, TUIs, web apps, and agents share 80% of the same DNA —
config loading, dependency wiring, logging, testing infrastructure,
command routing, middleware patterns. Pavona provides that 80% as
modular, composable layers, then stays silent for the other 20%.

The framework is built around 10 principles:

| #  | Principle                  | What it means                                                                                                                                                              |
|----|----------------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| 1  | **Developer happiness**    | The scaffold compiles on first try. Tests pass with no setup. Hot reload works out of the box. Every interaction leaves you satisfied, not frustrated.                     |
| 2  | **Surfaceless**            | No `init()` magic. No hidden goroutines. Every `Register`, `Start`, and `Stop` is explicit and traceable.                                                                  |
| 3  | **Peel away**              | Start with Pavona providing everything, then replace any component with your own. The framework cedes control gracefully.                                                  |
| 4  | **Compile-time > runtime** | If `go build` can catch it, it should. Generic `Register[T]`, sqlc queries, typed config structs. Zero panic-from-reflection paths.                                        |
| 5  | **Locality**               | Related code lives together. The scaffold suggests a layout but never enforces a folder religion. You can move a handler next to its tests without fighting import cycles. |
| 6  | **Testable by default**    | Every Pavona component produces a clean interface, godog BDD step helpers, and a test helper that works without network, filesystem, or global state.                                               |
| 7  | **5-minute onboarding**    | `pavona new` produces a project a new team member can understand in 5 minutes. The scaffold *is* the documentation.                                                    |
| 8  | **No lock-in exit**        | Drop `pavona` from `go.mod` and the project still compiles. Templates produce standard Go files you own.                                                                   |
| 9  | **Symmetry**               | Every `pavona add` has a corresponding `pavona remove`. What the scaffold creates it can also delete cleanly.                                                              |
| 10 | **Boring is beautiful**    | `net/http`, `slog`, `database/sql`, `text/template`. Not fashionable abstractions. Pavona reads like the stdlib.                                                           |

## Core Architecture

```
pavona/
├── cmd/           # Pavona CLI — scaffolding, dev server, codegen
├── app/           # Runtime kernel — bootstrap, lifecycle, shutdown
├── conf/          # Config loading (YAML/TOML/JSON/env, layered)
├── log/           # Structured logging (slog-based, level-filtered)
├── wire/          # Module registration & dependency graph declaration
├── serve/         # HTTP server (net/http, middleware stack, graceful shutdown)
├── agent/         # NATS Agent Protocol host: prompt/status/hb endpoints
├── pkg/site/      # Static site builder — public API (exported from internal)
├── cli/           # CLI framework (cobra-like, top-level subcommands)
├── tui/           # TUI framework (bubbletea-based, composable components)
├── pool/          # Worker pool / background job runner
├── db/            # Database abstractions (migrations, queries, transactions)
│   └── migrate/   # Migration runner (embedded, up/down/status)
├── test/          # Test helpers, golden file utils, DB fixtures, clock mocking, godog BDD steps
└── gen/           # Code generation & template engine (templates + overrides)
```

### The Kernel (`app`)

Every Pavona project has a single entry point that creates an `App`:

```go
app := pavona.New(
    pavona.WithConfig("config.yaml"),
    pavona.WithModules(api, cli, jobs),
    pavona.WithLogger(log.LevelInfo),
)

app.Start()   // parses flags, loads config, starts modules
app.Wait()    // blocks on signals (SIGINT/SIGTERM)
app.Stop()    // graceful shutdown in dependency order
```

`App` is not a framework — it's a **lifecycle manager**. Modules
register start/stop hooks, health checks, and depend on each
other. Pavona resolves the DAG and starts/stops in order.

## Six Project Types

### 1. CLI Tool

```
pavona new tool gh-deploy
```

Generates: single binary with subcommands (`deploy`, `rollback`,
`status`, `config`). Uses `pavona/cli` for command routing,
`pavona/conf` for config, and `pavona/pool` for concurrent API calls.
Includes a `features/` directory with godog suite.

```go
cmd := cli.New("gh-deploy").
    Add(deployCmd).
    Add(rollbackCmd).
    WithConfig(pavona.DefaultCLIConfig())

app := pavona.New(pavona.WithCLI(cmd))
```

### 2. Library

```
pavona new lib go-csvstream
```

Generates: minimal Go module with `pavona/test` helpers, godog BDD
scaffold, CI setup, and example-driven docs. The library uses zero
Pavona runtime dependencies in production — only the dev toolchain.
Includes a `features/` directory with godog suite.

### 3. Static Site

```
pavona new site blog                            # markdown (default)
pavona new site blog --format org               # org-mode
pavona new site blog --site-name "My Blog"      # custom display name
pavona new site blog --format org --site-name "My Blog" --theme modern
```

Generates: a Go module with content files, a `templ`-based theme, and
a `build.go` that renders the site. The project becomes self-contained —
you can `go run build.go` to produce `dist/` without the Pavona CLI.

```
blog/
├── content/
│   └── index.md              # or index.org
├── theme/
│   └── default.templ         # user-editable theme template (DaisyUI 5, dark mode toggle)
├── static/
│   └── style.css             # CSS entry point (Tailwind v4 + DaisyUI 5)
├── build.go                  # renders content through theme, builds CSS
├── package.json              # npm deps: tailwindcss v4, daisyui 5.6.3, @tailwindcss/cli
├── go.mod
└── features/
    └── .gitkeep
```

Content follows a **filesystem-as-URL** convention — every file maps
directly to a URL path:

```
content/
├── index.md              →  /              (home page)
├── about.md              →  /about.html    (flat page)
└── services/                                (section)
    ├── index.md          →  /services/      (section landing)
    ├── consulting.md     →  /services/consulting.html
    └── support.md        →  /services/support.html
```

Sections (directories with an `index.md`) produce clean `/section/` URLs
with no `.html` suffix. Flat files produce `/path.html`. The file tree
*is* the sitemap — no config file needed.

Markdown files support optional YAML frontmatter:

```yaml
---
title: Consulting Services
order: 2
---
```

The `order` field controls page position in navigation. Files without
frontmatter fall back to detecting the title from the first `# heading`
(markdown) or `#+TITLE` keyword (org-mode). The `draft: true` flag
excludes a page from the build but keeps it available in the dev server.

### Theming

Every site includes a `theme/` directory with a `templ` file that wraps
rendered content in a full HTML page. The theme has full access to `templ`
features: conditionals, loops, nested components, CSS/Tailwind classes.

The scaffolded theme uses **Tailwind CSS v4** with **DaisyUI 5.6.3**,
providing a built-in dark mode toggle. The toggle cycles through three
states (system, light, dark) and defaults to the user's OS preference.
The preference is persisted in `localStorage`.

**Theme contract** — the templ component receives three values:

- `title` — page title (from frontmatter or heading detection)
- `content` — rendered HTML body (from goldmark for Markdown or go-org for org-mode)
- `pages` — navigation tree passed as `[]site.TreeNode`

The scaffolded `theme/default.templ` defines the component signature:

```templ
package theme

type Page struct {
    URL   string
    Title string
}

templ Layout(title string, content string, pages []Page) {
    <!DOCTYPE html>
    <html lang="en" data-theme="system">
        <head>
            <meta charset="UTF-8"/>
            <meta name="viewport" content="width=device-width, initial-scale=1.0"/>
            <title>{ title } &mdash; {{.SiteName}}</title>
            <link rel="stylesheet" href="/style.css"/>
        </head>
        <body class="bg-base-200 min-h-screen">
            <header>
                <div>
                    <a href="/">&#127754; {{.SiteName}}</a>
                    <nav>
                        for _, p := range pages {
                            if p.URL != "index.html" {
                                <a href={ templ.URL(p.URL) }>{ p.Title }</a>
                            }
                        }
                    </nav>
                    <button onclick="cycleTheme()">...</button>
                </div>
            </header>
            <main>
                <article>{ content }</article>
            </main>
        </body>
    </html>
}
```

The `site.TreeNode` type is exported from Pavona's public `pkg/site/`
package (moved from `internal/site/`):

```go
package site

type TreeNode struct {
    Title    string
    URL      string
    Children []TreeNode
}
```

Users create custom themes by editing `theme/default.templ` or creating
new `.templ` files in the theme directory. The `--theme` flag selects
which directory to use:

```
pavona build --theme ./themes/docs
pavona serve --theme ./themes/blog --port 4000
```

**How it works** — the generated `build.go` ties content and theme
together at compile time:

```go
// build.go — generated by scaffold, owned by user
package main

import (
    "context"
    "os"
    "path/filepath"

    "github.com/qjcg/arcadia/exp/pavona/pkg/site"
    "github.com/user/blog/theme"
)

func main() {
    pages, nav, err := site.BuildPageTree("content")
    if err != nil { panic(err) }
    ctx := context.Background()

    for _, p := range pages {
        if p.Draft { continue }
        out := filepath.Join("dist", p.URL)
        os.MkdirAll(filepath.Dir(out), 0755)
        f, _ := os.Create(out)
        theme.Layout(p.Title, string(p.HTML), nav.Children).Render(ctx, f)
        f.Close()
    }
}
```

`site.BuildPageTree` walks `content/`, parses frontmatter, renders
Markdown/org to HTML, builds the navigation tree, and returns a flat
page list plus a rooted tree for nav. It's the same pipeline Pavona
uses internally — now exported for projects to call directly.

**Dev server (`pavona serve`):**

The dev server watches `content/` for changes to `.md` and `.org` files
using `fsnotify`. When a content file changes, the site is rebuilt
automatically (debounced at 200ms) and the browser picks up the new
HTML on next refresh. Only the content rebuild runs — the cached
renderer binary is reused for fast iteration.

The scaffolded site also builds CSS from Tailwind v4 + DaisyUI 5:
the `build.go` runs `npx @tailwindcss/cli` to compile `static/style.css`
into `dist/style.css` before rendering content.

For scaffolded sites, the build flow is:

```
cd blog && go run build.go
```

The generated `build.go` compiles CSS via `@tailwindcss/cli`, then walks
`content/` and renders each page through the theme. The project is fully
self-contained — no Pavona CLI needed after scaffold.

**Backward compatibility:** The Pavona CLI's `pavona build` command
uses an embedded `gohtml` fallback that works without a `theme/`
directory, for quick prototyping or existing projects without a custom
theme.

### 4. TUI App

```
pavona new tui chatmonitor
```

Generates: bubbletea-based terminal app with component structure
(views/, models/, commands/), keyboard-driven navigation, log viewer,
help overlay, and `Taskfile.yaml`. Uses `pavona/tui` for the component
model, layout primitives, and keybinding management. Includes a
`features/` directory with godog suite.

### 5. Full-Stack Web App

```
pavona new app acmecorp
```

Generates: HTTP server, static file serving, templ rendering (via
a-h/templ), HTMX + Alpine.js wiring, SQLite (with sqlc), CSS pipeline
(Tailwind via daisyui), `features/` directory with godog suite, Dockerfile,
and `Taskfile.yaml`.

### 6. Agent

```
pavona new agent triagebot
```

Generates: a Go service that speaks the NATS Agent Protocol — registering
as a NATS micro service named `agents` with `prompt`, `status`, and `hb`
(heartbeat) endpoints. Uses `pavona/agent` for the protocol implementation.
Includes a `features/` directory with godog feature files for BDD testing
of prompt/response flows.

```go
import "github.com/pavona/pavona/agent"

agt := agent.New("triagebot",
    agent.WithDescription("Triage incoming support tickets"),
    agent.WithPrompt(func(ctx context.Context, req agent.Request) agent.Response {
        // req.Text is the prompt, req.Attachments carries base64 blobs
        // Response can stream typed JSON chunks
        return agent.NewResponse().
            Write("Analyzing ticket contents...").
            Write(result)
    }),
    agent.WithStatus(func(ctx context.Context) agent.Status {
        return agent.Status{Healthy: true, Load: currentLoad}
    }),
)

app := pavona.New(
    pavona.WithConfig("config.yaml"),
    pavona.WithModules(agt),
    pavona.WithLogger(log.LevelInfo),
)
app.Start()
app.Wait()
app.Stop()
```

**Generated structure:**

```
agent/
├── main.go
├── agent/
│   └── handler.go      # prompt/status/hb logic
├── knowledge/           # reference docs, embeddings, tools
├── nats/
│   ├── conn.go          # connection lifecycle (start/stop hooks)
│   └── server.go        # embedded server in dev, remote in production
├── features/            # gherkin feature files + godog suite
├── Taskfile.yaml
├── go.mod
└── config.yaml
```

Every agent is automatically discoverable via `nats req '$SRV.INFO.agents'`.

## Templates

Every project type can be extended with a `--template` flag that wires in
a specific technology stack. Templates are versioned, community-contributed
packages stored in a registry. They overlay scaffolding on top of the base
project type — adding config, dependencies, wiring code, and examples.

```
pavona new app dashboard --template nats
pavona new tui chatmonitor --template nats
pavona new tool eventsink --template nats
```

### NATS Template

A "nothing but NATS" template that scaffolds a project with NATS as the sole
backing service. No database, no HTTP server (unless the project type
requires one) — just NATS for pub/sub, request/reply, and JetStream
persistence.

**What it wires:**
- `go get github.com/nats-io/nats.go` and embeds
  `github.com/nats-io/nats-server/v2` as a library
  (no external process needed)
- A `nats/` package with reusable connection lifecycle using the Pavona `app`
  kernel's start/stop hooks
- Embedded NATS server configured via `pavona/conf` (memory-backed by default,
  filesystem persistence opt-in)
- A `nats/streams/` directory for declarative JetStream stream and consumer
  definitions
- Three example patterns:
  - **pub/sub**: event bus for in-process communication
  - **request/reply**: RPC-style handlers wired to CLI subcommands or HTTP
    endpoints
  - **JetStream work queue**: durable consumer with Pavona's `pool/` worker
    pulling and ack'ing messages
- Test helpers: `nats/test` with an embedded server spun up per test suite,
  no network dependency

**Generated structure (app + nats template):**

```
app/
├── main.go
├── handlers/
├── nats/
│   ├── conn.go          # connection lifecycle (start/stop hooks)
│   ├── server.go        # embedded server setup
│   └── streams/
│       └── events.go    # stream & consumer definitions
├── pool/
│   └── workers.go       # JetStream consumers as app modules
├── db/                  # absent — NATS is the only backend
├── static/
├── features/            # gherkin feature files + godog suite
├── Taskfile.yaml
├── go.mod
└── config.yaml
```

**Why NATS as a template:**
NATS is the simplest possible backend that still has real teeth —
persistence, clustering, RPC, streaming, key-value, object store. A
NATS-only project is production-viable for event-driven services, lightweight
backends, and inter-service communication. The embedded server means zero
infrastructure to develop against.

### Extending with Templates

```
pavona template new my-stack       # create a new template from scratch
pavona template publish my-stack   # publish to the registry
pavona template search postgres    # discover community templates
```

Templates are Go modules with a `template.yaml` manifest that declares
dependencies, file templates (using Go's `text/template`), and hooks that
run after scaffolding. They compose with project types — a template can add
files to any project type that has the required hooks.

## Key Design Decisions

| Area           | Decision                                                                                             | Rationale                                                                    |
|----------------|------------------------------------------------------------------------------------------------------|------------------------------------------------------------------------------|
| **DI**         | No reflection. Modules self-register via `pavona.Register[T]`.                                       | Compile-time safety, no magic.                                               |
| **Config**     | `conf` merges: defaults → file → env → flags.                                                        | The Unix philosophy — layers override cleanly.                               |
| **DB**         | Pavona provides migration runner and connection lifecycle. Actual queries use **sqlc**.              | Type-safe SQL beats any ORM.                                                 |
| **Middleware** | Standard `net/http` middleware. Pavona provides `recover`, `request-id`, `access-log`, `rate-limit`. | No custom handler signature. Compatible with everything.                     |
| **TUI**        | `tui` package wraps bubbletea with layout primitives, keybinding registry, and screen primitives.    | Start a full-screen TUI in 10 lines. Components are reusable modules.        |
| **Testing**    | `test` package gives you: godog BDD suite with gherkin features, temp DB, golden file comparison, request recording, clock mocking. | The stuff you rewrite for every project.                                     |
| **Build**      | Single `go build`, no codegen step at runtime.                                                                                    | Fast iteration, simple debugging.                                            |
| **Scaffold**   | Codegen produces files once — you own them after.                                                    | No framework lock-in. You can delete `pavona` from go.mod and still compile. |
| **Agent**      | `agent` package implements the NATS Agent Protocol as a NATS micro service with prompt/status/hb.    | Agents are discoverable, addressable, and composable without a central registry. |

## Growth Path

```
pavona new site blog                        # generates: content/, static/, Tailwind config
cd blog && pavona add page about            # adds content/about.md
pavona build                                # produces dist/

pavona new app dashboard --template nats    # generates: main.go, handlers/, nats/
cd dashboard && pavona add handler auth     # adds handler + route + middleware
pavona add migration add_sessions           # generates migration SQL + sqlc query
pavona serve                                # dev server with hot reload
pavona build                                # single binary: ./bin/dashboard

pavona new agent triagebot                  # generates: main.go, agent/, nats/
pavona add tool escalate                    # adds CLI subcommand that prompts the agent
pavona add feature prompt_response          # adds gherkin feature + godog step definitions
```

The same binary can serve HTTP, expose CLI commands, launch a TUI, build a
static site, *or* speak the NATS Agent Protocol — all sharing the same domain code.

## Summary

Pavona is not the biggest or most opinionated framework. It is the
**most ergonomic** — designed to be picked up in 5 minutes and
outgrown gracefully. Like coral, it provides structure for life to
flourish, then becomes part of the reef.
