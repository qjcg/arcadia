# Pavona

> A Go framework that grows with you — from CLI tools to full-stack web apps.

Named after leaf coral of the *Pavona* genus: layered, branching, and symbiotic. You start with a single `main.go` and extend outward in predictable patterns.

## Philosophy

**"Pavona gets out of your way — and stays out."**

CLI tools, libraries, TUIs, and web apps share 80% of the same DNA —
config loading, dependency wiring, logging, testing infrastructure,
command routing, middleware patterns. Pavona provides that 80% as
modular, composable layers, then stays silent for the other 20%.

The framework is built around 10 principles:

| # | Principle | What it means |
|---|-----------|---------------|
| 1 | **Developer happiness** | The scaffold compiles on first try. Tests pass with no setup. Hot reload works out of the box. Every interaction leaves you satisfied, not frustrated. |
| 2 | **Surfaceless** | No `init()` magic. No hidden goroutines. Every `Register`, `Start`, and `Stop` is explicit and traceable. |
| 3 | **Peel away** | Start with Pavona providing everything, then replace any component with your own. The framework cedes control gracefully. |
| 4 | **Compile-time > runtime** | If `go build` can catch it, it should. Generic `Register[T]`, sqlc queries, typed config structs. Zero panic-from-reflection paths. |
| 5 | **Locality** | Related code lives together. The scaffold suggests a layout but never enforces a folder religion. You can move a handler next to its tests without fighting import cycles. |
| 6 | **Testable by default** | Every Pavona component produces a clean interface and a test helper that works without network, filesystem, or global state. |
| 7 | **5-minute onboarding** | `pavona new app` produces a project a new team member can understand in 5 minutes. The scaffold *is* the documentation. |
| 8 | **No lock-in exit** | Drop `pavona` from `go.mod` and the project still compiles. Templates produce standard Go files you own. |
| 9 | **Symmetry** | Every `pavona add` has a corresponding `pavona remove`. What the scaffold creates it can also delete cleanly. |
| 10 | **Boring is beautiful** | `net/http`, `slog`, `database/sql`, `text/template`. Not fashionable abstractions. Pavona reads like the stdlib. |

## Core Architecture

```
pavona/
├── cmd/           # Pavona CLI — scaffolding, dev server, codegen
├── app/           # Runtime kernel — bootstrap, lifecycle, shutdown
├── wire/          # Dependency injection & module registration
├── conf/          # Config loading (YAML/TOML/JSON/env, layered)
├── log/           # Structured logging (slog-based, level-filtered)
├── serve/         # HTTP server (net/http, middleware stack, graceful shutdown)
├── cli/           # CLI framework (cobra-like, top-level subcommands)
├── tui/           # TUI framework (bubbletea-based, composable components)
├── pool/          # Worker pool / background job runner
├── test/          # Test helpers, golden file utils, DB fixtures
├── type/          # Shared types & validation (by convention, not inheritance)
├── db/            # Database abstractions (migrations, queries, transactions)
│   └── migrate/   # Migration runner (embedded, up/down/status)
└── gen/           # Code generation framework (templates + overrides)
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

## Four Project Types

### 1. CLI Tool

```
pavona new tool gh-deploy
```

Generates: single binary with subcommands (`deploy`, `rollback`,
`status`, `config`). Uses `pavona/cli` for command routing,
`pavona/conf` for config, and `pavona/pool` for concurrent API calls.

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

Generates: minimal Go module with `pavona/test` helpers, CI setup, and
example-driven docs. The library uses zero Pavona runtime dependencies
in production — only the dev toolchain.

### 3. TUI App

```
pavona new tui chatmonitor
```

Generates: bubbletea-based terminal app with component structure
(views/, models/, commands/), keyboard-driven navigation, log viewer,
help overlay, and `Taskfile.yaml`. Uses `pavona/tui` for the component
model, layout primitives, and keybinding management.

### 4. Full-Stack Web App

```
pavona new app acmecorp
```

Generates: HTTP server, static file serving, templ rendering (via
a-h/templ), HTMX + Alpine.js wiring, SQLite (with sqlc), CSS pipeline
(Tailwind via daisyui), Dockerfile, and `Taskfile.yaml`.

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
| **Testing**    | `test` package gives you: temp DB, golden file comparison, request recording, clock mocking.         | The stuff you rewrite for every project.                                     |
| **Build**      | Single `go build`, no codegen step at runtime.                                                       | Fast iteration, simple debugging.                                            |
| **Scaffold**   | Codegen produces files once — you own them after.                                                    | No framework lock-in. You can delete `pavona` from go.mod and still compile. |

## What Makes It Beautiful

1. **No init() registration.** Pavona uses explicit `pavona.Register[T]()` calls in your `main.go` or module `init.go`. Static analysis can trace every registered type.

2. **Graceful shutdown is a first-class citizen.** Every module declares its shutdown dependencies. Pavona drains connections, finishes in-flight work, then exits. No global state cleanup needed.

3. **The "peel away" principle.** Start with `pavona new app` — you get everything. As you grow, you replace Pavona components one by one with your own. The framework doesn't fight you; it cedes control gracefully.

4. **The scaffold is the docs.** Every generated project includes a `docs/` directory with architecture decisions, a `README.md` that explains the layout, and inline comments on the generated files. You learn the patterns by reading the output, not a manual.

5. **Zero dependency in library mode.** If you `pavona new lib`, your `go.mod` only needs the standard library. Pavona's test helpers are a dev dependency via `_test.go` imports.

## Growth Path

```
pavona new app blog                     # generates: main.go, handlers/, db/, static/
cd blog && pavona add handler auth/login # adds handler + route + middleware
pavona add job cleanup                  # adds background worker
pavona add migration add_users_table    # generates migration SQL + sqlc query
pavona add cli admin:ban                # adds CLI subcommand to the same binary
pavona serve                            # runs dev server with hot reload (air)
pavona build                            # produces single binary: ./bin/blog
```

The same binary serves HTTP, exposes CLI commands, *and* can launch a TUI — all sharing the same domain code.

## Summary

Pavona is not the biggest or most opinionated framework. It is the
**most ergonomic** — designed to be picked up in 5 minutes and
outgrown gracefully. Like coral, it provides structure for life to
flourish, then becomes part of the reef.
