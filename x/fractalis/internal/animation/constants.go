package animation

const (
	// Timer durations (in ticks, at 50ms per tick)
	ScreenshotMessageDuration = 60  // ~3 seconds
	RandomMessageDuration     = 60  // ~3 seconds
	URLMessageDuration        = 120 // ~6 seconds
	TransitionMessageDuration = 30  // ~1.5 seconds
	VantageStatusDuration     = 30  // ~1.5 seconds

	// Default animation settings
	DefaultZoomSpeed            = 1.05 // 5% zoom per tick
	MinZoomSpeed                = 0.90 // Minimum zoom speed
	MaxZoomSpeed                = 1.5  // Maximum zoom speed
	DefaultVantageSceneDuration = 100  // ~5 seconds (100 ticks at 50ms)
	MinVantageSceneDuration     = 20   // ~1 second
	MaxVantageSceneDuration     = 600  // ~30 seconds

	// Pan speeds
	AutoPilotPanRate  = 0.05 // 5% of remaining distance per tick
	VantagePanRate    = 0.02 // 2% of remaining distance per tick (slower)
	ProgressIncrement = 0.02 // Progress increment per tick

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
