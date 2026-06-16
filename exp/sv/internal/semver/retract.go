package semver

import (
	"strings"

	"github.com/Masterminds/semver/v3"
)

// Retraction represents a single retracted version or range of versions.
type Retraction struct {
	Low  string // e.g., "v1.0.0"
	High string // e.g., "v1.9.9" (same as Low for single versions)
}

// ParseRetractions parses retract directives from go.mod content.
func ParseRetractions(goModContent string) []Retraction {
	lines := strings.Split(goModContent, "\n")
	var retractions []Retraction

	inBlock := false
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip blank lines and comments
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}

		if line == "retract (" {
			inBlock = true
			continue
		}

		if inBlock {
			if line == ")" {
				inBlock = false
				continue
			}
			// Parse individual version line inside block
			r := parseRetractLine(line)
			if r != nil {
				retractions = append(retractions, *r)
			}
			continue
		}

		// Outside block, look for "retract <spec>"
		if strings.HasPrefix(line, "retract ") {
			spec, _ := strings.CutPrefix(line, "retract ")
			spec = strings.TrimSpace(spec)
			r := parseRetractSpec(spec)
			if r != nil {
				retractions = append(retractions, *r)
			}
		}
	}

	return retractions
}

// parseRetractLine parses a single line from inside a retract block.
func parseRetractLine(line string) *Retraction {
	line = strings.TrimSpace(line)
	// Strip trailing comments
	if idx := strings.Index(line, "//"); idx >= 0 {
		line = strings.TrimSpace(line[:idx])
	}
	if line == "" {
		return nil
	}
	return parseRetractSpec(line)
}

// parseRetractSpec parses a version spec: "v1.0.0" or "[v1.0.0, v1.9.9]"
func parseRetractSpec(spec string) *Retraction {
	spec = strings.TrimSpace(spec)

	if strings.HasPrefix(spec, "[") && strings.HasSuffix(spec, "]") {
		inner := spec[1 : len(spec)-1]
		parts := strings.SplitN(inner, ",", 2)
		if len(parts) != 2 {
			return nil
		}
		low := strings.TrimSpace(parts[0])
		high := strings.TrimSpace(parts[1])
		if !strings.HasPrefix(low, "v") {
			low = "v" + low
		}
		if !strings.HasPrefix(high, "v") {
			high = "v" + high
		}
		return &Retraction{Low: low, High: high}
	}

	// Single version
	if !strings.HasPrefix(spec, "v") {
		spec = "v" + spec
	}
	return &Retraction{Low: spec, High: spec}
}

// IsVersionRetracted checks if a version (e.g., "v1.0.0") is covered by any retraction.
// The version should be a canonical semver without module path prefix.
func IsVersionRetracted(version string, retractions []Retraction) bool {
	v, err := semver.NewVersion(version)
	if err != nil {
		return false
	}

	for _, r := range retractions {
		low, errLow := semver.NewVersion(r.Low)
		high, errHigh := semver.NewVersion(r.High)
		if errLow != nil || errHigh != nil {
			continue
		}
		// Check if version is in [low, high]
		if (v.Equal(low) || v.GreaterThan(low) || v.Equal(high)) &&
			(v.Equal(high) || v.LessThan(high) || v.Equal(low)) {
			// Simplified: v >= low && v <= high
			if (v.GreaterThan(low) || v.Equal(low)) && (v.LessThan(high) || v.Equal(high)) {
				return true
			}
		}
	}

	return false
}
