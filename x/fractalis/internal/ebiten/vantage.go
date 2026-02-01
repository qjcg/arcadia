package ebiten

import (
	"math"
	"math/rand"
	"time"

	"github.com/qjcg/arcadia/x/fractalis/internal/core/search"
)

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

// updateVantage handles the vantage mode scene transitions
func (g *Game) updateVantage() {
	const dt = 1.0 / 60.0

	// Initialize first scene if not yet initialized
	if !g.vantageInitialized {
		g.randomVantage()
		g.vantageInitialized = true
		g.vantageTimer = 0
		return
	}

	g.vantageTimer += dt

	// Check if it's time to switch to next vantage
	if g.vantageTimer >= g.vantageSceneTime {
		g.vantageTimer = 0
		g.randomVantage()
	}

	// Apply slow pan if we have a target (2D mode)
	if g.fractalType >= 2 && g.vantageHasTarget && g.vantagePanProgress < 1.0 {
		deltaX := g.vantageTargetX - g.camX
		deltaZ := g.vantageTargetZ - g.camZ

		// Slowly move toward target
		g.camX += deltaX * 0.005
		g.camZ += deltaZ * 0.005

		g.vantagePanProgress += 0.002
		if g.vantagePanProgress > 1.0 {
			g.vantagePanProgress = 1.0
		}
	} else if g.vantagePanning {
		// 3D mode: gentle sway around current viewpoint
		g.vantagePanProgress += dt * 0.1
		// Add gentle sway to the view
		// We use small offsets to avoid drifting too far from curated points
		g.camYaw += math.Sin(g.vantagePanProgress) * 0.001
		g.camPitch += math.Cos(g.vantagePanProgress*0.7) * 0.0005
	}
}

// randomVantage selects a random interesting viewpoint
func (g *Game) randomVantage() {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Choose between 2D (40%) and 3D (60%)
	if rng.Float64() < 0.4 {
		// 2D Vantage
		cfg := g.getPersistenceConfig()
		search.RandomizeConfig(&cfg, g.calculator)
		g.applyPersistenceConfig(cfg)

		// Find interesting point to pan toward
		newX, newZ := g.calculator.FindInterestingPoint(search.DefaultSearchPasses(), cfg)
		g.vantageTargetX = newX
		g.vantageTargetZ = newZ
		g.vantageHasTarget = true
		g.vantagePanProgress = 0.0
		g.vantagePanning = false
	} else {
		// 3D Vantage from curated list
		g.vantageIndex = rng.Intn(len(g.vantageVantages))
		vp := g.vantageVantages[g.vantageIndex]
		g.moveToVantage(vp)
		g.vantageHasTarget = false
	}
}

// moveToVantage smoothly moves the camera to a vantage point
func (g *Game) moveToVantage(vp VantagePoint) {
	g.fractalType = vp.Type
	if vp.Type == 0 {
		g.power = vp.Scale
	} else {
		g.boxScale = vp.Scale
	}
	g.camX = vp.X
	g.camY = vp.Y
	g.camZ = vp.Z
	g.camPitch = vp.Pitch
	g.camYaw = vp.Yaw
	g.vantagePanning = true
	g.vantagePanProgress = 0
}
