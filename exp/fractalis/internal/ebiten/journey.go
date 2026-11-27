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

func (g *Game) tickFlatJourney() {
	id := pilot.ManualID
	if g.autopilotEnabled {
		id = pilot.AutoZoomID
	}
	if g.vantageEnabled {
		id = pilot.VantageID
	}
	if g.flatJourney == nil {
		rng := rand.New(rand.NewSource(time.Now().UnixNano()))
		g.flatJourney = transition.NewRunner(pilot.NewScheduler(rng), pilot.NewRandomDestination(search.AllFractalTypes, rng))
		g.flatURLPilot = id
	}
	if id != g.flatPilot {
		var active pilot.Pilot = pilot.Manual()
		switch id {
		case pilot.AutoZoomID:
			a := pilot.NewAutoZoom(flatSeeker{g})
			a.SetSpeed(g.config.PilotSpeed)
			a.SetDirection(g.config.PilotDirection)
			active = a
		case pilot.VantageID:
			active = pilot.NewVantage(pilot.VantageOptions{SceneDuration: g.config.PilotSceneTicks, Discover: flatSeeker{g}.FindInterestingPoint})
		}
		g.flatJourney.Activate(active)
		schedule := active.TransitionSchedule()
		if g.config.TransitionSchedule != nil && id == g.flatURLPilot {
			schedule = *g.config.TransitionSchedule
		}
		if g.flatOverride != nil {
			schedule = *g.flatOverride
		}
		g.flatSchedule = schedule
		g.flatJourney.SetSchedule(schedule)
		g.flatPilot = id
	}
	if g.cycleSchedule {
		g.cycleSchedule = false
		g.flatSchedule = pilot.NextPreset(g.flatSchedule)
		g.flatOverride = new(g.flatSchedule)
		g.flatJourney.SetSchedule(g.flatSchedule)
	}
	cfg := g.getPersistenceConfig()
	if a, ok := g.flatJourney.Active().(*pilot.AutoZoom); ok {
		if rate := g.autopilotSpeed; rate > 0 {
			base := g.config.PilotSpeed
			if base <= 0 {
				base = pilot.DefaultSpeed
			}
			a.SetSpeed(math.Pow(base, rate))
		}
	}
	view := pilot.View{FractalType: cfg.FractalType, CenterX: cfg.CenterX, CenterY: cfg.CenterY, Zoom: cfg.Zoom, MaxIter: cfg.MaxIter, JuliaCr: cfg.JuliaCr, JuliaCi: cfg.JuliaCi}
	next, crossing := g.flatJourney.Tick(view)
	if crossing != nil && !crossing.Progress.Landed() {
		g.flatCrossing = crossing
	} else {
		g.flatCrossing = nil
	}
	if next != view {
		g.applyFlatView(next)
	}
}

func (g *Game) applyFlatView(v pilot.View) {
	cfg := g.getPersistenceConfig()
	cfg.FractalType, cfg.CenterX, cfg.CenterY, cfg.Zoom, cfg.MaxIter = v.FractalType, v.CenterX, v.CenterY, v.Zoom, v.MaxIter
	g.applyPersistenceConfig(cfg)
}

func (g *Game) drawFlatCrossing(screen *ebiten.Image) {
	c := g.flatCrossing
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
	g.applyFlatView(frame.Source)
	g.drawFractal(screen)
	bounds := screen.Bounds()
	if g.crossingLayer == nil || g.crossingLayer.Bounds() != bounds {
		g.crossingLayer = ebiten.NewImage(bounds.Dx(), bounds.Dy())
	}
	g.crossingLayer.Clear()
	g.applyFlatView(frame.Destination)
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
