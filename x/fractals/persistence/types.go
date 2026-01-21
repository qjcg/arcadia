package persistence

// URLMode represents the type of fractal URL
type URLMode string

const (
	ModeStandard URLMode = "standard"
	ModeRandom   URLMode = "random"
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

// Transition animation modes
const (
	TransitionNone         = 0
	TransitionFade         = 1
	TransitionZoomOutIn    = 2
	TransitionRotate       = 3
	TransitionBreakthrough = 4
)
