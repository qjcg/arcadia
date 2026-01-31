package precision

import "math"

// FloatEqual compares two floats with an epsilon tolerance
func FloatEqual(a, b, epsilon float64) bool {
	return math.Abs(a-b) < epsilon
}

// IsNearZero checks if a float is close to zero
func IsNearZero(f, epsilon float64) bool {
	return math.Abs(f) < epsilon
}

// Statistics holds computed statistics for a dataset
type Statistics struct {
	Min      int
	Max      int
	Mean     float64
	Variance float64
	StdDev   float64
	Range    float64
}

// CalculateDistribution computes statistics for a slice of integers
func CalculateDistribution(samples []int) Statistics {
	if len(samples) == 0 {
		return Statistics{}
	}

	min := samples[0]
	max := samples[0]
	sum := 0

	for _, val := range samples {
		if val < min {
			min = val
		}
		if val > max {
			max = val
		}
		sum += val
	}

	mean := float64(sum) / float64(len(samples))
	variance := 0.0
	for _, val := range samples {
		diff := float64(val) - mean
		variance += diff * diff
	}
	variance /= float64(len(samples))
	stdDev := math.Sqrt(variance)

	return Statistics{
		Min:      min,
		Max:      max,
		Mean:     mean,
		Variance: variance,
		StdDev:   stdDev,
		Range:    float64(max - min),
	}
}
