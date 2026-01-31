package search

import (
	"math"

	"github.com/qjcg/arcadia/x/fractalis/internal/core/persistence"
	"github.com/qjcg/arcadia/x/fractalis/internal/core/precision"
)

// InterestCalculator computes how "interesting" a point is based on iteration variance
type InterestCalculator struct {
	calculateFractal func(cr, ci float64, cfg persistence.Config) float64
}

// NewInterestCalculator creates a calculator for scoring point interestingness
func NewInterestCalculator(
	calculateFractal func(cr, ci float64, cfg persistence.Config) float64,
) *InterestCalculator {
	return &InterestCalculator{
		calculateFractal: calculateFractal,
	}
}

// CalculateScore determines how "interesting" a point is by looking at iteration
// variation in its 3x3 neighborhood. Higher scores = more detail/variation = more interesting.
func (ic *InterestCalculator) CalculateScore(cx, cy float64, cfg persistence.Config) float64 {
	const neighborhoodSize = 9
	samples := make([]int, 0, neighborhoodSize)

	// Use a view-proportional delta that captures meaningful variation
	// at any zoom level. At zoom=1, view size is 3.5; at zoom=100, view size is 0.035.
	// We want samples spread across ~1/20th of the view.
	viewSize := 3.5 / cfg.Zoom
	delta := viewSize / 20.0

	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			cr := cx + float64(dx)*delta
			ci := cy + float64(dy)*delta/2.0
			iter := ic.calculateFractal(cr, ci, cfg)
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
	if stats.Mean > 0 && stats.Mean < float64(cfg.MaxIter) {
		normalizedIter := stats.Mean / float64(cfg.MaxIter)
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
func (ic *InterestCalculator) IsViewUniform(cfg persistence.Config) bool {
	const sampleSize = 25

	samples := make([]int, 0, sampleSize)
	// Use a view-proportional delta that captures meaningful variation
	// at any zoom level. Samples spread across ~1/10th of the view.
	viewSize := 3.5 / cfg.Zoom
	delta := viewSize / 10.0

	for i := range sampleSize {
		offsetX := delta * float64(i%5-2)
		offsetY := delta * float64(i/5-2) / 2.0

		cr := cfg.CenterX + offsetX
		ci := cfg.CenterY + offsetY

		iter := ic.calculateFractal(cr, ci, cfg)
		samples = append(samples, int(iter))
	}

	stats := precision.CalculateDistribution(samples)

	// A view is uniform if:
	// 1. The range is very small (less than 5% of maxIter), OR
	// 2. The standard deviation is very small (less than 10% of maxIter)
	rangeThreshold := float64(cfg.MaxIter) * 0.05
	varianceThreshold := float64(cfg.MaxIter) * 0.10

	return stats.Range < rangeThreshold || stats.StdDev < varianceThreshold
}

// SearchPass represents one pass of the spiral search
type SearchPass struct {
	NumCandidates int
	RadiusScale   float64
}

// FindInterestingPoint searches for a point with high variation/detail using multi-pass spiral search
func (ic *InterestCalculator) FindInterestingPoint(passes []SearchPass, cfg persistence.Config) (float64, float64) {
	bestX := cfg.CenterX
	bestY := cfg.CenterY
	bestScore := 0.0

	baseViewSize := 3.5 / cfg.Zoom

	for _, pass := range passes {
		searchRadius := baseViewSize * pass.RadiusScale
		for i := 0; i < pass.NumCandidates; i++ {
			angle := float64(i) * 2.399
			radius := searchRadius * math.Sqrt(float64(i)/float64(pass.NumCandidates))

			x := cfg.CenterX + radius*math.Cos(angle)
			y := cfg.CenterY + radius*math.Sin(angle)/2.0

			score := ic.CalculateScore(x, y, cfg)
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

// DescentSearchPasses returns a focused search configuration for auto-pilot descent mode.
// This searches much closer to the current center, preferring to zoom into local
// interesting details rather than jumping to distant features.
func DescentSearchPasses() []SearchPass {
	return []SearchPass{
		{NumCandidates: 30, RadiusScale: 0.15}, // Very close: focus on current feature
		{NumCandidates: 25, RadiusScale: 0.4},  // Nearby: slight exploration
		{NumCandidates: 20, RadiusScale: 0.8},  // Moderate: only if local area is boring
	}
}
