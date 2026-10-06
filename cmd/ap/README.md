# ap

**Agent Plugin packages, generated and spec-conformant.**

`ap` (Agent Plugins) generates and validates
[Agent Plugins](https://agent-plugins.org/specification) v1.0.0 packages from
[Agent Skills](https://agentskills.io/specification) source trees. It is the
engine behind `agents/plugins/` in this repo: skills come from hand-authored
trees or pinned upstream refs, and `ap` turns them into portable,
installable plugin packages. See
[ADR 0001](../../docs/adrs/0001-organize-agent-artifacts-under-agents.md) for
the layout decisions and the package contract.

## Key Features

- **Declarative sync specs**: each generated package is described by a small
  YAML spec (`agents/plugins/specs/*.yaml`) naming its manifest metadata and
  its skill sources — a local tree or a git repository pinned to a ref.
- **Flattening**: upstream bucket layouts like `skills/engineering/<name>/`
  become the spec-mandated `skills/<name>/` (immediate children only).
- **Frontmatter normalization**: non-standard frontmatter fields (e.g.
  `disable-model-invocation`, `argument-hint`) move into the Agent Skills
  `metadata` map as strings; nested maps flatten to dot-separated keys.
- **Conformance validation**: manifest fields, name rules, description
  limits, `name`↔directory matching, symlink bans, and plugin-root
  containment for Markdown links.
- **Deterministic output**: re-running `sync` on the same inputs produces a
  byte-identical tree, so generated packages can be committed and reviewed.

## Quickstart

### Regenerate all plugin packages

```bash
task plugins:sync
# equivalent to:
go tool ap sync
```

### Validate skills and packages

```bash
task skills:validate    # agents/skills against the Agent Skills spec
task plugins:validate   # agents/plugins against the Agent Plugins spec

# or directly, against any package root or bare skill tree:
go tool ap validate agents/plugins/mattpocock-skills
go tool ap validate agents/skills
```

## How sync works

For each spec in `--specs` (default `agents/plugins/specs`), `ap sync`:

1. Materializes the source tree (shallow-clones `source.repo` at
   `source.ref` for git sources, or uses `source.path` for local ones).
2. Copies the skills listed in `source.skills` (paths relative to the source
   tree root) into `<out>/<name>/skills/<basename>/`, skipping unlisted
   skills — promotion is explicit. All files in each skill directory travel
   along (`scripts/`, `references/`, `assets/`, …).
3. Copies `source.copy-root` files (e.g. `LICENSE`, `CHANGELOG.md`) to the
   package root.
4. Normalizes every `SKILL.md` frontmatter into Agent Skills form.
5. Writes `plugin.json` from the spec's manifest fields.

Before and after normalization:

```markdown
---                                   ---
name: handoff                         name: handoff
description: Compact the current ...  description: Compact the current ...
argument-hint: "What next?"     ->    metadata:
disable-model-invocation: true            argument-hint: What next?
---                                       disable-model-invocation: "true"
                                      ---
```

Nothing else changes: the Markdown body is preserved byte for byte, and
symlinks are rejected rather than copied (plugin installers do not reliably
preserve them).

## Validation rules

`ap validate PATH` accepts a plugin package root (contains `plugin.json`), a
bare skills tree (children contain `SKILL.md`), or a directory of package
roots. Checks follow Agent Plugins v1.0.0 and the Agent Skills spec:

- `plugin.json`: canonical `$schema`, required `name`, name character rules,
  closed top-level field set, `author`/`keywords`/`extensions` types.
- `skills/<name>/SKILL.md`: `name` matches its directory, `description` is
  1–1024 characters, frontmatter fields are spec-permitted, `metadata` values
  are strings.
- Containment: no symlinks anywhere in the package; Markdown link targets
  must resolve within the plugin root (§4.1).
- `mcp.json` (if present): canonical `$schema`, closed top-level field set.

Violations are listed one per line; any violation fails the run.

## Adding a plugin package

1. Add `agents/plugins/specs/<name>.yaml` (copy an existing spec).
2. Run `task plugins:sync` and `task plugins:validate`.
3. Commit the spec and the generated package together.

Never hand-edit the generated trees under `agents/plugins/<name>/`; they are
wiped and rebuilt by `sync`.

## Testing

Unit tests cover parsing, normalization, and validation edge cases; CLI
behavior is covered by [testscript](https://pkg.go.dev/github.com/rogpeppe/go-internal/testscript)
scenarios in [`testdata/`](testdata):

```bash
go test ./...
```
