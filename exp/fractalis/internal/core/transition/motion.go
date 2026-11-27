package transition

import (
	"math"

	"github.com/qjcg/arcadia/exp/fractalis/internal/core/pilot"
)

// ZoomOutIn pulls both views back around the midpoint of the crossing.
type ZoomOutIn struct{}

func (ZoomOutIn) Frame(src, dst pilot.View, p Progress) Frame {
	t := float64(p)
	scale := 1 - 0.9*math.Sin(math.Pi*t)
	src.Zoom *= scale
	dst.Zoom *= scale
	return Frame{Source: src, Destination: dst, Mix: t}
}

// Rotate sweeps the view center in an orbit during the crossing.
type Rotate struct{}

func (Rotate) Frame(src, dst pilot.View, p Progress) Frame {
	t := float64(p)
	angle := 2 * math.Pi * t
	offset := func(v pilot.View) pilot.View {
		radius := 0.5 / max(v.Zoom, pilot.MinZoom)
		v.CenterX += radius * (math.Cos(angle) - 1)
		v.CenterY += radius * math.Sin(angle)
		return v
	}
	return Frame{Source: offset(src), Destination: offset(dst), Mix: t}
}

// Breakthrough shakes the source view before revealing the destination.
type Breakthrough struct{}

func (Breakthrough) Frame(src, dst pilot.View, p Progress) Frame {
	t := float64(p)
	shake := math.Sin(12*math.Pi*t) * math.Sin(math.Pi*t)
	offset := shake * 0.1 / max(src.Zoom, pilot.MinZoom)
	src.CenterX += offset
	src.CenterY -= offset
	src.Zoom *= 1 - 0.5*math.Sin(math.Pi*t)
	return Frame{Source: src, Destination: dst, Mix: min(1, max(0, (t-0.3)/0.5))}
}
