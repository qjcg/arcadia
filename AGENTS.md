# Agents

## Build

- Use `task` via `github.com/go-task/task/v3/cmd/task`
  - When needed, create a `Taskfile.yaml`
  - `task` fully replaces `make` and `Makefiles`, NEVER use those.
- Where hot reload is needed, use `air` via `github.com/air-verse/air`
  - Use hot reload with:
    - Web UIs
	- Games
	- Backend servers
  - When using air, call it via task, as in `task dev`, which might call `go tool air`

## CI

- use `lefthook` for git hooks
  - for hooks that involve running multi-line shell commands, those shell commands should live in a separate directory
- lint via `golangci-lint`, called via `go run`

## Git

- Always commit your work as you go, in feature branches off of `main`
- Use [conventional commits](https://www.conventionalcommits.org/en/v1.0.0/)

## Docs

- Write docs in markdown format
- Put docs in a `docs` directory, except for the README.md
- The README.md should be concise, clear and useful.

## Backend

- Always use Go as the backend language
- db
  - Use sqlite via `github.com/ncruces/go-sqlite3`
  - Use `sqlc` via `github.com/sqlc-dev/sqlc/cmd/sqlc`
  - Use `migrate` via `github.com/golang-migrate/migrate/v4/cmd/migrate` for db migrations
  - Always embed migrations into the go binary
- authN
  - Use passkeys via `github.com/go-webauthn/webauthn`
- authZ
  - use `casbin` via `github.com/casbin/casbin/v3`
- For quick and dirty scripts, use https://github.com/bitfield/script

## Frontend

- Templ
- templUI
- HTMX
- Alpine.js
- tailwindcss

## Games

- Use ebiten via `github.com/hajimehoshi/ebiten/v2`
  - Build games as [WebAssembly](https://ebitengine.org/en/documents/webassembly.html) (Option 1. WasmServe)
	- Once the webassembly is built, serve it for browsers via `go run github.com/hajimehoshi/wasmserve@latest ./path/to/yourgame`

## CLI tools

- Use [cobra](https://github.com/spf13/cobra)
- Use [viper](https://github.com/spf13/viper)

### Testscript
- Write tests for CLI tools with [testscript](https://pkg.go.dev/github.com/rogpeppe/go-internal/testscript)
- When using testscript, your `main_test.go` file should follow the approach below (substituting "sv" for the name of the CLI tool in question):

```go
package main

import (
	"testing"

	"github.com/rogpeppe/go-internal/testscript"
)

func TestMain(m *testing.M) {
	testscript.Main(m, map[string]func(){
		"sv": main,
	})
}

func TestCLI(t *testing.T) {
	testscript.Run(t, testscript.Params{
		Dir: "testdata",
	})
}
```

## Text User Interface (TUI)

- Use `github.com/charmbracelet/bubbletea` (TUI framework)

## AI Agent Software

- For building AI Agent services, use `https://github.com/charmbracelet/fantasy`

## System architecture

- Prefer a "modular monolith" architecture to microservices

## Code organization

- Write clean, modular code.
- Generally go packages should be created under an `internal` subdirectory unless the library code is intended for export

## Issue Tracking

- Use `go tool bd` to manage tasks and implementation plans
- When creating work items, use these flags for maximum clarity:
  - `--type`: (bug|feature|task|epic|chore)
  - `--description`: Detailed explanation of the task
  - `--acceptance`: Clear criteria for when the task is considered done
  - `--deps`: Comma-separated dependencies (e.g., `sv-123`)
  - `--design`: Any critical technical design notes
  - `--estimate`: Time estimate in minutes. Estimate time for an agent such as yourself, NOT a human.
  - `--priority`: Priority level (P0-P4)
  - `--labels`: Comma-separated labels
  - `--parent`: Link to a parent epic or task for hierarchy
- Use `bd list` or `bd ready` to check current queue

## Testing

When writing tests, adhere to the following advice from the book Unit Testing by Vladimir Khorikov.

Focus on tests that align with the Four Pillars of a Valuable Test, and heed his guidance on What Not to Test.

### Four Pillars of a Valuable Test

Khorikov defines four key attributes for measuring the value of a unit test:

#### 1. Protection against Regressions
- Evaluates how effectively a test finds bugs
- Considers code complexity, coverage, and domain significance
- Focuses on catching meaningful business logic issues

#### 2. Resistance to Refactoring
- Determines whether a test can survive code changes without breaking
- Measures the coupling between test and implementation
- Aims to minimize false positives
- Considered a binary attribute: either a test resists refactoring or it doesn't

#### 3. Fast Feedback
- Measures how quickly a test can be executed
- Depends on test dependencies and size
- Balances thoroughness with performance

#### 4. Maintainability
- Assesses how easy the test is to set up and understand
- Considers test readability and complexity of dependencies

### What Not to Test

#### Avoid Testing Implementation Details

Khorikov emphasizes that tests should not focus on implementation details, but instead should verify:

- Observable behavior
- End results that a business person can understand
- Meaningful units of behavior

#### Don't Test Code That Doesn't Contribute to Observable Behavior

Khorikov suggests avoiding tests for code that doesn't:

- Expose an operation that helps the client achieve a goal
- Expose a state that helps the client achieve a goal

#### Steer Clear of Overly Fragile Tests

Avoid tests that:

- Break with every small refactoring
- Are tightly coupled to implementation
- Cannot be traced back to a business requirement

#### Minimize Excessive Mocking

Khorikov recommends limiting mock usage to:

- Shared, out-of-process dependencies
- External services

#### Red Flags for Tests to Avoid

Don't test code that:

- Doesn't have a clear connection to business requirements
- Is overly complex and highly collaborative
- Requires extensive mocking to test

#### Principles for Test Elimination

If a test:

- Generates frequent false alarms
- Is difficult to maintain
- Doesn't provide clear value in catching potential bugs

Then you should consider refactoring or deleting it.
