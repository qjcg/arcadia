# Tilde Expansion Implementation Plan

## Goal

Expand a leading unquoted `~` in a word to the current user's home directory
(`$HOME`), and `~user` to that user's home directory, as part of the word
expansion pipeline.

## Background

The expansion pipeline lives in `internal/expand/expand.go`:

```
input → parse → brace expand → variable expand → glob → execute
```

`Pipeline` (expand.go:14) runs brace → variable → glob on a single word.
`ExpandCommand` (expand.go:43) expands the command name and each arg through
`Pipeline`. Words carry a per-byte quoted `Mask` (`Word` struct, expand.go:6)
so quoted regions are never expanded.

Tilde expansion is a new step. Per POSIX it runs before parameter expansion;
per bash it runs after brace expansion. So the new order is:

```
brace expand → tilde expand → variable expand → glob
```

## Semantics

| Input (word starts with unquoted `~`) | Result |
|----------------------------------------|--------|
| `~` or `~/...`                         | `$HOME` (or `$HOME/...`) |
| `~user` or `~user/...`                 | that user's home dir |
| `~+` / `~-` (optional, bash ext)      | `$PWD` / `$OLDPWD` |
| `~` with `$HOME` unset                | left literal |
| `~user` where user not found          | left literal |
| quoted `~` (`'~'`, `"~"`, `\~`)       | left literal |

Rules:
- Only a `~` at index 0 of the word, with `mask[0]` unquoted (nil mask counts
  as unquoted), triggers expansion.
- The `~` must be followed by `/`, end-of-word, `+`, `-`, or a username
  (`[A-Za-z0-9_.-]*` up to the next `/` or end-of-word).
- The inserted home path bytes are marked **quoted** in the mask so glob
  metacharacters inside the home path are not re-globbed (matches bash).

## Implementation Steps

### 1. Add `Tilde` to `internal/expand/tilde.go` (new file)

```go
// Tilde expands a leading unquoted ~ in w to the home directory.
// Returns w unchanged when no expansion applies.
func Tilde(w Word) Word
```

- Check `w.Value[0] == '~'` and `mask[0]` is unquoted.
- Parse the prefix after `~`:
  - empty or `/` → use `$HOME` (via `os.Getenv("HOME")`); if empty, return w.
  - `+` / `-` → `$PWD` / `$OLDPWD` (optional; skip if out of scope).
  - username → `os/user.Lookup(name)`; on error return w.
- Rebuild the word: `home + rest`, with the home bytes marked quoted in the
  mask (append `len(home)` `true` entries), rest keeps its original mask.
- Use `os/user` (stdlib) for `~user`; fall back to `$HOME` for bare `~`.

### 2. Wire into `Pipeline` (expand.go:14)

Insert a tilde step between brace expansion and variable expansion:

```go
// Step 1.5: tilde expansion (for each expanded word)
for i := range expanded {
    expanded[i] = Tilde(expanded[i])
}
```

### 3. Apply to command name in `ExpandCommand` (expand.go:43)

For consistency, run `Tilde` on the expanded command name before variable
expansion. (Low priority; `~` as a command name is rare but harmless.)

### 4. Tests

- `internal/expand/tilde_test.go` — unit tests:
  - `~` → `$HOME`
  - `~/foo` → `$HOME/foo`
  - `~user` → user home (use a known user or skip if lookup fails)
  - `~user/foo`
  - quoted `'~'`, `"~"`, `\~` → unchanged
  - `$HOME` unset → literal `~`
  - unknown user → literal `~user`
  - home path containing `*` is not globbed (mask is quoted)
- `testdata/tilde.txtar` — end-to-end testscript cases (mirror `glob.txtar`).

### 5. Docs

- Update `docs/design.md` expansion-order diagram and add a "Tilde Expansion"
  section.
- Update `docs/plan.md` with a task row.

## Edge Cases

- `~` mid-word (`a~b`) → no expansion (only leading `~`).
- `~` after brace expansion, e.g. `{~,x}` → `~` and `x`; the `~` word expands.
- `~` in a variable value is **not** re-expanded (variable expansion runs
  after tilde; matches bash).
- Windows: `os/user` and `$HOME` behave differently; keep behavior
  `$HOME`-driven and document that `~user` may be unsupported there.

## Out of Scope

- `~+` / `~-` (PWD/OLDPWD) — optional bash extension, add only if desired.
- Tilde completion in the REPL — done in `internal/shell/completion.go`
  (`completeFileOrArg` expands a leading `~`, `~/`, or `~user` to the real
  path before reading the directory; a bare `~` completes to `$HOME/`).
