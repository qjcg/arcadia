package render

import (
	"fmt"
	"runtime"
	"strings"
	"sync"

	"github.com/qjcg/arcadia/x/fractalis/internal/core/persistence"
	"github.com/qjcg/arcadia/x/fractalis/internal/tui/cache"
	"github.com/qjcg/arcadia/x/fractalis/internal/tui/color"
	"github.com/qjcg/arcadia/x/fractalis/internal/tui/transition"
)

// Renderer handles ASCII fractal rendering with caching and parallel computation
type Renderer struct {
	config           persistence.Config
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
	// Check cache first
	if cached, found := r.iterationCache.Get(r.config); found {
		return r.renderFromIterations(cached)
	}

	// Calculate iterations (either parallel or serial)
	var iterGrid [][]int
	if r.enableParallel && r.config.Height*r.config.Width > 1000 { // Only parallelize for large grids
		iterGrid = r.calculateIterationsParallel()
	} else {
		iterGrid = r.calculateIterationsSerial()
	}

	// Cache the result
	r.iterationCache.Set(r.config, iterGrid)

	return r.renderFromIterations(iterGrid)
}

// calculateIterationsSerial computes iteration counts serially
func (r *Renderer) calculateIterationsSerial() [][]int {
	grid := make([][]int, r.config.Height)
	for row := range r.config.Height {
		grid[row] = make([]int, r.config.Width)
		for col := range r.config.Width {
			cr, ci := MapToComplex(col, row, r.config.Width, r.config.Height,
				r.config.CenterX, r.config.CenterY, r.config.Zoom)
			grid[row][col] = int(r.calculateFractal(cr, ci, r.config))
		}
	}
	return grid
}

// calculateIterationsParallel computes iteration counts in parallel using goroutines
func (r *Renderer) calculateIterationsParallel() [][]int {
	grid := make([][]int, r.config.Height)
	for i := range grid {
		grid[i] = make([]int, r.config.Width)
	}

	// Use a work queue pattern to distribute rows across workers
	rowChan := make(chan int, r.numWorkers*2)
	var wg sync.WaitGroup

	// Start worker goroutines
	for worker := 0; worker < r.numWorkers; worker++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for row := range rowChan {
				for col := range r.config.Width {
					cr, ci := MapToComplex(col, row, r.config.Width, r.config.Height,
						r.config.CenterX, r.config.CenterY, r.config.Zoom)
					grid[row][col] = int(r.calculateFractal(cr, ci, r.config))
				}
			}
		}()
	}

	// Send rows to workers
	for row := range r.config.Height {
		rowChan <- row
	}
	close(rowChan)
	wg.Wait()

	return grid
}

// renderFromIterations converts iteration grid to colored ASCII output
func (r *Renderer) renderFromIterations(iterGrid [][]int) string {
	// Create grid for character map
	grid := make([][]byte, r.config.Height)
	colorGrid := make([][]string, r.config.Height)
	for i := range grid {
		grid[i] = make([]byte, r.config.Width)
		colorGrid[i] = make([]string, r.config.Width)
	}

	// Convert iterations to characters and colors
	for row := range r.config.Height {
		for col := range r.config.Width {
			iter := iterGrid[row][col]
			char := GetChar(iter, r.config.MaxIter)
			var ansiColor string
			if r.dynamicColor {
				ansiColor = color.GetColorWithHueShift(iter, r.config.MaxIter, r.config.ColorScheme, r.hueShift)
			} else {
				ansiColor = color.GetColor(iter, r.config.MaxIter, r.config.ColorScheme)
			}

			grid[row][col] = char
			colorGrid[row][col] = ansiColor
		}
	}

	// Apply breakthrough transition overlay if active
	if r.breakthroughTr != nil {
		applyBreakthroughOverlay(grid, colorGrid, r.config.Width, r.config.Height, r.breakthroughTr)
	}

	// Convert grid to string output
	var output strings.Builder
	for row := range r.config.Height {
		for col := range r.config.Width {
			color := colorGrid[row][col]
			char := grid[row][col]

			if color != "" {
				fmt.Fprintf(&output, "%s%c", color, char)
			} else {
				output.WriteByte(char)
			}
		}

		// Reset color at end of line and add newline
		if r.config.ColorScheme != color.ColorGrayscale {
			output.WriteString("\033[0m")
		}
		if row < r.config.Height-1 {
			output.WriteByte('\n')
		}
	}

	return output.String()
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
