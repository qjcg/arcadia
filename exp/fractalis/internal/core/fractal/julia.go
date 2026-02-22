package fractal

// Julia calculates the Julia set for a given point z and constant c
// Uses smooth coloring for better visual quality
// Returns float64 for smooth iteration count interpolation
func Julia(zr, zi, juliaCr, juliaCi float64, maxIter int) float64 {
	calc := NewIterationCalculator(maxIter)
	return calc.JuliaOptimized(zr, zi, juliaCr, juliaCi)
}
