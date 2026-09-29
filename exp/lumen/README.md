# Lumen

> One calm, semantic token language for web and terminal interfaces.

Lumen is a design system for developer tools and the interfaces around them. It defines shared design tokens and interaction principles for browser-based and command-line experiences, so both surfaces communicate the same meaning with a consistent visual language.

## Design principles

- **Semantic tokens:** names such as `primary`, `danger`, and `surface` express purpose rather than fixed colors.
- **Consistent across surfaces:** the same token vocabulary maps to CSS and terminal rendering.
- **Quiet and legible:** hierarchy comes from structure, typography, and whitespace before color.
- **Accessible by default:** status and errors remain understandable without color, and focus is visible on both surfaces.
- **Respect the environment:** terminal output degrades across color capabilities and honors `NO_COLOR`, `TERM=dumb`, and non-TTY output.

## Surfaces

| Surface | Rendering approach |
| --- | --- |
| Web | CSS custom properties, typography, spacing, component states, and light/dark modes |
| Terminal | ANSI colors with graceful depth fallbacks, emphasis, glyphs, and text layout |

Both surfaces share tokens for color roles, typography, spacing, shapes, and components. The specification covers common elements such as actions, inputs, cards, tables, statuses, alerts, and progress indicators.

## Specification

Lumen uses `DESIGN.md` as its design system document, following the format and intended use described by Google Stitch. Its YAML frontmatter provides machine-readable design tokens, while the Markdown sections explain the design rationale and guidance for applying them. This makes the file useful as a shared source of truth for people and design-generation tools.

- [Lumen DESIGN.md](docs/DESIGN.md)
- [Google Stitch: DESIGN.md overview](https://stitch.withgoogle.com/docs/design-md/overview)
- [Google Stitch: DESIGN.md specification](https://stitch.withgoogle.com/docs/design-md/specification/)
