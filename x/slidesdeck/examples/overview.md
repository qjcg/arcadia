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

# Multi-Format Support
Slidesdeck natively supports both:
1. **Org-mode** (.org)
2. **Markdown** (.md)

Both formats use standard separators:
- First-level headings (e.g., `# Header` or `* Header`)
- Horizontal rules (e.g., `---` or `-----`)
- Custom separators via `--separator` flag

---

# Checklists and Lists
## Technical Task Tracking
- [x] Integrate Goldmark for Markdown
- [x] Add Org-mode support via go-org
- [x] Implement orderless search
- [ ] Add PDF export (coming soon)

## Nested Structures
- Level 1
    - Level 2
        - Level 3

---

# Tables for Technical Data
| Feature         | Markdown | Org-mode | Search |
|-----------------|:--------:|:--------:|:------:|
| Format Support  | ✅       | ✅       | Native |
| Syntax Highlighting | ✅   | ✅       | Chroma |
| Theme Palettes  | ✅       | ✅       | Built-in|
| Command Palette | ✅       | ✅       | FlexSearch|

---

# Data Modeling with CUE
CUE is powerful for configuration and validation.

```cue
#User: {
    id:   int
    name: string
    role: "admin" | "member"
    tags: [...string]
}

myUser: #User & {
    id:   101
    name: "Gopher"
    role: "admin"
    tags: ["core", "dev"]
}
```

---

# System Scripting: Python
Perfect for quick logic and data processing.

```python
import sys

def analyze_complexity(code):
    lines = code.split('\n')
    score = sum(1 for line in lines if "if" in line or "for" in line)
    return f"Complexity Score: {score}"

print(analyze_complexity("if True:\n  for i in range(10): pass"))
```

---

# Power of Rust
Memory safety and performance without a runtime.

```rust
fn main() {
    let languages = vec!["Go", "Rust", "Python", "CUE"];

    for lang in languages.iter() {
        println!("Slidesdeck supports: {} with high-quality highlighting", lang);
    }

    // Pattern matching example
    let status = Some(200);
    match status {
        Some(200..=299) => println!("Success!"),
        _ => println!("Request failed"),
    }
}
```

---

# Frontend Logic: JavaScript
Modern ESM and interactive features.

```javascript
export class SearchService {
    constructor(index) {
        this.index = index;
    }

    async search(query) {
        const tokens = query.toLowerCase().split(/\s+/);
        return this.index.filter(doc =>
            tokens.every(token => doc.text.includes(token))
        );
    }
}
```

---

# Backend Systems: Go
The engine behind Slidesdeck.

```go
package main

import "fmt"

func main() {
    msg := "Built with ❤️ using Charm components"
    for i := 0; i < 3; i++ {
        fmt.Printf("%d: %s\n", i+1, msg)
    }
}
```

---

# Command Palette Search
Press <kbd>/</kbd> at any time to open the **Search Palette**.

- Powered by **FlexSearch**
- **Orderless Logic**: Search "python logic" to find "System Scripting: Python"
- Automatic highlighting
- Instant jump to any slide

---

# Interactive Pause Mode
Manage break times professionally with <kbd>Shift+P</kbd>.

- **Countdown Message**: Customize what's shown during the break
- **Duration**: Choose "Until Time" or "In Minutes"
- Persistent state survives browser reload
- Full-screen high-contrast display

---

# Keyboard Shortcuts
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

# Go forth and Present!
Slidesdeck - Technical presentations for the modern era.

*Start your next talk with a simple command:*
```bash
slidesdeck overview.md
```
