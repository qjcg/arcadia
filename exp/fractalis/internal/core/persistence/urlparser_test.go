package persistence

import (
	"testing"

	"github.com/qjcg/arcadia/exp/fractalis/internal/core/pilot"
	"github.com/qjcg/arcadia/exp/fractalis/internal/tui/color"
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
				ColorTheme:  color.ColorGrayscale,
				Pilot:       pilot.ManualID,
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
				ColorTheme:  color.ColorRainbow,
				Pilot:       pilot.ManualID,
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
				ColorTheme:          color.ColorGrayscale,
				Pilot:               pilot.AutoZoomID,
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
				ColorTheme:  color.ColorGrayscale,
				Schedule:    pilot.Schedule{Transitions: []string{"fade"}, Traversal: pilot.Single},
				ScheduleSet: true,
				Pilot:       pilot.ManualID,
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
			if !schedulesEqual(got.Schedule, tt.want.Schedule) || got.ScheduleSet != tt.want.ScheduleSet {
				t.Errorf("Schedule = %+v (set %v), want %+v (set %v)",
					got.Schedule, got.ScheduleSet, tt.want.Schedule, tt.want.ScheduleSet)
			}
			if got.Pilot != tt.want.Pilot {
				t.Errorf("Pilot = %v, want %v", got.Pilot, tt.want.Pilot)
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
				return p.Mode == ModeRandom && p.ColorTheme == color.ColorRainbow
			},
		},
		{
			name:    "random with autopilot and dynamic color",
			url:     "fractal://random/?autopilot=t&dynamic_color=t",
			wantErr: false,
			checkFn: func(p *FractalURLParams) bool {
				return p.Mode == ModeRandom && p.Pilot == pilot.AutoZoomID && p.DynamicColorEnabled
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
				return err == nil && p.Pilot == pilot.AutoZoomID
			},
		},
		{
			name:    "autopilot=t sets to true",
			url:     "fractal://mandelbrot/-0.5/0.0/1.0/50/?autopilot=t",
			wantErr: false,
			check: func(p *FractalURLParams, err error) bool {
				return err == nil && p.Pilot == pilot.AutoZoomID
			},
		},
		{
			name:    "autopilot=f sets to false",
			url:     "fractal://mandelbrot/-0.5/0.0/1.0/50/?autopilot=f",
			wantErr: false,
			check: func(p *FractalURLParams, err error) bool {
				return err == nil && p.Pilot == pilot.ManualID
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
				return err == nil && p.Pilot == pilot.AutoZoomID && p.DynamicColorEnabled
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
				return err == nil && p.Mode == ModeRandom && p.Pilot == pilot.AutoZoomID
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
		ColorScheme: color.ColorGrayscale,
		JuliaCr:     -0.7,
		JuliaCi:     0.27015,
	}

	url := ConfigToFractalURL(config, pilot.ManualID, 1.05, pilot.DirectionIn, false, pilot.Schedule{})

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

	// Test with the Auto-zoom pilot and dynamic color enabled
	url = ConfigToFractalURL(config, pilot.AutoZoomID, 1.05, pilot.DirectionIn, true, pilot.Schedule{})
	if !contains(url, "pilot=autozoom") {
		t.Errorf("Generated URL should contain 'pilot=autozoom', got: %s", url)
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
	if params.Pilot != pilot.AutoZoomID {
		t.Errorf("Pilot should be autozoom")
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

func TestPilotParameter(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		want    pilot.ID
		wantErr bool
	}{
		{
			name: "pilot=manual",
			url:  "fractal://mandelbrot/-0.5/0.0/1.0/50/?pilot=manual",
			want: pilot.ManualID,
		},
		{
			name: "pilot=autozoom",
			url:  "fractal://mandelbrot/-0.5/0.0/1.0/50/?pilot=autozoom",
			want: pilot.AutoZoomID,
		},
		{
			name: "pilot=vantage",
			url:  "fractal://mandelbrot/-0.5/0.0/1.0/50/?pilot=vantage",
			want: pilot.VantageID,
		},
		{
			name: "default is the Manual pilot",
			url:  "fractal://mandelbrot/-0.5/0.0/1.0/50/",
			want: pilot.ManualID,
		},
		{
			name:    "unknown pilot errors",
			url:     "fractal://mandelbrot/-0.5/0.0/1.0/50/?pilot=warp",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseFractalURL(tt.url)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("ParseFractalURL() = %+v, want error", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseFractalURL() error: %v", err)
			}
			if got.Pilot != tt.want {
				t.Errorf("Pilot = %q, want %q", got.Pilot, tt.want)
			}
		})
	}
}

func TestAutopilotDeprecatedAlias(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		want    pilot.ID
		wantErr bool
	}{
		{name: "on maps to the Auto-zoom pilot", url: "fractal://mandelbrot/-0.5/0.0/1.0/50/?autopilot=on", want: pilot.AutoZoomID},
		{name: "off maps to the Manual pilot", url: "fractal://mandelbrot/-0.5/0.0/1.0/50/?autopilot=off", want: pilot.ManualID},
		{name: "legacy valueless form maps to the Auto-zoom pilot", url: "fractal://mandelbrot/-0.5/0.0/1.0/50/?autopilot", want: pilot.AutoZoomID},
		{name: "legacy t maps to the Auto-zoom pilot", url: "fractal://mandelbrot/-0.5/0.0/1.0/50/?autopilot=t", want: pilot.AutoZoomID},
		{name: "legacy f maps to the Manual pilot", url: "fractal://mandelbrot/-0.5/0.0/1.0/50/?autopilot=f", want: pilot.ManualID},
		{name: "invalid alias value errors", url: "fractal://mandelbrot/-0.5/0.0/1.0/50/?autopilot=invalid", wantErr: true},
		{name: "pilot= wins over the alias", url: "fractal://mandelbrot/-0.5/0.0/1.0/50/?pilot=manual&autopilot=on", want: pilot.ManualID},
		{name: "pilot= wins over the alias for vantage", url: "fractal://mandelbrot/-0.5/0.0/1.0/50/?autopilot=on&pilot=vantage", want: pilot.VantageID},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseFractalURL(tt.url)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("ParseFractalURL() = %+v, want error", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseFractalURL() error: %v", err)
			}
			if got.Pilot != tt.want {
				t.Errorf("Pilot = %q, want %q", got.Pilot, tt.want)
			}
		})
	}
}

func TestPilotRoundTrip(t *testing.T) {
	config := Config{
		FractalType: FractalMandelbrot,
		CenterX:     -0.5,
		CenterY:     0.0,
		Zoom:        1.0,
		MaxIter:     50,
		ColorScheme: color.ColorGrayscale,
	}

	for _, id := range pilot.IDs() {
		t.Run(string(id), func(t *testing.T) {
			url := ConfigToFractalURL(config, id, 1.05, pilot.DirectionIn, false, pilot.Schedule{})
			if !contains(url, "pilot="+string(id)) {
				t.Fatalf("generated URL %q should contain pilot=%s", url, id)
			}

			params, err := ParseFractalURL(url)
			if err != nil {
				t.Fatalf("Failed to parse generated URL: %v", err)
			}
			if params.Pilot != id {
				t.Errorf("Pilot = %q, want %q", params.Pilot, id)
			}
		})
	}
}

func TestAutoZoomPilotParameters(t *testing.T) {
	tests := []struct {
		name      string
		url       string
		wantSpeed float64
		wantDir   pilot.Direction
		wantErr   bool
	}{
		{
			name:      "defaults",
			url:       "fractal://mandelbrot/-0.5/0.0/1.0/50/",
			wantSpeed: 1.05,
			wantDir:   pilot.DirectionIn,
		},
		{
			name:      "pilot_speed",
			url:       "fractal://mandelbrot/-0.5/0.0/1.0/50/?pilot_speed=1.25",
			wantSpeed: 1.25,
			wantDir:   pilot.DirectionIn,
		},
		{
			name:      "pilot_direction=out",
			url:       "fractal://mandelbrot/-0.5/0.0/1.0/50/?pilot_direction=out",
			wantSpeed: 1.05,
			wantDir:   pilot.DirectionOut,
		},
		{
			name:      "both with the autozoom pilot",
			url:       "fractal://mandelbrot/-0.5/0.0/1.0/50/?pilot=autozoom&pilot_speed=1.5&pilot_direction=out",
			wantSpeed: 1.5,
			wantDir:   pilot.DirectionOut,
		},
		{name: "unparseable speed errors", url: "fractal://mandelbrot/-0.5/0.0/1.0/50/?pilot_speed=fast", wantErr: true},
		{name: "non-positive speed errors", url: "fractal://mandelbrot/-0.5/0.0/1.0/50/?pilot_speed=0", wantErr: true},
		{name: "unknown direction errors", url: "fractal://mandelbrot/-0.5/0.0/1.0/50/?pilot_direction=sideways", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseFractalURL(tt.url)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("ParseFractalURL() = %+v, want error", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseFractalURL() error: %v", err)
			}
			if got.PilotSpeed != tt.wantSpeed {
				t.Errorf("PilotSpeed = %v, want %v", got.PilotSpeed, tt.wantSpeed)
			}
			if got.PilotDirection != tt.wantDir {
				t.Errorf("PilotDirection = %v, want %v", got.PilotDirection, tt.wantDir)
			}
		})
	}
}

func TestAutoZoomParameterRoundTrip(t *testing.T) {
	config := Config{
		FractalType: FractalMandelbrot,
		CenterX:     -0.5,
		CenterY:     0.0,
		Zoom:        1.0,
		MaxIter:     50,
		ColorScheme: color.ColorGrayscale,
	}

	t.Run("steered parameters survive copy-URL", func(t *testing.T) {
		url := ConfigToFractalURL(config, pilot.AutoZoomID, 1.25, pilot.DirectionOut, false, pilot.Schedule{})
		if !contains(url, "pilot_speed=1.25") {
			t.Errorf("generated URL %q should contain pilot_speed=1.25", url)
		}
		if !contains(url, "pilot_direction=out") {
			t.Errorf("generated URL %q should contain pilot_direction=out", url)
		}

		params, err := ParseFractalURL(url)
		if err != nil {
			t.Fatalf("ParseFractalURL() error: %v", err)
		}
		if params.PilotSpeed != 1.25 || params.PilotDirection != pilot.DirectionOut {
			t.Errorf("round trip = (%v, %v), want (1.25, out)", params.PilotSpeed, params.PilotDirection)
		}
	})

	t.Run("default parameters stay implicit", func(t *testing.T) {
		url := ConfigToFractalURL(config, pilot.AutoZoomID, 1.05, pilot.DirectionIn, false, pilot.Schedule{})
		if contains(url, "pilot_speed") || contains(url, "pilot_direction") {
			t.Errorf("generated URL %q should omit default pilot parameters", url)
		}

		params, err := ParseFractalURL(url)
		if err != nil {
			t.Fatalf("ParseFractalURL() error: %v", err)
		}
		if params.PilotSpeed != 1.05 || params.PilotDirection != pilot.DirectionIn {
			t.Errorf("round trip = (%v, %v), want (1.05, in)", params.PilotSpeed, params.PilotDirection)
		}
	})
}
