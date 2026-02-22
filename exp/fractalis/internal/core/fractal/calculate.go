package fractal

// CalculateFractal dispatches to the appropriate fractal function based on fractal type
// Returns float64 to support smooth iteration counting for better color gradients
func CalculateFractal(cr, ci float64, fractalType string, maxIter int, juliaCr, juliaCi float64) float64 {
	switch fractalType {
	case "mandelbrot":
		return Mandelbrot(cr, ci, maxIter)
	case "julia":
		return Julia(cr, ci, juliaCr, juliaCi, maxIter)
	case "burningship":
		return BurningShip(cr, ci, maxIter)
	case "tricorn":
		return float64(Tricorn(cr, ci, maxIter))
	case "multibrot3":
		return float64(Multibrot3(cr, ci, maxIter))
	case "multibrot4":
		return float64(Multibrot4(cr, ci, maxIter))
	case "celtic":
		return float64(Celtic(cr, ci, maxIter))
	case "perpendicular":
		return float64(Perpendicular(cr, ci, maxIter))
	case "multibrot5":
		return float64(Multibrot5(cr, ci, maxIter))
	case "manhattan":
		return float64(Manhattan(cr, ci, maxIter))
	case "newton":
		return float64(Newton(cr, ci, maxIter))
	default:
		return Mandelbrot(cr, ci, maxIter)
	}
}
