package pilot

import "testing"

func TestDefaultRegistryRegistersManual(t *testing.T) {
	r := NewRegistry()

	got, ok := r.Lookup(ManualID)
	if !ok {
		t.Fatalf("Lookup(%q) not found, want the Manual pilot registered by default", ManualID)
	}
	if !got.TransitionSchedule().Empty() {
		t.Errorf("registered Manual pilot schedule = %+v, want empty", got.TransitionSchedule())
	}
}

func TestRegisterAndLookup(t *testing.T) {
	r := NewRegistry()

	r.Register(AutoZoomID, zoomingBy(2))
	got, ok := r.Lookup(AutoZoomID)
	if !ok {
		t.Fatalf("Lookup(%q) not found, want the registered pilot", AutoZoomID)
	}

	var e Engine
	e.Activate(got)
	v, _ := e.Tick(View{Zoom: 1, MaxIter: 50})
	if v.Zoom != 2 {
		t.Errorf("registered pilot did not drive the view: Zoom = %v, want 2", v.Zoom)
	}
}

func TestLookupUnknownPilot(t *testing.T) {
	r := NewRegistry()

	if got, ok := r.Lookup(VantageID); ok {
		t.Errorf("Lookup(%q) = %+v, want not found before its ticket registers it", VantageID, got)
	}
}
