package pilot

import (
	"math/rand"
	"testing"
)

// scriptedSource is an injected rand.Source whose draws are fixed, so
// traversal picks are deterministic under test. rand.Rand reads Int31
// from the high bits of Int63, so draws are shifted up.
type scriptedSource struct {
	draws []int64
	i     int
}

func (s *scriptedSource) Int63() int64 {
	v := s.draws[s.i%len(s.draws)]
	s.i++
	return v << 32
}

func (s *scriptedSource) Seed(int64) {}

func scriptedRand(draws ...int64) *rand.Rand {
	return rand.New(&scriptedSource{draws: draws})
}

func TestSchedulerSinglePicksTheFirstTransitionEveryCrossing(t *testing.T) {
	s := NewScheduler(scriptedRand())
	schedule := Schedule{Transitions: []string{"fade"}, Traversal: Single}

	for range 3 {
		if got := s.Next(schedule); got != "fade" {
			t.Fatalf("Next() = %q, want %q every crossing", got, "fade")
		}
	}
}

func TestSchedulerSequentialStepsThroughTheScheduleInOrder(t *testing.T) {
	s := NewScheduler(scriptedRand())
	schedule := Schedule{Transitions: []string{"fade", "rotate", "zoomout"}, Traversal: Sequential}

	want := []string{"fade", "rotate", "zoomout", "fade", "rotate"}
	for i, w := range want {
		if got := s.Next(schedule); got != w {
			t.Fatalf("crossing %d: Next() = %q, want %q", i+1, got, w)
		}
	}
}

func TestSchedulerRandomRerollsPerCrossing(t *testing.T) {
	s := NewScheduler(scriptedRand(0, 2, 1))
	schedule := Schedule{Transitions: []string{"fade", "rotate", "zoomout"}, Traversal: Random}

	want := []string{"fade", "zoomout", "rotate"}
	for i, w := range want {
		if got := s.Next(schedule); got != w {
			t.Fatalf("crossing %d: Next() = %q, want %q", i+1, got, w)
		}
	}
}

func TestSchedulerEmptySchedulePicksNothing(t *testing.T) {
	s := NewScheduler(scriptedRand())

	if got := s.Next(Schedule{}); got != "" {
		t.Fatalf("Next() on an empty schedule = %q, want none", got)
	}
}
