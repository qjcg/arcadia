# Mission: Pavona — template engine for real projects

## Why
Use Pavona's built-in templates to scaffold personal Go projects quickly, and build
and maintain custom Pavona templates so personal project setups stop being manual
repetition. Scaffolding should take seconds and stay consistent as projects evolve.

## Success looks like
- Scaffold a personal project from a built-in template (`tool`, `app`, `tui`, …) without
  consulting the README, and know which template fits which project shape.
- Create a custom template from scratch: `config.cue` variables, rendered `.tmpl` files,
  verbatim files, templated directory names.
- Maintain a custom template over time: change variables, add files, version it, and
  re-generate projects without breaking existing ones.
- Diagnose the common failure modes (template not found, CUE validation errors, quiet-mode
  defaults) unaided.

## Constraints
- Workspace lives at `exp/pavona/teach/` inside the arcadia repo.
- User is an experienced Go developer (repo contributor) — no need to teach Go, `text/template`,
  or basic CLI usage.
- Prefer short lessons with hands-on wins over long theory.

## Out of scope
- Contributing to / modifying the Pavona engine internals (yet) — usage only.
- CUE language deep dive — only the subset `config.cue` actually needs.
