package pilot

import "testing"

func TestRandomDestinationPicksADifferentFractalType(t *testing.T) {
	d := NewRandomDestination([]string{"mandelbrot", "julia", "celtic"}, scriptedRand(1))

	got := d.Choose(View{FractalType: "mandelbrot", Zoom: 1e15, MaxIter: 50})
	if got.FractalType != "celtic" {
		t.Fatalf("Choose() FractalType = %q, want %q", got.FractalType, "celtic")
	}
}

func TestRandomDestinationNeverReturnsTheCurrentType(t *testing.T) {
	types := []string{"tricorn", "julia", "burningship"}
	for draw := range int64(len(types)) {
		d := NewRandomDestination(types, scriptedRand(draw))
		got := d.Choose(View{FractalType: "tricorn", MaxIter: 50})
		if got.FractalType == "tricorn" {
			t.Fatalf("draw %d: Choose() returned the current fractal type", draw)
		}
	}
}

func TestRandomDestinationArrivesAtTheDefaultViewOfTheType(t *testing.T) {
	tests := []struct {
		name       string
		draw       int64
		wantType   string
		wantCenter [2]float64
	}{
		{"julia defaults to the origin", 1, "julia", [2]float64{0, 0}},
		{"burningship keeps its classic window", 2, "burningship", [2]float64{-0.5, -0.6}},
		{"others center on the mandelbrot view", 0, "tricorn", [2]float64{-0.5, 0}},
	}
	types := []string{"tricorn", "julia", "burningship"}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewRandomDestination(types, scriptedRand(tt.draw))
			got := d.Choose(View{FractalType: "mandelbrot", Zoom: 0.1, MaxIter: 200})

			if got.FractalType != tt.wantType {
				t.Errorf("FractalType = %q, want %q", got.FractalType, tt.wantType)
			}
			if got.CenterX != tt.wantCenter[0] || got.CenterY != tt.wantCenter[1] {
				t.Errorf("Center = (%v, %v), want (%v, %v)",
					got.CenterX, got.CenterY, tt.wantCenter[0], tt.wantCenter[1])
			}
			if got.Zoom != 1.0 {
				t.Errorf("Zoom = %v, want 1.0 at the destination", got.Zoom)
			}
			if got.MaxIter != 200 {
				t.Errorf("MaxIter = %v, want the current 200 carried over", got.MaxIter)
			}
		})
	}
}

func TestRandomDestinationWithOnlyTheCurrentTypeStaysPut(t *testing.T) {
	d := NewRandomDestination([]string{"mandelbrot"}, scriptedRand(0))

	got := d.Choose(View{FractalType: "mandelbrot", MaxIter: 50})
	if got.FractalType != "mandelbrot" {
		t.Fatalf("Choose() FractalType = %q, want no alternative to %q", got.FractalType, "mandelbrot")
	}
}
