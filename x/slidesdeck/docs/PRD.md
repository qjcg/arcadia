# Product Requirements Document (PRD) - Slidesdeck

## 1. Product Vision (Elevator Pitch)
For developers and technical writers who need a fast, portable way to transform Markdown or Org-mode notes into professional slideshows, Slidesdeck is a single-purpose CLI slideshow creator that generates self-contained, interactive HTML presentations with built-in break management. Unlike complex office suites or heavy converters like Pandoc, our product offers native Go performance, zero external dependencies, and first-class Org-mode support.

## 2. Product Overview

### 1.1 Problem Statement
Technical presentations are often created using either complex GUI
tools like PowerPoint/Keynote or heavy, complex CLI tools like Pandoc
that require significant setup. Developers and technical writers who
work primarily in Markdown or Emacs Org-mode need a simple,
single-purpose CLI tool to convert their notes into lightweight,
self-contained HTML slideshows.

### 1.2 Target Audience
- **Developers & Tech Writers**: Who prefer using their own text editors (VS Code, Emacs, Vim) and want to keep presentations under version control.
- **Conference Speakers**: Who need highly portable presentations that don't rely on internet connectivity or external assets.
- **Educators/Trainers**: Who need a fast way to turn lesson notes into visual aids.

### 1.3 Goals
- **Simplicity**: No complex configuration; one command to convert.
- **Portability**: Generate a single HTML file with CSS embedded. No need to ship images separately if using base64 (future feature) or absolute URLs.
- **Standardization**: Support the most common presentation separators in both Markdown and Org-mode.
- **Developer-Experience (DX)**: Provide clear errors and a fast feedback loop.

### 1.4 Success Metrics
- **Ease of Use**: A new user can generate their first presentation in under 3 minutes.
- **Efficiency**: Conversion of average slide decks (20-40 slides) in less than 200ms.
- **Self-Containment**: Output HTML requires zero external JS/CSS dependencies to render core styles.

## 2. Market Context & User Personas

### 2.1 Competitive Landscape
- **Pandoc**: The "Swiss Army Knife", powerful but complex with many flags and external dependencies (Lua filters, custom templates).
- **reveal-md**: Specifically for reveal.js; requires Node.js and a large `node_modules` folder.
- **Marp**: Great Markdown-based slideshow creator, but lacks native Org-mode support and can be heavy as an extension.
- **Slidesdeck Differentiation**: Native Go binary (fast, easy install), supports BOTH Markdown and Org-mode, produces ultra-lightweight vanilla HTML/CSS slides (not reliant on a framework unless requested).

### 2.2 User Personas
- **"The Org Oracle"**: An Emacs power user who has all their notes in `.org` files and wants to present them without switching to another format.
- **"The Markdown Minimalist"**: A developer on macOS or Linux who writes technical documentation in Markdown and wants to present it at a local meetup using a simple terminal command.

## 3. Product Features & Roadmap

### 3.1 Version 1.0 (MVP)
- **Multi-format Support**: Markdown (CommonMark) and Org-mode.
- **Self-contained Output**: Single HTML file with embedded Tailwind CSS (v4) and daisyUI (v5) styling. CSS and JS are bundled and optimized via **`esbuild`**.
- **Modular Design**: CSS themes are stored in separate files and bundled as needed.
- **Embedded Assets**: All optimized frontend resources are embedded into the Go binary for true single-executable portability.
- **Robust Testing**: Comprehensive CLI testing using `testscript` to ensure reliable file conversion and CLI behavior.
- **Interactivity**: Alpine.js (v3) for navigation (n/p/arrows/space), first/last slide (Shift+Alt+,/.), toggles (t for themes, f for fullscreen, N for line numbers, ? for help, / for search), and Shift+P for pause.
- **Search Feature**: Full-text slide search using `flexsearch` via a Command Palette interface.
- **Theme Support**: Command Palette based theme switching (using `flexsearch`). All daisyUI themes are included and categorized as `light:` or `dark:`. Default theme can be set via CLI flag.
- **Pause/Break Mode**: A visually impressive, interactive break screen (triggered by `Shift+P`) with a configurable countdown timer and custom message. The state is persisted in Local Storage to survive browser restarts.
- **Basic Styling**: A clean, readable default layout using daisyUI components.
- **Slide Separators**: Default separation by first-level headings (`#` or `*`). Support for explicit separators (e.g., `---`) and custom strings.
- **Code Highlighting**: High-quality syntax highlighting using **`chroma`**. Line numbers are displayed by default and can be toggled via keyboard.
- **Image Support**: Conversion of standard image syntax to `<img>` tags.

### 3.2 Version 2.0 (Planned)
- **Live Reload/Watch**: Automatically re-generate HTML when the input file changes.
- **Built-in Presentation Mode**: Keyboard controls (Arrow keys, Space) to navigate slides in the browser via minimal JS.
- **Custom Templates**: User-provided HTML templates.
- **Image Embedding**: Automatically base64-encode local images into the HTML file for true self-containment.

### 3.3 Out of Scope (For Now)
- **PDF Export**: Users can "Print to PDF" from their browser.
- **Animations & Transitions**: Fixed simple layouts for speed and portability.
- **Server Deployment**: This is a local CLI tool, not a hosting platform.

## 4. User Experience (UX)

### 4.1 Key Workflows
1. **The Default Conversion**: User runs `slidesdeck talk.md`. The tool creates `talk.html` in the same directory.
2. **Custom Output**: User runs `slidesdeck -o presentation.html docs/talk.md`. The tool creates `presentation.html` in the current working directory (regardless of the input file's path).
3. **Format Detection**: The tool automatically handles `.md` as Markdown and `.org` as Org-mode.

### 4.2 Error Scenarios
- **Unknown Input Format**: The tool should warn the user if it can't determine the format from the file extension or content.
- **Empty Output**: If the file contains no separators, the tool should suggest how to add them.

## 5. Prioritization (MoSCoW)

- **Must Have**: Markdown/Org parsing, CSS embedding, CLI interface, code blocks.
- **Should Have**: Basic navigation JS (Arrow keys), slide numbers, version/help commands.
- **Could Have**: Base64 image embedding, watch mode, custom templates.
- **Won't Have (v1)**: PDF generation, hosted version, 3D transitions.

## 6. Technical Assumptions
- Developed in Go (following arcadia/AGENTS.md).
- Frontend: Alpine.js (v3), Tailwind CSS (v4), daisyUI (v5).
- Templating: `templ` (external Go tool).
- Hot Reload: `air` (external Go tool).
- Parsers: `goldmark` for Markdown, `github.com/niklasfasching/go-org` for Org-mode, and `chroma` for syntax highlighting.

---
**Prepared by**: Crush (AI Assistant)
**Date**: February 7, 2026
