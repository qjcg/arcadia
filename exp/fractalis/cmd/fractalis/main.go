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

	"github.com/qjcg/arcadia/exp/fractalis/internal/core/fractal"
	"github.com/qjcg/arcadia/exp/fractalis/internal/core/persistence"
	"github.com/qjcg/arcadia/exp/fractalis/internal/core/pilot"
	"github.com/qjcg/arcadia/exp/fractalis/internal/core/search"
	"github.com/qjcg/arcadia/exp/fractalis/internal/core/transition"
	"github.com/qjcg/arcadia/exp/fractalis/internal/tui/animation"
	"github.com/qjcg/arcadia/exp/fractalis/internal/tui/color"
	renderlib "github.com/qjcg/arcadia/exp/fractalis/internal/tui/render"
	tuitransition "github.com/qjcg/arcadia/exp/fractalis/internal/tui/transition"
	"github.com/qjcg/arcadia/exp/fractalis/internal/web"
)

const (
	// ASCII characters from sparse to dense
	asciiChars = " .:-=+*#%@"

	// Transition animation modes
	TransitionNone         = persistence.TransitionNone
	TransitionFade         = tuitransition.TransitionFade
	TransitionZoomOutIn    = tuitransition.TransitionZoomOutIn
	TransitionRotate       = tuitransition.TransitionRotate
	TransitionBreakthrough = tuitransition.TransitionBreakthrough

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
	allFractalTypes = search.AllFractalTypes

	// Known interesting coordinates for random exploration
	// These are "seed points" near interesting fractal features
	interestingMandelbrot = search.InterestingMandelbrot
	interestingJulia      = search.InterestingJulia
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
	// Transition runner: boundary arrivals, schedules, crossings
	runner *transition.Runner
	// Effect registry for rendering crossing frames
	effects *transition.Registry
	// Effective transition schedule (URL-seeded; presets land in #7)
	schedule pilot.Schedule
	// Crossing in flight, if any
	crossing *transition.Crossing
	// The registered pilot currently driving the view
	pilotID pilot.ID
	// The Auto-zoom pilot instance the keys and URL parameters steer
	autozoom *pilot.AutoZoom
	// Seeker backing the Auto-zoom pilot's interesting-region search
	seek *autoZoomSeeker
	// The Vantage pilot: tours curated viewpoints on its scene timer
	vantage *pilot.Vantage
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

// autoZoomSeeker adapts the TUI's interest calculator to the
// Auto-zoom pilot's Seeker seam. Julia parameters live outside the
// pilot's View yet shape Julia regions, so the tick loop refreshes
// them before each step.
type autoZoomSeeker struct {
	calc    *search.InterestCalculator
	juliaCr float64
	juliaCi float64
}

func (s *autoZoomSeeker) FindInterestingPoint(v pilot.View) (float64, float64) {
	return s.calc.FindInterestingPoint(search.DescentSearchPasses(), s.configFor(v))
}

func (s *autoZoomSeeker) IsViewUniform(v pilot.View) bool {
	return s.calc.IsViewUniform(s.configFor(v))
}

func (s *autoZoomSeeker) configFor(v pilot.View) persistence.Config {
	return persistence.Config{
		FractalType: v.FractalType,
		CenterX:     v.CenterX,
		CenterY:     v.CenterY,
		Zoom:        v.Zoom,
		MaxIter:     v.MaxIter,
		JuliaCr:     s.juliaCr,
		JuliaCi:     s.juliaCi,
	}
}

// autoZoomPilot returns the Auto-zoom pilot instance, creating it on
// first use. Steering mutates the receiver, so call it on the model
// copy an Update returns (display code must use autoZoomDisplay).
func (m *model) autoZoomPilot() *pilot.AutoZoom {
	if m.autozoom == nil {
		if m.seek == nil {
			m.seek = &autoZoomSeeker{calc: m.calculator}
		}
		m.autozoom = pilot.NewAutoZoom(m.seek)
	}
	return m.autozoom
}

// autoZoomDisplay returns the steered Auto-zoom parameters and
// journey state for display. Display reads never create the pilot.
func (m model) autoZoomDisplay() (speed float64, direction pilot.Direction, panning bool) {
	if m.autozoom == nil {
		return pilot.DefaultSpeed, pilot.DirectionIn, false
	}
	return m.autozoom.Speed(), m.autozoom.Direction(), m.autozoom.Panning()
}

// Init initializes the Bubble Tea model
func (m model) Init() tea.Cmd {
	// The journey starts from a URL, a flag, or a keypress
	if m.activePilot() != pilot.ManualID {
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

	url := persistence.ConfigToFractalURL(m.config, m.activePilot(), m.autoZoomPilot().Speed(), m.autoZoomPilot().Direction(), m.sceneDuration(), m.animationState.Color.DynamicColor, m.schedule)

	bookmark := persistence.Bookmark{
		Name: name,
		URL:  url,
	}

	m.bookmarks = append(m.bookmarks, bookmark)
	return persistence.SaveBookmarks(m.bookmarks)
}

// activePilot returns the registered pilot currently driving the
// view. selectPilot records the id; the Manual pilot is the default.
func (m model) activePilot() pilot.ID {
	if m.pilotID == "" {
		return pilot.ManualID
	}
	return m.pilotID
}

// sceneDuration returns the Vantage pilot's scene duration, or zero
// when there is no Vantage pilot.
func (m model) sceneDuration() int {
	if m.vantage == nil {
		return 0
	}
	return m.vantage.SceneDuration()
}

// selectPilot makes id the one active pilot, replacing whatever was
// active. The engine seam runs the pilot's motion; switching pilots
// replaces the active one cleanly.
func (m *model) selectPilot(id pilot.ID) {
	m.pilotID = id

	switch id {
	case pilot.AutoZoomID:
		m.seam().Activate(m.autoZoomPilot())
		if m.animationState.AutoPilot.BaseMaxIter == 0 {
			m.animationState.AutoPilot.BaseMaxIter = m.config.MaxIter
		}
	case pilot.VantageID:
		if m.vantage == nil {
			m.vantage = pilot.NewVantage(pilot.VantageOptions{})
		}
		m.seam().Activate(m.vantage)
		m.animationState.Color.DynamicColor = true
		if m.animationState.AutoPilot.BaseMaxIter == 0 {
			m.animationState.AutoPilot.BaseMaxIter = m.config.MaxIter
		}
	default:
		m.seam().Activate(pilot.Manual())
	}
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

	m.selectPilot(params.Pilot)
	m.autoZoomPilot().SetSpeed(params.PilotSpeed)
	m.autoZoomPilot().SetDirection(params.PilotDirection)

	if m.vantage != nil {
		m.vantage.SetSceneDuration(params.SceneDuration)
	}

	m.animationState.Color.DynamicColor = params.DynamicColorEnabled
	if params.ScheduleSet {
		m.schedule = params.Schedule
		m.seam().SetSchedule(m.schedule)
	}

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
		m.animationState.Transition.FadeTransition = tuitransition.NewFadeTransition(allFractalTypes)
		m.animationState.Transition.FadeTransition.Start(m.config.FractalType)
		m.animationState.Messages.RandomMsg = m.animationState.Transition.FadeTransition.GetMessage()
	case TransitionZoomOutIn:
		m.animationState.Transition.ZoomOutInTransition = tuitransition.NewZoomOutInTransition(allFractalTypes)
		m.animationState.Transition.ZoomOutInTransition.Start(m.config.FractalType)
		m.animationState.Messages.RandomMsg = m.animationState.Transition.ZoomOutInTransition.GetMessage()
	case TransitionRotate:
		m.animationState.Transition.RotateTransition = tuitransition.NewRotateTransition(allFractalTypes)
		m.animationState.Transition.RotateTransition.Start(m.config.FractalType)
		m.animationState.Messages.RandomMsg = m.animationState.Transition.RotateTransition.GetMessage()
	case TransitionBreakthrough:
		m.animationState.Transition.BreakthroughTransition = tuitransition.NewBreakthroughTransition(allFractalTypes)
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
	// Random color scheme
	if m.animationState.Color.ColorMode {
		rng := rand.New(rand.NewSource(time.Now().UnixNano()))
		// Pick a random color scheme (excluding grayscale)
		schemes := color.AllColorSchemes[1:]
		m.config.ColorScheme = schemes[rng.Intn(len(schemes))]
	} else {
		m.config.ColorScheme = color.ColorGrayscale
	}

	// Use core randomization logic
	msg := search.RandomizeConfig(&m.config, m.calculator)
	m.animationState.Messages.RandomMsg = msg
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
		adaptiveIter := min(
			// Cap at reasonable maximum to avoid performance issues
			baseIter+int(logZoom*animation.IterationScaleFactor), animation.MaxIterationCap,
		)
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

// findInterestingPoint searches for a point with high variation/detail
// Delegates to the interest calculator
func (m model) findInterestingPoint() (float64, float64) {
	return m.calculator.FindInterestingPoint(search.DefaultSearchPasses(), m.config)
}

// view returns the current view state for the pilot engine seam.
func (m model) view() pilot.View {
	return pilot.View{
		FractalType: m.config.FractalType,
		CenterX:     m.config.CenterX,
		CenterY:     m.config.CenterY,
		Zoom:        m.config.Zoom,
		MaxIter:     m.config.MaxIter,
		JuliaCr:     m.config.JuliaCr,
		JuliaCi:     m.config.JuliaCi,
	}
}

// applyView writes a pilot-evolved view back to the config.
func (m *model) applyView(v pilot.View) {
	m.config.FractalType = v.FractalType
	m.config.CenterX = v.CenterX
	m.config.CenterY = v.CenterY
	m.config.Zoom = v.Zoom
	m.config.MaxIter = v.MaxIter
	m.config.JuliaCr = v.JuliaCr
	m.config.JuliaCi = v.JuliaCi
}

// seam returns the model's transition runner, building the default
// one on first use. Selection lives in the runner's scheduler and
// destination policy; randomness is seeded at the edge.
func (m *model) seam() *transition.Runner {
	if m.runner == nil {
		rng := rand.New(rand.NewSource(time.Now().UnixNano()))
		m.runner = transition.NewRunner(
			pilot.NewScheduler(rng),
			pilot.NewRandomDestination(allFractalTypes, rng),
		)
		m.effects = transition.NewRegistry()
		m.effects.Register(transition.FadeID, transition.Fade{})
	}
	return m.runner
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
		if m.activePilot() == pilot.AutoZoomID {
			// Update iteration count adaptively based on zoom level
			m.config.MaxIter = m.calculateAdaptiveMaxIter()
		}

		// Julia parameters live outside the pilot's View yet shape the
		// regions it seeks; refresh the seeker's copy before stepping.
		if m.seek != nil {
			m.seek.juliaCr, m.seek.juliaCi = m.config.JuliaCr, m.config.JuliaCi
		}

		// The active pilot evolves the view through the engine seam,
		// boundary rules included. A boundary arrival fires a
		// transition from the effective schedule; an empty schedule
		// clamps and the view stays put. The schedule's crossings
		// supersede the legacy Mode-driven transitions, which the
		// effect port retires.
		next, crossing := m.seam().Tick(m.view())
		m.applyView(next)
		m.crossing = crossing
		if crossing != nil {
			// A crossing freezes the view at its source and lands
			// the destination when progress completes.
			if crossing.Progress.Landed() {
				m.crossing = nil // the destination is now the view
				if m.config.FractalType == FractalJulia {
					// Arriving at Julia opens on its default parameters.
					m.config.JuliaCr = -0.7
					m.config.JuliaCi = 0.27015
				}
			}
			return m, tickCmd()
		}

		// Update dynamic color hue shift
		if m.animationState.Color.DynamicColor {
			m.animationState.Color.HueShift += 0.5 // Shift 0.5 degrees per tick (30 degrees/sec at 60fps, full cycle every 12 seconds)
			if m.animationState.Color.HueShift >= 360.0 {
				m.animationState.Color.HueShift -= 360.0
			}
		}

		// The Auto-zoom journey runs every tick
		if m.activePilot() == pilot.AutoZoomID {
			return m, tickCmd()
		}

		// Update dynamic color even when not auto-zooming (for manual navigation)
		if m.animationState.Color.DynamicColor {
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

		// Continue ticking while a pilot drives the view, or any timer
		// is active
		if m.activePilot() != pilot.ManualID ||
			m.animationState.Messages.ScreenshotTimer > 0 || m.animationState.Messages.RandomTimer > 0 || m.animationState.Messages.URLTimer > 0 {
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
			// Only start tick loop if not already on a journey
			if m.activePilot() == pilot.ManualID {
				return m, tickCmd()
			}
			return m, nil

		case "z":
			// Toggle the Auto-zoom pilot (replaces whatever was active)
			if m.activePilot() == pilot.AutoZoomID {
				m.selectPilot(pilot.ManualID)
				return m, nil
			}
			m.selectPilot(pilot.AutoZoomID)
			return m, tickCmd()

		case "r":
			// Reverse the Auto-zoom pilot's zoom direction
			m.autoZoomPilot().ReverseDirection()
			return m, nil

		case "V":
			// Vantage pilot: jump to the next scene when already active,
			// or activate it when not
			if m.activePilot() == pilot.VantageID {
				m.vantage.NextScene()
				m.animationState.Messages.RandomMsg = "Vantage: Next Scene"
				m.animationState.Messages.RandomTimer = 30
				return m, tickCmd()
			}

			m.selectPilot(pilot.VantageID)
			m.animationState.Messages.RandomMsg = fmt.Sprintf("Vantage Mode: ON (%.1fs/scene)", float64(m.sceneDuration())/20.0)
			m.animationState.Messages.RandomTimer = 30
			return m, tickCmd()

		case "v":
			// Toggle the Vantage pilot
			if m.activePilot() == pilot.VantageID {
				m.selectPilot(pilot.ManualID)
				m.animationState.Messages.RandomMsg = "Vantage Mode: OFF"
				m.animationState.Messages.RandomTimer = 30
				return m, nil
			}

			m.selectPilot(pilot.VantageID)
			m.animationState.Messages.RandomMsg = fmt.Sprintf("Vantage Mode: ON (%.1fs/scene)", float64(m.sceneDuration())/20.0)
			m.animationState.Messages.RandomTimer = 30
			return m, tickCmd()

		// Navigation - Pan
		case "up", "w":
			m.config.CenterY -= 0.1 / m.config.Zoom
			return m, nil
		case "down", "s":
			m.config.CenterY += 0.1 / m.config.Zoom
			return m, nil
		case "left", "a":
			m.config.CenterX -= 0.1 / m.config.Zoom
			return m, nil
		case "right", "d":
			m.config.CenterX += 0.1 / m.config.Zoom
			return m, nil

		// Zoom
		case "+", "=":
			// If the Auto-zoom pilot is active, steer the journey inward
			// Otherwise perform manual zoom in
			if m.activePilot() == pilot.AutoZoomID {
				m.autoZoomPilot().SetDirection(pilot.DirectionIn)
			} else {
				m.config.Zoom *= 1.2
			}
			return m, tickCmd()
		case "i":
			// Manual zoom in only
			m.config.Zoom *= 1.2
			return m, tickCmd()
		case "-", "_":
			// If the Auto-zoom pilot is active, steer the journey outward
			// Otherwise perform manual zoom out
			if m.activePilot() == pilot.AutoZoomID {
				m.autoZoomPilot().SetDirection(pilot.DirectionOut)
			} else {
				m.config.Zoom /= 1.2
			}
			return m, tickCmd()
		case "o":
			// Manual zoom out only
			m.config.Zoom /= 1.2
			return m, tickCmd()
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
			urlStr := persistence.ConfigToFractalURL(m.config, m.activePilot(), m.autoZoomPilot().Speed(), m.autoZoomPilot().Direction(), m.sceneDuration(), m.animationState.Color.DynamicColor, m.schedule)
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
			m.autoZoomPilot().SetSpeed(min(m.autoZoomPilot().Speed()+0.05, animation.MaxZoomSpeed))
			m.animationState.Messages.RandomMsg = fmt.Sprintf("Zoom speed: %.2fx", m.autoZoomPilot().Speed())
			m.animationState.Messages.RandomTimer = 30
			return m, nil
		case "{":
			// Decrease zoom speed
			m.autoZoomPilot().SetSpeed(max(m.autoZoomPilot().Speed()-0.05, animation.MinZoomSpeed))
			m.animationState.Messages.RandomMsg = fmt.Sprintf("Zoom speed: %.2fx", m.autoZoomPilot().Speed())
			m.animationState.Messages.RandomTimer = 30
			return m, nil

		// Vantage scene duration control
		case ">":
			// Increase vantage scene duration
			if m.vantage != nil {
				m.vantage.SetSceneDuration(min(m.vantage.SceneDuration()+20, 600)) // Max 30 seconds (600 ticks)
				sceneSec := float64(m.vantage.SceneDuration()) / 20.0
				m.animationState.Messages.RandomMsg = fmt.Sprintf("Vantage duration: %.1fs/scene", sceneSec)
				m.animationState.Messages.RandomTimer = 30
			}
			return m, nil
		case "<":
			// Decrease vantage scene duration
			if m.vantage != nil {
				m.vantage.SetSceneDuration(max(m.vantage.SceneDuration()-20, 20)) // Min 1 second (20 ticks)
				sceneSec := float64(m.vantage.SceneDuration()) / 20.0
				m.animationState.Messages.RandomMsg = fmt.Sprintf("Vantage duration: %.1fs/scene", sceneSec)
				m.animationState.Messages.RandomTimer = 30
			}
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
	help.WriteString("  V                  Start/jump vantage mode (random scenes with slow pan)\n")
	help.WriteString("  v                  Toggle vantage mode off/on\n")
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
	zoomSpeed, zoomDirection, panning := m.autoZoomDisplay()
	directionArrow := "\u2191" // ↑ (up arrow)
	if zoomDirection < 0 {
		directionArrow = "\u2193" // ↓ (down arrow)
	}

	// Build the active pilot indicator - only one pilot is active at
	// a time, and the status bar names it
	modeIndicator := ""
	switch id := m.activePilot(); id {
	case pilot.VantageID:
		// Active Vantage pilot: bright blue background
		vantageStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("white")).
			Background(lipgloss.Color("blue")).
			Bold(true)
		sceneDurationSec := float64(m.sceneDuration()) / 20.0
		statusText := fmt.Sprintf(" %s pilot (%.1fs/scene) ", pilot.Name(id), sceneDurationSec)
		modeIndicator = vantageStyle.Render(statusText)
	case pilot.AutoZoomID:
		// Active Auto-zoom pilot: bright green background
		autoZoomStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("white")).
			Background(lipgloss.Color("green")).
			Bold(true)

		statusText := fmt.Sprintf(" %s pilot %s (%.2fx) ", pilot.Name(id), directionArrow, zoomSpeed)
		if panning {
			// Show we're panning toward an interesting region
			statusText = fmt.Sprintf(" %s pilot \u2192 %s (%.2fx) ", pilot.Name(id), directionArrow, zoomSpeed) // → arrow
		}
		modeIndicator = autoZoomStyle.Render(statusText)
	default:
		// Manual pilot: subtle indicator
		inactiveStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")). // Gray
			Bold(false)
		modeIndicator = inactiveStyle.Render(fmt.Sprintf("%s pilot (z/v to toggle)", pilot.Name(id)))
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
	// A crossing in flight renders over the shared progress model:
	// the effect splits the frame between the two views.
	if c := m.crossing; c != nil {
		return m.renderCrossing(*c)
	}

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

// renderCrossing draws one frame of a boundary crossing: the effect
// named by the scheduler mixes the source and destination views over
// the shared progress model.
func (m model) renderCrossing(c transition.Crossing) string {
	var effect transition.Effect = transition.Fade{}
	if m.effects != nil {
		if e, ok := m.effects.Lookup(c.Effect); ok {
			effect = e
		}
	}
	frame := effect.Frame(c.Source, c.Destination, c.Progress)
	return m.renderer.RenderCrossfade(
		m.configForView(frame.Source),
		m.configForView(frame.Destination),
		frame.Mix,
	)
}

// configForView projects a pilot view onto the render config, keeping
// the terminal size and palette of the current config.
func (m model) configForView(v pilot.View) Config {
	cfg := m.config
	cfg.FractalType = v.FractalType
	cfg.CenterX = v.CenterX
	cfg.CenterY = v.CenterY
	cfg.Zoom = v.Zoom
	cfg.MaxIter = v.MaxIter
	return cfg
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
	engineFlag := flag.String("engine", persistence.EngineBubbleTea, engineChoiceHelp)
	flag.StringVar(engineFlag, "e", persistence.EngineBubbleTea, "Shorthand for --engine")

	fulscreenFlag := flag.Bool("fullscreen", false, "Start in fullscreen mode (ebiten only)")
	flag.BoolVar(fulscreenFlag, "f", false, "Shorthand for --fullscreen")

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
		for line := range strings.SplitSeq(logo, "\n") {
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
			{"-e, --engine <name>", engineChoiceHelp},
			{"-f, --fullscreen", "Start in fullscreen mode (ebiten only)"},
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
	urlPilot := pilot.ManualID
	urlPilotSpeed := pilot.DefaultSpeed
	urlPilotDirection := pilot.DirectionIn
	urlSceneDuration := pilot.DefaultSceneDuration
	urlDynamicColor := false
	urlSchedule := pilot.Schedule{}
	urlScheduleSet := false

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
		urlPilot = params.Pilot
		urlPilotSpeed = params.PilotSpeed
		urlPilotDirection = params.PilotDirection
		urlSceneDuration = params.SceneDuration
		urlDynamicColor = params.DynamicColorEnabled
		urlSchedule = params.Schedule
		urlScheduleSet = params.ScheduleSet

		if params.Engine != "" {
			config.Engine = params.Engine
		}

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

	// Check engine flag override
	if *engineFlag != persistence.EngineBubbleTea {
		config.Engine = *engineFlag
	}

	// Apply fullscreen flag
	config.Fullscreen = *fulscreenFlag

	// Validate engine
	if config.Engine == "" {
		config.Engine = persistence.EngineBubbleTea
	}

	// Adjust default center for Julia set if user didn't specify custom center
	if config.FractalType == FractalJulia && config.CenterX == -0.5 && config.CenterY == 0.0 {
		config.CenterX = 0.0
		config.CenterY = 0.0
	}

	// Apply random if requested
	if *randomMode {
		// Create temporary model to use random function
		tempModel := model{
			config: config,
			// Initialize the interest calculator (required by applyRandom)
			calculator: search.NewInterestCalculator(wrapCalculateFractalFloat64),
		}
		tempModel.applyRandom()
		config = tempModel.config
	}

	// Check which engine to use
	if config.Engine == persistence.EngineEbiten {
		// Apply flags to config for Ebiten
		config.AutoPilotEnabled = *autopilot || urlPilot == pilot.AutoZoomID
		config.VantageEnabled = *vantage || urlPilot == pilot.VantageID
		config.VantageSceneDuration = vantageSceneDuration

		// Run Ebiten engine (compiled in only with -tags ebiten)
		if err := runEbitenEngine(config); err != nil {
			fmt.Fprintf(os.Stderr, "Error running ebiten engine: %v\n", err)
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
	animState.Transition.Mode = TransitionFade
	animState.Transition.FadeTransition = tuitransition.NewFadeTransition(allFractalTypes)

	// The Vantage pilot tours curated viewpoints, panning each scene
	// toward a discovered interesting point.
	calculator := search.NewInterestCalculator(wrapCalculateFractalFloat64)
	vantagePilot := pilot.NewVantage(pilot.VantageOptions{
		SceneDuration: urlSceneDuration,
		Discover: func(v pilot.View) (float64, float64) {
			return calculator.FindInterestingPoint(search.DefaultSearchPasses(), persistence.Config{
				FractalType: v.FractalType,
				CenterX:     v.CenterX,
				CenterY:     v.CenterY,
				Zoom:        v.Zoom,
				MaxIter:     v.MaxIter,
				JuliaCr:     v.JuliaCr,
				JuliaCi:     v.JuliaCi,
			})
		},
	})

	m := model{
		config:            config,
		bookmarks:         bookmarks,
		bookmarksLoadedOk: bookmarksLoadedOk,
		animationState:    animState,
		renderer:          renderlib.NewRenderer(config, wrapCalculateFractal),

		// Initialize the interest calculator
		calculator: calculator,
		// The Vantage pilot running the curated tour
		vantage: vantagePilot,
	}

	// Apply URL-based state if provided
	if *urlFlag != "" {
		m.selectPilot(urlPilot)
		m.autoZoomPilot().SetSpeed(urlPilotSpeed)
		m.autoZoomPilot().SetDirection(urlPilotDirection)
		m.animationState.Color.DynamicColor = urlDynamicColor
		if urlScheduleSet {
			m.schedule = urlSchedule
			m.seam().SetSchedule(urlSchedule)
		}
	}

	// Apply CLI autopilot flag if provided
	if *autopilot {
		m.selectPilot(pilot.AutoZoomID)
	}

	// Apply CLI vantage flag if provided
	if *vantage {
		// vantage duration in seconds -> convert to ticks (50ms per tick)
		m.selectPilot(pilot.VantageID)
		m.vantage.SetSceneDuration(vantageSceneDuration * 20) // 20 ticks per second
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running interactive mode: %v\n", err)
		os.Exit(1)
	}
}
