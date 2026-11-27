//go:build ebiten || (js && wasm)

package ebiten

import (
	"image"
	"math"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/qjcg/arcadia/exp/fractalis/internal/core/persistence"
	"github.com/qjcg/arcadia/exp/fractalis/internal/core/pilot"
	"github.com/qjcg/arcadia/exp/fractalis/internal/core/search"
	"github.com/qjcg/arcadia/exp/fractalis/internal/core/transition"
)

type flatSeeker struct{ game *Game }

func (s flatSeeker) config(v pilot.View) persistence.Config {
	cfg := s.game.getPersistenceConfig()
	cfg.FractalType, cfg.CenterX, cfg.CenterY, cfg.Zoom, cfg.MaxIter = v.FractalType, v.CenterX, v.CenterY, v.Zoom, v.MaxIter
	return cfg
}

func (s flatSeeker) FindInterestingPoint(v pilot.View) (float64, float64) {
	return s.game.calculator.FindInterestingPoint(search.DescentSearchPasses(), s.config(v))
}

func (s flatSeeker) IsViewUniform(v pilot.View) bool {
	return s.game.calculator.IsViewUniform(s.config(v))
}

func (g *Game) nextVantageScene() {
	if g.journey != nil {
		if v, ok := g.journey.Active().(*pilot.Vantage); ok {
			v.NextScene()
		}
	}
}

func (g *Game) tickJourney() {
	id := pilot.ManualID
	if g.autopilotEnabled {
		id = pilot.AutoZoomID
	}
	if g.vantageEnabled {
		id = pilot.VantageID
	}
	if g.journey == nil {
		rng := rand.New(rand.NewSource(time.Now().UnixNano()))
		g.journey = transition.NewRunner(pilot.NewScheduler(rng), cameraDestination{flat: pilot.NewRandomDestination(search.AllFractalTypes, rng)})
		g.urlPilot = id
	}
	if id != g.activePilot {
		var active pilot.Pilot = pilot.Manual()
		switch id {
		case pilot.AutoZoomID:
			a := pilot.NewAutoZoom(flatSeeker{g})
			a.SetSpeed(g.config.PilotSpeed)
			a.SetDirection(g.config.PilotDirection)
			active = a
		case pilot.VantageID:
			opts := pilot.VantageOptions{SceneDuration: g.config.PilotSceneTicks, Discover: flatSeeker{g}.FindInterestingPoint}
			if g.fractalType < 2 {
				opts.Viewpoints = spatialViewpoints()
				opts.Discover = nil
			}
			active = pilot.NewVantage(opts)
		}
		g.journey.Activate(active)
		schedule := active.TransitionSchedule()
		if g.config.TransitionSchedule != nil && id == g.urlPilot {
			schedule = *g.config.TransitionSchedule
		}
		if g.sessionSchedule != nil {
			schedule = *g.sessionSchedule
		}
		g.schedule = schedule
		g.journey.SetSchedule(schedule)
		g.activePilot = id
	}
	if g.cycleSchedule {
		g.cycleSchedule = false
		g.schedule = pilot.NextPreset(g.schedule)
		g.sessionSchedule = new(g.schedule)
		g.journey.SetSchedule(g.schedule)
	}
	cfg := g.getPersistenceConfig()
	if a, ok := g.journey.Active().(*pilot.AutoZoom); ok {
		if rate := g.autopilotSpeed; rate > 0 {
			base := g.config.PilotSpeed
			if base <= 0 {
				base = pilot.DefaultSpeed
			}
			a.SetSpeed(math.Pow(base, rate))
		}
	}
	view := pilot.View{FractalType: cfg.FractalType, CenterX: cfg.CenterX, CenterY: cfg.CenterY, Zoom: cfg.Zoom, MaxIter: cfg.MaxIter, JuliaCr: cfg.JuliaCr, JuliaCi: cfg.JuliaCi}
	if g.fractalType < 2 {
		shape := g.power
		if g.fractalType == 1 {
			shape = g.boxScale
		}
		view = spatialView(g.fractalType, g.camX, g.camY, g.camZ, g.camPitch, g.camYaw, shape)
		view.MaxIter = cfg.MaxIter
	}
	next, crossing := g.journey.Tick(view)
	if crossing != nil && !crossing.Progress.Landed() {
		g.crossing = crossing
	} else {
		g.crossing = nil
	}
	if next != view {
		g.applyJourneyView(next)
	}
}

func (g *Game) applyJourneyView(v pilot.View) {
	if c := v.Camera; c.Spatial {
		if v.FractalType == persistence.FractalMandelbulb {
			g.fractalType, g.power = 0, c.Shape
		} else {
			g.fractalType, g.boxScale = 1, c.Shape
		}
		radius := math.Sqrt(c.X*c.X + c.Y*c.Y + c.Z*c.Z)
		if radius == 0 {
			c.Z, radius = 1, 1
		}
		scale := 1 / (max(v.Zoom, 0.001) * radius)
		g.camX, g.camY, g.camZ = c.X*scale, c.Y*scale, c.Z*scale
		g.camPitch, g.camYaw = c.Pitch, c.Yaw
		g.iterations = v.MaxIter
		return
	}
	cfg := g.getPersistenceConfig()
	cfg.FractalType, cfg.CenterX, cfg.CenterY, cfg.Zoom, cfg.MaxIter = v.FractalType, v.CenterX, v.CenterY, v.Zoom, v.MaxIter
	g.applyPersistenceConfig(cfg)
}

func (g *Game) drawCrossing(screen *ebiten.Image) {
	c := g.crossing
	var effect transition.Effect = transition.Fade{}
	switch c.Effect {
	case transition.ZoomOutID:
		effect = transition.ZoomOutIn{}
	case transition.RotateID:
		effect = transition.Rotate{}
	case transition.BreakthroughID:
		effect = transition.Breakthrough{}
	case transition.StarWipeID:
		effect = transition.StarWipe{}
	}
	frame := effect.Frame(c.Source, c.Destination, c.Progress)
	// Drawing samples temporary camera states; never let them affect Update.
	saved := *g
	defer func() { layer := g.crossingLayer; *g = saved; g.crossingLayer = layer }()
	g.applyJourneyView(frame.Source)
	g.drawFractal(screen)
	bounds := screen.Bounds()
	if g.crossingLayer == nil || g.crossingLayer.Bounds() != bounds {
		g.crossingLayer = ebiten.NewImage(bounds.Dx(), bounds.Dy())
	}
	g.crossingLayer.Clear()
	g.applyJourneyView(frame.Destination)
	g.drawFractal(g.crossingLayer)
	if frame.Star {
		for y := 0; y < bounds.Dy(); y++ {
			for x := 0; x < bounds.Dx(); {
				visible := func(x int) bool {
					return transition.StarThreshold((float64(x)+0.5)/float64(bounds.Dx()), (float64(y)+0.5)/float64(bounds.Dy())) < frame.Mix
				}
				if !visible(x) {
					x++
					continue
				}
				start := x
				for x < bounds.Dx() && visible(x) {
					x++
				}
				op := &ebiten.DrawImageOptions{}
				op.GeoM.Translate(float64(start), float64(y))
				screen.DrawImage(g.crossingLayer.SubImage(image.Rect(start, y, x, y+1)).(*ebiten.Image), op)
			}
		}
		return
	}
	op := &ebiten.DrawImageOptions{}
	op.ColorScale.ScaleAlpha(float32(frame.Mix))
	screen.DrawImage(g.crossingLayer, op)
}
