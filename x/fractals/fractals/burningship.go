package fractals

import "github.com/qjcg/arcadia/x/fractals/internal/optimize"

// BurningShip calculates the Burning Ship fractal
// Uses smooth coloring for better visual quality
// Returns float64 for smooth iteration count interpolation
func BurningShip(cr, ci float64, maxIter int) float64 {
	calc := optimize.NewIterationCalculator(maxIter)
	return calc.BurningShipOptimized(cr, ci)
}
