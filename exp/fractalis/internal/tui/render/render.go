package render

import (
	"fmt"
	"runtime"
	"strings"
	"sync"

	"github.com/qjcg/arcadia/exp/fractalis/internal/core/persistence"
	"github.com/qjcg/arcadia/exp/fractalis/internal/tui/cache"
	"github.com/qjcg/arcadia/exp/fractalis/internal/tui/color"
	"github.com/qjcg/arcadia/exp/fractalis/internal/tui/transition"
)

// Renderer handles ASCII fractal rendering with caching and parallel computation
type Renderer struct {
	config           persistence.Config
	colorMode        bool
	dynamicColor     bool
	hueShift         float64
	breakthroughTr   *transition.Breakthrough
	calculateFractal func(cr, ci float64, cfg persistence.Config) int
	// Performance optimization
	iterationCache *cache.IterationCache
	enableParallel bool
	numWorkers     int
}

// NewRenderer creates a new fractal renderer with performance optimizations
func NewRenderer(
	config persistence.Config,
	calculateFractal func(cr, ci float64, cfg persistence.Config) int,
) *Renderer {
	return &Renderer{
		config:           config,
		calculateFractal: calculateFractal,
		iterationCache:   cache.NewIterationCache(3), // Cache last 3 viewports
		enableParallel:   true,
		numWorkers:       runtime.NumCPU(),
	}
}

// SetDynamicColor enables/disables dynamic color mode
func (r *Renderer) SetDynamicColor(enabled bool, hueShift float64) {
	r.dynamicColor = enabled
	r.hueShift = hueShift
}

// SetColorMode enables/disables colorized output
func (r *Renderer) SetColorMode(enabled bool) {
	r.colorMode = enabled
}

// SetBreakthroughTransition sets the breakthrough transition animator
func (r *Renderer) SetBreakthroughTransition(tr *transition.Breakthrough) {
	r.breakthroughTr = tr
}

// SetConfig updates the fractal configuration
func (r *Renderer) SetConfig(config persistence.Config) {
	r.config = config
}

// SetParallel enables or disables parallel rendering
func (r *Renderer) SetParallel(enable bool) {
	r.enableParallel = enable
}

// ClearCache clears the iteration cache to free memory
func (r *Renderer) ClearCache() {
	r.iterationCache.Clear()
}

// CacheStats returns cache performance statistics (hits, misses, hit rate)
func (r *Renderer) CacheStats() (hits, misses int64, hitRate float64) {
	return r.iterationCache.Stats()
}

// RenderFractal generates the ASCII fractal output with caching and parallel computation
func (r *Renderer) RenderFractal() string {
	grid, colorGrid := r.cellGrid(r.config)
	return r.cellString(grid, colorGrid, r.config)
}

// RenderCrossfade dissolves between the renderings of two views at
// mix (0 = all source, 1 = all destination), picking per cell
// against a fixed dither so the dissolve is stable as mix grows.
func (r *Renderer) RenderCrossfade(src, dst persistence.Config, mix float64) string {
	srcGrid, srcColors := r.cellGrid(src)
	dstGrid, dstColors := r.cellGrid(dst)

	height := len(srcGrid)
	width := len(srcGrid[0])
	grid := make([][]byte, height)
	colorGrid := make([][]string, height)
	for row := range height {
		grid[row] = make([]byte, width)
		colorGrid[row] = make([]string, width)
		for col := range width {
			if cellDither(row, col) < mix {
				grid[row][col] = dstGrid[row][col]
				colorGrid[row][col] = dstColors[row][col]
			} else {
				grid[row][col] = srcGrid[row][col]
				colorGrid[row][col] = srcColors[row][col]
			}
		}
	}
	return r.cellString(grid, colorGrid, src)
}

// cellDither returns a stable pseudo-random threshold in [0,1) for a
// cell, so a dissolve reveals the same cells first as mix grows.
func cellDither(row, col int) float64 {
	h := uint32(row)*0x9E3779B9 ^ uint32(col)*0x85EBCA6B
	h ^= h >> 16
	h *= 0x7FEB352D
	h ^= h >> 15
	return float64(h) / float64(^uint32(0))
}

// cellGrid renders the config to per-cell characters and colors.
func (r *Renderer) cellGrid(cfg persistence.Config) (grid [][]byte, colorGrid [][]string) {
	iterGrid := r.iterationsFor(cfg)

	grid = make([][]byte, cfg.Height)
	colorGrid = make([][]string, cfg.Height)
	for i := range grid {
		grid[i] = make([]byte, cfg.Width)
		colorGrid[i] = make([]string, cfg.Width)
	}

	// Convert iterations to characters and colors
	for row := range cfg.Height {
		for col := range cfg.Width {
			iter := iterGrid[row][col]
			char := GetChar(iter, cfg.MaxIter)
			var ansiColor string

			scheme := cfg.ColorScheme
			if !r.colorMode {
				scheme = color.ColorGrayscale
			}

			if r.dynamicColor && r.colorMode {
				ansiColor = color.GetColorWithHueShift(iter, cfg.MaxIter, scheme, r.hueShift)
			} else {
				ansiColor = color.GetColor(iter, cfg.MaxIter, scheme)
			}

			grid[row][col] = char
			colorGrid[row][col] = ansiColor
		}
	}

	// Apply breakthrough transition overlay if active
	if r.breakthroughTr != nil {
		applyBreakthroughOverlay(grid, colorGrid, cfg.Width, cfg.Height, r.breakthroughTr)
	}
	return grid, colorGrid
}

// cellString serializes a cell grid to colored output.
func (r *Renderer) cellString(grid [][]byte, colorGrid [][]string, cfg persistence.Config) string {
	var output strings.Builder
	for row := range cfg.Height {
		for col := range cfg.Width {
			color := colorGrid[row][col]
			char := grid[row][col]

			if color != "" {
				fmt.Fprintf(&output, "%s%c", color, char)
			} else {
				output.WriteByte(char)
			}
		}

		// Reset color at end of line and add newline
		if cfg.ColorScheme != color.ColorGrayscale {
			output.WriteString("\033[0m")
		}
		if row < cfg.Height-1 {
			output.WriteByte('\n')
		}
	}

	return output.String()
}

// iterationsFor computes the iteration counts for cfg, reusing the
// cache of recent viewports when possible.
func (r *Renderer) iterationsFor(cfg persistence.Config) [][]int {
	if cached, found := r.iterationCache.Get(cfg); found {
		return cached
	}

	var iterGrid [][]int
	if r.enableParallel && cfg.Height*cfg.Width > 1000 { // Only parallelize for large grids
		iterGrid = r.calculateIterationsParallel(cfg)
	} else {
		iterGrid = r.calculateIterationsSerial(cfg)
	}

	r.iterationCache.Set(cfg, iterGrid)
	return iterGrid
}

// calculateIterationsSerial computes iteration counts serially
func (r *Renderer) calculateIterationsSerial(cfg persistence.Config) [][]int {
	grid := make([][]int, cfg.Height)
	for row := range cfg.Height {
		grid[row] = make([]int, cfg.Width)
		for col := range cfg.Width {
			cr, ci := MapToComplex(col, row, cfg.Width, cfg.Height,
				cfg.CenterX, cfg.CenterY, cfg.Zoom)
			grid[row][col] = int(r.calculateFractal(cr, ci, cfg))
		}
	}
	return grid
}

// calculateIterationsParallel computes iteration counts in parallel using goroutines
func (r *Renderer) calculateIterationsParallel(cfg persistence.Config) [][]int {
	grid := make([][]int, cfg.Height)
	for i := range grid {
		grid[i] = make([]int, cfg.Width)
	}

	// Use a work queue pattern to distribute rows across workers
	rowChan := make(chan int, r.numWorkers*2)
	var wg sync.WaitGroup

	// Start worker goroutines
	for worker := 0; worker < r.numWorkers; worker++ {
		wg.Go(func() {
			for row := range rowChan {
				for col := range cfg.Width {
					cr, ci := MapToComplex(col, row, cfg.Width, cfg.Height,
						cfg.CenterX, cfg.CenterY, cfg.Zoom)
					grid[row][col] = int(r.calculateFractal(cr, ci, cfg))
				}
			}
		})
	}

	// Send rows to workers
	for row := range cfg.Height {
		rowChan <- row
	}
	close(rowChan)
	wg.Wait()

	return grid
}

// MapToComplex converts terminal coordinates to complex plane coordinates
func MapToComplex(col, row, width, height int, centerX, centerY, zoom float64) (float64, float64) {
	// Normalized coordinates (0 to 1)
	normX := float64(col) / float64(width)
	normY := float64(row) / float64(height)

	// Default view is 3.5 units wide on real axis
	viewWidth := 3.5 / zoom
	viewHeight := viewWidth * float64(height) / float64(width) / 2.0 // Account for terminal aspect ratio

	// Convert to complex plane
	cr := centerX + (normX-0.5)*viewWidth
	ci := centerY + (normY-0.5)*viewHeight

	return cr, ci
}

// GetChar selects ASCII character based on iteration count
func GetChar(iter, maxIter int) byte {
	asciiChars := " .:-=+*#%@"
	if iter >= maxIter {
		return asciiChars[len(asciiChars)-1]
	}
	idx := (iter * len(asciiChars)) / maxIter
	return asciiChars[idx]
}

// applyBreakthroughOverlay applies the breakthrough transition visual effect
func applyBreakthroughOverlay(grid [][]byte, colorGrid [][]string, width, height int, b *transition.Breakthrough) {
	particles := b.GetParticles()
	crackMap := b.GetCrackMap()

	for _, p := range particles {
		col := int(p.X * float64(width))
		row := int(p.Y * float64(height))

		if row >= 0 && row < height && col >= 0 && col < width {
			if p.Opacity > 0.7 {
				grid[row][col] = '#'
				colorGrid[row][col] = "\033[90m"
			} else if p.Opacity > 0.4 {
				grid[row][col] = '%'
				colorGrid[row][col] = "\033[37m"
			} else {
				grid[row][col] = '+'
				colorGrid[row][col] = "\033[90m"
			}
		}
	}

	for coord := range crackMap {
		var x, y float64
		fmt.Sscanf(coord, "%f,%f", &x, &y)

		col := int(x * float64(width))
		row := int(y * float64(height))

		if row >= 0 && row < height && col >= 0 && col < width {
			grid[row][col] = '/'
			colorGrid[row][col] = "\033[37m"
		}
	}
}
