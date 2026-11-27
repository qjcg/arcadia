package transition

import "github.com/qjcg/arcadia/exp/fractalis/internal/core/pilot"

// Crossing is a boundary crossing in progress: the effect playing it,
// the views it crosses, and how far along the shared progress model it
// is.
type Crossing struct {
	// Effect is the registered ID of the effect playing this
	// crossing, picked by the scheduler.
	Effect string
	// Source is the boundary view the crossing starts from.
	Source pilot.View
	// Destination is the view the crossing arrives at.
	Destination pilot.View
	// Progress along the shared progress model, from Start to Done.
	Progress Progress
}

// Runner is the engine seam: it runs the active pilot over the view,
// fires a transition when the view arrives at a boundary while
// actively moving, and plays each crossing over the shared progress
// model. Selection lives in the schedule's scheduler and destination
// policy; pilots declare schedules only.
type Runner struct {
	engine      pilot.Engine
	scheduler   *pilot.Scheduler
	destination pilot.DestinationPolicy
	schedule    pilot.Schedule
	crossing    *Crossing
}

// NewRunner returns an engine seam that executes schedules with the
// given scheduler and arrives through the given destination policy.
func NewRunner(scheduler *pilot.Scheduler, destination pilot.DestinationPolicy) *Runner {
	return &Runner{scheduler: scheduler, destination: destination}
}

// Activate makes p the one active pilot, replacing whatever was
// active.
func (r *Runner) Activate(p pilot.Pilot) { r.engine.Activate(p) }

// Active returns the one pilot active at this moment.
func (r *Runner) Active() pilot.Pilot { return r.engine.Active() }

// SetSchedule installs the effective transition schedule: the one the
// scheduler executes at every boundary arrival. An empty schedule
// means none: boundary arrival clamps.
func (r *Runner) SetSchedule(schedule pilot.Schedule) { r.schedule = schedule }

// Tick advances the seam one tick and returns the resulting view and
// the crossing in flight, if any. A crossing freezes the view at its
// source while the effect plays, and lands the destination view when
// progress reaches Done.
func (r *Runner) Tick(v pilot.View) (pilot.View, *Crossing) {
	if r.crossing != nil {
		return r.advance()
	}

	next, arrival := r.engine.Tick(v)
	if arrival == nil {
		return next, nil
	}

	source := next
	lower, upper := source.ZoomBounds()
	switch *arrival {
	case pilot.TopBoundary:
		source.Zoom = upper
	case pilot.BottomBoundary:
		source.Zoom = lower
	}
	if r.schedule.Empty() {
		return source, nil
	}

	r.crossing = &Crossing{
		Effect:      r.scheduler.Next(r.schedule),
		Source:      source,
		Destination: r.destination.Choose(source),
		Progress:    Start,
	}
	return source, r.crossing.copy()
}

// advance moves an in-flight crossing along the shared progress
// model, landing the destination when it is done.
func (r *Runner) advance() (pilot.View, *Crossing) {
	r.crossing.Progress = r.crossing.Progress.Advance()
	c := r.crossing.copy()
	if r.crossing.Progress.Landed() {
		r.crossing = nil
		return c.Destination, c
	}
	return c.Source, c
}

func (c *Crossing) copy() *Crossing {
	copied := *c
	return &copied
}
