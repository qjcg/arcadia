//go:build ebiten || (js && wasm)

package ebiten

// VantagePoint represents an interesting camera viewpoint
type VantagePoint struct {
	Name        string
	Type        int // 0: Mandelbulb, 1: Mandelbox
	X, Y, Z     float64
	Pitch, Yaw  float64
	Scale       float64 // boxScale or bulb power
	Description string
}

// defaultVantagePoints returns the predefined vantage points
func defaultVantagePoints() []VantagePoint {
	return []VantagePoint{
		// Mandelbox Vantages (Now default)
		{"Box City", 1, 5.0, 2.0, 5.0, -0.3, 3.927, 2.7, "Inside the Mandelbox canyons"},
		{"Geometric Abyss", 1, 1.0, 0.5, 1.0, -1.0, 0.785, 2.7, "Staring into the box center"},
		{"The Grid", 1, 10.0, 10.0, 10.0, -0.7, 3.927, 2.7, "High altitude view of the box structure"},
		{"Symmetry Hall", 1, 0.0, 2.0, 0.0, -1.57, 0.0, 2.5, "Looking straight down the central axis"},
		{"Infinite Columns", 1, 3.2, 0.0, 3.2, 0.0, 3.927, 2.8, "Endless columns of recursion"},
		{"The Vault", 1, 0.5, 0.5, 0.5, 0.5, 0.785, 2.0, "Deep within the low-scale core"},

		// Mandelbulb Vantages
		{"Bulb Prime", 0, 0.0, 1.5, 6.0, -0.2, 3.14159, 8.0, "Classic exterior view of the Mandelbulb"},
		{"Power Canyon", 0, 0.8, 0.1, 0.8, -0.1, 3.927, 12.0, "Extreme detail in a high-power canyon"},
		{"Spiral Descent", 0, 1.2, 0.8, 1.2, -0.5, 3.5, 8.0, "Descending into the spiral arms"},
		{"Core Horizon", 0, 0.2, 0.0, 1.0, 0.1, 3.14159, 8.0, "Low-angle view of the bulb horizon"},
		{"Bulb Satellite", 0, 3.0, 3.0, 3.0, -0.6, 3.927, 8.0, "Distant view of the complex shell"},
		{"Micro-Buds", 0, 0.1, 0.05, 0.4, 1.2, 0.0, 9.5, "Close-up of the secondary bulb structures"},
	}
}
