package main

import (
	"flag"
	"fmt"
	"math"
	"math/rand"
	"os"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/qjcg/arcadia/x/fractals/colorthemes"
	"github.com/qjcg/arcadia/x/fractals/fractals"
	"github.com/qjcg/arcadia/x/fractals/persistence"
	"github.com/qjcg/arcadia/x/fractals/transitions"
)

const (
	// ASCII characters from sparse to dense
	asciiChars = " .:-=+*#%@"

	// Transition animation modes
	TransitionNone         = persistence.TransitionNone
	TransitionFade         = transitions.TransitionFade
	TransitionZoomOutIn    = transitions.TransitionZoomOutIn
	TransitionRotate       = transitions.TransitionRotate
	TransitionBreakthrough = transitions.TransitionBreakthrough

	// Fractal types
	FractalMandelbrot    = persistence.FractalMandelbrot
	FractalJulia         = persistence.FractalJulia
	FractalBurningShip   = persistence.FractalBurningShip
	FractalTricorn       = persistence.FractalTricorn
	FractalMultibrot3    = persistence.FractalMultibrot3
	FractalMultibrot4    = persistence.FractalMultibrot4
	FractalCeltic        = persistence.FractalCeltic
	FractalPerpendicular = persistence.FractalPerpendicular
	FractalMultibrot5    = persistence.FractalMultibrot5
	FractalManhattan     = persistence.FractalManhattan
	FractalNewton        = persistence.FractalNewton

	// URL modes
	ModeRandom   = persistence.ModeRandom
	ModeStandard = persistence.ModeStandard
)

var (
	showVersion bool

	// All fractal types for random selection
	allFractalTypes = []string{
		FractalMandelbrot, FractalJulia, FractalBurningShip, FractalTricorn,
		FractalMultibrot3, FractalMultibrot4, FractalCeltic, FractalPerpendicular,
		FractalNewton,
	}

	// Known interesting coordinates for random exploration
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

// Use Config from persistence package
type Config = persistence.Config

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
	transitionMode      int     // 0=none, 1=fade, 2=zoom_out_in, 3=rotate, 4=breakthrough
	transitionProgress  float64 // 0.0 to 1.0, progress through transition
	transitionTarget    string  // Target fractal type for transition
	transitionZoomStart float64 // Starting zoom level for transition

	// New transitions package
	fadeTransition         *transitions.Fade
	zoomOutInTransition    *transitions.ZoomOutIn
	rotateTransition       *transitions.Rotate
	breakthroughTransition *transitions.Breakthrough
	// Dynamic color state
	dynamicColor bool    // Enable smooth hue rotation
	hueShift     float64 // Current hue shift in degrees (0-360)
	// Bookmark state
	showBookmarks     bool                   // Show bookmark list
	bookmarks         []persistence.Bookmark // Loaded bookmarks
	bookmarkCursor    int                    // Selected bookmark in list
	savingBookmark    bool                   // Prompting for bookmark name
	bookmarkInput     string                 // User input for bookmark name
	suggestedBookmark string                 // Auto-generated bookmark name suggestion
	// Screenshot state
	screenshotMsg   string // Message to display after screenshot
	screenshotTimer int    // Countdown for hiding screenshot message
	// Random state
	randomMsg   string // Message to display after randomization
	randomTimer int    // Countdown for hiding random message
	// URL launch state
	urlMsg   string // Message to display when copying URL
	urlTimer int    // Countdown for hiding URL message
	// Zoom speed control
	zoomSpeed float64 // Multiplier for zoom speed (default 1.05, adjustable 0.9-1.5)
}

// Init initializes the Bubble Tea model
func (m model) Init() tea.Cmd {
	// If auto-zoom is enabled (e.g., from URL), start the tick loop
	if m.autoZoom {
		return tickCmd()
	}
	return nil
}

// addBookmark adds a new bookmark and saves to file
func (m *model) addBookmark(name string) error {
	url := persistence.ConfigToFractalURL(m.config, m.autoZoom, m.dynamicColor, m.transitionMode)

	bookmark := persistence.Bookmark{
		Name:                name,
		URL:                 url,
		FractalType:         m.config.FractalType,
		CenterX:             m.config.CenterX,
		CenterY:             m.config.CenterY,
		Zoom:                m.config.Zoom,
		MaxIter:             m.config.MaxIter,
		ColorScheme:         m.config.ColorScheme,
		JuliaCr:             m.config.JuliaCr,
		JuliaCi:             m.config.JuliaCi,
		AutopilotEnabled:    m.autoZoom,
		DynamicColorEnabled: m.dynamicColor,
		TransitionMode:      persistence.TransitionModeToString(m.transitionMode),
	}

	m.bookmarks = append(m.bookmarks, bookmark)
	return persistence.SaveBookmarks(m.bookmarks)
}

// applyParamsToModel applies FractalURLParams to a model
func applyParamsToModel(m *model, params persistence.FractalURLParams) {
	m.config.FractalType = params.FractalType
	m.config.CenterX = params.CenterX
	m.config.CenterY = params.CenterY
	m.config.Zoom = params.Zoom
	m.config.MaxIter = params.MaxIter
	m.config.ColorScheme = params.ColorTheme
	m.config.JuliaCr = params.JuliaCr
	m.config.JuliaCi = params.JuliaCi

	m.autoZoom = params.AutopilotEnabled
	if params.AutopilotEnabled {
		m.autoZoomDirection = 1
	}

	m.dynamicColor = params.DynamicColorEnabled
	m.transitionMode = persistence.StringToTransitionMode(params.Transition)

	m.baseMaxIter = params.MaxIter
}

// loadBookmark applies a bookmark to the current config
func (m *model) loadBookmark(index int) {
	if index < 0 || index >= len(m.bookmarks) {
		return
	}

	bm := m.bookmarks[index]

	// If URL present, try to parse from it (prefer URL)
	if bm.URL != "" {
		params, err := persistence.ParseFractalURL(bm.URL)
		if err == nil {
			applyParamsToModel(m, params)
			return
		}
		// If URL parse fails, fall back to individual fields
	}

	// Fallback: load from individual fields (backward compat)
	m.config.FractalType = bm.FractalType
	m.config.CenterX = bm.CenterX
	m.config.CenterY = bm.CenterY
	m.config.Zoom = bm.Zoom
	m.config.MaxIter = bm.MaxIter
	m.config.ColorScheme = bm.ColorScheme
	m.config.JuliaCr = bm.JuliaCr
	m.config.JuliaCi = bm.JuliaCi

	// Restore autopilot/dynamic color/transition if saved
	if bm.AutopilotEnabled {
		m.autoZoom = true
		m.autoZoomDirection = 1
	} else {
		m.autoZoom = false
	}

	if bm.DynamicColorEnabled {
		m.dynamicColor = true
	} else {
		m.dynamicColor = false
	}

	if bm.TransitionMode != "" {
		m.transitionMode = persistence.StringToTransitionMode(bm.TransitionMode)
	}

	// Reset base max iter for adaptive scaling
	m.baseMaxIter = bm.MaxIter
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

	// Initialize the appropriate transition based on mode
	switch m.transitionMode {
	case TransitionFade:
		m.fadeTransition = transitions.NewFadeTransition(allFractalTypes)
		m.fadeTransition.Start(m.config.FractalType)
		m.randomMsg = m.fadeTransition.GetMessage()
	case TransitionZoomOutIn:
		m.zoomOutInTransition = transitions.NewZoomOutInTransition(allFractalTypes)
		m.zoomOutInTransition.Start(m.config.FractalType)
		m.randomMsg = m.zoomOutInTransition.GetMessage()
	case TransitionRotate:
		m.rotateTransition = transitions.NewRotateTransition(allFractalTypes)
		m.rotateTransition.Start(m.config.FractalType)
		m.randomMsg = m.rotateTransition.GetMessage()
	case TransitionBreakthrough:
		m.breakthroughTransition = transitions.NewBreakthroughTransition(allFractalTypes)
		m.breakthroughTransition.Start(m.config.FractalType)
		m.randomMsg = m.breakthroughTransition.GetMessage()
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

// applyRandom generates a completely random interesting view
func (m *model) applyRandom() {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	const maxAttempts = 5

	for attempt := 0; attempt < maxAttempts; attempt++ {
		// Random fractal type
		fractalType := allFractalTypes[rng.Intn(len(allFractalTypes))]
		m.config.FractalType = fractalType

		// Random color scheme
		m.config.ColorScheme = colorthemes.AllColorSchemes[rng.Intn(len(colorthemes.AllColorSchemes))]

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
			m.randomMsg = fmt.Sprintf("Random: %s @ %.1fx", fractalType, zoom)
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
	m.config.ColorScheme = colorthemes.AllColorSchemes[rng.Intn(len(colorthemes.AllColorSchemes))]

	m.randomMsg = fmt.Sprintf("Random: %s @ %.1fx (fallback)", m.config.FractalType, m.config.Zoom)
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
	return persistence.SaveBookmarks(m.bookmarks)
}

// saveScreenshot saves the current fractal view to a text file
func (m *model) saveScreenshot() error {
	// Render the fractal
	fractalOutput := m.renderFractal()

	// Format Julia parameters as strings
	juliaCrStr := fmt.Sprintf("%.6f", m.config.JuliaCr)
	juliaCiStr := fmt.Sprintf("%.6f", m.config.JuliaCi)

	// Build the screenshot content with metadata
	content := persistence.BuildScreenshotContent(
		m.config.FractalType,
		m.config.CenterX,
		m.config.CenterY,
		m.config.Zoom,
		m.config.MaxIter,
		m.config.ColorScheme,
		juliaCrStr,
		juliaCiStr,
		m.config.Width,
		m.config.Height,
		fractalOutput,
	)

	// Save to file
	filename, err := persistence.SaveScreenshot(m.config.FractalType, content)
	if err != nil {
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

			// Zoom based on direction (speed configurable via zoomSpeed)
			// Default: 1.05x per tick, ~60fps = 1.05^20 ≈ 2.65x per second
			if m.autoZoomDirection > 0 {
				// Zoom in
				m.config.Zoom *= m.zoomSpeed
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
				m.config.Zoom /= m.zoomSpeed
				// Minimum zoom limit
				if m.config.Zoom < 0.1 {
					m.config.Zoom = 0.1
					// Trigger transition if enabled
					if m.transitionMode > 0 && m.transitionProgress == 0.0 {
						m.startFractalTransition()
					}
				}
			}

			// Handle transition animations using the new transitions package
			if m.transitionProgress > 0.0 {
				var completed bool
				var message string

				// Use the appropriate transition based on mode
				switch m.transitionMode {
				case TransitionFade:
					if m.fadeTransition == nil {
						m.fadeTransition = transitions.NewFadeTransition(allFractalTypes)
						m.fadeTransition.Start(m.config.FractalType)
					}
					completed, message = m.fadeTransition.Update()
					if completed {
						m.config.FractalType = m.fadeTransition.Target
						m.resetToDefaultState(m.fadeTransition.Target)
						m.transitionProgress = 0.0
						m.randomMsg = message
					}
				case TransitionZoomOutIn:
					if m.zoomOutInTransition == nil {
						m.zoomOutInTransition = transitions.NewZoomOutInTransition(allFractalTypes)
						m.zoomOutInTransition.Start(m.config.FractalType)
					}
					completed, zoomLevel, message := m.zoomOutInTransition.Update()
					m.config.Zoom = zoomLevel
					if completed {
						m.config.FractalType = m.zoomOutInTransition.Target
						m.resetToDefaultState(m.zoomOutInTransition.Target)
						m.transitionProgress = 0.0
						m.randomMsg = message
					}
				case TransitionRotate:
					if m.rotateTransition == nil {
						m.rotateTransition = transitions.NewRotateTransition(allFractalTypes)
						m.rotateTransition.Start(m.config.FractalType)
					}
					completed, centerX, centerY, message := m.rotateTransition.Update()
					m.config.CenterX = centerX
					m.config.CenterY = centerY
					if completed {
						m.config.FractalType = m.rotateTransition.Target
						m.resetToDefaultState(m.rotateTransition.Target)
						m.transitionProgress = 0.0
						m.randomMsg = message
					}
				case TransitionBreakthrough:
					if m.breakthroughTransition == nil {
						m.breakthroughTransition = transitions.NewBreakthroughTransition(allFractalTypes)
						m.breakthroughTransition.Start(m.config.FractalType)
					}
					completed, centerX, centerY, zoomLevel, message := m.breakthroughTransition.Update()
					m.config.CenterX = centerX
					m.config.CenterY = centerY
					m.config.Zoom = zoomLevel
					if completed {
						m.config.FractalType = m.breakthroughTransition.Target
						m.resetToDefaultState(m.breakthroughTransition.Target)
						m.transitionProgress = 0.0
						m.randomMsg = message
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

		// Handle URL message timer
		if m.urlTimer > 0 {
			m.urlTimer--
			if m.urlTimer == 0 {
				m.urlMsg = ""
			}
		}

		// Continue ticking if any timer is active
		if m.screenshotTimer > 0 || m.randomTimer > 0 || m.urlTimer > 0 {
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
			m.suggestedBookmark = persistence.GenerateBookmarkName()
			return m, nil

		case "l":
			// Show bookmark list
			// Reload bookmarks from file
			if bookmarks, err := persistence.LoadBookmarks(); err == nil {
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
			// Random - random everything with interesting view
			m.applyRandom()
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
			for i, cs := range colorthemes.AllColorSchemes {
				if cs == m.config.ColorScheme {
					currentIndex = i
					break
				}
			}
			if currentIndex == -1 {
				// Unknown color scheme, reset to first
				m.config.ColorScheme = colorthemes.AllColorSchemes[0]
			} else {
				// Move to next color scheme
				m.config.ColorScheme = colorthemes.AllColorSchemes[(currentIndex+1)%len(colorthemes.AllColorSchemes)]
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
			m.transitionMode = (m.transitionMode + 1) % 5
			if m.transitionMode == 0 {
				m.randomMsg = "Transition: None"
			} else if m.transitionMode == 1 {
				m.randomMsg = "Transition: Fade"
			} else if m.transitionMode == 2 {
				m.randomMsg = "Transition: Zoom Out/In"
			} else if m.transitionMode == 3 {
				m.randomMsg = "Transition: Rotate"
			} else if m.transitionMode == 4 {
				m.randomMsg = "Transition: Breakthrough"
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

		// Copy current location as URL
		case "U":
			urlStr := persistence.ConfigToFractalURL(m.config, m.autoZoom, m.dynamicColor, m.transitionMode)
			if err := clipboard.WriteAll(urlStr); err != nil {
				m.urlMsg = fmt.Sprintf("Error copying URL: %v", err)
			} else {
				m.urlMsg = "URL copied to clipboard"
			}
			m.urlTimer = 120
			return m, tickCmd()

		// Zoom speed control
		case "}":
			// Increase zoom speed
			m.zoomSpeed = math.Min(m.zoomSpeed+0.05, 1.5)
			m.randomMsg = fmt.Sprintf("Zoom speed: %.2fx", m.zoomSpeed)
			m.randomTimer = 30
			return m, nil
		case "{":
			// Decrease zoom speed
			m.zoomSpeed = math.Max(m.zoomSpeed-0.05, 0.90)
			m.randomMsg = fmt.Sprintf("Zoom speed: %.2fx", m.zoomSpeed)
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
	help.WriteString("  }, {               Increase/decrease auto-pilot zoom speed\n")
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
	help.WriteString("  p       Save screenshot (text file with fractal + metadata)\n")
	help.WriteString("  U       Copy current location as shareable fractal:// URL\n\n")

	help.WriteString(keyStyle.Render("Random Exploration:") + "\n")
	help.WriteString("  R       Random (completely random fractal + interesting view)\n")
	help.WriteString("          Uses AI to find non-uniform, visually interesting regions\n")
	help.WriteString("  T       Cycle transition modes (None/Fade/Zoom Out-In/Rotate/Breakthrough)\n")
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

		statusText := fmt.Sprintf(" AUTO-PILOT %s (%.2fx) ", directionArrow, m.zoomSpeed)
		if m.hasTarget && m.panProgress < 1.0 {
			// Show we're panning toward an interesting region
			statusText = fmt.Sprintf(" AUTO-PILOT \u2192 %s (%.2fx) ", directionArrow, m.zoomSpeed) // → arrow
		}
		autoZoomIndicator = autoZoomStyle.Render(statusText)
	} else {
		// Auto-pilot off: subtle indicator showing ready direction
		inactiveStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")). // Gray
			Bold(false)
		autoZoomIndicator = inactiveStyle.Render(fmt.Sprintf("Auto: %s (%.2fx)", directionArrow, m.zoomSpeed))
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

	// URL message indicator
	urlIndicator := ""
	if m.urlMsg != "" {
		msgStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("white")).
			Background(lipgloss.Color("blue")).
			Bold(true)
		urlIndicator = msgStyle.Render(" " + m.urlMsg + " ")
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
	// URL message (only when present)
	if urlIndicator != "" {
		statusBar += urlIndicator + " "
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
	// Create a 2D grid to store characters for potential breakthrough overlay
	grid := make([][]byte, m.config.Height)
	colorGrid := make([][]string, m.config.Height)
	for i := range grid {
		grid[i] = make([]byte, m.config.Width)
		colorGrid[i] = make([]string, m.config.Width)
	}

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
				color = colorthemes.GetColorWithHueShift(iter, m.config.MaxIter, m.config.ColorScheme, m.hueShift)
			} else {
				color = colorthemes.GetColor(iter, m.config.MaxIter, m.config.ColorScheme)
			}

			grid[row][col] = char
			colorGrid[row][col] = color
		}
	}

	// Apply breakthrough transition overlay if active
	if m.transitionMode == TransitionBreakthrough && m.breakthroughTransition != nil {
		applyBreakthroughOverlay(grid, colorGrid, m.config.Width, m.config.Height, m.breakthroughTransition)
	}

	// Convert grid to string output
	var output strings.Builder
	for row := 0; row < m.config.Height; row++ {
		for col := 0; col < m.config.Width; col++ {
			color := colorGrid[row][col]
			char := grid[row][col]

			// Print with or without color
			if color != "" {
				output.WriteString(fmt.Sprintf("%s%c", color, char))
			} else {
				output.WriteByte(char)
			}
		}

		// Reset color at end of line and add newline
		if m.config.ColorScheme != colorthemes.ColorGrayscale {
			output.WriteString("\033[0m")
		}
		if row < m.config.Height-1 {
			output.WriteByte('\n')
		}
	}

	return output.String()
}

// applyBreakthroughOverlay applies the breakthrough transition visual effect to the grid
func applyBreakthroughOverlay(grid [][]byte, colorGrid [][]string, width, height int, b *transitions.Breakthrough) {
	particles := b.GetParticles()
	crackMap := b.GetCrackMap()

	// Draw particles (falling glass shards)
	for _, p := range particles {
		col := int(p.X * float64(width))
		row := int(p.Y * float64(height))

		if row >= 0 && row < height && col >= 0 && col < width {
			// Use ASCII symbols for particle with opacity effect
			if p.Opacity > 0.7 {
				grid[row][col] = '#'
				colorGrid[row][col] = "\033[90m" // dark gray
			} else if p.Opacity > 0.4 {
				grid[row][col] = '%'
				colorGrid[row][col] = "\033[37m" // light gray
			} else {
				grid[row][col] = '+'
				colorGrid[row][col] = "\033[90m" // dark gray (very faint)
			}
		}
	}

	// Draw cracks
	for coord := range crackMap {
		// Parse coordinate string "x.xx,y.yy"
		parts := strings.Split(coord, ",")
		if len(parts) != 2 {
			continue
		}
		var x, y float64
		fmt.Sscanf(parts[0], "%f", &x)
		fmt.Sscanf(parts[1], "%f", &y)

		col := int(x * float64(width))
		row := int(y * float64(height))

		if row >= 0 && row < height && col >= 0 && col < width {
			grid[row][col] = '/'
			colorGrid[row][col] = "\033[37m" // light gray cracks
		}
	}
}

// looksLikeURL checks if a string looks like a fractal URL (without the fractal:// prefix)
// It matches patterns like "mandelbrot", "random", "mandelbrot/-0.5/0.0/1.0/50/", or "random/?..."
func looksLikeURL(s string) bool {
	// Check for random with or without further components
	if s == "random" || strings.HasPrefix(s, "random/") || strings.HasPrefix(s, "random?") {
		return true
	}

	// Extract the first component (before / or ?)
	endIdx := len(s)
	for i, c := range s {
		if c == '/' || c == '?' {
			endIdx = i
			break
		}
	}

	firstComponent := s[:endIdx]

	// Check if it's a valid fractal type
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
	return validTypes[firstComponent]
}

// expandURLWithDefaults expands a minimal URL (just fractal type or random)
// with default values for missing components
func expandURLWithDefaults(urlPart string) string {
	// If it's just a fractal type name, add defaults
	if urlPart == "random" {
		return "random/"
	}

	// Check if it's just a fractal name (no / or ?)
	hasPath := strings.ContainsAny(urlPart, "/?")
	if !hasPath {
		// Just a fractal name - add defaults
		// Default center depends on fractal type
		centerX := "-0.5"
		centerY := "0.0"
		if urlPart == "julia" {
			centerX = "0.0"
			centerY = "0.0"
		} else if urlPart == "burningship" {
			centerX = "-0.5"
			centerY = "-0.6"
		}
		return fmt.Sprintf("%s/%s/%s/1.0/50/", urlPart, centerX, centerY)
	}

	// Already has path components, return as-is
	return urlPart
}

func main() {
	// Parse CLI flags
	config := Config{
		MaxIter:     50,
		CenterX:     -0.5,
		CenterY:     0.0,
		Zoom:        1.0,
		ColorScheme: colorthemes.ColorGrayscale,
		FractalType: FractalMandelbrot,
		JuliaCr:     -0.7,
		JuliaCi:     0.27015,
	}

	// Add an interactive mode flag (default true)
	interactive := flag.Bool("interactive", true, "Run in interactive mode")
	urlFlag := flag.String("url", "", "Fractal URL (fractal://...)")
	flag.BoolVar(&showVersion, "v", false, "print the version")
	flag.BoolVar(&showVersion, "version", false, "print the version")

	legacyMode := flag.Bool("static", false, "Run in legacy static mode (disables interactive)")
	randomMode := flag.Bool("random", false, "Start with a completely random interesting view")
	flag.BoolVar(randomMode, "r", false, "Start with a completely random interesting view (shorthand)")

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
	flag.StringVar(&config.ColorScheme, "c", colorthemes.ColorGrayscale, "Color scheme: grayscale, blue, rainbow")
	flag.StringVar(&config.ColorScheme, "color", colorthemes.ColorGrayscale, "Color scheme: grayscale, blue, rainbow")
	flag.StringVar(&config.FractalType, "t", FractalMandelbrot, "Fractal type: mandelbrot, julia, burningship, tricorn, multibrot3, multibrot4, celtic, perpendicular")
	flag.StringVar(&config.FractalType, "type", FractalMandelbrot, "Fractal type: mandelbrot, julia, burningship, tricorn, multibrot3, multibrot4, celtic, perpendicular")
	flag.Float64Var(&config.JuliaCr, "jr", -0.7, "Julia set real parameter")
	flag.Float64Var(&config.JuliaCi, "ji", 0.27015, "Julia set imaginary parameter")

	flag.Parse()

	// Check for URL argument (positional) or --url flag
	urlAutopilot := false
	urlDynamicColor := false
	urlTransitionMode := TransitionNone

	args := flag.Args()
	if len(args) > 0 {
		arg := args[0]

		// Handle URLs with or without fractal:// prefix
		if strings.HasPrefix(arg, "fractal://") {
			// URL provided as is
			*urlFlag = arg
		} else if looksLikeURL(arg) {
			// URL without prefix - expand with defaults if needed, then add prefix
			expanded := expandURLWithDefaults(arg)
			*urlFlag = "fractal://" + expanded
		}
	}

	if *urlFlag != "" {
		// Parse URL and extract config
		params, err := persistence.ParseFractalURL(*urlFlag)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing URL: %v\n", err)
			os.Exit(1)
		}

		// Apply URL parameters to config
		config.FractalType = params.FractalType
		config.CenterX = params.CenterX
		config.CenterY = params.CenterY
		config.Zoom = params.Zoom
		config.MaxIter = params.MaxIter
		config.ColorScheme = params.ColorTheme
		config.JuliaCr = params.JuliaCr
		config.JuliaCi = params.JuliaCi

		// Save URL state for interactive mode
		urlAutopilot = params.AutopilotEnabled
		urlDynamicColor = params.DynamicColorEnabled
		urlTransitionMode = persistence.StringToTransitionMode(params.Transition)

		// If random mode, we'll apply it after TUI initialization
		if params.Mode == ModeRandom {
			*randomMode = true
		}
	}

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
	if !persistence.IsValidColorTheme(config.ColorScheme) {
		fmt.Fprintf(os.Stderr, "Invalid color scheme: %s. Using grayscale.\n", config.ColorScheme)
		config.ColorScheme = colorthemes.ColorGrayscale
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

	// Apply random if requested
	if *randomMode {
		// Create temporary model to use random function
		tempModel := model{config: config}
		tempModel.applyRandom()
		config = tempModel.config
	}

	if runInteractive {
		// Load bookmarks
		bookmarks, err := persistence.LoadBookmarks()
		if err != nil {
			// Non-fatal - just start with empty bookmarks
			bookmarks = []persistence.Bookmark{}
		}

		// Run interactive TUI mode
		m := model{
			config:            config,
			interactive:       true,
			baseMaxIter:       config.MaxIter, // Store base for adaptive iteration
			bookmarks:         bookmarks,
			autoZoomDirection: 1,              // Default to zoom in
			transitionMode:    TransitionFade, // Default to Fade transition
			fadeTransition:    transitions.NewFadeTransition(allFractalTypes),
			zoomSpeed:         1.05, // Default zoom speed
		}

		// Apply URL-based state if provided
		if *urlFlag != "" {
			m.autoZoom = urlAutopilot
			m.dynamicColor = urlDynamicColor
			m.transitionMode = urlTransitionMode
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

// calculateFractal dispatches to the appropriate fractal function based on config
func calculateFractal(cr, ci float64, config Config) int {
	switch config.FractalType {
	case FractalMandelbrot:
		return fractals.Mandelbrot(cr, ci, config.MaxIter)
	case FractalJulia:
		return fractals.Julia(cr, ci, config.JuliaCr, config.JuliaCi, config.MaxIter)
	case FractalBurningShip:
		return fractals.BurningShip(cr, ci, config.MaxIter)
	case FractalTricorn:
		return fractals.Tricorn(cr, ci, config.MaxIter)
	case FractalMultibrot3:
		return fractals.Multibrot3(cr, ci, config.MaxIter)
	case FractalMultibrot4:
		return fractals.Multibrot4(cr, ci, config.MaxIter)
	case FractalCeltic:
		return fractals.Celtic(cr, ci, config.MaxIter)
	case FractalPerpendicular:
		return fractals.Perpendicular(cr, ci, config.MaxIter)
	case FractalMultibrot5:
		return fractals.Multibrot5(cr, ci, config.MaxIter)
	case FractalManhattan:
		return fractals.Manhattan(cr, ci, config.MaxIter)
	case FractalNewton:
		return fractals.Newton(cr, ci, config.MaxIter)
	default:
		return fractals.Mandelbrot(cr, ci, config.MaxIter)
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
			color := colorthemes.GetColor(iter, config.MaxIter, config.ColorScheme)

			// Print with or without color
			if color != "" {
				fmt.Printf("%s%c", color, char)
			} else {
				fmt.Printf("%c", char)
			}
		}

		// Reset color at end of line and add newline
		if config.ColorScheme != colorthemes.ColorGrayscale {
			fmt.Print("\033[0m")
		}
		fmt.Println()
	}
}
