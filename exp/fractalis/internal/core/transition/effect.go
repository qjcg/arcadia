package transition

import "github.com/qjcg/arcadia/exp/fractalis/internal/core/pilot"

// Effect IDs. The v1 vocabulary lives with the Schedule in the pilot
// package — schedules name these; the effect registry keys on them.
const (
	FadeID         = pilot.FadeID
	ZoomOutID      = pilot.ZoomOutID
	RotateID       = pilot.RotateID
	BreakthroughID = pilot.BreakthroughID
	StarWipeID     = pilot.StarWipeID
)

// EffectIDs returns every effect ID in the v1 vocabulary.
func EffectIDs() []string { return pilot.AllEffectIDs() }

// Effect renders a crossing from a source view to a destination view
// over the shared progress model. Frontends render the frames; effects
// never touch the screen.
type Effect interface {
	// Frame renders one frame of the crossing at progress p.
	Frame(src, dst pilot.View, p Progress) Frame
}

// Frame is one step of a crossing for a frontend to render: the two
// views being crossed and how much of the destination shows.
type Frame struct {
	// Source is the view the crossing starts from.
	Source pilot.View
	// Destination is the view the crossing arrives at.
	Destination pilot.View
	// Mix in [0,1] splits the frame between the two views: 0 shows
	// only the source, 1 only the destination.
	Mix float64
}

// Registry maps effect IDs to effect implementations.
type Registry struct {
	byID map[string]Effect
}

// NewRegistry returns an empty effect registry.
func NewRegistry() *Registry {
	return &Registry{byID: map[string]Effect{}}
}

// Register makes e available under id.
func (r *Registry) Register(id string, e Effect) { r.byID[id] = e }

// Lookup returns the effect registered under id.
func (r *Registry) Lookup(id string) (Effect, bool) {
	e, ok := r.byID[id]
	return e, ok
}
