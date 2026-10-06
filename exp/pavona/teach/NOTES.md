# Notes

## Workspace
- Teaching workspace lives at `exp/pavona/teach/` (user's explicit choice, 2026-06).
- Lessons link repo sources with relative `../../docs/...` / `../../README.md` paths.

## User preferences
- Expressed mission (2026-06): use built-in templates + build/maintain custom templates for
  personal projects. Excluded Pavona internals/contribution for now.
- Experienced Go developer; skip Go/text/template fundamentals.

## Teaching notes
- All Pavona facts must be verified against `exp/pavona` source (README, docs/design.md,
  internal/) — the engine is young and README wording can drift from behavior. Verified so far:
  positional `[name]` sets both output dir and `project_name` prefill (internal/cli/template.go);
  quiet mode uses defaults then fails required-validation (internal/scaffold/prompt.go).
