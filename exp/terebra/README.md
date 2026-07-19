# Terebra

A modern shell in Go — long, slender, drills deep.

*Named after the **Terebra** genus (auger shells).*

## Quick Start

```bash
# REPL mode
go run .

# Inline script
go run . -c 'echo hello'

# Script mode
go run . script.trb

# Script with shebang
chmod +x script.trb
./script.trb

# Build a script to a standalone binary
go run . build script.trb myapp
./myapp
```

## Features

### Core Shell
- REPL with readline (history, search, editing, persistent history, fuzzy Ctrl+R search)
- Built-in commands: `cd`, `pwd`, `echo`, `exit`, `help`, `type`, `which`,
  `export`, `unset`, `set`, `readonly`, `alias`, `unalias`, `history`, `source`,
  `jobs`, `fg`, `bg`, `drill`, `cue`, `plugin`, `state`
- External command execution via `$PATH`
- Pipes `|`, redirects `>`, `>>`, `<`, `2>`, `2>>`, `2>&1`, heredocs `<<`, `<<-`
- Command chaining: `&&`, `||`, `;`
- Job control: `&`, `jobs`, `fg`, `bg`
- Variables: `$VAR`, `${VAR}`, `$?`, `$$`, local/exported/read-only
- Brace expansion: `{a,b,c}`, `{1..5}`, `{a..z..2}`, `{01..05}`, nested, cartesian
- Globbing: `*`, `?`, `[...]`, `**` (recursive)
- Arithmetic expansion: `$((2 + 3))`, `$((x * 2))`
- Arrays: `arr=(1 2 3)`, `${arr[0]}`, `${arr[@]}`, `${#arr[@]}`
- Associative arrays: `declare -A map`, `map[key]=val`, `${map[key]}`, `${!map[@]}`
- String manipulation: `${var:0:5}`, `${var/old/new}`, `${var^^}`, `${var,}`,
  `${var#pat}`, `${var%pat}`, `${#var}`
- Command substitution: `$(cmd)`, `` `cmd` ``
- Tab completion: commands, built-ins, files, PATH
- Quoting: single, double, escape, backtick
- Debug mode: `set -x` / `set +x`
- Vi/Emacs keybindings: `set -o vi` / `set -o emacs`
- Prompt customization: `PS1` with `{{.}}`, `{{exit}}`, `{{exitcode}}` templates
- `--explain` dry-run flag
- `~/.terebrarc` startup script

### Scripting Language (`.trb`)
- `if`/`then`/`elif`/`else`/`fi` conditionals
- `for`/`in`/`do`/`done` loops
- `while`/`do`/`done` and `until`/`do`/`done` loops
- `function name {}` and `name() {}` function definitions
- `try`/`catch`/`end` error handling
- `source` to include script files
- Shebang support (`#!/usr/bin/env terebra`)

### CUE Integration
- `drill cue <file>` — evaluate, validate, extract, unify CUE values
- `drill fs <path>` — filesystem metadata as CUE
- `drill proc <pid>` — process information as CUE
- `drill net <host>` — network inspection as CUE
- `|>` auger pipe — passes structured CUE values between commands
- `|>json`, `|>yaml`, `|>cue` — encode CUE output as JSON/YAML/CUE
- `cue eval`, `vet`, `export`, `def`, `fmt`, `trim` — CUE operations
- `state` — export/import shell state as CUE with save/load checkpointing

### Plugin System
- Load Go `.so` plugins at runtime
- `plugin load <path>` — load a plugin
- `plugin list` — list loaded plugins
- Plugin search path: `~/.terebra/plugins/`, `/usr/lib/terebra/plugins/`, `$TEREBRA_PLUGIN_PATH`
- Optional interfaces: `Completer`, `PromptFunc`, `CUEEncoder`, `CUEDecoder`

### Script Compilation
- `terebra build <script.trb> [output]` — compile `.trb` scripts to standalone Go binaries

## Project Structure

```
main.go              # Entry point (REPL, script, build, version, -c)
build.go             # Script compilation
main_test.go         # Testscript integration tests
testdata/            # 16 test suites
internal/
  shell/             # REPL loop + command execution
  parser/            # Shell syntax parser (lexer, parser, AST)
  expand/            # Word expansion (brace, var, glob)
  builtins/          # Built-in command handlers
  script/            # Scripting language (parser, interpreter, AST)
  drill/             # Drill subsystem (cue, fs, proc, net)
  cueutil/           # CUE helpers
  plugin/            # Plugin loading
```

## Build

```bash
go build -o terebra .
```

## Test

```bash
go test ./...
go test . -run TestTerebra -v
```
