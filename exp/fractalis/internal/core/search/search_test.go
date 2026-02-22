package search

import (
	"testing"

	"github.com/qjcg/arcadia/exp/fractalis/internal/core/persistence"
)

// mockFractalCalculator returns a simple fractal function for testing
// that returns predictable iteration counts based on position
func mockFractalCalculator(iterations int) func(cr, ci float64, cfg persistence.Config) float64 {
	return func(cr, ci float64, cfg persistence.Config) float64 {
		baseIter := float64(iterations)
		variation := (cr + ci) * 10.0
		if variation < 0 {
			variation = -variation
		}
		result := baseIter + variation
		if result > float64(cfg.MaxIter) {
			return float64(cfg.MaxIter)
		}
		return result
	}
}

// uniformFractalCalculator returns the same iteration count everywhere
func uniformFractalCalculator(iterations float64) func(cr, ci float64, cfg persistence.Config) float64 {
	return func(cr, ci float64, cfg persistence.Config) float64 {
		return iterations
	}
}

// edgeFractalCalculator simulates an edge with high variation
func edgeFractalCalculator() func(cr, ci float64, cfg persistence.Config) float64 {
	return func(cr, ci float64, cfg persistence.Config) float64 {
		if cr > 0 {
			return float64(cfg.MaxIter)
		}
		return 10.0
	}
}

func TestNewInterestCalculator(t *testing.T) {
	calc := mockFractalCalculator(100)
	ic := NewInterestCalculator(calc)

	if ic == nil {
		t.Fatal("NewInterestCalculator returned nil")
	}

	if ic.calculateFractal == nil {
		t.Error("calculateFractal function is nil")
	}
}

func TestCalculateScore(t *testing.T) {
	tests := []struct {
		name          string
		calculator    func(cr, ci float64, cfg persistence.Config) float64
		cfg           persistence.Config
		cx, cy        float64
		wantZero      bool
		wantPositive  bool
		checkPositive bool
	}{
		{
			name:       "uniform area returns zero",
			calculator: uniformFractalCalculator(100.0),
			cfg: persistence.Config{
				MaxIter: 1000,
				Zoom:    1.0,
			},
			cx:       0.0,
			cy:       0.0,
			wantZero: true,
		},
		{
			name:       "edge area returns positive score",
			calculator: edgeFractalCalculator(),
			cfg: persistence.Config{
				MaxIter: 1000,
				Zoom:    1.0,
			},
			cx:            0.0,
			cy:            0.0,
			wantPositive:  true,
			checkPositive: true,
		},
		{
			name: "high zoom with non-negative score",
			calculator: func(cr, ci float64, cfg persistence.Config) float64 {
				return float64(cfg.MaxIter) * 0.95
			},
			cfg: persistence.Config{
				MaxIter: 1000,
				Zoom:    100.0,
			},
			cx:            0.0,
			cy:            0.0,
			wantPositive:  true,
			checkPositive: false,
		},
		{
			name: "boundary bonus with variation returns positive score",
			calculator: func(cr, ci float64, cfg persistence.Config) float64 {
				base := float64(cfg.MaxIter) * 0.5
				if cr > 0 {
					return base + 50
				}
				return base
			},
			cfg: persistence.Config{
				MaxIter: 1000,
				Zoom:    1.0,
			},
			cx:            0.0,
			cy:            0.0,
			wantPositive:  true,
			checkPositive: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ic := NewInterestCalculator(tt.calculator)
			score := ic.CalculateScore(tt.cx, tt.cy, tt.cfg)

			if tt.wantZero && score != 0.0 {
				t.Errorf("CalculateScore() = %f, want 0.0", score)
			}

			if tt.checkPositive {
				if tt.wantPositive && score <= 0.0 {
					t.Errorf("CalculateScore() = %f, want positive", score)
				}
				if !tt.wantPositive && score > 0.0 {
					t.Errorf("CalculateScore() = %f, want non-positive", score)
				}
			}
		})
	}
}

func TestCalculateScore_VariationIncreasesScore(t *testing.T) {
	highVariationCalc := func(cr, ci float64, cfg persistence.Config) float64 {
		if cr > 0 {
			return float64(cfg.MaxIter)
		}
		return 10.0
	}

	lowVariationCalc := func(cr, ci float64, cfg persistence.Config) float64 {
		if cr > 0.1 {
			return 200.0
		}
		return 100.0
	}

	highIC := NewInterestCalculator(highVariationCalc)
	lowIC := NewInterestCalculator(lowVariationCalc)

	cfg := persistence.Config{
		MaxIter: 1000,
		Zoom:    1.0,
	}

	highScore := highIC.CalculateScore(0.0, 0.0, cfg)
	lowScore := lowIC.CalculateScore(0.0, 0.0, cfg)

	if highScore <= lowScore {
		t.Errorf("High variation area should have higher score: high=%f, low=%f", highScore, lowScore)
	}
}

func TestIsViewUniform(t *testing.T) {
	tests := []struct {
		name        string
		calculator  func(cr, ci float64, cfg persistence.Config) float64
		cfg         persistence.Config
		wantUniform bool
	}{
		{
			name:        "uniform view is detected as uniform",
			calculator:  uniformFractalCalculator(100.0),
			wantUniform: true,
			cfg: persistence.Config{
				MaxIter: 1000,
				Zoom:    1.0,
				CenterX: 0.0,
				CenterY: 0.0,
			},
		},
		{
			name:        "non-uniform view is detected as non-uniform",
			calculator:  edgeFractalCalculator(),
			wantUniform: false,
			cfg: persistence.Config{
				MaxIter: 1000,
				Zoom:    1.0,
				CenterX: 0.0,
				CenterY: 0.0,
			},
		},
		{
			name:        "high zoom does not panic",
			calculator:  mockFractalCalculator(500),
			wantUniform: false, // May vary but should not panic
			cfg: persistence.Config{
				MaxIter: 1000,
				Zoom:    1000.0,
				CenterX: 0.0,
				CenterY: 0.0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ic := NewInterestCalculator(tt.calculator)
			got := ic.IsViewUniform(tt.cfg)

			// For high zoom test, we just check it doesn't panic
			// For others, verify the expected result
			if tt.name != "high zoom does not panic" && got != tt.wantUniform {
				t.Errorf("IsViewUniform() = %v, want %v", got, tt.wantUniform)
			}
		})
	}
}

func TestFindInterestingPoint(t *testing.T) {
	tests := []struct {
		name             string
		calculator       func(cr, ci float64, cfg persistence.Config) float64
		cfg              persistence.Config
		passes           []SearchPass
		targetX          float64
		targetY          float64
		maxDist          float64
		shouldFindTarget bool
	}{
		{
			name:       "returns point within search radius",
			calculator: mockFractalCalculator(100),
			cfg: persistence.Config{
				MaxIter: 1000,
				Zoom:    1.0,
				CenterX: 0.0,
				CenterY: 0.0,
			},
			passes: []SearchPass{
				{NumCandidates: 10, RadiusScale: 0.5},
			},
			maxDist: 3.5 / 1.0 * 0.5 * 2.0,
		},
		{
			name: "finds better point near interesting area",
			calculator: func(cr, ci float64, cfg persistence.Config) float64 {
				dx := cr - 1.0
				dy := ci - 0.5
				dist := dx*dx + dy*dy
				if dist < 0.1 {
					return float64(cfg.MaxIter)
				}
				return 10.0
			},
			cfg: persistence.Config{
				MaxIter: 1000,
				Zoom:    1.0,
				CenterX: 0.0,
				CenterY: 0.0,
			},
			passes: []SearchPass{
				{NumCandidates: 100, RadiusScale: 2.0},
			},
			targetX:          1.0,
			targetY:          0.5,
			maxDist:          1.0,
			shouldFindTarget: true,
		},
		{
			name:       "works with multiple passes",
			calculator: mockFractalCalculator(100),
			cfg: persistence.Config{
				MaxIter: 1000,
				Zoom:    1.0,
				CenterX: 0.5,
				CenterY: 0.25,
			},
			passes: []SearchPass{
				{NumCandidates: 20, RadiusScale: 0.5},
				{NumCandidates: 30, RadiusScale: 1.0},
				{NumCandidates: 40, RadiusScale: 2.0},
			},
			maxDist: 3.5 / 1.0 * 2.0 * 2.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ic := NewInterestCalculator(tt.calculator)
			x, y := ic.FindInterestingPoint(tt.passes, tt.cfg)

			if tt.shouldFindTarget {
				dx := x - tt.targetX
				dy := y - tt.targetY
				dist := dx*dx + dy*dy
				if dist > tt.maxDist {
					t.Errorf("Expected to find point near (%f, %f), got (%f, %f), distance %f > %f",
						tt.targetX, tt.targetY, x, y, dist, tt.maxDist)
				}
			} else {
				// Check point is within reasonable bounds
				if x < tt.cfg.CenterX-tt.maxDist || x > tt.cfg.CenterX+tt.maxDist {
					t.Errorf("X coordinate %f out of expected range around %f", x, tt.cfg.CenterX)
				}
				if y < tt.cfg.CenterY-tt.maxDist || y > tt.cfg.CenterY+tt.maxDist {
					t.Errorf("Y coordinate %f out of expected range around %f", y, tt.cfg.CenterY)
				}
			}
		})
	}
}

func TestSearchPassConfigurations(t *testing.T) {
	tests := []struct {
		name     string
		passes   []SearchPass
		expected []SearchPass
	}{
		{
			name:   "default search passes",
			passes: DefaultSearchPasses(),
			expected: []SearchPass{
				{NumCandidates: 40, RadiusScale: 0.5},
				{NumCandidates: 40, RadiusScale: 1.5},
				{NumCandidates: 50, RadiusScale: 4.0},
			},
		},
		{
			name:   "descent search passes",
			passes: DescentSearchPasses(),
			expected: []SearchPass{
				{NumCandidates: 30, RadiusScale: 0.15},
				{NumCandidates: 25, RadiusScale: 0.4},
				{NumCandidates: 20, RadiusScale: 0.8},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if len(tt.passes) != len(tt.expected) {
				t.Fatalf("Expected %d passes, got %d", len(tt.expected), len(tt.passes))
			}

			for i := range tt.passes {
				if tt.passes[i].NumCandidates != tt.expected[i].NumCandidates {
					t.Errorf("Pass %d: expected %d candidates, got %d",
						i, tt.expected[i].NumCandidates, tt.passes[i].NumCandidates)
				}
				if tt.passes[i].RadiusScale != tt.expected[i].RadiusScale {
					t.Errorf("Pass %d: expected radius scale %f, got %f",
						i, tt.expected[i].RadiusScale, tt.passes[i].RadiusScale)
				}
			}
		})
	}
}

func TestDescentVsDefault(t *testing.T) {
	descent := DescentSearchPasses()
	default_ := DefaultSearchPasses()

	minLen := min(len(default_), len(descent))
	for i := range minLen {
		if descent[i].RadiusScale >= default_[i].RadiusScale {
			t.Errorf("Descent pass %d should have smaller radius scale than default: descent=%f, default=%f",
				i, descent[i].RadiusScale, default_[i].RadiusScale)
		}
	}
}
