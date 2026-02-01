# jmacs

An Emacs-like text editor written in Elbereth (a Lisp dialect that compiles to Go).

## About

jmacs brings the power and elegance of Emacs to Elbereth. Written in Lisp but compiled to efficient Go, it provides:

- **Modal Editing**: Keyboard-driven interface with Emacs-style keybindings
- **Multiple Buffers**: Edit multiple files seamlessly
- **Extensibility**: Write extensions in Elbereth
- **Performance**: Compiled to native Go, no runtime overhead
- **TUI Interface**: Full-featured terminal user interface

## Features

- [x] Buffer management
- [x] Basic editor commands
- [x] Keybinding system
- [ ] Syntax highlighting
- [ ] Undo/redo
- [ ] Search and replace
- [ ] File operations
- [ ] Extensibility through Elbereth scripts

## Quick Start

```bash
# Build jmacs
task build

# Run jmacs
./jmacs
```

## Keybindings

- `C-x C-c` - Exit
- `C-?` - Help
- Navigation with arrow keys
- Standard Emacs movement keys (C-n, C-p, C-f, C-b)

## Architecture

jmacs is written entirely in Elbereth and demonstrates:

1. **TUI Framework Integration** - Using bubbletea for terminal UI
2. **State Management** - Buffer and editor state in Lisp
3. **Event Handling** - Keyboard and mouse events
4. **Go Interoperability** - Seamless integration with Go packages

## Building an Editor in Elbereth

This project showcases:

- Using the `charmbracelet/bubbletea` Go package from Elbereth
- Functional programming patterns for editor state
- Immutable data structures for buffers
- Pattern matching for keyboard events

See `IMPLEMENTATION.md` for technical details on how the editor is built.
