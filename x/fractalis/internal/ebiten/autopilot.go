package ebiten

import (
	"math"
	"math/rand"
	"time"

	"github.com/qjcg/arcadia/x/fractalis/internal/core/persistence"
	"github.com/qjcg/arcadia/x/fractalis/internal/core/search"
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
	} else if g.fractalType >= 2 {
		// 2D fractal autopilot: Intelligent zoom and interest-based panning
		// Similar to TUI mode behavior

		// 1. Zoom in
		// In Ebiten, power is log2(zoom)
		// Base zoom speed scaled by autopilotSpeed
		zoomStep := 1.0 + (0.02 * g.autopilotSpeed)
		g.power += math.Log2(zoomStep)

		// 2. Adaptive iterations
		zoom := math.Pow(2.0, g.power)
		baseIter := g.autopilotIterBase
		if baseIter == 0 {
			baseIter = 50
		}
		logZoom10 := math.Log10(zoom)
		if logZoom10 < 0 {
			logZoom10 = 0
		}
		g.iterations = baseIter + int(logZoom10*20.0)
		if g.iterations > 1000 {
			g.iterations = 1000
		}

		// 3. Target management
		cfg := g.getPersistenceConfig()
		shouldFindNewTarget := false
		if !g.autopilotHasTarget {
			shouldFindNewTarget = true
		} else if g.autopilotPanProgress >= 1.0 {
			shouldFindNewTarget = true
		} else if g.autopilotPanProgress > 0.8 && g.calculator.IsViewUniform(cfg) {
			shouldFindNewTarget = true
		}

		if shouldFindNewTarget {
			newX, newZ := g.calculator.FindInterestingPoint(search.DescentSearchPasses(), cfg)
			g.autopilotTargetX = newX
			g.autopilotTargetZ = newZ
			g.autopilotHasTarget = true
			g.autopilotPanProgress = 0.0
		}

		// 4. Smooth pan toward target
		if g.autopilotHasTarget && g.autopilotPanProgress < 1.0 {
			deltaX := g.autopilotTargetX - g.camX
			deltaZ := g.autopilotTargetZ - g.camZ

			// Pan rate scaled by autopilotSpeed
			panRate := 0.04 * g.autopilotSpeed
			g.camX += deltaX * panRate
			g.camZ += deltaZ * panRate

			// Update progress
			viewSize := 3.5 / zoom
			threshold := (viewSize / 1000.0) * (viewSize / 1000.0)
			distSq := deltaX*deltaX + deltaZ*deltaZ

			if distSq < threshold {
				g.autopilotPanProgress = 1.0
			} else {
				g.autopilotPanProgress += 0.015 * g.autopilotSpeed
			}
		}

		// 5. Check if we hit maximum zoom and should transition (similar to TUI mode)
		// Reduced threshold to 22.0 (approx 4 million zoom) to stay within float32 precision limits
		if g.power > 22.0 {
			// Pick a random 2D fractal type (2-10) different from current
			rng := rand.New(rand.NewSource(time.Now().UnixNano()))
			next := g.fractalType
			for next == g.fractalType {
				next = 2 + rng.Intn(9) // 2 to 10
			}
			g.switchToFractal(next)
		}
	}
}

// switchToFractal resets the game state for a new fractal type
func (g *Game) switchToFractal(fractalType int) {
	g.fractalType = fractalType
	g.autopilotHasTarget = false

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
		fractalTypeName = persistence.FractalCeltic
	case 9:
		fractalTypeName = persistence.FractalPerpendicular
	case 10:
		fractalTypeName = persistence.FractalManhattan
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
