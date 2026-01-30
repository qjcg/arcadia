package fractals

import "testing"

func TestBurningShip(t *testing.T) {
	tests := []struct {
		name      string
		cr, ci    float64
		maxIter   int
		expected  float64
		tolerance int
	}{
		{"Origin should be in set", 0.0, 0.0, 100, 100, 0},
		{"Far point diverges", 10.0, 10.0, 100, 1, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BurningShip(tt.cr, tt.ci, tt.maxIter)
			if int(result) < int(tt.expected)-tt.tolerance || int(result) > int(tt.expected)+tt.tolerance {
				t.Errorf("BurningShip(%f, %f, %d) = %f, want ~%f", tt.cr, tt.ci, tt.maxIter, result, tt.expected)
			}
		})
	}
}
