package fractals

import "testing"

func TestMandelbrot(t *testing.T) {
	tests := []struct {
		name     string
		cr, ci   float64
		maxIter  int
		expected int
	}{
		{"Origin (0,0) is in set", 0.0, 0.0, 100, 100},
		{"Point (-1,0) is in set", -1.0, 0.0, 100, 100},
		{"Point (0.25,0) is in set", 0.25, 0.0, 100, 100},
		{"Distant point (2,0) diverges quickly", 2.0, 0.0, 100, 2},
		{"Distant point (0,2) diverges quickly", 0.0, 2.0, 100, 2},
		{"Point (-2,-2) diverges", -2.0, -2.0, 100, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Mandelbrot(tt.cr, tt.ci, tt.maxIter)
			if result != tt.expected {
				t.Errorf("Mandelbrot(%f, %f, %d) = %d, want %d", tt.cr, tt.ci, tt.maxIter, result, tt.expected)
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
			// Should diverge quickly (within first few iterations)
			if result > 5 {
				t.Errorf("Mandelbrot(%f, %f, %d) = %d, expected to diverge quickly (<= 5)", tt.cr, tt.ci, tt.maxIter, result)
			}
		})
	}
}
