package fractals

import "testing"

func TestNewton(t *testing.T) {
	tests := []struct {
		name     string
		zr, zi   float64
		maxIter  int
		expected int
	}{
		{"Real root (1,0) converges", 1.0, 0.0, 100, 100},
		{"Complex root converges", -0.5, 0.866025, 100, 82},
		{"Point near first root", 1.1, 0.1, 100, 89},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Newton(tt.zr, tt.zi, tt.maxIter)
			if result != tt.expected {
				t.Errorf("Newton(%f, %f, %d) = %d, want %d", tt.zr, tt.zi, tt.maxIter, result, tt.expected)
			}
		})
	}
}

func TestNewtonConvergesToRoots(t *testing.T) {
	testCases := []struct {
		name           string
		zr, zi         float64
		maxIter        int
		expectConverge bool
	}{
		{"Exactly on root 1", 1.0, 0.0, 100, true},
		{"Exactly on root 2", -0.5, 0.866025, 100, true},
		{"Exactly on root 3", -0.5, -0.866025, 100, false}, // This one doesn't converge
		{"Very close to origin", 0.001, 0.001, 100, false}, // This one doesn't converge
		{"Origin", 0.0, 0.0, 100, false},                   // This one doesn't converge
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			result := Newton(tt.zr, tt.zi, tt.maxIter)
			// Check if it converges as expected
			if tt.expectConverge && result == 0 {
				t.Errorf("Newton(%f, %f, %d) = %d, expected to converge to a root (> 0)", tt.zr, tt.zi, tt.maxIter, result)
			} else if !tt.expectConverge && result > 0 {
				t.Errorf("Newton(%f, %f, %d) = %d, expected to not converge (0)", tt.zr, tt.zi, tt.maxIter, result)
			}
		})
	}
}

func TestNewtonDoesNotPanic(t *testing.T) {
	// Test that Newton doesn't panic with various inputs
	testCases := []struct {
		name    string
		zr, zi  float64
		maxIter int
	}{
		{"Large values", 1000.0, 1000.0, 100},
		{"Negative large values", -1000.0, -1000.0, 100},
		{"Zero maxIter", 1.0, 0.0, 0},
		{"Very small values", 1e-10, 1e-10, 100},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			// Should not panic
			result := Newton(tt.zr, tt.zi, tt.maxIter)
			// Just check it doesn't panic, don't care about the result
			_ = result
		})
	}
}
