//go:build ebiten

package ebiten

import (
	"testing"

	"github.com/qjcg/arcadia/exp/fractalis/internal/core/persistence"
	"github.com/qjcg/arcadia/exp/fractalis/internal/core/pilot"
)

func TestSpatialJourneyUsesURLScheduleAtBoundary(t *testing.T) {
	g := NewGame(persistence.Config{FractalType: persistence.FractalMandelbox, AutoPilotEnabled: true, PilotSpeed: 1.2, PilotDirection: pilot.DirectionIn, TransitionSchedule: &pilot.Schedule{Transitions: []string{pilot.StarWipeID}}})
	g.camX, g.camY, g.camZ = 0, 0, 0.65
	g.tickJourney()
	if g.crossing == nil || g.crossing.Effect != pilot.StarWipeID {
		t.Fatalf("crossing = %+v, want star wipe", g.crossing)
	}
	for i := 0; i < 30 && g.crossing != nil; i++ {
		g.tickJourney()
	}
	if g.crossing != nil || g.fractalType != 0 {
		t.Fatalf("did not land on Mandelbulb: type %d, crossing %+v", g.fractalType, g.crossing)
	}
	g.cycleSchedule = true
	g.tickJourney()
	got := persistence.ScheduleString(g.schedule)
	g.autopilotEnabled = false
	g.tickJourney()
	if after := persistence.ScheduleString(g.schedule); after != got {
		t.Fatalf("schedule after pilot switch = %q, want %q", after, got)
	}
}
