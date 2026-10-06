# Command lifecycle: experiments in `exp/`, promotion to `cmd/` at v1.0.0

## Status

Accepted (2026-10-05)

## Context and Problem Statement

Arcadia is a Go monorepo with independent Go modules versioned by path-scoped `sv`
tags. Its `exp/` and `cmd/` directories have grown organically: some experiments
are modules with their own `v0.x` tags, some are packages inside the stable root
module, and `cmd/` contains commands that have not reached v1.0.0. We need one
lifecycle that makes experimentation cheap, stable command paths deliberate, and
tags understandable to both Go tooling and humans.

The decision is how commands begin, what their tags mean before and after
stabilization, how they cross the boundary, and how the repo prevents the old
ambiguity from returning.

## Decision Drivers

* A command is experimental until a human deliberately promises stability.
* Go modules need independent path-scoped versions in this monorepo.
* A module-path change is a breaking change; experimental consumers must expect it.
* `cmd/` must mean stable v1.0.0-or-later commands, with no grandfathered exceptions.
* The workflow must be explicit for humans and deterministic for agents.
* Experiments may not create source-code dependency pressure on stable modules.

## Considered Options

### Where commands start

1. **Start commands directly in `cmd/` and use v0.x to signal instability.**
2. **Start each command in `exp/` as its own module; move it to `cmd/` when it is
   ready for v1.0.0.**
3. **Keep a mixed `exp/`: some commands as root-module packages, others as modules.**

### What promotion means for versions

1. **Promotion cuts `cmd/<name>/v1.0.0`.**
2. Carry the `exp/<name>` version number forward into a new `cmd/<name>` tag line.
3. Start the new `cmd/<name>` line at v0.1.0 and promote to v1.0.0 later.

### What to do with existing pre-1.0 commands in `cmd/`

1. Grandfather them in place until they reach v1.0.0.
2. Move them to `exp/`, preserving old tags and seeding matching version tags in
the new `exp/<name>` namespace.

## Decision Outcome

Chosen option: **commands start in `exp/` as independent modules and are promoted
into `cmd/` at v1.0.0**. The lifecycle is normative in
[`docs/agents/command-lifecycle.md`](../agents/command-lifecycle.md).

### Normative lifecycle contract

* Every new Go command starts at `exp/<name>` with module path
  `github.com/qjcg/arcadia/exp/<name>`, a `go.mod`, and a `go.work` entry. Its
  version tags use `exp/<name>/v0.x.y`, managed by `sv` and the normal release
  workflow. Non-command material under `exp/` is outside this lifecycle.
* An experiment makes no compatibility promise: APIs and flags may change, the
  import path may change at promotion, and the command may be removed.
* Go dependencies point from experiments toward stable code, never the reverse.
  Stable Go source must not import an experiment, and stable modules must not
  require an experiment in `go.mod` (including as a `tool`). Experiments do not
  depend on one another. Run repository experiments as workspace packages with
  `go run ./exp/<name>` instead.
* Promotion is an explicit human decision, never an automatic consequence of a
  metric. It moves `exp/<name>` to `cmd/<name>`, updates the module path, workspace
  membership, and source imports, and creates the annotated
  `cmd/<name>/v1.0.0` tag. **This is the only way a command enters `cmd/`.**
* Promotion requires the checklist in the workflow: passing tests/lint, useful
  README and changelog, real-use soak, and no known breaking-change debt. The
  script enforces mechanical checks; the human owns the judgement calls.
* Retirement is deletion. Remove the experiment and its workspace entry; there is
  no graveyard or ceremony. Git history and existing tags remain.
* `cmd/` contains only Go modules whose latest version is v1.0.0 or later. A
  policy check runs under `task lint` to enforce this, the experiment-module
  rule, source dependency direction, and experiment independence.

### One-time migration

* Move `cmd/ap`, `cmd/awesome-lint`, and `cmd/horeb` to `exp/`; keep their old
  `cmd/<name>/v0.x.y` tags as frozen history and seed matching
  `exp/<name>/v0.x.y` annotated tags at the migration commit. `cmd/sv` remains in
  place because it is already v1.0.0 or later.
* Give the existing root-module Go commands under `exp/` their own `go.mod` and
  `go.work` entries. This removes experimental code from the root module's
  release line and makes each command independently taggable.

### Consequences

* Good, because `exp/` and `cmd/` now communicate lifecycle state, while every
  command has its own `sv` tag line.
* Good, because promotion's new module path and v1.0.0 tag form one explicit
  stability boundary; old exp tags remain available as history.
* Good, because experiments cannot force stable modules to depend on an
  unstable import or tool path.
* Neutral, because promotion is a breaking import-path change for exp consumers;
  the experiment contract makes that explicit.
* Bad, because each experiment needs a small `go.mod` and `go.work` entry from
  birth. This is intentional overhead for independent versioning and isolation.
* Bad, because the one-time move changes the module paths for three existing
  commands; old cmd tags remain frozen and the exp namespace receives continuity
  tags.

### Confirmation

* `task check:policy` passes after the one-time migration is complete.
* `go test work` and `go build work` cover all modules in `go.work`.
* A promotion testscript verifies the move, import/module-path rewrites, workspace
  update, commit, and v1.0.0 tag.
* Retirement and policy-check test scripts cover their success and failure paths.

## Pros and Cons of the Options

### Start commands directly in `cmd/` and use v0.x

* Good, because there is no directory move before 1.0.
* Bad, because `cmd/` does not reveal whether a command is stable; its contents
  repeat the current mixture.

### Mixed module and root-package experiments

* Good, because tiny experiments need no module setup.
* Bad, because they share the root module's release line and cannot be
  independently versioned; this is already the source of ambiguous history.

### Carry the exp version number forward at promotion

* Good, because it preserves a human-readable maturity counter.
* Bad, because `cmd/<name>` is a new Go module path and would start with an
  unexplained gap in its own tag history.

### Restart the cmd line at v0.1.0

* Good, because each module path gets a conventional first tag.
* Bad, because `cmd/` would again contain pre-1.0 commands and promotion would no
  longer coincide with the stability promise.

### Grandfather existing pre-1.0 cmd modules

* Good, because it avoids moving existing module paths.
* Bad, because it makes `cmd/`'s central invariant conditional and leaves an
  exception agents could copy for future commands.

### Leave experiments untagged or add their module boundary later

* Good, because it reduces setup for disposable code.
* Bad, because experiments cannot be independently versioned from birth, and
  deciding when a command becomes taggable becomes subjective.
