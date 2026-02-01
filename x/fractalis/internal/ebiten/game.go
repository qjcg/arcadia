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

// PaintMode represents the visual enhancement mode for fractal interior/exterior
type PaintMode int

const (
	PaintModeNone PaintMode = iota
	// Inpainting modes - enhance interior of fractal
	InpaintSolidMode   // Solid color fill
	InpaintNoisyMode   // Noisy turbulence pattern
	InpaintIridMode    // Iridescent color shifting
	InpaintFractalMode // Recursive fractal-in-fractal pattern
	InpaintFireMode    // Molten fire effect inside
	// Outpainting modes - enhance exterior of fractal
	OutpaintGlowMode     // Soft glow halo around fractal
	OutpaintRippleMode   // Ripple/wave distortion fields
	OutpaintFogMode      // Foggy mist effect radiating outward
	OutpaintElectricMode // Electric discharge-like arcs around boundary
	OutpaintFireMode     // Realistic flames shooting outward
)

// Game represents the fractal engine game
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
	vantageInitialized bool
	vantageTargetX     float64
	vantageTargetZ     float64
	vantageHasTarget   bool

	// Paint mode state
	paintMode PaintMode
}

// NewGame creates a new fractal engine game
func NewGame(config persistence.Config) *Game {
	// Map fractal type string to engine type ID
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
	case persistence.FractalMultibrot5:
		fType = 8
	case persistence.FractalCeltic:
		fType = 9
	case persistence.FractalPerpendicular:
		fType = 10
	case persistence.FractalManhattan:
		fType = 11
	case persistence.FractalNewton:
		fType = 12
	}

	// Default camera settings vary by fractal type
	camX, camY, camZ := 5.0, 2.0, 5.0
	pitch, yaw := -0.3, 3.927
	power := 8.0

	if fType >= 2 {
		// 2D fractal: use config values and look straight down
		camX = config.CenterX
		camY = 0.5
		camZ = config.CenterY
		pitch = -1.570796 // Look straight down
		yaw = 0.0
		// Ebiten zoom is 2^power, so power is log2(zoom)
		if config.Zoom > 0 {
			power = math.Log2(config.Zoom)
		} else {
			power = 0.0
		}
	} else if fType == 0 {
		// Mandelbulb defaults
		camX, camY, camZ = 0.0, 1.5, 6.0
		pitch, yaw = 0.0, 3.14159
		power = 8.0
	}

	g := &Game{
		config:               config,
		width:                screenWidth,
		height:               screenHeight,
		camX:                 camX,
		camY:                 camY,
		camZ:                 camZ,
		camPitch:             pitch,
		camYaw:               yaw,
		forwardSpeed:         0.15,
		sideSpeed:            0.15,
		rotSpeed:             3.0,
		power:                power,
		boxScale:             2.7, // Mandelbox scale
		fractalType:          fType,
		iterations:           32,
		colorShift:           0.0,
		startTime:            time.Now(),
		autopilotEnabled:     config.AutoPilotEnabled,
		autopilotTime:        0,
		autopilotSpeed:       1.0,
		autopilotRadius:      2.5,
		autopilotHeight:      0.5,
		autopilotIterBase:    config.MaxIter,
		autopilotHasTarget:   false,
		autopilotPanProgress: 0.0,
		vantageEnabled:       config.VantageEnabled,
		vantageSceneTime:     float64(config.VantageSceneDuration),
		vantageVantages:      defaultVantagePoints(),
		paintMode:            PaintModeNone,
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
		"PaintMode":   float32(g.paintMode),
	}

	op := &ebiten.DrawRectShaderOptions{}
	op.Uniforms = uniforms
	op.Images[0] = nil

	screen.DrawRectShader(screen.Bounds().Dx(), screen.Bounds().Dy(), g.shader, op)
}

// Run starts the 3D game loop
func (g *Game) Run() error {
	if g.config.Fullscreen {
		ebiten.SetFullscreen(true)
	} else {
		ebiten.SetWindowSize(g.width, g.height)
	}

	title := "Fractalis"
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
	if !g.config.Fullscreen {
		ebiten.MaximizeWindow()
	}

	g.initShader()

	return ebiten.RunGame(g)
}

// getPaintModeName returns the name of the current paint mode
func (g *Game) getPaintModeName() string {
	switch g.paintMode {
	case InpaintSolidMode:
		return "Solid Inpaint"
	case InpaintNoisyMode:
		return "Noisy Inpaint"
	case InpaintIridMode:
		return "Iridescent Inpaint"
	case InpaintFractalMode:
		return "Fractal Inpaint"
	case InpaintFireMode:
		return "Fire Inpaint"
	case OutpaintGlowMode:
		return "Glow Outpaint"
	case OutpaintRippleMode:
		return "Ripple Outpaint"
	case OutpaintFogMode:
		return "Fog Outpaint"
	case OutpaintElectricMode:
		return "Electric Outpaint"
	case OutpaintFireMode:
		return "Fire Outpaint"
	default:
		return "None"
	}
}
