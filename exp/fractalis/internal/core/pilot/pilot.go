// Package pilot holds the policies that evolve a fractal view over
// time and the engine seam that runs them.
package pilot

import (
	"fmt"
	"slices"
)

// ID identifies a registered pilot.
type ID string

// Registered pilot IDs.
const (
	ManualID   ID = "manual"
	AutoZoomID ID = "autozoom"
	VantageID  ID = "vantage"
)

// IDs returns every registered pilot ID.
func IDs() []ID {
	return []ID{ManualID, AutoZoomID, VantageID}
}

// ParseID validates a registered pilot ID.
func ParseID(s string) (ID, error) {
	id := ID(s)
	if slices.Contains(IDs(), id) {
		return id, nil
	}
	return "", fmt.Errorf("invalid pilot: %s", s)
}

// Name returns the display name of a registered pilot.
func Name(id ID) string {
	switch id {
	case ManualID:
		return "Manual"
	case AutoZoomID:
		return "Auto-zoom"
	case VantageID:
		return "Vantage"
	default:
		return string(id)
	}
}

// View is the fractal view state a pilot evolves each tick.
type View struct {
	FractalType string
	CenterX     float64
	CenterY     float64
	Zoom        float64
	MaxIter     int
}

// Pilot is the policy that evolves a fractal view over time. Exactly
// one pilot is active at any moment.
type Pilot interface {
	// Step evolves the view one tick under this pilot's policy.
	Step(v View) View
	// TransitionSchedule returns the transitions this pilot uses when
	// the view arrives at a boundary.
	TransitionSchedule() Schedule
}

// TraversalMode is how a transition schedule picks its next
// transition: single, sequential, or random.
type TraversalMode int

const (
	Single TraversalMode = iota
	Sequential
	Random
)

// Schedule is the ordered list of transitions a pilot uses, together
// with its traversal mode. An empty schedule means no transition:
// boundary arrival clamps and the view stays put.
type Schedule struct {
	Transitions []string
	Traversal   TraversalMode
}

// Empty reports whether the schedule lists no transitions.
func (s Schedule) Empty() bool { return len(s.Transitions) == 0 }
