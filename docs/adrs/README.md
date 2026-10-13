# Architecture Decision Records

MADR-style ([MADR][madr]) decision records for arcadia.

## Conventions

- One ADR per file: `NNNN-kebab-title.md`, numbered sequentially
  starting at `0001`.
- Use the MADR 3.0 sections: Status / Context and Problem Statement /
  Decision Drivers / Considered Options / Decision Outcome (with
  Consequences and Confirmation) / Pros and Cons of the Options.
  Extra normative sections are allowed when a decision carries a
  contract (see `0001`).
- Status lifecycle: Proposed -> Accepted (or Rejected), later
  Deprecated or Superseded by a new ADR (record the superseding ADR
  number in Status).

## Index

- [0001. Organize agent artifacts under `agents/`][0001]
- [0002. Command lifecycle: `exp/` to `cmd/` at v1.0.0][0002]
- [0003. Agent worktrees: one per concurrent session][0003]
- [0004. Terebra core: a superset of mvdan/sh][0004]

## Link references

[madr]: https://adr.github.io/madr/
[0001]: 0001-organize-agent-artifacts-under-agents.md
[0002]: 0002-command-lifecycle-exp-to-cmd.md
[0003]: 0003-agent-worktrees-for-concurrent-agents.md
[0004]: 0004-terebra-superset-of-mvdan-sh.md
