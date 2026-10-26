package animation

const (
	// Timer durations (in ticks, at 50ms per tick)
	ScreenshotMessageDuration = 60  // ~3 seconds
	RandomMessageDuration     = 60  // ~3 seconds
	URLMessageDuration        = 120 // ~6 seconds
	TransitionMessageDuration = 30  // ~1.5 seconds
	VantageStatusDuration     = 30  // ~1.5 seconds

	// Zoom speed steering bounds for the Auto-zoom journey
	MinZoomSpeed = 0.90 // Minimum zoom speed
	MaxZoomSpeed = 1.5  // Maximum zoom speed

	// Search intensity
	RandomSearchMaxAttempts = 5

	// Zoom limits
	MaxZoomLimit = 1e15
	MinZoomLimit = 0.1

	// Iteration scaling
	IterationScaleFactor  = 20.0 // Add 20 iterations per decade of zoom
	MaxIterationCap       = 2000 // Cap iterations at this value
	DefaultBaseIterations = 50
)
