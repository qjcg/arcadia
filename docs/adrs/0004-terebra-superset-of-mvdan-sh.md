# Terebra core: a superset of mvdan/sh with `interp` as the engine

## Status

Accepted (2026-10-12)

## Context and Problem Statement

Terebra (`exp/terebra`) hand-rolls its shell semantics. An interactive
pipeline language lives in `internal/parser` (lexer, parser, AST), a
separate `.trb` script language in `internal/script`, and word expansion
in `internal/expand` (brace, glob, tilde). Together that is roughly
2,500 lines of non-test code maintaining bash-like behavior across two
divergent ASTs, and it grows with every quoting, expansion, and
redirection edge case users already expect from a modern shell.

Terebra's product goal is to be a superset of mvdan/sh [mvdan-sh]: every
feature of its `syntax`, `expand`, and `interp` packages must be
available in terebra, and bash scripts must run unchanged. mvdan/sh has
no extension surface for non-bash grammar, and terebra has exactly one
family of non-bash syntax: the auger pipe vocabulary (`|>`, `|>json`,
`|>yaml`, `|>cue`, and planned `|>.field` extraction).

The decision is how much of mvdan/sh to adopt, what terebra code
survives around it, and how auger pipes coexist with a bash grammar
without forking mvdan/sh.

## Decision Drivers

* Superset goal: all mvdan/sh features available in terebra; real bash
  scripts run unchanged.
* Delete hand-rolled semantics instead of maintaining parallel ones.
* Terebra's identity survives: auger pipes, `drill`/`cue` structured
  output, job control, and the REPL experience.
* mvdan/sh stays a plain dependency, never a fork.
* Custom syntax must never silently change meaning under bash rules.
* The `testdata/*.txtar` suite is the acceptance gate.

## Considered Options

### How much of mvdan/sh to adopt

1. **Partial adoption: `syntax` + `expand` only.** Keep terebra's
   executor walking the `syntax` AST, including a new control-flow
   walker replacing `internal/script`.
2. **Full adoption: `syntax` + `expand` + `interp`.** A persistent
   `interp.Runner` is the execution core; terebra becomes middleware
   around it.
3. **Stay hand-rolled.** Keep `internal/parser`, `internal/script`,
   and `internal/expand` as they are.

### How to support `|>` without forking

1. **Desugar at the boundary (A).** A quote-aware scan rewrites
   unquoted `|>` into a reserved `__terebra_auger__` pipeline stage
   before parsing.
2. **Split-and-stitch (B).** Split source at top-level `|>`, parse
   each side with `syntax.Parser`, and stitch the pieces into
   terebra's pipeline container.
3. **Vendored grammar patch (C).** Fork `syntax` and add a `|>` token
   to the lexer and grammar.

## Decision Outcome

Chosen: **full adoption with the `|>` desugar (A).** Terebra is a
superset of mvdan/sh; mvdan/sh is the engine, and terebra is the shell
around it.

### Normative contract

* `interp.Runner` executes all shell semantics: pipelines,
  redirections, control flow, functions, traps, `set`/`shopt` options,
  field splitting, and every builtin. One persistent runner per REPL
  session holds variables, functions, and exit status.
* Terebra middleware owns: the `|>` desugar; the `__terebra_auger__`
  stage (CUE decode, reformat, and `--enc=` encoder endpoints); the
  `drill`, `cue`, `plugin`, and `state` builtins; job control
  (`jobs`, `fg`, `bg`, pty); the REPL experience (readline, history,
  prompts, completion); and `build.go`.
* Conflict rule: **bash wins** on builtin collisions (`cd`, `echo`,
  `set`, and friends). Terebra extras get their own names; wrapping
  runs through exec-handler middleware or `HandlerContext.Builtin`
  [sh-issue-705].
* `try`/`catch`/`end` is deprecated in favor of `set -e` and `trap`.
* The desugar and the parser swap land in the same commit. Under bash
  rules `cmd1 |> cmd2` parses silently as `cmd1 | >cmd2`; without the
  desugar the meaning changes without an error.
* Deleted after the swap: `internal/parser` and `internal/script`
  immediately; `internal/expand` except `brace.go`, which goes only
  after a compatibility check against `expand.BracesSeq` [pkg-expand]
  covers cartesian, nested, zero-padded, and stepped forms
  [sh-issue-1042].

### Confirmation

* The `testdata/*.txtar` suite passes, except diffs explicitly
  reviewed and accepted as bash-semantics wins.
* A corpus of real bash scripts runs unchanged.
* The job-control spike proves `fg`/`bg` over `interp` exec handlers
  before the migration begins.
* No remaining imports of the deleted packages.

## Consequences

Good:

* Roughly 2,500 lines of hand-rolled parsing, expansion, and
  interpretation are deleted, and the replacement is a maintained
  dependency.
* Bash compatibility comes from upstream and stays current, instead of
  being re-earned one edge case at a time.
* One AST and one expansion pipeline serve the REPL, scripts, and
  completion alike.

Bad:

* Core semantics couple to mvdan/sh's API evolution. This is a
  dependency risk, not a fork risk; pin versions deliberately.
* Job control (`fg`/`bg`, terminal handling) is the one genuinely hard
  piece terebra keeps, because `interp` has no job tables or SIGTSTP
  handling.
* Builtin behavior where terebra currently deviates from bash becomes
  bash behavior. Testdata changes must be reviewed deliberately.

Neutral:

* The `|>` desugar means auger pipes are invisible to mvdan/sh tools
  such as `shfmt`; acceptable, since `.trb` is a superset dialect.
* A future script mode running `.trb` as bash costs nothing extra; the
  engine is already there.

## Pros and Cons of the Options

### Partial adoption (`syntax` + `expand`, own executor)

* Good, because terebra keeps full control of execution semantics and
  builtin behavior.
* Good, because job control keeps working on terebra's own terms.
* Bad, because the superset goal then requires reimplementing `interp`
  semantics in the walker: the large half of bash, forever.

### Full adoption (`interp` as the engine)

* Good, because the superset goal is met by construction.
* Good, because it deletes the most code, including the control-flow
  walker and most of `executor.go`.
* Bad, because builtin collisions resolve to bash [sh-issue-705], and
  job control remains custom and hard.

### Stay hand-rolled

* Good, because nothing changes and terebra's semantics stay fully
  bespoke.
* Bad, because the superset goal is unreachable at reasonable cost and
  edge-case maintenance never ends.

### Desugar at the boundary (A)

* Good, because mvdan/sh stays unforked and the AST is pure bash.
* Good, because it scales to `|>.field` and future auger syntax
  without grammar work.
* Bad, because a small quote-aware scanner must be exactly right, and
  it must ship atomically with the parser swap.

### Split-and-stitch (B)

* Good, because no source rewriting happens at all.
* Bad, because terebra keeps a wrapper AST and parsing is split across
  segment boundaries.

### Vendored grammar patch (C)

* Good, because `|>` is first-class: errors, printing, and completion
  all see it.
* Bad, because it is a permanent fork to rebase; upstream accepts bash
  compatibility only.

## Link references

[mvdan-sh]: https://github.com/mvdan/sh
[pkg-expand]: https://pkg.go.dev/mvdan.cc/sh/v3/expand
[sh-issue-705]: https://github.com/mvdan/sh/issues/705
[sh-issue-1042]: https://github.com/mvdan/sh/issues/1042
