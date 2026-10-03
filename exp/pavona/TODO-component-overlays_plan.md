# Plan: Component overlays

## Purpose

Introduce reusable components that can be layered onto a base template and invoked repeatedly with different arguments. Examples include adding a README to a template that lacks one, or generating multiple documentation pages from a `page` component.

## Proposed design

- Components are independently discoverable packages, not files that must be copied into each template.
- Resolve components by built-in name, explicit directory path, or `$XDG_DATA_HOME/pavona/components/<name>`, following the existing template-resolution pattern.
- A component has a `config.cue` and a file tree. Its configuration defines metadata, arguments, and an optional `repeatable` setting that defaults to false.
- Add repeatable `--component <name>` options to `pavona new`; preserve selection order.
- Keep component arguments namespaced from base-template variables. Component templates receive the base variables plus their invocation's values under a `.component` namespace.
- Support explicit per-invocation values with a syntax such as `--set component.page[0].slug=about`. Interactive use prompts separately for each selected invocation; quiet mode uses defaults and errors if a required value is missing.
- Render component files into the same output root as the base template. Shared directories are allowed, but two sources producing the same file path are an error rather than an overwrite.
- Validate values and rendered paths, detect collisions before finalizing output, and avoid leaving a partial generated project on failure.
- With no components selected, existing template generation behavior remains unchanged.

## Example usage

```sh
pavona new foo project --component readme
pavona new docs-site site --component page --component page \
  --set component.page[0].slug=about \
  --set component.page[1].slug=gallery
```

The repeated `page` component invocation should prompt for its own arguments separately in interactive mode. In quiet mode, required arguments must be supplied explicitly or generation fails with a clear error.

## Work sequence

1. Write an ADR describing the motivation, decision, examples, alternatives, tradeoffs, and implementation/testing implications.
2. Define the component package format and CUE schema, including arguments and repeatability rules.
3. Implement component discovery and resolution for built-ins, paths, and XDG data.
4. Extend `pavona new` to accept repeated component selections and associate values with individual invocations.
5. Add interactive prompting, defaults, quiet-mode validation, and clear errors for missing or invalid values.
6. Extend rendering to apply base and component trees in order with isolated component argument namespaces.
7. Add preflight validation for rendered paths and file collisions; ensure generation failures do not leave partial output.
8. Add unit and integration coverage, then document authoring and CLI usage.

## Acceptance criteria

- A template can opt into one or more components without modifying the template's own files.
- A component declared non-repeatable cannot be selected more than once; a repeatable component can be selected multiple times.
- Distinct invocations of a repeated component render using their own argument values.
- Base and component variables cannot silently shadow one another.
- Shared directories do not conflict, but duplicate output file paths fail clearly and deterministically.
- Missing required values and invalid output paths fail without leaving partial output.
- Existing generation behavior is unchanged when no components are selected.
- Tests cover single-use and repeatable components, multiple invocation arguments, quiet mode, collision errors, and no-component behavior.

## Risks and alternatives to record in the ADR

- Template-local-only components reduce cross-template reuse; independently discoverable components are preferred.
- Implicit overwrite ordering makes output depend on application order and can hide mistakes; collisions should be errors.
- Flattening component values into the base variable map risks accidental shadowing; a separate namespace is preferred.
- Atomic output handling and path normalization need careful implementation to prevent partial output and paths escaping the output root.
- CLI syntax for per-invocation values and the exact CUE schema should be finalized in the ADR before implementation.
