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
		g.fractalType = 0 // Mandelbulb
		g.camX, g.camY, g.camZ = 0.0, 1.5, 6.0
		g.camPitch, g.camYaw = 0.0, 3.14159
	}
	if inpututil.IsKeyJustPressed(ebiten.Key2) {
		g.fractalType = 1 // Mandelbox
		g.camX, g.camY, g.camZ = 5.0, 2.0, 5.0
		g.camPitch, g.camYaw = -0.3, 3.927
		g.boxScale = 2.7
	}

	// Parameter adjustment
	if ebiten.IsKeyPressed(ebiten.KeyBracketRight) {
		if g.fractalType == 0 {
			g.power += 0.05
		} else {
			g.boxScale += 0.01
		}
	}
	if ebiten.IsKeyPressed(ebiten.KeyBracketLeft) {
		if g.fractalType == 0 {
			g.power -= 0.05
		} else {
			g.boxScale -= 0.01
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

	// Toggle autopilot on P (re-mapped from A to avoid conflict with movement)
	if inpututil.IsKeyJustPressed(ebiten.KeyP) {
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
