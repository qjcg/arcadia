package main

import (
	"flag"
	"fmt"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"gopkg.in/yaml.v3"
)

const (
	Version = "v1.0.0"

	// ASCII characters from sparse to dense
	asciiChars = " .:-=+*#%@"

	// Transition animation modes
	TransitionNone      = 0
	TransitionFade      = 1
	TransitionZoomOutIn = 2
	TransitionRotate    = 3
)

// Color scheme types
const (
	ColorGrayscale = "grayscale"
	ColorBlue      = "blue"
	ColorRainbow   = "rainbow"
	ColorFire      = "fire"
	ColorPurple    = "purple"
	ColorGreen     = "green"
	ColorGold      = "gold"
	ColorCyan      = "cyan"
)

// Fractal types
const (
	FractalMandelbrot    = "mandelbrot"
	FractalJulia         = "julia"
	FractalBurningShip   = "burningship"
	FractalTricorn       = "tricorn"
	FractalMultibrot3    = "multibrot3"
	FractalMultibrot4    = "multibrot4"
	FractalCeltic        = "celtic"
	FractalPerpendicular = "perpendicular"
	FractalMultibrot5    = "multibrot5"
	FractalManhattan     = "manhattan"
	FractalNewton        = "newton"
)

// Word lists for auto-generated bookmark names
var (
	showVersion bool

	adjectives = []string{
		"uncharted", "distant", "forgotten", "ancient", "sacred", "secret",
		"mystic", "endless", "infinite", "twilight", "crystal", "phantom",
		"hidden", "silent", "veiled", "ethereal", "serene", "tranquil",
		"arcane", "celestial", "luminous", "radiant", "shadowed", "misty",
		"frosted", "gilded", "bronze", "silver", "crimson", "azure",
		"jade", "amber", "obsidian", "prismatic", "shimmering", "glowing",
		"gleaming", "whispering", "echoing", "wandering",
	}

	journeyNouns = []string{
		"path", "journey", "expedition", "quest", "voyage", "passage",
		"gateway", "portal", "crossing", "frontier", "realm", "domain",
		"territory", "landscape", "horizon", "threshold", "border", "edge",
		"brink", "verge", "summit", "peak", "valley", "canyon",
		"cavern", "hollow", "chamber", "sanctum", "haven", "refuge",
		"shelter", "oasis", "crossroads", "junction", "nexus", "confluence",
		"convergence", "labyrinth", "maze", "corridor",
	}

	// All fractal types for random selection
	allFractalTypes = []string{
		FractalMandelbrot, FractalJulia, FractalBurningShip, FractalTricorn,
		FractalMultibrot3, FractalMultibrot4, FractalCeltic, FractalPerpendicular,
		FractalNewton,
	}

	// All color schemes for random selection
	allColorSchemes = []string{
		ColorGrayscale, ColorBlue, ColorRainbow, ColorFire,
		ColorPurple, ColorGreen, ColorGold, ColorCyan,
	}

	// Known interesting coordinates for hyperrandom exploration
	// These are "seed points" near interesting fractal features
	interestingMandelbrot = []struct{ x, y float64 }{
		{-0.7436, 0.1314}, // Spiral region
		{-0.16, 1.0405},   // Elephant valley
		{-0.7269, 0.1889}, // Seahorse valley
		{-0.7453, 0.1127}, // Double spiral
		{-0.1592, 1.0317}, // North region
		{-0.7746, 0.1152}, // Tendril region
		{0.2805, 0.0089},  // East mini-mandelbrot
		{-0.1011, 0.9563}, // North filament
		{-0.7500, 0.0000}, // Main body edge
		{-0.1600, 1.0400}, // Antenna region
	}

	interestingJulia = []struct{ cr, ci float64 }{
		{-0.4, 0.6},       // Dendrite
		{0.285, 0.01},     // Douady rabbit
		{-0.8, 0.156},     // San Marco dragon
		{-0.7, 0.27},      // Classic
		{-0.835, -0.2321}, // Siegel disk
		{-0.123, 0.745},   // Douady rabbit variant
		{0.3, 0.5},        // Swirls
		{-0.79, 0.15},     // Dragon variant
		{-0.162, 1.04},    // Spiral
		{0.355, 0.355},    // Cross pattern
	}
)

// Config holds the rendering configuration
type Config struct {
	Width       int
	Height      int
	MaxIter     int
	CenterX     float64
	CenterY     float64
	Zoom        float64
	ColorScheme string
	FractalType string
	// Julia set parameters (c = JuliaCr + JuliaCi*i)
	JuliaCr float64
	JuliaCi float64
}

// Bookmark represents a saved fractal location
type Bookmark struct {
	Name        string  `yaml:"name"`
	FractalType string  `yaml:"fractal_type"`
	CenterX     float64 `yaml:"center_x"`
	CenterY     float64 `yaml:"center_y"`
	Zoom        float64 `yaml:"zoom"`
	MaxIter     int     `yaml:"max_iter"`
	ColorScheme string  `yaml:"color_scheme"`
	JuliaCr     float64 `yaml:"julia_cr,omitempty"`
	JuliaCi     float64 `yaml:"julia_ci,omitempty"`
}

// BookmarkList holds all bookmarks
type BookmarkList struct {
	Bookmarks []Bookmark `yaml:"bookmarks"`
}

// tickMsg is sent on each animation tick for auto-zoom
type tickMsg time.Time

// model holds the Bubble Tea application state
type model struct {
	config            Config
	showHelp          bool
	ready             bool
	termWidth         int
	termHeight        int
	lastRender        string
	interactive       bool // When false, run in legacy mode
	autoZoom          bool // Auto-zoom mode active
	autoZoomDirection int  // +1 for zoom in, -1 for zoom out
	// Auto-pilot state
	targetX     float64 // Target point to pan toward
	targetY     float64
	hasTarget   bool    // Whether we have a target to pan toward
	panProgress float64 // 0.0 to 1.0, how far we've panned toward target
	baseMaxIter int     // Base iteration count (for adaptive scaling)
	// Transition animation state
	transitionMode      int     // 0=none, 1=fade, 2=zoom_out_in, 3=rotate
	transitionProgress  float64 // 0.0 to 1.0, progress through transition
	transitionTarget    string  // Target fractal type for transition
	transitionZoomStart float64 // Starting zoom level for transition
	// Dynamic color state
	dynamicColor bool    // Enable smooth hue rotation
	hueShift     float64 // Current hue shift in degrees (0-360)
	// Bookmark state
	showBookmarks     bool       // Show bookmark list
	bookmarks         []Bookmark // Loaded bookmarks
	bookmarkCursor    int        // Selected bookmark in list
	savingBookmark    bool       // Prompting for bookmark name
	bookmarkInput     string     // User input for bookmark name
	suggestedBookmark string     // Auto-generated bookmark name suggestion
	// Screenshot state
	screenshotMsg   string // Message to display after screenshot
	screenshotTimer int    // Countdown for hiding screenshot message
	// Random state
	randomMsg   string // Message to display after randomization
	randomTimer int    // Countdown for hiding random message
}

// Init initializes the Bubble Tea model
func (m model) Init() tea.Cmd {
	return nil
}

// getBookmarkPath returns the path to the bookmarks file
func getBookmarkPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	configDir := filepath.Join(homeDir, ".config", "fractals")

	// Create directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		return "", err
	}

	return filepath.Join(configDir, "bookmarks.yaml"), nil
}

// loadBookmarks reads bookmarks from the YAML file
func loadBookmarks() ([]Bookmark, error) {
	path, err := getBookmarkPath()
	if err != nil {
		return nil, err
	}

	// If file doesn't exist, return empty list
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return []Bookmark{}, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var list BookmarkList
	if err := yaml.Unmarshal(data, &list); err != nil {
		return nil, err
	}

	return list.Bookmarks, nil
}

// saveBookmarks writes bookmarks to the YAML file
func saveBookmarks(bookmarks []Bookmark) error {
	path, err := getBookmarkPath()
	if err != nil {
		return err
	}

	list := BookmarkList{Bookmarks: bookmarks}
	data, err := yaml.Marshal(&list)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0o644)
}

// addBookmark adds a new bookmark and saves to file
func (m *model) addBookmark(name string) error {
	bookmark := Bookmark{
		Name:        name,
		FractalType: m.config.FractalType,
		CenterX:     m.config.CenterX,
		CenterY:     m.config.CenterY,
		Zoom:        m.config.Zoom,
		MaxIter:     m.config.MaxIter,
		ColorScheme: m.config.ColorScheme,
		JuliaCr:     m.config.JuliaCr,
		JuliaCi:     m.config.JuliaCi,
	}

	m.bookmarks = append(m.bookmarks, bookmark)
	return saveBookmarks(m.bookmarks)
}

// loadBookmark applies a bookmark to the current config
func (m *model) loadBookmark(index int) {
	if index < 0 || index >= len(m.bookmarks) {
		return
	}

	bm := m.bookmarks[index]
	m.config.FractalType = bm.FractalType
	m.config.CenterX = bm.CenterX
	m.config.CenterY = bm.CenterY
	m.config.Zoom = bm.Zoom
	m.config.MaxIter = bm.MaxIter
	m.config.ColorScheme = bm.ColorScheme
	m.config.JuliaCr = bm.JuliaCr
	m.config.JuliaCi = bm.JuliaCi

	// Reset base max iter for adaptive scaling
	m.baseMaxIter = bm.MaxIter
}

// generateBookmarkName creates a random name in format: adjective_noun
func generateBookmarkName() string {
	// Seed random with current time for variety
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	adjective := adjectives[rng.Intn(len(adjectives))]
	noun := journeyNouns[rng.Intn(len(journeyNouns))]

	return adjective + "_" + noun
}

// applyRandomFractal switches to a random fractal type with default settings
func (m *model) applyRandomFractal() {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Pick random fractal type
	fractalType := allFractalTypes[rng.Intn(len(allFractalTypes))]
	m.config.FractalType = fractalType

	// Reset to default position and zoom for that fractal
	m.config.Zoom = 1.0
	if fractalType == FractalJulia {
		m.config.CenterX = 0.0
		m.config.CenterY = 0.0
		// Use default Julia parameters
		m.config.JuliaCr = -0.7
		m.config.JuliaCi = 0.27015
	} else if fractalType == FractalBurningShip {
		m.config.CenterX = -0.5
		m.config.CenterY = -0.6
	} else if fractalType == FractalNewton {
		m.config.CenterX = 0.0
		m.config.CenterY = 0.0
	} else if fractalType == FractalMultibrot5 || fractalType == FractalManhattan {
		m.config.CenterX = -0.5
		m.config.CenterY = 0.0
	} else {
		m.config.CenterX = -0.5
		m.config.CenterY = 0.0
	}

	// Keep current iteration and color settings
	m.randomMsg = fmt.Sprintf("Random: %s", fractalType)
	m.randomTimer = 60
}

// Test helper function to verify transition functionality
func (m *model) testTransitionSetup() {
	// This is a test helper that sets up a transition for testing
	if len(allFractalTypes) > 1 {
		m.transitionMode = TransitionFade
		m.transitionTarget = allFractalTypes[0]
		if m.config.FractalType == m.transitionTarget && len(allFractalTypes) > 1 {
			m.transitionTarget = allFractalTypes[1]
		}
		m.transitionProgress = 0.5
		m.transitionZoomStart = m.config.Zoom
	}
}

// startFractalTransition initiates a transition to a new fractal type
func (m *model) startFractalTransition() {
	rng := rand.New(rand.New(rand.NewSource(time.Now().UnixNano())))

	// Pick a random fractal type (different from current)
	currentIndex := -1
	for i, ft := range allFractalTypes {
		if ft == m.config.FractalType {
			currentIndex = i
			break
		}
	}

	// Find a different fractal type
	targetIndex := currentIndex
	for targetIndex == currentIndex {
		targetIndex = rng.Intn(len(allFractalTypes))
	}

	m.transitionTarget = allFractalTypes[targetIndex]
	m.transitionProgress = 0.01 // Start slightly above 0 so animation triggers
	m.transitionZoomStart = m.config.Zoom

	// Set appropriate message based on transition type
	if m.transitionMode == TransitionFade {
		m.randomMsg = fmt.Sprintf("Transitioning to %s (Fade)", m.transitionTarget)
	} else if m.transitionMode == TransitionZoomOutIn {
		m.randomMsg = fmt.Sprintf("Transitioning to %s (Zoom Out/In)", m.transitionTarget)
	} else if m.transitionMode == TransitionRotate {
		m.randomMsg = fmt.Sprintf("Transitioning to %s (Rotate)", m.transitionTarget)
	}
	m.randomTimer = 60
}

// resetToDefaultState resets the fractal configuration to default values for a given fractal type
func (m *model) resetToDefaultState(fractalType string) {
	m.config.Zoom = 1.0
	// Reset iteration count to base value
	if m.baseMaxIter > 0 {
		m.config.MaxIter = m.baseMaxIter
	} else {
		m.config.MaxIter = 50 // Default base iterations
	}
	// Reset to default position for the fractal type
	if fractalType == FractalJulia {
		m.config.CenterX = 0.0
		m.config.CenterY = 0.0
		// Reset Julia parameters to defaults
		m.config.JuliaCr = -0.7
		m.config.JuliaCi = 0.27015
	} else if fractalType == FractalBurningShip {
		m.config.CenterX = -0.5
		m.config.CenterY = -0.6
	} else if fractalType == FractalNewton {
		m.config.CenterX = 0.0
		m.config.CenterY = 0.0
	} else {
		m.config.CenterX = -0.5
		m.config.CenterY = 0.0
	}
}

// applyHyperrandom generates a completely random interesting view
func (m *model) applyHyperrandom() {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	const maxAttempts = 5

	for attempt := 0; attempt < maxAttempts; attempt++ {
		// Random fractal type
		fractalType := allFractalTypes[rng.Intn(len(allFractalTypes))]
		m.config.FractalType = fractalType

		// Random color scheme
		m.config.ColorScheme = allColorSchemes[rng.Intn(len(allColorSchemes))]

		// Random zoom (log-uniform between 1.0 and 1000.0, weighted toward mid-range)
		// Using exponential distribution: zoom = 10^(uniform(0, 3))
		logZoom := rng.Float64() * 3.0 // 0 to 3
		zoom := math.Pow(10.0, logZoom)
		m.config.Zoom = zoom

		// Random iterations appropriate for zoom level
		// Base: 50-150, plus bonus for higher zoom
		baseIter := 50 + rng.Intn(100)
		zoomBonus := int(math.Log10(zoom) * 20.0)
		m.config.MaxIter = baseIter + zoomBonus
		if m.config.MaxIter > 300 {
			m.config.MaxIter = 300
		}
		m.baseMaxIter = baseIter

		// Random position based on fractal type
		if fractalType == FractalJulia {
			// Random Julia parameter from interesting set
			juliaPt := interestingJulia[rng.Intn(len(interestingJulia))]
			m.config.JuliaCr = juliaPt.cr
			m.config.JuliaCi = juliaPt.ci

			// Random viewing position with small offset
			offset := 0.5 / zoom
			m.config.CenterX = (rng.Float64()*2.0 - 1.0) * offset
			m.config.CenterY = (rng.Float64()*2.0 - 1.0) * offset

		} else if fractalType == FractalMandelbrot || fractalType == FractalMultibrot3 ||
			fractalType == FractalMultibrot4 || fractalType == FractalCeltic ||
			fractalType == FractalPerpendicular {
			// Use interesting Mandelbrot coordinates as seed
			seedPt := interestingMandelbrot[rng.Intn(len(interestingMandelbrot))]

			// Add random offset proportional to zoom
			maxOffset := 0.3 / zoom
			m.config.CenterX = seedPt.x + (rng.Float64()*2.0-1.0)*maxOffset
			m.config.CenterY = seedPt.y + (rng.Float64()*2.0-1.0)*maxOffset

		} else if fractalType == FractalBurningShip {
			// Burning ship interesting region
			m.config.CenterX = -0.5 + (rng.Float64()*2.0-1.0)*0.3/zoom
			m.config.CenterY = -0.6 + (rng.Float64()*2.0-1.0)*0.3/zoom

		} else if fractalType == FractalTricorn {
			// Tricorn interesting region
			m.config.CenterX = -0.5 + (rng.Float64()*2.0-1.0)*0.3/zoom
			m.config.CenterY = 0.0 + (rng.Float64()*2.0-1.0)*0.3/zoom
		}

		// Check if the result is interesting
		if !m.isViewUniform() {
			// Success! Found an interesting view
			m.randomMsg = fmt.Sprintf("Hyperrandom: %s @ %.1fx", fractalType, zoom)
			m.randomTimer = 60
			return
		}
	}

	// If all attempts failed, fall back to a known-good interesting point
	// Use first Mandelbrot interesting point
	seedPt := interestingMandelbrot[0]
	m.config.FractalType = FractalMandelbrot
	m.config.CenterX = seedPt.x
	m.config.CenterY = seedPt.y
	m.config.Zoom = 10.0 + rng.Float64()*90.0 // 10x to 100x
	m.config.MaxIter = 100 + rng.Intn(100)
	m.config.ColorScheme = allColorSchemes[rng.Intn(len(allColorSchemes))]

	m.randomMsg = fmt.Sprintf("Hyperrandom: %s @ %.1fx (fallback)", m.config.FractalType, m.config.Zoom)
	m.randomTimer = 60
}

// deleteBookmark removes a bookmark at the specified index and saves the updated list
func (m *model) deleteBookmark(index int) error {
	if index < 0 || index >= len(m.bookmarks) {
		return fmt.Errorf("invalid bookmark index: %d", index)
	}

	// Remove bookmark at index
	m.bookmarks = append(m.bookmarks[:index], m.bookmarks[index+1:]...)

	// Adjust cursor position if needed
	if m.bookmarkCursor >= len(m.bookmarks) && m.bookmarkCursor > 0 {
		m.bookmarkCursor = len(m.bookmarks) - 1
	}

	// Save updated bookmarks to file
	return saveBookmarks(m.bookmarks)
}

// getScreenshotPath returns the directory path for screenshots
func getScreenshotPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	screenshotDir := filepath.Join(homeDir, ".config", "fractals", "screenshots")

	// Create directory if it doesn't exist
	if err := os.MkdirAll(screenshotDir, 0o755); err != nil {
		return "", err
	}

	return screenshotDir, nil
}

// saveScreenshot saves the current fractal view to a text file
func (m *model) saveScreenshot() error {
	dir, err := getScreenshotPath()
	if err != nil {
		return err
	}

	// Generate filename with timestamp and fractal type
	now := time.Now()
	timestamp := now.Format("2006-01-02_150405")
	filename := fmt.Sprintf("%s_%s.txt", m.config.FractalType, timestamp)
	fullPath := filepath.Join(dir, filename)

	// Check if file exists and add counter if needed
	counter := 1
	for {
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			break
		}
		filename = fmt.Sprintf("%s_%s_%d.txt", m.config.FractalType, timestamp, counter)
		fullPath = filepath.Join(dir, filename)
		counter++
	}

	// Render the fractal
	fractalOutput := m.renderFractal()

	// Create metadata header
	var metadata strings.Builder
	metadata.WriteString("=" + strings.Repeat("=", 78) + "\n")
	metadata.WriteString(fmt.Sprintf("Fractal Screenshot - %s\n", now.Format("2006-01-02 15:04:05")))
	metadata.WriteString("=" + strings.Repeat("=", 78) + "\n")
	metadata.WriteString(fmt.Sprintf("Fractal Type: %s\n", m.config.FractalType))
	metadata.WriteString(fmt.Sprintf("Center: (%.10f, %.10f)\n", m.config.CenterX, m.config.CenterY))

	// Format zoom display
	var zoomStr string
	if m.config.Zoom >= 10000.0 {
		zoomStr = fmt.Sprintf("%.6e", m.config.Zoom)
	} else {
		zoomStr = fmt.Sprintf("%.6f", m.config.Zoom)
	}
	metadata.WriteString(fmt.Sprintf("Zoom: %sx\n", zoomStr))
	metadata.WriteString(fmt.Sprintf("Max Iterations: %d\n", m.config.MaxIter))
	metadata.WriteString(fmt.Sprintf("Color Scheme: %s\n", m.config.ColorScheme))

	if m.config.FractalType == FractalJulia {
		metadata.WriteString(fmt.Sprintf("Julia Parameters: c = %.6f + %.6fi\n", m.config.JuliaCr, m.config.JuliaCi))
	}

	metadata.WriteString(fmt.Sprintf("Resolution: %dx%d\n", m.config.Width, m.config.Height))
	metadata.WriteString("=" + strings.Repeat("=", 78) + "\n\n")

	// Combine metadata and fractal output
	content := metadata.String() + fractalOutput

	// Write to file
	if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
		return err
	}

	// Set success message
	m.screenshotMsg = fmt.Sprintf("Screenshot saved: %s", filename)
	m.screenshotTimer = 60 // Show message for ~3 seconds (60 ticks at 50ms each)

	return nil
}

// tickCmd sends a tick message after a delay for smooth animation
func tickCmd() tea.Cmd {
	return tea.Tick(time.Millisecond*50, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// calculateAdaptiveMaxIter returns an appropriate iteration count based on zoom level
// As we zoom deeper, we need more iterations to reveal detail at the boundary
func (m model) calculateAdaptiveMaxIter() int {
	// Base iteration count
	baseIter := m.baseMaxIter
	if baseIter == 0 {
		baseIter = 50 // Default if not set
	}

	// Increase iterations logarithmically with zoom
	// Formula: baseIter + log10(zoom) * scaleFactor
	// At zoom=1: baseIter
	// At zoom=100: baseIter + 40
	// At zoom=10000: baseIter + 80
	// At zoom=1000000: baseIter + 120
	if m.config.Zoom > 1.0 {
		logZoom := math.Log10(m.config.Zoom)
		scaleFactor := 20.0 // Add 20 iterations per decade of zoom
		adaptiveIter := baseIter + int(logZoom*scaleFactor)

		// Cap at reasonable maximum to avoid performance issues
		if adaptiveIter > 2000 {
			adaptiveIter = 2000
		}
		return adaptiveIter
	}

	return baseIter
}

// getEffectiveSearchDelta returns a search distance that's meaningful at current zoom
// Ensures we don't lose precision at extreme zoom levels
func (m model) getEffectiveSearchDelta() float64 {
	// At high zoom, floating point precision becomes an issue
	// We need to ensure our delta is large enough to actually change the value
	// float64 has ~15-17 decimal digits of precision

	baseViewSize := 3.5 / m.config.Zoom // Current view width

	// Use a delta that's at least 1/1000 of view size, but not smaller than
	// what float64 can reliably represent relative to current center coordinates
	minDelta := 1e-14 // Near the limit of float64 precision

	// Scale delta with view size
	delta := baseViewSize / 20.0 // 5% of view size

	// But ensure it's meaningful given our coordinate magnitudes
	coordMagnitude := m.config.CenterX
	if coordMagnitude < 0 {
		coordMagnitude = -coordMagnitude
	}
	if coordMagnitude == 0 {
		coordMagnitude = 1.0
	}

	// Ensure delta is at least 1e-12 relative to coordinate magnitude
	relativeDelta := coordMagnitude * 1e-12
	if delta < relativeDelta {
		delta = relativeDelta
	}
	if delta < minDelta {
		delta = minDelta
	}

	return delta
}

// isViewUniform checks if the current view is mostly uniform (boring)
// by sampling points and checking variation in iteration counts
func (m model) isViewUniform() bool {
	const sampleSize = 25 // Sample 25 points

	samples := make([]int, 0, sampleSize)

	// Get effective delta for sampling at current zoom
	delta := m.getEffectiveSearchDelta()

	// Sample points around the current center
	for i := 0; i < sampleSize; i++ {
		// Generate offsets using deterministic pattern
		offsetX := delta * float64(i%5-2)
		offsetY := delta * float64(i/5-2) / 2.0 // Account for aspect ratio

		cr := m.config.CenterX + offsetX
		ci := m.config.CenterY + offsetY

		iter := calculateFractal(cr, ci, m.config)
		samples = append(samples, iter)
	}

	// Calculate variance in iteration counts
	if len(samples) == 0 {
		return false
	}

	// Find min and max for range calculation
	minIter := samples[0]
	maxIter := samples[0]
	sum := 0
	for _, val := range samples {
		sum += val
		if val < minIter {
			minIter = val
		}
		if val > maxIter {
			maxIter = val
		}
	}
	mean := float64(sum) / float64(len(samples))

	// Calculate range and variance
	iterRange := float64(maxIter - minIter)

	variance := 0.0
	for _, val := range samples {
		diff := float64(val) - mean
		variance += diff * diff
	}
	variance /= float64(len(samples))
	stdDev := math.Sqrt(variance)

	// A view is uniform if:
	// 1. The range is very small (less than 5% of maxIter), OR
	// 2. The standard deviation is very small (less than 10% of maxIter)
	rangeThreshold := float64(m.config.MaxIter) * 0.05
	varianceThreshold := float64(m.config.MaxIter) * 0.10

	isUniform := iterRange < rangeThreshold || stdDev < varianceThreshold

	return isUniform
}

// findInterestingPoint searches for a point with high variation/detail
// Returns coordinates of an interesting point, or current center if none found
func (m model) findInterestingPoint() (float64, float64) {
	bestX := m.config.CenterX
	bestY := m.config.CenterY
	bestScore := 0.0

	// Get base view size for search radius calculation
	baseViewSize := 3.5 / m.config.Zoom

	// Multi-pass search with increasing radius
	// Pass 1: Local search (within 50% of view)
	searchRadius1 := baseViewSize * 0.5
	const numCandidates1 = 40

	for i := 0; i < numCandidates1; i++ {
		angle := float64(i) * 2.399                                             // Golden angle for good coverage
		radius := searchRadius1 * math.Sqrt(float64(i)/float64(numCandidates1)) // Spiral

		x := m.config.CenterX + radius*math.Cos(angle)
		y := m.config.CenterY + radius*math.Sin(angle)/2.0 // Account for aspect ratio

		score := m.calculateInterestScore(x, y)
		if score > bestScore {
			bestScore = score
			bestX = x
			bestY = y
		}
	}

	// Pass 2: Medium search if we didn't find much (within 150% of view)
	if bestScore < float64(m.config.MaxIter)*5.0 {
		searchRadius2 := baseViewSize * 1.5
		const numCandidates2 = 40

		for i := 0; i < numCandidates2; i++ {
			angle := float64(i) * 2.399
			radius := searchRadius2 * math.Sqrt(float64(i)/float64(numCandidates2))

			x := m.config.CenterX + radius*math.Cos(angle)
			y := m.config.CenterY + radius*math.Sin(angle)/2.0

			score := m.calculateInterestScore(x, y)
			if score > bestScore {
				bestScore = score
				bestX = x
				bestY = y
			}
		}
	}

	// Pass 3: Wide search if still not finding interesting areas (within 400% of view)
	if bestScore < float64(m.config.MaxIter)*2.0 {
		searchRadius3 := baseViewSize * 4.0
		const numCandidates3 = 50

		for i := 0; i < numCandidates3; i++ {
			angle := float64(i) * 2.399
			radius := searchRadius3 * math.Sqrt(float64(i)/float64(numCandidates3))

			x := m.config.CenterX + radius*math.Cos(angle)
			y := m.config.CenterY + radius*math.Sin(angle)/2.0

			score := m.calculateInterestScore(x, y)
			if score > bestScore {
				bestScore = score
				bestX = x
				bestY = y
			}
		}
	}

	return bestX, bestY
}

// calculateInterestScore determines how "interesting" a point is
// by looking at iteration variation in its neighborhood
// Higher scores = more detail/variation = more interesting
func (m model) calculateInterestScore(cx, cy float64) float64 {
	const neighborhoodSize = 9
	samples := make([]int, 0, neighborhoodSize)

	// Use effective delta for neighborhood sampling
	delta := m.getEffectiveSearchDelta() * 0.3

	// Sample a 3x3 grid around the point
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			cr := cx + float64(dx)*delta
			ci := cy + float64(dy)*delta/2.0 // Account for aspect ratio
			iter := calculateFractal(cr, ci, m.config)
			samples = append(samples, iter)
		}
	}

	// Find min and max iteration counts
	minIter := samples[0]
	maxIter := samples[0]
	sum := 0
	for _, val := range samples {
		sum += val
		if val < minIter {
			minIter = val
		}
		if val > maxIter {
			maxIter = val
		}
	}
	mean := float64(sum) / float64(len(samples))

	// Calculate actual range of iteration values (edge detection)
	iterRange := float64(maxIter - minIter)

	// Calculate variance for texture detection
	variance := 0.0
	for _, val := range samples {
		diff := float64(val) - mean
		variance += diff * diff
	}
	variance /= float64(len(samples))
	stdDev := math.Sqrt(variance)

	// Score components:
	// 1. Range score: High range = edge/boundary (most important)
	rangeScore := iterRange * 100.0

	// 2. Variance score: Variation within neighborhood
	varianceScore := stdDev * 20.0

	// 3. Boundary bonus: Prefer areas with medium iteration counts
	boundaryScore := 0.0
	avgIter := mean
	if avgIter > 0 && avgIter < float64(m.config.MaxIter) {
		// Avoid deep interior (maxIter) and far exterior (0)
		normalizedIter := avgIter / float64(m.config.MaxIter)
		// Peak score between 10-90% of maxIter
		if normalizedIter > 0.1 && normalizedIter < 0.9 {
			boundaryScore = 500.0
		} else if normalizedIter > 0.05 && normalizedIter < 0.95 {
			boundaryScore = 200.0
		}
	}

	// Heavily penalize completely uniform areas (all same iteration count)
	if iterRange == 0 {
		return 0.0
	}

	return rangeScore + varianceScore + boundaryScore
}

// Update handles messages and updates the model
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.termWidth = msg.Width
		m.termHeight = msg.Height
		m.config.Width = msg.Width
		m.config.Height = msg.Height - 3 // Reserve space for status bar and help
		m.ready = true
		return m, nil

	case tickMsg:
		// Handle auto-zoom animation tick with intelligent panning
		if m.autoZoom {
			// Update iteration count adaptively based on zoom level
			m.config.MaxIter = m.calculateAdaptiveMaxIter()

			// Check if we need a new target (no target, or reached target, or view is uniform)
			if !m.hasTarget || m.panProgress >= 1.0 || (m.panProgress > 0.5 && m.isViewUniform()) {
				// Find a new interesting point to pan toward
				newX, newY := m.findInterestingPoint()
				m.targetX = newX
				m.targetY = newY
				m.hasTarget = true
				m.panProgress = 0.0
			}

			// Smooth pan toward target
			if m.hasTarget && m.panProgress < 1.0 {
				// Move toward target gradually
				// Just move a fraction of the remaining distance each frame
				deltaX := m.targetX - m.config.CenterX
				deltaY := m.targetY - m.config.CenterY

				// Move 5% of the remaining distance per tick (smooth exponential approach)
				m.config.CenterX += deltaX * 0.05
				m.config.CenterY += deltaY * 0.05

				// Update progress based on distance remaining
				// Use precision-aware threshold
				effectiveDelta := m.getEffectiveSearchDelta()
				precisionThreshold := effectiveDelta * effectiveDelta * 0.01

				remainingDist := deltaX*deltaX + deltaY*deltaY
				if remainingDist < precisionThreshold {
					// Close enough, mark as reached
					m.panProgress = 1.0
				} else {
					// Progress increases faster at first, slower as we approach
					m.panProgress += 0.02
				}
			}

			// Zoom based on direction (about 1.05x per tick, ~60fps = 1.05^20 ≈ 2.65x per second)
			if m.autoZoomDirection > 0 {
				// Zoom in
				m.config.Zoom *= 1.05
				// Check if we hit maximum zoom and should trigger transition
				if m.config.Zoom > 1e15 {
					m.config.Zoom = 1e15
					// Trigger transition if enabled
					if m.transitionMode > 0 && m.transitionProgress == 0.0 {
						m.startFractalTransition()
					}
				}
			} else {
				// Zoom out
				m.config.Zoom *= 0.95
				// Minimum zoom limit
				if m.config.Zoom < 0.1 {
					m.config.Zoom = 0.1
					// Trigger transition if enabled
					if m.transitionMode > 0 && m.transitionProgress == 0.0 {
						m.startFractalTransition()
					}
				}
			}

			// Handle transition animations
			if m.transitionProgress > 0.0 {
				if m.transitionProgress < 1.0 {
					m.transitionProgress += 0.05 // Progress transition at 5% per tick
				}

				// Apply different transition effects based on mode
				if m.transitionMode == TransitionFade {
					// Fade transition: gradually change fractal type
					if m.transitionProgress >= 1.0 {
						// Complete the transition
						m.config.FractalType = m.transitionTarget
						m.resetToDefaultState(m.transitionTarget)
						m.transitionProgress = 0.0
					}
				} else if m.transitionMode == TransitionZoomOutIn {
					// Zoom out/in transition
					if m.transitionProgress < 0.5 {
						// Zoom out phase
						progress := m.transitionProgress / 0.5
						m.config.Zoom = m.transitionZoomStart * (1.0 - progress*0.9) // Zoom out to 10% of start
					} else {
						// Zoom in phase
						progress := (m.transitionProgress - 0.5) / 0.5
						if m.transitionProgress >= 1.0 {
							// Complete the transition
							m.config.FractalType = m.transitionTarget
							m.resetToDefaultState(m.transitionTarget)
							m.transitionProgress = 0.0
						} else {
							m.config.Zoom = 0.1*m.transitionZoomStart + progress*0.9 // Zoom in from 10% to 90% of start
						}
					}
				} else if m.transitionMode == TransitionRotate {
					// Rotate transition: spin the view while changing fractal
					if m.transitionProgress >= 1.0 {
						// Complete the transition
						m.config.FractalType = m.transitionTarget
						m.resetToDefaultState(m.transitionTarget)
						m.transitionProgress = 0.0
					} else {
						// Rotate the view during transition
						angle := m.transitionProgress * math.Pi * 2
						radius := 0.5 / m.config.Zoom
						m.config.CenterX = -0.5 + radius*math.Cos(angle)
						m.config.CenterY = 0.0 + radius*math.Sin(angle)
					}
				}
			}

			// Update dynamic color hue shift
			if m.dynamicColor {
				m.hueShift += 0.5 // Shift 0.5 degrees per tick (30 degrees/sec at 60fps, full cycle every 12 seconds)
				if m.hueShift >= 360.0 {
					m.hueShift -= 360.0
				}
			}

			// Continue the animation by sending another tick
			return m, tickCmd()
		}

		// Update dynamic color even when not auto-zooming (for manual navigation)
		if m.dynamicColor {
			m.hueShift += 0.5
			if m.hueShift >= 360.0 {
				m.hueShift -= 360.0
			}
			// Continue ticking to keep colors changing
			return m, tickCmd()
		}

		// Handle screenshot message timer even when not auto-zooming
		if m.screenshotTimer > 0 {
			m.screenshotTimer--
			if m.screenshotTimer == 0 {
				m.screenshotMsg = ""
			}
		}

		// Handle random message timer
		if m.randomTimer > 0 {
			m.randomTimer--
			if m.randomTimer == 0 {
				m.randomMsg = ""
			}
		}

		// Continue ticking if any timer is active
		if m.screenshotTimer > 0 || m.randomTimer > 0 {
			return m, tickCmd()
		}

		return m, nil

	case tea.KeyMsg:
		// Handle bookmark name input mode
		if m.savingBookmark {
			switch msg.String() {
			case "enter":
				// Use custom input if provided, otherwise use suggested name
				nameToSave := m.bookmarkInput
				if nameToSave == "" {
					nameToSave = m.suggestedBookmark
				}

				if nameToSave != "" {
					if err := m.addBookmark(nameToSave); err != nil {
						// Error saving - could show error message
						// For now just exit save mode
					}
					m.savingBookmark = false
					m.bookmarkInput = ""
					m.suggestedBookmark = ""
				}
				return m, nil
			case "esc":
				m.savingBookmark = false
				m.bookmarkInput = ""
				m.suggestedBookmark = ""
				return m, nil
			case "backspace":
				if len(m.bookmarkInput) > 0 {
					m.bookmarkInput = m.bookmarkInput[:len(m.bookmarkInput)-1]
				}
				return m, nil
			default:
				// Add character to input if it's a single character
				if len(msg.String()) == 1 {
					m.bookmarkInput += msg.String()
				}
				return m, nil
			}
		}

		// Handle bookmark list mode
		if m.showBookmarks {
			switch msg.String() {
			case "q", "esc":
				m.showBookmarks = false
				return m, nil
			case "up", "k":
				if m.bookmarkCursor > 0 {
					m.bookmarkCursor--
				}
				return m, nil
			case "down", "j":
				if m.bookmarkCursor < len(m.bookmarks)-1 {
					m.bookmarkCursor++
				}
				return m, nil
			case "enter":
				if len(m.bookmarks) > 0 {
					m.loadBookmark(m.bookmarkCursor)
					m.showBookmarks = false
				}
				return m, nil
			case "d", "x", "delete":
				// Delete the currently selected bookmark
				if len(m.bookmarks) > 0 {
					if err := m.deleteBookmark(m.bookmarkCursor); err != nil {
						// Error deleting - could show error message
						// For now just continue
					}
					// If we deleted all bookmarks, close the list
					if len(m.bookmarks) == 0 {
						m.showBookmarks = false
					}
				}
				return m, nil
			case "1", "2", "3", "4", "5", "6", "7", "8", "9":
				// Quick jump to bookmark by number
				num, _ := strconv.Atoi(msg.String())
				if num > 0 && num <= len(m.bookmarks) {
					m.loadBookmark(num - 1)
					m.showBookmarks = false
				}
				return m, nil
			}
			return m, nil
		}

		// Normal mode key handling
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return m, tea.Quit

		case "?":
			m.showHelp = !m.showHelp
			return m, nil

		case "b":
			// Start saving a bookmark with auto-generated name suggestion
			m.savingBookmark = true
			m.bookmarkInput = ""
			m.suggestedBookmark = generateBookmarkName()
			return m, nil

		case "l":
			// Show bookmark list
			// Reload bookmarks from file
			if bookmarks, err := loadBookmarks(); err == nil {
				m.bookmarks = bookmarks
			}
			m.showBookmarks = true
			m.bookmarkCursor = 0
			return m, nil

		case "p":
			// Save screenshot
			if err := m.saveScreenshot(); err != nil {
				m.screenshotMsg = fmt.Sprintf("Error saving screenshot: %v", err)
				m.screenshotTimer = 60
			}
			// Start tick to show and hide the message
			return m, tickCmd()

		case "R":
			// Random fractal type
			m.applyRandomFractal()
			return m, tickCmd()

		case "H":
			// Hyperrandom - random everything with interesting view
			m.applyHyperrandom()
			return m, tickCmd()

		case "z":
			// Toggle auto-zoom mode
			m.autoZoom = !m.autoZoom
			if m.autoZoom {
				// Store base iteration count for adaptive scaling
				if m.baseMaxIter == 0 {
					m.baseMaxIter = m.config.MaxIter
				}
				// Initialize zoom direction if not set (default to zoom in)
				if m.autoZoomDirection == 0 {
					m.autoZoomDirection = 1
				}
				// Start the animation
				return m, tickCmd()
			}
			return m, nil

		case "r":
			// Toggle/reverse auto-pilot zoom direction
			if m.autoZoomDirection >= 0 {
				m.autoZoomDirection = -1
			} else {
				m.autoZoomDirection = 1
			}
			return m, nil

		// Navigation - Pan
		case "up", "w":
			m.config.CenterY += 0.1 / m.config.Zoom
			return m, nil
		case "down", "s":
			m.config.CenterY -= 0.1 / m.config.Zoom
			return m, nil
		case "left", "a":
			m.config.CenterX -= 0.1 / m.config.Zoom
			return m, nil
		case "right", "d":
			m.config.CenterX += 0.1 / m.config.Zoom
			return m, nil

		// Zoom
		case "+", "=":
			// If auto-pilot is active, set zoom direction to IN
			// Otherwise perform manual zoom in
			if m.autoZoom {
				m.autoZoomDirection = 1
			} else {
				m.config.Zoom *= 1.2
			}
			return m, nil
		case "i":
			// Manual zoom in only
			m.config.Zoom *= 1.2
			return m, nil
		case "-", "_":
			// If auto-pilot is active, set zoom direction to OUT
			// Otherwise perform manual zoom out
			if m.autoZoom {
				m.autoZoomDirection = -1
			} else {
				m.config.Zoom /= 1.2
			}
			return m, nil
		case "o":
			// Manual zoom out only
			m.config.Zoom /= 1.2
			return m, nil
		case "0":
			// Reset zoom, center, and iteration depth to defaults
			m.config.Zoom = 1.0
			if m.config.FractalType == FractalJulia {
				m.config.CenterX = 0.0
				m.config.CenterY = 0.0
			} else {
				m.config.CenterX = -0.5
				m.config.CenterY = 0.0
			}
			// Reset iteration count to the base value (user's starting value)
			if m.baseMaxIter > 0 {
				m.config.MaxIter = m.baseMaxIter
			} else {
				m.config.MaxIter = 50 // Fallback to default
			}
			return m, nil

		// Iteration depth
		case "]":
			m.config.MaxIter += 10
			return m, nil
		case "[":
			if m.config.MaxIter > 10 {
				m.config.MaxIter -= 10
			}
			return m, nil

		// Fractal type switching
		case "1":
			m.config.FractalType = FractalMandelbrot
			m.config.CenterX = -0.5
			m.config.CenterY = 0.0
			return m, nil
		case "2":
			m.config.FractalType = FractalJulia
			m.config.CenterX = 0.0
			m.config.CenterY = 0.0
			return m, nil
		case "3":
			m.config.FractalType = FractalBurningShip
			m.config.CenterX = -0.5
			m.config.CenterY = -0.6
			return m, nil
		case "4":
			m.config.FractalType = FractalTricorn
			m.config.CenterX = -0.5
			m.config.CenterY = 0.0
			return m, nil
		case "5":
			m.config.FractalType = FractalMultibrot3
			m.config.CenterX = -0.5
			m.config.CenterY = 0.0
			return m, nil
		case "6":
			m.config.FractalType = FractalMultibrot4
			m.config.CenterX = -0.5
			m.config.CenterY = 0.0
			return m, nil
		case "7":
			m.config.FractalType = FractalCeltic
			m.config.CenterX = -0.5
			m.config.CenterY = 0.0
			return m, nil
		case "8":
			m.config.FractalType = FractalPerpendicular
			m.config.CenterX = -0.5
			m.config.CenterY = 0.0
			return m, nil
		case "9":
			m.config.FractalType = FractalMultibrot5
			m.config.CenterX = -0.5
			m.config.CenterY = 0.0
			return m, nil
		case "m":
			m.config.FractalType = FractalManhattan
			m.config.CenterX = -0.5
			m.config.CenterY = 0.0
			return m, nil
		case "n":
			m.config.FractalType = FractalNewton
			m.config.CenterX = 0.0
			m.config.CenterY = 0.0
			return m, nil
		case "c":
			// Cycle through color schemes
			currentIndex := -1
			for i, cs := range allColorSchemes {
				if cs == m.config.ColorScheme {
					currentIndex = i
					break
				}
			}
			if currentIndex == -1 {
				// Unknown color scheme, reset to first
				m.config.ColorScheme = allColorSchemes[0]
			} else {
				// Move to next color scheme
				m.config.ColorScheme = allColorSchemes[(currentIndex+1)%len(allColorSchemes)]
			}
			return m, nil

		// Julia set parameter adjustment
		case "J":
			m.config.JuliaCr += 0.05
			return m, nil
		case "j":
			m.config.JuliaCr -= 0.05
			return m, nil
		case "K":
			m.config.JuliaCi += 0.05
			return m, nil
		case "k":
			m.config.JuliaCi -= 0.05
			return m, nil
		// Transition mode switching
		case "T":
			// Cycle through transition modes
			m.transitionMode = (m.transitionMode + 1) % 4
			if m.transitionMode == 0 {
				m.randomMsg = "Transition: None"
			} else if m.transitionMode == 1 {
				m.randomMsg = "Transition: Fade"
			} else if m.transitionMode == 2 {
				m.randomMsg = "Transition: Zoom Out/In"
			} else if m.transitionMode == 3 {
				m.randomMsg = "Transition: Rotate"
			}
			m.randomTimer = 30
			return m, nil

		// Dynamic color mode toggle
		case "C":
			m.dynamicColor = !m.dynamicColor
			if m.dynamicColor {
				m.randomMsg = "Dynamic Color: ON"
				// Start the tick animation to keep colors changing
				return m, tickCmd()
			} else {
				m.randomMsg = "Dynamic Color: OFF"
				m.hueShift = 0.0 // Reset hue shift
			}
			m.randomTimer = 30
			return m, nil
		}
	}

	return m, nil
}

// View renders the current state
func (m model) View() string {
	if !m.ready {
		return "Initializing fractal viewer..."
	}

	if m.showHelp {
		return m.renderHelp()
	}

	if m.showBookmarks {
		return m.renderBookmarks()
	}

	if m.savingBookmark {
		return m.renderBookmarkInput()
	}

	// Render the fractal
	fractal := m.renderFractal()
	statusBar := m.renderStatusBar()

	return fractal + "\n" + statusBar
}

// renderHelp displays the help screen
func (m model) renderHelp() string {
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("cyan")).
		Bold(true)

	keyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("yellow")).
		Bold(true)

	var help strings.Builder
	help.WriteString(helpStyle.Render("Interactive Fractal Viewer - Keyboard Controls") + "\n\n")

	help.WriteString(keyStyle.Render("Navigation:") + "\n")
	help.WriteString("  Arrow keys / wasd  Pan view\n")
	help.WriteString("  i / o              Zoom in / out (manual)\n")
	help.WriteString("  +, =               Zoom in (or set auto-pilot → zoom in)\n")
	help.WriteString("  -, _               Zoom out (or set auto-pilot → zoom out)\n")
	help.WriteString("  z                  Toggle auto-pilot mode (intelligent zoom + pan)\n")
	help.WriteString("                     Automatically finds and explores interesting regions\n")
	help.WriteString("  r                  Reverse auto-pilot zoom direction (↑ ↔ ↓)\n")
	help.WriteString("                     Direction indicator always visible in status bar\n")
	help.WriteString("  0                  Reset view (position, zoom, and iteration depth)\n\n")

	help.WriteString(keyStyle.Render("Fractal Types:") + "\n")
	help.WriteString("  1  Mandelbrot      2  Julia          3  Burning Ship\n")
	help.WriteString("  4  Tricorn         5  Multibrot-3    6  Multibrot-4\n")
	help.WriteString("  7  Celtic          8  Perpendicular  9  Multibrot-5\n")
	help.WriteString("  m  Manhattan       n  Newton\n\n")

	help.WriteString(keyStyle.Render("Settings:") + "\n")
	help.WriteString("  c       Cycle color schemes (grayscale/blue/rainbow/fire/purple/green/gold/cyan)\n")
	help.WriteString("  C       Toggle dynamic color mode (smooth hue rotation - great with autopilot!)\n")
	help.WriteString("  [ ]     Decrease/increase iteration depth\n")
	help.WriteString("  J/j     Adjust Julia real parameter (for Julia set)\n")
	help.WriteString("  K/k     Adjust Julia imaginary parameter (for Julia set)\n\n")

	help.WriteString(keyStyle.Render("Bookmarks & Screenshots:") + "\n")
	help.WriteString("  b       Save current location as bookmark\n")
	help.WriteString("  l       Load bookmark (shows list, d/x to delete)\n")
	help.WriteString("  p       Save screenshot (text file with fractal + metadata)\n\n")

	help.WriteString(keyStyle.Render("Random Exploration:") + "\n")
	help.WriteString("  R       Random fractal type (keeps current zoom/position settings)\n")
	help.WriteString("  H       Hyperrandom (random fractal + interesting random view)\n")
	help.WriteString("          Uses AI to find non-uniform, visually interesting regions\n")
	help.WriteString("  T       Cycle transition modes (None/Fade/Zoom Out-In/Rotate)\n")
	help.WriteString("          When auto-pilot hits zoom limits, transition to new fractal\n\n")

	help.WriteString(keyStyle.Render("Other:") + "\n")
	help.WriteString("  ?       Toggle this help\n")
	help.WriteString("  q / Esc Quit\n\n")

	help.WriteString(lipgloss.NewStyle().Faint(true).Render("Press ? to return to fractal view"))

	return help.String()
}

// renderBookmarks displays the bookmark list
func (m model) renderBookmarks() string {
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("cyan")).
		Bold(true)

	selectedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("black")).
		Background(lipgloss.Color("cyan")).
		Bold(true)

	normalStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("white"))

	faintStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))

	var output strings.Builder
	output.WriteString(titleStyle.Render("Saved Bookmarks") + "\n\n")

	if len(m.bookmarks) == 0 {
		output.WriteString(faintStyle.Render("No bookmarks saved yet. Press 'b' to save current location.") + "\n\n")
	} else {
		for i, bm := range m.bookmarks {
			prefix := fmt.Sprintf("%d. ", i+1)

			// Format zoom display
			var zoomStr string
			if bm.Zoom >= 10000.0 {
				zoomStr = fmt.Sprintf("%.1e", bm.Zoom)
			} else {
				zoomStr = fmt.Sprintf("%.2f", bm.Zoom)
			}

			line := fmt.Sprintf("%s%-25s %s @ (%.4f, %.4f) zoom:%sx",
				prefix, bm.Name, bm.FractalType, bm.CenterX, bm.CenterY, zoomStr)

			if i == m.bookmarkCursor {
				output.WriteString(selectedStyle.Render(line) + "\n")
			} else {
				output.WriteString(normalStyle.Render(line) + "\n")
			}
		}
		output.WriteString("\n")
	}

	output.WriteString(faintStyle.Render("↑/↓ or j/k: Navigate | Enter: Load | d/x: Delete | 1-9: Quick load | Esc: Cancel") + "\n")

	return output.String()
}

// renderBookmarkInput displays the bookmark name input prompt
func (m model) renderBookmarkInput() string {
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("cyan")).
		Bold(true)

	promptStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("white"))

	inputStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("yellow")).
		Bold(true)

	faintStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))

	var output strings.Builder
	output.WriteString(titleStyle.Render("Save Bookmark") + "\n\n")

	// Show current location info
	var zoomStr string
	if m.config.Zoom >= 10000.0 {
		zoomStr = fmt.Sprintf("%.1e", m.config.Zoom)
	} else {
		zoomStr = fmt.Sprintf("%.2f", m.config.Zoom)
	}

	output.WriteString(faintStyle.Render(fmt.Sprintf("Location: %s @ (%.4f, %.4f) zoom:%sx\n\n",
		m.config.FractalType, m.config.CenterX, m.config.CenterY, zoomStr)))

	// Show prompt with suggested name in brackets
	suggestedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("242"))

	output.WriteString(promptStyle.Render("Name "))
	output.WriteString(suggestedStyle.Render("[" + m.suggestedBookmark + "]"))
	output.WriteString(promptStyle.Render(": "))

	// Show user input or cursor
	if m.bookmarkInput != "" {
		output.WriteString(inputStyle.Render(m.bookmarkInput+"_") + "\n\n")
	} else {
		output.WriteString(inputStyle.Render("_") + "\n\n")
	}

	output.WriteString(faintStyle.Render("Enter: Use suggested name | Type to override | Esc: Cancel") + "\n")

	return output.String()
}

// renderStatusBar creates the status bar with current settings
func (m model) renderStatusBar() string {
	infoStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("cyan"))

	// Determine direction arrow: ↑ for zoom in, ↓ for zoom out
	directionArrow := "\u2191" // ↑ (up arrow)
	if m.autoZoomDirection < 0 {
		directionArrow = "\u2193" // ↓ (down arrow)
	}

	autoZoomIndicator := ""
	if m.autoZoom {
		// Active auto-pilot: bright green background
		autoZoomStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("white")).
			Background(lipgloss.Color("green")).
			Bold(true)

		statusText := " AUTO-PILOT " + directionArrow + " "
		if m.hasTarget && m.panProgress < 1.0 {
			// Show we're panning toward an interesting region
			statusText = " AUTO-PILOT \u2192 " + directionArrow + " " // → arrow
		}
		autoZoomIndicator = autoZoomStyle.Render(statusText)
	} else {
		// Auto-pilot off: subtle indicator showing ready direction
		inactiveStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")). // Gray
			Bold(false)
		autoZoomIndicator = inactiveStyle.Render("Auto: " + directionArrow)
	}

	// Format zoom level - use scientific notation for high values
	var zoomStr string
	if m.config.Zoom >= 10000.0 {
		zoomStr = fmt.Sprintf("%.1e", m.config.Zoom)
	} else {
		zoomStr = fmt.Sprintf("%.2f", m.config.Zoom)
	}

	status := fmt.Sprintf(
		" %s | Center: (%.4f, %.4f) | Zoom: %sx | Iter: %d | Color: %s ",
		m.config.FractalType,
		m.config.CenterX,
		m.config.CenterY,
		zoomStr,
		m.config.MaxIter,
		m.config.ColorScheme,
	)

	if m.config.FractalType == FractalJulia {
		status += fmt.Sprintf("| Julia: (%.3f, %.3f) ", m.config.JuliaCr, m.config.JuliaCi)
	}

	helpText := infoStyle.Render("[? for help]")

	// Screenshot message indicator
	screenshotIndicator := ""
	if m.screenshotMsg != "" {
		msgStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("white")).
			Background(lipgloss.Color("green")).
			Bold(true)
		screenshotIndicator = msgStyle.Render(" " + m.screenshotMsg + " ")
	}

	// Random message indicator
	randomIndicator := ""
	if m.randomMsg != "" {
		msgStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("white")).
			Background(lipgloss.Color("magenta")).
			Bold(true)
		randomIndicator = msgStyle.Render(" " + m.randomMsg + " ")
	}

	// Build the complete status bar
	statusBar := status
	// Auto-pilot direction indicator (always shown)
	statusBar += autoZoomIndicator + " "
	// Screenshot message (only when present)
	if screenshotIndicator != "" {
		statusBar += screenshotIndicator + " "
	}
	// Random message (only when present)
	if randomIndicator != "" {
		statusBar += randomIndicator + " "
	}
	statusBar += helpText

	// Apply background styling to the entire status bar
	statusStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("black")).
		Background(lipgloss.Color("white")).
		Bold(true)

	return statusStyle.Render(statusBar)
}

// renderFractal generates the ASCII fractal view
func (m model) renderFractal() string {
	var output strings.Builder

	for row := 0; row < m.config.Height; row++ {
		for col := 0; col < m.config.Width; col++ {
			// Map terminal coordinates to complex plane
			cr, ci := mapToComplex(col, row, m.config.Width, m.config.Height,
				m.config.CenterX, m.config.CenterY, m.config.Zoom)

			// Calculate iteration count
			iter := calculateFractal(cr, ci, m.config)

			// Get character and color
			char := getChar(iter, m.config.MaxIter)
			var color string
			if m.dynamicColor {
				color = getColorWithHueShift(iter, m.config.MaxIter, m.config.ColorScheme, m.hueShift)
			} else {
				color = getColor(iter, m.config.MaxIter, m.config.ColorScheme)
			}

			// Print with or without color
			if color != "" {
				output.WriteString(fmt.Sprintf("%s%c", color, char))
			} else {
				output.WriteByte(char)
			}
		}

		// Reset color at end of line and add newline
		if m.config.ColorScheme != ColorGrayscale {
			output.WriteString("\033[0m")
		}
		if row < m.config.Height-1 {
			output.WriteByte('\n')
		}
	}

	return output.String()
}

func main() {
	// Parse CLI flags
	config := Config{
		MaxIter:     50,
		CenterX:     -0.5,
		CenterY:     0.0,
		Zoom:        1.0,
		ColorScheme: ColorGrayscale,
		FractalType: FractalMandelbrot,
		JuliaCr:     -0.7,
		JuliaCi:     0.27015,
	}

	// Add an interactive mode flag (default true)
	interactive := flag.Bool("interactive", true, "Run in interactive mode")
	flag.BoolVar(&showVersion, "v", false, "print the version")
	flag.BoolVar(&showVersion, "version", false, "print the version")

	legacyMode := flag.Bool("static", false, "Run in legacy static mode (disables interactive)")
	randomMode := flag.Bool("random", false, "Start with a random fractal type")
	flag.BoolVar(randomMode, "r", false, "Start with a random fractal type (shorthand)")
	hyperrandomMode := flag.Bool("hyperrandom", false, "Start with a random interesting view")
	flag.BoolVar(hyperrandomMode, "hyper", false, "Start with a random interesting view (shorthand)")

	flag.IntVar(&config.Width, "w", 0, "Terminal width (0 = auto)")
	flag.IntVar(&config.Width, "width", 0, "Terminal width (0 = auto)")
	flag.IntVar(&config.Height, "h", 0, "Terminal height (0 = auto)")
	flag.IntVar(&config.Height, "height", 0, "Terminal height (0 = auto)")
	flag.IntVar(&config.MaxIter, "i", 50, "Max iterations")
	flag.IntVar(&config.MaxIter, "iterations", 50, "Max iterations")
	flag.Float64Var(&config.CenterX, "x", -0.5, "Center X (real axis)")
	flag.Float64Var(&config.CenterY, "y", 0.0, "Center Y (imaginary axis)")
	flag.Float64Var(&config.Zoom, "z", 1.0, "Zoom level")
	flag.Float64Var(&config.Zoom, "zoom", 1.0, "Zoom level")
	flag.StringVar(&config.ColorScheme, "c", ColorGrayscale, "Color scheme: grayscale, blue, rainbow")
	flag.StringVar(&config.ColorScheme, "color", ColorGrayscale, "Color scheme: grayscale, blue, rainbow")
	flag.StringVar(&config.FractalType, "t", FractalMandelbrot, "Fractal type: mandelbrot, julia, burningship, tricorn, multibrot3, multibrot4, celtic, perpendicular")
	flag.StringVar(&config.FractalType, "type", FractalMandelbrot, "Fractal type: mandelbrot, julia, burningship, tricorn, multibrot3, multibrot4, celtic, perpendicular")
	flag.Float64Var(&config.JuliaCr, "jr", -0.7, "Julia set real parameter")
	flag.Float64Var(&config.JuliaCi, "ji", 0.27015, "Julia set imaginary parameter")

	flag.Parse()

	if showVersion {
		bi, ok := debug.ReadBuildInfo()
		if !ok {
			fmt.Println("fractals dev")
			os.Exit(0)
		}
		version := "dev"
		for _, s := range bi.Settings {
			if s.Key == "vcs.revision" {
				version = s.Value[:7]
				break
			} else if s.Key == "GOOS" {
				// Skip
			}
		}
		// If tagged, use module version
		if bi.Main.Version != "" && bi.Main.Version != "(devel)" {
			version = bi.Main.Version
		}
		fmt.Printf("fractals %s\n", version)
		os.Exit(0)
	}

	// Determine if we should run in interactive mode
	// If --static is specified, or if any non-default flags are set, use static mode
	runInteractive := *interactive && !*legacyMode

	// Check if user provided custom flags (indicates they want static mode)
	if flag.NFlag() > 0 && !*interactive {
		runInteractive = false
	}

	// Validate color scheme
	if config.ColorScheme != ColorGrayscale && config.ColorScheme != ColorBlue && config.ColorScheme != ColorRainbow {
		fmt.Fprintf(os.Stderr, "Invalid color scheme: %s. Using grayscale.\n", config.ColorScheme)
		config.ColorScheme = ColorGrayscale
	}

	// Validate fractal type
	validTypes := map[string]bool{
		FractalMandelbrot:    true,
		FractalJulia:         true,
		FractalBurningShip:   true,
		FractalTricorn:       true,
		FractalMultibrot3:    true,
		FractalMultibrot4:    true,
		FractalCeltic:        true,
		FractalPerpendicular: true,
		FractalMultibrot5:    true,
		FractalManhattan:     true,
		FractalNewton:        true,
	}
	if !validTypes[config.FractalType] {
		fmt.Fprintf(os.Stderr, "Invalid fractal type: %s. Using mandelbrot.\n", config.FractalType)
		config.FractalType = FractalMandelbrot
	}

	// Adjust default center for Julia set if user didn't specify custom center
	if config.FractalType == FractalJulia && config.CenterX == -0.5 && config.CenterY == 0.0 {
		config.CenterX = 0.0
		config.CenterY = 0.0
	}

	// Apply random/hyperrandom if requested
	// Hyperrandom takes precedence over random
	if *hyperrandomMode {
		// Create temporary model to use hyperrandom function
		tempModel := model{config: config}
		tempModel.applyHyperrandom()
		config = tempModel.config
	} else if *randomMode {
		// Random fractal type only
		rng := rand.New(rand.NewSource(time.Now().UnixNano()))
		fractalType := allFractalTypes[rng.Intn(len(allFractalTypes))]
		config.FractalType = fractalType

		// Set default position for random fractal
		if fractalType == FractalJulia {
			config.CenterX = 0.0
			config.CenterY = 0.0
		} else if fractalType == FractalBurningShip {
			config.CenterX = -0.5
			config.CenterY = -0.6
		} else {
			config.CenterX = -0.5
			config.CenterY = 0.0
		}
	}

	if runInteractive {
		// Load bookmarks
		bookmarks, err := loadBookmarks()
		if err != nil {
			// Non-fatal - just start with empty bookmarks
			bookmarks = []Bookmark{}
		}

		// Run interactive TUI mode
		m := model{
			config:            config,
			interactive:       true,
			baseMaxIter:       config.MaxIter, // Store base for adaptive iteration
			bookmarks:         bookmarks,
			autoZoomDirection: 1,              // Default to zoom in
			transitionMode:    TransitionFade, // Default to Fade transition
		}

		p := tea.NewProgram(m, tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Error running interactive mode: %v\n", err)
			os.Exit(1)
		}
	} else {
		// Run legacy static mode
		// Auto-detect terminal size if not specified
		if config.Width == 0 || config.Height == 0 {
			// Try to get terminal size, but don't use the term package
			// since we're already using Bubble Tea's approach
			if config.Width == 0 {
				config.Width = 80
			}
			if config.Height == 0 {
				config.Height = 40
			}
		}

		// Render the fractal once and exit
		render(config)
	}
}

// mandelbrot calculates the number of iterations for a given complex number c
// Returns the iteration count when |z|² > 4 (diverged) or maxIter if in the set
func mandelbrot(cr, ci float64, maxIter int) int {
	zr, zi := 0.0, 0.0

	for i := range maxIter {
		// Calculate z² = (zr + zi*i)²
		zr2 := zr * zr
		zi2 := zi * zi

		// Check if diverged (|z|² > 4)
		if zr2+zi2 > 4.0 {
			return i
		}

		// z = z² + c
		zi = 2.0*zr*zi + ci
		zr = zr2 - zi2 + cr
	}

	return maxIter
}

// julia calculates the Julia set for a given point z and constant c
// z iterates: z = z² + c where c is constant (juliaCr, juliaCi)
func julia(zr, zi, juliaCr, juliaCi float64, maxIter int) int {
	for i := range maxIter {
		zr2 := zr * zr
		zi2 := zi * zi

		if zr2+zi2 > 4.0 {
			return i
		}

		// z = z² + c
		zi = 2.0*zr*zi + juliaCi
		zr = zr2 - zi2 + juliaCr
	}

	return maxIter
}

// burningShip calculates the Burning Ship fractal
// Uses absolute values: z = (|Re(z)| + i|Im(z)|)² + c
func burningShip(cr, ci float64, maxIter int) int {
	zr, zi := 0.0, 0.0

	for i := range maxIter {
		zr2 := zr * zr
		zi2 := zi * zi

		if zr2+zi2 > 4.0 {
			return i
		}

		// Take absolute values before squaring
		if zr < 0 {
			zr = -zr
		}
		if zi < 0 {
			zi = -zi
		}

		// z = z² + c
		zi = 2.0*zr*zi + ci
		zr = zr2 - zi2 + cr
	}

	return maxIter
}

// tricorn (Mandelbar) calculates the Tricorn fractal
// Uses conjugate: z = conj(z)² + c
func tricorn(cr, ci float64, maxIter int) int {
	zr, zi := 0.0, 0.0

	for i := range maxIter {
		zr2 := zr * zr
		zi2 := zi * zi

		if zr2+zi2 > 4.0 {
			return i
		}

		// z = conj(z)² + c
		// conj(z)² = (zr - zi*i)² = zr² - zi² - 2*zr*zi*i
		zi = -2.0*zr*zi + ci // Note the negative sign
		zr = zr2 - zi2 + cr
	}

	return maxIter
}

// multibrot3 calculates the Multibrot set with power 3
// z = z³ + c
func multibrot3(cr, ci float64, maxIter int) int {
	zr, zi := 0.0, 0.0

	for i := range maxIter {
		zr2 := zr * zr
		zi2 := zi * zi

		if zr2+zi2 > 4.0 {
			return i
		}

		// z³ = (zr + zi*i)³ = zr³ + 3*zr²*zi*i - 3*zr*zi² - zi³*i
		//    = (zr³ - 3*zr*zi²) + i(3*zr²*zi - zi³)
		newZr := zr*zr2 - 3.0*zr*zi2 + cr
		newZi := 3.0*zr2*zi - zi*zi2 + ci
		zr = newZr
		zi = newZi
	}

	return maxIter
}

// multibrot4 calculates the Multibrot set with power 4
// z = z⁴ + c
func multibrot4(cr, ci float64, maxIter int) int {
	zr, zi := 0.0, 0.0

	for i := range maxIter {
		zr2 := zr * zr
		zi2 := zi * zi

		if zr2+zi2 > 4.0 {
			return i
		}

		// z⁴ = ((zr + zi*i)²)²
		// First: z² = zr² - zi² + 2*zr*zi*i
		zr_temp := zr2 - zi2
		zi_temp := 2.0 * zr * zi
		// Second: (z²)²
		newZr := zr_temp*zr_temp - zi_temp*zi_temp + cr
		newZi := 2.0*zr_temp*zi_temp + ci
		zr = newZr
		zi = newZi
	}

	return maxIter
}

// celtic calculates the Celtic Mandelbrot fractal
// Uses z = (|Re(z²)| + i*Im(z²)) + c
func celtic(cr, ci float64, maxIter int) int {
	zr, zi := 0.0, 0.0

	for i := range maxIter {
		zr2 := zr * zr
		zi2 := zi * zi

		if zr2+zi2 > 4.0 {
			return i
		}

		// Calculate z²
		newZr := zr2 - zi2
		newZi := 2.0 * zr * zi

		// Take absolute value of real part
		if newZr < 0 {
			newZr = -newZr
		}

		// Add c
		zr = newZr + cr
		zi = newZi + ci
	}

	return maxIter
}

// perpendicular calculates the Perpendicular Mandelbrot fractal
// Uses z = (Re(z²) - |Im(z²)|*i) + c
func perpendicular(cr, ci float64, maxIter int) int {
	zr, zi := 0.0, 0.0

	for i := range maxIter {
		zr2 := zr * zr
		zi2 := zi * zi

		if zr2+zi2 > 4.0 {
			return i
		}

		// Calculate z²
		newZr := zr2 - zi2
		newZi := 2.0 * zr * zi

		// Take absolute value of imaginary part and negate
		if newZi < 0 {
			newZi = -newZi
		}
		newZi = -newZi

		// Add c
		zr = newZr + cr
		zi = newZi + ci
	}

	return maxIter
}

func multibrot5(cr, ci float64, maxIter int) int {
	zr, zi := 0.0, 0.0

	for i := 0; i < maxIter; i++ {
		zr2 := zr * zr
		zi2 := zi * zi

		if zr2+zi2 > 4.0 {
			return i
		}

		z2r := zr2 - zi2
		z2i := 2.0 * zr * zi

		z4r := z2r*z2r - z2i*z2i
		z4i := 2.0 * z2r * z2i

		newZr := z4r*zr - z4i*zi + cr
		newZi := z4r*zi + z4i*zr + ci
		zr = newZr
		zi = newZi
	}

	return maxIter
}

func manhattan(cr, ci float64, maxIter int) int {
	zr, zi := 0.0, 0.0

	for i := 0; i < maxIter; i++ {
		zr2 := zr * zr
		zi2 := zi * zi

		if math.Abs(zr)+math.Abs(zi) > 4.0 {
			return i
		}

		zi = 2.0*zr*zi + ci
		zr = zr2 - zi2 + cr
	}

	return maxIter
}

// newton calculates the Newton fractal for z^3 - 1 = 0
// Shows basins of attraction for the three cube roots of unity
// Uses Newton's method: z_new = z - f(z)/f'(z) = z - (z^3 - 1)/(3z^2)
func newton(zr, zi float64, maxIter int) int {
	const tolerance = 1e-6

	// Three roots of z^3 - 1 = 0 (cube roots of unity)
	// root1 = 1 + 0i
	// root2 = -0.5 + 0.866i
	// root3 = -0.5 - 0.866i

	for i := 0; i < maxIter; i++ {
		// Calculate z^2
		zr2 := zr * zr
		zi2 := zi * zi

		// Calculate z^3 = z * z^2
		z3r := zr*zr2 - zr*zi2 - zr*zi2
		z3i := zi*zr2 + zi*zr2 + zr*zi2

		// f(z) = z^3 - 1
		fr := z3r - 1.0
		fi := z3i

		// f'(z) = 3z^2
		fpr := 3.0 * (zr2 - zi2)
		fpi := 6.0 * zr * zi

		// Calculate f(z)/f'(z) using complex division
		// (a + bi) / (c + di) = ((ac + bd) + (bc - ad)i) / (c^2 + d^2)
		denom := fpr*fpr + fpi*fpi
		if denom < 1e-10 {
			return i
		}

		divr := (fr*fpr + fi*fpi) / denom
		divi := (fi*fpr - fr*fpi) / denom

		// z_new = z - f(z)/f'(z)
		newZr := zr - divr
		newZi := zi - divi

		// Check convergence
		dr := newZr - zr
		di := newZi - zi
		if dr*dr+di*di < tolerance*tolerance {
			// Converged - determine which root and color based on it
			// Check distance to each root
			d1 := (newZr-1.0)*(newZr-1.0) + newZi*newZi
			d2 := (newZr+0.5)*(newZr+0.5) + (newZi-0.866025)*(newZi-0.866025)
			d3 := (newZr+0.5)*(newZr+0.5) + (newZi+0.866025)*(newZi+0.866025)

			// Return different values based on which root we converged to
			// This creates the three-fold symmetric basins
			if d1 < d2 && d1 < d3 {
				return maxIter - i // Root 1 (real)
			} else if d2 < d3 {
				return maxIter - i*2/3 // Root 2 (upper)
			} else {
				return maxIter - i/3 // Root 3 (lower)
			}
		}

		zr = newZr
		zi = newZi
	}

	return 0 // Did not converge
}

// calculateFractal dispatches to the appropriate fractal function based on config
func calculateFractal(cr, ci float64, config Config) int {
	switch config.FractalType {
	case FractalMandelbrot:
		return mandelbrot(cr, ci, config.MaxIter)
	case FractalJulia:
		return julia(cr, ci, config.JuliaCr, config.JuliaCi, config.MaxIter)
	case FractalBurningShip:
		return burningShip(cr, ci, config.MaxIter)
	case FractalTricorn:
		return tricorn(cr, ci, config.MaxIter)
	case FractalMultibrot3:
		return multibrot3(cr, ci, config.MaxIter)
	case FractalMultibrot4:
		return multibrot4(cr, ci, config.MaxIter)
	case FractalCeltic:
		return celtic(cr, ci, config.MaxIter)
	case FractalPerpendicular:
		return perpendicular(cr, ci, config.MaxIter)
	case FractalMultibrot5:
		return multibrot5(cr, ci, config.MaxIter)
	case FractalManhattan:
		return manhattan(cr, ci, config.MaxIter)
	case FractalNewton:
		return newton(cr, ci, config.MaxIter)
	default:
		return mandelbrot(cr, ci, config.MaxIter)
	}
}

// mapToComplex converts terminal coordinates to complex plane coordinates
func mapToComplex(col, row, width, height int, centerX, centerY, zoom float64) (float64, float64) {
	// Default view spans 3.5 units on real axis, 2.5 units on imaginary axis
	// Real axis: -2.5 to 1.0
	// Imaginary axis: -1.25 to 1.25
	realSpan := 3.5 / zoom
	imagSpan := 2.5 / zoom

	// Adjust for aspect ratio (characters are taller than wide, roughly 2:1)
	aspectRatio := 2.0

	// Map column to real axis
	cr := centerX + (float64(col)/float64(width)-0.5)*realSpan

	// Map row to imaginary axis (inverted, top is positive)
	ci := centerY - (float64(row)/float64(height)-0.5)*imagSpan/aspectRatio

	return cr, ci
}

// getChar returns the ASCII character for a given iteration count
func getChar(iter, maxIter int) byte {
	if iter == maxIter {
		return asciiChars[len(asciiChars)-1]
	}

	// Map iteration count to character index
	idx := int(float64(iter) / float64(maxIter) * float64(len(asciiChars)-1))
	return asciiChars[idx]
}

// hueShiftColor applies a hue rotation to an RGB color
// hueShift is in degrees (0-360)
func hueShiftColor(r, g, b int, hueShift float64) (int, int, int) {
	// Convert RGB to HSV
	rf, gf, bf := float64(r)/255.0, float64(g)/255.0, float64(b)/255.0

	max := math.Max(rf, math.Max(gf, bf))
	min := math.Min(rf, math.Min(gf, bf))
	delta := max - min

	var h, s, v float64
	v = max

	if delta < 0.00001 {
		s = 0
		h = 0
	} else {
		s = delta / max

		if rf == max {
			h = 60.0 * (math.Mod((gf-bf)/delta, 6.0))
		} else if gf == max {
			h = 60.0 * (((bf - rf) / delta) + 2.0)
		} else {
			h = 60.0 * (((rf - gf) / delta) + 4.0)
		}

		if h < 0 {
			h += 360.0
		}
	}

	// Apply hue shift
	h = math.Mod(h+hueShift, 360.0)

	// Convert HSV back to RGB
	c := v * s
	x := c * (1.0 - math.Abs(math.Mod(h/60.0, 2.0)-1.0))
	m := v - c

	var rPrime, gPrime, bPrime float64

	if h < 60 {
		rPrime, gPrime, bPrime = c, x, 0
	} else if h < 120 {
		rPrime, gPrime, bPrime = x, c, 0
	} else if h < 180 {
		rPrime, gPrime, bPrime = 0, c, x
	} else if h < 240 {
		rPrime, gPrime, bPrime = 0, x, c
	} else if h < 300 {
		rPrime, gPrime, bPrime = x, 0, c
	} else {
		rPrime, gPrime, bPrime = c, 0, x
	}

	// Convert back to 0-255 range
	rOut := int((rPrime + m) * 255.0)
	gOut := int((gPrime + m) * 255.0)
	bOut := int((bPrime + m) * 255.0)

	return rOut, gOut, bOut
}

// ansiColorFromRGB converts RGB values to the closest ANSI 256-color code
func ansiColorFromRGB(r, g, b int) int {
	// Use the 216-color cube (colors 16-231)
	// Each component can be 0-5 (6 levels)
	rIdx := int(float64(r) / 255.0 * 5.0)
	gIdx := int(float64(g) / 255.0 * 5.0)
	bIdx := int(float64(b) / 255.0 * 5.0)

	return 16 + (36 * rIdx) + (6 * gIdx) + bIdx
}

// getColor returns the ANSI color code for a given iteration count and color scheme
func getColor(iter, maxIter int, scheme string) string {
	return getColorWithHueShift(iter, maxIter, scheme, 0.0)
}

// getColorWithHueShift returns the ANSI color code with optional hue rotation
func getColorWithHueShift(iter, maxIter int, scheme string, hueShift float64) string {
	if scheme == ColorGrayscale {
		return ""
	}

	if iter == maxIter {
		// Points in the set are black
		return "\033[38;5;0m"
	}

	// Normalize iteration count to 0-1
	t := float64(iter) / float64(maxIter)

	// Generate base RGB color based on scheme
	var r, g, b int

	switch scheme {
	case ColorBlue:
		// Blue gradient: dark blue to bright blue
		r = int(t * 100)
		g = int(t * 150)
		b = int(100 + t*155)

	case ColorRainbow:
		// Rainbow gradient: red -> yellow -> green -> cyan -> blue
		if t < 0.2 {
			r, g, b = 255, int(t/0.2*255), 0 // Red to yellow
		} else if t < 0.4 {
			r, g, b = int((1.0-(t-0.2)/0.2)*255), 255, 0 // Yellow to green
		} else if t < 0.6 {
			r, g, b = 0, 255, int((t-0.4)/0.2*255) // Green to cyan
		} else if t < 0.8 {
			r, g, b = 0, int((1.0-(t-0.6)/0.2)*255), 255 // Cyan to blue
		} else {
			r, g, b = int((t-0.8)/0.2*128), 0, 255 // Blue to purple
		}

	case ColorFire:
		// Fire gradient: black -> red -> orange -> yellow -> white
		if t < 0.25 {
			r, g, b = int(t/0.25*255), 0, 0 // Black to red
		} else if t < 0.5 {
			r, g, b = 255, int((t-0.25)/0.25*165), 0 // Red to orange
		} else if t < 0.75 {
			r, g, b = 255, int(165+(t-0.5)/0.25*90), int((t-0.5)/0.25*50) // Orange to yellow
		} else {
			base := int((t - 0.75) / 0.25 * 255)
			r, g, b = 255, 255, base // Yellow to white
		}

	case ColorPurple:
		// Purple/magenta gradient
		r = int(128 + t*127)
		g = int(t * 100)
		b = int(200 + t*55)

	case ColorGreen:
		// Green gradient: dark green to bright green
		r = int(t * 100)
		g = int(100 + t*155)
		b = int(t * 100)

	case ColorGold:
		// Gold/amber gradient: brown -> gold -> yellow
		if t < 0.5 {
			r, g, b = int(139+t*116), int(69+t*146), int(19+t*30) // Brown to gold
		} else {
			r, g, b = int(255), int(215+(t-0.5)*80), int(0+(t-0.5)*200) // Gold to yellow
		}

	case ColorCyan:
		// Cyan/aqua gradient: dark cyan to bright cyan
		r = int(t * 100)
		g = int(139 + t*116)
		b = int(139 + t*116)

	default:
		return ""
	}

	// Apply hue shift if enabled
	if hueShift != 0.0 {
		r, g, b = hueShiftColor(r, g, b, hueShift)
	}

	// Convert RGB to ANSI 256 color
	colorIdx := ansiColorFromRGB(r, g, b)
	return fmt.Sprintf("\033[38;5;%dm", colorIdx)
}

// render generates and displays the selected fractal
func render(config Config) {
	for row := 0; row < config.Height; row++ {
		for col := 0; col < config.Width; col++ {
			// Map terminal coordinates to complex plane
			cr, ci := mapToComplex(col, row, config.Width, config.Height,
				config.CenterX, config.CenterY, config.Zoom)

			// Calculate iteration count
			iter := calculateFractal(cr, ci, config)

			// Get character and color
			char := getChar(iter, config.MaxIter)
			color := getColor(iter, config.MaxIter, config.ColorScheme)

			// Print with or without color
			if color != "" {
				fmt.Printf("%s%c", color, char)
			} else {
				fmt.Printf("%c", char)
			}
		}

		// Reset color at end of line and add newline
		if config.ColorScheme != ColorGrayscale {
			fmt.Print("\033[0m")
		}
		fmt.Println()
	}
}
