# Command lifecycle: experiment, promote, retire

This is the working procedure for Go commands in Arcadia. The lifecycle decision
is recorded in [ADR 0002](../adrs/0002-command-lifecycle-exp-to-cmd.md); use the
terms in [`GLOSSARY.md`](../../GLOSSARY.md).

## Start a command as an experiment

Every new command starts under `exp/`, never directly under `cmd/`:

1. Create `exp/<name>/` and initialize an independent module:
   `module github.com/qjcg/arcadia/exp/<name>`.
2. Add it to the workspace with `go work use ./exp/<name>`.
3. Add a README with purpose and usage, tests for its behavior, and any required
   docs. Keep dependencies flowing from the experiment toward stable code; stable
   Go source and stable `go.mod` files must not depend on an experiment. Experiments
   must not depend on each other. To run a repository experiment as a tool, use it
   from the workspace with `go run ./exp/<name>`, not a stable module's `tool`
   directive.
4. Run `task test` and `task lint`. Inspect the proposed version with
   `go tool sv next --path exp/<name>`; `task release` creates the normal
   path-scoped `exp/<name>/v0.x.y` tags.

An experiment has no compatibility promise: breaking changes are allowed, its
module path may change on promotion, and it may be deleted. Consumers should pin
a version and treat the path as temporary.

## Promote an experiment

Promotion is the explicit human decision to promise stability. It is the only
route into `cmd/` and cuts `cmd/<name>/v1.0.0` as part of the move.

Before promoting, confirm all of the following:

- `task test` and `task lint` pass across the workspace.
- The README documents real usage; a changelog is present. If it needs to be
  generated, run `task changelogs` and commit that update before promotion.
- The command has seen real use for a meaningful soak period.
- There is no known breaking-change debt that should be resolved before making a
  stability promise.
- You explicitly choose to take responsibility for the stable import path and
  behavior. The checklist does not promote anything automatically.

With a clean working tree, run:

```sh
task promote -- <name>
```

The task verifies the module, README, changelog, tests, and lifecycle policy; moves
`exp/<name>` to `cmd/<name>`; rewrites the module path, source imports, and
workspace membership; runs `go mod tidy`, tests, and vet; then creates
one conventional commit and the annotated `cmd/<name>/v1.0.0` tag. Inspect the
result before pushing the branch and tag through the repository's normal review
and release workflow. The exp module's old tags remain in Git as historical
versions of the old module path.

The move changes the Go module path. Any consumer of `exp/<name>` must update its
imports and dependency path; that break is expected and is why experiments carry
no stability promise.

## Retire an experiment

If an experiment will not be promoted, delete it without an archive or ceremony:

```sh
task retire -- <name>
```

The task removes `exp/<name>`, drops its `go.work` entry, syncs the workspace,
and commits the deletion. Existing tags and Git history remain available.

## Policy checks

`task check:policy` is also part of `task lint`. It checks that:

- every `exp/*` directory containing Go sources has its own `go.mod`;
- every `cmd/*` module's latest version tag is v1.0.0 or later; and
- Go source outside `exp/` does not import an `exp/` package; and
- no `go.mod` outside `exp/` requires an `exp/` module (including tool pins); and
- no experiment module requires another experiment module.

A new directory under `cmd/` without a v1.0.0-or-later tag is a policy failure.
Non-command material under `exp/` (for example, docs or test assets) is outside
the command lifecycle.
