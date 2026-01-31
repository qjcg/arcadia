package ebiten

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

// handleInput processes keyboard and mouse input
// Returns true if the game should quit
func (g *Game) handleInput() bool {
	const dt = 1.0 / 60.0 // Assume 60 FPS

	// Mouse capture toggle
	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		g.mouseCaptured = !g.mouseCaptured
		if g.mouseCaptured {
			ebiten.SetCursorMode(ebiten.CursorModeHidden)
			g.lastMouseX, g.lastMouseY = g.width/2, g.height/2
		} else {
			ebiten.SetCursorMode(ebiten.CursorModeVisible)
		}
	}

	// Movement
	moveX := 0.0
	moveY := 0.0
	moveZ := 0.0

	if ebiten.IsKeyPressed(ebiten.KeyW) {
		moveZ -= g.forwardSpeed * dt
	}
	if ebiten.IsKeyPressed(ebiten.KeyS) {
		moveZ += g.forwardSpeed * dt
	}
	if ebiten.IsKeyPressed(ebiten.KeyA) {
		moveX -= g.sideSpeed * dt
	}
	if ebiten.IsKeyPressed(ebiten.KeyD) {
		moveX += g.sideSpeed * dt
	}
	if ebiten.IsKeyPressed(ebiten.KeySpace) {
		moveY += g.forwardSpeed * dt
	}
	if ebiten.IsKeyPressed(ebiten.KeyShiftLeft) {
		moveY -= g.forwardSpeed * dt
	}

	// Rotate movement by camera yaw
	sinYaw := math.Sin(g.camYaw)
	cosYaw := math.Cos(g.camYaw)

	g.camX += moveX*cosYaw + moveZ*sinYaw
	g.camY += moveY
	g.camZ += -moveX*sinYaw + moveZ*cosYaw

	// Mouse look (when captured)
	if g.mouseCaptured {
		mx, my := ebiten.CursorPosition()
		dx := float64(mx - g.lastMouseX)
		dy := float64(my - g.lastMouseY)

		g.camYaw -= dx * 0.002
		g.camPitch -= dy * 0.002

		// Clamp pitch to avoid flipping
		g.camPitch = math.Max(-math.Pi/2+0.1, math.Min(math.Pi/2-0.1, g.camPitch))
	}

	// Rotation controls (arrow keys)
	if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) {
		g.camYaw += g.rotSpeed * dt
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowRight) {
		g.camYaw -= g.rotSpeed * dt
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowUp) {
		g.camPitch = math.Max(-math.Pi/2+0.1, g.camPitch-g.rotSpeed*dt)
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowDown) {
		g.camPitch = math.Min(math.Pi/2-0.1, g.camPitch+g.rotSpeed*dt)
	}

	// Toggle autopilot on A
	if ebiten.IsKeyPressed(ebiten.KeyA) {
		// Simple debounce - only toggle if we're not already processing a key
		if !g.autopilotEnabled && !g.vantageEnabled {
			g.autopilotEnabled = true
			g.autopilotTime = 0
		} else if g.autopilotEnabled {
			g.autopilotEnabled = false
		}
	}

	// Toggle vantage mode on V
	if ebiten.IsKeyPressed(ebiten.KeyV) {
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

	// Quit on Q
	if ebiten.IsKeyPressed(ebiten.KeyQ) {
		return true
	}

	return false
}
