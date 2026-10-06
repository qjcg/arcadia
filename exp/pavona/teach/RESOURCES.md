# Pavona Resources

## Knowledge

- [Pavona README](../README.md)
  The canonical quick reference: installation, commands, built-in templates, template file
  rules, and resolution order. Use for: any "how do I invoke X" question.
- [Pavona design doc](../docs/design.md)
  Full architecture: `config.cue` format, how CUE types map to prompts, prompt flow,
  hydration algorithm, error handling. Use for: anything about *why* Pavona behaves as it
  does, and the exact semantics of variables and rendering.
- [Pavona implementation: `scaffold.go` and `config.go`](../internal/scaffold/scaffold.go)
  Current behavior for resolution, CUE parsing, path/content rendering, and hydration. Use for:
  edge cases or discrepancies where the README/design doc may be out of date; config parsing
  details are in sibling `config.go`.
- [Built-in templates as examples](../templates/)
  `tool/`, `app/`, `site/`, `pavona/`, … are worked examples of real templates. Use for:
  modelling custom template structure on something proven.
- [CUE documentation — language basics](https://cuelang.org/docs/concepts/logic/)
  Upstream CUE docs. Use for: `config.cue` syntax questions beyond what Pavona's design doc
  covers (disjunctions, defaults with `*`, optional fields with `?`).
- Go stdlib [`text/template`](https://pkg.go.dev/text/template)
  The rendering engine for `*.tmpl` files. Use for: template syntax inside `.tmpl` files
  (variables, pipelines, conditionals).

## Wisdom (Communities)

- [qjcg/arcadia issue tracker](https://github.com/qjcg/arcadia/issues)
  Pavona is developed here; bugs and feature discussion happen as GitHub issues. Use for:
  reporting Pavona bugs and seeing how the tool is evolving.
- Local: ask the teaching agent (in these sessions)
  Use for: anything unclear, exercise feedback, and template review.

## Gaps

- No dedicated external Pavona community (it is a young in-repo tool). CUE's own community
  (CUE Slack / GitHub discussions at cuelang/cue) is the fallback for `config.cue` questions
  that the design doc does not answer.
