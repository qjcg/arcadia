package transition

import (
	"fmt"
	"math/rand"
	"time"
)

// ZoomOutIn transition constants
const (
	TransitionZoomOutIn = 2
)

// ZoomOutIn represents the zoom out/in transition state
type ZoomOutIn struct {
	Mode            int
	Progress        float64
	Target          string
	ZoomStart       float64
	AllFractalTypes []string
}

// NewZoomOutIn creates a new zoom out/in transition
func NewZoomOutInTransition(allFractalTypes []string) *ZoomOutIn {
	return &ZoomOutIn{
		Mode:            TransitionZoomOutIn,
		Progress:        0.0,
		Target:          "",
		ZoomStart:       1.0,
		AllFractalTypes: allFractalTypes,
	}
}

// Start initiates the zoom out/in transition to a new fractal type
func (z *ZoomOutIn) Start(currentFractal string) {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Pick a random fractal type (different from current)
	currentIndex := -1
	for i, ft := range z.AllFractalTypes {
		if ft == currentFractal {
			currentIndex = i
			break
		}
	}

	// Find a different fractal type
	targetIndex := currentIndex
	for targetIndex == currentIndex {
		targetIndex = rng.Intn(len(z.AllFractalTypes))
	}

	z.Target = z.AllFractalTypes[targetIndex]
	z.Progress = 0.01 // Start slightly above 0 so animation triggers
	z.ZoomStart = 1.0 // This should be set to actual zoom level by caller
}

// Update progresses the zoom out/in transition
func (z *ZoomOutIn) Update() (completed bool, zoomLevel float64, message string) {
	if z.Progress < 1.0 {
		z.Progress += 0.05 // Progress transition at 5% per tick
	}

	// Apply zoom out/in effect
	if z.Progress < 0.5 {
		// Zoom out phase
		progress := z.Progress / 0.5
		return false, z.ZoomStart * (1.0 - progress*0.9), "" // Zoom out to 10% of start
	} else {
		// Zoom in phase
		progress := (z.Progress - 0.5) / 0.5
		if z.Progress >= 1.0 {
			return true, 0.1*z.ZoomStart + progress*0.9, fmt.Sprintf("Transitioning to %s (Zoom Out/In)", z.Target)
		}
		return false, 0.1*z.ZoomStart + progress*0.9, "" // Zoom in from 10% to 90% of start
	}
}

// GetMessage returns the transition message
func (z *ZoomOutIn) GetMessage() string {
	if z.Mode == TransitionZoomOutIn {
		return fmt.Sprintf("Transitioning to %s (Zoom Out/In)", z.Target)
	}
	return ""
}
