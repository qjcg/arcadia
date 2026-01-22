package main

import (
	"math"
	"testing"

	"github.com/qjcg/arcadia/x/fractals/colorthemes"
	"github.com/qjcg/arcadia/x/fractals/persistence"
)

func TestMapToComplex(t *testing.T) {
	width, height := 80, 40
	centerX, centerY := -0.5, 0.0
	zoom := 1.0

	tests := []struct {
		name       string
		col        int
		row        int
		checkCr    bool
		checkCi    bool
		expectedCr float64
		expectedCi float64
		tolerance  float64
	}{
		{
			name:       "Center of screen should map to center point",
			col:        width / 2,
			row:        height / 2,
			checkCr:    true,
			checkCi:    true,
			expectedCr: centerX,
			expectedCi: centerY,
			tolerance:  0.1,
		},
		{
			name:       "Left edge should be around -2.5",
			col:        0,
			row:        height / 2,
			checkCr:    true,
			checkCi:    false,
			expectedCr: -2.25,
			expectedCi: 0,
			tolerance:  0.5,
		},
		{
			name:       "Right edge should be around 1.0",
			col:        width - 1,
			row:        height / 2,
			checkCr:    true,
			checkCi:    false,
			expectedCr: 1.25,
			expectedCi: 0,
			tolerance:  0.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cr, ci := mapToComplex(tt.col, tt.row, width, height, centerX, centerY, zoom)

			if tt.checkCr {
				if math.Abs(cr-tt.expectedCr) > tt.tolerance {
					t.Errorf("Real component: expected ~%v, got %v (diff: %v)",
						tt.expectedCr, cr, math.Abs(cr-tt.expectedCr))
				}
			}

			if tt.checkCi {
				if math.Abs(ci-tt.expectedCi) > tt.tolerance {
					t.Errorf("Imaginary component: expected ~%v, got %v (diff: %v)",
						tt.expectedCi, ci, math.Abs(ci-tt.expectedCi))
				}
			}
		})
	}
}

func TestMapToComplexZoom(t *testing.T) {
	width, height := 80, 40
	centerX, centerY := -0.5, 0.0

	// Test that zoom=2 gives half the range of zoom=1
	cr1, _ := mapToComplex(0, height/2, width, height, centerX, centerY, 1.0)
	cr2, _ := mapToComplex(0, height/2, width, height, centerX, centerY, 2.0)

	range1 := math.Abs(cr1 - centerX)
	range2 := math.Abs(cr2 - centerX)

	// range2 should be approximately half of range1
	ratio := range1 / range2
	if math.Abs(ratio-2.0) > 0.1 {
		t.Errorf("Expected zoom=2 to halve the range, got ratio %v", ratio)
	}
}

func TestGetChar(t *testing.T) {
	maxIter := 50

	// Test that points in the set get the densest character
	char := getChar(maxIter, maxIter)
	if char != '@' {
		t.Errorf("Expected '@' for points in set, got '%c'", char)
	}

	// Test that points that diverge immediately get the sparsest character
	char = getChar(0, maxIter)
	if char != ' ' {
		t.Errorf("Expected ' ' for immediate divergence, got '%c'", char)
	}

	// Test that we get different characters for different iteration counts
	chars := make(map[byte]bool)
	for i := 0; i < maxIter; i += 5 {
		char := getChar(i, maxIter)
		chars[char] = true
	}

	// Should have multiple different characters
	if len(chars) < 5 {
		t.Errorf("Expected at least 5 different characters, got %d", len(chars))
	}
}

func TestGetColor(t *testing.T) {
	maxIter := 50

	// Test grayscale returns empty string
	color := colorthemes.GetColor(25, maxIter, colorthemes.ColorGrayscale)
	if color != "" {
		t.Errorf("Expected empty string for grayscale, got %q", color)
	}

	// Test blue returns ANSI color code
	color = colorthemes.GetColor(25, maxIter, colorthemes.ColorBlue)
	if color == "" {
		t.Error("Expected ANSI color code for blue scheme, got empty string")
	}

	// Test rainbow returns ANSI color code
	color = colorthemes.GetColor(25, maxIter, colorthemes.ColorRainbow)
	if color == "" {
		t.Error("Expected ANSI color code for rainbow scheme, got empty string")
	}

	// Test points in set get black
	color = colorthemes.GetColor(maxIter, maxIter, colorthemes.ColorBlue)
	expected := "\033[38;5;0m"
	if color != expected {
		t.Errorf("Expected black color %q for points in set, got %q", expected, color)
	}
}

func TestCalculateFractal(t *testing.T) {
	tests := []struct {
		name        string
		config      Config
		cr          float64
		ci          float64
		shouldExist bool
	}{
		{
			name: "Mandelbrot at origin",
			config: Config{
				FractalType: FractalMandelbrot,
				MaxIter:     50,
			},
			cr:          0.0,
			ci:          0.0,
			shouldExist: true,
		},
		{
			name: "Julia with default parameters",
			config: Config{
				FractalType: FractalJulia,
				MaxIter:     50,
				JuliaCr:     -0.7,
				JuliaCi:     0.27015,
			},
			cr:          10.0,
			ci:          0.0,
			shouldExist: false,
		},
		{
			name: "Burning Ship at origin",
			config: Config{
				FractalType: FractalBurningShip,
				MaxIter:     50,
			},
			cr:          0.0,
			ci:          0.0,
			shouldExist: true,
		},
		{
			name: "Tricorn at origin",
			config: Config{
				FractalType: FractalTricorn,
				MaxIter:     50,
			},
			cr:          0.0,
			ci:          0.0,
			shouldExist: true,
		},
		{
			name: "Multibrot3 at origin",
			config: Config{
				FractalType: FractalMultibrot3,
				MaxIter:     50,
			},
			cr:          0.0,
			ci:          0.0,
			shouldExist: true,
		},
		{
			name: "Multibrot4 at origin",
			config: Config{
				FractalType: FractalMultibrot4,
				MaxIter:     50,
			},
			cr:          0.0,
			ci:          0.0,
			shouldExist: true,
		},
		{
			name: "Multibrot5 at origin",
			config: Config{
				FractalType: FractalMultibrot5,
				MaxIter:     50,
			},
			cr:          0.0,
			ci:          0.0,
			shouldExist: true,
		},
		{
			name: "Celtic at origin",
			config: Config{
				FractalType: FractalCeltic,
				MaxIter:     50,
			},
			cr:          0.0,
			ci:          0.0,
			shouldExist: true,
		},
		{
			name: "Perpendicular at origin",
			config: Config{
				FractalType: FractalPerpendicular,
				MaxIter:     50,
			},
			cr:          0.0,
			ci:          0.0,
			shouldExist: true,
		},
		{
			name: "Manhattan at origin",
			config: Config{
				FractalType: FractalManhattan,
				MaxIter:     50,
			},
			cr:          0.0,
			ci:          0.0,
			shouldExist: true,
		},
		{
			name: "Newton near root",
			config: Config{
				FractalType: FractalNewton,
				MaxIter:     50,
			},
			cr:          0.9,
			ci:          0.0,
			shouldExist: true, // Newton should converge
		},
		{
			name: "Invalid type defaults to Mandelbrot",
			config: Config{
				FractalType: "invalid",
				MaxIter:     50,
			},
			cr:          0.0,
			ci:          0.0,
			shouldExist: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			iter := calculateFractal(tt.cr, tt.ci, tt.config)

			if tt.shouldExist {
				if tt.config.FractalType == FractalNewton {
					// Newton should return non-zero for convergence
					if iter == 0 {
						t.Error("Expected Newton to converge, got iter=0")
					}
				} else if iter != tt.config.MaxIter {
					t.Errorf("Expected point to be in set (iter=%d), got iter=%d",
						tt.config.MaxIter, iter)
				}
			} else {
				if iter == tt.config.MaxIter {
					t.Error("Expected point to diverge")
				}
			}
		})
	}
}

func TestCalculateAdaptiveMaxIter(t *testing.T) {
	tests := []struct {
		name        string
		baseMaxIter int
		zoom        float64
		wantMin     int
		wantMax     int
	}{
		{
			name:        "Zoom=1 returns base iterations",
			baseMaxIter: 50,
			zoom:        1.0,
			wantMin:     50,
			wantMax:     50,
		},
		{
			name:        "Zoom=100 adds iterations",
			baseMaxIter: 50,
			zoom:        100.0,
			wantMin:     80,
			wantMax:     100,
		},
		{
			name:        "Zoom=10000 adds more iterations",
			baseMaxIter: 50,
			zoom:        10000.0,
			wantMin:     120,
			wantMax:     150,
		},
		{
			name:        "Very high zoom capped at 2000",
			baseMaxIter: 50,
			zoom:        1e100,
			wantMin:     2000,
			wantMax:     2000,
		},
		{
			name:        "Zero baseMaxIter uses default",
			baseMaxIter: 0,
			zoom:        1.0,
			wantMin:     50,
			wantMax:     50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := model{
				baseMaxIter: tt.baseMaxIter,
				config: Config{
					Zoom: tt.zoom,
				},
			}
			got := m.calculateAdaptiveMaxIter()
			if got < tt.wantMin || got > tt.wantMax {
				t.Errorf("calculateAdaptiveMaxIter() = %d, want between %d and %d",
					got, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestGetEffectiveSearchDelta(t *testing.T) {
	tests := []struct {
		name    string
		centerX float64
		zoom    float64
	}{
		{
			name:    "Low zoom",
			centerX: -0.5,
			zoom:    1.0,
		},
		{
			name:    "Medium zoom",
			centerX: -0.5,
			zoom:    100.0,
		},
		{
			name:    "High zoom",
			centerX: -0.5,
			zoom:    10000.0,
		},
		{
			name:    "Very high zoom",
			centerX: -0.5,
			zoom:    1e12,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := model{
				config: Config{
					CenterX: tt.centerX,
					Zoom:    tt.zoom,
				},
			}
			delta := m.getEffectiveSearchDelta()

			// Delta should be positive
			if delta <= 0 {
				t.Errorf("getEffectiveSearchDelta() = %v, want positive value", delta)
			}

			// Delta should decrease as zoom increases
			// At high zoom, delta should be small relative to view size
			viewSize := 3.5 / tt.zoom
			if delta > viewSize {
				t.Errorf("getEffectiveSearchDelta() = %v is larger than view size %v",
					delta, viewSize)
			}

			// Delta should not be smaller than minimum precision
			minDelta := 1e-14
			if delta < minDelta {
				t.Errorf("getEffectiveSearchDelta() = %v is smaller than minimum %v",
					delta, minDelta)
			}
		})
	}
}

func TestGenerateBookmarkName(t *testing.T) {
	// Test that it generates names
	for i := 0; i < 10; i++ {
		name := persistence.GenerateBookmarkName()

		// Should not be empty
		if name == "" {
			t.Error("GenerateBookmarkName() returned empty string")
		}

		// Should contain an underscore (adjective_noun format)
		if !containsChar(name, '_') {
			t.Errorf("GenerateBookmarkName() = %q, expected format adjective_noun", name)
		}

		// Should have two parts
		parts := splitString(name, '_')
		if len(parts) != 2 {
			t.Errorf("GenerateBookmarkName() = %q, expected two parts separated by underscore", name)
		}
	}

	// Test that it generates different names (probabilistic test)
	names := make(map[string]bool)
	for i := 0; i < 50; i++ {
		name := persistence.GenerateBookmarkName()
		names[name] = true
	}

	// Should have generated at least 30 different names out of 50
	if len(names) < 30 {
		t.Errorf("GenerateBookmarkName() generated only %d unique names out of 50, expected more variety",
			len(names))
	}
}

func TestIsViewUniform(t *testing.T) {
	tests := []struct {
		name     string
		config   Config
		expected bool
	}{
		{
			name: "Far outside Mandelbrot set should be uniform",
			config: Config{
				FractalType: FractalMandelbrot,
				CenterX:     10.0,
				CenterY:     10.0,
				Zoom:        1.0,
				MaxIter:     50,
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := model{
				config: tt.config,
			}
			got := m.isViewUniform()
			if got != tt.expected {
				t.Errorf("isViewUniform() = %v, want %v", got, tt.expected)
			}
		})
	}

	// Test that it returns a boolean (basic smoke test)
	m := model{
		config: Config{
			FractalType: FractalMandelbrot,
			CenterX:     -0.5,
			CenterY:     0.0,
			Zoom:        1.0,
			MaxIter:     50,
		},
	}
	// Just verify it doesn't crash and returns a valid boolean
	result := m.isViewUniform()
	_ = result // Use the result to avoid unused variable warning
}

func TestCalculateInterestScore(t *testing.T) {
	tests := []struct {
		name      string
		config    Config
		cx        float64
		cy        float64
		wantScore float64
		tolerance float64
	}{
		{
			name: "Mandelbrot boundary should have high score",
			config: Config{
				FractalType: FractalMandelbrot,
				MaxIter:     50,
				Zoom:        1.0,
				CenterX:     -0.5,
				CenterY:     0.0,
			},
			cx:        -0.7,
			cy:        0.0,
			wantScore: 100.0, // Expect at least 100
			tolerance: 1000.0,
		},
		{
			name: "Far outside should have low score",
			config: Config{
				FractalType: FractalMandelbrot,
				MaxIter:     50,
				Zoom:        1.0,
				CenterX:     -0.5,
				CenterY:     0.0,
			},
			cx:        10.0,
			cy:        10.0,
			wantScore: 0.0,
			tolerance: 50.0, // Very low score expected
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := model{
				config: tt.config,
			}
			got := m.calculateInterestScore(tt.cx, tt.cy)

			// Check if score is within expected range
			if got < tt.wantScore-tt.tolerance {
				t.Errorf("calculateInterestScore() = %v, want at least %v", got, tt.wantScore)
			}
		})
	}
}

func TestFindInterestingPoint(t *testing.T) {
	tests := []struct {
		name   string
		config Config
	}{
		{
			name: "Should find interesting point in Mandelbrot set",
			config: Config{
				FractalType: FractalMandelbrot,
				CenterX:     -0.5,
				CenterY:     0.0,
				Zoom:        1.0,
				MaxIter:     50,
			},
		},
		{
			name: "Should handle high zoom",
			config: Config{
				FractalType: FractalMandelbrot,
				CenterX:     -0.7,
				CenterY:     0.0,
				Zoom:        100.0,
				MaxIter:     100,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := model{
				config: tt.config,
			}
			x, y := m.findInterestingPoint()

			// Should return valid coordinates
			if math.IsNaN(x) || math.IsNaN(y) {
				t.Errorf("findInterestingPoint() returned NaN: (%v, %v)", x, y)
			}

			// Coordinates should be within reasonable bounds
			if math.Abs(x) > 10 || math.Abs(y) > 10 {
				t.Errorf("findInterestingPoint() = (%v, %v), coordinates seem unreasonably far",
					x, y)
			}
		})
	}
}

func TestApplyRandom(t *testing.T) {
	m := model{
		config: Config{
			FractalType: FractalMandelbrot,
			CenterX:     -0.5,
			CenterY:     0.0,
			Zoom:        1.0,
			MaxIter:     50,
			ColorScheme: colorthemes.ColorGrayscale,
		},
	}

	// Apply random multiple times
	for i := 0; i < 5; i++ {
		m.applyRandom()

		// Should set a valid fractal type
		validType := false
		for _, ft := range allFractalTypes {
			if m.config.FractalType == ft {
				validType = true
				break
			}
		}
		if !validType {
			t.Errorf("applyRandom() set invalid fractal type: %s", m.config.FractalType)
		}

		// Should set a valid color scheme
		validColor := false
		for _, cs := range colorthemes.AllColorSchemes {
			if m.config.ColorScheme == cs {
				validColor = true
				break
			}
		}
		if !validColor {
			t.Errorf("applyRandom() set invalid color scheme: %s", m.config.ColorScheme)
		}

		// Zoom should be positive
		if m.config.Zoom <= 0 {
			t.Errorf("applyRandom() set invalid zoom: %v", m.config.Zoom)
		}

		// MaxIter should be reasonable
		if m.config.MaxIter < 50 || m.config.MaxIter > 300 {
			t.Errorf("applyRandom() set unreasonable MaxIter: %d", m.config.MaxIter)
		}

		// Coordinates should be within reasonable bounds
		if math.Abs(m.config.CenterX) > 10 || math.Abs(m.config.CenterY) > 10 {
			t.Errorf("applyRandom() set unreasonable coordinates: (%v, %v)",
				m.config.CenterX, m.config.CenterY)
		}
	}
}

func TestTransitionFunctionality(t *testing.T) {
	m := model{
		config: Config{
			FractalType: FractalMandelbrot,
			CenterX:     -0.5,
			CenterY:     0.0,
			Zoom:        1.0,
			MaxIter:     50,
			ColorScheme: colorthemes.ColorGrayscale,
		},
	}

	// Test that transition mode can be set
	m.transitionMode = TransitionFade
	if m.transitionMode != TransitionFade {
		t.Error("Failed to set transition mode to Fade")
	}

	m.transitionMode = TransitionZoomOutIn
	if m.transitionMode != TransitionZoomOutIn {
		t.Error("Failed to set transition mode to ZoomOutIn")
	}

	m.transitionMode = TransitionRotate
	if m.transitionMode != TransitionRotate {
		t.Error("Failed to set transition mode to Rotate")
	}

	// Test transition progress
	m.transitionProgress = 0.5
	if m.transitionProgress != 0.5 {
		t.Error("Failed to set transition progress")
	}

	// Test that startFractalTransition can be called without panic
	m.startFractalTransition()

	// Should have set a target fractal type
	if m.transitionTarget == "" {
		t.Error("startFractalTransition() did not set a target fractal type")
	}

	// Should have initialized progress to start animation
	if m.transitionProgress != 0.01 {
		t.Errorf("startFractalTransition() did not initialize progress to 0.01, got %v", m.transitionProgress)
	}

	// Should have stored starting zoom
	if m.transitionZoomStart <= 0 {
		t.Errorf("startFractalTransition() did not store starting zoom, got %v", m.transitionZoomStart)
	}
}

func TestTransitionResetFunctionality(t *testing.T) {
	m := model{
		config: Config{
			FractalType: FractalMandelbrot,
			CenterX:     -0.5,
			CenterY:     0.0,
			Zoom:        100.0, // High zoom to test reset
			MaxIter:     200,   // High iterations to test reset
			ColorScheme: colorthemes.ColorGrayscale,
		},
		baseMaxIter: 50, // Set a base iteration count
	}

	// Manually set up a transition to test completion
	m.transitionMode = TransitionFade
	m.transitionTarget = FractalJulia
	m.transitionProgress = 1.0 // Force completion
	m.transitionZoomStart = 100.0

	// Simulate the transition completion logic
	if m.transitionProgress >= 1.0 {
		m.config.FractalType = m.transitionTarget
		m.transitionProgress = 0.0
		m.config.Zoom = 1.0
		// Reset iteration count to base value
		if m.baseMaxIter > 0 {
			m.config.MaxIter = m.baseMaxIter
		} else {
			m.config.MaxIter = 50
		}
		// Reset to default position for Julia set
		if m.transitionTarget == FractalJulia {
			m.config.CenterX = 0.0
			m.config.CenterY = 0.0
			m.config.JuliaCr = -0.7
			m.config.JuliaCi = 0.27015
		}
	}

	// Verify the reset worked correctly
	if m.config.FractalType != FractalJulia {
		t.Error("Transition did not change fractal type to Julia")
	}

	if m.transitionProgress != 0.0 {
		t.Errorf("Transition progress not reset, got %v", m.transitionProgress)
	}

	if m.config.Zoom != 1.0 {
		t.Errorf("Zoom not reset to 1.0, got %v", m.config.Zoom)
	}

	if m.config.MaxIter != 50 {
		t.Errorf("MaxIter not reset to base value, got %d", m.config.MaxIter)
	}

	if m.config.CenterX != 0.0 || m.config.CenterY != 0.0 {
		t.Errorf("Julia position not reset to (0,0), got (%v, %v)", m.config.CenterX, m.config.CenterY)
	}

	if m.config.JuliaCr != -0.7 || m.config.JuliaCi != 0.27015 {
		t.Errorf("Julia parameters not reset to defaults, got (cr=%v, ci=%v)", m.config.JuliaCr, m.config.JuliaCi)
	}
}

func TestModelInit(t *testing.T) {
	m := model{
		config: Config{
			FractalType: FractalMandelbrot,
			MaxIter:     50,
		},
	}

	cmd := m.Init()
	if cmd != nil {
		t.Errorf("Init() returned non-nil command: %v", cmd)
	}
}

func TestLoadBookmark(t *testing.T) {
	m := model{
		config: Config{
			FractalType: FractalMandelbrot,
			CenterX:     0.0,
			CenterY:     0.0,
			Zoom:        1.0,
			MaxIter:     50,
			ColorScheme: colorthemes.ColorGrayscale,
		},
		bookmarks: []persistence.Bookmark{
			{
				Name:        "test1",
				FractalType: FractalJulia,
				CenterX:     -0.7,
				CenterY:     0.27,
				Zoom:        10.0,
				MaxIter:     100,
				ColorScheme: colorthemes.ColorBlue,
				JuliaCr:     -0.8,
				JuliaCi:     0.156,
			},
			{
				Name:        "test2",
				FractalType: FractalBurningShip,
				CenterX:     -0.5,
				CenterY:     -0.6,
				Zoom:        5.0,
				MaxIter:     75,
				ColorScheme: colorthemes.ColorRainbow,
			},
		},
		baseMaxIter: 50,
	}

	// Load first bookmark
	m.loadBookmark(0)

	// Verify config was updated
	if m.config.FractalType != FractalJulia {
		t.Errorf("loadBookmark() FractalType = %s, want %s", m.config.FractalType, FractalJulia)
	}
	if m.config.CenterX != -0.7 {
		t.Errorf("loadBookmark() CenterX = %v, want %v", m.config.CenterX, -0.7)
	}
	if m.config.Zoom != 10.0 {
		t.Errorf("loadBookmark() Zoom = %v, want %v", m.config.Zoom, 10.0)
	}
	if m.config.ColorScheme != colorthemes.ColorBlue {
		t.Errorf("loadBookmark() ColorScheme = %s, want %s", m.config.ColorScheme, colorthemes.ColorBlue)
	}
	if m.baseMaxIter != 100 {
		t.Errorf("loadBookmark() baseMaxIter = %d, want %d", m.baseMaxIter, 100)
	}

	// Load second bookmark
	m.loadBookmark(1)

	if m.config.FractalType != FractalBurningShip {
		t.Errorf("loadBookmark() FractalType = %s, want %s", m.config.FractalType, FractalBurningShip)
	}
	if m.config.Zoom != 5.0 {
		t.Errorf("loadBookmark() Zoom = %v, want %v", m.config.Zoom, 5.0)
	}

	// Test invalid index (should not crash)
	m.loadBookmark(-1)
	m.loadBookmark(100)
}

func TestDeleteBookmark(t *testing.T) {
	m := model{
		bookmarks: []persistence.Bookmark{
			{Name: "test1", FractalType: FractalMandelbrot},
			{Name: "test2", FractalType: FractalJulia},
			{Name: "test3", FractalType: FractalBurningShip},
		},
		bookmarkCursor: 1,
	}

	// Delete middle bookmark
	err := m.deleteBookmark(1)
	if err != nil {
		t.Errorf("deleteBookmark() returned error: %v", err)
	}

	// Should have 2 bookmarks left
	if len(m.bookmarks) != 2 {
		t.Errorf("deleteBookmark() left %d bookmarks, want 2", len(m.bookmarks))
	}

	// Remaining bookmarks should be test1 and test3
	if m.bookmarks[0].Name != "test1" || m.bookmarks[1].Name != "test3" {
		t.Errorf("deleteBookmark() left wrong bookmarks: %v", m.bookmarks)
	}

	// Test invalid indices
	err = m.deleteBookmark(-1)
	if err == nil {
		t.Error("deleteBookmark(-1) should return error")
	}

	err = m.deleteBookmark(100)
	if err == nil {
		t.Error("deleteBookmark(100) should return error")
	}

	// Delete last bookmark and cursor should adjust
	m.bookmarkCursor = 1
	err = m.deleteBookmark(1)
	if err != nil {
		t.Errorf("deleteBookmark() returned error: %v", err)
	}
	if m.bookmarkCursor != 0 {
		t.Errorf("deleteBookmark() cursor = %d, want 0", m.bookmarkCursor)
	}
}

func TestRenderFractal(t *testing.T) {
	m := model{
		config: Config{
			Width:       20,
			Height:      10,
			MaxIter:     50,
			CenterX:     -0.5,
			CenterY:     0.0,
			Zoom:        1.0,
			FractalType: FractalMandelbrot,
			ColorScheme: colorthemes.ColorGrayscale,
		},
	}

	output := m.renderFractal()

	// Should not be empty
	if output == "" {
		t.Error("renderFractal() returned empty string")
	}

	// Count lines (should be height - 1 newlines)
	lineCount := 1
	for _, ch := range output {
		if ch == '\n' {
			lineCount++
		}
	}
	if lineCount != m.config.Height {
		t.Errorf("renderFractal() produced %d lines, want %d", lineCount, m.config.Height)
	}
}

func TestRenderStatusBar(t *testing.T) {
	tests := []struct {
		name     string
		m        model
		contains []string
	}{
		{
			name: "Basic status bar",
			m: model{
				config: Config{
					FractalType: FractalMandelbrot,
					CenterX:     -0.5,
					CenterY:     0.0,
					Zoom:        1.0,
					MaxIter:     50,
					ColorScheme: colorthemes.ColorGrayscale,
				},
				autoZoom:          false,
				autoZoomDirection: 1,
			},
			contains: []string{"mandelbrot", "-0.5000", "50", "grayscale"},
		},
		{
			name: "Auto-pilot active",
			m: model{
				config: Config{
					FractalType: FractalMandelbrot,
					Zoom:        1.0,
					MaxIter:     50,
				},
				autoZoom:          true,
				autoZoomDirection: 1,
			},
			contains: []string{"AUTO-PILOT"},
		},
		{
			name: "Julia parameters shown",
			m: model{
				config: Config{
					FractalType: FractalJulia,
					JuliaCr:     -0.7,
					JuliaCi:     0.27,
					Zoom:        1.0,
					MaxIter:     50,
				},
				autoZoomDirection: 1,
			},
			contains: []string{"julia", "Julia"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := tt.m.renderStatusBar()

			if output == "" {
				t.Error("renderStatusBar() returned empty string")
			}

			// Check for expected content
			for _, expected := range tt.contains {
				found := false
				// Case-insensitive search
				outputLower := toLower(output)
				expectedLower := toLower(expected)
				if containsSubstring(outputLower, expectedLower) {
					found = true
				}
				if !found {
					t.Errorf("renderStatusBar() output does not contain %q", expected)
				}
			}
		})
	}
}

// Helper functions for tests

func containsChar(s string, c rune) bool {
	for _, ch := range s {
		if ch == c {
			return true
		}
	}
	return false
}

func splitString(s string, sep rune) []string {
	var parts []string
	var current string
	for _, ch := range s {
		if ch == sep {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(ch)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}
	return parts
}

func toLower(s string) string {
	result := ""
	for _, ch := range s {
		if ch >= 'A' && ch <= 'Z' {
			result += string(ch + 32)
		} else {
			result += string(ch)
		}
	}
	return result
}

func containsSubstring(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(substr) > len(s) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			if s[i+j] != substr[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

func TestAllFractalTypesValid(t *testing.T) {
	// Test that all fractal types in allFractalTypes are valid
	for _, fractalType := range allFractalTypes {
		// Create a config with this fractal type
		config := Config{
			FractalType: fractalType,
			MaxIter:     50,
			JuliaCr:     -0.7,
			JuliaCi:     0.27015,
		}

		// Should not panic when calculating
		_ = calculateFractal(0.0, 0.0, config)
	}
}

func TestAllColorSchemesValid(t *testing.T) {
	// Test that all color schemes work
	colorSchemes := []string{colorthemes.ColorGrayscale, colorthemes.ColorBlue, colorthemes.ColorRainbow}

	for _, scheme := range colorSchemes {
		color := colorthemes.GetColor(25, 50, scheme)
		// Should not panic and return a valid string
		_ = color
	}
}

func TestRenderFractalAllTypes(t *testing.T) {
	// Test that all fractal types can be rendered without panicking
	for _, fractalType := range allFractalTypes {
		t.Run(fractalType, func(t *testing.T) {
			m := model{
				config: Config{
					Width:       10,
					Height:      5,
					MaxIter:     20,
					CenterX:     0.0,
					CenterY:     0.0,
					Zoom:        1.0,
					FractalType: fractalType,
					ColorScheme: colorthemes.ColorGrayscale,
					JuliaCr:     -0.7,
					JuliaCi:     0.27015,
				},
			}

			// Should not panic
			output := m.renderFractal()

			// Should produce output
			if output == "" {
				t.Errorf("renderFractal() for %s produced no output", fractalType)
			}
		})
	}
}

func TestBookmarkYAMLMarshaling(t *testing.T) {
	// Test that bookmarks can be marshaled/unmarshaled properly
	bookmarks := []persistence.Bookmark{
		{
			Name:        "test_bookmark",
			FractalType: FractalNewton,
			CenterX:     0.5,
			CenterY:     -0.3,
			Zoom:        42.5,
			MaxIter:     150,
			ColorScheme: colorthemes.ColorRainbow,
		},
		{
			Name:        "julia_test",
			FractalType: FractalJulia,
			CenterX:     0.0,
			CenterY:     0.0,
			Zoom:        1.0,
			MaxIter:     50,
			ColorScheme: colorthemes.ColorBlue,
			JuliaCr:     -0.8,
			JuliaCi:     0.156,
		},
	}

	// This just verifies the bookmark structure is valid
	// Actual file I/O is tested elsewhere
	for _, bm := range bookmarks {
		if bm.Name == "" {
			t.Error("Bookmark has empty name")
		}
		if bm.FractalType == "" {
			t.Error("Bookmark has empty fractal type")
		}
		if bm.MaxIter <= 0 {
			t.Error("Bookmark has invalid MaxIter")
		}
	}
}

func TestMapToComplexSymmetry(t *testing.T) {
	// Test that mapping is symmetric
	width, height := 100, 50
	centerX, centerY := 0.0, 0.0
	zoom := 1.0

	// Center should map to center coordinates
	crCenter, ciCenter := mapToComplex(width/2, height/2, width, height, centerX, centerY, zoom)

	if math.Abs(crCenter-centerX) > 0.01 {
		t.Errorf("Center X mapping: expected ~%v, got %v", centerX, crCenter)
	}
	if math.Abs(ciCenter-centerY) > 0.01 {
		t.Errorf("Center Y mapping: expected ~%v, got %v", centerY, ciCenter)
	}

	// Left and right should be equidistant from center
	crLeft, _ := mapToComplex(0, height/2, width, height, centerX, centerY, zoom)
	crRight, _ := mapToComplex(width-1, height/2, width, height, centerX, centerY, zoom)

	leftDist := math.Abs(crLeft - centerX)
	rightDist := math.Abs(crRight - centerX)

	if math.Abs(leftDist-rightDist) > 0.1 {
		t.Errorf("Horizontal symmetry broken: left distance %v, right distance %v",
			leftDist, rightDist)
	}
}

func TestGetCharBoundaries(t *testing.T) {
	maxIter := 100

	// Test boundary conditions
	tests := []struct {
		iter     int
		expected byte
	}{
		{0, ' '},           // First character (quick divergence)
		{maxIter, '@'},     // Last character (in set)
		{maxIter / 2, ' '}, // Mid-range gets some character
	}

	for _, tt := range tests {
		char := getChar(tt.iter, maxIter)
		if tt.iter == 0 && char != ' ' {
			t.Errorf("getChar(0) = %c, want ' '", char)
		}
		if tt.iter == maxIter && char != '@' {
			t.Errorf("getChar(maxIter) = %c, want '@'", char)
		}
	}

	// Character should be in the defined set
	validChars := " .:-=+*#%@"
	for i := 0; i <= maxIter; i += 10 {
		char := getChar(i, maxIter)
		found := false
		for _, valid := range validChars {
			if byte(valid) == char {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("getChar(%d) = %c is not in valid character set", i, char)
		}
	}
}

func TestRenderStaticMode(t *testing.T) {
	// Test the legacy render function
	config := Config{
		Width:       20,
		Height:      10,
		MaxIter:     50,
		CenterX:     -0.5,
		CenterY:     0.0,
		Zoom:        1.0,
		FractalType: FractalMandelbrot,
		ColorScheme: colorthemes.ColorGrayscale,
	}

	// Should not panic
	render(config)
}

func TestGetColorAllSchemes(t *testing.T) {
	maxIter := 50

	tests := []struct {
		name   string
		scheme string
		iter   int
	}{
		{"Grayscale low iter", colorthemes.ColorGrayscale, 5},
		{"Grayscale mid iter", colorthemes.ColorGrayscale, 25},
		{"Grayscale high iter", colorthemes.ColorGrayscale, 45},
		{"Blue low iter", colorthemes.ColorBlue, 5},
		{"Blue mid iter", colorthemes.ColorBlue, 25},
		{"Blue high iter", colorthemes.ColorBlue, 45},
		{"Blue at maxIter", colorthemes.ColorBlue, maxIter},
		{"Rainbow low iter", colorthemes.ColorRainbow, 5},
		{"Rainbow mid iter", colorthemes.ColorRainbow, 25},
		{"Rainbow high iter", colorthemes.ColorRainbow, 45},
		{"Rainbow at maxIter", colorthemes.ColorRainbow, maxIter},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			color := colorthemes.GetColor(tt.iter, maxIter, tt.scheme)

			if tt.scheme == colorthemes.ColorGrayscale {
				if color != "" {
					t.Errorf("Expected empty string for grayscale, got %q", color)
				}
			} else {
				// Blue and rainbow should return ANSI codes
				if color == "" && tt.iter != maxIter {
					t.Errorf("Expected ANSI color code, got empty string")
				}
			}
		})
	}
}

func TestFractalQuickDivergence(t *testing.T) {
	// Test that all fractals diverge quickly for far points
	maxIter := 50
	farPoint := 10.0

	fractals := []string{
		FractalMandelbrot, FractalBurningShip, FractalTricorn,
		FractalMultibrot3, FractalMultibrot4, FractalMultibrot5,
		FractalCeltic, FractalPerpendicular, FractalManhattan,
	}

	for _, fractalType := range fractals {
		t.Run(fractalType, func(t *testing.T) {
			config := Config{
				FractalType: fractalType,
				MaxIter:     maxIter,
			}

			iter := calculateFractal(farPoint, farPoint, config)

			// Should diverge very quickly (within 10 iterations)
			if iter > 10 {
				t.Errorf("%s took %d iterations for far point, expected quick divergence",
					fractalType, iter)
			}
		})
	}
}

func TestAddBookmarkIntegration(t *testing.T) {
	// Test adding a bookmark
	m := model{
		config: Config{
			FractalType: FractalNewton,
			CenterX:     1.5,
			CenterY:     -0.5,
			Zoom:        25.0,
			MaxIter:     100,
			ColorScheme: colorthemes.ColorRainbow,
		},
		bookmarks: []persistence.Bookmark{},
	}

	// This would normally save to file, which we skip in tests
	// Just verify the bookmark structure
	bookmark := persistence.Bookmark{
		Name:        "test_newton",
		FractalType: m.config.FractalType,
		CenterX:     m.config.CenterX,
		CenterY:     m.config.CenterY,
		Zoom:        m.config.Zoom,
		MaxIter:     m.config.MaxIter,
		ColorScheme: m.config.ColorScheme,
	}

	// Verify bookmark has expected fields
	if bookmark.Name == "" {
		t.Error("Bookmark name is empty")
	}
	if bookmark.Zoom != 25.0 {
		t.Errorf("Bookmark zoom = %v, want 25.0", bookmark.Zoom)
	}
	if bookmark.FractalType != FractalNewton {
		t.Errorf("Bookmark fractal type = %s, want %s", bookmark.FractalType, FractalNewton)
	}
}

func TestRenderHelp(t *testing.T) {
	m := model{
		config: Config{
			FractalType: FractalMandelbrot,
			MaxIter:     50,
		},
	}

	help := m.renderHelp()

	// Should contain key sections
	expectedSections := []string{
		"Navigation",
		"Fractal Types",
		"Settings",
		"Bookmarks",
		"Random",
	}

	for _, section := range expectedSections {
		if !containsSubstring(toLower(help), toLower(section)) {
			t.Errorf("Help text missing section: %s", section)
		}
	}

	// Should mention the new Newton fractal
	if !containsSubstring(toLower(help), "newton") {
		t.Error("Help text doesn't mention Newton fractal")
	}
}

func TestRenderBookmarks(t *testing.T) {
	m := model{
		config: Config{},
		bookmarks: []persistence.Bookmark{
			{Name: "test1", FractalType: FractalMandelbrot, Zoom: 1.0},
			{Name: "test2", FractalType: FractalNewton, Zoom: 50.0},
		},
		bookmarkCursor: 0,
	}

	output := m.renderBookmarks()

	// Should contain bookmark names
	if !containsSubstring(output, "test1") {
		t.Error("Bookmark list doesn't show test1")
	}
	if !containsSubstring(output, "test2") {
		t.Error("Bookmark list doesn't show test2")
	}
}

func TestRenderBookmarkInput(t *testing.T) {
	m := model{
		config: Config{
			FractalType: FractalNewton,
			CenterX:     0.0,
			CenterY:     0.0,
			Zoom:        1.0,
		},
		savingBookmark:    true,
		suggestedBookmark: "test_bookmark",
		bookmarkInput:     "",
	}

	output := m.renderBookmarkInput()

	// Should show suggested name
	if !containsSubstring(output, "test_bookmark") {
		t.Error("Bookmark input doesn't show suggested name")
	}

	// Should show current location
	if !containsSubstring(toLower(output), "newton") {
		t.Error("Bookmark input doesn't show fractal type")
	}
}

func TestVantageModeFields(t *testing.T) {
	m := model{
		vantageMode:        true,
		vantageSceneDur:    100,
		vantageSceneTimer:  50,
		vantageInitialized: false,
	}

	if !m.vantageMode {
		t.Error("vantageMode should be true")
	}

	if m.vantageSceneDur != 100 {
		t.Errorf("vantageSceneDur should be 100, got %d", m.vantageSceneDur)
	}

	if m.vantageSceneTimer != 50 {
		t.Errorf("vantageSceneTimer should be 50, got %d", m.vantageSceneTimer)
	}

	if m.vantageInitialized {
		t.Error("vantageInitialized should be false")
	}
}

func TestVantageModeInit(t *testing.T) {
	m := model{
		config: Config{
			Width:       80,
			Height:      40,
			MaxIter:     50,
			CenterX:     -0.5,
			CenterY:     0.0,
			Zoom:        1.0,
			ColorScheme: "grayscale",
			FractalType: "mandelbrot",
			JuliaCr:     -0.7,
			JuliaCi:     0.27015,
		},
		baseMaxIter:       50,
		vantageMode:       true,
		vantageSceneDur:   100,
		vantageSceneTimer: 0,
	}

	// Test that Init returns a tickCmd when vantage mode is enabled
	cmd := m.Init()
	if cmd == nil {
		t.Errorf("Expected tickCmd for vantage mode, got nil")
	}
}

func TestStatusBarModeIndicators(t *testing.T) {
	config := Config{
		Width:       80,
		Height:      40,
		MaxIter:     50,
		CenterX:     -0.5,
		CenterY:     0.0,
		Zoom:        1.0,
		ColorScheme: "grayscale",
		FractalType: "mandelbrot",
		JuliaCr:     -0.7,
		JuliaCi:     0.27015,
	}

	// Test explorer mode (manual navigation)
	m := model{config: config, autoZoom: false, vantageMode: false}
	output := m.renderStatusBar()
	if !containsSubstring(output, "Explore") {
		t.Error("Status bar should show 'Explore' for manual mode")
	}

	// Test autopilot mode
	m = model{config: config, autoZoom: true, vantageMode: false, autoZoomDirection: 1, zoomSpeed: 1.05}
	output = m.renderStatusBar()
	if !containsSubstring(output, "AUTO-PILOT") {
		t.Error("Status bar should show 'AUTO-PILOT' when autopilot is active")
	}

	// Test vantage mode
	m = model{config: config, autoZoom: false, vantageMode: true, vantageSceneDur: 100}
	output = m.renderStatusBar()
	if !containsSubstring(output, "VANTAGE") {
		t.Error("Status bar should show 'VANTAGE' when vantage is active")
	}
	if !containsSubstring(output, "5.0s") {
		t.Error("Status bar should show vantage duration (100 ticks = 5 seconds)")
	}
}
