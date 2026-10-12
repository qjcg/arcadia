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
}

// Vantage is the pilot that tours curated viewpoints on its scene
// timer. Scene switches are the pilot moving, not transitions.
type Vantage struct {
	sceneDuration int
	viewpoints    []Viewpoint
	next          int

	sceneTimer int
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
// then moves it to the next viewpoint on the tour.
func (p *Vantage) Step(v View) View {
	if p.sceneTimer > 0 {
		p.sceneTimer--
		return v
	}
	v = p.viewpoints[p.next].View
	p.next = (p.next + 1) % len(p.viewpoints)
	p.sceneTimer = p.sceneDuration - 1
	return v
}

// TransitionSchedule returns the Vantage pilot's default: a single
// fade. Scene switches never play it; only boundary crossings do.
func (p *Vantage) TransitionSchedule() Schedule {
	return Schedule{Transitions: []string{"fade"}, Traversal: Single}
}
