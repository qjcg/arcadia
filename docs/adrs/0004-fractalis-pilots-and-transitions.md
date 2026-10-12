# 0004. Fractalis: pilots, transitions, and transition schedules

## Status

Accepted (2026-10-13)

## Context and Problem Statement

Fractalis grew three movement behaviors (auto-zoom, vantage tour,
manual) and four transition effects as independent, parallel code —
each effect struct re-implements selection logic, each frontend
re-implements the behaviors, and the transition trigger, destination
choice, and visual effect are bundled in one path. We are
introducing the **Pilot** and **Transition** concepts and must decide
where transition *selection* lives and what events count as
transitions at all.

## Decision

A **pilot** owns two things: how the view moves each tick, and its
**transition schedule** — a declared list of transitions plus a
traversal mode (single, sequential, or random). The schedule is plain
data executed by one shared scheduler; pilots never implement their
own selection logic. Configuration precedence: URL beats pilot
default; a session override (`T`) applies on top.

A **transition** is defined narrowly: the visual effect played while
the view crosses a **boundary** (top or bottom of a fractal). Other
view changes — notably vantage scene switches — are the pilot moving,
not transitions.

## Considered Options

- Letting each pilot pick transitions itself (rejected: the current
  per-effect selection logic is exactly the duplication we are
  removing).
- Defining transitions as *any* change of fractal view (rejected:
  over-broad; vantage scene switches would force every motion to be
  an effect, and "effect during boundary crossing" is the concept
  users described).

## Consequences

- New effects and new pilots plug into one registry each; a pilot can
  never misbehave relative to the scheduler.
- "Random" is a traversal mode, not an effect.
- Schedules must round-trip through `fractal://` URLs losslessly,
  pinning the URL grammar (`pilot=`, `pilot_*`, `transition=`).
