package animation

import "github.com/qjcg/arcadia/x/fractalis/transitions"

// AutoPilotState holds auto-pilot specific animation state
type AutoPilotState struct {
	Enabled       bool    // Auto-zoom mode active
	ZoomDirection int     // +1 for zoom in, -1 for zoom out
	ZoomSpeed     float64 // Multiplier for zoom speed (default 1.05)
	TargetX       float64 // Target point to pan toward
	TargetY       float64
	HasTarget     bool    // Whether we have a target to pan toward
	PanProgress   float64 // 0.0 to 1.0, how far we've panned toward target
	BaseMaxIter   int     // Base iteration count (for adaptive scaling)
}

// VantageState holds vantage mode specific animation state
type VantageState struct {
	Enabled       bool // Vantage mode active
	SceneDuration int  // Number of ticks per scene
	SceneTimer    int  // Countdown for current scene
	Initialized   bool // Track if mode has been initialized
	// Shared pan state (uses same fields as AutoPilot)
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
	FadeTransition         *transitions.Fade
	ZoomOutInTransition    *transitions.ZoomOutIn
	RotateTransition       *transitions.Rotate
	BreakthroughTransition *transitions.Breakthrough
}

// ColorState holds dynamic color animation state
type ColorState struct {
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
		AutoPilot: AutoPilotState{
			ZoomSpeed: DefaultZoomSpeed,
		},
		Vantage: VantageState{
			SceneDuration: DefaultVantageSceneDuration,
		},
	}
}
