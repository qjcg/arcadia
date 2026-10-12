package pilot

import "math/rand"

// Scheduler executes transition schedules: it picks the effect for
// each boundary crossing according to the schedule's traversal mode.
// One scheduler runs every schedule, so a pilot can never get
// transition selection wrong.
type Scheduler struct {
	rng *rand.Rand
	seq int
}

// NewScheduler returns a schedule executor drawing its random picks
// from rng.
func NewScheduler(rng *rand.Rand) *Scheduler {
	return &Scheduler{rng: rng}
}

// Next picks the transition for the next boundary crossing: single
// always takes the first entry, sequential steps through the list in
// order and wraps, and random re-rolls uniformly per crossing. An
// empty schedule picks nothing.
func (s *Scheduler) Next(schedule Schedule) string {
	if schedule.Empty() {
		return ""
	}
	switch schedule.Traversal {
	case Sequential:
		id := schedule.Transitions[s.seq%len(schedule.Transitions)]
		s.seq++
		return id
	case Random:
		return schedule.Transitions[s.rng.Intn(len(schedule.Transitions))]
	default: // Single
		return schedule.Transitions[0]
	}
}
