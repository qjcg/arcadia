package persistence

import (
	"fmt"
	"slices"
	"strings"

	"github.com/qjcg/arcadia/exp/fractalis/internal/core/pilot"
	"github.com/qjcg/arcadia/exp/fractalis/internal/core/transition"
)

// ParseSchedule parses the transition= schedule grammar: none, a
// single effect ID, a comma list (sequential), random (all effects),
// or random: plus a comma list (random over a subset).
func ParseSchedule(value string) (pilot.Schedule, error) {
	value = strings.TrimSpace(value)
	if value == "" || value == "none" {
		return pilot.Schedule{}, nil
	}

	if rest, ok := strings.CutPrefix(value, "random:"); ok {
		transitions, err := parseEffectList(rest)
		if err != nil {
			return pilot.Schedule{}, err
		}
		return pilot.Schedule{Transitions: transitions, Traversal: pilot.Random}, nil
	}
	if value == "random" {
		return pilot.Schedule{Transitions: transition.EffectIDs(), Traversal: pilot.Random}, nil
	}

	transitions, err := parseEffectList(value)
	if err != nil {
		return pilot.Schedule{}, err
	}
	traversal := pilot.Single
	if len(transitions) > 1 {
		traversal = pilot.Sequential
	}
	return pilot.Schedule{Transitions: transitions, Traversal: traversal}, nil
}

// ScheduleString renders a schedule in the transition= grammar.
func ScheduleString(schedule pilot.Schedule) string {
	if schedule.Empty() {
		return "none"
	}
	switch schedule.Traversal {
	case pilot.Random:
		if equalEffectList(schedule.Transitions, transition.EffectIDs()) {
			return "random"
		}
		return "random:" + strings.Join(schedule.Transitions, ",")
	case pilot.Sequential:
		return strings.Join(schedule.Transitions, ",")
	default: // Single plays its one effect
		return schedule.Transitions[0]
	}
}

// parseEffectList splits and validates a comma list of effect IDs.
func parseEffectList(list string) ([]string, error) {
	var transitions []string
	for id := range strings.SplitSeq(list, ",") {
		id = canonicalEffectID(id)
		if !isValidEffectID(id) {
			return nil, fmt.Errorf("invalid transition: %s", id)
		}
		transitions = append(transitions, id)
	}
	if len(transitions) == 0 {
		return nil, fmt.Errorf("empty transition list")
	}
	return transitions, nil
}

// canonicalEffectID folds the legacy zoom_out spelling onto its
// effect ID.
func canonicalEffectID(id string) string {
	if id == "zoom_out" {
		return transition.ZoomOutID
	}
	return id
}

func isValidEffectID(id string) bool {
	return slices.Contains(transition.EffectIDs(), id)
}

func equalEffectList(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
