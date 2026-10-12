# 0003. Agent worktrees: one worktree per concurrent agent session

## Status

Accepted (2026-10-12)

## Context and Problem Statement

Arcadia is increasingly worked by several agents at once — parallel
harness sessions, delegated subagents, and humans interleaved with
them. All of them land in one clone and immediately collide: two
sessions editing the same working tree, `sv` release tags created
twice, one agent rebasing under another's feet. We need a convention
that lets concurrent agent sessions work in parallel without
coordinating every keystroke, and that names who may touch the state
they nonetheless share.

## Decision Drivers

* Git worktrees share one history *and one set of refs* — tags and
  branches are global across all worktrees of a clone.
* `sv` path-scoped release tags (`exp/<name>/v0.x.y`) are git refs, so
  concurrent releasers conflict even from separate worktrees.
* The tracked `go.work` makes each worktree a self-contained Go
  workspace with zero setup.
* `git worktree list` is already a complete, authoritative registry of
  worktrees; a second inventory would only go stale.
* `Taskfile.yaml` is the single authoritative source for repo tasks;
  procedures belong there or in `docs/agents/`, not in ad-hoc scripts.
* Agent harnesses (e.g. pig) can isolate subagents in managed
  worktrees at paths the repo does not control.

## Considered Options

1. Share the main checkout, serialize with turn-taking or a lockfile
2. One full clone per agent session
3. Convention: sibling-directory worktrees plus task targets
   (chosen)
4. Harness-managed worktrees only

## Decision Outcome

Chosen option: **"Convention: sibling-directory worktrees plus task
targets"**, because worktrees give isolation at near-zero cost while
sharing one object store, and the dangerous shared state is small and
nameable.

* Each concurrent session works in its own worktree at
  `../arcadia.worktrees/<slug>`, slug
  `<issue-or-task-id>-<short-desc>`, branch named after the slug.
  `git worktree list` is the registry.
* A worktree is **branch-only territory**: it may branch, commit,
  build, test, and lint, but never create tags, never run `task
  release`, and never merge or push `main`. Those belong to the main
  checkout.
* The main checkout is single-tenant: the session integrating or
  releasing holds it; everyone else is in a worktree.
* Integration is rebase onto `origin/main`, then `git merge --ff-only`
  from the main checkout. Linear history keeps `sv changelog` output
  clean.
* Lifecycle is `task worktree:new|clean|ls`, backed by
  `internal/worktrees/*.sh`. `worktree:clean` refuses unmerged or dirty
  worktrees unless `FORCE=1`. No automatic sweeps.
* Harness-managed worktrees (pig subagent isolation) follow the same
  rules wherever they live; the layout convention is for sessions
  humans and agents create themselves.

The working procedure lives in `docs/agents/worktrees.md`.

### Consequences

* Good, because concurrent sessions never share a working tree, so
  file-level conflicts become merge-time review instead of corrupted
  checkouts.
* Good, because the shared surface is reduced to named refs and the
  main checkout, each with exactly one owner.
* Good, because a worktree is self-contained (`go.work` is tracked) and
  cheap enough that isolation is the default, not an ask.
* Neutral, because `git worktree list` discipline is required to find
  and clean up leaked worktrees.
* Bad, because ref discipline is convention, not enforcement: nothing
  physically stops `git tag` in a worktree. The task targets make the
  safe path the easy path.
* Bad, because the main checkout becomes a handoff point; solo sessions
  pay a small ceremony cost even when nobody else is working.

### Confirmation

* `task worktree:new -- demo-slug` creates
  `../arcadia.worktrees/demo-slug` on branch `demo-slug` off `main`;
  `task worktree:clean -- demo-slug` removes it and prunes.
* `task worktree:clean` refuses an unmerged branch without `FORCE=1`.
* `task validate-conventional-commit` passes on this change's commits.

## Pros and Cons of the Options

### Share the main checkout with turn-taking or a lockfile

* Good, because zero extra disk, zero new concepts.
* Good, because no shared-ref discipline needed — there is only one
  checkout.
* Bad, because a lock serializes all work; an agent's long build or
  half-finished edit blocks everyone.
* Bad, because a lockfile around git operations does not stop two
  editors racing on the same files.

### One full clone per agent session

* Good, because refs and working tree are fully independent.
* Good, because no worktree ref-discipline caveats.
* Bad, because N clones duplicate the object store and per-clone
  caches.
* Bad, because clones drift: `main` goes stale independently and
  re-syncing is manual in each clone.

### Harness-managed worktrees only

* Good, because the harness owns the lifecycle end to end.
* Bad, because harness worktrees live at harness-chosen paths and are
  reaped on the harness's schedule, not the repo's.
* Bad, because it does nothing for independent sessions or humans who
  are not harness children.

## Link references

- [Git worktrees][git-worktrees]
- [ADR 0001: organize agent artifacts under `agents/`][adr-0001]

[git-worktrees]: https://git-scm.com/docs/git-worktree
[adr-0001]: 0001-organize-agent-artifacts-under-agents.md
