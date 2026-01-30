package fractals

import "math"

func Manhattan(cr, ci float64, maxIter int) int {
	zr, zi := 0.0, 0.0

	for i := 0; i < maxIter; i++ {
		zr2 := zr * zr
		zi2 := zi * zi

		if math.Abs(zr)+math.Abs(zi) > 4.0 {
			return i
		}

		zi = 2.0*zr*zi + ci
		zr = zr2 - zi2 + cr
	}

	return maxIter
}
