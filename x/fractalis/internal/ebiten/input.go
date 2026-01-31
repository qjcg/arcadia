package ebiten

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// handleInput processes keyboard and mouse input
func (g *Game) handleInput() bool {
	const dt = 1.0 / 60.0 // Assume 60 FPS

	// Fractal switching
	if inpututil.IsKeyJustPressed(ebiten.Key1) {
		g.switchToFractal(2) // Mandelbrot
	}
	if inpututil.IsKeyJustPressed(ebiten.Key2) {
		g.switchToFractal(3) // Julia
	}
	if inpututil.IsKeyJustPressed(ebiten.Key3) {
		g.switchToFractal(4) // Burning Ship
	}
	if inpututil.IsKeyJustPressed(ebiten.Key4) {
		g.switchToFractal(5) // Tricorn
	}
	if inpututil.IsKeyJustPressed(ebiten.Key5) {
		g.switchToFractal(6) // Multibrot 3
	}
	if inpututil.IsKeyJustPressed(ebiten.Key6) {
		g.switchToFractal(7) // Multibrot 4
	}
	if inpututil.IsKeyJustPressed(ebiten.Key7) {
		g.switchToFractal(8) // Celtic
	}
	if inpututil.IsKeyJustPressed(ebiten.Key8) {
		g.switchToFractal(9) // Perpendicular
	}
	if inpututil.IsKeyJustPressed(ebiten.Key9) {
		g.switchToFractal(10) // Manhattan
	}
	if inpututil.IsKeyJustPressed(ebiten.Key0) {
		g.switchToFractal(g.fractalType)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyF1) {
		g.fractalType = 0 // Mandelbulb
		g.camX, g.camY, g.camZ = 0.0, 1.5, 6.0
		g.camPitch, g.camYaw = 0.0, 3.14159
		g.autopilotHasTarget = false
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyF2) {
		g.fractalType = 1 // Mandelbox
		g.camX, g.camY, g.camZ = 5.0, 2.0, 5.0
		g.camPitch, g.camYaw = -0.3, 3.927
		g.boxScale = 2.7
		g.autopilotHasTarget = false
	}

	// Parameter adjustment
	if ebiten.IsKeyPressed(ebiten.KeyBracketRight) {
		if g.fractalType == 0 {
			g.power += 0.05
		} else if g.fractalType == 1 {
			g.boxScale += 0.01
		} else {
			g.iterations += 1
		}
	}
	if ebiten.IsKeyPressed(ebiten.KeyBracketLeft) {
		if g.fractalType == 0 {
			g.power -= 0.05
		} else if g.fractalType == 1 {
			g.boxScale -= 0.01
		} else {
			if g.iterations > 1 {
				g.iterations -= 1
			}
		}
	}

	// Autopilot speed adjustment
	if ebiten.IsKeyPressed(ebiten.KeyEqual) {
		g.autopilotSpeed += 0.01
	}
	if ebiten.IsKeyPressed(ebiten.KeyMinus) {
		g.autopilotSpeed -= 0.01
		if g.autopilotSpeed < 0 {
			g.autopilotSpeed = 0
		}
	}

	// Zoom (Ebiten version uses camY/Power for zoom in 2D)
	if ebiten.IsKeyPressed(ebiten.KeyI) {
		if g.fractalType >= 2 {
			g.power += 0.05
		}
	}
	if ebiten.IsKeyPressed(ebiten.KeyO) {
		if g.fractalType >= 2 {
			g.power -= 0.05
		}
	}

	// Mouse click to capture
	if !g.mouseCaptured && inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		g.mouseCaptured = true
		ebiten.SetCursorMode(ebiten.CursorModeCaptured)
		g.lastMouseX, g.lastMouseY = ebiten.CursorPosition()
	}

	// Mouse capture toggle
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		g.mouseCaptured = false
		ebiten.SetCursorMode(ebiten.CursorModeVisible)
	}

	// Movement forces
	forceX := 0.0
	forceY := 0.0
	forceZ := 0.0

	if ebiten.IsKeyPressed(ebiten.KeyW) {
		forceZ -= g.forwardSpeed
	}
	if ebiten.IsKeyPressed(ebiten.KeyS) {
		forceZ += g.forwardSpeed
	}
	if ebiten.IsKeyPressed(ebiten.KeyA) {
		forceX -= g.sideSpeed
	}
	if ebiten.IsKeyPressed(ebiten.KeyD) {
		forceX += g.sideSpeed
	}
	if ebiten.IsKeyPressed(ebiten.KeySpace) {
		forceY += g.forwardSpeed
	}
	if ebiten.IsKeyPressed(ebiten.KeyShiftLeft) {
		forceY -= g.forwardSpeed
	}

	// Physics constants
	const friction = 0.90
	const rotFriction = 0.85

	// Update velocities
	g.velX = (g.velX + forceX) * friction
	g.velY = (g.velY + forceY) * friction
	g.velZ = (g.velZ + forceZ) * friction

	// Rotate movement by camera yaw
	sinYaw := math.Sin(g.camYaw)
	cosYaw := math.Cos(g.camYaw)

	g.camX += (g.velX*cosYaw + g.velZ*sinYaw) * dt
	g.camY += g.velY * dt
	g.camZ += (-g.velX*sinYaw + g.velZ*cosYaw) * dt

	// Mouse look (when captured)
	if g.mouseCaptured {
		mx, my := ebiten.CursorPosition()
		dx := float64(mx - g.lastMouseX)
		dy := float64(my - g.lastMouseY)

		g.velYaw -= dx * 0.015
		g.velPitch -= dy * 0.015

		g.lastMouseX, g.lastMouseY = mx, my
	}

	// Rotation controls (arrow keys)
	if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) {
		g.velYaw += g.rotSpeed
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowRight) {
		g.velYaw -= g.rotSpeed
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowUp) {
		g.velPitch -= g.rotSpeed
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowDown) {
		g.velPitch += g.rotSpeed
	}

	// Apply rotation velocity
	g.camYaw += g.velYaw * dt
	g.camPitch += g.velPitch * dt
	g.velYaw *= rotFriction
	g.velPitch *= rotFriction

	// Clamp pitch to avoid flipping
	g.camPitch = math.Max(-math.Pi/2+0.1, math.Min(math.Pi/2-0.1, g.camPitch))

	// Toggle autopilot on Z (was P, re-mapped further below)
	if inpututil.IsKeyJustPressed(ebiten.KeyZ) {
		// Simple debounce - only toggle if we're not already processing a key
		if !g.autopilotEnabled && !g.vantageEnabled {
			g.autopilotEnabled = true
			g.autopilotTime = 0
		} else if g.autopilotEnabled {
			g.autopilotEnabled = false
		}
	}

	// Toggle vantage mode on V
	if inpututil.IsKeyJustPressed(ebiten.KeyV) {
		if !g.vantageEnabled && !g.autopilotEnabled {
			g.vantageEnabled = true
			g.vantageIndex = 0
			g.vantageTimer = 0
			g.vantagePanning = false
			g.moveToVantage(g.vantageVantages[0])
		} else if g.vantageEnabled {
			g.vantageEnabled = false
		}
	}

	// Toggle fullscreen on F
	if inpututil.IsKeyJustPressed(ebiten.KeyF) {
		ebiten.SetFullscreen(!ebiten.IsFullscreen())
	}

	// Quit on Q
	if ebiten.IsKeyPressed(ebiten.KeyQ) {
		return true
	}

	return false
}
