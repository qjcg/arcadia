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

	"github.com/qjcg/arcadia/x/fractalis/internal/core/fractal"
	"github.com/qjcg/arcadia/x/fractalis/internal/core/persistence"
	"github.com/qjcg/arcadia/x/fractalis/internal/core/search"
	"github.com/qjcg/arcadia/x/fractalis/internal/ebiten"
	"github.com/qjcg/arcadia/x/fractalis/internal/tui/animation"
	"github.com/qjcg/arcadia/x/fractalis/internal/tui/color"
	renderlib "github.com/qjcg/arcadia/x/fractalis/internal/tui/render"
	"github.com/qjcg/arcadia/x/fractalis/internal/tui/transition"
	"github.com/qjcg/arcadia/x/fractalis/internal/web"
)

const (
	// ASCII characters from sparse to dense
	asciiChars = " .:-=+*#%@"

	// Transition animation modes
	TransitionNone         = persistence.TransitionNone
	TransitionFade         = transition.TransitionFade
	TransitionZoomOutIn    = transition.TransitionZoomOutIn
	TransitionRotate       = transition.TransitionRotate
	TransitionBreakthrough = transition.TransitionBreakthrough

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
	FractalMandelbulb    = persistence.FractalMandelbulb
	FractalMandelbox     = persistence.FractalMandelbox

	// URL modes
	ModeRandom   = persistence.ModeRandom
	ModeStandard = persistence.ModeStandard
	ModeVantage  = persistence.ModeVantage
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
	config     Config
	showHelp   bool
	ready      bool
	termWidth  int
	termHeight int
	lastRender string
	// Interest calculator for finding interesting points
	calculator *search.InterestCalculator
	// Animation and UI state
	animationState animation.AnimationState
	// Rendering engine with caching and parallel computation
	renderer *renderlib.Renderer
	// Bookmark state
	showBookmarks     bool                   // Show bookmark list
	bookmarks         []persistence.Bookmark // Loaded bookmarks
	bookmarkCursor    int                    // Selected bookmark in list
	savingBookmark    bool                   // Prompting for bookmark name
	bookmarkInput     string                 // User input for bookmark name
	suggestedBookmark string                 // Auto-generated bookmark name suggestion
	bookmarksLoadedOk bool                   // Whether bookmarks were successfully loaded from file
}

// Init initializes the Bubble Tea model
func (m model) Init() tea.Cmd {
	// If auto-zoom is enabled (e.g., from URL), start the tick loop
	if m.animationState.AutoPilot.Enabled {
		return tickCmd()
	}
	// If vantage mode is enabled, start the tick loop
	if m.animationState.Vantage.Enabled {
		return tickCmd()
	}
	return nil
}

// addBookmark adds a new bookmark and saves to file
func (m *model) addBookmark(name string) error {
	// Reload bookmarks from disk first to avoid overwriting concurrent changes
	if loaded, err := persistence.LoadBookmarks(); err == nil {
		m.bookmarks = loaded
	}

	url := persistence.ConfigToFractalURL(m.config, m.animationState.AutoPilot.Enabled, m.animationState.Color.DynamicColor, m.animationState.Transition.Mode)

	bookmark := persistence.Bookmark{
		Name: name,
		URL:  url,
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

	m.animationState.AutoPilot.Enabled = params.AutopilotEnabled
	if params.AutopilotEnabled {
		m.animationState.AutoPilot.ZoomDirection = 1
	}

	m.animationState.Color.DynamicColor = params.DynamicColorEnabled
	m.animationState.Transition.Mode = persistence.StringToTransitionMode(params.Transition)

	m.animationState.AutoPilot.BaseMaxIter = params.MaxIter
}

// loadBookmark applies a bookmark to the current config
func (m *model) loadBookmark(index int) {
	if index < 0 || index >= len(m.bookmarks) {
		return
	}

	bm := m.bookmarks[index]

	// Parse bookmark from URL
	if bm.URL == "" {
		return
	}

	params, err := persistence.ParseFractalURL(bm.URL)
	if err != nil {
		return
	}

	applyParamsToModel(m, params)
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

	m.animationState.Transition.Target = allFractalTypes[targetIndex]
	m.animationState.Transition.Progress = 0.01 // Start slightly above 0 so animation triggers
	m.animationState.Transition.ZoomStart = m.config.Zoom

	// Initialize the appropriate transition based on mode
	switch m.animationState.Transition.Mode {
	case TransitionFade:
		m.animationState.Transition.FadeTransition = transition.NewFadeTransition(allFractalTypes)
		m.animationState.Transition.FadeTransition.Start(m.config.FractalType)
		m.animationState.Messages.RandomMsg = m.animationState.Transition.FadeTransition.GetMessage()
	case TransitionZoomOutIn:
		m.animationState.Transition.ZoomOutInTransition = transition.NewZoomOutInTransition(allFractalTypes)
		m.animationState.Transition.ZoomOutInTransition.Start(m.config.FractalType)
		m.animationState.Messages.RandomMsg = m.animationState.Transition.ZoomOutInTransition.GetMessage()
	case TransitionRotate:
		m.animationState.Transition.RotateTransition = transition.NewRotateTransition(allFractalTypes)
		m.animationState.Transition.RotateTransition.Start(m.config.FractalType)
		m.animationState.Messages.RandomMsg = m.animationState.Transition.RotateTransition.GetMessage()
	case TransitionBreakthrough:
		m.animationState.Transition.BreakthroughTransition = transition.NewBreakthroughTransition(allFractalTypes)
		m.animationState.Transition.BreakthroughTransition.Start(m.config.FractalType)
		m.animationState.Messages.RandomMsg = m.animationState.Transition.BreakthroughTransition.GetMessage()
	}
	m.animationState.Messages.RandomTimer = 60
}

// resetToDefaultState resets the fractal configuration to default values for a given fractal type
func (m *model) resetToDefaultState(fractalType string) {
	m.config.Zoom = 1.0
	// Reset iteration count to base value
	if m.animationState.AutoPilot.BaseMaxIter > 0 {
		m.config.MaxIter = m.animationState.AutoPilot.BaseMaxIter
	} else {
		m.config.MaxIter = animation.DefaultBaseIterations // Default base iterations
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
		if m.animationState.Color.ColorMode {
			// Pick a random color scheme (excluding grayscale)
			schemes := color.AllColorSchemes[1:]
			m.config.ColorScheme = schemes[rng.Intn(len(schemes))]
		} else {
			m.config.ColorScheme = color.ColorGrayscale
		}

		// Random zoom (log-uniform between 1.0 and 1000.0, weighted toward mid-range)
		// Using exponential distribution: zoom = 10^(uniform(0, 3))
		logZoom := rng.Float64() * 3.0 // 0 to 3
		zoom := math.Pow(10.0, logZoom)
		m.config.Zoom = zoom

		// Random iterations appropriate for zoom level
		baseIter := animation.DefaultBaseIterations + rng.Intn(100)
		zoomBonus := int(math.Log10(zoom) * animation.IterationScaleFactor)
		m.config.MaxIter = baseIter + zoomBonus
		if m.config.MaxIter > 300 {
			m.config.MaxIter = 300
		}
		m.animationState.AutoPilot.BaseMaxIter = baseIter

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
			m.animationState.Messages.RandomMsg = fmt.Sprintf("Random: %s @ %.1fx", fractalType, zoom)
			m.animationState.Messages.RandomTimer = 60
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
	if m.animationState.Color.ColorMode {
		schemes := color.AllColorSchemes[1:]
		m.config.ColorScheme = schemes[rng.Intn(len(schemes))]
	} else {
		m.config.ColorScheme = color.ColorGrayscale
	}

	m.animationState.Messages.RandomMsg = fmt.Sprintf("Random: %s @ %.1fx (fallback)", m.config.FractalType, m.config.Zoom)
	m.animationState.Messages.RandomTimer = 60
}

// deleteBookmark removes a bookmark at the specified index and saves the updated list
func (m *model) deleteBookmark(index int) error {
	// Reload bookmarks from disk first to avoid overwriting concurrent changes
	if loaded, err := persistence.LoadBookmarks(); err == nil {
		m.bookmarks = loaded
	}

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
	m.animationState.Messages.ScreenshotMsg = fmt.Sprintf("Screenshot saved: %s", filename)
	m.animationState.Messages.ScreenshotTimer = 60 // Show message for ~3 seconds (60 ticks at 50ms each)

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
	baseIter := m.animationState.AutoPilot.BaseMaxIter
	if baseIter == 0 {
		baseIter = animation.DefaultBaseIterations // Default if not set
	}

	// Increase iterations logarithmically with zoom
	// Formula: baseIter + log10(zoom) * scaleFactor
	// At zoom=1: baseIter
	// At zoom=100: baseIter + 40
	// At zoom=10000: baseIter + 80
	// At zoom=1000000: baseIter + 120
	if m.config.Zoom > 1.0 {
		logZoom := math.Log10(m.config.Zoom)
		adaptiveIter := baseIter + int(logZoom*animation.IterationScaleFactor)

		// Cap at reasonable maximum to avoid performance issues
		if adaptiveIter > animation.MaxIterationCap {
			adaptiveIter = animation.MaxIterationCap
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
// Delegates to the interest calculator
func (m model) isViewUniform() bool {
	return m.calculator.IsViewUniform(m.config)
}

// findInterestingPoint searches for a point with high variation/detail
// Delegates to the interest calculator
func (m model) findInterestingPoint() (float64, float64) {
	return m.calculator.FindInterestingPoint(search.DefaultSearchPasses(), m.config)
}

// findDescentTarget searches for a nearby interesting point for auto-pilot descent mode.
// This uses a focused search radius to encourage zooming into local details rather
// than jumping to distant features, creating a smoother "descending" experience.
func (m model) findDescentTarget() (float64, float64) {
	return m.calculator.FindInterestingPoint(search.DescentSearchPasses(), m.config)
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
		// Handle vantage mode - continuous random scenes with slow pan
		if m.animationState.Vantage.Enabled {
			// Initialize first scene if not yet initialized
			if !m.animationState.Vantage.Initialized {
				m.applyRandom()
				m.animationState.Vantage.SceneTimer = m.animationState.Vantage.SceneDuration
				m.animationState.Vantage.Initialized = true
				// Enable slow panning by setting up a fake autopilot target
				newX, newY := m.findInterestingPoint()
				m.animationState.AutoPilot.TargetX = newX
				m.animationState.AutoPilot.TargetY = newY
				m.animationState.AutoPilot.HasTarget = true
				m.animationState.AutoPilot.PanProgress = 0.0
				// Ensure dynamic color and no auto-zoom
				m.animationState.Color.DynamicColor = true
				m.animationState.AutoPilot.Enabled = false
				return m, tickCmd()
			}

			// Decrement scene timer
			m.animationState.Vantage.SceneTimer--

			// Apply slow pan if we have a target
			if m.animationState.AutoPilot.HasTarget && m.animationState.AutoPilot.PanProgress < 1.0 {
				// Move toward target gradually (slower than autopilot)
				deltaX := m.animationState.AutoPilot.TargetX - m.config.CenterX
				deltaY := m.animationState.AutoPilot.TargetY - m.config.CenterY

				m.config.CenterX += deltaX * animation.VantagePanRate
				m.config.CenterY += deltaY * animation.VantagePanRate

				// Update progress based on distance remaining
				effectiveDelta := m.getEffectiveSearchDelta()
				precisionThreshold := effectiveDelta * effectiveDelta * 0.01

				remainingDist := deltaX*deltaX + deltaY*deltaY
				if remainingDist < precisionThreshold {
					m.animationState.AutoPilot.PanProgress = 1.0
				} else {
					m.animationState.AutoPilot.PanProgress += 0.01 // Also slower progress update
				}
			}

			// Update dynamic color hue shift
			if m.animationState.Color.DynamicColor {
				m.animationState.Color.HueShift += 0.5
				if m.animationState.Color.HueShift >= 360.0 {
					m.animationState.Color.HueShift -= 360.0
				}
			}

			// Check if it's time to switch to a new scene
			if m.animationState.Vantage.SceneTimer <= 0 {
				m.applyRandom()
				m.animationState.Vantage.SceneTimer = m.animationState.Vantage.SceneDuration
				// Find new interesting point to pan to
				newX, newY := m.findInterestingPoint()
				m.animationState.AutoPilot.TargetX = newX
				m.animationState.AutoPilot.TargetY = newY
				m.animationState.AutoPilot.HasTarget = true
				m.animationState.AutoPilot.PanProgress = 0.0
			}

			// Continue the animation
			return m, tickCmd()
		}

		// Handle auto-zoom animation tick with intelligent panning
		if m.animationState.AutoPilot.Enabled {
			// Update iteration count adaptively based on zoom level
			m.config.MaxIter = m.calculateAdaptiveMaxIter()

			// Check if we need a new target
			// Only switch targets when:
			// 1. No target exists yet, OR
			// 2. We've fully reached the current target, OR
			// 3. The view has become completely uniform (boring) AND we've made significant progress
			// This prevents jerky mid-pan jumps and encourages descent into local details
			shouldFindNewTarget := false
			if !m.animationState.AutoPilot.HasTarget {
				shouldFindNewTarget = true
			} else if m.animationState.AutoPilot.PanProgress >= 1.0 {
				shouldFindNewTarget = true
			} else if m.animationState.AutoPilot.PanProgress > 0.8 && m.isViewUniform() {
				// Only jump if we're mostly done AND the view is boring
				shouldFindNewTarget = true
			}

			if shouldFindNewTarget {
				// Find a new interesting point using descent-focused search
				// This searches close to center, encouraging zooming into local details
				newX, newY := m.findDescentTarget()
				m.animationState.AutoPilot.TargetX = newX
				m.animationState.AutoPilot.TargetY = newY
				m.animationState.AutoPilot.HasTarget = true
				m.animationState.AutoPilot.PanProgress = 0.0
			}

			// Smooth pan toward target
			if m.animationState.AutoPilot.HasTarget && m.animationState.AutoPilot.PanProgress < 1.0 {
				// Move toward target gradually
				// Just move a fraction of the remaining distance each frame
				deltaX := m.animationState.AutoPilot.TargetX - m.config.CenterX
				deltaY := m.animationState.AutoPilot.TargetY - m.config.CenterY

				// Move a fraction of the remaining distance per tick (smooth exponential approach)
				m.config.CenterX += deltaX * animation.AutoPilotPanRate
				m.config.CenterY += deltaY * animation.AutoPilotPanRate

				// Update progress based on distance remaining
				// Use precision-aware threshold
				effectiveDelta := m.getEffectiveSearchDelta()
				precisionThreshold := effectiveDelta * effectiveDelta * 0.01

				remainingDist := deltaX*deltaX + deltaY*deltaY
				if remainingDist < precisionThreshold {
					// Close enough, mark as reached
					m.animationState.AutoPilot.PanProgress = 1.0
				} else {
					// Progress increases faster at first, slower as we approach
					m.animationState.AutoPilot.PanProgress += 0.02
				}
			}

			// Zoom based on direction (speed configurable via zoomSpeed)
			// Default: 1.05x per tick, ~60fps = 1.05^20 ≈ 2.65x per second
			if m.animationState.AutoPilot.ZoomDirection > 0 {
				// Zoom in
				m.config.Zoom *= m.animationState.AutoPilot.ZoomSpeed
				// Check if we hit maximum zoom and should trigger transition
				if m.config.Zoom > 1e15 {
					m.config.Zoom = 1e15
					// Trigger transition if enabled
					if m.animationState.Transition.Mode > 0 && m.animationState.Transition.Progress == 0.0 {
						m.startFractalTransition()
					}
				}
			} else {
				// Zoom out
				m.config.Zoom /= m.animationState.AutoPilot.ZoomSpeed
				// Minimum zoom limit
				if m.config.Zoom < 0.1 {
					m.config.Zoom = 0.1
					// Trigger transition if enabled
					if m.animationState.Transition.Mode > 0 && m.animationState.Transition.Progress == 0.0 {
						m.startFractalTransition()
					}
				}
			}

			// Handle transition animations using the new transitions package
			if m.animationState.Transition.Progress > 0.0 {
				var completed bool
				var message string

				// Use the appropriate transition based on mode
				switch m.animationState.Transition.Mode {
				case TransitionFade:
					if m.animationState.Transition.FadeTransition == nil {
						m.animationState.Transition.FadeTransition = transition.NewFadeTransition(allFractalTypes)
						m.animationState.Transition.FadeTransition.Start(m.config.FractalType)
					}
					completed, message = m.animationState.Transition.FadeTransition.Update()
					if completed {
						m.config.FractalType = m.animationState.Transition.FadeTransition.Target
						m.resetToDefaultState(m.animationState.Transition.FadeTransition.Target)
						m.animationState.Transition.Progress = 0.0
						m.animationState.Messages.RandomMsg = message
					}
				case TransitionZoomOutIn:
					if m.animationState.Transition.ZoomOutInTransition == nil {
						m.animationState.Transition.ZoomOutInTransition = transition.NewZoomOutInTransition(allFractalTypes)
						m.animationState.Transition.ZoomOutInTransition.Start(m.config.FractalType)
					}
					completed, zoomLevel, message := m.animationState.Transition.ZoomOutInTransition.Update()
					m.config.Zoom = zoomLevel
					if completed {
						m.config.FractalType = m.animationState.Transition.ZoomOutInTransition.Target
						m.resetToDefaultState(m.animationState.Transition.ZoomOutInTransition.Target)
						m.animationState.Transition.Progress = 0.0
						m.animationState.Messages.RandomMsg = message
					}
				case TransitionRotate:
					if m.animationState.Transition.RotateTransition == nil {
						m.animationState.Transition.RotateTransition = transition.NewRotateTransition(allFractalTypes)
						m.animationState.Transition.RotateTransition.Start(m.config.FractalType)
					}
					completed, centerX, centerY, message := m.animationState.Transition.RotateTransition.Update()
					m.config.CenterX = centerX
					m.config.CenterY = centerY
					if completed {
						m.config.FractalType = m.animationState.Transition.RotateTransition.Target
						m.resetToDefaultState(m.animationState.Transition.RotateTransition.Target)
						m.animationState.Transition.Progress = 0.0
						m.animationState.Messages.RandomMsg = message
					}
				case TransitionBreakthrough:
					if m.animationState.Transition.BreakthroughTransition == nil {
						m.animationState.Transition.BreakthroughTransition = transition.NewBreakthroughTransition(allFractalTypes)
						m.animationState.Transition.BreakthroughTransition.Start(m.config.FractalType)
					}
					completed, centerX, centerY, zoomLevel, message := m.animationState.Transition.BreakthroughTransition.Update()
					m.config.CenterX = centerX
					m.config.CenterY = centerY
					m.config.Zoom = zoomLevel
					if completed {
						m.config.FractalType = m.animationState.Transition.BreakthroughTransition.Target
						m.resetToDefaultState(m.animationState.Transition.BreakthroughTransition.Target)
						m.animationState.Transition.Progress = 0.0
						m.animationState.Messages.RandomMsg = message
					}
				}
			}

			// Update dynamic color hue shift
			if m.animationState.Color.DynamicColor {
				m.animationState.Color.HueShift += 0.5 // Shift 0.5 degrees per tick (30 degrees/sec at 60fps, full cycle every 12 seconds)
				if m.animationState.Color.HueShift >= 360.0 {
					m.animationState.Color.HueShift -= 360.0
				}
			}

			// Continue the animation by sending another tick
			return m, tickCmd()
		}

		// Update dynamic color even when not auto-zooming (for manual navigation)
		if m.animationState.Color.DynamicColor {
			m.animationState.Color.HueShift += 0.5
			if m.animationState.Color.HueShift >= 360.0 {
				m.animationState.Color.HueShift -= 360.0
			}
			// Continue ticking to keep colors changing
			return m, tickCmd()
		}

		// Handle screenshot message timer even when not auto-zooming
		if m.animationState.Messages.ScreenshotTimer > 0 {
			m.animationState.Messages.ScreenshotTimer--
			if m.animationState.Messages.ScreenshotTimer == 0 {
				m.animationState.Messages.ScreenshotMsg = ""
			}
		}

		// Handle random message timer
		if m.animationState.Messages.RandomTimer > 0 {
			m.animationState.Messages.RandomTimer--
			if m.animationState.Messages.RandomTimer == 0 {
				m.animationState.Messages.RandomMsg = ""
			}
		}

		// Handle URL message timer
		if m.animationState.Messages.URLTimer > 0 {
			m.animationState.Messages.URLTimer--
			if m.animationState.Messages.URLTimer == 0 {
				m.animationState.Messages.URLMsg = ""
			}
		}

		// Continue ticking if any timer is active
		if m.animationState.Messages.ScreenshotTimer > 0 || m.animationState.Messages.RandomTimer > 0 || m.animationState.Messages.URLTimer > 0 {
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
				m.animationState.Messages.ScreenshotMsg = fmt.Sprintf("Error saving screenshot: %v", err)
				m.animationState.Messages.ScreenshotTimer = 60
			}
			// Start tick to show and hide the message
			return m, tickCmd()

		case "R":
			// Random - random everything with interesting view
			m.applyRandom()
			// Only start tick loop if not already in an animation mode
			if !m.animationState.AutoPilot.Enabled && !m.animationState.Vantage.Enabled {
				return m, tickCmd()
			}
			return m, nil

		case "z":
			// Toggle auto-pilot mode (mutually exclusive with vantage mode)
			if m.animationState.Vantage.Enabled {
				// Exit vantage mode when entering autopilot
				m.animationState.Vantage.Enabled = false
				m.animationState.Vantage.Initialized = false
			}
			m.animationState.AutoPilot.Enabled = !m.animationState.AutoPilot.Enabled
			if m.animationState.AutoPilot.Enabled {
				// Store base iteration count for adaptive scaling
				if m.animationState.AutoPilot.BaseMaxIter == 0 {
					m.animationState.AutoPilot.BaseMaxIter = m.config.MaxIter
				}
				// Initialize zoom direction if not set (default to zoom in)
				if m.animationState.AutoPilot.ZoomDirection == 0 {
					m.animationState.AutoPilot.ZoomDirection = 1
				}
				// Start the animation
				return m, tickCmd()
			}
			return m, nil

		case "r":
			// Toggle/reverse auto-pilot zoom direction
			if m.animationState.AutoPilot.ZoomDirection >= 0 {
				m.animationState.AutoPilot.ZoomDirection = -1
			} else {
				m.animationState.AutoPilot.ZoomDirection = 1
			}
			return m, nil

		case "V":
			// Toggle vantage mode (mutually exclusive with autopilot)
			if m.animationState.AutoPilot.Enabled {
				// Exit autopilot when entering vantage mode
				m.animationState.AutoPilot.Enabled = false
			}
			m.animationState.Vantage.Enabled = !m.animationState.Vantage.Enabled
			if m.animationState.Vantage.Enabled {
				// Entering vantage mode
				// Store base iteration count for adaptive scaling
				if m.animationState.AutoPilot.BaseMaxIter == 0 {
					m.animationState.AutoPilot.BaseMaxIter = m.config.MaxIter
				}
				// Initialize if needed
				if !m.animationState.Vantage.Initialized {
					m.animationState.Vantage.SceneTimer = 0
					// Set default scene duration if not already set
					if m.animationState.Vantage.SceneDuration == 0 {
						m.animationState.Vantage.SceneDuration = 100 // Default 5 seconds (100 ticks at 50ms)
					}
				}
				// Force dynamic color on, autopilot off
				m.animationState.Color.DynamicColor = true
				m.animationState.AutoPilot.Enabled = false
				m.animationState.Messages.RandomMsg = fmt.Sprintf("Vantage Mode: ON (%.1fs/scene)", float64(m.animationState.Vantage.SceneDuration)/20.0)
				m.animationState.Messages.RandomTimer = 30
				// Start the animation
				return m, tickCmd()
			} else {
				// Exiting vantage mode
				m.animationState.Vantage.Initialized = false
				m.animationState.Messages.RandomMsg = "Vantage Mode: OFF"
				m.animationState.Messages.RandomTimer = 30
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
			if m.animationState.AutoPilot.Enabled {
				m.animationState.AutoPilot.ZoomDirection = 1
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
			if m.animationState.AutoPilot.Enabled {
				m.animationState.AutoPilot.ZoomDirection = -1
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
			if m.animationState.AutoPilot.BaseMaxIter > 0 {
				m.config.MaxIter = m.animationState.AutoPilot.BaseMaxIter
			} else {
				m.config.MaxIter = animation.DefaultBaseIterations // Fallback to default
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
			if !m.animationState.Color.ColorMode {
				m.animationState.Messages.RandomMsg = "Color Mode is OFF. Press 'X' to enable colors."
				m.animationState.Messages.RandomTimer = 30
				return m, nil
			}

			currentIndex := -1
			for i, cs := range color.AllColorSchemes {
				if cs == m.config.ColorScheme {
					currentIndex = i
					break
				}
			}

			// If color mode is ON, we want to skip grayscale (index 0)
			if currentIndex == -1 {
				m.config.ColorScheme = color.AllColorSchemes[1] // Start at first color
			} else {
				// Move to next color scheme, skipping index 0
				nextIndex := (currentIndex + 1) % len(color.AllColorSchemes)
				if nextIndex == 0 {
					nextIndex = 1
				}
				m.config.ColorScheme = color.AllColorSchemes[nextIndex]
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
			m.animationState.Transition.Mode = (m.animationState.Transition.Mode + 1) % 5
			if m.animationState.Transition.Mode == 0 {
				m.animationState.Messages.RandomMsg = "Transition: None"
			} else if m.animationState.Transition.Mode == 1 {
				m.animationState.Messages.RandomMsg = "Transition: Fade"
			} else if m.animationState.Transition.Mode == 2 {
				m.animationState.Messages.RandomMsg = "Transition: Zoom Out/In"
			} else if m.animationState.Transition.Mode == 3 {
				m.animationState.Messages.RandomMsg = "Transition: Rotate"
			} else if m.animationState.Transition.Mode == 4 {
				m.animationState.Messages.RandomMsg = "Transition: Breakthrough"
			}
			m.animationState.Messages.RandomTimer = 30
			return m, nil

		// Dynamic color mode toggle
		case "C":
			m.animationState.Color.DynamicColor = !m.animationState.Color.DynamicColor
			if m.animationState.Color.DynamicColor {
				m.animationState.Messages.RandomMsg = "Dynamic Color: ON"
				// Start the tick animation to keep colors changing
				return m, tickCmd()
			} else {
				m.animationState.Messages.RandomMsg = "Dynamic Color: OFF"
				m.animationState.Color.HueShift = 0.0 // Reset hue shift
			}
			m.animationState.Messages.RandomTimer = 30
			return m, nil

		// Color mode toggle (Black & White vs Color)
		case "X":
			m.animationState.Color.ColorMode = !m.animationState.Color.ColorMode
			if m.animationState.Color.ColorMode {
				m.animationState.Messages.RandomMsg = "Color Mode: ON"
				// If we were in grayscale, pick a random color scheme
				if m.config.ColorScheme == color.ColorGrayscale {
					schemes := color.AllColorSchemes[1:]
					m.config.ColorScheme = schemes[rand.Intn(len(schemes))]
				}
			} else {
				m.animationState.Messages.RandomMsg = "Color Mode: OFF (Black & White)"
			}
			m.animationState.Messages.RandomTimer = 30
			return m, nil

		// Copy current location as URL
		case "U":
			urlStr := persistence.ConfigToFractalURL(m.config, m.animationState.AutoPilot.Enabled, m.animationState.Color.DynamicColor, m.animationState.Transition.Mode)
			if err := clipboard.WriteAll(urlStr); err != nil {
				m.animationState.Messages.URLMsg = fmt.Sprintf("Error copying URL: %v", err)
			} else {
				m.animationState.Messages.URLMsg = "URL copied to clipboard"
			}
			m.animationState.Messages.URLTimer = 120
			return m, tickCmd()

		// Zoom speed control (autopilot)
		case "}":
			// Increase zoom speed
			m.animationState.AutoPilot.ZoomSpeed = math.Min(m.animationState.AutoPilot.ZoomSpeed+0.05, 1.5)
			m.animationState.Messages.RandomMsg = fmt.Sprintf("Zoom speed: %.2fx", m.animationState.AutoPilot.ZoomSpeed)
			m.animationState.Messages.RandomTimer = 30
			return m, nil
		case "{":
			// Decrease zoom speed
			m.animationState.AutoPilot.ZoomSpeed = math.Max(m.animationState.AutoPilot.ZoomSpeed-0.05, 0.90)
			m.animationState.Messages.RandomMsg = fmt.Sprintf("Zoom speed: %.2fx", m.animationState.AutoPilot.ZoomSpeed)
			m.animationState.Messages.RandomTimer = 30
			return m, nil

		// Vantage scene duration control
		case ">":
			// Increase vantage scene duration
			m.animationState.Vantage.SceneDuration = int(math.Min(float64(m.animationState.Vantage.SceneDuration)+20, 600)) // Max 30 seconds (600 ticks)
			sceneSec := float64(m.animationState.Vantage.SceneDuration) / 20.0
			m.animationState.Messages.RandomMsg = fmt.Sprintf("Vantage duration: %.1fs/scene", sceneSec)
			m.animationState.Messages.RandomTimer = 30
			return m, nil
		case "<":
			// Decrease vantage scene duration
			m.animationState.Vantage.SceneDuration = int(math.Max(float64(m.animationState.Vantage.SceneDuration)-20, 20)) // Min 1 second (20 ticks)
			sceneSec := float64(m.animationState.Vantage.SceneDuration) / 20.0
			m.animationState.Messages.RandomMsg = fmt.Sprintf("Vantage duration: %.1fs/scene", sceneSec)
			m.animationState.Messages.RandomTimer = 30
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
	help.WriteString("  V                  Toggle vantage mode (random scenes with slow pan)\n")
	help.WriteString("                     Dynamic color ON, auto-pilot OFF, new scene every N seconds\n")
	help.WriteString("  >, <               Increase/decrease vantage scene duration (1-30 seconds)\n")
	help.WriteString("  0                  Reset view (position, zoom, and iteration depth)\n\n")

	help.WriteString(keyStyle.Render("Fractal Types:") + "\n")
	help.WriteString("  1  Mandelbrot      2  Julia          3  Burning Ship\n")
	help.WriteString("  4  Tricorn         5  Multibrot-3    6  Multibrot-4\n")
	help.WriteString("  7  Celtic          8  Perpendicular  9  Multibrot-5\n")
	help.WriteString("  m  Manhattan       n  Newton\n\n")

	help.WriteString(keyStyle.Render("Settings:") + "\n")
	help.WriteString("  c       Cycle color schemes (grayscale/blue/rainbow/fire/purple/green/gold/cyan)\n")
	help.WriteString("  C       Toggle dynamic color mode (smooth hue rotation)\n")
	help.WriteString("  X       Toggle color mode (On: color / Off: black & white)\n")
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

			// Parse URL to get display info
			var displayLine string
			if params, err := persistence.ParseFractalURL(bm.URL); err == nil {
				// Format zoom display
				var zoomStr string
				if params.Zoom >= 10000.0 {
					zoomStr = fmt.Sprintf("%.1e", params.Zoom)
				} else {
					zoomStr = fmt.Sprintf("%.2f", params.Zoom)
				}
				displayLine = fmt.Sprintf("%s%-25s %s @ (%.4f, %.4f) zoom:%sx",
					prefix, bm.Name, params.FractalType, params.CenterX, params.CenterY, zoomStr)
			} else {
				// Fallback: just show name if URL can't be parsed
				displayLine = fmt.Sprintf("%s%s (invalid URL)", prefix, bm.Name)
			}

			if i == m.bookmarkCursor {
				output.WriteString(selectedStyle.Render(displayLine) + "\n")
			} else {
				output.WriteString(normalStyle.Render(displayLine) + "\n")
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
	if m.animationState.AutoPilot.ZoomDirection < 0 {
		directionArrow = "\u2193" // ↓ (down arrow)
	}

	// Build mode indicator - only one mode can be active at once
	modeIndicator := ""
	if m.animationState.Vantage.Enabled {
		// Active vantage mode: bright blue background
		vantageStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("white")).
			Background(lipgloss.Color("blue")).
			Bold(true)
		sceneDurationSec := float64(m.animationState.Vantage.SceneDuration) / 20.0
		statusText := fmt.Sprintf(" VANTAGE (%.1fs/scene) ", sceneDurationSec)
		modeIndicator = vantageStyle.Render(statusText)
	} else if m.animationState.AutoPilot.Enabled {
		// Active auto-pilot: bright green background
		autoZoomStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("white")).
			Background(lipgloss.Color("green")).
			Bold(true)

		statusText := fmt.Sprintf(" AUTO-PILOT %s (%.2fx) ", directionArrow, m.animationState.AutoPilot.ZoomSpeed)
		if m.animationState.AutoPilot.HasTarget && m.animationState.AutoPilot.PanProgress < 1.0 {
			// Show we're panning toward an interesting region
			statusText = fmt.Sprintf(" AUTO-PILOT \u2192 %s (%.2fx) ", directionArrow, m.animationState.AutoPilot.ZoomSpeed) // → arrow
		}
		modeIndicator = autoZoomStyle.Render(statusText)
	} else {
		// Manual mode: subtle indicator
		inactiveStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")). // Gray
			Bold(false)
		modeIndicator = inactiveStyle.Render(fmt.Sprintf("Explore (z/V to toggle)"))
	}

	// Format zoom level - use scientific notation for high values
	var zoomStr string
	if m.config.Zoom >= 10000.0 {
		zoomStr = fmt.Sprintf("%.1e", m.config.Zoom)
	} else {
		zoomStr = fmt.Sprintf("%.2f", m.config.Zoom)
	}

	colorSchemeDisplay := m.config.ColorScheme
	if !m.animationState.Color.ColorMode {
		colorSchemeDisplay = "grayscale (B&W)"
	}

	status := fmt.Sprintf(
		" %s | Center: (%.4f, %.4f) | Zoom: %sx | Iter: %d | Color: %s ",
		m.config.FractalType,
		m.config.CenterX,
		m.config.CenterY,
		zoomStr,
		m.config.MaxIter,
		colorSchemeDisplay,
	)

	if m.config.FractalType == FractalJulia {
		status += fmt.Sprintf("| Julia: (%.3f, %.3f) ", m.config.JuliaCr, m.config.JuliaCi)
	}

	helpText := infoStyle.Render("[? for help]")

	// Screenshot message indicator
	screenshotIndicator := ""
	if m.animationState.Messages.ScreenshotMsg != "" {
		msgStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("white")).
			Background(lipgloss.Color("green")).
			Bold(true)
		screenshotIndicator = msgStyle.Render(" " + m.animationState.Messages.ScreenshotMsg + " ")
	}

	// Random message indicator
	randomIndicator := ""
	if m.animationState.Messages.RandomMsg != "" {
		msgStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("white")).
			Background(lipgloss.Color("magenta")).
			Bold(true)
		randomIndicator = msgStyle.Render(" " + m.animationState.Messages.RandomMsg + " ")
	}

	// URL message indicator
	urlIndicator := ""
	if m.animationState.Messages.URLMsg != "" {
		msgStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("white")).
			Background(lipgloss.Color("blue")).
			Bold(true)
		urlIndicator = msgStyle.Render(" " + m.animationState.Messages.URLMsg + " ")
	}

	// Build the complete status bar
	statusBar := status
	// Mode indicator (always shown)
	statusBar += modeIndicator + " "
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

// renderFractal generates the ASCII fractal view using the cached renderer
func (m model) renderFractal() string {
	// Update renderer with current config and animation state
	m.renderer.SetConfig(m.config)
	m.renderer.SetColorMode(m.animationState.Color.ColorMode)
	m.renderer.SetDynamicColor(m.animationState.Color.DynamicColor, m.animationState.Color.HueShift)
	if m.animationState.Transition.Mode == TransitionBreakthrough {
		m.renderer.SetBreakthroughTransition(m.animationState.Transition.BreakthroughTransition)
	} else {
		m.renderer.SetBreakthroughTransition(nil)
	}

	// Use the renderer (leverages caching and parallel computation)
	return m.renderer.RenderFractal()
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

// wrapCalculateFractal wraps fractal.CalculateFractal for the render package (returns int)
func wrapCalculateFractal(cr, ci float64, cfg Config) int {
	return int(fractal.CalculateFractal(cr, ci, cfg.FractalType, cfg.MaxIter, cfg.JuliaCr, cfg.JuliaCi))
}

// wrapCalculateFractalFloat64 wraps fractal.CalculateFractal for the search package (returns float64)
func wrapCalculateFractalFloat64(cr, ci float64, cfg persistence.Config) float64 {
	return fractal.CalculateFractal(cr, ci, cfg.FractalType, cfg.MaxIter, cfg.JuliaCr, cfg.JuliaCi)
}

func main() {
	// Parse CLI flags
	config := Config{
		MaxIter:     50,
		CenterX:     -0.5,
		CenterY:     0.0,
		Zoom:        1.0,
		ColorScheme: color.ColorBlue,
		FractalType: FractalMandelbrot,
		JuliaCr:     -0.7,
		JuliaCi:     0.27015,
	}

	// --- Group 1: Display Modes ---
	randomMode := flag.Bool("random", false, "Start with a random interesting view")
	flag.BoolVar(randomMode, "r", false, "Shorthand for --random")

	autopilot := flag.Bool("autopilot", false, "Enable auto-zoom and intelligent exploration")
	flag.BoolVar(autopilot, "a", false, "Shorthand for --autopilot")

	vantage := flag.Bool("vantage", false, "Enable vantage mode (scenic tour of random scenes)")
	flag.BoolVar(vantage, "V", false, "Shorthand for --vantage")

	vantageSceneDurationSpec := flag.Int("vantage-duration", 5, "Seconds per scene in vantage mode")
	mode3d := flag.Bool("3d", false, "GPU-accelerated 3D mode (Mandelbulb)")
	flag.BoolVar(mode3d, "3", false, "Shorthand for --3d")

	// --- Group 2: Navigation & URL ---
	urlFlag := flag.String("url", "", "Load a fractal:// URL")
	flag.StringVar(urlFlag, "u", "", "Shorthand for --url")

	flag.Float64Var(&config.CenterX, "x", -0.5, "Center X coordinate (real axis)")
	flag.Float64Var(&config.CenterY, "y", 0.0, "Center Y coordinate (imaginary axis)")
	flag.Float64Var(&config.Zoom, "zoom", 1.0, "Initial zoom level")
	flag.Float64Var(&config.Zoom, "z", 1.0, "Shorthand for --zoom")

	// --- Group 3: Rendering & Quality ---
	flag.StringVar(&config.FractalType, "type", FractalMandelbrot, "Fractal type")
	flag.StringVar(&config.FractalType, "t", FractalMandelbrot, "Shorthand for --type")

	flag.IntVar(&config.MaxIter, "iterations", 50, "Maximum iterations (detail level)")
	flag.IntVar(&config.MaxIter, "i", 50, "Shorthand for --iterations")

	flag.StringVar(&config.ColorScheme, "color", color.ColorBlue, "Color theme (grayscale, blue, rainbow, etc.)")
	flag.StringVar(&config.ColorScheme, "c", color.ColorBlue, "Shorthand for --color")

	flag.IntVar(&config.Width, "width", 0, "Width override (0=auto)")
	flag.IntVar(&config.Width, "w", 0, "Shorthand for --width")
	flag.IntVar(&config.Height, "height", 0, "Height override (0=auto)")
	flag.IntVar(&config.Height, "h", 0, "Shorthand for --height")

	// --- Group 4: Fractal Parameters ---
	flag.Float64Var(&config.JuliaCr, "julia-re", -0.7, "Julia set real component")
	flag.Float64Var(&config.JuliaCr, "jr", -0.7, "Shorthand for --julia-re")
	flag.Float64Var(&config.JuliaCi, "julia-im", 0.27015, "Julia set imaginary component")
	flag.Float64Var(&config.JuliaCi, "ji", 0.27015, "Shorthand for --julia-im")

	// --- Group 5: Info ---
	flag.BoolVar(&showVersion, "version", false, "Show version information")
	flag.BoolVar(&showVersion, "v", false, "Shorthand for --version")

	flag.Usage = func() {
		logo := `
                              / \
                             /   \
                            /     \
                       ____/       \____
                       \               /
                        \     * *     /
                         \  * F R *  /
                          \ * A C * /
                           \ * T * /
                            \* A */
                             \ L /
                              \I/
                               S
                               v
`
		logoStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("81"))
		for _, line := range strings.Split(logo, "\n") {
			fmt.Fprintln(os.Stderr, logoStyle.Render(line))
		}
		fmt.Fprintln(os.Stderr, lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Italic(true).Render("      ~ deep iterative crystalline exploration ~\n"))
		fmt.Fprintln(os.Stderr, "Usage:")
		fmt.Fprintf(os.Stderr, "  fractalis [options] [url]\n")
		fmt.Fprintf(os.Stderr, "  fractalis serve [port]\n\n")

		titleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("81")).Bold(true)
		flagStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("226"))
		descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("250"))

		printGroup := func(title string, flags [][2]string) {
			fmt.Fprintln(os.Stderr, titleStyle.Render(title))
			for _, f := range flags {
				fmt.Fprintf(os.Stderr, "  %-25s %s\n", flagStyle.Render(f[0]), descStyle.Render(f[1]))
			}
			fmt.Fprintln(os.Stderr)
		}

		printGroup("DISPLAY MODES:", [][2]string{
			{"-r, --random", "Start with a random interesting view"},
			{"-a, --autopilot", "Enable auto-zoom and intelligent exploration"},
			{"-V, --vantage", "Enable vantage mode (scenic tour)"},
			{"--vantage-duration <sec>", "Seconds per scene in vantage mode (default 5)"},
			{"-3, --3d", "GPU-accelerated 3D mode (Mandelbulb)"},
		})

		printGroup("NAVIGATION & URL:", [][2]string{
			{"-u, --url <url>", "Load a fractal:// URL (also positional)"},
			{"-x <float>", "Center X coordinate (real axis)"},
			{"-y <float>", "Center Y coordinate (imaginary axis)"},
			{"-z, --zoom <float>", "Initial zoom level (default 1.0)"},
		})

		printGroup("RENDERING & QUALITY:", [][2]string{
			{"-t, --type <type>", "Fractal type (mandelbrot, julia, mandelbulb, mandelbox...)"},
			{"-i, --iterations <n>", "Maximum iterations (detail level, default 50)"},
			{"-c, --color <theme>", "Color theme (grayscale, blue, rainbow...)"},
			{"-w, --width <pixels>", "Width override (0=auto)"},
			{"-h, --height <pixels>", "Height override (0=auto)"},
		})

		printGroup("FRACTAL PARAMETERS:", [][2]string{
			{"--julia-re <float>", "Julia set real component (default -0.7)"},
			{"--julia-im <float>", "Julia set imaginary component (default 0.27)"},
			{"--jr, --ji", "Shorthands for the above"},
		})

		printGroup("INFO:", [][2]string{
			{"-v, --version", "Show version information"},
			{"-h, --help", "Show this help menu"},
		})
	}

	flag.Parse()

	// Apply vantage duration
	vantageSceneDuration := *vantageSceneDurationSpec

	// Check for URL argument (positional) or --url flag
	urlAutopilot := false
	urlDynamicColor := false
	urlTransitionMode := TransitionNone

	args := flag.Args()
	if len(args) > 0 {
		arg := args[0]

		if arg == "serve" {
			port := 8080
			if len(args) > 1 {
				p, err := strconv.Atoi(args[1])
				if err == nil {
					port = p
				}
			}
			if err := web.Serve(port); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			return
		}

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

	// Validate color scheme
	if !persistence.IsValidColorTheme(config.ColorScheme) {
		fmt.Fprintf(os.Stderr, "Invalid color scheme: %s. Using grayscale.\n", config.ColorScheme)
		config.ColorScheme = color.ColorGrayscale
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
		FractalMandelbulb:    true,
		FractalMandelbox:     true,
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
		// Initialize the interest calculator (required by applyRandom -> isViewUniform)
		tempModel.calculator = search.NewInterestCalculator(wrapCalculateFractalFloat64)
		tempModel.applyRandom()
		config = tempModel.config
	}

	// Check if 3D mode is requested
	if *mode3d {
		// Run 3D mode with Ebiten
		game := ebiten.NewGame(config)
		if err := game.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Error running 3D mode: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Always run in interactive mode
	// Load bookmarks
	bookmarks, err := persistence.LoadBookmarks()
	bookmarksLoadedOk := err == nil
	if err != nil {
		// Non-fatal - just start with empty bookmarks
		bookmarks = []persistence.Bookmark{}
	}

	// Run interactive TUI mode
	animState := animation.NewAnimationState()
	animState.AutoPilot.BaseMaxIter = config.MaxIter
	animState.AutoPilot.ZoomDirection = 1
	animState.Transition.Mode = TransitionFade
	animState.Transition.FadeTransition = transition.NewFadeTransition(allFractalTypes)

	m := model{
		config:            config,
		bookmarks:         bookmarks,
		bookmarksLoadedOk: bookmarksLoadedOk,
		animationState:    animState,
		renderer:          renderlib.NewRenderer(config, wrapCalculateFractal),
	}

	// Initialize the interest calculator
	m.calculator = search.NewInterestCalculator(wrapCalculateFractalFloat64)

	// Apply URL-based state if provided
	if *urlFlag != "" {
		m.animationState.AutoPilot.Enabled = urlAutopilot
		m.animationState.Color.DynamicColor = urlDynamicColor
		m.animationState.Transition.Mode = urlTransitionMode
	}

	// Apply CLI autopilot flag if provided
	if *autopilot {
		m.animationState.AutoPilot.Enabled = true
	}

	// Apply CLI vantage flag if provided
	if *vantage {
		m.animationState.Vantage.Enabled = true
		// vantage duration in seconds -> convert to ticks (50ms per tick)
		m.animationState.Vantage.SceneDuration = vantageSceneDuration * 20 // 20 ticks per second
		m.animationState.Vantage.SceneTimer = 0                            // Will trigger first scene immediately
		// Force dynamic color on and autopilot off
		m.animationState.Color.DynamicColor = true
		m.animationState.AutoPilot.Enabled = false
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running interactive mode: %v\n", err)
		os.Exit(1)
	}
}
