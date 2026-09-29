# Bash vs Terebra

A feature-by-feature comparison of **bash** and **terebra** based on what is
actually implemented in this checkout. Terebra's `docs/design.md` describes a
broader vision (including some features marked "planned"); this document only
covers behavior that exists in the code today and is exercised by
`testdata/*.txtar` or the package unit tests.

Terebra is a shell written in Go. It deliberately reuses bash's syntax for the
common cases, so most one-liners are portable, but it is **not** a bash clone:
it omits a large amount of bash builtins, script syntax, and expansions, and it
adds a structured-data layer (CUE) that bash has no equivalent for.

## At a glance

| Area                                                                | Bash                      | Terebra                 |
|---------------------------------------------------------------------|---------------------------|-------------------------|
| Language                                                            | C                         | Go                      |
| Execution of external commands via `PATH`                           | Yes                       | Yes                     |
| Pipes `\|`                                                          | Yes                       | Yes                     |
| `\|&` (pipe stdout+stderr)                                          | Yes                       | Yes                     |
| Redirect `>`, `>>`, `<`, `2>`, `2>>`, `2>&1`                        | Yes                       | Yes                     |
| Redirect `&>`, `&>>`                                                | Yes (bash 4+)             | Yes                     |
| Heredoc `<<`, `<<-`, here-string `<<<`                              | Yes                       | Yes                     |
| Arbitrary fd redirection (`3>`, `n>&m`)                             | Yes                       | No (only fd 2)          |
| Chaining `;`, `&&`, `\|\|`                                          | Yes                       | Yes                     |
| Background `&`, `jobs`, `fg`, `bg`                                  | Full job control          | Basic (see below)       |
| Variables `$VAR`, `${VAR}`, `$?`, `$$`                              | Yes                       | Yes                     |
| Positional params `$0..$9`, `$@`, `$#`, `$*`                        | Yes                       | No                      |
| Default/alternate expansions `${v:-}`, `${v:=}`, `${v:+}`, `${v:?}` | Yes                       | No                      |
| Arrays (indexed + associative)                                      | Yes                       | Yes (subset)            |
| Arithmetic `$(( ))`                                                 | Full C-like               | Integer subset          |
| Command substitution `$( )`, `` ` ` ``                              | Yes                       | Yes                     |
| Brace / tilde / glob expansion                                      | Yes                       | Yes                     |
| Recursive glob `**`                                                 | `globstar` shopt          | Yes (always on)         |
| Quoting `'…'`, `"…"`, `\`                                           | Yes                       | Yes                     |
| ANSI-C quoting `$'…'`                                               | Yes                       | No                      |
| Process substitution `<( )`, `>( )`                                 | Yes                       | No                      |
| `[[ ]]`, `(( ))`, `case`, `select`                                  | Yes                       | No                      |
| Script control flow (`if`, `for`, `while`, functions, `try`)        | Yes                       | Yes (own `.trb` syntax) |
| User-defined function params / locals / `return`                    | Yes                       | No                      |
| History expansion (`!!`, `!$`)                                      | Yes                       | No                      |
| Programmable completion                                             | Yes                       | No (fixed strategy)     |
| Startup file                                                        | `~/.bashrc`, `~/.profile` | `~/.terebrarc`          |
| Structured data (CUE) integration                                   | No                        | Yes                     |
| Compile scripts to standalone binaries                              | No                        | Yes                     |

## Invocation and modes

| Mode                  | Bash                  | Terebra                                 |
|-----------------------|-----------------------|-----------------------------------------|
| Interactive REPL      | `bash`                | `terebra`                               |
| Run a script          | `bash script.sh`      | `terebra script.trb`                    |
| Inline command string | `bash -c '…'`         | `terebra -c '…'`                        |
| Shebang               | `#!/usr/bin/env bash` | `#!/usr/bin/env terebra`                |
| Dry-run               | `bash -x` (trace)     | `terebra --explain <cmd>` (per command) |
| Compile to binary     | No                    | `terebra build script.trb myapp`        |
| Version               | `bash --version`      | `terebra --version` / `-v`              |

`--explain` emits `# would execute: …` instead of running the command. Terebra's
debug trace (`set -x`) is the closest analogue to `bash -x`.

## Execution, operators, and redirection

Supported operators (verified in `internal/parser/lexer.go` and
`testdata/chaining.txtar`, `pipes.txtar`, `errredir.txtar`,
`bothredir.txtar`):

- `cmd1 | cmd2` — pipe stdout.
- `cmd1 |& cmd2` — pipe stdout **and** stderr (bash 4+ syntax).
- `cmd1 ; cmd2` — sequential.
- `cmd1 && cmd2`, `cmd1 || cmd2` — short-circuit.
- `cmd &` — background (not supported inside pipelines, see job control).

Redirections:

| Form           | Meaning                     | Terebra                                         |
|----------------|-----------------------------|-------------------------------------------------|
| `> f`          | stdout truncate             | Yes                                             |
| `>> f`         | stdout append               | Yes                                             |
| `< f`          | stdin                       | Yes                                             |
| `2> f`         | stderr truncate             | Yes                                             |
| `2>> f`        | stderr append               | Yes                                             |
| `2>&1`         | stderr to stdout            | Yes                                             |
| `&> f`         | stdout+stderr truncate      | Yes                                             |
| `&>> f`        | stdout+stderr append        | Yes                                             |
| `<< EOF`       | heredoc                     | Yes (variable-expanded unless delimiter quoted) |
| `<<- EOF`      | heredoc, strip leading tabs | Yes                                             |
| `<<< word`     | here-string                 | Yes                                             |
| `3> f`, `n>&m` | other file descriptors      | No                                              |
| `>& f`         | bash shorthand for `&> f`   | No (use `&>`)                                   |

There is no `set -o noclobber` support, so `>` always truncates.

## Variables

| Feature                                      | Bash | Terebra                         |
|----------------------------------------------|------|---------------------------------|
| Assign `NAME=value`                          | Yes  | Yes                             |
| `export NAME` / `export NAME=value`          | Yes  | Yes                             |
| `readonly`                                   | Yes  | Yes                             |
| `unset`                                      | Yes  | Yes                             |
| Temp env for one command `FOO=b cmd`         | Yes  | Yes                             |
| `$?` last exit code                          | Yes  | Yes                             |
| `$$` shell PID                               | Yes  | Yes                             |
| `$0`, `$1`…`$9`, `$@`, `$#`, `$*`            | Yes  | **No**                          |
| `${v:-d}`, `${v:=d}`, `${v:+a}`, `${v:?err}` | Yes  | **No**                          |
| Indirect `${!var}`                           | Yes  | **No** (only `${!arr[@]}` keys) |
| `declare` / `typeset` / `local`              | Yes  | **No**                          |
| `IFS`, `$RANDOM`, `$SECONDS`, `$LINENO`      | Yes  | **No**                          |
| `env`, `set` listing                         | Yes  | `set` lists variables           |

Terebra variables live in an in-process map; `export` mirrors them into the
process environment. There is no notion of `local` scope or positional
parameters.

### String manipulation (inside `${…}`)

Implemented (see `testdata/strings.txtar` and `internal/shell/completion.go`):

| Form                                 | Effect                                 |
|--------------------------------------|----------------------------------------|
| `${v:n}`, `${v:n:len}`               | substring (negative offsets supported) |
| `${v/old/new}`                       | replace first                          |
| `${v//old/new}`                      | replace all                            |
| `${v#pat}`, `${v##pat}`              | remove shortest/longest prefix         |
| `${v%pat}`, `${v%%pat}`              | remove shortest/longest suffix         |
| `${v^}`, `${v^^}`, `${v,}`, `${v,,}` | case conversion                        |
| `${#v}`                              | length                                 |

The `#`/`%` prefix/suffix patterns are literal prefix/suffix matches, not glob
patterns as in bash.

### Arrays

| Feature                                  | Bash         | Terebra                         |
|------------------------------------------|--------------|---------------------------------|
| `arr=(a b c)`                            | Yes          | Yes                             |
| `arr[i]=x`                               | Yes          | Yes                             |
| `${arr[i]}`, `$arr[i]`                   | Yes          | Yes                             |
| `${arr[@]}`                              | Yes          | Yes                             |
| `${#arr[@]}`                             | Yes          | Yes                             |
| `${!arr[@]}` (keys/indices)              | Yes          | Yes                             |
| Associative arrays                       | `declare -A` | Auto-created on non-numeric key |
| `${arr[@]:1:2}` slicing, `${arr[@]/x/y}` | Yes          | No                              |
| `declare -a` / `declare -A` (explicit)   | Yes          | **No**                          |

Terebra auto-promotes an array to associative when a key is not numeric
(`internal/shell/shell.go` `getArrayVar`/assignment handling), so
`colors[red]=r; echo ${colors[red]}` works without `declare -A`.

### Arithmetic

`$(( … ))` is supported with `+ - * / %`, parentheses, unary `+`/`-`, and
variable references (`internal/shell/completion.go` `evalExpr`/`evalTerm`/
`evalFactor`). Integer division by zero yields `0` rather than an error. Not
supported: comparisons, `&&`/`||`, bitwise operators, `**`, `++`/`--`,
hex/octal literals, or the `(( ))` compound command.

## Quoting and expansion order

| Feature                                   | Bash | Terebra                      |
|-------------------------------------------|------|------------------------------|
| Single quotes (literal)                   | Yes  | Yes                          |
| Double quotes (allow `$`, `` ` ``, `\`)   | Yes  | Yes                          |
| Backslash escaping                        | Yes  | Yes                          |
| `$'…'` ANSI-C quoting                     | Yes  | **No**                       |
| `$"…"` locale strings                     | Yes  | **No**                       |
| Command substitution `$( )` and `` ` ` `` | Yes  | Yes (nested `$()` supported) |

Terebra's expansion pipeline is fixed (`docs/design.md`): parse → **brace** →
**tilde** → **variable** → **glob** → execute.

Brace expansion covers comma lists, numeric/alpha ranges with steps, zero
padding, reverse ranges, nesting, and cartesian products of adjacent groups
(`testdata/brace.txtar`). Tilde expansion covers `~` and `~user`; quoted or
mid-word `~` is left literal (`testdata/tilde.txtar`).

Globbing supports `*`, `?`, `[...]`, and recursive `**`. Unlike bash, `**` is
always enabled (no `globstar` option). `set -o nullglob` and `set -o dotglob`
control unmatched-glob and leading-dot behavior.

## Control flow and scripting

Terebra executes scripts in two ways:

1. **Shell-parser path** — a single line (or `-c` without newlines) is parsed as
   a shell command line: pipes, redirects, chaining, expansion.
2. **Scripting-language path** (`internal/script`) — multi-line scripts and
   `.trb` files are parsed by a dedicated interpreter supporting:

| Construct      | Syntax                                             |
|----------------|----------------------------------------------------|
| Conditional    | `if <cmd>; then … elif … else … fi`                |
| For loop       | `for x in a b c; do … done`                        |
| While / Until  | `while <cmd>; do … done`, `until <cmd>; do … done` |
| Functions      | `function name { … }` or `name() { … }`            |
| Error handling | `try … catch … end`                                |
| Include        | `source file.trb`                                  |
| Shebang        | `#!/usr/bin/env terebra`                           |

Not supported in terebra's scripting language: `case`/`esac`, `select`,
`break`/`continue`/`return`, C-style `for (( ))`, `[[ ]]`, arithmetic
`(( ))`, function parameters, local scope, or `getopts`.

Because control flow lives in the scripting-language parser, a one-line
`terebra -c 'if …; then …; fi'` is **not** interpreted as a conditional. Use a
`.trb` file or a `-c` string containing newlines.

## Job control and signals

| Feature                          | Bash | Terebra                                              |
|----------------------------------|------|------------------------------------------------------|
| `cmd &` background               | Yes  | Yes                                                  |
| `jobs`                           | Yes  | Yes (list with id/state)                             |
| `fg [n]`, `bg [n]`               | Yes  | Yes (numeric id only, no `%1` specs)                 |
| Ctrl+C (SIGINT) forwarding       | Yes  | Yes                                                  |
| Ctrl+Z (SIGTSTP) stop + resume   | Yes  | Yes                                                  |
| Background inside a pipeline     | Yes  | **No** (`background not supported in pipelines yet`) |
| `disown`, `wait`, `kill` builtin | Yes  | **No** (`kill` only as external command)             |

## Interactive REPL

| Feature                          | Bash                      | Terebra                                                       |
|----------------------------------|---------------------------|---------------------------------------------------------------|
| Line editing                     | Readline                  | `chzyer/readline`                                             |
| Emacs keybindings                | Yes (default)             | Yes (default)                                                 |
| Vi keybindings                   | `set -o vi`               | `set -o vi` (also `set -o emacs`)                             |
| Persistent history file          | `~/.bash_history`         | `~/.terebra_history`                                          |
| Reverse history search           | Ctrl+R (incremental)      | Ctrl+R **fuzzy TUI** (bubbletea)                              |
| History expansion `!!`, `!$`     | Yes                       | **No**                                                        |
| `history -c` / `-d` / `-w`       | Yes                       | **No**; `history [n]` lists only                              |
| Tab completion                   | Programmable (`complete`) | Fixed: commands, builtins, files, `PATH`, `~`                 |
| Flag/option completion           | Programmable              | No                                                            |
| Prompt `PS1`                     | Escape codes              | `{{.}}`, `{{exit}}`, `{{exitcode}}` templates + `$( )` output |
| Multi-line prompt                | Escapes                   | Supported                                                     |
| Syntax highlighting while typing | Via plugins               | No                                                            |
| `--explain` dry run              | No                        | Yes                                                           |
| Startup file                     | `~/.bashrc`/`~/.profile`  | `~/.terebrarc`                                                |

`set` options in terebra: `-x`/`+x` (trace), `-o vi`/`-o emacs`,
`-o nullglob`/`-o no-nullglob`, `-o dotglob`/`-o no-dotglob`. Bash options such
as `-e` (errexit), `-u` (nounset), `-a` (allexport), `-o pipefail`, and `shopt`
have no terebra equivalent.

## Built-in commands

Bash builtins (non-exhaustive) include `alias`, `bg`, `bind`, `break`,
`builtin`, `caller`, `cd`, `command`, `compgen`, `complete`, `declare`, `dirs`,
`disown`, `echo`, `enable`, `eval`, `exec`, `exit`, `export`, `fc`, `fg`,
`getopts`, `hash`, `help`, `history`, `jobs`, `kill`, `let`, `local`, `logout`,
`mapfile`, `popd`, `printf`, `pushd`, `pwd`, `read`, `readonly`, `return`,
`set`, `shift`, `shopt`, `source`, `suspend`, `test`/`[`, `times`, `trap`,
`type`, `typeset`, `ulimit`, `umask`, `unalias`, `unset`, `wait`.

Terebra builtins (the complete set):

| Builtin | Notes |
|---------|-------|
| `cd [dir]` | Defaults to `$HOME`; supports `cd -` and sets `PWD`/`OLDPWD` |
| `pwd` | Print working directory |
| `echo [args]` | Joins args with spaces; **no** `-n`/`-e` flags |
| `exit [code]` | Exit shell |
| `exec [-a name] <cmd>` | Replace the shell via `syscall.Exec` (never returns) |
| `help [cmd\|topic]` | Built-in help |
| `type [cmd]` | Reports builtin vs `PATH` location |
| `which [cmd]` | Locate in `PATH` |
| `export [name[=value]]` | Set/list environment |
| `unset <name>` | Remove variable |
| `set [-x\|+x] [-o opt]` | List vars / toggle options |
| `readonly [name[=value]]` | Mark read-only |
| `alias`, `unalias` | First-word alias expansion |
| `history [n]` | List history from the history file |
| `source <file>` | Execute a script file |
| `jobs`, `fg [n]`, `bg [n]` | Job control |
| `drill <sub>` | Terebra-only structured inspection |
| `cue <sub>` | Terebra-only CUE operations |
| `plugin <sub>` | Terebra-only Go plugin management |
| `state [save\|load]` | Terebra-only shell state as CUE |

Notably absent (present in bash): `read`, `printf`, `test`/`[`, `eval`,
`trap`, `wait`, `kill`, `umask`, `declare`/`typeset`, `local`, `let`, `shift`,
`command`, `builtin`, `pushd`/`popd`/`dirs`, `hash`, `fc`, `getopts`, `bind`,
`compgen`, `complete`, `shopt`, `ulimit`, `times`, `callout` helpers. Terebra
also has no built-in `ls`; it must be the external binary.

## Terebra-only capabilities

These have no bash equivalent and are the reason to use terebra:

- **`drill`** — inspect structured data and emit CUE:
  - `drill cue <file>` with `-e 'path'` (extract), `-v` (validate),
    `--export json`, `--unify other.cue`.
  - `drill fs <path>` (`-r` recursive, `-l` follow symlinks) — file metadata
    as CUE.
  - `drill proc <pid>` — process info from `/proc` (Linux).
  - `drill net <host>` — interfaces, DNS lookup, TCP connect probe.
- **`cue`** — `eval`, `vet`, `export`, `def`, `fmt`, `trim`, usable as pipe
  filters.
- **Auger pipe `|>`** — treats the left-hand output as CUE and passes it to the
  right-hand command (`cmd1 |> cmd2`). Encoders at the pipe end:
  `|>json`, `|>cue`, `|>yaml`. Current limitations: only the simple two-stage
  `cmd1 |> cmd2` form is special-cased; mixed `|`/`|>` pipelines fall back to
  plain pipes. There is no field-extraction syntax (`|>.field`) or `select`/
  `where`/`sort`/`group`/`join` filter set, and `|>yaml` currently emits
  CUE-formatted text rather than YAML.
- **`state`** — export/import shell variables as a self-validating CUE document;
  `state save <name>` / `state load <name>` under `~/.terebra/states/`.
- **Go plugins** — load `.so` plugins (`plugin load`, `plugin list`) from
  `~/.terebra/plugins`, `/usr/lib/terebra/plugins`, `/usr/local/lib/terebra/plugins`,
  or `$TEREBRA_PLUGIN_PATH`. Plugins can add commands, completions, prompt
  segments, and CUE encoders/decoders.
- **Compilation** — `terebra build script.trb` embeds the script in a Go
  program and compiles a standalone binary.
- **Fuzzy Ctrl+R** — a full-screen fuzzy finder over history, rather than bash's
  incremental reverse search.

## Side-by-side examples

Portable as-is:

```bash
FOO=bar ; echo "$FOO"
arr=(a b c) ; echo "${arr[@]}"
echo {1..5}
echo $((2 + 3))
echo hello | wc -c
echo hello 2>&1 | cat
cat <<< hello
set -x ; echo hello
```

Bash works, terebra does not:

```bash
echo "${FOO:-default}"          # default expansion — unsupported
echo "$1 $@ $#"                 # positional params — unsupported
printf '%s\n' "$FOO"            # printf builtin — use external printf
read -r line                    # read builtin — unsupported
[[ -f /etc/hosts ]] && echo ok  # [[ ]] — unsupported
for ((i=0;i<3;i++)); do …       # C-style for — unsupported
eval "$cmd"                     # eval builtin — unsupported
trap 'cleanup' EXIT             # trap — unsupported
echo $'\n'                      # $'…' — unsupported
diff <(a) <(b)                  # process substitution — unsupported
!!                              # history expansion — unsupported
set -euo pipefail               # set options — unsupported
case $x in a) ;; esac           # case — unsupported
```

Terebra works, bash does not:

```bash
drill proc $$ |> json                       # process info as structured data
drill fs . -r |> cue vet ./schema.cue       # validate filesystem facts
echo 'name: "x"' |> json                    # CUE in, JSON out
state save mysession ; state load mysession # CUE-backed shell state
terebra build script.trb myapp              # standalone binary
```

## Portability checklist

When moving a bash script to terebra, expect to rewrite:

1. `read`, `printf`, `test`/`[`, `eval`, `trap`, `wait`, `kill`, `umask` — call
   external binaries where they exist, or restructure.
2. Parameter defaults (`${v:-…}` and friends) — expand manually or guard with
   `if`.
3. Positional parameters and function arguments — not available.
4. `case`, `[[ ]]`, `(( ))`, C-style `for`, `break`/`continue`/`return` — not
   available in the scripting language.
5. History expansion and `history -c`/`-d`.
6. Process substitution and `$'…'`.
7. `set -e`/`-u`/`-o pipefail` — no equivalents.

Expect to gain: structured CUE pipelines, `drill`/`cue` introspection, `state`
checkpointing, Go plugins, fuzzy history search, and script compilation.

## How this document was verified

Feature claims were checked against:

- `testdata/*.txtar` (execution, expansion, redirection, chaining, arrays,
  strings, tilde, brace, glob, heredoc, auger).
- `internal/parser/{lexer,ast}.go` (operators, redirect tokens).
- `internal/shell/{executor,job,completion,shell}.go` (builtins, jobs,
  arithmetic, string ops, prompt, options).
- `internal/script/{parser,ast,interpreter}.go` (control flow, functions).
- `internal/drill/*.go` and `internal/builtins/cue.go` (structured commands).
- `internal/plugin/plugin.go` and `build.go` (plugins, compilation).

Items described in `docs/design.md` that are **not** implemented and are
therefore excluded above: `declare -A`, `history -c`/`-d`, TUI syntax
highlighting, interactive `drill` viewer, progress display, auger filters
(`select`/`where`/`sort`/`group`/`join`), `|>.field` extraction, CUE-typed
variables (`let port: int & >1024 = 8080`), function local scope, and remote
shell state.
