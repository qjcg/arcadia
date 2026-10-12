package transition

import (
	"testing"

	"github.com/qjcg/arcadia/exp/fractalis/internal/core/pilot"
)

var (
	srcView = pilot.View{
		FractalType: "mandelbrot",
		CenterX:     -0.5,
		CenterY:     0.0,
		Zoom:        0.1,
		MaxIter:     50,
	}
	dstView = pilot.View{
		FractalType: "julia",
		CenterX:     0.0,
		CenterY:     0.0,
		Zoom:        1.0,
		MaxIter:     50,
	}
)

func TestFadeRendersAcrossTheSharedProgressModel(t *testing.T) {
	var e Fade

	for _, p := range []Progress{Start, 0.5, Done} {
		frame := e.Frame(srcView, dstView, p)
		if frame.Source != srcView {
			t.Errorf("frame at %v: Source = %+v, want %+v", p, frame.Source, srcView)
		}
		if frame.Destination != dstView {
			t.Errorf("frame at %v: Destination = %+v, want %+v", p, frame.Destination, dstView)
		}
		if frame.Mix != float64(p) {
			t.Errorf("frame at %v: Mix = %v, want %v", p, frame.Mix, p)
		}
	}
}

func TestFadeStartsOnTheSourceAndEndsOnTheDestination(t *testing.T) {
	var e Fade

	if got := e.Frame(srcView, dstView, Start).Mix; got >= 0.5 {
		t.Errorf("first frame Mix = %v, want the source view dominant", got)
	}
	if got := e.Frame(srcView, dstView, Done).Mix; got != 1 {
		t.Errorf("last frame Mix = %v, want 1 (the destination view)", got)
	}
}

func TestRegistryRegistersAndLooksUpEffects(t *testing.T) {
	r := NewRegistry()
	r.Register(FadeID, Fade{})

	got, ok := r.Lookup(FadeID)
	if !ok {
		t.Fatalf("Lookup(%q) not found, want the registered effect", FadeID)
	}
	if frame := got.Frame(srcView, dstView, Done); frame.Mix != 1 {
		t.Errorf("registered effect Mix = %v, want 1", frame.Mix)
	}
}

func TestRegistryLookupUnknownEffect(t *testing.T) {
	r := NewRegistry()

	if got, ok := r.Lookup(StarWipeID); ok {
		t.Errorf("Lookup(%q) = %v, want not found before its ticket registers it", StarWipeID, got)
	}
}
