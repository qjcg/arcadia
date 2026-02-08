# Slidesdeck

**Plain text, polished slides.**

Slidesdeck is a single-purpose CLI tool that transforms your Markdown or Emacs Org-mode notes into professional, self-contained HTML slideshows. Designed for developers and technical writers who want to stay in their flow, it combines the simplicity of plain text with the polish of modern web technologies.

## Features

- **Multi-format Support**: Convert `.md` (CommonMark) and `.org` (Org-mode) files.
- **Self-contained HTML**: Styles (Tailwind CSS, daisyUI) and interactivity (Alpine.js) are bundled into a single file—no external dependencies or internet connection required.
- **Command Palette**: Switch between dozens of daisyUI themes at runtime using a high-performance search interface.
- **Font Themes**: Choose from 16 professionally curated typography pairings using Google Fonts.
- **Pause Mode**: Manage presentation breaks with a configurable, persistent countdown timer (`Shift+P`).
- **Syntax Highlighting**: Built-in high-quality highlighting with Chroma, line numbers included by default.
- **Keyboard-First**: Navigate slides, toggle fullscreen, and manage themes entirely via the keyboard.
- **Customizable**: Supply [custom theme and font theme CSS files](docs/CUSTOM_THEMES.md).

## Installation

```bash
go install github.com/charmbracelet/arcadia/x/slidesdeck@latest
```

## Quickstart

1. **Create your notes**: Use first-level headings (`#` or `*`) to separate slides.
2. **Convert to HTML**:
   ```bash
   slidesdeck presentation.md
   ```
3. **Open the browser**: `presentation.html` is created in the same directory.

