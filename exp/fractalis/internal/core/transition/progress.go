// Package transition holds the effects shown while a view crosses a
// boundary into another fractal: the Effect interface, the effect
// registry, and the shared progress model.
package transition

// Progress is the shared progress model every effect renders over: a
// crossing starts slightly above zero and advances by Step per tick
// until it lands at Done.
type Progress float64

// The shared progress model: a crossing takes 20 ticks (one second at
// the 50ms tick) from just above zero to done.
const (
	// Start is the progress at a crossing's first frame, slightly
	// above 0 so animation triggers.
	Start Progress = 0.01
	// Step is the progress a crossing advances per tick.
	Step Progress = 0.05
	// Done is the progress at which a crossing lands.
	Done Progress = 1
)

// Advance moves progress along the shared model, never past Done.
func (p Progress) Advance() Progress {
	p += Step
	if p > Done {
		p = Done
	}
	return p
}

// Landed reports whether progress has reached the end of the
// crossing.
func (p Progress) Landed() bool { return p >= Done }
