package fractal

import "testing"

func TestMandelbrot(t *testing.T) {
	tests := []struct {
		name      string
		cr, ci    float64
		maxIter   int
		expected  float64
		tolerance int
	}{
		{"Origin (0,0) is in set", 0.0, 0.0, 100, 100, 0},
		{"Point (-1,0) is in set", -1.0, 0.0, 100, 100, 0},
		{"Point (0.25,0) is in set", 0.25, 0.0, 100, 100, 0},
		{"Distant point (2,0) diverges quickly", 2.0, 0.0, 100, 2, 1},
		{"Distant point (0,2) diverges quickly", 0.0, 2.0, 100, 2, 1},
		{"Point (-2,-2) diverges", -2.0, -2.0, 100, 1, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Mandelbrot(tt.cr, tt.ci, tt.maxIter)
			if int(result) < int(tt.expected)-tt.tolerance || int(result) > int(tt.expected)+tt.tolerance {
				t.Errorf("Mandelbrot(%f, %f, %d) = %f, want ~%f", tt.cr, tt.ci, tt.maxIter, result, tt.expected)
			}
		})
	}
}

func TestMandelbrotQuickDivergence(t *testing.T) {
	tests := []struct {
		name    string
		cr, ci  float64
		maxIter int
	}{
		{"Far right", 10.0, 0.0, 100},
		{"Far left", -10.0, 0.0, 100},
		{"Far up", 0.0, 10.0, 100},
		{"Far down", 0.0, -10.0, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Mandelbrot(tt.cr, tt.ci, tt.maxIter)
			if int(result) > 5 {
				t.Errorf("Mandelbrot(%f, %f, %d) = %f, expected to diverge quickly (<= 5)", tt.cr, tt.ci, tt.maxIter, result)
			}
		})
	}
}
