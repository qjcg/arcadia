package fractals

// julia calculates the Julia set for a given point z and constant c
// z iterates: z = z² + c where c is constant (juliaCr, juliaCi)
func Julia(zr, zi, juliaCr, juliaCi float64, maxIter int) int {
	for i := range maxIter {
		zr2 := zr * zr
		zi2 := zi * zi

		if zr2+zi2 > 4.0 {
			return i
		}

		// z = z² + c
		zi = 2.0*zr*zi + juliaCi
		zr = zr2 - zi2 + juliaCr
	}

	return maxIter
}
