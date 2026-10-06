# Architecture Decision Records

MADR-style ([madrgithub.io](https://adr.github.io/madr/)) decision records for arcadia.

## Conventions

- One ADR per file: `NNNN-kebab-title.md`, numbered sequentially starting at `0001`.
- Use the MADR 3.0 sections: Status / Context and Problem Statement / Decision Drivers /
  Considered Options / Decision Outcome (with Consequences and Confirmation) /
  Pros and Cons of the Options. Extra normative sections are allowed when a decision
  carries a contract (see `0001`).
- Status lifecycle: Proposed → Accepted (or Rejected), later Deprecated or Superseded
  by a new ADR (record the superseding ADR number in Status).

## Index

- [0001. Organize agent artifacts under `agents/` with generated plugin packages](0001-organize-agent-artifacts-under-agents.md)
- [0002. Command lifecycle: experiments in `exp/`, promotion to `cmd/` at v1.0.0](0002-command-lifecycle-exp-to-cmd.md)
