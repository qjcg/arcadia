package pilot

// The v1 effect vocabulary: the IDs a Schedule names when it picks a
// transition for a boundary crossing. The vocabulary lives with the
// Schedule because schedules are what name these; the effect registry
// in the transition package keys on the same strings.
const (
	FadeID         = "fade"
	ZoomOutID      = "zoomout"
	RotateID       = "rotate"
	BreakthroughID = "breakthrough"
	StarWipeID     = "starwipe"
)

// AllEffectIDs returns every effect ID in the v1 vocabulary.
func AllEffectIDs() []string {
	return []string{FadeID, ZoomOutID, RotateID, BreakthroughID, StarWipeID}
}
