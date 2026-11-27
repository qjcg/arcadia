package animation

// AutoPilotState holds the iteration bookkeeping the journeys share.
type AutoPilotState struct {
	BaseMaxIter int // Base iteration count (for adaptive scaling)
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
	AutoPilot AutoPilotState
	Color     ColorState
	Messages  MessageState
}

// NewAnimationState creates initialized animation state
func NewAnimationState() AnimationState {
	return AnimationState{
		Color: ColorState{
			ColorMode: true,
		},
	}
}
