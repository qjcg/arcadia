package pilot

import "slices"

// NextPreset advances the schedule cycle shared by all frontends.
func NextPreset(current Schedule) Schedule {
	ids := AllEffectIDs()
	presets := make([]Schedule, 0, len(ids)+3)
	for _, id := range ids {
		presets = append(presets, Schedule{Transitions: []string{id}, Traversal: Single})
	}
	presets = append(presets, Schedule{Transitions: ids, Traversal: Sequential}, Schedule{Transitions: ids, Traversal: Random}, Schedule{})
	for i, preset := range presets {
		if preset.Traversal == current.Traversal && slices.Equal(preset.Transitions, current.Transitions) {
			return presets[(i+1)%len(presets)]
		}
	}
	return presets[0]
}
