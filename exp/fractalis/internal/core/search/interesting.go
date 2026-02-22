package search

import (
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/qjcg/arcadia/exp/fractalis/internal/core/persistence"
)

// InterestingPoint represents a known interesting location in a fractal
type InterestingPoint struct {
	X, Y float64
}

// InterestingJuliaPoint represents a known interesting Julia set parameter
type InterestingJuliaPoint struct {
	Cr, Ci float64
}

var (
	// InterestingMandelbrot coordinates for random exploration
	InterestingMandelbrot = []InterestingPoint{
		{-0.7436, 0.1314}, // Spiral region
		{-0.16, 1.0405},   // Elephant valley
		{-0.7269, 0.1889}, // Seahorse valley
		{-0.7453, 0.1127}, // Double spiral
		{-0.1592, 1.0317}, // North region
		{-0.7746, 0.1152}, // Tendril region
		{0.2805, 0.0089},  // East mini-mandelbrot
		{-0.1011, 0.9563}, // North filament
		{-0.7500, 0.0000}, // Main body edge
		{-0.1600, 1.0400}, // Antenna region
	}

	// InterestingJulia constants for variety
	InterestingJulia = []InterestingJuliaPoint{
		{-0.4, 0.6},       // Dendrite
		{0.285, 0.01},     // Douady rabbit
		{-0.8, 0.156},     // San Marco dragon
		{-0.7, 0.27},      // Classic
		{-0.835, -0.2321}, // Siegel disk
		{-0.123, 0.745},   // Douady rabbit variant
		{0.3, 0.5},        // Swirls
		{-0.79, 0.15},     // Dragon variant
		{-0.162, 1.04},    // Spiral
		{0.355, 0.355},    // Cross pattern
	}

	// AllFractalTypes lists all available 2D fractal types
	AllFractalTypes = []string{
		persistence.FractalMandelbrot,
		persistence.FractalJulia,
		persistence.FractalBurningShip,
		persistence.FractalTricorn,
		persistence.FractalMultibrot3,
		persistence.FractalMultibrot4,
		persistence.FractalMultibrot5,
		persistence.FractalCeltic,
		persistence.FractalPerpendicular,
		persistence.FractalManhattan,
		persistence.FractalNewton,
	}
)

// RandomizeConfig randomizes a fractal configuration with interesting parameters.
// It returns a message describing the new configuration.
func RandomizeConfig(cfg *persistence.Config, ic *InterestCalculator) string {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	const maxAttempts = 5
	var resultMsg string

	for attempt := 0; attempt < maxAttempts; attempt++ {
		// Random fractal type
		fractalType := AllFractalTypes[rng.Intn(len(AllFractalTypes))]
		cfg.FractalType = fractalType

		// Random zoom (log-uniform between 1.0 and 1000.0, weighted toward mid-range)
		logZoom := rng.Float64() * 3.0 // 0 to 3
		zoom := math.Pow(10.0, logZoom)
		cfg.Zoom = zoom

		// Random iterations appropriate for zoom level
		const defaultBaseIterations = 50
		const iterationScaleFactor = 20.0
		baseIter := defaultBaseIterations + rng.Intn(100)
		zoomBonus := int(math.Log10(zoom) * iterationScaleFactor)
		cfg.MaxIter = baseIter + zoomBonus
		if cfg.MaxIter > 300 {
			cfg.MaxIter = 300
		}

		// Random position based on fractal type
		if fractalType == persistence.FractalJulia {
			juliaPt := InterestingJulia[rng.Intn(len(InterestingJulia))]
			cfg.JuliaCr = juliaPt.Cr
			cfg.JuliaCi = juliaPt.Ci

			offset := 0.5 / zoom
			cfg.CenterX = (rng.Float64()*2.0 - 1.0) * offset
			cfg.CenterY = (rng.Float64()*2.0 - 1.0) * offset

		} else if isMandelbrotLike(fractalType) {
			seedPt := InterestingMandelbrot[rng.Intn(len(InterestingMandelbrot))]
			maxOffset := 0.3 / zoom
			cfg.CenterX = seedPt.X + (rng.Float64()*2.0-1.0)*maxOffset
			cfg.CenterY = seedPt.Y + (rng.Float64()*2.0-1.0)*maxOffset

		} else if fractalType == persistence.FractalBurningShip {
			cfg.CenterX = -0.5 + (rng.Float64()*2.0-1.0)*0.3/zoom
			cfg.CenterY = -0.6 + (rng.Float64()*2.0-1.0)*0.3/zoom

		} else if fractalType == persistence.FractalTricorn {
			cfg.CenterX = -0.5 + (rng.Float64()*2.0-1.0)*0.3/zoom
			cfg.CenterY = 0.0 + (rng.Float64()*2.0-1.0)*0.3/zoom
		}

		// Check if the result is interesting if we have a calculator
		if ic == nil || !ic.IsViewUniform(*cfg) {
			resultMsg = fmt.Sprintf("Random: %s @ %.1fx", fractalType, zoom)
			return resultMsg
		}
	}

	// Fallback to a known-good interesting point
	seedPt := InterestingMandelbrot[0]
	cfg.FractalType = persistence.FractalMandelbrot
	cfg.CenterX = seedPt.X
	cfg.CenterY = seedPt.Y
	cfg.Zoom = 10.0 + rng.Float64()*90.0
	cfg.MaxIter = 100 + rng.Intn(100)

	return fmt.Sprintf("Random: %s @ %.1fx (fallback)", cfg.FractalType, cfg.Zoom)
}

func isMandelbrotLike(ft string) bool {
	return ft == persistence.FractalMandelbrot ||
		ft == persistence.FractalMultibrot3 ||
		ft == persistence.FractalMultibrot4 ||
		ft == persistence.FractalMultibrot5 ||
		ft == persistence.FractalCeltic ||
		ft == persistence.FractalPerpendicular
}
