package ebiten

import (
	"math"
	"strings"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/qjcg/arcadia/x/fractalis/internal/core/fractal"
	"github.com/qjcg/arcadia/x/fractalis/internal/core/persistence"
	"github.com/qjcg/arcadia/x/fractalis/internal/core/search"
)

const (
	screenWidth  = 800
	screenHeight = 600
)

// Game represents the 3D fractal viewer
type Game struct {
	config      persistence.Config
	width       int
	height      int
	palette     []uint32
	shader      *ebiten.Shader
	initialized bool
	startTime   time.Time

	// Camera state
	camX, camY, camZ float64
	camPitch, camYaw float64
	velX, velY, velZ float64
	velPitch, velYaw float64
	forwardSpeed     float64
	sideSpeed        float64
	rotSpeed         float64

	// Mouse state
	lastMouseX, lastMouseY int
	mouseCaptured          bool

	// Fractal parameters
	power       float64
	boxScale    float64
	fractalType int // 0: Mandelbulb, 1: Mandelbox
	iterations  int
	colorShift  float64

	// Autopilot state
	autopilotEnabled     bool
	autopilotTime        float64
	autopilotSpeed       float64
	autopilotRadius      float64
	autopilotHeight      float64
	autopilotTargetX     float64
	autopilotTargetZ     float64
	autopilotHasTarget   bool
	autopilotPanProgress float64
	autopilotIterBase    int
	calculator           *search.InterestCalculator

	// Vantage state
	vantageEnabled     bool
	vantageVantages    []VantagePoint
	vantageIndex       int
	vantageTimer       float64
	vantageSceneTime   float64
	vantagePanning     bool
	vantagePanProgress float64
}

// NewGame creates a new 3D fractal game
func NewGame(config persistence.Config) *Game {
	// Map fractal type string to 3D engine type ID
	fType := 1 // Default to Mandelbox
	switch config.FractalType {
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
	case persistence.FractalCeltic:
		fType = 8
	case persistence.FractalPerpendicular:
		fType = 9
	case persistence.FractalManhattan:
		fType = 10
	}

	g := &Game{
		config:            config,
		width:             screenWidth,
		height:            screenHeight,
		camX:              5.0,
		camY:              2.0,
		camZ:              5.0,
		camPitch:          -0.3,
		camYaw:            3.927, // looking toward origin from positive X/Z
		forwardSpeed:      2.0,
		sideSpeed:         1.5,
		rotSpeed:          3.0,
		power:             8.0, // Mandelbulb power
		boxScale:          2.7, // Mandelbox scale
		fractalType:       fType,
		iterations:        32,
		colorShift:        0.0,
		startTime:         time.Now(),
		autopilotSpeed:    1.0,
		autopilotRadius:   2.5,
		autopilotHeight:   0.5,
		autopilotIterBase: config.MaxIter,
		vantageSceneTime:  8.0, // 8 seconds per vantage point
		vantageVantages:   defaultVantagePoints(),
	}

	g.calculator = search.NewInterestCalculator(func(cr, ci float64, cfg persistence.Config) float64 {
		return fractal.CalculateFractal(cr, ci, cfg.FractalType, cfg.MaxIter, cfg.JuliaCr, cfg.JuliaCi)
	})

	return g
}

// Layout implements ebiten.Game
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	g.width = outsideWidth
	g.height = outsideHeight
	return outsideWidth, outsideHeight
}

// Update implements ebiten.Game
func (g *Game) Update() error {
	// Handle keyboard input
	if g.handleInput() {
		return ebiten.Termination
	}

	// Update autopilot mode
	if g.autopilotEnabled {
		g.updateAutopilot()
	}

	// Update vantage mode
	if g.vantageEnabled {
		g.updateVantage()
	}

	// Update color shift for animation
	g.colorShift += 0.5

	return nil
}

// Draw implements ebiten.Game
func (g *Game) Draw(screen *ebiten.Image) {
	if g.shader == nil {
		g.drawLoading(screen)
		return
	}

	g.drawFractal(screen)
	g.drawUI(screen)
}

// drawLoading shows loading screen
func (g *Game) drawLoading(screen *ebiten.Image) {
	screen.Fill(colorBlack)
	ebitenutil.DebugPrint(screen, "Loading 3D shader...")
}

// splitFloat64 splits a float64 into two float32s for emulated double precision in shaders
func splitFloat64(f float64) [2]float32 {
	hi := float32(f)
	lo := float32(f - float64(hi))
	return [2]float32{hi, lo}
}

// drawFractal renders the shader
func (g *Game) drawFractal(screen *ebiten.Image) {
	zoom := 1.0
	if g.power > 0.0 {
		zoom = math.Pow(2.0, g.power)
	}
	invZoom := 1.0 / zoom // Base zoom

	// Coordinate mapping constants matching TUI/Core logic
	// viewWidth := 3.5 / zoom
	// cr := centerX + (normX-0.5)*viewWidth
	// normX := col / width
	// cr := centerX + (col/width - 0.5) * (3.5/zoom)
	// cr := centerX + (col - 0.5*width) * (3.5 / (width * zoom))

	unitToSize := 3.5 / (float64(g.width) * zoom)
	aspect := float64(g.height) / float64(g.width)

	uniforms := map[string]interface{}{
		"Time":        float32(time.Since(g.startTime).Seconds()),
		"Resolution":  [2]float32{float32(g.width), float32(g.height)},
		"CamPos":      [3]float32{float32(g.camX), float32(g.camY), float32(g.camZ)},
		"CamPosHi":    [2]float32{float32(g.camX), float32(g.camZ)},
		"CamPosLo":    [2]float32{float32(g.camX - float64(float32(g.camX))), float32(g.camZ - float64(float32(g.camZ)))},
		"InvZoom":     splitFloat64(invZoom),
		"UnitToSize":  splitFloat64(unitToSize),
		"Aspect":      float32(aspect),
		"CamPitch":    float32(g.camPitch),
		"CamYaw":      float32(g.camYaw),
		"Power":       float32(g.power),
		"BoxScale":    float32(g.boxScale),
		"FractalType": float32(g.fractalType),
		"Iterations":  float32(g.iterations),
		"ColorShift":  float32(g.colorShift),
	}

	op := &ebiten.DrawRectShaderOptions{}
	op.Uniforms = uniforms
	op.Images[0] = nil

	screen.DrawRectShader(screen.Bounds().Dx(), screen.Bounds().Dy(), g.shader, op)
}

// Run starts the 3D game loop
func (g *Game) Run() error {
	ebiten.SetWindowSize(g.width, g.height)

	title := "Fractalis 3D"
	if g.config.FractalType != "" {
		// Capitalize first letter
		name := g.config.FractalType
		if len(name) > 0 {
			name = strings.ToUpper(name[:1]) + name[1:]
		}
		title += " - " + name
	}
	ebiten.SetWindowTitle(title)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.MaximizeWindow()

	g.initShader()

	return ebiten.RunGame(g)
}
