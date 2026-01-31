package ebiten

import "math"

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
	} else {
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
