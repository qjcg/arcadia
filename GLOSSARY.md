# Arcadia

Arcadia is a polyglot monorepo of Go commands and experiments. This glossary pins
the domain language around the command lifecycle.

## Language

**Experiment**:
A command in its pre-1.0 incubation phase, where breaking changes are free and the
command may be modified arbitrarily or removed without notice.
_Avoid_: prototype, WIP command

**Retirement**:
The act of removing an experiment that will not be promoted. It is deletion —
simply and without ceremony; history and old tags preserve what mattered.
_Avoid_: deprecation, archiving, sunsetting

**Stable command**:
A command that has been promoted. It lives under `cmd/` and carries a v1.0.0-or-later
version, promising import-path and API stability.
_Avoid_: finished command, production command

**Promotion**:
The act of declaring an experiment stable: it moves out of incubation and its
v1.0.0 is cut as part of the move. The only way a command may enter `cmd/`.
_Avoid_: stabilization
