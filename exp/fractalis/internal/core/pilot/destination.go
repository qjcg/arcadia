package pilot

import "math/rand"

// DestinationPolicy is the rule that chooses the fractal a transition
// arrives at.
type DestinationPolicy interface {
	// Choose returns the destination view for a crossing from the
	// current view.
	Choose(current View) View
}

// randomDestination is the v1 destination policy: a random fractal
// type different from the current one, shared by all pilots.
type randomDestination struct {
	types []string
	rng   *rand.Rand
}

// NewRandomDestination returns the v1 destination policy: a random
// fractal type different from the current one, drawn from types.
func NewRandomDestination(types []string, rng *rand.Rand) DestinationPolicy {
	return &randomDestination{types: types, rng: rng}
}

// Choose picks a random fractal type different from the current one
// and arrives at its default view, carrying the iteration count over.
func (d *randomDestination) Choose(current View) View {
	candidates := make([]string, 0, len(d.types))
	for _, t := range d.types {
		if t != current.FractalType {
			candidates = append(candidates, t)
		}
	}
	if len(candidates) == 0 {
		return current
	}

	arrival := defaultView(candidates[d.rng.Intn(len(candidates))])
	arrival.MaxIter = current.MaxIter
	return arrival
}

// defaultView returns the home view a fractal type opens on.
func defaultView(fractalType string) View {
	v := View{FractalType: fractalType, CenterX: -0.5, Zoom: 1.0}
	switch fractalType {
	case "julia", "newton":
		v.CenterX = 0
	case "burningship":
		v.CenterY = -0.6
	}
	return v
}
