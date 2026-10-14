package pilot

import (
	"math"
	"testing"
)

// stubSeeker is a scripted interesting-region service for exercising
// the Auto-zoom pilot through the engine seam.
type stubSeeker struct {
	targetX float64
	targetY float64
	uniform bool
	finds   int
}

func (s *stubSeeker) FindInterestingPoint(View) (float64, float64) {
	s.finds++
	return s.targetX, s.targetY
}

func (s *stubSeeker) IsViewUniform(View) bool { return s.uniform }

func TestAutoZoomDrivesZoomPacing(t *testing.T) {
	var e Engine
	e.Activate(NewAutoZoom(&stubSeeker{}))

	// The default journey zooms in at 1.05x per tick.
	v := View{FractalType: "mandelbrot", CenterX: -0.5, Zoom: 1, MaxIter: 50}
	for tick := 1; tick <= 3; tick++ {
		next, arrival := e.Tick(v)
		want := math.Pow(1.05, float64(tick))
		if math.Abs(next.Zoom-want) > 1e-12 {
			t.Fatalf("tick %d: Zoom = %v, want %v", tick, next.Zoom, want)
		}
		if arrival != nil {
			t.Fatalf("tick %d: boundary arrival = %v, want none", tick, *arrival)
		}
		v = next
	}
}

func TestAutoZoomDirectionAndSpeedSteerMotion(t *testing.T) {
	t.Run("zooming out divides by speed each tick", func(t *testing.T) {
		az := NewAutoZoom(&stubSeeker{})
		az.SetDirection(DirectionOut)

		var e Engine
		e.Activate(az)

		v := View{FractalType: "mandelbrot", Zoom: 8, MaxIter: 50}
		for tick := 1; tick <= 2; tick++ {
			v, _ = e.Tick(v)
			want := 8 / math.Pow(1.05, float64(tick))
			if math.Abs(v.Zoom-want) > 1e-12 {
				t.Fatalf("tick %d: Zoom = %v, want %v", tick, v.Zoom, want)
			}
		}
	})

	t.Run("reversing direction flips the zoom motion", func(t *testing.T) {
		az := NewAutoZoom(&stubSeeker{})
		az.ReverseDirection()
		if got := az.Direction(); got != DirectionOut {
			t.Fatalf("ReverseDirection() left direction = %v, want out", got)
		}

		var e Engine
		e.Activate(az)

		v := View{FractalType: "mandelbrot", Zoom: 8, MaxIter: 50}
		next, _ := e.Tick(v)
		if next.Zoom >= v.Zoom {
			t.Fatalf("Zoom = %v, want less than %v after reversing", next.Zoom, v.Zoom)
		}

		az.ReverseDirection()
		next, _ = e.Tick(next)
		if next.Zoom <= 8/1.05 {
			t.Fatalf("Zoom = %v, want more than %v after reversing back", next.Zoom, 8/1.05)
		}
	})

	t.Run("speed sets the zoom pacing", func(t *testing.T) {
		az := NewAutoZoom(&stubSeeker{})
		az.SetSpeed(1.5)

		var e Engine
		e.Activate(az)

		v := View{FractalType: "mandelbrot", Zoom: 1, MaxIter: 50}
		next, _ := e.Tick(v)
		if next.Zoom != 1.5 {
			t.Fatalf("Zoom = %v, want 1.5 at speed 1.5", next.Zoom)
		}
	})
}

func TestAutoZoomSeeksInterestingRegions(t *testing.T) {
	t.Run("pans toward the region the seeker finds", func(t *testing.T) {
		seeker := &stubSeeker{targetX: 1, targetY: -1}
		var e Engine
		e.Activate(NewAutoZoom(seeker))

		v := View{FractalType: "mandelbrot", CenterX: 0, CenterY: 0, Zoom: 1, MaxIter: 50}
		next, _ := e.Tick(v)
		if next.CenterX != 0.05 || next.CenterY != -0.05 {
			t.Fatalf("center = (%v, %v), want (0.05, -0.05) panning toward the region",
				next.CenterX, next.CenterY)
		}
		if seeker.finds != 1 {
			t.Fatalf("FindInterestingPoint called %d times, want 1", seeker.finds)
		}
	})

	t.Run("seeks a fresh region once the target is reached", func(t *testing.T) {
		seeker := &stubSeeker{targetX: 0, targetY: 0}
		var e Engine
		e.Activate(NewAutoZoom(seeker))

		v := View{FractalType: "mandelbrot", CenterX: 0, CenterY: 0, Zoom: 1, MaxIter: 50}
		for range 3 {
			v, _ = e.Tick(v)
		}
		if seeker.finds < 2 {
			t.Fatalf("FindInterestingPoint called %d times, want a fresh region once reached", seeker.finds)
		}
	})

	t.Run("seeks a fresh region when the view goes uniform", func(t *testing.T) {
		seeker := &stubSeeker{targetX: 1, targetY: 1, uniform: true}
		var e Engine
		e.Activate(NewAutoZoom(seeker))

		v := View{FractalType: "mandelbrot", Zoom: 1, MaxIter: 50}
		for range 50 {
			v, _ = e.Tick(v)
		}
		if seeker.finds < 2 {
			t.Fatalf("FindInterestingPoint called %d times, want a fresh region on a uniform view", seeker.finds)
		}
	})

	t.Run("seeks a fresh region after crossing into a new fractal", func(t *testing.T) {
		seeker := &stubSeeker{targetX: 1, targetY: 1}
		var e Engine
		e.Activate(NewAutoZoom(seeker))

		v := View{FractalType: "mandelbrot", Zoom: 1, MaxIter: 50}
		v, _ = e.Tick(v)
		if seeker.finds != 1 {
			t.Fatalf("FindInterestingPoint called %d times, want 1", seeker.finds)
		}

		v.FractalType = "julia"
		v, _ = e.Tick(v)
		if seeker.finds != 2 {
			t.Fatalf("FindInterestingPoint called %d times, want a fresh region on the new fractal", seeker.finds)
		}
	})
}

func TestSwitchingPilotsReplacesTheActiveOne(t *testing.T) {
	var e Engine

	e.Activate(NewAutoZoom(&stubSeeker{}))
	moving, _ := e.Tick(View{FractalType: "mandelbrot", Zoom: 1, MaxIter: 50})
	if moving.Zoom == 1 {
		t.Fatal("Auto-zoom pilot did not move the view")
	}

	e.Activate(Manual())
	resting, _ := e.Tick(moving)
	if resting != moving {
		t.Fatalf("Manual pilot moved the view: got %+v, want %+v", resting, moving)
	}

	e.Activate(NewAutoZoom(&stubSeeker{}))
	moving, _ = e.Tick(resting)
	if moving.Zoom == resting.Zoom {
		t.Fatal("Auto-zoom pilot did not replace the Manual pilot")
	}
}
