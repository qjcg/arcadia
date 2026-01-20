package transitions

import (
	"fmt"
	"math"
	"math/rand"
	"time"
)

// Breakthrough transition constants
const (
	TransitionBreakthrough = 4
)

// Breakthrough represents the breakthrough transition state
type Breakthrough struct {
	Mode            int
	Progress        float64
	Target          string
	ZoomStart       float64
	AllFractalTypes []string
	CenterX         float64
	CenterY         float64
}

// NewBreakthrough creates a new breakthrough transition
func NewBreakthroughTransition(allFractalTypes []string) *Breakthrough {
	return &Breakthrough{
		Mode:            TransitionBreakthrough,
		Progress:        0.0,
		Target:          "",
		ZoomStart:       1.0,
		AllFractalTypes: allFractalTypes,
		CenterX:         -0.5,
		CenterY:         0.0,
	}
}

// Start initiates the breakthrough transition to a new fractal type
func (b *Breakthrough) Start(currentFractal string) {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Pick a random fractal type (different from current)
	currentIndex := -1
	for i, ft := range b.AllFractalTypes {
		if ft == currentFractal {
			currentIndex = i
			break
		}
	}

	// Find a different fractal type
	targetIndex := currentIndex
	for targetIndex == currentIndex {
		targetIndex = rng.Intn(len(b.AllFractalTypes))
	}

	b.Target = b.AllFractalTypes[targetIndex]
	b.Progress = 0.01 // Start slightly above 0 so animation triggers
	b.ZoomStart = 1.0 // This should be set to actual zoom level by caller
}

// Update progresses the breakthrough transition
func (b *Breakthrough) Update() (completed bool, centerX, centerY, zoomLevel float64, message string) {
	if b.Progress < 1.0 {
		b.Progress += 0.05 // Progress transition at 5% per tick
	}

	if b.Progress >= 1.0 {
		return true, -0.5, 0.0, 1.0, fmt.Sprintf("Transitioning to %s (Breakthrough)", b.Target)
	}

	// Create dramatic zoom-out, then sudden zoom-in with position shift
	if b.Progress < 0.3 {
		// Phase 1: Rapid zoom out (30% of transition time)
		progress := b.Progress / 0.3
		b.ZoomStart = b.ZoomStart * (1.0 - progress*0.8) // Zoom out to 20% of start
	} else if b.Progress < 0.6 {
		// Phase 2: Sudden position shift and zoom in (next 30% of time)
		progress := (b.Progress - 0.3) / 0.3
		// Dramatic position shift - "break through" to new location
		b.CenterX = -0.5 + 1.0*math.Cos(b.Progress*20.0)
		b.CenterY = 0.0 + 1.0*math.Sin(b.Progress*20.0)
		b.ZoomStart = 0.2*b.ZoomStart + progress*1.5 // Zoom in rapidly
	} else {
		// Phase 3: Final zoom in and stabilization (last 40% of time)
		progress := (b.Progress - 0.6) / 0.4
		b.ZoomStart = 1.7 * b.ZoomStart * (1.0 + progress*0.5) // Final zoom adjustment
		// Smooth out position to target
		b.CenterX = -0.5 + (1.0-progress)*0.5*math.Cos(b.Progress*10.0)
		b.CenterY = 0.0 + (1.0-progress)*0.5*math.Sin(b.Progress*10.0)
	}

	return false, b.CenterX, b.CenterY, b.ZoomStart, ""
}

// GetMessage returns the transition message
func (b *Breakthrough) GetMessage() string {
	if b.Mode == TransitionBreakthrough {
		return fmt.Sprintf("Transitioning to %s (Breakthrough)", b.Target)
	}
	return ""
}
