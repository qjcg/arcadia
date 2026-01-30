package transitions

import (
	"fmt"
	"math/rand"
	"time"
)

// Fade transition constants
const (
	TransitionFade = 1
)

// Fade represents the fade transition state
type Fade struct {
	Mode            int
	Progress        float64
	Target          string
	ZoomStart       float64
	AllFractalTypes []string
}

// NewFade creates a new fade transition
type NewFade struct {
	Mode            int
	Progress        float64
	Target          string
	ZoomStart       float64
	AllFractalTypes []string
}

// NewFade creates a new fade transition
func NewFadeTransition(allFractalTypes []string) *Fade {
	return &Fade{
		Mode:            TransitionFade,
		Progress:        0.0,
		Target:          "",
		ZoomStart:       1.0,
		AllFractalTypes: allFractalTypes,
	}
}

// Start initiates the fade transition to a new fractal type
func (f *Fade) Start(currentFractal string) {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Pick a random fractal type (different from current)
	currentIndex := -1
	for i, ft := range f.AllFractalTypes {
		if ft == currentFractal {
			currentIndex = i
			break
		}
	}

	// Find a different fractal type
	targetIndex := currentIndex
	for targetIndex == currentIndex {
		targetIndex = rng.Intn(len(f.AllFractalTypes))
	}

	f.Target = f.AllFractalTypes[targetIndex]
	f.Progress = 0.01 // Start slightly above 0 so animation triggers
	f.ZoomStart = 1.0 // This should be set to actual zoom level by caller
}

// Update progresses the fade transition
func (f *Fade) Update() (completed bool, message string) {
	if f.Progress < 1.0 {
		f.Progress += 0.05 // Progress transition at 5% per tick
	}

	if f.Progress >= 1.0 {
		return true, fmt.Sprintf("Transitioning to %s (Fade)", f.Target)
	}

	return false, ""
}

// GetMessage returns the transition message
func (f *Fade) GetMessage() string {
	if f.Mode == TransitionFade {
		return fmt.Sprintf("Transitioning to %s (Fade)", f.Target)
	}
	return ""
}
