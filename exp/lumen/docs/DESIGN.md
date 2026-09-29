---
version: alpha
name: Lumen
description: One calm, semantic token language that renders to both web and terminal surfaces.

colors:
  surface: "#FBFAF7"
  surface-raised: "#FFFFFF"
  surface-sunken: "#F2F0EB"
  on-surface: "#17191C"
  on-surface-muted: "#6C7278"
  border: "#E4E1DA"
  border-strong: "#C9C5BB"
  primary: "#2E6F95"
  primary-hover: "#255A78"
  primary-soft: "#E4EEF4"
  accent: "#C9772E"
  focus: "#2E6F95"
  danger: "#B8422E"
  danger-soft: "#F7E6E2"
  warning: "#8A6414"
  success: "#2F7D5B"
  success-soft: "#E3F0E9"

typography:
  display:
    fontFamily: Fraunces, Newsreader, Georgia, serif
    fontSize: 40px
    fontWeight: 500
    lineHeight: 1.12
    letterSpacing: -0.02em
  title:
    fontFamily: Inter, system-ui, sans-serif
    fontSize: 22px
    fontWeight: 600
    lineHeight: 1.3
    letterSpacing: -0.01em
  body:
    fontFamily: Inter, system-ui, sans-serif
    fontSize: 15px
    fontWeight: 400
    lineHeight: 1.6
    letterSpacing: 0
  label:
    fontFamily: Inter, system-ui, sans-serif
    fontSize: 13px
    fontWeight: 500
    lineHeight: 1.4
    letterSpacing: 0.01em
  mono:
    fontFamily: JetBrains Mono, Berkeley Mono, ui-monospace, monospace
    fontSize: 13px
    fontWeight: 400
    lineHeight: 1.5
    letterSpacing: 0

rounded:
  sm: 4px
  md: 8px
  lg: 14px
  pill: 9999px

spacing:
  xs: 4px
  sm: 8px
  md: 16px
  lg: 24px
  xl: 40px

components:
  button-primary:
    backgroundColor: "{colors.primary}"
    textColor: "{colors.surface-raised}"
    rounded: "{rounded.md}"
    padding: "{spacing.sm} {spacing.md}"
    typography: label
  button-primary-hover:
    backgroundColor: "{colors.primary-hover}"
  button-ghost:
    backgroundColor: transparent
    textColor: "{colors.primary}"
    rounded: "{rounded.md}"
    padding: "{spacing.sm} {spacing.md}"
    typography: label
  input:
    backgroundColor: "{colors.surface-raised}"
    textColor: "{colors.on-surface}"
    rounded: "{rounded.sm}"
    padding: "{spacing.sm} {spacing.md}"
    borderColor: "{colors.border-strong}"
  card:
    backgroundColor: "{colors.surface-raised}"
    textColor: "{colors.on-surface}"
    rounded: "{rounded.lg}"
    padding: "{spacing.lg}"
  table:
    textColor: "{colors.on-surface}"
    borderColor: "{colors.border}"
    padding: "{spacing.sm}"
  badge:
    backgroundColor: "{colors.primary-soft}"
    textColor: "{colors.primary}"
    rounded: "{rounded.pill}"
    padding: "{spacing.xs} {spacing.sm}"
  alert-danger:
    backgroundColor: "{colors.danger-soft}"
    textColor: "{colors.danger}"
    rounded: "{rounded.md}"
    padding: "{spacing.md}"
  alert-success:
    backgroundColor: "{colors.success-soft}"
    textColor: "{colors.success}"
    rounded: "{rounded.md}"
    padding: "{spacing.md}"

x-modes:
  dark:
    colors:
      surface: "#14161A"
      surface-raised: "#1B1E23"
      surface-sunken: "#0F1114"
      on-surface: "#E8E7E2"
      on-surface-muted: "#9AA0A6"
      border: "#2A2E34"
      border-strong: "#3C424A"
      primary: "#7FB4D6"
      primary-hover: "#9AC5E0"
      primary-soft: "#1E2B34"
      accent: "#E0A458"
      focus: "#7FB4D6"
      danger: "#E08A78"
      danger-soft: "#33211E"
      warning: "#D8B45E"
      success: "#7FB89A"
      success-soft: "#1B2A22"

x-terminal:
  color-depth: [truecolor, 256, 16, none]
  background: auto
  width: 80
  indent: 2
  glyphs:
    border-h: "─"
    border-v: "│"
    corner-tl: "╭"
    corner-tr: "╮"
    corner-bl: "╰"
    corner-br: "╯"
    bullet: "•"
    arrow: "▸"
    check: "✓"
    cross: "✗"
    warn: "!"
    ellipsis: "…"
    spinner: "⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏"
    bar-full: "█"
    bar-empty: "░"
  emphasis:
    strong: bold
    muted: dim
    focus: reverse
    link: underline
  ansi:
    primary: "38;2;46;111;149"
    primary-hover: "38;2;37;90;120"
    accent: "38;2;201;119;46"
    focus: "38;2;46;111;149"
    danger: "38;2;184;66;46"
    warning: "38;2;138;100;20"
    success: "38;2;47;125;91"
    muted: "38;2;108;114;120"
    border: "38;2;201;197;187"
    on-surface: "38;2;23;25;28"
---

# Lumen

*A calm, semantic token language that renders the same design to a web page and a terminal.*

## Overview

Lumen is a design system for tools: the web frontends and the command line interfaces
that surround them. It is built on one idea: **tokens are semantic, renderers are
interchangeable.** A token named `danger` means the same thing in a Go CLI and a
templ component. One file, `DESIGN.md`, is the source of truth; two renderers turn it
into CSS custom properties and ANSI escape sequences.

**Personality.** Calm, precise, luminous. Warm paper against cool ink. Generous
whitespace. A single accent, spent rarely. Nothing decorative that does not also
inform.

**Audience.** Developers and operators who read dense output all day, on a laptop, over
SSH, and in a browser. The system assumes both.

## Laws

These are non-negotiable. When a rule and a Law conflict, the Law wins.

1. **Legible in monochrome.** Hierarchy must survive `NO_COLOR`, a 16-color terminal,
   and a black-and-white printout. Color is emphasis, never the only signal.
2. **One accent, spent rarely.** A second accent is a bug in the information design.
3. **Whitespace is a material.** Blank lines in the terminal, margins on the web. Both
   are load-bearing.
4. **Structure before style.** Size, weight, position, and order carry hierarchy long
   before color does.
5. **Quiet by default, detailed on request.** Show the answer; reveal the reasoning
   behind `--verbose`.
6. **Same words, both surfaces.** A token name has one meaning across CSS and ANSI.
7. **Nothing extra.** Every token, glyph, and rule earns its place or is deleted.
8. **Calm errors.** Errors say what happened, why, and what to do next. Never a stack
   trace.

## Surfaces

Lumen defines a **surface** as a rendering target with a known character budget and a
known color capability. Two surfaces ship today: `web` and `terminal`. Both consume the
same tokens; only the renderer differs.

| Token group | Web renderer                     | Terminal renderer                        |
|-------------|----------------------------------|------------------------------------------|
| colors      | CSS custom properties, dark mode | SGR codes, depth-degraded                |
| typography  | font stacks, size, weight        | weight and case only (one font exists)   |
| rounded     | `border-radius`                  | corner glyph set                         |
| spacing     | rem/px scale                     | blank lines, cell padding, indent        |
| components  | templ/HTML elements              | prompt, table, list, progress, alert     |

### Degradation ladder

Color depth is authored at the top and degrades automatically. Never author below the
top rung; the renderer walks down.

| Depth     | Source                                   | Fallback rule                                  |
|-----------|------------------------------------------|------------------------------------------------|
| truecolor | exact 24-bit RGB from `x-terminal.ansi`  | primary target                                 |
| 256       | nearest xterm-256 index                  | RGB cube + grayscale ramp, pick nearest        |
| 16        | nearest classic ANSI                     | preserve role, accept hue drift                |
| none      | emphasis + glyphs only                   | Law 1 applies; meaning must still be readable  |

Color is disabled when `NO_COLOR` is set, when `TERM=dumb`, or when stdout is not a
TTY. Structured output (`--json`, `--output yaml`) is always uncolored.

## Colors

Tokens are roles, not hues. The renderer chooses the value.

| Role              | Meaning                                | Paired signal  |
|-------------------|----------------------------------------|----------------|
| `surface`         | page or terminal background            | none           |
| `on-surface`      | primary text                           | none           |
| `on-surface-muted`| secondary text, metadata, timestamps   | none           |
| `border`          | hairline dividers                      | `─` / `│`      |
| `primary`         | interactive, selected, links           | `▸` / underline|
| `accent`          | one rare highlight, never for status   | noun marker    |
| `focus`           | keyboard focus                         | reverse video  |
| `danger`          | failure, destructive action            | `✗`            |
| `warning`         | caution, degraded state                | `!`            |
| `success`         | completion, healthy state              | `✓`            |

**Contrast.** Text meets WCAG AA (4.5:1 body, 3:1 large text and graphical objects)
in both light and dark modes. Terminal color is only applied against both an assumed
black and an assumed white background; if either fails, fall to the next depth rung
rather than emit an unreadable role.

**Rule.** A semantic color never travels alone. Every status token has a glyph, a word,
or both. This is what makes the 16-color and `NO_COLOR` rungs survivable.

## Typography & Emphasis

The web surface has faces; the terminal surface has weight. Both encode the same
hierarchy, so a heading reads as a heading in either place.

| Level     | Web                              | Terminal                    |
|-----------|----------------------------------|-----------------------------|
| `display` | Fraunces/newreader, 40/1.12      | bold, uppercase, blank line |
| `title`   | Inter 22/1.3, 600                | bold                        |
| `body`    | Inter 15/1.6, 400                | plain                       |
| `label`   | Inter 13/1.4, 500, +0.01em       | dim or bold (never both)    |
| `mono`    | JetBrains Mono 13/1.5            | the native font             |

Emphasis is a closed set: `strong` (bold), `muted` (dim), `focus` (reverse),
`link` (underline). Do not combine more than two on one span. Italic is reserved for
citations, not emphasis.

## Layout & Spacing

The web scale is the terminal scale measured in real units.

| Token | Web  | Terminal                        |
|-------|------|---------------------------------|
| `xs`  | 4px  | padding within a cell group     |
| `sm`  | 8px  | one blank line between items    |
| `md`  | 16px | one blank line between blocks   |
| `lg`  | 24px | three blank lines, section break |
| `xl`  | 40px | not used; terminal sections breathe via `lg` |

**Terminal geometry.** Default width is 80 columns; respect `COLUMNS` when present.
Target widths are 72, 80, 100, and 120. Content wraps; tables truncate middle with
`…` rather than overflow. Indent is 2 spaces per nesting level, maximum depth 3;
beyond that, stop nesting and use a heading.

**Web geometry.** Single-column reading measure of 60 to 75 characters. Grid gutters
take `md`; section spacing takes `xl`.

## Elevation & Depth

The web expresses depth with shadow; the terminal cannot. Depth in a terminal is
structural, not optical.

| Level | Web                          | Terminal                                  |
|-------|------------------------------|-------------------------------------------|
| 0     | flat, `surface`              | no border                                 |
| 1     | raised card, soft shadow     | `╭─╮` single border, no fill              |
| 2     | popover, stronger shadow     | nested border + `muted` label             |
| 3     | modal, backdrop              | full-screen pane, reverse title bar       |

Shadows are soft and low-contrast; never stack more than two elevations in view.
Terminal depth never invents a background fill, because the terminal background is
borrowed, not owned.

## Shapes & Glyphs

| Token  | Web radius | Terminal corners         |
|--------|------------|--------------------------|
| `sm`   | 4px        | sharp, minimal framing   |
| `md`   | 8px        | `╭ ╮ ╰ ╯` rounded frame  |
| `lg`   | 14px       | not used                 |
| `pill` | 9999px     | not used                 |

One border weight. A single box-drawing character set is chosen per block (`light`,
`heavy`, or `double`), never mixed inside one frame. Glyphs live in
`x-terminal.glyphs`; the check, cross, arrow, bullet, and ellipsis are the only
ornamental characters permitted in output.

## Components

Every component has a web form and a terminal form that carry the same semantics.

| Component       | Web                       | Terminal                                   |
|-----------------|---------------------------|--------------------------------------------|
| Primary action  | `button-primary`          | `[run]` bracket action, `primary` + bold   |
| Secondary action| `button-ghost`            | dim bracket action                         |
| Input           | `input`                   | prompt line, `▸ ` marker, reverse cursor   |
| Container       | `card`                    | framed block or blank-line-delimited block |
| Data            | `table`                   | aligned columns, `─` rule under header     |
| Status          | `badge`                   | `✓` / `!` / `✗` prefix + word              |
| Alert           | `alert-danger/success`    | `✗ ` / `✓ ` first line, dim detail below   |
| Progress        | meter/bar                 | `bar-full`/`bar-empty` or `spinner`        |

**States.** Interactive components define the same four states on both surfaces:
rest, hover (web) or selected (terminal), focus, and disabled (dim, no color). Disabled
is expressed by `muted` plus the absence of the accent, never by removing the label.

## Accessibility & Interaction

- **Never color alone.** Status, selection, and errors carry a glyph or a word.
- **Contrast checked.** Both `x-modes` palettes are validated at build time; a failing
  token fails the build.
- **Respect the environment.** Honor `NO_COLOR`, `TERM=dumb`, non-TTY stdout, and
  `--plain`. Honor `prefers-reduced-motion` by replacing the spinner with a static
  line and step counter.
- **Focus is visible.** Web focus draws the `focus` ring; terminal focus uses reverse
  video so it survives every color depth.
- **Machine output is a contract.** `--json` is stable, complete, and uncolored.
  Human output is never the only way to get data.

## Do's and Don'ts

**Do**

- Reach for weight, spacing, and order before reaching for color.
- Keep one accent on screen at a time.
- Let the terminal be quiet: white space is the primary device.
- Pair every status color with `✓`, `!`, or `✗`.
- Truncate with `…` and align columns on the decimal or the left edge.

**Don't**

- Don't encode meaning in hue alone; it dies at the 16-color rung.
- Don't nest terminal frames more than twice.
- Don't introduce a new color, glyph, or radius without a token.
- Don't use italic, blink, or overline in terminal output.
- Don't print raw stack traces, ANSI color in `--json`, or unaligned tables.
