package persistence

import (
	"testing"

	"github.com/qjcg/arcadia/exp/fractalis/internal/core/pilot"
	"github.com/qjcg/arcadia/exp/fractalis/internal/core/transition"
	"github.com/qjcg/arcadia/exp/fractalis/internal/tui/color"
)

func TestParseScheduleGrammar(t *testing.T) {
	allRandom := pilot.Schedule{Transitions: transition.EffectIDs(), Traversal: pilot.Random}

	tests := []struct {
		name    string
		input   string
		want    pilot.Schedule
		wantErr bool
	}{
		{name: "none", input: "none", want: pilot.Schedule{}},
		{name: "empty is none", input: "", want: pilot.Schedule{}},
		{name: "single effect", input: "fade", want: pilot.Schedule{
			Transitions: []string{"fade"}, Traversal: pilot.Single,
		}},
		{name: "legacy zoom_out spelling", input: "zoom_out", want: pilot.Schedule{
			Transitions: []string{"zoomout"}, Traversal: pilot.Single,
		}},
		{name: "comma list is sequential", input: "fade,rotate", want: pilot.Schedule{
			Transitions: []string{"fade", "rotate"}, Traversal: pilot.Sequential,
		}},
		{name: "random over all effects", input: "random", want: allRandom},
		{name: "random over a subset", input: "random:fade,zoomout", want: pilot.Schedule{
			Transitions: []string{"fade", "zoomout"}, Traversal: pilot.Random,
		}},
		{name: "unknown effect", input: "warp", wantErr: true},
		{name: "unknown effect in a list", input: "fade,warp", wantErr: true},
		{name: "none is not an effect", input: "none,fade", wantErr: true},
		{name: "random needs a subset", input: "random:", wantErr: true},
		{name: "trailing comma", input: "fade,", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseSchedule(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("ParseSchedule(%q) = %+v, want error", tt.input, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseSchedule(%q) error: %v", tt.input, err)
			}
			if !schedulesEqual(got, tt.want) {
				t.Errorf("ParseSchedule(%q) = %+v, want %+v", tt.input, got, tt.want)
			}
		})
	}
}

func TestScheduleStringRendersTheGrammar(t *testing.T) {
	tests := []struct {
		name     string
		schedule pilot.Schedule
		want     string
	}{
		{name: "empty is none", schedule: pilot.Schedule{}, want: "none"},
		{name: "single effect", schedule: pilot.Schedule{
			Transitions: []string{"fade"}, Traversal: pilot.Single,
		}, want: "fade"},
		{name: "comma list is sequential", schedule: pilot.Schedule{
			Transitions: []string{"fade", "rotate"}, Traversal: pilot.Sequential,
		}, want: "fade,rotate"},
		{name: "random over all effects", schedule: pilot.Schedule{
			Transitions: transition.EffectIDs(), Traversal: pilot.Random,
		}, want: "random"},
		{name: "random over a subset", schedule: pilot.Schedule{
			Transitions: []string{"fade", "zoomout"}, Traversal: pilot.Random,
		}, want: "random:fade,zoomout"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ScheduleString(tt.schedule); got != tt.want {
				t.Errorf("ScheduleString(%+v) = %q, want %q", tt.schedule, got, tt.want)
			}
		})
	}
}

func TestScheduleGrammarRoundTrips(t *testing.T) {
	// Every canonical spelling parses and renders back to itself.
	for _, spelling := range []string{
		"none", "fade", "zoomout", "rotate", "breakthrough", "starwipe",
		"fade,rotate,zoomout", "random", "random:fade,zoomout",
	} {
		schedule, err := ParseSchedule(spelling)
		if err != nil {
			t.Fatalf("ParseSchedule(%q) error: %v", spelling, err)
		}
		if got := ScheduleString(schedule); got != spelling {
			t.Errorf("round trip of %q rendered %q", spelling, got)
		}
	}
}

func schedulesEqual(a, b pilot.Schedule) bool {
	if a.Traversal != b.Traversal || len(a.Transitions) != len(b.Transitions) {
		return false
	}
	for i := range a.Transitions {
		if a.Transitions[i] != b.Transitions[i] {
			return false
		}
	}
	return true
}

func TestTransitionParamCarriesTheScheduleGrammar(t *testing.T) {
	tests := []struct {
		name      string
		url       string
		want      pilot.Schedule
		wantSet   bool
		wantError bool
	}{
		{
			name: "sequential list",
			url:  "fractal://mandelbrot/-0.5/0.0/1.0/50/?transition=fade,zoomout",
			want: pilot.Schedule{
				Transitions: []string{"fade", "zoomout"}, Traversal: pilot.Sequential,
			},
			wantSet: true,
		},
		{
			name:    "absent leaves no schedule opinion",
			url:     "fractal://mandelbrot/-0.5/0.0/1.0/50/",
			want:    pilot.Schedule{},
			wantSet: false,
		},
		{
			name:    "explicit none is an empty schedule",
			url:     "fractal://mandelbrot/-0.5/0.0/1.0/50/?transition=none",
			want:    pilot.Schedule{},
			wantSet: true,
		},
		{
			name:      "unknown effect errors",
			url:       "fractal://mandelbrot/-0.5/0.0/1.0/50/?transition=warp",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseFractalURL(tt.url)
			if tt.wantError {
				if err == nil {
					t.Fatalf("ParseFractalURL(%q) = %+v, want error", tt.url, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseFractalURL(%q) error: %v", tt.url, err)
			}
			if !schedulesEqual(got.Schedule, tt.want) {
				t.Errorf("Schedule = %+v, want %+v", got.Schedule, tt.want)
			}
			if got.ScheduleSet != tt.wantSet {
				t.Errorf("ScheduleSet = %v, want %v", got.ScheduleSet, tt.wantSet)
			}
		})
	}
}

func TestCopyURLRoundTripsTheTransitionSchedule(t *testing.T) {
	config := Config{
		FractalType: FractalMandelbrot,
		CenterX:     -0.5,
		CenterY:     0.0,
		Zoom:        1.0,
		MaxIter:     50,
		ColorScheme: color.ColorGrayscale,
	}

	builders := map[string]func(pilot.Schedule) string{
		"ConfigToFractalURL": func(s pilot.Schedule) string {
			return ConfigToFractalURL(config, pilot.AutoZoomID, false, s)
		},
		"RandomToFractalURL": func(s pilot.Schedule) string {
			return RandomToFractalURL("", pilot.AutoZoomID, false, s)
		},
	}
	for _, spelling := range []string{
		"none", "fade", "fade,rotate", "random", "random:fade,zoomout",
	} {
		schedule, err := ParseSchedule(spelling)
		if err != nil {
			t.Fatalf("ParseSchedule(%q) error: %v", spelling, err)
		}
		for name, build := range builders {
			url := build(schedule)
			params, err := ParseFractalURL(url)
			if err != nil {
				t.Fatalf("%s copy of %q did not parse: %v", name, spelling, err)
			}
			if !schedulesEqual(params.Schedule, schedule) {
				t.Errorf("%s copy of %q parsed %+v, want %+v",
					name, spelling, params.Schedule, schedule)
			}
			if got := ScheduleString(params.Schedule); got != spelling {
				t.Errorf("%s copy of %q rendered back as %q", name, spelling, got)
			}
		}
	}
}
