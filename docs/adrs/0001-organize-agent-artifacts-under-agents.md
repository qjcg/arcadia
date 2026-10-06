# 0001. Organize agent artifacts under `agents/` with generated plugin packages

## Status

Accepted (2026-10-05)

## Context and Problem Statement

Arcadia contains five first-party Agent Skills (formerly in `skills/`) and wants to
add distributable [Agent Plugins](https://agent-plugins.org/specification) v1.0.0,
starting with a packaging of [mattpocock/skills](https://github.com/mattpocock/skills).
We must decide where skills and plugins live in this polyglot repo, what is
hand-authored versus generated, and how the tracked tree relates to the gitignored,
machine-local `.agents/` directory. The wrong layout either breaks skill discovery
conventions or makes plugin packages drag the entire repo into a client's plugin store.

## Decision Drivers

* Agent Plugins v1.0.0 discovers skills only at `<plugin-root>/skills/<name>/SKILL.md`
  (flat, immediate children, no recursion).
* Plugin installers copy the whole plugin root — everything under it ships to users.
* Minimize churn: existing `skills/` paths were referenced in only two Taskfile lines.
* Arcadia is a polyglot playground; agent artifacts will grow beyond skills
  (e.g. MCP servers, prompts/commands).
* Single source of truth — no duplicated, hand-maintained skill trees.

## Considered Options

1. Root-level `skills/` + `plugins/`
2. Umbrella `agents/skills/` + `agents/plugins/`
3. Repo root as the plugin root (root `plugin.json`)

## Decision Outcome

Chosen option: **"Umbrella `agents/skills/` + `agents/plugins/`"**, because it groups
all agent-facing artifacts under one namespace with room to grow, and the migration
cost is near zero.

* `agents/skills/` is the hand-authored source of truth (moved from `skills/`),
  one skill per `agents/skills/<name>/SKILL.md`, conforming to the
  [Agent Skills specification](https://agentskills.io/specification).
* `agents/plugins/<name>/` holds **generated** Agent Plugin packages. Each is a
  self-contained plugin root (`plugin.json` + `skills/<name>/SKILL.md`), produced by
  `task plugins:sync` from source trees:
  * first-party skills: copied from `agents/skills/`;
  * vendored skills (e.g. `mattpocock-skills`): copied from a pinned upstream git ref,
    with only the promoted set included.
* Plugin trees are **copied, never symlinked**: plugin installers do not reliably
  preserve symlinks (upstream `mattpocock/skills` ADR-0002 documents installers
  dropping symlinks, leaving skills empty).
* `.agents/` remains gitignored, machine-local harness state. It is never committed
  and never a build input.

### Consequences

* Good, because the repo root stays uncluttered and future artifact types slot into
  `agents/` without another reshuffle.
* Good, because generated plugin trees keep one source of truth while shipping
  spec-conformant flat layouts.
* Neutral, because `agents/` versus `.agents/` requires a loud documented boundary
  (in `AGENTS.md` and the README).
* Bad, because `skills/` at the repo root was a recognizable ecosystem convention;
  moving it trades familiarity for consistency.

### Confirmation

* `git mv skills agents/skills` is a pure rename; only the two
  `go tool sv changelog … --exclude` references needed updating (they match module
  paths by subtree prefix, so `agents/skills` still excludes the `godog-examples`
  modules nested under the moved tree).
* `task plugins:validate` passes on every `agents/plugins/*` package.
* No tracked file references `skills/` at the repo root.

## Pros and Cons of the Options

### Root-level `skills/` + `plugins/`

* Good, because `skills/` matches the ecosystem's most recognizable convention.
* Good, because zero migration.
* Bad, because agent artifacts scatter across the root as they multiply.

### Repo root as the plugin root

* Good, because zero duplication for shipping arcadia's own skills.
* Bad, because installing the plugin would copy `exp/`, `talks/`, `gym/`, and Go
  source — the entire playground — into a client's plugin store.

## Plugin package contract (normative for `agents/plugins/`)

* Manifest: `plugin.json` per Agent Plugins v1.0.0 — required `$schema`
  (`https://agent-plugins.org/schemas/1.0.0/plugin.schema.json`) and `name`; no inline
  component configuration (no `skills` arrays); unknown top-level fields forbidden.
* Skills: discovered at `skills/<name>/SKILL.md`; `name` matches its directory;
  `description` ≤ 1024 characters.
* Vendored skills are flattened from upstream bucket layouts
  (e.g. `skills/{engineering,productivity}/<name>/` → `skills/<name>/`) at sync time.
* Non-standard frontmatter fields (e.g. `disable-model-invocation`, `argument-hint`)
  are normalized into the spec's `metadata` map as string values.
* `mcp.json` is omitted when a package has no MCP servers; a missing fixed component
  location is valid.
* Validation: `plugin.json` against `schemas/1.0.0/plugin.schema.json`, every skill
  through `skills-ref validate`, and every plugin-relative path checked to resolve
  within the plugin root.
