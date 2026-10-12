package pilot

import "testing"

func TestParseID(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    ID
		wantErr bool
	}{
		{name: "manual", input: "manual", want: ManualID},
		{name: "autozoom", input: "autozoom", want: AutoZoomID},
		{name: "vantage", input: "vantage", want: VantageID},
		{name: "unknown", input: "warp", wantErr: true},
		{name: "empty", input: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseID(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("ParseID(%q) = %q, want error", tt.input, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseID(%q) error: %v", tt.input, err)
			}
			if got != tt.want {
				t.Errorf("ParseID(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestRegisteredIDsRoundTripThroughName(t *testing.T) {
	for _, id := range IDs() {
		if Name(id) == "" {
			t.Errorf("Name(%q) is empty, want a display name", id)
		}
		if _, err := ParseID(string(id)); err != nil {
			t.Errorf("ParseID(%q) error: %v", id, err)
		}
	}
}

func TestName(t *testing.T) {
	tests := []struct {
		id   ID
		want string
	}{
		{ManualID, "Manual"},
		{AutoZoomID, "Auto-zoom"},
		{VantageID, "Vantage"},
	}

	for _, tt := range tests {
		if got := Name(tt.id); got != tt.want {
			t.Errorf("Name(%q) = %q, want %q", tt.id, got, tt.want)
		}
	}
}

func TestManualPilotIsNoOp(t *testing.T) {
	v := View{
		FractalType: "mandelbrot",
		CenterX:     -0.75,
		CenterY:     0.1,
		Zoom:        123.4,
		MaxIter:     200,
	}

	got := Manual().Step(v)
	if got != v {
		t.Errorf("Manual().Step() changed the view: got %+v, want %+v", got, v)
	}
}

func TestManualPilotHasEmptySchedule(t *testing.T) {
	if got := Manual().TransitionSchedule(); !got.Empty() {
		t.Errorf("Manual().TransitionSchedule() = %+v, want empty", got)
	}
}

func TestScheduleEmpty(t *testing.T) {
	tests := []struct {
		name     string
		schedule Schedule
		want     bool
	}{
		{name: "no transitions", schedule: Schedule{}, want: true},
		{name: "single transition", schedule: Schedule{Transitions: []string{"fade"}}, want: false},
		{
			name:     "sequential traversal without transitions",
			schedule: Schedule{Traversal: Sequential},
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.schedule.Empty(); got != tt.want {
				t.Errorf("Schedule%+v.Empty() = %v, want %v", tt.schedule, got, tt.want)
			}
		})
	}
}
