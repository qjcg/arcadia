package ebiten

import (
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
		fractalType:       1,   // Start with Mandelbox
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

// drawFractal renders the shader
func (g *Game) drawFractal(screen *ebiten.Image) {
	uniforms := map[string]interface{}{
		"Time":        float32(time.Since(g.startTime).Seconds()),
		"Resolution":  [2]float32{float32(g.width), float32(g.height)},
		"CamPos":      [3]float32{float32(g.camX), float32(g.camY), float32(g.camZ)},
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
	ebiten.SetWindowTitle("Fractalis 3D - Mandelbox")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.MaximizeWindow()

	g.initShader()

	return ebiten.RunGame(g)
}
