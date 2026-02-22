package render

import (
	"testing"

	"github.com/qjcg/arcadia/exp/fractalis/internal/core/fractal"
	"github.com/qjcg/arcadia/exp/fractalis/internal/core/persistence"
)

// Simple mock calculateFractal for benchmarking
func mockCalculateFractal(cr, ci float64, cfg persistence.Config) int {
	switch cfg.FractalType {
	case "mandelbrot":
		return int(fractal.Mandelbrot(cr, ci, cfg.MaxIter))
	default:
		return 0
	}
}

func BenchmarkRenderSerialVsParallel(b *testing.B) {
	config := persistence.Config{
		FractalType: "mandelbrot",
		CenterX:     -0.5,
		CenterY:     0.0,
		Zoom:        1.0,
		MaxIter:     50,
		Width:       80,
		Height:      40,
		ColorScheme: "grayscale",
	}

	b.Run("Serial", func(b *testing.B) {
		r := NewRenderer(config, mockCalculateFractal)
		r.SetParallel(false)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = r.RenderFractal()
		}
	})

	b.Run("Parallel", func(b *testing.B) {
		r := NewRenderer(config, mockCalculateFractal)
		r.SetParallel(true)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = r.RenderFractal()
		}
	})

	b.Run("ParallelLargeGrid", func(b *testing.B) {
		largeConfig := config
		largeConfig.Width = 160
		largeConfig.Height = 80 // 12800 pixels

		r := NewRenderer(largeConfig, mockCalculateFractal)
		r.SetParallel(true)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = r.RenderFractal()
		}
	})
}

func BenchmarkCacheEffectiveness(b *testing.B) {
	config := persistence.Config{
		FractalType: "mandelbrot",
		CenterX:     -0.5,
		CenterY:     0.0,
		Zoom:        1.0,
		MaxIter:     50,
		Width:       80,
		Height:      40,
		ColorScheme: "grayscale",
	}

	b.Run("WithCache", func(b *testing.B) {
		r := NewRenderer(config, mockCalculateFractal)
		r.SetParallel(false)
		b.ResetTimer()
		// First render fills cache
		_ = r.RenderFractal()
		b.ResetTimer() // Reset after first render

		for i := 0; i < b.N; i++ {
			_ = r.RenderFractal() // Subsequent renders hit cache
		}
	})

	b.Run("WithoutCache", func(b *testing.B) {
		r := NewRenderer(config, mockCalculateFractal)
		r.SetParallel(false)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = r.RenderFractal()
			r.ClearCache() // Force cache miss each time
		}
	})
}

func BenchmarkCalculateIterationsParallel(b *testing.B) {
	config := persistence.Config{
		FractalType: "mandelbrot",
		CenterX:     -0.5,
		CenterY:     0.0,
		Zoom:        1.0,
		MaxIter:     50,
		Width:       80,
		Height:      40,
	}

	r := NewRenderer(config, mockCalculateFractal)

	b.Run("Serial", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = r.calculateIterationsSerial()
		}
	})

	b.Run("Parallel", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = r.calculateIterationsParallel()
		}
	})
}
