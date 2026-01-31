package ebiten

import (
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/qjcg/arcadia/x/fractalis/internal/core/persistence"
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
	forwardSpeed     float64
	sideSpeed        float64
	rotSpeed         float64

	// Mouse state
	lastMouseX, lastMouseY int
	mouseCaptured          bool

	// Fractal parameters
	power      float64
	iterations int
	colorShift float64

	// Autopilot state
	autopilotEnabled bool
	autopilotTime    float64
	autopilotSpeed   float64
	autopilotRadius  float64
	autopilotHeight  float64

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
		config:           config,
		width:            screenWidth,
		height:           screenHeight,
		camX:             0.0,
		camY:             1.5,
		camZ:             6.0,
		camPitch:         0.0,
		camYaw:           3.14159, // pi, looking toward -Z (at origin)
		forwardSpeed:     0.3,
		sideSpeed:        0.2,
		rotSpeed:         2.0,
		power:            8.0, // Mandelbulb power
		iterations:       10,
		colorShift:       0.0,
		startTime:        time.Now(),
		autopilotSpeed:   0.5,
		autopilotRadius:  2.5,
		autopilotHeight:  0.5,
		vantageSceneTime: 8.0, // 8 seconds per vantage point
		vantageVantages:  defaultVantagePoints(),
	}

	return g
}

// Layout implements ebiten.Game
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return g.width, g.height
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
		"Time":       float32(time.Since(g.startTime).Seconds()),
		"Resolution": [2]float32{float32(g.width), float32(g.height)},
		"CamPos":     [3]float32{float32(g.camX), float32(g.camY), float32(g.camZ)},
		"CamPitch":   float32(g.camPitch),
		"CamYaw":     float32(g.camYaw),
		"Power":      float32(g.power),
		"Iterations": float32(g.iterations),
		"ColorShift": float32(g.colorShift),
	}

	op := &ebiten.DrawRectShaderOptions{}
	op.Uniforms = uniforms
	op.Images[0] = nil

	screen.DrawRectShader(screen.Bounds().Dx(), screen.Bounds().Dy(), g.shader, op)
}

// Run starts the 3D game loop
func (g *Game) Run() error {
	ebiten.SetWindowSize(g.width, g.height)
	ebiten.SetWindowTitle("Fractalis 3D - Mandelbulb")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	g.initShader()

	return ebiten.RunGame(g)
}
