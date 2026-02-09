# Agents

## Foundations

### System Architecture
- Prefer a "modular monolith" architecture over microservices.

### Code Organization
- Write clean, modular code.
- Organize Go packages under an `internal` subdirectory unless the library code is intended for export.

### Git
- Commit work incrementally in feature branches off `main`.
- Follow [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/).

### Documentation
- Write documentation in Markdown format.
- Store documentation in a `docs/` directory, except for `README.md`.
- Maintain a concise, clear, and useful `README.md`.

## Workflow

### Issue Tracking
- ALWAYS use `go tool bd` to manage tasks and implementation plans.
- Apply these flags when creating work items for maximum clarity:
  - `--type`: (bug|feature|task|epic|chore)
  - `--description`: Detailed explanation of the task
  - `--acceptance`: Clear criteria for when the task is considered done
  - `--deps`: Comma-separated dependencies (e.g., `sv-123`)
  - `--design`: Any critical technical design notes
  - `--estimate`: Time estimate in minutes. Estimate time for an agent such as yourself, NOT a human.
  - `--priority`: Priority level (P0-P4)
  - `--labels`: Comma-separated labels
  - `--parent`: Link to a parent epic or task for hierarchy
- Use `bd list` or `bd ready` to check the current queue.

### Build & Development
- Use `task` via `github.com/go-task/task/v3/cmd/task`.
  - Create a `Taskfile.yaml` when needed.
  - Replace `make` and `Makefiles` entirely with `task`.
- Use `air` via `github.com/air-verse/air` for hot reloading.
  - Apply hot reload to Web UIs, Games, and Backend servers.
  - Trigger `air` via `task` (e.g., `task dev` calling `go tool air`).

### CI/CD
- Use `github.com/evilmartians/lefthook` for git hooks.
  - Store multi-line shell commands for hooks in a separate directory.
- Lint using `github.com/golangci-lint/golangci-lint` via `go run`.

## Tech Stacks

### Backend
- Build backends exclusively with Go.
- Database:
  - Use `github.com/ncruces/go-sqlite3` for SQLite.
  - Use `github.com/sqlc-dev/sqlc/cmd/sqlc` for type-safe SQL.
  - Use `github.com/golang-migrate/migrate/v4/cmd/migrate` for migrations.
  - Embed migrations into the Go binary.
- Authentication:
  - Use `github.com/go-webauthn/webauthn` for passkeys.
- Authorization:
  - Use `github.com/casbin/casbin/v3` for RBAC/ABAC.
- Scripting:
  - Use `github.com/bitfield/script` for rapid shell-like scripting.

### Frontend
- Use `github.com/a-h/templ` for server-side templates.
- Use `HTMX` via `htmx.org` for dynamic interactions.
- Use `Alpine.js` via `alpinejs.dev` for client-side state.
- Use `tailwindcss` for styling.
- Use `daisyui` for components.

## Project Types

### CLI Tools
- Use `github.com/spf13/cobra` for command structure.
- Use `github.com/spf13/viper` for configuration.
- Test CLI tools with `github.com/rogpeppe/go-internal/testscript`.
- Follow this `main_test.go` pattern for `testscript`:

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

### TUI (Text User Interface)
- Use `github.com/charmbracelet/bubbletea` as the TUI framework.

### AI Agent Software
- Use `github.com/charmbracelet/fantasy` for building AI Agent services.

### Games
- Use `github.com/hajimehoshi/ebiten/v2` for game development.
- Target WebAssembly (WasmServe option).
- Serve for browsers via `go run github.com/hajimehoshi/wasmserve@latest ./path/to/game`.

## Quality

### Testing

When writing tests, adhere to the following advice from the book Unit Testing by Vladimir Khorikov.

Focus on tests that align with the Four Pillars of a Valuable Test, and heed his guidance on What Not to Test.

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
