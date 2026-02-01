# jmacs Implementation Guide

## Overview

jmacs is an Emacs-like text editor written entirely in Elbereth and compiled to Go. It demonstrates how functional programming patterns from Lisp can be effectively used to build interactive terminal applications.

## Architecture

### Core Components

#### 1. **Data Structures** (Immutable State)
All editor state is represented as immutable maps (Elbereth hash maps):

```elbereth
; Buffer - represents a file being edited
{:filename "file.txt"
 :content "file content"
 :lines ["line 1" "line 2" "line 3"]
 :cursor-row 0
 :cursor-col 0
 :modified false}

; Editor - main editor state
{:buffers [buffer1 buffer2]
 :active-buffer 0
 :mode :normal
 :message "Status message"
 :width 80
 :height 24}
```

#### 2. **Buffer Operations**
Functions that transform buffer state immutably:

- `make-buffer` - Create new buffer
- `buffer-insert-char` - Insert character at cursor
- `buffer-delete-char` - Delete character at cursor
- `buffer-move-cursor` - Move cursor position

All operations follow functional programming principles:
- **No side effects** - Pure functions only
- **Immutable updates** - Use `merge` to create new state
- **Composability** - Operations can be chained

#### 3. **Editor Commands**
Higher-level commands that modify editor state:

- `cmd-quit` - Exit the editor
- `cmd-help` - Display help
- `cmd-move-forward` - Move right
- `cmd-move-backward` - Move left
- `cmd-move-next-line` - Move down
- `cmd-move-prev-line` - Move up

Commands return new editor state, following the Redux pattern.

#### 4. **BubbleTea Integration**
jmacs uses the `charmbracelet/bubbletea` framework for TUI rendering. Three key functions:

- `model-init` - Initialize the model (called once at startup)
- `model-update` - Handle messages (keypresses, resize events, etc.)
- `model-view` - Render the current state to terminal

### The BubbleTea Pattern

BubbleTea follows this loop:
```
[Model] -> View -> Render
  ^                   |
  |-- Update <- Messages (keyboard, resize, etc.)
```

In jmacs:
1. **Model** is the `editor` map containing all state
2. **View** function (`model-view`) renders the current model
3. **Update** function (`model-update`) processes events and returns new model

### Key Design Decisions

#### 1. **Functional State Management**
Instead of mutable buffers, jmacs uses immutable data structures. Each operation creates a new buffer state:

```elbereth
; Before: :cursor-col = 5
; After:  :cursor-col = 6
(buffer-move-cursor buf 1 0)  ; returns new buffer
```

This makes:
- **Undo/Redo** trivial (just keep state history)
- **Debugging** easier (can inspect any past state)
- **Concurrency** safe (no races)

#### 2. **Map-Based Configuration**
Use maps instead of objects for configuration and state:

```elbereth
(merge editor {:message "New message"})  ; Clean updates
```

Benefits:
- No object hierarchy
- Easy to inspect
- Natural pattern matching
- Extensible

#### 3. **Go Interoperability**
Direct access to Go packages via Elbereth import:

```elbereth
(import [tea "github.com/charmbracelet/bubbletea"]
        [lipgloss "github.com/charmbracelet/lipgloss"])
```

This allows:
- Using proven Go libraries from Lisp
- No FFI overhead
- Type safety through Go's system

## Compilation Pipeline

```
Elbereth Source (.elb)
        |
        v
Lexer (tokenization)
        |
        v
Parser (S-expression to AST)
        |
        v
Evaluator (interpret or compile)
        |
        v
Code Generator (to Go .go file)
        |
        v
Go Compiler (go build)
        |
        v
Native Binary
```

## Future Enhancements

### Phase 1: Core Editing (Current)
- [x] Basic buffer management
- [x] Cursor movement
- [ ] Text insertion and deletion

### Phase 2: Navigation
- [ ] Syntax highlighting
- [ ] Line numbering
- [ ] Status bar improvements

### Phase 3: Advanced Features
- [ ] Undo/redo with command history
- [ ] Search and replace
- [ ] Multiple window layouts
- [ ] Plugin system

### Phase 4: Extensibility
- [ ] Elbereth-based config file
- [ ] User-defined keybindings
- [ ] Custom commands

## Testing

Build and test the editor:

```bash
# Build
task build

# Run
./jmacs

# Generate code to inspect
task gen
```

## Performance Considerations

Since jmacs compiles to Go:
- **Fast startup** - No interpreter boot time
- **Low memory** - No runtime overhead
- **Efficient editing** - Go's optimized standard library
- **Responsive UI** - Compiled code with minimal GC pressure

## Bridging Lisp and Go

jmacs demonstrates several patterns:

### 1. **Functional Go**
Elbereth's functional paradigm (immutable state, pure functions) works well with Go's concurrency model.

### 2. **Type System**
Despite Lisp's dynamism, jmacs can leverage Go's type system:
```elbereth
; Go types available through Elbereth
(tea/NewProgram model)  ; Returns tea.Program
(.Run program)          ; Call method on Go object
```

### 3. **Error Handling**
Elbereth maps to Go's error convention:
```elbereth
(if-let [err (.Run program)]
  (println (str "Error: " err)))
```

## References

- **Elbereth**: `../elbereth/SPEC.md`
- **BubbleTea**: https://github.com/charmbracelet/bubbletea
- **Lipgloss**: https://github.com/charmbracelet/lipgloss
- **Emacs**: https://www.gnu.org/software/emacs/

## Contributing

To extend jmacs:

1. Add new commands in the "Editor Commands" section
2. Handle them in `model-update`
3. Display results in `model-view`
4. Test with `task run`

Remember: Keep functions pure and state immutable!
