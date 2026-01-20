package fractals

// multibrot3 calculates the Multibrot set with power 3
// z = z³ + c
func Multibrot3(cr, ci float64, maxIter int) int {
	zr, zi := 0.0, 0.0

	for i := range maxIter {
		zr2 := zr * zr
		zi2 := zi * zi

		if zr2+zi2 > 4.0 {
			return i
		}

		// z³ = (zr + zi*i)³ = zr³ + 3*zr²*zi*i - 3*zr*zi² - zi³*i
		//    = (zr³ - 3*zr*zi²) + i(3*zr²*zi - zi³)
		newZr := zr*zr2 - 3.0*zr*zi2 + cr
		newZi := 3.0*zr2*zi - zi*zi2 + ci
		zr = newZr
		zi = newZi
	}

	return maxIter
}
