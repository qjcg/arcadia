package fractals

import "testing"

func TestJulia(t *testing.T) {
	tests := []struct {
		name      string
		zr, zi    float64
		juliaCr   float64
		juliaCi   float64
		maxIter   int
		expected  float64
		tolerance int
	}{
		{"Classic Julia set c=-0.7+0.27015i, origin is in set", 0.0, 0.0, -0.7, 0.27015, 100, 96, 5},
		{"Distant point diverges quickly", 10.0, 10.0, -0.7, 0.27015, 100, 0, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Julia(tt.zr, tt.zi, tt.juliaCr, tt.juliaCi, tt.maxIter)
			if int(result) < int(tt.expected)-tt.tolerance || int(result) > int(tt.expected)+tt.tolerance {
				t.Errorf("Julia(%f, %f, %f, %f, %d) = %f, want ~%f", tt.zr, tt.zi, tt.juliaCr, tt.juliaCi, tt.maxIter, result, tt.expected)
			}
		})
	}
}
