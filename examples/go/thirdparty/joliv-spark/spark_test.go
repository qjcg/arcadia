package spark

import (
	"testing"

	"github.com/joliv/spark"
)

func TestLine(t *testing.T) {
	testData := []struct{
		description string
		data []float64
		want string
	}{
		{"empty", []float64{}, ""},
		{"basic-ascending", []float64{1, 2, 3, 4, 5, 6, 7, 8}, "▁▂▃▄▅▆▇█"},
		{"basic-descending", []float64{8, 7, 6, 5, 4, 3, 2, 1}, "█▇▆▅▄▃▂▁"},
		{"basic-trio", []float64{1, 2, 3}, "▁▅█"},
		{"temps", []float64{67, 71, 77, 85, 95, 104, 106, 105, 100, 89, 76, 66}, "▁▂▃▄▆███▇▅▃▁"},

	}

	for _, tc := range testData {
		t.Run(tc.description, func(t *testing.T) {
			got := spark.Line(tc.data)

			if got != tc.want {
				t.Fatalf("want %q got %q", tc.want, got)
			}

			t.Logf("data %v, sparkline %q", tc.data, got)

		})
	}
}
