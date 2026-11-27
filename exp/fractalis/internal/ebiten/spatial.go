//go:build ebiten || (js && wasm)

package ebiten

import (
	"math"

	"github.com/qjcg/arcadia/exp/fractalis/internal/core/persistence"
	"github.com/qjcg/arcadia/exp/fractalis/internal/core/pilot"
)

func spatialView(id int, x, y, z, pitch, yaw, shape float64) pilot.View {
	name := persistence.FractalMandelbox
	if id == 0 {
		name = persistence.FractalMandelbulb
	}
	return pilot.View{FractalType: name, Zoom: 1 / max(0.001, math.Sqrt(x*x+y*y+z*z)), MaxIter: 32, Camera: pilot.Camera{Spatial: true, X: x, Y: y, Z: z, Pitch: pitch, Yaw: yaw, Shape: shape}}
}

func spatialViewpoints() []pilot.Viewpoint {
	points := defaultVantagePoints()
	views := make([]pilot.Viewpoint, 0, len(points))
	for _, p := range points {
		views = append(views, pilot.Viewpoint{Name: p.Name, View: spatialView(p.Type, p.X, p.Y, p.Z, p.Pitch, p.Yaw, p.Scale)})
	}
	return views
}

type cameraDestination struct{ flat pilot.DestinationPolicy }

func (d cameraDestination) Choose(v pilot.View) pilot.View {
	if !v.Camera.Spatial {
		return d.flat.Choose(v)
	}
	if v.FractalType == persistence.FractalMandelbulb {
		return spatialView(1, 5, 2, 5, -0.3, 3.927, 2.7)
	}
	return spatialView(0, 0, 1.5, 6, 0, math.Pi, 8)
}
