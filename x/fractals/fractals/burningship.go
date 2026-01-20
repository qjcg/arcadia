package fractals

// burningShip calculates the Burning Ship fractal
// Uses absolute values: z = (|Re(z)| + i|Im(z)|)² + c
func BurningShip(cr, ci float64, maxIter int) int {
	zr, zi := 0.0, 0.0

	for i := range maxIter {
		zr2 := zr * zr
		zi2 := zi * zi

		if zr2+zi2 > 4.0 {
			return i
		}

		// Take absolute values before squaring
		if zr < 0 {
			zr = -zr
		}
		if zi < 0 {
			zi = -zi
		}

		// z = z² + c
		zi = 2.0*zr*zi + ci
		zr = zr2 - zi2 + cr
	}

	return maxIter
}
