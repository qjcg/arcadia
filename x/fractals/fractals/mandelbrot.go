package fractals

// mandelbrot calculates the number of iterations for a given complex number c
// Returns the iteration count when |z|² > 4 (diverged) or maxIter if in the set
func Mandelbrot(cr, ci float64, maxIter int) int {
	zr, zi := 0.0, 0.0

	for i := range maxIter {
		// Calculate z² = (zr + zi*i)²
		zr2 := zr * zr
		zi2 := zi * zi

		// Check if diverged (|z|² > 4)
		if zr2+zi2 > 4.0 {
			return i
		}

		// z = z² + c
		zi = 2.0*zr*zi + ci
		zr = zr2 - zi2 + cr
	}

	return maxIter
}
