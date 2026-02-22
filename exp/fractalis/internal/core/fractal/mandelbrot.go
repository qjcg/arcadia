package fractal

// Mandelbrot calculates the number of iterations for a given complex number c
// Uses smooth coloring for better visual quality
// Returns float64 for smooth iteration count interpolation
func Mandelbrot(cr, ci float64, maxIter int) float64 {
	calc := NewIterationCalculator(maxIter)
	return calc.MandelbrotOptimized(cr, ci)
}
