# Arcadia

Arcadia is a polyglot monorepo of Go commands and experiments. This
glossary pins the domain language around the command lifecycle and
concurrent agent work.

## Language

**Experiment**:
A command in its pre-1.0 incubation phase, where breaking changes are
free and the command may be modified arbitrarily or removed without
notice.
_Avoid_: prototype, WIP command

**Retirement**:
The act of removing an experiment that will not be promoted. It is deletion —
simply and without ceremony; history and old tags preserve what mattered.
_Avoid_: deprecation, archiving, sunsetting

**Stable command**:
A command that has been promoted. It lives under `cmd/` and carries a
v1.0.0-or-later version, promising import-path and API stability.
_Avoid_: finished command, production command

**Promotion**:
The act of declaring an experiment stable: it moves out of incubation and its
v1.0.0 is cut as part of the move. The only way a command may enter `cmd/`.
_Avoid_: stabilization

**Worktree**:
An isolated working directory of the repository that shares one history and one
set of refs with every other working directory. The unit of isolation for
concurrent agent sessions. _Avoid_: clone, checkout copy

**Main checkout**:
The single-tenant working directory where integration and releases happen; the
only place shared refs may be advanced. _Avoid_: primary clone, main clone

**Integration**:
The act of landing a worktree's branch into the main branch through the main
checkout, keeping history linear. _Avoid_: merge-back
