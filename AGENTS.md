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

- Use `cobra` via `github.com/spf13/cobra`
- Use `viper` via `github.com/spf13/viper`
- Write tests for CLI tools with [testscript](https://pkg.go.dev/github.com/rogpeppe/go-internal/testscript)

## Text User Interface (TUI)

- Use `github.com/charmbracelet/bubbletea` (TUI framework)

## AI Agent Software

- For building AI Agent services, use `https://github.com/charmbracelet/fantasy`

## System architecture

- Prefer a "modular monolith" architecture to microservices

## Code organization

- Write clean, modular code.
- Generally go packages should be created under an `internal` subdirectory unless the library code is intended for export
