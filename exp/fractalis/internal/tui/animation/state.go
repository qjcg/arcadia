package animation

import "github.com/qjcg/arcadia/exp/fractalis/internal/tui/transition"

// AutoPilotState holds the iteration bookkeeping the journeys share.
type AutoPilotState struct {
	BaseMaxIter int // Base iteration count (for adaptive scaling)
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
	Transition TransitionState
	Color      ColorState
	Messages   MessageState
}

// NewAnimationState creates initialized animation state
func NewAnimationState() AnimationState {
	return AnimationState{
		Color: ColorState{
			ColorMode: true,
		},
	}
}
