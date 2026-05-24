package transition

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

// Particle represents a falling piece of glass/wall
type Particle struct {
	X, Y          float64
	VelX, VelY    float64
	Width, Height float64
	Opacity       float64
	Symbol        rune
	Color         string
}

// Breakthrough represents the breakthrough transition state
type Breakthrough struct {
	Mode            int
	Progress        float64
	Target          string
	ZoomStart       float64
	AllFractalTypes []string
	CenterX         float64
	CenterY         float64
	Particles       []*Particle
	RNG             *rand.Rand
	CrackMap        map[string]bool // marks which cells have cracks
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
		Particles:       make([]*Particle, 0),
		RNG:             rand.New(rand.NewSource(time.Now().UnixNano())),
		CrackMap:        make(map[string]bool),
	}
}

// Start initiates the breakthrough transition to a new fractal type
func (b *Breakthrough) Start(currentFractal string) {
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
		targetIndex = b.RNG.Intn(len(b.AllFractalTypes))
	}

	b.Target = b.AllFractalTypes[targetIndex]
	b.Progress = 0.01 // Start slightly above 0 so animation triggers
	b.ZoomStart = 1.0 // This should be set to actual zoom level by caller
	b.Particles = make([]*Particle, 0)
	b.CrackMap = make(map[string]bool)
}

// Update progresses the breakthrough transition
func (b *Breakthrough) Update() (completed bool, centerX, centerY, zoomLevel float64, message string) {
	if b.Progress >= 1.0 {
		return true, -0.5, 0.0, 1.0, fmt.Sprintf("Transitioning to %s (Breakthrough)", b.Target)
	}

	// Generate initial crack pattern at start of phase 1
	if b.Progress < 0.05 && len(b.Particles) == 0 {
		b.generateCrackPatternAndParticles()
	}

	if b.Progress < 1.0 {
		b.Progress += 0.05 // Progress transition at 5% per tick (20 ticks ≈ 0.33 seconds at 60fps)
	}

	if b.Progress >= 1.0 {
		return true, -0.5, 0.0, 1.0, fmt.Sprintf("Transitioning to %s (Breakthrough)", b.Target)
	}

	// Update particle physics
	b.updateParticles()

	// Create camera shake and position shift synchronized with particle destruction
	if b.Progress < 0.3 {
		// Phase 1: Impact - heavy shaking as wall cracks, zoom out
		progress := b.Progress / 0.3
		shakeIntensity := math.Sin(progress*math.Pi*12) * (1.0 - progress*0.4)
		b.ZoomStart = 1.0 * (1.0 - progress*0.5 + math.Abs(shakeIntensity)*0.1) // Zoom out with shake
		b.CenterX = -0.5 + shakeIntensity*0.3
		b.CenterY = 0.0 + shakeIntensity*0.3
	} else if b.Progress < 0.7 {
		// Phase 2: Pieces falling - gradual view transition
		progress := (b.Progress - 0.3) / 0.4
		b.ZoomStart = 0.5 + progress*0.5 + math.Sin(progress*math.Pi*4)*0.08
		b.CenterX = -0.5 + math.Sin(progress*math.Pi*2)*0.1
		b.CenterY = 0.0 + math.Cos(progress*math.Pi*2)*0.1
	} else {
		// Phase 3: Final settling
		progress := (b.Progress - 0.7) / 0.3
		b.ZoomStart = 1.0 - progress*0.1 + math.Sin(progress*math.Pi*2)*0.05
		b.CenterX = -0.5
		b.CenterY = 0.0
	}

	return false, b.CenterX, b.CenterY, b.ZoomStart, ""
}

// generateCrackPatternAndParticles creates the initial crack pattern and particles
func (b *Breakthrough) generateCrackPatternAndParticles() {
	// Generate a radial crack pattern from center with branches
	centerX, centerY := 0.5, 0.5

	// Create main crack rays from center
	numRays := 3 + b.RNG.Intn(3)
	for ray := range numRays {
		angle := (float64(ray) / float64(numRays)) * 2 * math.Pi
		// Trace crack ray outward
		for dist := range 15 {
			x := centerX + math.Cos(angle)*float64(dist)/15
			y := centerY + math.Sin(angle)*float64(dist)/15
			// Add perpendicular branches
			if dist%3 == 0 {
				for branch := -1; branch <= 1; branch += 2 {
					branchAngle := angle + math.Pi/2*float64(branch)
					for blen := 1; blen < 5; blen++ {
						bx := x + math.Cos(branchAngle)*float64(blen)/10
						by := y + math.Sin(branchAngle)*float64(blen)/10
						b.CrackMap[fmt.Sprintf("%.2f,%.2f", bx, by)] = true
					}
				}
			}
			b.CrackMap[fmt.Sprintf("%.2f,%.2f", x, y)] = true
		}
	}

	// Create glass shards as particles from crack regions
	for range 30 {
		x := b.RNG.Float64()
		y := b.RNG.Float64()

		// Particles spawn from crack areas
		p := &Particle{
			X:       x,
			Y:       y,
			VelX:    (b.RNG.Float64() - 0.5) * 1.0,
			VelY:    b.RNG.Float64() * 0.2, // slight downward bias
			Width:   0.03 + b.RNG.Float64()*0.05,
			Height:  0.03 + b.RNG.Float64()*0.05,
			Opacity: 1.0,
			Symbol:  '█',
			Color:   "\033[90m", // dark gray for glass
		}
		b.Particles = append(b.Particles, p)
	}
}

// updateParticles updates particle physics and removes far particles
func (b *Breakthrough) updateParticles() {
	gravity := 0.1

	for i := len(b.Particles) - 1; i >= 0; i-- {
		p := b.Particles[i]

		// Apply gravity
		p.VelY += gravity * 0.05

		// Update position
		p.X += p.VelX
		p.Y += p.VelY

		// Fade out over time
		p.Opacity -= 0.02

		// Remove particles that have fallen off screen or faded
		if p.Y > 1.2 || p.Opacity <= 0 {
			b.Particles = append(b.Particles[:i], b.Particles[i+1:]...)
		}
	}
}

// GetParticles returns the current particles for rendering
func (b *Breakthrough) GetParticles() []*Particle {
	return b.Particles
}

// GetCrackMap returns the crack pattern
func (b *Breakthrough) GetCrackMap() map[string]bool {
	return b.CrackMap
}

// GetMessage returns the transition message
func (b *Breakthrough) GetMessage() string {
	if b.Mode == TransitionBreakthrough {
		return fmt.Sprintf("Transitioning to %s (Breakthrough)", b.Target)
	}
	return ""
}
