package ebiten

import "math"

// VantagePoint represents an interesting camera viewpoint
type VantagePoint struct {
	Name        string
	X, Y, Z     float64
	Pitch, Yaw  float64
	Description string
}

// defaultVantagePoints returns the predefined vantage points
func defaultVantagePoints() []VantagePoint {
	return []VantagePoint{
		{"Outside Front", 0.0, 1.5, 6.0, -0.2, 3.14159, "Classic exterior view"},
		{"Outside Side", 2.5, 0.0, 0.0, 0.0, 4.71239, "Side profile view"},
		{"Outside Top", 0.0, 2.5, 0.5, -1.2, 3.14159, "Looking down from above"},
		{"Deep Inside", 0.0, 0.0, 0.0, 0.0, 0.0, "Inside the fractal core"},
		{"Canyon Pass", 0.8, 0.0, 0.8, 0.0, 3.927, "Through a Mandelbulb canyon"},
		{"Spiral Arm", 1.5, 0.5, 1.5, -0.3, 3.5, "Along a spiral arm"},
		{"Close Detail", 0.3, 0.0, 1.2, 0.2, 3.0, "Close-up surface detail"},
		{"Wide Orbit", 0.0, 1.0, 4.0, -0.2, 3.14159, "Wide orbital view"},
	}
}

// updateVantage handles the vantage mode scene transitions
func (g *Game) updateVantage() {
	const dt = 1.0 / 60.0
	g.vantageTimer += dt

	// Time to switch to next vantage?
	if g.vantageTimer >= g.vantageSceneTime {
		g.vantageTimer = 0
		g.vantageIndex = (g.vantageIndex + 1) % len(g.vantageVantages)
		g.moveToVantage(g.vantageVantages[g.vantageIndex])
	}

	// Slowly pan around the current vantage point
	if g.vantagePanning {
		g.vantagePanProgress += dt * 0.1
		vp := g.vantageVantages[g.vantageIndex]
		// Add gentle sway to the view
		g.camYaw = vp.Yaw + math.Sin(g.vantagePanProgress)*0.3
		g.camPitch = vp.Pitch + math.Cos(g.vantagePanProgress*0.7)*0.1
	}
}

// moveToVantage smoothly moves the camera to a vantage point
func (g *Game) moveToVantage(vp VantagePoint) {
	// Just set the position directly for now (could be interpolated)
	g.camX = vp.X
	g.camY = vp.Y
	g.camZ = vp.Z
	g.camPitch = vp.Pitch
	g.camYaw = vp.Yaw
	g.vantagePanning = true
	g.vantagePanProgress = 0
}
