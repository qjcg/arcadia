package fractal

// newton calculates the Newton fractal for z^3 - 1 = 0
// Shows basins of attraction for the three cube roots of unity
// Uses Newton's method: z_new = z - f(z)/f'(z) = z - (z^3 - 1)/(3z^2)
func Newton(zr, zi float64, maxIter int) int {
	const tolerance = 1e-6

	// Three roots of z^3 - 1 = 0 (cube roots of unity)
	// root1 = 1 + 0i
	// root2 = -0.5 + 0.866i
	// root3 = -0.5 - 0.866i

	for i := range maxIter {
		// Calculate z^2
		zr2 := zr * zr
		zi2 := zi * zi

		// Calculate z^3 = z * z^2
		z3r := zr*zr2 - zr*zi2 - zr*zi2
		z3i := zi*zr2 + zi*zr2 + zr*zi2

		// f(z) = z^3 - 1
		fr := z3r - 1.0
		fi := z3i

		// f'(z) = 3z^2
		fpr := 3.0 * (zr2 - zi2)
		fpi := 6.0 * zr * zi

		// Calculate f(z)/f'(z) using complex division
		// (a + bi) / (c + di) = ((ac + bd) + (bc - ad)i) / (c^2 + d^2)
		denom := fpr*fpr + fpi*fpi
		if denom < 1e-10 {
			return i
		}

		divr := (fr*fpr + fi*fpi) / denom
		divi := (fi*fpr - fr*fpi) / denom

		// z_new = z - f(z)/f'(z)
		newZr := zr - divr
		newZi := zi - divi

		// Check convergence
		dr := newZr - zr
		di := newZi - zi
		if dr*dr+di*di < tolerance*tolerance {
			// Converged - determine which root and color based on it
			// Check distance to each root
			d1 := (newZr-1.0)*(newZr-1.0) + newZi*newZi
			d2 := (newZr+0.5)*(newZr+0.5) + (newZi-0.866025)*(newZi-0.866025)
			d3 := (newZr+0.5)*(newZr+0.5) + (newZi+0.866025)*(newZi+0.866025)

			// Return different values based on which root we converged to
			// This creates the three-fold symmetric basins
			if d1 < d2 && d1 < d3 {
				return maxIter - i // Root 1 (real)
			} else if d2 < d3 {
				return maxIter - i*2/3 // Root 2 (upper)
			} else {
				return maxIter - i/3 // Root 3 (lower)
			}
		}

		zr = newZr
		zi = newZi
	}

	return 0 // Did not converge
}
