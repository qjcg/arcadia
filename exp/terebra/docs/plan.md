# Implementation Plan

## Phase 1: Core Shell Gaps (foundation) ✅

| # | Task | Status |
|---|------|--------|
| 1.1 | **Globbing** — `*`, `?`, `[...]`, `**` matching via `filepath.Match` + `**` recursive walk | ✅ |
| 1.2 | **Command chaining** — `&&`, `||`, `;` token types + parser + executor loop | ✅ |
| 1.3 | **Heredocs** — `<<EOF`, `<<-EOF`, `<<'EOF'` token + redirect type + lexer accumulation | ✅ |
| 1.4 | **`alias` builtin** — `alias [name=value]`, `unalias`, expansion during execution | ✅ |
| 1.5 | **`history` builtin** — `history [n]`, `history -c`, `history -d n` wrapping readline | ✅ |
| 1.6 | **Backtick command substitution** — `` `cmd` `` in lexer → execute + capture stdout | ✅ |
| 1.7 | **Read-only variables** — `readonly`/`declare -r` flag on vars, reject set attempts | ✅ |

## Phase 2: CUE Expansion ✅

| # | Task | Status |
|---|------|--------|
| 2.1 | **CUE pipe encoders** — `|>json`, `|>yaml`, `|>cue` as pipe endpoint modifiers | ✅ |
| 2.2 | **CUE unification pipe** — `|>` then `&` to merge two CUE streams | ⏳ (future) |
| 2.3 | **CUE pipe filters** — `select`, `where`, `sort`, `group`, `join` as filter commands | ⏳ (future) |
| 2.4 | **`cue` subcommand group** — `eval`, `vet`, `export`, `def`, `fmt`, `trim` as builtins | ✅ |
| 2.5 | **`drill cue --unify`** flag | ✅ |

## Phase 3: Scripting & State ✅

| # | Task | Status |
|---|------|--------|
| 3.1 | **CUE-typed variables** — `let name: type = value` syntax in script parser + interpreter | ⏳ (future) |
| 3.2 | **Self-validating state schema** — embedded CUE schema for shell state, validate on load | ✅ |
| 3.3 | **Session checkpointing** — `state save <name>` / `state load <name>` | ✅ |
| 3.4 | **Plugin extension points** — plugins register completions, prompt funcs, CUE codecs | ✅ |

## Phase 4: Developer Experience ✅

| # | Task | Status |
|---|------|--------|
| 4.1 | **Error suggestions** — Levenshtein-based "did you mean `drill proc`?" | ✅ |
| 4.2 | **`--explain` flag** — dry-run mode showing what a command will do | ✅ |
| 4.3 | **Debug mode** — `set -x` tracing of commands before execution | ✅ |
| 4.4 | **Rich help** — help with examples per command, not flat text dump | ✅ |
| 4.5 | **Prompt customization** — `$PS1` env var with Go template support | ✅ |
| 4.6 | **Vi-mode** — `set -o vi` / `set -o emacs` toggle | ✅ |

## Phase 5: TUI ✅

| # | Task | Status |
|---|------|--------|
| 5.1 | **Syntax-highlighted input** — token-based coloring in readline | ⏳ (future) |
| 5.2 | **Fuzzy command search** — Ctrl+R with preview via readline API | ⏳ (future) |
| 5.3 | **Interactive drill viewer** — expand/collapse, scroll for CUE drill output | ⏳ (future) |
| 5.4 | **Progress display** — spinner/progress bar for long-running commands | ⏳ (future) |

## Phase 6: Consolidation

| # | Task | Status |
|---|------|--------|
| 6.1 | **Align file structure with spec** — create `pipeline.go`, `value.go`, `pipe.go`, `schema.go`, `complete.go`, `prompt.go`, `state.go` as shim/reorg | ⏳ (future) |
| 6.2 | **End-to-end tests** — integration tests for pipes, chains, scripts, CUE pipeline | ⏳ (future) |
| 6.3 | **Final polish** — edge cases, error messages, performance | ⏳ (future) |

