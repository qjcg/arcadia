package pilot

// Seeker is the interesting-region search service the Auto-zoom
// pilot travels toward. Frontends adapt their search module to it.
type Seeker interface {
	// FindInterestingPoint returns the interesting region to travel
	// toward near the given view.
	FindInterestingPoint(v View) (x, y float64)
	// IsViewUniform reports whether the view has gone visually
	// uniform and needs a fresh region.
	IsViewUniform(v View) bool
}

// Zoom directions the Auto-zoom pilot travels.
type Direction int

const (
	DirectionIn  Direction = 1
	DirectionOut Direction = -1
)

// DefaultSpeed is the Auto-zoom pilot's default zoom multiplier per
// tick.
const DefaultSpeed = 1.05

// Motion pacing for the Auto-zoom journey.
const (
	// PanRate is the fraction of the remaining distance to the target
	// panned each tick.
	PanRate = 0.05
	// PanProgressStep is how much pan progress advances each tick.
	PanProgressStep = 0.02
	// UniformRetargetProgress is the pan progress past which a uniform
	// view seeks a fresh region.
	UniformRetargetProgress = 0.8
)

// AutoZoom is the Auto-zoom pilot: it continuously zooms toward
// interesting regions. It owns its behavior state and its steerable
// speed and direction parameters.
type AutoZoom struct {
	seeker      Seeker
	speed       float64
	direction   Direction
	fractalType string
	targetX     float64
	targetY     float64
	hasTarget   bool
	panProgress float64
}

// NewAutoZoom returns an Auto-zoom pilot traveling inward at the
// default speed, seeking interesting regions with seeker.
func NewAutoZoom(seeker Seeker) *AutoZoom {
	return &AutoZoom{seeker: seeker, speed: DefaultSpeed, direction: DirectionIn}
}

// Speed returns the zoom multiplier per tick.
func (a *AutoZoom) Speed() float64 { return a.speed }

// SetSpeed sets the zoom multiplier per tick.
func (a *AutoZoom) SetSpeed(speed float64) { a.speed = speed }

// Direction returns the direction the pilot zooms toward.
func (a *AutoZoom) Direction() Direction { return a.direction }

// SetDirection sets the direction the pilot zooms toward.
func (a *AutoZoom) SetDirection(d Direction) { a.direction = d }

// ReverseDirection flips the zoom direction.
func (a *AutoZoom) ReverseDirection() { a.direction = -a.direction }

// Step steers the journey one tick: it seeks interesting regions,
// pans toward the current one, and zooms in the steered direction at
// the steered speed.
func (a *AutoZoom) Step(v View) View {
	if v.FractalType != a.fractalType {
		// A new fractal deserves a fresh region.
		a.fractalType = v.FractalType
		a.hasTarget = false
		a.panProgress = 0
	}
	if !a.hasTarget || a.panProgress >= 1 ||
		(a.panProgress > UniformRetargetProgress && a.seeker.IsViewUniform(v)) {
		a.targetX, a.targetY = a.seeker.FindInterestingPoint(v)
		a.hasTarget = true
		a.panProgress = 0
	}
	if a.hasTarget && a.panProgress < 1 {
		dx := a.targetX - v.CenterX
		dy := a.targetY - v.CenterY
		v.CenterX += dx * PanRate
		v.CenterY += dy * PanRate
		if dx*dx+dy*dy < searchDelta(v)*searchDelta(v)*0.01 {
			a.panProgress = 1
		} else {
			a.panProgress += PanProgressStep
		}
	}
	if a.direction > 0 {
		v.Zoom *= a.speed
	} else {
		v.Zoom /= a.speed
	}
	return v
}

// searchDelta returns a search distance that is meaningful at the
// view's zoom: large enough to change the center under float64
// precision at extreme zoom levels.
func searchDelta(v View) float64 {
	delta := (3.5 / v.Zoom) / 20.0
	coordMagnitude := v.CenterX
	if coordMagnitude < 0 {
		coordMagnitude = -coordMagnitude
	}
	if coordMagnitude == 0 {
		coordMagnitude = 1.0
	}
	delta = max(delta, coordMagnitude*1e-12)
	return max(delta, 1e-14)
}

// TransitionSchedule returns the Auto-zoom pilot's transition
// schedule. It stays empty until a scheduler executes schedules, so
// the engine keeps clamping at boundaries.
func (a *AutoZoom) TransitionSchedule() Schedule { return Schedule{} }
