# TODO: Component overlays

Design and implement reusable components that can be overlaid onto a template, including components that can be invoked multiple times with different arguments.

## Design decisions to settle

- Component discovery and packaging: built-ins, explicit directory paths, and `$XDG_DATA_HOME/pavona/components/<name>` are the proposed sources.
- Component configuration: use `config.cue` for metadata and arguments; components are single-use by default and may opt into repetition.
- CLI selection: add repeatable `--component <name>` options, preserving invocation order.
- Argument scoping: keep component arguments separate from template variables, for example `.component.slug`; support an explicit per-invocation assignment syntax such as `--set component.page[0].slug=about`.
- Collision policy: allow shared directories, but reject distinct sources that produce the same file path. Never silently overwrite files.
- Output safety: validate arguments and rendered paths, detect conflicts, and avoid leaving partial output on failure.

## Implementation checklist

- [ ] Write an ADR covering the motivation, chosen model, alternatives, examples, and tradeoffs.
- [ ] Define and parse component metadata and arguments in CUE.
- [ ] Add component resolution and registration, matching template discovery conventions where appropriate.
- [ ] Add repeated component selection and per-invocation argument handling to `pavona new`.
- [ ] Prompt separately for each component invocation; in quiet mode, apply defaults and report missing required values.
- [ ] Render base template and component trees into one output with shared template variables and namespaced component arguments.
- [ ] Detect invalid paths and file collisions before writing the final output; ensure errors do not leave a partial project.
- [ ] Add tests for single-use and repeatable components, multiple invocations with distinct arguments, quiet mode, collision errors, and unchanged behavior when no components are selected.
- [ ] Document component authoring and CLI usage.
