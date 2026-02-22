package transition

import (
	"fmt"
	"math"
	"math/rand"
	"time"
)

// Rotate transition constants
const (
	TransitionRotate = 3
)

// Rotate represents the rotate transition state
type Rotate struct {
	Mode            int
	Progress        float64
	Target          string
	ZoomStart       float64
	AllFractalTypes []string
	CenterX         float64
	CenterY         float64
}

// NewRotate creates a new rotate transition
func NewRotateTransition(allFractalTypes []string) *Rotate {
	return &Rotate{
		Mode:            TransitionRotate,
		Progress:        0.0,
		Target:          "",
		ZoomStart:       1.0,
		AllFractalTypes: allFractalTypes,
		CenterX:         -0.5,
		CenterY:         0.0,
	}
}

// Start initiates the rotate transition to a new fractal type
func (r *Rotate) Start(currentFractal string) {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Pick a random fractal type (different from current)
	currentIndex := -1
	for i, ft := range r.AllFractalTypes {
		if ft == currentFractal {
			currentIndex = i
			break
		}
	}

	// Find a different fractal type
	targetIndex := currentIndex
	for targetIndex == currentIndex {
		targetIndex = rng.Intn(len(r.AllFractalTypes))
	}

	r.Target = r.AllFractalTypes[targetIndex]
	r.Progress = 0.01 // Start slightly above 0 so animation triggers
	r.ZoomStart = 1.0 // This should be set to actual zoom level by caller
}

// Update progresses the rotate transition
func (r *Rotate) Update() (completed bool, centerX, centerY float64, message string) {
	if r.Progress < 1.0 {
		r.Progress += 0.05 // Progress transition at 5% per tick
	}

	if r.Progress >= 1.0 {
		return true, -0.5, 0.0, fmt.Sprintf("Transitioning to %s (Rotate)", r.Target)
	}

	// Rotate the view during transition
	angle := r.Progress * math.Pi * 2
	radius := 0.5 / r.ZoomStart
	r.CenterX = -0.5 + radius*math.Cos(angle)
	r.CenterY = 0.0 + radius*math.Sin(angle)

	return false, r.CenterX, r.CenterY, ""
}

// GetMessage returns the transition message
func (r *Rotate) GetMessage() string {
	if r.Mode == TransitionRotate {
		return fmt.Sprintf("Transitioning to %s (Rotate)", r.Target)
	}
	return ""
}
