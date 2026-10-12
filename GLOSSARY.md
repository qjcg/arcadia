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

### Fractalis

**Pilot**:
The policy that evolves a fractal view over time. Exactly one pilot is
active at any moment.
_Avoid_: mode, driver, camera mode

**Manual pilot**:
The default pilot, whose policy is to leave the view unchanged.
_Avoid_: manual mode, no pilot

**Auto-zoom pilot**:
The pilot that continuously zooms toward interesting regions.
_Avoid_: autopilot, auto-zoom

**Vantage pilot**:
The pilot that tours curated and discovered viewpoints.
_Avoid_: vantage mode, vantage tour

**Boundary**:
The top or bottom edge of a fractal, where one fractal gives way to
another. _Avoid_: zoom limit, edge

**Transition**:
The visual effect shown while the view crosses from one fractal to
another. Only boundary crossings involve transitions.
_Avoid_: effect, warp, animation

**Transition schedule**:
The ordered list of transitions a pilot uses, together with its
traversal mode.
_Avoid_: playlist, transition list

**Traversal mode**:
How a transition schedule picks its next transition: single,
sequential, or random.
_Avoid_: playback mode, cycle mode

**Destination policy**:
The rule that chooses the fractal a transition arrives at.
_Avoid_: next-fractal picker
