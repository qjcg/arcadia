# Worktrees for concurrent agent sessions

This is the working procedure for running several agents against Arcadia
at the same time without stepping on each other. The decision is
recorded in
[ADR 0003](../adrs/0003-agent-worktrees-for-concurrent-agents.md);
use the terms in [`GLOSSARY.md`](../../GLOSSARY.md).

## Layout

Each concurrent agent session works in its own **worktree** — a second
working directory of the same repository, sharing one history and one
set of refs. All worktrees live under one directory per repository:

    ~/src/worktrees/<repo>/<slug>

for this repo: `~/src/worktrees/arcadia/<slug>`. This is the same
directory the pi-subagents harness uses for its managed worktrees, so
every working copy of a repo — session worktrees and harness
worktrees alike — shows up in one place.

`<slug>` is lowercase letters, digits, and hyphens, shaped
`<issue-or-task-id>-<short-desc>` (e.g. `142-worktree-docs`). The
worktree's branch is named after the slug unless you pass an explicit
branch name. `git worktree list` is the registry; there is no other
inventory to keep in sync.

Session notes and other out-of-repo artifacts for a repo live under:

    ~/src/worktrees/<repo>/.notes/<feature>/

The `go.work` workspace file is tracked, so every worktree is a
self-contained Go workspace with no setup step.

Create and remove worktrees through the task targets:

```sh
task worktree:new -- <slug>           # branch off main (or reattach)
task worktree:new -- <slug> <branch>  # explicit branch name
task worktree:ls                      # git worktree list
task worktree:clean -- <slug>         # remove + prune; see Cleanup
```

`worktree:new` reattaches an existing branch instead of failing, so a
crashed session can pick its work back up in a fresh worktree.

## Branch-only territory

All worktrees share one set of git refs, so a worktree is **branch-only
territory**. It may branch, commit, build, test, and run `task lint` /
`task test` freely, but these belong to the **main checkout** only:

- Creating tags: no `git tag`, no `sv` tagging subcommands, no
  `task release`. Release tags are global refs; two agents tagging at
  once would fight over them.
- Merging to `main` and pushing `main`.
- Anything else that writes a shared ref outside the worktree's own
  branch.

Pushing the worktree's own branch (e.g. before opening a PR) is fine:
branch names are the namespace that keeps concurrent agents apart.

## The main checkout is single-tenant

Whoever is integrating or releasing holds the main checkout; everyone
else works in a worktree. `task worktree:new` is cheap, so no two
sessions should ever be doing feature work in the main checkout at the
same time.

## Integration

1. In the worktree: `git fetch origin && git rebase origin/main`.
2. Make sure `task test` and `task lint` pass.
3. Hand off to whoever holds the main checkout.
4. There: `git merge --ff-only <branch>`, then push.

Rebase before handing off so `main` stays linear — this is what keeps
`sv changelog` output clean. If a change needs review first, open a PR
with `gh`; the worktree rules are the same either way.

## Cleanup

At the end of a session (or when the work is done):

    task worktree:clean -- <slug>

It removes the worktree, prunes stale registrations, and deletes the
branch. It refuses when the branch has commits that are not in `main` or
the worktree is dirty; pass `FORCE=1` only when you mean it.

There are no automatic sweeps. Deleting an agent's in-flight work is
worse than a leaked directory; leaked worktrees surface in
`git worktree list` and `git worktree prune` tidies up registrations
after the fact.

## Harness-managed worktrees

Agent harnesses (e.g. pig subagents run with worktree isolation) may
create managed worktrees in the same directory, named
`pi-worktree-*`. They follow the same rules as session worktrees —
any directory holding an agent session is branch-only territory —
but their lifecycle belongs to the harness: do not `task worktree:clean`
them, and do not reuse their names for session worktrees.

## Shared machine state

`bin/`, `dist/`, and `go` build caches are per-worktree or
machine-global-but-concurrency-safe; concurrent `task build`,
`task lint`, and `task test` runs need no coordination. GitHub issues
and PRs are shared state too: one session owns an issue's state changes
at a time, just like code review comments.

## Link references

- [ADR 0003: Agent worktrees][adr-0003]
- [GLOSSARY.md][glossary]

[adr-0003]: ../adrs/0003-agent-worktrees-for-concurrent-agents.md
[glossary]: ../../GLOSSARY.md
