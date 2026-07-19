# PS1 & oh-my-posh Plan

## Problem

Currently `PS1` supports:
- Static text: `PS1=">> "`
- Template vars: `{{.}}` (cwd), `{{exit}}` (`$`/`!`), `{{exitcode}}` (numeric code)

oh-my-posh is invoked via `$(oh-my-posh print ...)` in bash/zsh — the shell executes the command on every prompt render and substitutes its output (ANSI-colored prompt string). Terebra's `prompt()` has no `$(...)` substitution, so `PS1='$(oh-my-posh print ...)'` renders the literal command text.

## Steps

### 1. Add `$(...)` command substitution to `prompt()`

The `prompt()` function already runs on every REPL iteration. After reading PS1 (from `s.vars` or `os.Getenv`), scan for `$(...)` expressions:

```
func expandPromptCmd(s string) string {
    // Find $(...) and execute each command
    // Replace with captured stdout
    // Return the expanded string
}
```

- Handle nesting: `$(echo $(date))`
- Handle errors: if command fails, substitute empty string or `(error)`
- Use `exec.Command("sh", "-c", cmdStr).Output()` for full shell evaluation (so pipes, redirects, etc. work inside `$(...)`)
- Escape the result properly for the terminal (ANSI codes pass through)

### 2. Works with oh-my-posh directly

User sets:
```
PS1='$(oh-my-posh print --shell=generic)'
```

On each prompt render:
1. `prompt()` reads PS1 → `$(oh-my-posh print --shell=generic)`
2. `expandPromptCmd` executes `oh-my-posh print --shell=generic`
3. Captures stdout (ANSI-colored prompt string)
4. Returns the raw string → readline displays it

### 3. Cursor position

oh-my-posh output includes ANSI escape codes. The readline library needs to know the visible width of the prompt (excluding ANSI codes) to position the cursor correctly. readline handles this via `\x1b[...` sequences — it counts them as zero-width. oh-my-posh's output is already ANSI-compatible, so this should work out of the box.

### 4. Performance

`$(oh-my-posh print ...)` runs on every prompt render (every REPL iteration). oh-my-posh is fast (~5-20ms), so this is acceptable. If it becomes an issue, add a cache with a short TTL (e.g., 100ms).

### 5. Template vars still work

The existing `{{.}}`, `{{exit}}`, `{{exitcode}}` template vars continue to be processed before `$(...)` expansion, so users can mix both:
```
PS1='{{.}} $(oh-my-posh print ...)'
```

### 6. Backward compatibility

All existing PS1 usage (static text, template vars) is unaffected — `$(...)` expansion only triggers when `$(` is present in the string.

## Files to change

| File | Change |
|------|--------|
| `internal/shell/shell.go` | Add `expandPromptCmd()` function; call it in `prompt()` after reading PS1 |
| `internal/shell/executor.go` | Reuse `captureCommandOutput()` or add a helper for shell cmd execution |

## Test

```bash
PS1='$(oh-my-posh print --shell=generic)' terebra
# or inside terebra:
# PS1='$(oh-my-posh print --shell=generic)'
```

## Future

- `oh-my-posh init terebra` support (transient prompt, right-prompt, tooltips)
- Caching for slow prompt commands
- Async prompt rendering (render new prompt while previous command runs)

