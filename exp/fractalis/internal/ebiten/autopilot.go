//go:build ebiten || (js && wasm)

package ebiten

import (
	"math"

	"github.com/qjcg/arcadia/exp/fractalis/internal/core/persistence"
)

// updateAutopilot moves the camera in an interesting pattern
func (g *Game) updateAutopilot() {
	const dt = 1.0 / 60.0
	g.autopilotTime += dt * g.autopilotSpeed

	if g.fractalType == 0 {
		// Mandelbulb: Orbital path with varying height
		angle := g.autopilotTime
		radius := g.autopilotRadius + math.Sin(g.autopilotTime*0.5)*0.5
		height := g.autopilotHeight * math.Sin(g.autopilotTime*0.3)

		// Calculate target position
		targetX := math.Sin(angle) * radius
		targetZ := math.Cos(angle) * radius
		targetY := height

		// Smoothly interpolate current position to target
		g.camX += (targetX - g.camX) * 0.05
		g.camY += (targetY - g.camY) * 0.05
		g.camZ += (targetZ - g.camZ) * 0.05

		// Look at origin
		g.camYaw = math.Atan2(g.camX, g.camZ) + math.Pi
		g.camPitch = math.Atan2(g.camY, math.Sqrt(g.camX*g.camX+g.camZ*g.camZ)) * -0.5
	} else if g.fractalType == 1 {
		// Mandelbox: Infinite descent / zoom
		// We use a periodic zoom into a self-similar point
		// Mandelbox (scale 2) has roughly a factor of 2 self-similarity
		zoomSpeed := 0.2
		zoomCycle := math.Mod(g.autopilotTime*zoomSpeed, 1.0)
		// Zoom from 10.0 down to 0.625 (16x zoom)
		zoomLevel := 10.0 * math.Pow(2.0, -4.0*zoomCycle)

		// Target a point that often has interesting self-similarity in Mandelbox
		targetX := 0.0
		targetY := 0.0
		targetZ := 0.0

		// Spiral motion that also scales with zoom to keep features in view
		rotAngle := g.autopilotTime * 0.4
		rotRadius := zoomLevel

		targetCamX := targetX + math.Cos(rotAngle)*rotRadius
		targetCamY := targetY + math.Sin(rotAngle*0.7)*rotRadius*0.5
		targetCamZ := targetZ + math.Sin(rotAngle)*rotRadius

		// Interpolate for smoothness but keep it responsive
		g.camX += (targetCamX - g.camX) * 0.1
		g.camY += (targetCamY - g.camY) * 0.1
		g.camZ += (targetCamZ - g.camZ) * 0.1

		// Look towards center
		targetYaw := math.Atan2(targetX-g.camX, targetZ-g.camZ)
		distXZ := math.Sqrt((targetX-g.camX)*(targetX-g.camX) + (targetZ-g.camZ)*(targetZ-g.camZ))
		targetPitch := math.Atan2(g.camY-targetY, distXZ) * -1.0

		g.camYaw = targetYaw
		g.camPitch = targetPitch
	}
}

func (g *Game) switchToFractal(fractalType int) {
	g.fractalType = fractalType
	g.flatPilot = ""

	// Set default positions based on fractal type
	if g.fractalType >= 2 {
		g.power = 0.0
		if g.fractalType == 3 { // Julia
			g.camX, g.camY, g.camZ = 0.0, 0.5, 0.0
		} else if g.fractalType == 4 { // Burning Ship
			g.camX, g.camY, g.camZ = -0.5, 0.5, -0.6
		} else {
			g.camX, g.camY, g.camZ = -0.5, 0.5, 0.0
		}
		g.camPitch, g.camYaw = -1.57, 0.0
	} else if g.fractalType == 0 { // Mandelbulb
		g.camX, g.camY, g.camZ = 0.0, 1.5, 6.0
		g.camPitch, g.camYaw = 0.0, 3.14159
		g.power = 8.0
	} else if g.fractalType == 1 { // Mandelbox
		g.camX, g.camY, g.camZ = 5.0, 2.0, 5.0
		g.camPitch, g.camYaw = -0.3, 3.927
		g.boxScale = 2.7
	}

	// Reset velocities
	g.velX, g.velY, g.velZ = 0, 0, 0
	g.velPitch, g.velYaw = 0, 0
}

// applyPersistenceConfig updates game state from a persistence.Config
func (g *Game) applyPersistenceConfig(cfg persistence.Config) {
	// Map fractal type string to ID
	fType := 1
	switch cfg.FractalType {
	case persistence.FractalMandelbulb:
		fType = 0
	case persistence.FractalMandelbox:
		fType = 1
	case persistence.FractalMandelbrot:
		fType = 2
	case persistence.FractalJulia:
		fType = 3
	case persistence.FractalBurningShip:
		fType = 4
	case persistence.FractalTricorn:
		fType = 5
	case persistence.FractalMultibrot3:
		fType = 6
	case persistence.FractalMultibrot4:
		fType = 7
	case persistence.FractalMultibrot5:
		fType = 8
	case persistence.FractalCeltic:
		fType = 9
	case persistence.FractalPerpendicular:
		fType = 10
	case persistence.FractalManhattan:
		fType = 11
	case persistence.FractalNewton:
		fType = 12
	}

	g.fractalType = fType
	g.iterations = cfg.MaxIter
	if cfg.Zoom > 0 {
		g.power = math.Log2(cfg.Zoom)
	} else {
		g.power = 0
	}

	if fType >= 2 {
		g.camX = cfg.CenterX
		g.camZ = cfg.CenterY
		g.camY = 0.5
		g.camPitch = -1.570796
		g.camYaw = 0.0
	}
}

// getPersistenceConfig creates a core Config object from current game state
func (g *Game) getPersistenceConfig() persistence.Config {
	fractalTypeName := persistence.FractalMandelbrot
	switch g.fractalType {
	case 2:
		fractalTypeName = persistence.FractalMandelbrot
	case 3:
		fractalTypeName = persistence.FractalJulia
	case 4:
		fractalTypeName = persistence.FractalBurningShip
	case 5:
		fractalTypeName = persistence.FractalTricorn
	case 6:
		fractalTypeName = persistence.FractalMultibrot3
	case 7:
		fractalTypeName = persistence.FractalMultibrot4
	case 8:
		fractalTypeName = persistence.FractalMultibrot5
	case 9:
		fractalTypeName = persistence.FractalCeltic
	case 10:
		fractalTypeName = persistence.FractalPerpendicular
	case 11:
		fractalTypeName = persistence.FractalManhattan
	case 12:
		fractalTypeName = persistence.FractalNewton
	}

	zoom := math.Pow(2.0, g.power)

	return persistence.Config{
		FractalType: fractalTypeName,
		CenterX:     g.camX,
		CenterY:     g.camZ, // In shader, camZ is used as the imaginary axis (Y in 2D)
		Zoom:        zoom,
		MaxIter:     g.iterations,
		JuliaCr:     -0.7, // Default Julia parameters as in shader
		JuliaCi:     0.27015,
	}
}
