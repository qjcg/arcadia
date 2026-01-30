package search

import (
	"math"

	"github.com/qjcg/arcadia/x/fractalis/internal/precision"
	"github.com/qjcg/arcadia/x/fractalis/persistence"
)

// InterestCalculator computes how "interesting" a point is based on iteration variance
type InterestCalculator struct {
	config            persistence.Config
	getEffectiveDelta func() float64
	calculateFractal  func(cr, ci float64, cfg persistence.Config) float64
}

// NewInterestCalculator creates a calculator for scoring point interestingness
func NewInterestCalculator(
	config persistence.Config,
	getEffectiveDelta func() float64,
	calculateFractal func(cr, ci float64, cfg persistence.Config) float64,
) *InterestCalculator {
	return &InterestCalculator{
		config:            config,
		getEffectiveDelta: getEffectiveDelta,
		calculateFractal:  calculateFractal,
	}
}

// CalculateScore determines how "interesting" a point is by looking at iteration
// variation in its 3x3 neighborhood. Higher scores = more detail/variation = more interesting.
func (ic *InterestCalculator) CalculateScore(cx, cy float64) float64 {
	const neighborhoodSize = 9
	samples := make([]int, 0, neighborhoodSize)

	delta := ic.getEffectiveDelta() * 0.3

	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			cr := cx + float64(dx)*delta
			ci := cy + float64(dy)*delta/2.0
			iter := ic.calculateFractal(cr, ci, ic.config)
			samples = append(samples, int(iter))
		}
	}

	stats := precision.CalculateDistribution(samples)

	// Score components:
	// 1. Range score: High range = edge/boundary (most important)
	rangeScore := stats.Range * 100.0

	// 2. Variance score: Variation within neighborhood
	varianceScore := stats.StdDev * 20.0

	// 3. Boundary bonus: Prefer areas with medium iteration counts
	boundaryScore := 0.0
	if stats.Mean > 0 && stats.Mean < float64(ic.config.MaxIter) {
		normalizedIter := stats.Mean / float64(ic.config.MaxIter)
		if normalizedIter > 0.1 && normalizedIter < 0.9 {
			boundaryScore = 500.0
		} else if normalizedIter > 0.05 && normalizedIter < 0.95 {
			boundaryScore = 200.0
		}
	}

	// Heavily penalize completely uniform areas
	if stats.Range == 0 {
		return 0.0
	}

	return rangeScore + varianceScore + boundaryScore
}

// IsViewUniform checks if the current view is mostly uniform (boring)
// by sampling points and checking variation in iteration counts
func (ic *InterestCalculator) IsViewUniform() bool {
	const sampleSize = 25

	samples := make([]int, 0, sampleSize)
	delta := ic.getEffectiveDelta()

	for i := range sampleSize {
		offsetX := delta * float64(i%5-2)
		offsetY := delta * float64(i/5-2) / 2.0

		cr := ic.config.CenterX + offsetX
		ci := ic.config.CenterY + offsetY

		iter := ic.calculateFractal(cr, ci, ic.config)
		samples = append(samples, int(iter))
	}

	stats := precision.CalculateDistribution(samples)

	// A view is uniform if:
	// 1. The range is very small (less than 5% of maxIter), OR
	// 2. The standard deviation is very small (less than 10% of maxIter)
	rangeThreshold := float64(ic.config.MaxIter) * 0.05
	varianceThreshold := float64(ic.config.MaxIter) * 0.10

	return stats.Range < rangeThreshold || stats.StdDev < varianceThreshold
}

// SearchPass represents one pass of the spiral search
type SearchPass struct {
	NumCandidates int
	RadiusScale   float64
}

// FindInterestingPoint searches for a point with high variation/detail using multi-pass spiral search
func (ic *InterestCalculator) FindInterestingPoint(passes []SearchPass) (float64, float64) {
	bestX := ic.config.CenterX
	bestY := ic.config.CenterY
	bestScore := 0.0

	baseViewSize := 3.5 / ic.config.Zoom

	for _, pass := range passes {
		searchRadius := baseViewSize * pass.RadiusScale
		for i := 0; i < pass.NumCandidates; i++ {
			angle := float64(i) * 2.399
			radius := searchRadius * math.Sqrt(float64(i)/float64(pass.NumCandidates))

			x := ic.config.CenterX + radius*math.Cos(angle)
			y := ic.config.CenterY + radius*math.Sin(angle)/2.0

			score := ic.CalculateScore(x, y)
			if score > bestScore {
				bestScore = score
				bestX = x
				bestY = y
			}
		}
	}

	return bestX, bestY
}

// DefaultSearchPasses returns the standard 3-pass search configuration
func DefaultSearchPasses() []SearchPass {
	return []SearchPass{
		{NumCandidates: 40, RadiusScale: 0.5},
		{NumCandidates: 40, RadiusScale: 1.5},
		{NumCandidates: 50, RadiusScale: 4.0},
	}
}
