package fractals

// tricorn (Mandelbar) calculates the Tricorn fractal
// Uses conjugate: z = conj(z)² + c
func Tricorn(cr, ci float64, maxIter int) int {
	zr, zi := 0.0, 0.0

	for i := range maxIter {
		zr2 := zr * zr
		zi2 := zi * zi

		if zr2+zi2 > 4.0 {
			return i
		}

		// z = conj(z)² + c
		// conj(z)² = (zr - zi*i)² = zr² - zi² - 2*zr*zi*i
		zi = -2.0*zr*zi + ci // Note the negative sign
		zr = zr2 - zi2 + cr
	}

	return maxIter
}
