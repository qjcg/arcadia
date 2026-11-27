package main

import (
	"testing"

	"github.com/qjcg/arcadia/exp/fractalis/internal/core/persistence"
	"github.com/qjcg/arcadia/exp/fractalis/internal/core/pilot"
)

func TestSchedulePrecedence(t *testing.T) {
	m := newSteeredModel()
	check := func(want string) {
		t.Helper()
		if got := persistence.ScheduleString(m.schedule); got != want {
			t.Fatalf("schedule = %q, want %q", got, want)
		}
	}
	m.selectPilot(pilot.AutoZoomID)
	check("random")
	m.setURLSchedule(pilot.AutoZoomID, pilot.Schedule{}, true)
	check("none")
	m.selectPilot(pilot.VantageID)
	check("fade")
	m.selectPilot(pilot.AutoZoomID)
	check("none")
	m = keyRunes(m, "T")
	check("fade")
	m.selectPilot(pilot.ManualID)
	check("fade")
}

func TestSchedulePresets(t *testing.T) {
	m := newSteeredModel()
	for _, want := range []string{"fade", "zoomout", "rotate", "breakthrough", "starwipe", "fade,zoomout,rotate,breakthrough,starwipe", "random", "none", "fade"} {
		m = keyRunes(m, "T")
		if got := persistence.ScheduleString(m.schedule); got != want {
			t.Fatalf("schedule = %q, want %q", got, want)
		}
	}
}
