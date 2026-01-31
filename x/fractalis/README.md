# fractals

An interactive terminal-based fractal viewer with ASCII art rendering and color scheme support.

## Features

### Interactive Mode (Default)
- Real-time fractal exploration with keyboard navigation
- Zoom and pan controls for deep exploration
- Auto-pilot mode with intelligent exploration of interesting regions
- Random exploration for discovering new fractals (generates completely random interesting views)
- Switch between 8 different fractal types on the fly
- Eight color schemes: grayscale, blue, rainbow, fire, purple, green, gold, and cyan
- Adjustable iteration depth for detail control
- Interactive Julia set parameter adjustment
- Bookmark system to save and load favorite locations
- Screenshot feature to capture views as text files with metadata
- Built-in help system
- Full-screen TUI powered by Bubble Tea

### 3D Mode (Separate Command)
- GPU-accelerated 3D fractal rendering available via `cmd/3d`
- Real-time ray marching of Mandelbulb fractal at 60 FPS
- Full 3D navigation with camera controls
- Soft shadows and ambient occlusion for realistic lighting
- Dynamic color animation with orbit trap coloring
- HDR tone mapping and gamma correction
- Graphical window with interactive controls

## Installation

### Using go install

```bash
go install github.com/qjcg/arcadia/x/fractalis@latest
```

### Building from source

```bash
cd x/fractalis
go tool task build
# Binary will be in bin/fractals
```

### Installing locally

```bash
cd x/fractalis
go tool task install
```

## Usage

### Interactive Mode (Default)

Launch the interactive fractal viewer:

```bash
fractals
```

Once in the interactive viewer, you can:
- Navigate with arrow keys or WASD
- Zoom in/out with +/- or i/o
- Enable auto-pilot mode with 'z' for automatic exploration
- Control auto-pilot zoom direction with 'r' (toggle), '+' (in), or '-' (out)
- Get random fractal with 'R' or completely random interesting view with 'H'
- Switch fractals with number keys (1-8)
- Cycle color schemes with 'c'
- Adjust iteration depth with [ and ]
- For Julia sets, adjust parameters with J/j (real) and K/k (imaginary)
- Save interesting locations with 'b' (bookmarks)
- Load saved locations with 'l'
- Copy current view as shareable URL with 'U'
- Capture screenshots with 'p' (saves text file with fractal + metadata)
- Press '?' for help
- Press 'q' or ESC to quit

### Shareable URLs

Launch fractals directly from shareable `fractal://` URLs that encode the complete interactive state:

```bash
# Basic Mandelbrot
fractals fractal://mandelbrot/-0.5/0.0/1.0/50/

# Julia set with parameters
fractals fractal://julia/0.0/0.0/1.0/50/?julia_cr=-0.7&julia_ci=0.27015&color_theme=blue

# With autopilot and dynamic color
fractals fractal://mandelbrot/-0.7436/0.1314/50.0/100/?color_theme=fire&autopilot=on&dynamic_color=on&transition=fade

# Random mode
fractals fractal://random/?color_theme=rainbow&autopilot=on

# Using --url flag
fractals --url "fractal://mandelbrot/-0.5/0.0/1.0/50/?autopilot=on"
```

**URL Format**: `fractal://$FRACTAL_TYPE/$CENTER_X/$CENTER_Y/$ZOOM/$ITERATIONS/?[options]`

**Query Parameters** (all optional):
- `color_theme`: grayscale, blue, rainbow, fire, purple, green, gold, cyan
- `autopilot`: on/off
- `dynamic_color`: on/off (smooth hue rotation)
- `transition`: none, fade, zoomout, rotate, breakthrough
- `julia_cr`: Julia set real parameter (for Julia sets)
- `julia_ci`: Julia set imaginary parameter (for Julia sets)

**Random URLs**: Use `fractal://random/` to launch with a random fractal at an interesting location

**Copying URLs**: Press **U** in interactive mode to copy the current view as a shareable URL (displayed in status bar)

**Bookmarks Integration**: When you save a bookmark, it automatically generates and stores a shareable URL, making bookmarks portable and shareable with others.

### 3D Mode

GPU-accelerated 3D Mandelbulb rendering is available as a separate command:

```bash
# Run the 3D viewer directly
go run ./cmd/3d

# Or build and install it
go build -o fractalis-3d ./cmd/3d
./fractalis-3d
```

**3D Mode Controls:**
- **WASD** - Move in X/Z plane
- **Space/Shift** - Move up/down in Y
- **Mouse** - Look around (click window to capture mouse, ESC to release)
- **Arrow keys** - Alternative camera look control
- **Q** - Quit

**3D Mode Features:**
- Real-time ray marching at 60 FPS
- Soft shadows and ambient occlusion
- Dynamic color shifting animation
- HDR tone mapping and gamma correction
- Orbit trap coloring for fractal detail

**Development:**
```bash
# Run with hot reload during development
task dev-3d

# Serve as WebAssembly in browser (runs on http://localhost:8080)
task wasm-3d
```

**WebAssembly Mode:**
The 3D mode can be built as WebAssembly to run in any modern browser:

```bash
# Manual build
GOOS=js GOARCH=wasm go build -o fractalis.wasm ./cmd/3d

# Serve with wasmserve
go run github.com/hajimehoshi/wasmserve@latest ./cmd/3d
```

The WebAssembly version provides the same GPU-accelerated ray marching experience as the native version, accessible directly in your browser at http://localhost:8080.

### Advanced Usage

#### Fractal Types

Choose from multiple fractal types using the `-t` or `--type` flag:

```bash
# Mandelbrot set (default)
fractals -t mandelbrot

# Julia set with default parameters (-0.7 + 0.27015i)
fractals -t julia

# Julia set with custom parameters
fractals -t julia -jr -0.4 -ji 0.6

# Burning Ship fractal
fractals -t burningship -x -0.5 -y -0.6

# Tricorn (Mandelbar) fractal
fractals -t tricorn

# Multibrot set with power 3
fractals -t multibrot3

# Multibrot set with power 4
fractals -t multibrot4

# Celtic Mandelbrot fractal
fractals -t celtic

# Perpendicular Mandelbrot fractal
fractals -t perpendicular
```

#### Color Schemes

Choose from eight color schemes:

```bash
# Grayscale (default) - ASCII characters only, no color codes
fractals -c grayscale

# Blue gradient - dark to bright blue
fractals -c blue

# Rainbow gradient - full spectrum
fractals -c rainbow

# Fire gradient - black to red to orange to yellow to white
fractals -c fire

# Purple/magenta gradient
fractals -c purple

# Green gradient - dark to bright green
fractals -c green

# Gold/amber gradient - brown to gold to yellow
fractals -c gold

# Cyan/aqua gradient - dark to bright cyan
fractals -c cyan
```

All color schemes support dynamic color mode (Shift+C) for smooth hue rotation.

#### Zoom and Pan

Explore different regions of the Mandelbrot set:

```bash
# Zoom in 2x at default center
fractals -z 2.0

# Pan to a different location
fractals -x -0.7 -y 0.0

# Combine zoom and pan with rainbow colors
fractals -x -0.7 -y 0.0 -z 2.0 -c rainbow
```

#### Custom Resolution

Override auto-detected terminal size:

```bash
# Set specific width and height
fractals -w 120 -h 50

# Use more iterations for smoother detail
fractals -i 100 -c rainbow
```

## Interactive Controls

When running in interactive mode, use these keyboard shortcuts:

### Navigation
- **Arrow keys** or **WASD** - Pan the view around the complex plane
- **i** / **o** - Zoom in / out (manual control)
- **+**, **=** - Zoom in manually, or set auto-pilot to zoom in direction (↑)
- **-**, **_** - Zoom out manually, or set auto-pilot to zoom out direction (↓)
- **z** - Toggle auto-pilot mode (automatic exploration with intelligent panning)
- **r** - Reverse/toggle auto-pilot zoom direction (↑ ↔ ↓)
- **0** - Reset view to default (position, zoom, and iteration depth)

### Fractal Types
- **1** - Mandelbrot set
- **2** - Julia set
- **3** - Burning Ship
- **4** - Tricorn (Mandelbar)
- **5** - Multibrot-3 (power 3)
- **6** - Multibrot-4 (power 4)
- **7** - Celtic Mandelbrot
- **8** - Perpendicular Mandelbrot
- **9** - Multibrot-5 (power 5)
- **m** - Manhattan distance variant
- **n** - Newton fractal

### Settings
- **c** - Cycle through color schemes (grayscale → blue → rainbow → fire → purple → green → gold → cyan)
- **C** (Shift+c) - Toggle dynamic color mode (smooth hue rotation - mesmerizing with autopilot!)
- **[** and **]** - Decrease/increase iteration depth
- **J** and **j** - Increase/decrease Julia set real parameter
- **K** and **k** - Increase/decrease Julia set imaginary parameter

### Bookmarks & Screenshots
- **b** - Save current location as a bookmark (prompts for name)
- **l** - Load bookmark (shows interactive list)
- In bookmark list:
  - **↑/↓** or **j/k** - Navigate through bookmarks
  - **Enter** - Load selected bookmark
  - **d** or **x** - Delete selected bookmark
  - **1-9** - Quick load bookmark by number
  - **Esc** - Cancel and return to fractal view
- **p** - Save screenshot (captures current view to text file with metadata)

### Random Exploration
- **R** (Shift+r) - Random (generates completely random fractal with interesting view)
  - Randomly selects fractal type, zoom level, position, color scheme, and iterations
  - Uses intelligent algorithms to find visually interesting, non-uniform regions
  - Draws from curated list of known interesting coordinates as seed points
  - Attempts multiple random configurations to ensure engaging results
- **T** (Shift+t) - Cycle through transition modes (None/Fade/Zoom Out-In/Rotate)
  - When auto-pilot hits zoom limits, automatically transition to a new fractal type with animation
  - **Fade**: Smoothly crossfade between fractal types
  - **Zoom Out-In**: Zoom out then into the new fractal for dramatic effect
  - **Rotate**: Spin the view while transitioning to the new fractal

### Other
- **?** - Toggle help screen
- **q** or **Esc** - Quit

## CLI Options

| Flag | Long Form | Description | Default |
|------|-----------|-------------|---------|
| `-r` | `--random` | Start with a completely random interesting view | false |
| `-t` | `--type` | Fractal type: mandelbrot, julia, burningship, tricorn, multibrot3, multibrot4, multibrot5, celtic, perpendicular, manhattan, newton | mandelbrot |
| `-c` | `--color` | Color scheme: grayscale, blue, rainbow, fire, purple, green, gold, cyan | grayscale |
| `-w` | `--width` | Terminal width (0 = auto-detect) | 0 |
| `-h` | `--height` | Terminal height (0 = auto-detect) | 0 |
| `-i` | `--iterations` | Maximum iterations for convergence test | 50 |
| `-x` | | Center X coordinate (real axis) | -0.5 |
| `-y` | | Center Y coordinate (imaginary axis) | 0.0 |
| `-z` | `--zoom` | Zoom level (higher = closer) | 1.0 |
| `-jr` | | Julia set real parameter | -0.7 |
| `-ji` | | Julia set imaginary parameter | 0.27015 |

## Examples

### Interactive Mode

```bash
# Launch interactive viewer (default)
fractals

# Launch with a specific starting fractal and color
fractals -t julia -c rainbow

# Launch zoomed into an interesting region
fractals -t mandelbrot -x -0.75 -y 0.1 -z 3.0 -c rainbow

# Start with a completely random interesting view
fractals --random
# or
fractals -r
```

## Fractal Algorithms

### Mandelbrot Set

The Mandelbrot set is a mathematical set of complex numbers with fascinating fractal properties. For each point `c` in the complex plane, we iterate the formula:

```
z(n+1) = z(n)² + c
```

Starting with `z(0) = 0`, we check if the sequence remains bounded. If `|z|² > 4` after some iterations, the point diverges and is not in the set. Points that don't diverge within the maximum iteration count are considered part of the set.

### Julia Set

The Julia set is similar to the Mandelbrot set, but instead of varying `c` across the complex plane, we fix `c` to a constant and vary the starting point `z(0)`:

```
z(n+1) = z(n)² + c (where c is constant)
```

Different values of `c` produce dramatically different Julia sets, from connected islands to dust-like patterns.

### Burning Ship

Uses absolute values before squaring, creating a fractal that resembles a burning ship:

```
z(n+1) = (|Re(z(n))| + i|Im(z(n))|)² + c
```

### Tricorn (Mandelbar)

Uses the complex conjugate before squaring:

```
z(n+1) = conj(z(n))² + c
```

### Multibrot Sets

Generalizations of the Mandelbrot set using higher powers:

```
z(n+1) = z(n)^d + c (where d = 3, 4, etc.)
```

### Celtic and Perpendicular

Variations that apply absolute value transformations to components of z² before adding c, creating unique symmetrical patterns.

### Visualization

The visualization maps:
- **Real axis** (X): -2.5 to 1.0 (default view)
- **Imaginary axis** (Y): -1.25 to 1.25 (default view)
- **Iteration count**: Determines the character/color intensity
- **Character density**: Space (quick divergence) to @ (in the set)

Color schemes use ANSI 256-color codes to create gradients based on how quickly points diverge.

## Development

### Running Tests

```bash
go tool task test
```

### Building

```bash
go tool task build
```

### Viewing All Color Schemes

```bash
go tool task demo
```

### Cleaning Build Artifacts

```bash
go tool task clean
```

## Bookmarks

The fractal viewer includes a bookmark system to save and revisit interesting locations.

### Saving Bookmarks

1. Navigate to an interesting location
2. Press **b** to save a bookmark
3. A suggested name is auto-generated (e.g., "ethereal_threshold", "frosted_sanctuary")
4. Options:
   - Press **Enter** immediately to use the suggested name
   - Type a custom name to override the suggestion
   - Press **Esc** to cancel
5. Press **Enter** to save

The auto-generated names use evocative words that capture the mystical nature of fractal exploration:
- **Adjectives**: uncharted, ancient, ethereal, luminous, prismatic, wandering, and more
- **Journey Nouns**: path, realm, threshold, sanctuary, nexus, labyrinth, and more
- **Format**: `adjective_noun` (e.g., `crystal_gateway`, `shadowed_labyrinth`)

Bookmarks are stored in `~/.config/fractalis/bookmarks.yaml` in a simple format:

```yaml
bookmarks:
  - name: ethereal_threshold
    url: "fractal://mandelbrot/-0.7436/0.1314/10.0/100/?color_theme=rainbow"
  - name: crystal_gateway
    url: "fractal://julia/0.0/0.0/5.0/50/?julia_cr=-0.4&julia_ci=0.6&color_theme=blue"
```

Each bookmark stores all state (fractal type, coordinates, zoom, colors, etc.) in the URL string, making bookmarks portable and shareable.

### Loading Bookmarks

1. Press **l** to open the bookmark list
2. Use **↑/↓** or **j/k** to navigate through saved bookmarks
3. Press **Enter** to load the selected bookmark
4. Or press **1-9** to quickly jump to a bookmark by number
5. Press **Esc** to cancel and return to exploring

### Deleting Bookmarks

1. Press **l** to open the bookmark list
2. Navigate to the bookmark you want to delete
3. Press **d** or **x** to delete the selected bookmark
4. The bookmark is immediately removed from the list and saved to disk
5. If you delete the last bookmark, the list closes automatically

### Managing Bookmarks

Bookmarks are stored in a human-readable YAML format at `~/.config/fractals/bookmarks.yaml`. You can:
- Edit the file manually to organize or rename bookmarks
- Delete bookmarks directly from the list view (press **d** or **x**)
- Share bookmarks with others by copying the YAML entries
- Back up your favorite locations by copying the bookmarks file

## Screenshots

The fractal viewer allows you to save "screenshots" of your current view as text files.

### Taking Screenshots

1. Navigate to the view you want to capture
2. Press **p** to save a screenshot
3. A confirmation message appears briefly in the status bar
4. The screenshot is saved to `~/.config/fractals/screenshots/`

### Screenshot Files

Screenshots are saved with descriptive filenames:
- Format: `{fractal-type}_{timestamp}.txt`
- Example: `mandelbrot_2026-01-17_143022.txt`
- If a file exists, a counter is added: `mandelbrot_2026-01-17_143022_1.txt`

Each screenshot file contains:
- **Metadata header** with complete location information:
  - Fractal type
  - Center coordinates (high precision)
  - Zoom level
  - Max iterations
  - Color scheme
  - Julia parameters (if applicable)
  - Resolution (terminal size)
- **The rendered fractal** exactly as shown on screen with colors

### Using Screenshots

Screenshots are useful for:
- **Documentation** - Capture interesting fractal features for reference
- **Sharing** - Share ASCII art fractals with others via text files
- **Comparison** - Save multiple views to compare different settings
- **Recreation** - Use the metadata to manually return to the exact location
- **Analysis** - Study fractal patterns in detail at your own pace
- **Art** - Create collections of beautiful fractal ASCII art

The screenshot directory is created automatically when you save your first screenshot.

## Tips for Interactive Exploration

1. **Use auto-pilot mode** - Press 'z' to enable automatic exploration that intelligently finds and zooms into interesting regions
2. **Control zoom direction** - Press 'r' to toggle between zooming in (↑) and zooming out (↓), or use '+'/'-' to set direction explicitly. The direction indicator is always visible in the status bar (even when auto-pilot is off)
3. **Explore in reverse** - After zooming deep into a fractal with auto-pilot, press 'r' to reverse direction and watch it explore back out with fresh perspective
4. **Random exploration** - Press 'R' to explore completely random interesting views with changing fractals, positions, and colors. Great for discovering new fractals and getting inspiration
6. **Bookmark interesting discoveries** - When you find something beautiful, press 'b' to save it for later
7. **Take screenshots of favorites** - Press 'p' to save the current view as a text file with complete metadata
8. **Start with low zoom and gradually zoom in** - Use the + key repeatedly to dive deeper into interesting features
9. **Adjust iteration depth when zoomed in** - Press ] to increase iterations for more detail at high zoom levels (auto-pilot does this automatically)
10. **Explore different color schemes** - Press 'c' to cycle through 8 color schemes and find the most pleasing visualization
11. **Enable dynamic colors** - Press 'C' (Shift+c) to toggle smooth hue rotation. Incredibly mesmerizing with autopilot enabled! Colors continuously shift through the spectrum while maintaining the gradient structure
12. **Julia set exploration** - Switch to Julia set (press 2) then use J/j and K/k to adjust parameters and discover different patterns
13. **Enable automatic fractal transitions** - Press 'T' to cycle through transition modes. When auto-pilot hits zoom limits, it will automatically switch to a new fractal type with smooth animations instead of stopping
14. **Reset when lost** - Press 0 to return to the default view (position, zoom, and iteration depth) for any fractal type
15. **Check the status bar** - The bottom bar shows your current position, zoom, and settings. The auto-pilot direction indicator (↑/↓) is always visible: bright green when active, subtle gray when inactive

## Dependencies

- Standard library: `flag`, `fmt`, `os`, `strings`, `math`, `path/filepath`
- `github.com/charmbracelet/bubbletea`: TUI framework for interactive mode
- `github.com/charmbracelet/lipgloss`: Terminal styling
- `gopkg.in/yaml.v3`: YAML parsing for bookmark storage

## License

Part of the Arcadia project.
