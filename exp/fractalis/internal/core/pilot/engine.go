package pilot

// Boundary is the top or bottom edge of a fractal, where one fractal
// gives way to another.
type Boundary int

// The two boundaries a view can arrive at.
const (
	TopBoundary    Boundary = iota + 1 // the zooming-in limit
	BottomBoundary                     // the zooming-out limit
)

// Zoom limits that define the boundaries. These mirror the zoom range
// the engines have always enforced.
const (
	MaxZoom = 1e15
	MinZoom = 0.1
)

// Engine runs the active pilot against a view and enforces boundary
// rules. There is no "no pilot" state: the Manual pilot is active
// until another one is activated.
type Engine struct {
	active   Pilot
	seen     bool
	lastZoom float64
}

// Active returns the one pilot active at this moment.
func (e *Engine) Active() Pilot {
	if e.active == nil {
		return Manual()
	}
	return e.active
}

// Activate makes p the one active pilot, replacing whatever was
// active.
func (e *Engine) Activate(p Pilot) { e.active = p }

// Observe records the view before a frontend applies manual zoom input.
func (e *Engine) Observe(v View) { e.seen, e.lastZoom = true, v.Zoom }

// Tick advances the view one tick under the active pilot and enforces
// boundary rules. It returns the boundary actively arrived at, if
// any. With an empty transition schedule the view clamps and stays
// put at the boundary; a view resting at a boundary is left alone.
// A non-empty schedule plays a transition across the boundary
// instead, which the scheduler executes when transitions land.
// Movement toward a boundary counts from any source — the pilot or
// manual zoom — so externally driven arrivals fire too.
func (e *Engine) Tick(v View) (next View, boundary *Boundary) {
	next = e.Active().Step(v)
	moving := next.Zoom != v.Zoom || (e.seen && v.Zoom != e.lastZoom)
	defer func() { e.Observe(next) }()
	if !moving || next.FractalType != v.FractalType {
		// Not moving toward a boundary; a resting view is left alone.
		return next, nil
	}

	lower, upper := next.ZoomBounds()
	var arrived *Boundary
	switch {
	case next.Zoom >= upper:
		arrived = new(TopBoundary)
		if e.Active().TransitionSchedule().Empty() {
			next.Zoom = upper
		}
	case next.Zoom <= lower:
		arrived = new(BottomBoundary)
		if e.Active().TransitionSchedule().Empty() {
			next.Zoom = lower
		}
	}
	return next, arrived
}

//go:fix inline
func boundaryPtr(b Boundary) *Boundary { return new(b) }
