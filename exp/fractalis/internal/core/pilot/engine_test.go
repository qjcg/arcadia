package pilot

import "testing"

// stubPilot is a scripted pilot used to exercise the engine seam.
type stubPilot struct {
	step     func(View) View
	schedule Schedule
}

func (p stubPilot) Step(v View) View {
	if p.step == nil {
		return v
	}
	return p.step(v)
}

func (p stubPilot) TransitionSchedule() Schedule { return p.schedule }

// zoomingBy returns a stub pilot whose policy multiplies zoom by rate
// each tick: rate > 1 zooms in toward the top boundary, rate < 1
// zooms out toward the bottom boundary.
func zoomingBy(rate float64) stubPilot {
	return stubPilot{step: func(v View) View {
		v.Zoom *= rate
		return v
	}}
}

// panningBy returns a stub pilot whose policy pans center X by dx
// each tick without touching zoom.
func panningBy(dx float64) stubPilot {
	return stubPilot{step: func(v View) View {
		v.CenterX += dx
		return v
	}}
}

// atRest returns a stub pilot whose policy never moves the view.
func atRest() stubPilot { return stubPilot{} }

func TestManualIsTheDefaultActivePilot(t *testing.T) {
	var e Engine

	v := View{FractalType: "mandelbrot", CenterX: -0.5, Zoom: 4, MaxIter: 50}
	for range 5 {
		next, arrival := e.Tick(v)
		if next != v {
			t.Fatalf("default active pilot moved the view: got %+v, want %+v", next, v)
		}
		if arrival != nil {
			t.Fatalf("default active pilot reported boundary arrival %v, want none", *arrival)
		}
	}

	if got := e.Active().TransitionSchedule(); !got.Empty() {
		t.Errorf("default active pilot schedule = %+v, want empty", got)
	}
}

func TestViewEvolvesPerPilotPolicy(t *testing.T) {
	var e Engine
	e.Activate(zoomingBy(2))

	v := View{FractalType: "mandelbrot", Zoom: 1, MaxIter: 50}
	for i := range 3 {
		var arrival *Boundary
		v, arrival = e.Tick(v)
		if want := pow2(float64(i + 1)); v.Zoom != want {
			t.Fatalf("tick %d: Zoom = %v, want %v", i+1, v.Zoom, want)
		}
		if arrival != nil {
			t.Fatalf("tick %d: boundary arrival = %v, want none", i+1, *arrival)
		}
	}

	e.Activate(panningBy(0.25))
	p := View{FractalType: "julia", CenterX: 0, Zoom: 3, MaxIter: 50}
	for i := range 2 {
		p, _ = e.Tick(p)
		if want := 0.25 * float64(i+1); p.CenterX != want {
			t.Fatalf("tick %d: CenterX = %v, want %v", i+1, p.CenterX, want)
		}
		if p.Zoom != 3 {
			t.Fatalf("tick %d: Zoom = %v, want 3", i+1, p.Zoom)
		}
	}
}

func TestActivateReplacesTheActivePilot(t *testing.T) {
	var e Engine

	e.Activate(zoomingBy(2))
	v, _ := e.Tick(View{Zoom: 1, MaxIter: 50})
	if v.Zoom != 2 {
		t.Fatalf("first pilot not active: Zoom = %v, want 2", v.Zoom)
	}

	e.Activate(panningBy(0.5))
	v, _ = e.Tick(v)
	if v.Zoom != 2 {
		t.Fatalf("first pilot still active: Zoom = %v, want 2", v.Zoom)
	}
	if v.CenterX != 0.5 {
		t.Fatalf("second pilot not active: CenterX = %v, want 0.5", v.CenterX)
	}
}

func TestBoundaryArrivalClampsOnEmptySchedule(t *testing.T) {
	t.Run("top", func(t *testing.T) {
		var e Engine
		// Overshooting zoom-in with an empty schedule: the view clamps
		// at the boundary and stays put.
		e.Activate(zoomingBy(4))

		v := View{Zoom: MaxZoom / 2, MaxIter: 50}
		next, arrival := e.Tick(v)
		if next.Zoom != MaxZoom {
			t.Fatalf("Zoom = %v, want clamped to %v", next.Zoom, MaxZoom)
		}
		if arrival == nil || *arrival != TopBoundary {
			t.Fatalf("arrival = %v, want TopBoundary", arrival)
		}

		for range 3 {
			next, arrival = e.Tick(next)
			if next.Zoom != MaxZoom {
				t.Fatalf("view did not stay put: Zoom = %v, want %v", next.Zoom, MaxZoom)
			}
			if arrival == nil || *arrival != TopBoundary {
				t.Fatalf("arrival = %v, want TopBoundary while actively moving", arrival)
			}
		}
	})

	t.Run("bottom", func(t *testing.T) {
		var e Engine
		e.Activate(zoomingBy(0.1))

		v := View{Zoom: 1, MaxIter: 50}
		next, arrival := e.Tick(v)
		if next.Zoom != MinZoom {
			t.Fatalf("Zoom = %v, want clamped to %v", next.Zoom, MinZoom)
		}
		if arrival == nil || *arrival != BottomBoundary {
			t.Fatalf("arrival = %v, want BottomBoundary", arrival)
		}

		for range 3 {
			next, arrival = e.Tick(next)
			if next.Zoom != MinZoom {
				t.Fatalf("view did not stay put: Zoom = %v, want %v", next.Zoom, MinZoom)
			}
		}
	})

	t.Run("reaching the limit exactly", func(t *testing.T) {
		var e Engine
		e.Activate(zoomingBy(2))

		next, arrival := e.Tick(View{Zoom: MaxZoom / 2, MaxIter: 50})
		if next.Zoom != MaxZoom {
			t.Fatalf("Zoom = %v, want %v", next.Zoom, MaxZoom)
		}
		if arrival == nil || *arrival != TopBoundary {
			t.Fatalf("arrival = %v, want TopBoundary", arrival)
		}
	})
}

func TestBoundaryArrivalOnlyWhileActivelyMoving(t *testing.T) {
	t.Run("view resting at the boundary is left alone", func(t *testing.T) {
		var e Engine // Manual pilot at rest
		v := View{Zoom: MaxZoom, MaxIter: 50}

		next, arrival := e.Tick(v)
		if next != v {
			t.Fatalf("resting view changed: got %+v, want %+v", next, v)
		}
		if arrival != nil {
			t.Fatalf("arrival = %v, want none for a resting view", *arrival)
		}
	})

	t.Run("panning at the boundary is not arrival", func(t *testing.T) {
		var e Engine
		e.Activate(panningBy(0.25))

		v := View{CenterX: -0.5, Zoom: MaxZoom, MaxIter: 50}
		next, arrival := e.Tick(v)
		if next.Zoom != MaxZoom {
			t.Fatalf("Zoom = %v, want %v", next.Zoom, MaxZoom)
		}
		if next.CenterX == v.CenterX {
			t.Fatal("panning pilot did not move the view")
		}
		if arrival != nil {
			t.Fatalf("arrival = %v, want none while not moving toward a boundary", *arrival)
		}
	})

	t.Run("actively moving at the boundary still reports arrival", func(t *testing.T) {
		var e Engine
		e.Activate(zoomingBy(2))

		next, arrival := e.Tick(View{Zoom: MaxZoom, MaxIter: 50})
		if next.Zoom != MaxZoom {
			t.Fatalf("Zoom = %v, want %v", next.Zoom, MaxZoom)
		}
		if arrival == nil || *arrival != TopBoundary {
			t.Fatalf("arrival = %v, want TopBoundary while actively moving", arrival)
		}
	})
}

func TestBoundaryArrivalReportedRegardlessOfSchedule(t *testing.T) {
	// A non-empty schedule plays a transition instead of the bare
	// clamp; the scheduler lands with the transition core (#6). The
	// seam contract: arrival is reported either way.
	var e Engine
	e.Activate(stubPilot{
		step:     zoomingBy(4).step,
		schedule: Schedule{Transitions: []string{"fade"}, Traversal: Single},
	})

	_, arrival := e.Tick(View{Zoom: MaxZoom / 2, MaxIter: 50})
	if arrival == nil || *arrival != TopBoundary {
		t.Fatalf("arrival = %v, want TopBoundary", arrival)
	}
}

func pow2(n float64) float64 {
	p := 1.0
	for range int(n) {
		p *= 2
	}
	return p
}
