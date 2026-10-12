package pilot

// DefaultSceneDuration is how many ticks the Vantage pilot rests on
// each scene when no duration is given (~5 seconds at 20 ticks per
// second).
const DefaultSceneDuration = 100

// Viewpoint is a named curated composition the Vantage pilot tours.
type Viewpoint struct {
	Name string
	View View
}

// VantageOptions configures a Vantage pilot.
type VantageOptions struct {
	// SceneDuration is how many ticks the view rests on each scene.
	// Zero selects DefaultSceneDuration.
	SceneDuration int
	// Viewpoints is the curated tour the pilot runs.
	Viewpoints []Viewpoint
	// Discover finds the interesting point a scene pans toward. With
	// no Discover, a scene holds its curated composition.
	Discover func(View) (float64, float64)
}

// The slow scene pan: a fraction of the remaining distance per tick,
// on a fixed progress clock so a pan ends within a scene.
const (
	vantagePanRate = 0.02
	vantagePanStep = 0.01
)

// Vantage is the pilot that tours curated viewpoints on its scene
// timer. Scene switches are the pilot moving, not transitions.
type Vantage struct {
	sceneDuration int
	viewpoints    []Viewpoint
	discover      func(View) (float64, float64)
	next          int

	sceneTimer  int
	targetX     float64
	targetY     float64
	panProgress float64
}

// NewVantage returns a Vantage pilot touring the given viewpoints,
// or CuratedViewpoints() when none are given.
func NewVantage(opts VantageOptions) *Vantage {
	viewpoints := opts.Viewpoints
	if len(viewpoints) == 0 {
		viewpoints = CuratedViewpoints()
	}
	return &Vantage{
		sceneDuration: sceneTicks(opts.SceneDuration),
		viewpoints:    viewpoints,
		discover:      opts.Discover,
	}
}

// sceneTicks normalizes a scene duration to ticks per scene.
func sceneTicks(d int) int {
	if d < 1 {
		return DefaultSceneDuration
	}
	return d
}

// SceneDuration returns the ticks the view rests on each scene.
func (p *Vantage) SceneDuration() int { return p.sceneDuration }

// SetSceneDuration retimes the tour: the new duration applies from
// the next scene.
func (p *Vantage) SetSceneDuration(d int) { p.sceneDuration = sceneTicks(d) }

// NextScene makes the following Step skip ahead to the next viewpoint
// on the tour.
func (p *Vantage) NextScene() { p.sceneTimer = 0 }

// Step rests the view on each curated scene for SceneDuration ticks,
// panning it slowly toward its discovered point, then moves it to the
// next viewpoint on the tour.
func (p *Vantage) Step(v View) View {
	if p.sceneTimer > 0 {
		p.sceneTimer--
		return p.pan(v)
	}
	v = p.startScene(v)
	p.sceneTimer = p.sceneDuration - 1
	return v
}

// startScene lands the view on the next curated viewpoint and points
// its slow pan at the discovered interesting point.
func (p *Vantage) startScene(v View) View {
	scene := p.viewpoints[p.next]
	p.next = (p.next + 1) % len(p.viewpoints)
	p.panProgress = 0
	p.targetX, p.targetY = scene.View.CenterX, scene.View.CenterY
	if p.discover != nil {
		p.targetX, p.targetY = p.discover(scene.View)
	}
	return scene.View
}

// pan creeps the view toward the scene's discovered point.
func (p *Vantage) pan(v View) View {
	if p.discover == nil || p.panProgress >= 1 {
		return v
	}
	v.CenterX += (p.targetX - v.CenterX) * vantagePanRate
	v.CenterY += (p.targetY - v.CenterY) * vantagePanRate
	p.panProgress += vantagePanStep
	if p.panProgress > 1 {
		p.panProgress = 1
	}
	return v
}

// TransitionSchedule returns the Vantage pilot's default: a single
// fade. Scene switches never play it; only boundary crossings do.
func (p *Vantage) TransitionSchedule() Schedule {
	return Schedule{Transitions: []string{"fade"}, Traversal: Single}
}
