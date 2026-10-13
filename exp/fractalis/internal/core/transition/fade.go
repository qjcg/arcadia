package transition

import "github.com/qjcg/arcadia/exp/fractalis/internal/core/pilot"

// Fade dissolves from the source view to the destination view over
// the shared progress model.
type Fade struct{}

// Frame mixes the two views by the crossing progress.
func (Fade) Frame(src, dst pilot.View, p Progress) Frame {
	return Frame{Source: src, Destination: dst, Mix: float64(p)}
}
