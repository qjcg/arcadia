# Software Requirements Specification
## Fractals - Interactive Terminal-Based Fractal Viewer

**Version:** 1.0.0 | **Date:** January 20, 2026 | **Status:** Active Development

---

## 1. Overview

Fractals is a Go-based terminal application for interactive exploration of mathematical fractal sets. Users navigate the complex plane with keyboard controls, discover interesting regions via auto-pilot, save discoveries as bookmarks/screenshots, and share state via URLs.

### Key Features
- Interactive navigation (pan, zoom, type switching)
- 11 fractal algorithms including Mandelbrot, Julia, Burning Ship, Tricorn, Multibrot variants
- Auto-pilot mode with intelligent interest detection
- 8 color schemes with dynamic hue rotation
- Bookmarks, screenshots, shareable `fractal://` URLs
- Animated transitions between fractal types
- Works on Linux, macOS, Windows (via WSL)

### Constraints
- ASCII-only rendering, 256-color terminal
- Float64 precision (practical zoom limit: ~1e15x)
- 20 FPS minimum (50ms per frame)
- Terminal minimum: 40×10 characters

---

## 2. User Personas & Use Cases

### Personas

| Role | Goals | Needs |
|------|-------|-------|
| **Researcher** (Dr. Elena Chen) | Study fractal properties, validate hypotheses | Extreme zoom, precise parameter control, URL sharing, bookmark reproducibility |
| **Artist** (Maya Rodriguez) | Generate artwork, discover beautiful regions | Auto-pilot, color schemes, smooth animations, screenshot capture |
| **Developer** (Alex Kim) | Understand algorithms, tinker with code | Clean modular codebase, extensible interfaces, documentation |
| **Educator** (James Wu) | Teach fractal concepts, classroom demos | Sensible defaults, minimal setup, help text, smooth auto-pilot for projection |
| **Performance Enthusiast** (Sarah Thompson) | Test limits, optimize rendering | Frame timing metrics, large terminal support, memory efficiency |

**Use Cases:**
- **Casual Exploration**: Launch → auto-pilot → cycle colors → save screenshot
- **Targeted Navigation**: Launch with coordinates → manual pan/zoom → bookmark
- **Julia Set Exploration**: Switch to Julia → adjust parameters → bookmark discoveries
- **Classroom Demo**: Launch → explain → show auto-pilot/colors → interactive student exploration
- **Research Sharing**: Discover region → copy URL → colleague opens identical state

---

## 3. Functional Requirements

### 3.1 Fractal Rendering
- **Mandelbrot Set**: Classic z(n+1) = z(n)² + c, centered at -0.5+0i
- **Julia Set**: Parameter-adjustable c, centered at 0+0i, default c = -0.7 + 0.27015i
- **Additional Types**: Burning Ship, Tricorn, Multibrot-3/4/5, Celtic, Perpendicular, Manhattan, Newton
- **Iteration Calculation**: Escape-time algorithm (|z|² > 4 diverges), adjustable 10-500 iterations (default: 50)
- **ASCII Mapping**: Characters " .:-=+*#%@" (10 density levels) scaled to iteration count
- **Color Schemes**: 8 schemes (grayscale, blue, rainbow, fire, purple, green, gold, cyan)
  - Each maps iterations to ANSI 256-color codes
  - Dynamic hue rotation (0.5°/tick) available via toggle
- **Terminal Adaptation**: Auto-detect size; CLI flags -w/-h override; minimum 40×10

### 3.2 Navigation & Manipulation
- **Manual Pan**: Arrow/WASD keys move 10% of viewport per keystroke
- **Manual Zoom**: i/o zoom 1.2x ±; +/- explicit zoom or set direction
- **Type Switching**: 1-9 (types 1-9), m (Manhattan), n (Newton); reset zoom/center to type defaults
- **Iteration Depth**: [ ] decrease/increase by 10 (min 10, max 500 manual / 2000 auto)
- **Reset**: 0 key restores defaults (zoom 1.0x, type-default center, iter 50)
- **Julia Parameters**: J/j ±0.05 real, K/k ±0.05 imaginary (no bounds)

### 3.3 Auto-Pilot Mode
- **Toggle**: z key enables continuous zoom (direction set by r or +/-)
- **Interest Detection**: Valleys in fractal structure
  - Sample ~25 points around candidate center
  - Score: range (×100) + variance (×20) + boundary bonus (500 if 10-90% at maxIter)
  - Uniform penalty: zero score if all iterations match
- **Point Search**: Multi-pass spiral (local 50%, medium 150%, wide 400% of view)
- **Smooth Panning**: Move 5% of remaining distance toward target per tick
- **Adaptive Iterations**: Scale up ~20 per zoom decade
- **Speed Control**: { } adjust multiplier (0.90-1.50x, default 1.05x)
- **Transitions**: At zoom limit, trigger fractal switch if enabled; otherwise pause zoom

### 3.4 Transitions (Animated Fractal Switching)
- **Modes** (cycle via T key):
  - None: instant switch
  - Fade: 1-2 sec linear crossfade
  - Zoom Out-In: zoom to 0.1x then 1.0x in new fractal
  - Rotate: 90-frame center rotation
  - Breakthrough: particle effect with gravity (100 frames)
- **Trigger**: When auto-pilot reaches zoom limit (0.1x or 1e15x)
- **Effect**: Switch type → reset zoom to 1.0x → continue auto-pilot

### 3.5 State Management

#### Bookmarks
- b: Save current state with auto-generated name (`{adjective}_{noun}`)
- l: Load bookmark from list (navigate with arrows/j/k/1-9, delete with d/x)
- Storage: `~/.config/fractals/bookmarks.yaml` (auto-created)
- Restore: Type, center (X,Y), zoom, iterations, colors, Julia params, modes

#### Screenshots
- p: Capture as text file with metadata header
- Filename: `{type}_{YYYYMMDD_HHMMSS}[_N].txt` to `~/.config/fractals/screenshots/`
- Header: Type, coords (10 decimal places), zoom, iterations, color, resolution

#### Shareable URLs
- U: Generate and copy `fractal://{TYPE}/{X}/{Y}/{ZOOM}/{ITER}/?[params]`
  - Query params: `color_theme`, `autopilot`, `dynamic_color`, `transition`, `julia_cr`, `julia_ci`
- Launch via: `fractals --url 'fractal://...'` or as positional argument
- R: Random state (random type, interesting coords, random colors)

### 3.6 Help & Input

#### Help
- ? toggles full-screen modal with all keyboard bindings

#### Bookmark Naming
- Prompt accepts alphanumeric/underscore/hyphen; backspace deletes; Esc cancels

#### Bookmark List
- Navigate with arrows/j/k/1-9, Enter loads, d/x deletes, Esc exits

### 3.7 CLI & Static Mode

#### Arguments
- `-t/--type`: Fractal type (default: mandelbrot)
- `-c/--color`: Color scheme
- `-w/--width`, `-h/--height`: Terminal size (0 = auto)
- `-i/--iterations`: Max iterations
- `-x/-y`: Center coordinates
- `-z/--zoom`: Zoom level
- `-jr/-ji`: Julia parameters
- `-r/--random`: Random state
- `--url`: Load from fractal:// URL
- `--static`: Force single-shot rendering (automatic if params given; no interaction)
- `--interactive`: Force interactive mode

---

## 4. Non-Functional Requirements

| Requirement | Specification |
|---|---|
| **Rendering Speed** | ≥20 FPS (≤50ms/frame) at 80×24; maintain up to 200×50 |
| **Startup Time** | <100ms launch, <50ms bookmark load, <200ms first frame |
| **Memory** | <50MB base; no growth during extended sessions; no leaks |
| **Precision** | Float64 coords; graceful degradation at zoom >1e14 |
| **Responsiveness** | User actions respond within 1 tick (50ms) |
| **Error Handling** | Graceful on invalid URLs, bookmark failures, terminal resize; no panics |
| **Accessibility** | Mandatory grayscale scheme; distinct schemes for color-blind |
| **Portability** | Linux, macOS, Windows (WSL); ANSI 256-color terminals |
| **Code Quality** | Modular (render/nav/state/UI); unit tests on algorithms; documented |
| **File Safety** | Restrict paths to `~/.config/fractals/`; validate filenames; atomic writes |

---

## 5. External Interfaces

### Terminal UI
- Fractals viewport (main, height -3 lines)
- Status bar (1 line): `{FRACTAL} | Center: (X, Y) | Zoom: Zx | Iter: I | Color: S | Auto: ↑/↓`
- Help overlay, bookmark list, input prompts (modal)

### Files
- Bookmarks: `~/.config/fractals/bookmarks.yaml` (YAML; auto-created)
- Screenshots: `~/.config/fractals/screenshots/*.txt`

### APIs (interfaces for extensibility)
- Fractal calculator: Calculate(cr, ci, maxIter) → iterationCount
- Color scheme: GetColor(iter, maxIter, scheme) → ANSICode
- Transition: Start/Update/GetMessage

---

## 6. Data Structures

```go
// Core state
Config struct {
    Width, Height int
    MaxIter int
    CenterX, CenterY float64
    Zoom float64
    ColorScheme, FractalType string
    JuliaCr, JuliaCi float64
}

// Bookmark persistence
Bookmark struct {
    Name string
    URL string  // fractal:// encoding all state
    FractalType string
    CenterX, CenterY, Zoom float64
    MaxIter int
    ColorScheme string
    JuliaCr, JuliaCi float64
    AutopilotEnabled, DynamicColorEnabled bool
    TransitionMode string
}

// Model (UI state)
model struct {
    config Config
    showHelp, autoZoom, hasTarget bool
    targetX, targetY float64
    transitionMode int
    transitionProgress float64
    dynamicColor bool
    hueShift float64
    bookmarks []Bookmark
    bookmarkCursor int
    zoomSpeed float64
    // ... status messages, timers
}
```

---

## 7. Constraints & Assumptions

### Technical
- Float64: ~15-17 significant digits; practical zoom limit ~1e15x
- ANSI 256-color palette required
- VT100 escape sequences
- Go 1.18+ (generics)
- XDG config dir writable

### Design
- Single-threaded event loop (Bubble Tea)
- Keyboard-only interaction (no mouse)
- Modal dialogs acceptable
- File ops synchronous but fast

### Known Limitations
- ASCII rendering less detailed than pixel graphics
- No anti-aliasing
- Floating-point precision visible above 1e13-1e14 zoom
- No GPU acceleration
- No fractal customization (hardcoded algorithms)

---

## Appendix A: Keyboard Controls

| Key(s) | Action |
|--------|--------|
| Arrow/WASD | Pan |
| i/o | Zoom in/out |
| +/= or -/_ | Zoom in/out or set direction |
| z | Toggle auto-pilot |
| r | Reverse auto-pilot direction |
| {/} | Decrease/increase speed |
| 0 | Reset defaults |
| 1-9, m, n | Switch fractal type |
| c | Cycle color scheme |
| C | Toggle dynamic hue rotation |
| [/] | Decrease/increase iterations |
| J/j, K/k | Adjust Julia parameters |
| T | Cycle transition modes |
| b | Save bookmark |
| l | Load bookmark |
| p | Save screenshot |
| U | Copy URL to clipboard |
| R | Random fractal |
| ? | Help |
| q, Esc, Ctrl-C | Quit |

---

## Appendix B: Algorithms

### Escape-Time (all fractals)
```
for pixel (x,y):
  z = 0
  iter = 0
  while |z|² ≤ 4 and iter < maxIter:
    z = fractal_formula(z, c)
    iter++
  return iter
```

### Coordinate Mapping
```
width_ratio = 3.5 / zoom
height_ratio = 2.5 / zoom
cr = centerX - width_ratio/2 + (col/width) × width_ratio
ci = centerY + height_ratio/2 - (row/height) × height_ratio
```

### Interest Score (find interesting points)
```
score = 0
score += iterRange × 100        // Edge detection
score += stdDev × 20            // Texture
if avgIter in [10-90% of maxIter]: score += 500
if avgIter in [5-95% of maxIter]: score += 200
if iterRange == 0: score = 0    // Uniform penalty
```

---

**End of SRS**
