package fractal

// multibrot4 calculates the Multibrot set with power 4
// z = z⁴ + c
func Multibrot4(cr, ci float64, maxIter int) int {
	zr, zi := 0.0, 0.0

	for i := range maxIter {
		zr2 := zr * zr
		zi2 := zi * zi

		if zr2+zi2 > 4.0 {
			return i
		}

		// z⁴ = ((zr + zi*i)²)²
		// First: z² = zr² - zi² + 2*zr*zi*i
		zr_temp := zr2 - zi2
		zi_temp := 2.0 * zr * zi
		// Second: (z²)²
		newZr := zr_temp*zr_temp - zi_temp*zi_temp + cr
		newZi := 2.0*zr_temp*zi_temp + ci
		zr = newZr
		zi = newZi
	}

	return maxIter
}
