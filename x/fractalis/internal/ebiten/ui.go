package ebiten

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"golang.org/x/image/colornames"
)

var colorBlack = colornames.Black

// drawUI renders the user interface overlay
func (g *Game) drawUI(screen *ebiten.Image) {
	modeStr := "Manual"
	if g.autopilotEnabled {
		modeStr = "AUTOPILOT"
	} else if g.vantageEnabled {
		modeStr = "VANTAGE"
	}

	vantageInfo := ""
	if g.vantageEnabled && len(g.vantageVantages) > 0 {
		vp := g.vantageVantages[g.vantageIndex]
		remaining := g.vantageSceneTime - g.vantageTimer
		vantageInfo = fmt.Sprintf("\nVantage: %s (%.1fs)\n%s\n", vp.Name, remaining, vp.Description)
	}

	uiText := fmt.Sprintf(
		"3D Fractal Viewer - Mandelbulb [%s]\n\n"+
			"Controls:\n"+
			"WASD      Move X/Z\n"+
			"Space/Shift Move Y\n"+
			"Mouse/Arrows Look around (Click to capture, ESC to release)\n"+
			"A          Toggle Autopilot (orbital flight)\n"+
			"V          Toggle Vantage mode (scene tour)\n"+
			"Q          Quit\n"+
			"%s"+
			"Position: (%.2f, %.2f, %.2f)\n"+
			"Pitch/Yaw: (%.2f, %.2f)\n"+
			"Power: %.1f | Iterations: %d\n\n"+
			"FPS: %.1f",
		modeStr,
		vantageInfo,
		g.camX, g.camY, g.camZ,
		g.camPitch, g.camYaw,
		g.power, g.iterations,
		ebiten.ActualFPS(),
	)
	ebitenutil.DebugPrint(screen, uiText)
}
