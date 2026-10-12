package animation

import "github.com/qjcg/arcadia/exp/fractalis/internal/tui/transition"

// AutoPilotState holds the iteration bookkeeping the journeys share.
type AutoPilotState struct {
	BaseMaxIter int // Base iteration count (for adaptive scaling)
}

// VantageState holds vantage mode specific animation state
type VantageState struct {
	Enabled       bool // Vantage mode active
	SceneDuration int  // Number of ticks per scene
	SceneTimer    int  // Countdown for current scene
	Initialized   bool // Track if mode has been initialized
	// Tour pan state (collapses into the Vantage pilot with its ticket)
	TargetX     float64
	TargetY     float64
	HasTarget   bool
	PanProgress float64
	BaseMaxIter int
}

// TransitionState holds fractal transition animation state
type TransitionState struct {
	Mode      int     // 0=none, 1=fade, 2=zoom_out_in, 3=rotate, 4=breakthrough
	Progress  float64 // 0.0 to 1.0, progress through transition
	Target    string  // Target fractal type for transition
	ZoomStart float64 // Starting zoom level for transition

	// Transition implementations
	FadeTransition         *transition.Fade
	ZoomOutInTransition    *transition.ZoomOutIn
	RotateTransition       *transition.Rotate
	BreakthroughTransition *transition.Breakthrough
}

// ColorState holds dynamic color animation state
type ColorState struct {
	ColorMode    bool    // Enable colorized output (default true)
	DynamicColor bool    // Enable smooth hue rotation
	HueShift     float64 // Current hue shift in degrees (0-360)
}

// MessageState holds temporary message display state
type MessageState struct {
	ScreenshotMsg   string // Message to display after screenshot
	ScreenshotTimer int    // Countdown for hiding screenshot message
	RandomMsg       string // Message to display after randomization
	RandomTimer     int    // Countdown for hiding random message
	URLMsg          string // Message to display when copying URL
	URLTimer        int    // Countdown for hiding URL message
}

// AnimationState holds all animation-related state
type AnimationState struct {
	AutoPilot  AutoPilotState
	Vantage    VantageState
	Transition TransitionState
	Color      ColorState
	Messages   MessageState
}

// NewAnimationState creates initialized animation state
func NewAnimationState() AnimationState {
	return AnimationState{
		Vantage: VantageState{
			SceneDuration: DefaultVantageSceneDuration,
		},
		Color: ColorState{
			ColorMode: true,
		},
	}
}
