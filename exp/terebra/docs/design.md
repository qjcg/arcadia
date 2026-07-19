# Terebra — A Modern Shell in Go

*Named after the **Terebra** genus (auger shells) — long, slender, drills deep.*

## Design Philosophy

| Principle                 | Meaning                                                            |
|---------------------------|--------------------------------------------------------------------|
| **Drills deep**           | Powerful built-in data inspection (files, processes, network, CUE) |
| **Slim profile**          | Minimal, fast startup, sensible defaults                           |
| **Go-native**             | Extensible via Go, not shell scripts                               |
| **Structured by default** | Commands output structured data, not just text                     |

## Feature Set

### Core Shell

- REPL with `chzyer/readline` (history, search, editing, persistent history)
- Built-in commands: `cd`, `pwd`, `echo`, `ls`, `source`, `export`, `alias`, `type`, `which`, `exit`, `help`, `history`
- External command execution via `$PATH`
- Pipes `|`, redirects `>`, `>>`, `<`, `2>`, `2>&1`, heredocs
- Job control: `&`, `jobs`, `fg`, `bg`
- Variables: `$VAR`, `${VAR}`, local/exported/read-only
- Quoting: single, double, backtick, escape
- **Brace expansion** — `{a,b,c}` lists, `{1..5}` ranges, `{01..05}` zero-padded, `{a..z..2}` steps, reverse ranges, nesting, cartesian product of adjacent groups
- Globbing: `*`, `?`, `[...]`, `**`
- Command chaining: `&&`, `||`, `;`
- Exit codes (`$?`)
- Tab completion: files, commands, built-ins, flags, PATH
- Script execution: `terebra script.trb` or `./script.trb`
- Prompt customization via Go template + env vars
- Vi-mode keybindings (insert/normal)

#### Expansion Pipeline

Word expansion happens in a fixed order after parsing, before execution:

```
input → parse → brace expand → variable expand → glob → execute
```

1. **Brace expansion** — `{a,b,c}`, `{1..5}` generate multiple words
2. **Variable expansion** — `$VAR`, `${VAR}`, `$?`, `$$` replaced with values
3. **Globbing** — `*`, `?`, `[...]` patterns replaced with matching filenames

This order ensures `${...}` is never confused with `{...}` and that glob patterns in expanded variables are still expanded.

### Brace Expansion

Brace expansion generates multiple words from `{...}` patterns before variable expansion and globbing.

| Pattern           | Example        | Result              |
|-------------------|----------------|---------------------|
| Comma list        | `{a,b,c}`      | `a b c`             |
| Numeric range     | `{1..5}`       | `1 2 3 4 5`         |
| Alpha range       | `{a..f}`       | `a b c d e f`       |
| Step              | `{1..10..3}`   | `1 4 7 10`          |
| Zero-padded       | `{01..05}`     | `01 02 03 04 05`    |
| Reverse           | `{z..u}`       | `z y x w v u`       |
| Prefix/suffix     | `pre{1,2}post` | `pre1post pre2post` |
| Cartesian product | `{a,b}{1,2}`   | `a1 a2 b1 b2`       |
| Nested            | `{a,{b,c}}`    | `a b c`             |

**Rules:**
- `{}` with no comma or range → literal `{}`
- Escaped `\{` or quoted `'{'` → no expansion
- Invalid ranges like `{1..a}` or overflow → literal
- Unclosed `{` → literal

**Implementation:** `internal/expand/` package. The `Expand(word string) []string` function parses the word into a tree of brace groups and expands it into zero or more words.

CUE's lattice type system (types + values unified) is the internal currency of the structured pipeline.

- `|>` auger pipe — passes CUE values between commands, preserving types and constraints
- Automatic decode: `ls |>.name` extracts the `name` field from each CUE value
- CUE unification: `cmd1 |>.ports | cmd2 |>.ports |> &` merges port configs with validation
- Encoders: `|>json`, `|>yaml`, `|>cue` at pipe endpoints
- Built-in filters: `select`, `where`, `sort`, `group`, `join` operate on CUE values

### "Drill" Commands (signature feature)

- `drill fs <path>` — drill into filesystem metadata, disk usage, file types (outputs CUE)
- `drill proc <pid>` — drill into processes, open files, env, memory (outputs CUE)
- `drill net <host:port>` — drill into network connections, TLS, HTTP headers (outputs CUE)
- `drill cue <file>` — drill into CUE files: evaluate, validate, walk the value tree
  - `drill cue -e 'app.ports'` — extract a path
  - `drill cue -v` — validate against schema
  - `drill cue --export json` — evaluate and export as JSON
  - `drill cue --unify extra.cue` — unify with another file

### Built-in CUE Commands

- `cue eval <file>` — evaluate and print CUE
- `cue vet <data> <schema>` — validate data against schema
- `cue export <file>` — evaluate to concrete JSON/YAML
- `cue def <file>` — print consolidated definition
- `cue fmt <file>` — format CUE
- `cue trim <file>` — remove redundant values
- All work as pipe filters: `drill fs . |> cue vet ./fs-schema.cue`

### Shell State Management

- `terebra state` — export/import shell state as **CUE** (not plain JSON)
- State schema is self-validating — loading a corrupted state file fails gracefully
- Session checkpointing — save and restore shell sessions
- Remote state — share shell state across machines

### Scripting Language (`.trb`)

- Functions with local scope
- `if`/`elif`/`else`, `for`, `while`, `until`
- Arrays (indexed + associative)
- Arithmetic expansion `$((expr))`
- String manipulation (slice, replace, regex)
- `try`/`catch` error handling
- **CUE-typed variables**: `let port: int & >1024 = 8080` (CUE constraint on a shell variable)
- Shebang support (`#!/usr/bin/env terebra`)
- Compile `.trb` → standalone Go binary

### Plugin System

- Load Go plugins (`.so`) at runtime
- Register new built-ins, completions, prompt functions, **CUE encoders/decoders**
- Plugin discovery via `$TEREBRA_PLUGIN_PATH`

### TUI Enhancements

- Syntax-highlighted input line
- Fuzzy command search (Ctrl+R with preview)
- Interactive `drill` output viewer (expand/collapse, scroll)
- Progress display for long-running commands

### Developer Experience

- `help` with examples, not just man-page style
- Errors include suggestions ("did you mean `drill proc`?")
- `--explain` flag on any command shows what it will do
- Debug mode (`set -x` equivalent)

## Architecture

```
cmd/terebra/
├── main.go
├── internal/
│   ├── shell/
│   │   ├── repl.go
│   │   ├── executor.go
│   │   ├── pipeline.go          # | and |> (auger) pipe handling
│   │   └── job.go
│   ├── parser/
│   │   ├── lexer.go
│   │   ├── parser.go
│   │   └── ast.go
│   ├── expand/                   # Word expansion pipeline
│   │   ├── expand.go             # Pipeline: brace → var → glob
│   │   ├── brace.go              # Brace expansion: {a,b}, {1..5}
│   │   └── brace_test.go
│   ├── builtins/
│   ├── drill/                   # "drill" subsystem
│   │   ├── drill.go
│   │   ├── fs.go
│   │   ├── proc.go
│   │   ├── net.go
│   │   └── cue.go               # drill cue — evaluate, validate, walk
│   ├── cueutil/                 # CUE helpers (shared across packages)
│   │   ├── value.go             # cue.Value helpers (lookup, format, unify)
│   │   ├── pipe.go              # CUE pipe encoding/decoding
│   │   └── schema.go            # Embedded shell state schema
│   ├── complete/
│   │   └── complete.go
│   ├── prompt/
│   │   └── prompt.go
│   ├── state/
│   │   └── state.go             # Shell state serialized as CUE
│   ├── scripting/
│   │   └── interpreter.go       # .trb interpreter with CUE-typed vars
│   └── plugin/
│       └── plugin.go
├── go.mod
├── go.sum
├── Taskfile.yaml
├── CHANGELOG.md
├── README.md
└── docs/
    ├── design.md
    ├── SPEC.md
    └── GUIDE.md
```

## Implementation Order

| Phase | What                                                           |
|-------|----------------------------------------------------------------|
| **1** | REPL + built-ins + external commands + `\|` pipes + redirects  | ✅ |
| **2** | Tab completion + history + job control + variables             | ✅ |
| **3** | CUE integration: `drill cue`, `\|>`, CUE state, CUE-typed vars | ✅ |
| **4** | Scripting language + functions + control flow                  | ✅ |
| **5** | Remaining drill commands (fs, proc, net) + plugins + TUI       | ✅ |
| **6** | `.trb` → Go compilation + polish                               | ✅ |
