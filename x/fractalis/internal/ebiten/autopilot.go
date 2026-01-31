package ebiten

import "math"

// updateAutopilot moves the camera in an orbital pattern around the Mandelbulb
func (g *Game) updateAutopilot() {
	const dt = 1.0 / 60.0
	g.autopilotTime += dt * g.autopilotSpeed

	// Orbital path with varying height
	angle := g.autopilotTime
	radius := g.autopilotRadius + math.Sin(g.autopilotTime*0.5)*0.5
	height := g.autopilotHeight * math.Sin(g.autopilotTime*0.3)

	// Calculate target position
	targetX := math.Sin(angle) * radius
	targetZ := math.Cos(angle) * radius
	targetY := height

	// Smoothly interpolate current position to target
	g.camX += (targetX - g.camX) * 0.02
	g.camY += (targetY - g.camY) * 0.02
	g.camZ += (targetZ - g.camZ) * 0.02

	// Look at origin
	g.camYaw = math.Atan2(g.camX, g.camZ) + math.Pi
	g.camPitch = math.Atan2(g.camY, math.Sqrt(g.camX*g.camX+g.camZ*g.camZ)) * -0.3
}
