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

	fractalName := "Mandelbulb"
	paramName := "Power"
	paramVal := g.power
	if g.fractalType == 1 {
		fractalName = "Mandelbox"
		paramName = "Scale"
		paramVal = g.boxScale
	}

	uiText := fmt.Sprintf(
		"3D Fractal Viewer - %s [%s]\n\n"+
			"Controls:\n"+
			"1 / 2      Switch Fractal (Mandelbulb / Mandelbox)\n"+
			"WASD       Move X/Z\n"+
			"Space/Shift Move Y\n"+
			"Mouse/Arrows Look around (Click to capture, ESC to release)\n"+
			"[]         Adjust %s\n"+
			"P          Toggle Autopilot (orbital flight)\n"+
			"V          Toggle Vantage mode (tour)\n"+
			"F          Toggle Fullscreen\n"+
			"Q          Quit\n"+
			"%s"+
			"Position: (%.2f, %.2f, %.2f)\n"+
			"Pitch/Yaw: (%.2f, %.2f)\n"+
			"%s: %.2f | Iterations: %d\n\n"+
			"FPS: %.1f",
		fractalName, modeStr,
		paramName,
		vantageInfo,
		g.camX, g.camY, g.camZ,
		g.camPitch, g.camYaw,
		paramName, paramVal, g.iterations,
		ebiten.ActualFPS(),
	)
	ebitenutil.DebugPrint(screen, uiText)
}
