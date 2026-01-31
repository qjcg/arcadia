package fractal

import "testing"

func TestPerpendicular(t *testing.T) {
	tests := []struct {
		name     string
		cr, ci   float64
		maxIter  int
		expected int
	}{
		{"Origin should be in set", 0.0, 0.0, 100, 100},
		{"Far point diverges", 10.0, 10.0, 100, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Perpendicular(tt.cr, tt.ci, tt.maxIter)
			if result != tt.expected {
				t.Errorf("Perpendicular(%f, %f, %d) = %d, want %d", tt.cr, tt.ci, tt.maxIter, result, tt.expected)
			}
		})
	}
}
