# Software Requirements Specification (SRS) - Slidesdeck

## 1. Introduction

### 1.1 Purpose
For developers and technical writers who need a fast, portable way to transform Markdown or Org-mode notes into professional slideshows, Slidesdeck is a single-purpose CLI slideshow creator that generates self-contained, interactive HTML presentations with built-in break management. Unlike complex office suites or heavy converters like Pandoc, our product offers native Go performance, zero external dependencies, and first-class Org-mode support.

This document specifies the requirements for `slidesdeck`, a command-line tool designed to convert Markdown and Org-mode files into self-contained HTML slideshows.

### 1.2 Scope
`slidesdeck` is a CLI tool that parses structured text files and generates HTML5 presentations. It supports basic Markdown (CommonMark) and Org-mode syntax, using specific separators to define slide boundaries.

## 2. Overall Description

### 2.1 Product Perspective
`slidesdeck` is a standalone utility written in Go. It does not require a web server to host the generated slides, as all styling and structure are embedded in the output HTML file.

### 2.2 Product Functions
- Parse Markdown files.
- Parse Org-mode files.
- Split content into slides based on configurable separators.
- Generate valid HTML5 output with embedded CSS.
- Support common formatting: headings, lists, code blocks, images, links.

### 2.3 User Classes and Characteristics
Target users are developers, technical writers, and presenters who prefer text-based workflows and command-line tools.

### 2.4 Operating Environment
- Compatible with Linux, macOS, and Windows.
- No external runtime dependencies (self-contained binary).

## 3. Specific Requirements

### 3.1 Functional Requirements

#### 3.1.1 General
**FR-1**: The tool shall accept a single input file as a positional argument.
**FR-2**: The tool shall detect the input format based on the file extension: `.md` for Markdown and `.org` for Org-mode.
**FR-3**: By default, the tool shall generate an output file in the same directory as the input file, with the same base name and the `.html` extension (e.g., `path/to/slides.md` -> `path/to/slides.html`).

#### 3.1.2 Markdown Parsing
**FR-4**: The tool shall support CommonMark headings (#, ##, etc.).
**FR-5**: The tool shall support bold (**text**) and italic (*text*) formatting.
**FR-6**: The tool shall support fenced code blocks with language identifiers.
**FR-7**: The tool shall support inline code (`code`).
**FR-8**: The tool shall support images (![alt](url)).
**FR-9**: The tool shall support hyperlinks ([text](url)).
**FR-10**: The tool shall support lists (ordered and unordered).
**FR-11**: The tool shall support blockquotes.

#### 3.1.3 Org-mode Parsing
**FR-12**: The tool shall support org-mode headings (*, **, etc.).
**FR-13**: The tool shall support bold (*text*) and italic (/text/) formatting.
**FR-14**: The tool shall support code blocks (#+BEGIN_SRC ... #+END_SRC).
**FR-15**: The tool shall support inline code (~code~ or =code=).
**FR-16**: The tool shall support images ([[url]]).
**FR-17**: The tool shall support links ([[url][description]]).
**FR-18**: The tool shall support lists (+, -, 1.).

#### 3.1.4 Slide Separation
**FR-19**: By default, the tool shall use first-level headings (`#` in Markdown, `*` in Org-mode) as slide separators.
**FR-20**: The tool shall continue to support explicit horizontal rules (`---` in Markdown, `-----` in Org-mode) as additional slide separators.
**FR-21**: The tool shall allow users to specify a custom separator via the `--separator` flag, which will take precedence over the default heading-based separation.

#### 3.1.5 HTML Generation
**FR-22**: The tool shall wrap each slide in a `<section class="slide">` element.
**FR-23**: The tool shall bundle Tailwind CSS (v4) and daisyUI (v5) styles into the generated HTML.
**FR-24**: The tool shall generate a responsive layout that fits standard screen resolutions.
**FR-25**: The tool shall use Alpine.js (v3) for client-side interactivity, including the following keyboard shortcuts:
  - `n`, `Right Arrow`, `Space`: Next slide.
  - `p`, `Left Arrow`: Previous slide.
  - `Shift+Alt+,`: First slide.
  - `Shift+Alt+.`: Last slide.
  - `t`: Toggle Theme Command Palette.
  - `f`: Toggle browser fullscreen.
  - `N`: Toggle line numbers in code blocks.
  - `?`: Toggle help screen.
  - `/`: Toggle slide search palette.
  - `Shift+P`: Toggle Pause Mode.
**FR-26**: The tool shall use `github.com/alecthomas/chroma/v2` for high-quality syntax highlighting in code blocks.
**FR-27**: Code blocks shall display line numbers by default.
**FR-28**: The syntax highlighting should be compatible with the selected daisyUI themes.

#### 3.1.6 CLI Interface
**FR-29**: The tool shall accept the input file as a positional argument.
**FR-30**: The tool shall support the `-o` or `--output` flag to specify an explicit output path.
**FR-31**: If the `-o` or `--output` flag is provided, the tool shall write the result to the specified path, overriding the default naming convention.
**FR-32**: The tool shall accept the output flag regardless of its position relative to the input file (e.g., `slidesdeck -o out.html in.md` or `slidesdeck in.md -o out.html`).
**FR-33**: The tool shall provide a `--help` command with usage instructions.
**FR-34**: The tool shall support selecting the default daisyUI theme via a `-t` or `--theme` flag.
**FR-35**: The generated HTML shall include a "Command Palette" interface for switching themes at runtime.
**FR-36**: The Command Palette shall list all available daisyUI themes, prefixed with `light:` or `dark:` (e.g., `light: cupcake`, `dark: dracula`).
**FR-37**: The Command Palette shall use latest **`flexsearch` (v0.8.2)** for high-performance theme searching and filtering.
**FR-38**: The theme palette shall always open with the current theme selected/highlighted.
**FR-39**: If search text is typed and then cleared, the selection shall revert to the currently active theme.
**FR-40**: Pressing the `Escape` key shall close the palette without changing the current theme.
**FR-41**: The tool shall apply the selected daisyUI theme by setting the `data-theme` attribute on the root HTML element.

#### 3.1.7 Presentation Features
**FR-42**: The tool shall include a "Pause Mode" feature, triggered by the `Shift+P` key in the browser.
**FR-43**: In "Pause Mode", the tool shall provide a configuration screen to set either a countdown timer (e.g., 5 minutes) or a "target time" (e.g., 2:00 PM).
**FR-44**: The configuration screen shall allow the presenter to enter a custom message (e.g., "Lunch Break", "Technical Discussion").
**FR-45**: Upon starting the timer, the tool shall display a full-screen, visually impressive view with the live countdown and the associated message in large, clear text.
**FR-46**: The tool shall store the active countdown state (target end time and message) in browser Local Storage.
**FR-47**: The active countdown shall persist across browser restarts, page reloads, and toggling Pause Mode on/off.
**FR-48**: If Pause Mode is toggled off while a countdown is active, and then toggled back on, the countdown shall display the remaining time without being reset.
**FR-49**: The countdown screen shall include a "Reset" button that stops the active countdown and removes it from Local Storage.
**FR-50**: The tool shall include a help screen, toggled by the `?` key, displaying the tool name, a short description, and an overview of all keyboard shortcuts.
**FR-51**: The tool shall include a "Slide Search" Command Palette, toggled by the `/` key, which shall display all slides as a default listing upon being opened.
**FR-52**: The Slide Search Palette shall use latest **`flexsearch` (v0.8.2)** to provide fast, full-text search capability. Matches shall be returned asynchronously.
**FR-53**: The search shall index both slide titles and slide content. Matching slide titles shall have priority in the results list, appearing above content-only matches. Highlighting of the first result shall be automatic as users type.
**FR-54**: Search results shall include a slide number, title, and a subtitle featuring a preview of the slide content. Selecting a result and pressing `Enter` shall immediately jump to the slide and close the palette.

### 3.2 Non-Functional Requirements
**NFR-1**: Performance: Conversion should take less than 1 second for files up to 100 slides.
**NFR-2**: Portability: The binary should be statically linked for easy distribution.
**NFR-3**: Security: The tool should not include or execute remote scripts during conversion.

### 3.3 Development Requirements
**DR-1**: The project shall use `templ` for all backend HTML templates.
**DR-2**: The project shall include an `air` configuration for hot reloading during development.
**DR-3**: `templ`, `air`, and `esbuild` shall be managed using the native Go tool management (`go get -tool <package>`) to ensure version consistency.
**DR-4**: The frontend asset pipeline shall use `esbuild` to bundle and optimize CSS and JavaScript files into `assets/dist`.
**DR-5**: daisyUI themes shall be managed modularly, with each theme stored in its own CSS file under `assets/src/css/themes/`.
**DR-6**: All optimized assets in `assets/dist/` shall be embedded into the Go binary using `go:embed` at build time.
**DR-7**: CLI integration and functional tests shall be written using `github.com/rogpeppe/go-internal/testscript`.
**DR-8**: All tests, including CLI testscripts, shall be executable via the standard `go test` command.

## 4. System Architecture
The tool will follow a standard compiler-like architecture:
1. **Frontend**: Alpine.js (v3) for interactivity, Tailwind CSS (v4) with daisyUI (v5) for styling. Assets are modular, processed by the Tailwind CLI (CSS) and **`esbuild`** (JS), and embedded into the binary.
2. **Backend**: Go with `templ` for HTML components.
3. **Parsers**: `goldmark` for Markdown, `github.com/niklasfasching/go-org` for Org-mode.
4. **Syntax Highlighting**: `chroma` (alecthomas/chroma) for source code presentation.
5. **Development**: `air` for hot reloading.
6. **Build Pipeline**: Tailwind CLI for CSS, and `esbuild` for JS bundling and minification.

## 5. Appendix
### 5.1 Sample Commands
- `slidesdeck talk.md` -> Creates `talk.html` in the same directory.
- `slidesdeck sub/deck.org` -> Creates `sub/deck.html`.
- `slidesdeck -o presentation.html docs/talk.md` -> Creates `presentation.html` in the current directory.
- `slidesdeck docs/talk.md -o presentation.html` -> Creates `presentation.html` in the current directory.
