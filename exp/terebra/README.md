# Terebra

A modern shell in Go — long, slender, drills deep.

*Named after the **Terebra** genus (auger shells).*

## Quick Start

```bash
# REPL mode
go run ./cmd/terebra

# Script mode
go run ./cmd/terebra script.trb

# Script with shebang
chmod +x script.trb
./script.trb

# Build a script to a standalone binary
go run ./cmd/terebra build script.trb myapp
./myapp
```

## Features

### Core Shell
- REPL with readline (history, search, editing, persistent history)
- Built-in commands: `cd`, `pwd`, `echo`, `ls`, `exit`, `help`, `type`, `which`
- External command execution via `$PATH`
- Pipes `|`, redirects `>`, `>>`, `<`, `2>`, `2>&1`
- Job control: `&`, `jobs`, `fg`, `bg`
- Variables: `$VAR`, `${VAR}`, `$?`, `$$`, local/exported
- Brace expansion: `{a,b,c}`, `{1..5}`, `{a..z..2}`, `{01..05}`
- Arithmetic expansion: `$((2 + 3))`, `$((x * 2))`
- Arrays: `arr=(1 2 3)`, `${arr[0]}`, `${arr[@]}`, `${#arr[@]}`
- String manipulation: `${var:0:5}`, `${var/old/new}`, `${var^^}`
- Tab completion: commands, built-ins, files
- Quoting: single, double, escape
- Comments (`#`)

### Scripting Language (`.trb`)
- `if`/`then`/`elif`/`else`/`fi` conditionals
- `for`/`in`/`do`/`done` loops
- `while`/`do`/`done` and `until`/`do`/`done` loops
- `function name {}` and `name() {}` function definitions
- `try`/`catch`/`end` error handling
- `source` to include script files
- Shebang support (`#!/usr/bin/env terebra`)

### CUE Integration
- `drill cue <file>` — evaluate, validate, and extract CUE values
- `|>` auger pipe — passes structured CUE values between commands
- `state` — export shell state as CUE
- `drill fs <path>` — filesystem metadata as CUE
- `drill proc <pid>` — process information as CUE
- `drill net <host>` — network inspection as CUE

### Plugin System
- Load Go `.so` plugins at runtime
- `plugin load <path>` — load a plugin
- `plugin list` — list loaded plugins
- Plugin search path: `~/.terebra/plugins/`, `/usr/lib/terebra/plugins/`, `$TEREBRA_PLUGIN_PATH`

### Script Compilation
- `terebra build <script.trb> [output]` — compile `.trb` scripts to standalone Go binaries

## Project Structure

```
cmd/terebra/
  main.go            # Entry point (REPL, script, build, version)
  build.go           # Script compilation
internal/
  shell/             # REPL loop + command execution
  parser/            # Shell syntax parser (lexer, parser, AST)
  expand/            # Word expansion (brace expansion)
  builtins/          # Built-in command handlers
  script/            # Scripting language (parser, interpreter, AST)
  drill/             # Drill subsystem (cue, fs, proc, net)
  cueutil/           # CUE helpers
  plugin/            # Plugin loading
```

## Build

```bash
go build -o terebra ./cmd/terebra
```

## Test

```bash
go test ./...
```
