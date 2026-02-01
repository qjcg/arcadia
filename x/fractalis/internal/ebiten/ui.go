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
	if g.vantageEnabled {
		remaining := g.vantageSceneTime - g.vantageTimer
		if g.fractalType < 2 && g.vantageIndex >= 0 && g.vantageIndex < len(g.vantageVantages) {
			vp := g.vantageVantages[g.vantageIndex]
			vantageInfo = fmt.Sprintf("\nVantage: %s (%.1fs)\n%s\n", vp.Name, remaining, vp.Description)
		} else {
			vantageInfo = fmt.Sprintf("\nVantage: Random Scene (%.1fs)\nExploring interesting features...\n", remaining)
		}
	}

	fractalName := "Mandelbulb"
	paramName := "Power"
	paramVal := g.power
	if g.fractalType == 1 {
		fractalName = "Mandelbox"
		paramName = "Scale"
		paramVal = g.boxScale
	} else if g.fractalType == 2 {
		fractalName = "Mandelbrot"
		paramName = "ZoomExp"
	} else if g.fractalType == 3 {
		fractalName = "Julia"
		paramName = "ZoomExp"
	} else if g.fractalType == 4 {
		fractalName = "Burning Ship"
		paramName = "ZoomExp"
	} else if g.fractalType == 5 {
		fractalName = "Tricorn"
		paramName = "ZoomExp"
	} else if g.fractalType == 6 {
		fractalName = "Multibrot 3"
		paramName = "ZoomExp"
	} else if g.fractalType == 7 {
		fractalName = "Multibrot 4"
		paramName = "ZoomExp"
	} else if g.fractalType == 8 {
		fractalName = "Multibrot 5"
		paramName = "ZoomExp"
	} else if g.fractalType == 9 {
		fractalName = "Celtic"
		paramName = "ZoomExp"
	} else if g.fractalType == 10 {
		fractalName = "Perpendicular"
		paramName = "ZoomExp"
	} else if g.fractalType == 11 {
		fractalName = "Manhattan"
		paramName = "ZoomExp"
	} else if g.fractalType == 12 {
		fractalName = "Newton"
		paramName = "ZoomExp"
	}

	uiText := fmt.Sprintf(
		"3D Fractal Viewer - %s [%s]\n\n"+
			"Controls:\n"+
			"1-9        Switch Fractal (Mandelbrot, Julia, ...)\n"+
			"F1/F2      3D Fractals (Mandelbulb / Mandelbox)\n"+
			"WASD       Move X/Z (or Pan in 2D)\n"+
			"Space/Shift Move Y\n"+
			"I/O        Zoom In/Out (2D)\n"+
			"Mouse/Arrows Look around (Click to capture, ESC to release)\n"+
			"[]         Adjust %s / Iterations\n"+
			"+/-        Adjust Autopilot Speed\n"+
			"Z          Toggle Autopilot\n"+
			"V          Toggle Vantage mode (tour)\n"+
			"0          Reset View\n"+
			"F          Toggle Fullscreen\n"+
			"Q          Quit\n"+
			"%s"+
			"Position: (%.2f, %.2f, %.2f)\n"+
			"Pitch/Yaw: (%.2f, %.2f)\n"+
			"%s: %.2f | Iterations: %d\n"+
			"Autopilot Speed: %.2f\n\n"+
			"FPS: %.1f",
		fractalName, modeStr,
		paramName,
		vantageInfo,
		g.camX, g.camY, g.camZ,
		g.camPitch, g.camYaw,
		paramName, paramVal, g.iterations,
		g.autopilotSpeed,
		ebiten.ActualFPS(),
	)
	ebitenutil.DebugPrint(screen, uiText)
}
