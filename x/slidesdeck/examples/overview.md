# Welcome to Slidesdeck
*Plain text, polished slides.*

Slidesdeck is a single-purpose CLI tool that transforms your notes into professional, self-contained HTML slideshows.

- Type <kbd>n</kbd> or <kbd>p</kbd> to navigate
- Type <kbd>?</kbd> to see all help
- Type <kbd>/</kbd> to search slides
- Type <kbd>t</kbd> to change color themes
- Type <kbd>T</kbd> to change font themes

---

# Core Philosophies
## Simplicity
- No complex configuration
- One command to convert
## Portability
- Single HTML file output
- Zero external CSS/JS dependencies (all embedded)
## Developer Experience
- Native Go performance
- Clear errors and fast feedback loop

---

# Standard Formatting
Markdown supports various formatting options:
- **Bold** text
- *Italic* text
- ~~Strikethrough~~ text
- `Code` and inline snippets

---

# Multi-Format Support
Slidesdeck natively supports both:
1. **Org-mode** (.org)
2. **Markdown** (.md)

Both formats use standard separators:
- First-level headings (e.g., `* Header`)
- Horizontal rules (e.g., `-----` or `---`)
- Custom separators via `--separator` flag

---

# Lists and Deep Nesting
## Unordered Lists
- Simple item
- Nested item
    - Even deeper
## Ordered Lists
1. First item
2. Second item
    1. Sub-item
## Checklists
- [x] Feature A
- [x] Feature B
- [ ] Feature C

---

# Tables for Technical Data
| Feature         | Org-mode | Markdown |
|-----------------|----------|----------|
| Tables          | Supported| Supported|
| Lists           | Supported| Supported|
| Code Blocks     | Supported| Supported|
| Search          | Included | Included |
| Themes          | 30+      | 30+      |

---

# Syntax Highlighting
Slidesdeck uses Chroma for high-quality syntax highlighting.

```go
package main

import "fmt"

func main() {
    // High-performance Go code
    fmt.Println("Hello from Slidesdeck!")
}
```

- Toggle line numbers with <kbd>N</kbd>
- Built-in "Copy" button for code blocks

---

# Command Palette Search
Press <kbd>/</kbd> at any time to open the **Search Palette**.

- Powered by **FlexSearch** (v0.8.2)
- High-performance full-text search
- Priority given to slide titles
- Automatic highlighting as you type
- Instant jump to any slide

---

# Color Theme Palette
Press <kbd>t</kbd> to open the **Theme Switcher**.

- 30+ built-in daisyUI themes
- Search by name or category:
    - 'dark: ' for dark themes
    - 'light: ' for light themes
- Current theme is always pre-selected
- Apply custom CSS via `--theme-file`

---

# Font Theme Palette
Press <kbd>Shift+T</kbd> (<kbd>T</kbd>) to open the **Font Switcher**.

- 16 professionally curated font themes
- Pairings for headings, body, and code
- Example themes:
    - *Elegant*: Playfair Display / Source Sans
    - *Tech*: Space Grotesk / JetBrains Mono
    - *Brutalist*: Archivo Black / Space Grotesk
- Apply custom fonts via `--fonttheme-file`

---

# Interactive Pause Mode
Manage break times professionally with <kbd>Shift+P</kbd>.

- Configurable countdown timer
- Custom break messages
- Persistent state (survives browser close/reload)
- Full-screen, high-contrast display
- "Reset" button to clear active timers

---

# Keyboard Shortcuts Cheat Sheet
| Key                                            | Action                  |
|------------------------------------------------|-------------------------|
| <kbd>n</kbd> / <kbd>→</kbd> / <kbd>Spc</kbd> | Next Slide              |
| <kbd>p</kbd> / <kbd>←</kbd> | Previous Slide          |
| <kbd>/</kbd>                                   | Search Palette          |
| <kbd>t</kbd> / <kbd>T</kbd>                    | Theme / Font Palettes   |
| <kbd>f</kbd>                                   | Toggle Fullscreen       |
| <kbd>N</kbd>                                   | Toggle Line Numbers     |
| <kbd>Shift+P</kbd>                             | Pause Mode              |
| <kbd>?</kbd>                                   | Help Screen             |

---

# Portable and Fast
- Generated as a single HTML file
- Zero external dependencies
- Powered by Go, Tailwind CSS v4, and Alpine.js
- Blazing fast conversion (<200ms)

---

# Go forth and Present!
Slidesdeck - Technical presentations for the modern era.

*Start your next talk with a simple command:*
```bash
slidesdeck overview.org
```
