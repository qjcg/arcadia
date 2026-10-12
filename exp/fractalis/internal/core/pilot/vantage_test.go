package pilot

import "testing"

// viewpointAt returns a curated viewpoint at the given center.
func viewpointAt(name string, x, y, zoom float64) Viewpoint {
	return Viewpoint{
		Name: name,
		View: View{
			FractalType: "mandelbrot",
			CenterX:     x,
			CenterY:     y,
			Zoom:        zoom,
			MaxIter:     100,
		},
	}
}

// tickN runs the engine seam n ticks and records the resulting view
// after each tick.
func tickN(e *Engine, v View, n int) []View {
	views := make([]View, 0, n)
	for range n {
		v, _ = e.Tick(v)
		views = append(views, v)
	}
	return views
}

func TestVantageSwitchesScenesOnItsSceneTimer(t *testing.T) {
	vantage := NewVantage(VantageOptions{
		SceneDuration: 3,
		Viewpoints: []Viewpoint{
			viewpointAt("first", 1, 1, 10),
			viewpointAt("second", 2, 2, 20),
			viewpointAt("third", 3, 3, 30),
		},
	})

	var e Engine
	e.Activate(vantage)

	// Each scene rests exactly SceneDuration ticks before the pilot
	// moves the view to the next one.
	views := tickN(&e, View{FractalType: "mandelbrot", Zoom: 1, MaxIter: 50}, 9)

	want := []View{
		{FractalType: "mandelbrot", CenterX: 1, CenterY: 1, Zoom: 10, MaxIter: 100},
		{FractalType: "mandelbrot", CenterX: 1, CenterY: 1, Zoom: 10, MaxIter: 100},
		{FractalType: "mandelbrot", CenterX: 1, CenterY: 1, Zoom: 10, MaxIter: 100},
		{FractalType: "mandelbrot", CenterX: 2, CenterY: 2, Zoom: 20, MaxIter: 100},
		{FractalType: "mandelbrot", CenterX: 2, CenterY: 2, Zoom: 20, MaxIter: 100},
		{FractalType: "mandelbrot", CenterX: 2, CenterY: 2, Zoom: 20, MaxIter: 100},
		{FractalType: "mandelbrot", CenterX: 3, CenterY: 3, Zoom: 30, MaxIter: 100},
		{FractalType: "mandelbrot", CenterX: 3, CenterY: 3, Zoom: 30, MaxIter: 100},
		{FractalType: "mandelbrot", CenterX: 3, CenterY: 3, Zoom: 30, MaxIter: 100},
	}
	for i, got := range views {
		if got != want[i] {
			t.Fatalf("tick %d: view = %+v, want %+v", i+1, got, want[i])
		}
	}
}

func TestVantageSceneDurationParameter(t *testing.T) {
	vantage := NewVantage(VantageOptions{
		SceneDuration: 2,
		Viewpoints: []Viewpoint{
			viewpointAt("first", 1, 1, 10),
			viewpointAt("second", 2, 2, 20),
		},
	})

	if got := vantage.SceneDuration(); got != 2 {
		t.Fatalf("SceneDuration() = %d, want 2", got)
	}

	// The parameter retimes the tour live: after two ticks on the
	// first scene, a longer scene duration stretches the second one.
	var e Engine
	e.Activate(vantage)
	v := View{FractalType: "mandelbrot", Zoom: 1, MaxIter: 50}
	v, _ = e.Tick(v)
	v, _ = e.Tick(v)
	vantage.SetSceneDuration(5)

	views := tickN(&e, v, 7)
	want := []View{
		{FractalType: "mandelbrot", CenterX: 2, CenterY: 2, Zoom: 20, MaxIter: 100},
		{FractalType: "mandelbrot", CenterX: 2, CenterY: 2, Zoom: 20, MaxIter: 100},
		{FractalType: "mandelbrot", CenterX: 2, CenterY: 2, Zoom: 20, MaxIter: 100},
		{FractalType: "mandelbrot", CenterX: 2, CenterY: 2, Zoom: 20, MaxIter: 100},
		{FractalType: "mandelbrot", CenterX: 2, CenterY: 2, Zoom: 20, MaxIter: 100},
		{FractalType: "mandelbrot", CenterX: 1, CenterY: 1, Zoom: 10, MaxIter: 100},
		{FractalType: "mandelbrot", CenterX: 1, CenterY: 1, Zoom: 10, MaxIter: 100},
	}
	for i, got := range views {
		if got != want[i] {
			t.Fatalf("tick %d: view = %+v, want %+v", i+1, got, want[i])
		}
	}
}

func TestVantageNextSceneSkipsAhead(t *testing.T) {
	vantage := NewVantage(VantageOptions{
		SceneDuration: 10,
		Viewpoints: []Viewpoint{
			viewpointAt("first", 1, 1, 10),
			viewpointAt("second", 2, 2, 20),
		},
	})

	var e Engine
	e.Activate(vantage)
	v := View{FractalType: "mandelbrot", Zoom: 1, MaxIter: 50}
	v, _ = e.Tick(v)

	vantage.NextScene()
	v, _ = e.Tick(v)
	want := View{FractalType: "mandelbrot", CenterX: 2, CenterY: 2, Zoom: 20, MaxIter: 100}
	if v != want {
		t.Fatalf("view after NextScene = %+v, want %+v", v, want)
	}
}

func TestVantageToursCuratedViewpoints(t *testing.T) {
	// With no tour injected, the pilot runs its curated viewpoints.
	vantage := NewVantage(VantageOptions{SceneDuration: 2})

	var e Engine
	e.Activate(vantage)

	curated := CuratedViewpoints()
	if len(curated) < 2 {
		t.Fatalf("CuratedViewpoints() = %d viewpoints, want a tour of at least 2", len(curated))
	}

	// Two ticks per scene, one tick past the end of the tour to
	// assert it wraps around.
	views := tickN(&e, View{FractalType: "mandelbrot", Zoom: 1, MaxIter: 50}, 2*len(curated)+1)

	for i, got := range views {
		want := curated[i/2%len(curated)].View
		if got != want {
			t.Fatalf("tick %d: view = %+v, want curated viewpoint %q %+v", i+1, got, curated[i/2%len(curated)].Name, want)
		}
	}
}

func TestVantagePansTowardDiscoveredPoints(t *testing.T) {
	// The pilot pans each scene slowly toward a discovered
	// interesting point, one pan per scene.
	discovered := 0
	vantage := NewVantage(VantageOptions{
		SceneDuration: 2,
		Viewpoints: []Viewpoint{
			viewpointAt("first", 0, 0, 10),
			viewpointAt("second", 0, 0, 20),
		},
		Discover: func(View) (float64, float64) {
			discovered++
			return 1, 1
		},
	})

	var e Engine
	e.Activate(vantage)
	views := tickN(&e, View{FractalType: "mandelbrot", Zoom: 1, MaxIter: 50}, 4)

	if discovered != 2 {
		t.Fatalf("Discover called %d times, want once per scene", discovered)
	}
	if views[0].CenterX != 0 || views[0].CenterY != 0 {
		t.Fatalf("scene tick moved the view to %+v, want the scene itself", views[0])
	}
	if views[1].CenterX <= 0 || views[1].CenterY <= 0 {
		t.Fatalf("pan tick did not move toward the discovered point: %+v", views[1])
	}
	if views[2].CenterX != 0 || views[3].CenterX <= views[1].CenterX {
		t.Fatalf("new scene did not restart its own pan: %+v, %+v", views[2], views[3])
	}
}

func TestVantageScheduleIsSingleFade(t *testing.T) {
	vantage := NewVantage(VantageOptions{
		Viewpoints: []Viewpoint{viewpointAt("only", 0, 0, 1)},
	})

	got := vantage.TransitionSchedule()
	want := Schedule{Transitions: []string{"fade"}, Traversal: Single}
	if len(got.Transitions) != 1 || got.Transitions[0] != want.Transitions[0] || got.Traversal != want.Traversal {
		t.Fatalf("TransitionSchedule() = %+v, want %+v", got, want)
	}
}
