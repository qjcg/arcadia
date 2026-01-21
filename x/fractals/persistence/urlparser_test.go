package persistence

import (
	"testing"

	"github.com/qjcg/arcadia/x/fractals/colorthemes"
)

func TestParseStandardFractalURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		want    FractalURLParams
		wantErr bool
	}{
		{
			name: "basic mandelbrot",
			url:  "fractal://mandelbrot/-0.5/0.0/1.0/50/",
			want: FractalURLParams{
				Mode:        ModeStandard,
				FractalType: FractalMandelbrot,
				CenterX:     -0.5,
				CenterY:     0.0,
				Zoom:        1.0,
				MaxIter:     50,
				ColorTheme:  colorthemes.ColorGrayscale,
				Transition:  "none",
			},
			wantErr: false,
		},
		{
			name: "julia with parameters",
			url:  "fractal://julia/0.0/0.0/1.0/50/?julia_cr=-0.7&julia_ci=0.27015&color_theme=rainbow",
			want: FractalURLParams{
				Mode:        ModeStandard,
				FractalType: FractalJulia,
				CenterX:     0.0,
				CenterY:     0.0,
				Zoom:        1.0,
				MaxIter:     50,
				ColorTheme:  colorthemes.ColorRainbow,
				Transition:  "none",
				JuliaCr:     -0.7,
				JuliaCi:     0.27015,
			},
			wantErr: false,
		},
		{
			name: "with autopilot and dynamic color",
			url:  "fractal://mandelbrot/-0.5/0.0/1.0/50/?autopilot=t&dynamic_color=t",
			want: FractalURLParams{
				Mode:                ModeStandard,
				FractalType:         FractalMandelbrot,
				CenterX:             -0.5,
				CenterY:             0.0,
				Zoom:                1.0,
				MaxIter:             50,
				ColorTheme:          colorthemes.ColorGrayscale,
				Transition:          "none",
				AutopilotEnabled:    true,
				DynamicColorEnabled: true,
			},
			wantErr: false,
		},
		{
			name: "with transition",
			url:  "fractal://mandelbrot/-0.5/0.0/1.0/50/?transition=fade",
			want: FractalURLParams{
				Mode:        ModeStandard,
				FractalType: FractalMandelbrot,
				CenterX:     -0.5,
				CenterY:     0.0,
				Zoom:        1.0,
				MaxIter:     50,
				ColorTheme:  colorthemes.ColorGrayscale,
				Transition:  "fade",
			},
			wantErr: false,
		},
		{
			name:    "invalid fractal type",
			url:     "fractal://invalid/-0.5/0.0/1.0/50/",
			wantErr: true,
		},
		{
			name:    "invalid center_x",
			url:     "fractal://mandelbrot/abc/0.0/1.0/50/",
			wantErr: true,
		},
		{
			name:    "invalid zoom",
			url:     "fractal://mandelbrot/-0.5/0.0/abc/50/",
			wantErr: true,
		},
		{
			name:    "invalid max_iter",
			url:     "fractal://mandelbrot/-0.5/0.0/1.0/abc/",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseFractalURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseFractalURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}

			if got.Mode != tt.want.Mode {
				t.Errorf("Mode = %v, want %v", got.Mode, tt.want.Mode)
			}
			if got.FractalType != tt.want.FractalType {
				t.Errorf("FractalType = %v, want %v", got.FractalType, tt.want.FractalType)
			}
			if got.CenterX != tt.want.CenterX {
				t.Errorf("CenterX = %v, want %v", got.CenterX, tt.want.CenterX)
			}
			if got.CenterY != tt.want.CenterY {
				t.Errorf("CenterY = %v, want %v", got.CenterY, tt.want.CenterY)
			}
			if got.Zoom != tt.want.Zoom {
				t.Errorf("Zoom = %v, want %v", got.Zoom, tt.want.Zoom)
			}
			if got.MaxIter != tt.want.MaxIter {
				t.Errorf("MaxIter = %v, want %v", got.MaxIter, tt.want.MaxIter)
			}
			if got.ColorTheme != tt.want.ColorTheme {
				t.Errorf("ColorTheme = %v, want %v", got.ColorTheme, tt.want.ColorTheme)
			}
			if got.Transition != tt.want.Transition {
				t.Errorf("Transition = %v, want %v", got.Transition, tt.want.Transition)
			}
			if got.AutopilotEnabled != tt.want.AutopilotEnabled {
				t.Errorf("AutopilotEnabled = %v, want %v", got.AutopilotEnabled, tt.want.AutopilotEnabled)
			}
			if got.DynamicColorEnabled != tt.want.DynamicColorEnabled {
				t.Errorf("DynamicColorEnabled = %v, want %v", got.DynamicColorEnabled, tt.want.DynamicColorEnabled)
			}
		})
	}
}

func TestParseHyperrandURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
		checkFn func(*FractalURLParams) bool
	}{
		{
			name:    "basic random",
			url:     "fractal://random/",
			wantErr: false,
			checkFn: func(p *FractalURLParams) bool {
				return p.Mode == ModeRandom
			},
		},
		{
			name:    "random with color theme",
			url:     "fractal://random/?color_theme=rainbow",
			wantErr: false,
			checkFn: func(p *FractalURLParams) bool {
				return p.Mode == ModeRandom && p.ColorTheme == colorthemes.ColorRainbow
			},
		},
		{
			name:    "random with autopilot and dynamic color",
			url:     "fractal://random/?autopilot=t&dynamic_color=t",
			wantErr: false,
			checkFn: func(p *FractalURLParams) bool {
				return p.Mode == ModeRandom && p.AutopilotEnabled && p.DynamicColorEnabled
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseFractalURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseFractalURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}

			if !tt.checkFn(&got) {
				t.Errorf("Parsed URL does not match expected criteria: %+v", got)
			}
		})
	}
}

func TestParseBooleanParameters(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
		check   func(*FractalURLParams, error) bool
	}{
		{
			name:    "autopilot without value defaults to true",
			url:     "fractal://mandelbrot/-0.5/0.0/1.0/50/?autopilot",
			wantErr: false,
			check: func(p *FractalURLParams, err error) bool {
				return err == nil && p.AutopilotEnabled
			},
		},
		{
			name:    "autopilot=t sets to true",
			url:     "fractal://mandelbrot/-0.5/0.0/1.0/50/?autopilot=t",
			wantErr: false,
			check: func(p *FractalURLParams, err error) bool {
				return err == nil && p.AutopilotEnabled
			},
		},
		{
			name:    "autopilot=f sets to false",
			url:     "fractal://mandelbrot/-0.5/0.0/1.0/50/?autopilot=f",
			wantErr: false,
			check: func(p *FractalURLParams, err error) bool {
				return err == nil && !p.AutopilotEnabled
			},
		},
		{
			name:    "dynamic_color without value defaults to true",
			url:     "fractal://mandelbrot/-0.5/0.0/1.0/50/?dynamic_color",
			wantErr: false,
			check: func(p *FractalURLParams, err error) bool {
				return err == nil && p.DynamicColorEnabled
			},
		},
		{
			name:    "dynamic_color=t sets to true",
			url:     "fractal://mandelbrot/-0.5/0.0/1.0/50/?dynamic_color=t",
			wantErr: false,
			check: func(p *FractalURLParams, err error) bool {
				return err == nil && p.DynamicColorEnabled
			},
		},
		{
			name:    "dynamic_color=f sets to false",
			url:     "fractal://mandelbrot/-0.5/0.0/1.0/50/?dynamic_color=f",
			wantErr: false,
			check: func(p *FractalURLParams, err error) bool {
				return err == nil && !p.DynamicColorEnabled
			},
		},
		{
			name:    "both boolean params without values",
			url:     "fractal://mandelbrot/-0.5/0.0/1.0/50/?autopilot&dynamic_color",
			wantErr: false,
			check: func(p *FractalURLParams, err error) bool {
				return err == nil && p.AutopilotEnabled && p.DynamicColorEnabled
			},
		},
		{
			name:    "autopilot=invalid should error",
			url:     "fractal://mandelbrot/-0.5/0.0/1.0/50/?autopilot=invalid",
			wantErr: true,
			check: func(p *FractalURLParams, err error) bool {
				return err != nil
			},
		},
		{
			name:    "random with autopilot without value",
			url:     "fractal://random/?autopilot",
			wantErr: false,
			check: func(p *FractalURLParams, err error) bool {
				return err == nil && p.Mode == ModeRandom && p.AutopilotEnabled
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseFractalURL(tt.url)
			if !tt.check(&got, err) {
				t.Errorf("ParseFractalURL() = %+v, err = %v, check failed", got, err)
			}
		})
	}
}

func TestConfigToFractalURL(t *testing.T) {
	config := Config{
		FractalType: FractalMandelbrot,
		CenterX:     -0.5,
		CenterY:     0.0,
		Zoom:        1.0,
		MaxIter:     50,
		ColorScheme: colorthemes.ColorGrayscale,
		JuliaCr:     -0.7,
		JuliaCi:     0.27015,
	}

	url := ConfigToFractalURL(config, false, false, TransitionNone)

	// Parse it back and verify
	params, err := ParseFractalURL(url)
	if err != nil {
		t.Errorf("Failed to parse generated URL: %v", err)
		return
	}

	if params.FractalType != config.FractalType {
		t.Errorf("FractalType mismatch: got %v, want %v", params.FractalType, config.FractalType)
	}
	if params.CenterX != config.CenterX {
		t.Errorf("CenterX mismatch: got %v, want %v", params.CenterX, config.CenterX)
	}
	if params.Zoom != config.Zoom {
		t.Errorf("Zoom mismatch: got %v, want %v", params.Zoom, config.Zoom)
	}

	// Test with autopilot and dynamic color enabled
	url = ConfigToFractalURL(config, true, true, TransitionNone)
	if !contains(url, "autopilot=t") {
		t.Errorf("Generated URL should contain 'autopilot=t', got: %s", url)
	}
	if !contains(url, "dynamic_color=t") {
		t.Errorf("Generated URL should contain 'dynamic_color=t', got: %s", url)
	}

	// Parse and verify boolean flags
	params, err = ParseFractalURL(url)
	if err != nil {
		t.Errorf("Failed to parse generated URL: %v", err)
		return
	}
	if !params.AutopilotEnabled {
		t.Errorf("AutopilotEnabled should be true")
	}
	if !params.DynamicColorEnabled {
		t.Errorf("DynamicColorEnabled should be true")
	}
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestTransitionModeConversion(t *testing.T) {
	tests := []struct {
		name     string
		mode     int
		want     string
		backMode int
	}{
		{"none", TransitionNone, "none", TransitionNone},
		{"fade", TransitionFade, "fade", TransitionFade},
		{"zoomout", TransitionZoomOutIn, "zoomout", TransitionZoomOutIn},
		{"rotate", TransitionRotate, "rotate", TransitionRotate},
		{"breakthrough", TransitionBreakthrough, "breakthrough", TransitionBreakthrough},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			str := TransitionModeToString(tt.mode)
			if str != tt.want {
				t.Errorf("TransitionModeToString(%d) = %v, want %v", tt.mode, str, tt.want)
			}

			mode := StringToTransitionMode(str)
			if mode != tt.backMode {
				t.Errorf("StringToTransitionMode(%v) = %d, want %d", str, mode, tt.backMode)
			}
		})
	}
}
