package transition

import "testing"

func TestProgressAdvancesFromStartToDoneInTwentyTicks(t *testing.T) {
	p := Start
	ticks := 0
	for !p.Landed() {
		p = p.Advance()
		ticks++
		if ticks > 20 {
			t.Fatalf("progress not done after 20 ticks: %v", p)
		}
	}

	if ticks != 20 {
		t.Fatalf("ticks to land = %d, want 20", ticks)
	}
	if p != Done {
		t.Fatalf("progress after landing = %v, want %v", p, Done)
	}
}

func TestProgressNeverOvershootsDone(t *testing.T) {
	p := Done
	if got := p.Advance(); got != Done {
		t.Fatalf("Advance() past the end = %v, want %v", got, Done)
	}
}

func TestProgressStartsAboveZeroSoAnimationTriggers(t *testing.T) {
	if Start <= 0 {
		t.Fatalf("Start = %v, want above 0 so animation triggers", Start)
	}
	if Start >= Done {
		t.Fatalf("Start = %v, want below %v", Start, Done)
	}
}
