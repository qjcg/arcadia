package pilot

// manual is the Manual pilot, the default pilot, whose policy is to
// leave the view unchanged.
type manual struct{}

// Manual returns the Manual pilot.
func Manual() Pilot { return manual{} }

// Step leaves the view unchanged.
func (manual) Step(v View) View { return v }

// TransitionSchedule returns the Manual pilot's default, empty
// schedule: boundary arrival clamps.
func (manual) TransitionSchedule() Schedule { return Schedule{} }
