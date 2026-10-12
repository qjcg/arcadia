package transition

import (
	"math"
	"math/rand"
	"testing"

	"github.com/qjcg/arcadia/exp/fractalis/internal/core/pilot"
)

// stubPilot is a scripted pilot used to exercise the engine seam.
type stubPilot struct {
	step     func(pilot.View) pilot.View
	schedule pilot.Schedule
}

func (p stubPilot) Step(v pilot.View) pilot.View {
	if p.step == nil {
		return v
	}
	return p.step(v)
}

func (p stubPilot) TransitionSchedule() pilot.Schedule { return p.schedule }

// zoomingOut returns a stub pilot whose policy multiplies zoom by
// rate each tick, driving toward the bottom boundary.
func zoomingOut(rate float64) stubPilot {
	return stubPilot{step: func(v pilot.View) pilot.View {
		v.Zoom *= rate
		return v
	}}
}

// scriptedSource is an injected rand.Source whose draws are fixed, so
// traversal and destination picks are deterministic under test.
// rand.Rand reads Int31 from the high bits of Int63, so draws are
// shifted up.
type scriptedSource struct {
	draws []int64
	i     int
}

func (s *scriptedSource) Int63() int64 {
	if len(s.draws) == 0 {
		return 0
	}
	v := s.draws[s.i%len(s.draws)]
	s.i++
	return v << 32
}

func (s *scriptedSource) Seed(int64) {}

func scriptedRand(draws ...int64) *rand.Rand {
	return rand.New(&scriptedSource{draws: draws})
}

func fadeSchedule() pilot.Schedule {
	return pilot.Schedule{Transitions: []string{FadeID}, Traversal: pilot.Single}
}

// atBottom returns a view resting one tick from the bottom boundary
// under the given zoom rate.
func atBottom() pilot.View {
	return pilot.View{FractalType: "mandelbrot", Zoom: pilot.MinZoom, MaxIter: 50}
}

// collectStarts ticks the seam until it has started want crossings
// and returns the effect picked for each.
func collectStarts(t *testing.T, r *Runner, v pilot.View, want int) []string {
	t.Helper()
	var effects []string
	for ticks := 0; len(effects) < want; ticks++ {
		if ticks > 1000 {
			t.Fatalf("seam started %d crossings, want %d", len(effects), want)
		}
		var c *Crossing
		v, c = r.Tick(v)
		if c != nil && c.Progress == Start {
			effects = append(effects, c.Effect)
		}
	}
	return effects
}

func TestEngineSeamFiresATransitionOnActiveBoundaryArrival(t *testing.T) {
	r := NewRunner(pilot.NewScheduler(scriptedRand()), pilot.NewRandomDestination(nil, scriptedRand()))
	r.Activate(zoomingOut(0.5))
	r.SetSchedule(fadeSchedule())

	v, c := r.Tick(atBottom())
	if c == nil {
		t.Fatal("boundary arrival while actively moving started no transition")
	}
	if c.Effect != FadeID {
		t.Errorf("crossing effect = %q, want %q", c.Effect, FadeID)
	}
	if c.Source.Zoom != pilot.MinZoom {
		t.Errorf("crossing source Zoom = %v, want clamped to %v", c.Source.Zoom, pilot.MinZoom)
	}
	if v != c.Source {
		t.Errorf("view during the crossing = %+v, want the frozen source %+v", v, c.Source)
	}
	if c.Progress != Start {
		t.Errorf("crossing progress = %v, want %v at the first frame", c.Progress, Start)
	}
}

func TestEngineSeamLeavesARestingBoundaryViewAlone(t *testing.T) {
	r := NewRunner(pilot.NewScheduler(scriptedRand()), pilot.NewRandomDestination(nil, scriptedRand()))
	r.SetSchedule(fadeSchedule())

	// The Manual pilot at rest: no transition, no clamp-and-move.
	v, c := r.Tick(atBottom())
	if c != nil {
		t.Fatalf("resting view started a transition: %+v", c)
	}
	if v.Zoom != pilot.MinZoom {
		t.Fatalf("resting view moved: Zoom = %v, want %v", v.Zoom, pilot.MinZoom)
	}
}

func TestEmptyScheduleClampsAtTheBoundary(t *testing.T) {
	r := NewRunner(pilot.NewScheduler(scriptedRand()), pilot.NewRandomDestination(nil, scriptedRand()))
	// The pilot declares a schedule, but the effective schedule is
	// empty: boundary arrival clamps.
	r.Activate(stubPilot{step: zoomingOut(0.5).step, schedule: fadeSchedule()})
	r.SetSchedule(pilot.Schedule{})

	v, c := r.Tick(atBottom())
	if c != nil {
		t.Fatalf("empty schedule started a transition: %+v", c)
	}
	if v.Zoom != pilot.MinZoom {
		t.Fatalf("Zoom = %v, want clamped to %v", v.Zoom, pilot.MinZoom)
	}
}

func TestCrossingPlaysOverTheSharedProgressModel(t *testing.T) {
	r := NewRunner(pilot.NewScheduler(scriptedRand()),
		pilot.NewRandomDestination([]string{"julia"}, scriptedRand(0)))
	r.Activate(zoomingOut(0.5))
	r.SetSchedule(fadeSchedule())

	v, c := r.Tick(atBottom())
	source := v
	for i := 0; !c.Progress.Landed(); i++ {
		want := Start + Step*Progress(i)
		if math.Abs(float64(c.Progress-want)) > 1e-9 {
			t.Fatalf("tick %d: progress = %v, want %v", i, c.Progress, want)
		}
		if v != source {
			t.Fatalf("tick %d: view moved mid-crossing: %+v", i, v)
		}
		v, c = r.Tick(v)
	}

	if v != c.Destination {
		t.Fatalf("landing view = %+v, want the destination %+v", v, c.Destination)
	}
	if v.FractalType != "julia" {
		t.Fatalf("landing FractalType = %q, want %q", v.FractalType, "julia")
	}
}

func TestSingleTraversalPicksTheSameEffectEveryCrossing(t *testing.T) {
	r := NewRunner(pilot.NewScheduler(scriptedRand()), pilot.NewRandomDestination(nil, scriptedRand()))
	r.Activate(zoomingOut(0.05))
	r.SetSchedule(fadeSchedule())

	got := collectStarts(t, r, atBottom(), 3)
	for i, effect := range got {
		if effect != FadeID {
			t.Fatalf("crossing %d: effect = %q, want %q", i+1, effect, FadeID)
		}
	}
}

func TestSequentialTraversalStepsThroughTheSchedule(t *testing.T) {
	r := NewRunner(pilot.NewScheduler(scriptedRand()), pilot.NewRandomDestination(nil, scriptedRand()))
	r.Activate(zoomingOut(0.05))
	r.SetSchedule(pilot.Schedule{
		Transitions: []string{FadeID, RotateID, ZoomOutID},
		Traversal:   pilot.Sequential,
	})

	got := collectStarts(t, r, atBottom(), 4)
	want := []string{FadeID, RotateID, ZoomOutID, FadeID}
	for i, effect := range want {
		if got[i] != effect {
			t.Fatalf("crossing %d: effect = %q, want %q", i+1, got[i], effect)
		}
	}
}

func TestRandomTraversalRerollsPerCrossing(t *testing.T) {
	r := NewRunner(pilot.NewScheduler(scriptedRand(0, 2, 1)), pilot.NewRandomDestination(nil, scriptedRand()))
	r.Activate(zoomingOut(0.05))
	r.SetSchedule(pilot.Schedule{
		Transitions: []string{FadeID, RotateID, ZoomOutID},
		Traversal:   pilot.Random,
	})

	got := collectStarts(t, r, atBottom(), 3)
	want := []string{FadeID, ZoomOutID, RotateID}
	for i, effect := range want {
		if got[i] != effect {
			t.Fatalf("crossing %d: effect = %q, want %q", i+1, got[i], effect)
		}
	}
}

func TestDestinationIsARandomFractalTypeDifferentFromTheCurrentOne(t *testing.T) {
	r := NewRunner(pilot.NewScheduler(scriptedRand()),
		pilot.NewRandomDestination([]string{"mandelbrot", "julia", "burningship"}, scriptedRand(1)))
	r.Activate(zoomingOut(0.5))
	r.SetSchedule(fadeSchedule())

	_, c := r.Tick(atBottom())
	for !c.Progress.Landed() {
		_, c = r.Tick(atBottom())
	}

	if c.Destination.FractalType == c.Source.FractalType {
		t.Fatalf("destination %q equals the source fractal type", c.Destination.FractalType)
	}
	if c.Destination.FractalType != "burningship" {
		t.Fatalf("destination FractalType = %q, want the injected draw %q", c.Destination.FractalType, "burningship")
	}
}
