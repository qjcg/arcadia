package fractal

// celtic calculates the Celtic Mandelbrot fractal
// Uses z = (|Re(z²)| + i*Im(z²)) + c
func Celtic(cr, ci float64, maxIter int) int {
	zr, zi := 0.0, 0.0

	for i := range maxIter {
		zr2 := zr * zr
		zi2 := zi * zi

		if zr2+zi2 > 4.0 {
			return i
		}

		// Calculate z²
		newZr := zr2 - zi2
		newZi := 2.0 * zr * zi

		// Take absolute value of real part
		if newZr < 0 {
			newZr = -newZr
		}

		// Add c
		zr = newZr + cr
		zi = newZi + ci
	}

	return maxIter
}
