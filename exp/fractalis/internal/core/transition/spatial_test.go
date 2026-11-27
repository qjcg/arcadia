package transition

import (
	"math/rand"
	"testing"

	"github.com/qjcg/arcadia/exp/fractalis/internal/core/pilot"
)

type fixedCameraDestination struct{ view pilot.View }

func (d fixedCameraDestination) Choose(pilot.View) pilot.View { return d.view }

func TestSpatialAutoZoomCrossesAndLands(t *testing.T) {
	src := pilot.View{FractalType: "box", Zoom: 1.55, Camera: pilot.Camera{Spatial: true, X: 1}}
	dst := pilot.View{FractalType: "bulb", Zoom: 0.2, Camera: pilot.Camera{Spatial: true, Z: 5}}
	r := NewRunner(pilot.NewScheduler(rand.New(rand.NewSource(1))), fixedCameraDestination{dst})
	r.Activate(pilot.NewAutoZoom(nil))
	r.SetSchedule(pilot.Schedule{Transitions: []string{RotateID}})
	next, c := r.Tick(src)
	if c == nil || c.Effect != RotateID || next.Zoom != 1.6 {
		t.Fatalf("arrival = %+v, crossing = %+v", next, c)
	}
	if frame := (Rotate{}).Frame(c.Source, c.Destination, 0.25); frame.Source.Camera == c.Source.Camera {
		t.Fatal("rotate did not move spatial camera")
	}
	for i := 0; i < 30 && !c.Progress.Landed(); i++ {
		next, c = r.Tick(next)
	}
	if next != dst || c == nil || !c.Progress.Landed() {
		t.Fatalf("landing = %+v, crossing = %+v", next, c)
	}
}
