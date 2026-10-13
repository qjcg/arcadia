package transition

import "github.com/qjcg/arcadia/exp/fractalis/internal/core/pilot"

// Registered effect IDs. These are the v1 vocabulary: the schedule
// grammar and the effect registry both key on them.
const (
	FadeID         = "fade"
	ZoomOutID      = "zoomout"
	RotateID       = "rotate"
	BreakthroughID = "breakthrough"
	StarWipeID     = "starwipe"
)

// EffectIDs returns every effect ID in the v1 vocabulary.
func EffectIDs() []string {
	return []string{FadeID, ZoomOutID, RotateID, BreakthroughID, StarWipeID}
}

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
