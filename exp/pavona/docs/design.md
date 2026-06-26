# Pavona

> A Go framework that grows with you — from CLI tools to full-stack web apps.

Named after leaf coral of the *Pavona* genus: layered, branching, and symbiotic. You start with a single `main.go` and extend outward in predictable patterns.

## Philosophy

**"A framework that gets out of your way the moment you know what you're doing."**

CLI tools, libraries, and web apps share 80% of the same DNA — config loading, dependency wiring, logging, testing infrastructure, command routing, middleware patterns. Pavona provides that 80% as modular, composable layers, then stays silent for the other 20%.

## Core Architecture

```
pavona/
├── cmd/           # Pavona CLI — scaffolding, dev server, codegen
├── app/           # Runtime kernel — bootstrap, lifecycle, shutdown
├── wire/          # Dependency injection & module registration
├── conf/          # Config loading (YAML/TOML/JSON/env, layered)
├── log/           # Structured logging (zerolog-based, level-filtered)
├── serve/         # HTTP server (net/http, middleware stack, graceful shutdown)
├── cli/           # CLI framework (cobra-like, top-level subcommands)
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

`App` is not a framework — it's a **lifecycle manager**. Modules register start/stop hooks, health checks, and depend on each other. Pavona resolves the DAG and starts/stops in order.

## Three Project Types

### 1. CLI Tool

```
pavona new tool gh-deploy
```

Generates: single binary with subcommands (`deploy`, `rollback`, `status`, `config`). Uses `pavona/cli` for command routing, `pavona/conf` for config, and `pavona/pool` for concurrent API calls.

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

Generates: minimal Go module with `pavona/test` helpers, CI setup, and example-driven docs. The library uses zero Pavona runtime dependencies in production — only the dev toolchain.

### 3. Full-Stack Web App

```
pavona new app acmecorp
```

Generates: HTTP server, static file serving, templ rendering (via a-h/templ), HTMX + Alpine.js wiring, SQLite (with sqlc), CSS pipeline (Tailwind via daisyui), Dockerfile, and `Taskfile.yaml`.

## Key Design Decisions

| Area           | Decision                                                                                             | Rationale                                                                    |
|----------------|------------------------------------------------------------------------------------------------------|------------------------------------------------------------------------------|
| **DI**         | No reflection. Modules self-register via `pavona.Register[T]`.                                       | Compile-time safety, no magic.                                               |
| **Config**     | `conf` merges: defaults → file → env → flags.                                                        | The Unix philosophy — layers override cleanly.                               |
| **DB**         | Pavona provides migration runner and connection lifecycle. Actual queries use **sqlc**.              | Type-safe SQL beats any ORM.                                                 |
| **Middleware** | Standard `net/http` middleware. Pavona provides `recover`, `request-id`, `access-log`, `rate-limit`. | No custom handler signature. Compatible with everything.                     |
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

The same binary serves HTTP *and* exposes CLI commands — useful for admin operations, data migrations, or cron jobs that share the same domain code.

## Summary

Pavona is not the biggest or most opinionated framework. It is the **most ergonomic** — designed to be picked up in 5 minutes and outgrown gracefully. Like coral, it provides structure for life to flourish, then becomes part of the reef.
