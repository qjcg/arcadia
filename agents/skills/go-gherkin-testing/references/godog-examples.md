---
description: "Complete, runnable reference projects from the godog repository. Each example demonstrates a real godog usage pattern. Download URL is the canonical source."
source: "https://github.com/cucumber/godog/tree/26931e66028d28bc7522af082fb55f1d57628ceb/_examples"
---

# godog examples reference

These are the official godog examples, downloaded from the godog repository.
Each is a complete, standalone Go project with its own feature files, step
definitions, and test suite. Study them to see real working patterns.

## Index

| Directory                      | What it demonstrates                                                           | When to study it                                                             |
|--------------------------------|--------------------------------------------------------------------------------|------------------------------------------------------------------------------|
| `godogs/`                      | Minimal complete project — one `.go` file, one test file, two `.feature` files | First time setting up godog. Shows the bare minimum wiring.                  |
| `api/`                         | HTTP API testing with `httptest` and JSON response assertions                  | Writing step definitions for HTTP servers. Includes a detailed README.       |
| `db/`                          | Database state management in BDD tests (PostgreSQL via `testcontainers-go`)    | Testing against a real database. Shows setup/teardown in hooks.              |
| `assert-godogs/`               | Using godog's built-in `godog` package assertions                              | Writing reusable assertion helpers. Simpler variant of the godogs example.   |
| `attachments/`                 | Attaching data (screenshots, logs) to godog output                             | Debugging failing scenarios in CI. Shows `Attachment` API.                   |
| `custom-formatter/`            | Custom formatter (emoji-based output)                                          | Advanced reporting needs. Shows the `Format` interface.                      |
| `incorrect-project-structure/` | A deliberately wrong project layout (feature files in the test package root)   | Teaching/troubleshooting. Shows what happens when godog can't find features. |

## How to study these

Each directory is a self-contained Go module. To run an example:

```bash
cd <example-dir>
go test
```

For the **api** and **db** examples, read the README first — they have extra
setup steps (starting servers, database containers).

For **godogs**, the simplest path is:
1. Read `features/godogs.feature` — see the Gherkin
2. Read `godogs_test.go` — see the step definitions and suite wire-up
3. Read `godogs.go` — see the production code being tested

## Key patterns visible across examples

- **Feature files live in `features/`**, step definitions and test suite live
  together in a test package at the project root (or `features/steps/` for
  larger projects)
- **`ScenarioInitializer`** is the standard way to register steps (godog v0.15+)
- **`ctx.Before`** resets state before each scenario (godogs, db, api all do this)
- **Build tags and hooks** manage expensive setup like database containers (db example)
- **Custom formatters** implement a `FormatterFunc` interface (custom-formatter)
- **Attachments** capture debugging data without changing assertion logic (attachments)
