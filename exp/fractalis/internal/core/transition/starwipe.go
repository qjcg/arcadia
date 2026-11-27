package transition

import (
	"math"

	"github.com/qjcg/arcadia/exp/fractalis/internal/core/pilot"
)

// StarWipe reveals the destination through an expanding five-point star.
type StarWipe struct{}

func (StarWipe) Frame(src, dst pilot.View, p Progress) Frame {
	return Frame{Source: src, Destination: dst, Mix: float64(p), Star: true}
}

// StarThreshold gives the reveal progress for normalized screen coordinates.
func StarThreshold(x, y float64) float64 {
	x, y = x-0.5, y-0.5
	const sector = math.Pi / 5
	angle := math.Mod(math.Atan2(y, x)+math.Pi/2+2*math.Pi, 2*math.Pi)
	a := math.Mod(angle, 2*sector)
	if a > sector {
		a = 2*sector - a
	}
	const inner = 0.45
	edge := inner * math.Sin(sector) / (inner*math.Sin(sector-a) + math.Sin(a))
	return min(1, math.Hypot(x, y)*inner/(edge*math.Sqrt(0.5)))
}
