package fractal

func Multibrot5(cr, ci float64, maxIter int) int {
	zr, zi := 0.0, 0.0

	for i := range maxIter {
		zr2 := zr * zr
		zi2 := zi * zi

		if zr2+zi2 > 4.0 {
			return i
		}

		z2r := zr2 - zi2
		z2i := 2.0 * zr * zi

		z4r := z2r*z2r - z2i*z2i
		z4i := 2.0 * z2r * z2i

		newZr := z4r*zr - z4i*zi + cr
		newZi := z4r*zi + z4i*zr + ci
		zr = newZr
		zi = newZi
	}

	return maxIter
}
