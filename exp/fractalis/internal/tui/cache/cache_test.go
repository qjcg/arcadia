package cache

import (
	"testing"

	"github.com/qjcg/arcadia/exp/fractalis/internal/core/persistence"
)

func TestIterationCacheBasic(t *testing.T) {
	cache := NewIterationCache(2)

	config := persistence.Config{
		FractalType: "mandelbrot",
		CenterX:     -0.5,
		CenterY:     0.0,
		Zoom:        1.0,
		MaxIter:     50,
		Width:       80,
		Height:      40,
	}

	// Test miss on empty cache
	_, found := cache.Get(config)
	if found {
		t.Error("Expected miss on empty cache")
	}

	// Store a grid
	grid := make([][]int, 40)
	for i := range grid {
		grid[i] = make([]int, 80)
		for j := range grid[i] {
			grid[i][j] = i * j % 50
		}
	}

	cache.Set(config, grid)

	// Test hit
	retrieved, found := cache.Get(config)
	if !found {
		t.Error("Expected hit after Set")
	}

	if len(retrieved) != 40 || len(retrieved[0]) != 80 {
		t.Error("Retrieved grid has wrong dimensions")
	}
}

func TestIterationCacheEviction(t *testing.T) {
	cache := NewIterationCache(2) // Max 2 entries

	config1 := persistence.Config{
		FractalType: "mandelbrot",
		CenterX:     -0.5,
		CenterY:     0.0,
		Zoom:        1.0,
		MaxIter:     50,
		Width:       80,
		Height:      40,
	}

	config2 := persistence.Config{
		FractalType: "mandelbrot",
		CenterX:     -0.4,
		CenterY:     0.1,
		Zoom:        2.0,
		MaxIter:     50,
		Width:       80,
		Height:      40,
	}

	config3 := persistence.Config{
		FractalType: "mandelbrot",
		CenterX:     -0.3,
		CenterY:     0.2,
		Zoom:        4.0,
		MaxIter:     50,
		Width:       80,
		Height:      40,
	}

	grid := make([][]int, 40)
	for i := range grid {
		grid[i] = make([]int, 80)
	}

	// Store 3 configs but cache can only hold 2
	cache.Set(config1, grid)
	cache.Set(config2, grid)
	cache.Set(config3, grid) // Should evict config1

	// Verify cache contains only 2 items
	if len(cache.data) != 2 {
		t.Errorf("Expected cache to contain 2 items, got %d", len(cache.data))
	}

	// Verify config3 is still in cache
	_, found := cache.Get(config3)
	if !found {
		t.Error("Expected config3 to be in cache")
	}

	// Verify config1 was evicted (trying to get it should be a miss)
	_, found = cache.Get(config1)
	if found {
		t.Error("Expected config1 to be evicted from cache")
	}
}

func TestColorCacheBasic(t *testing.T) {
	cc := NewColorCache(10)

	// Test miss
	_, found := cc.Get("test_key")
	if found {
		t.Error("Expected miss on empty cache")
	}

	// Store value
	cc.Set("test_key", "\033[31m")

	// Test hit
	val, found := cc.Get("test_key")
	if !found {
		t.Error("Expected hit after Set")
	}

	if val != "\033[31m" {
		t.Errorf("Got wrong value: %s", val)
	}
}

func BenchmarkIterationCacheHit(b *testing.B) {
	cache := NewIterationCache(3)

	config := persistence.Config{
		FractalType: "mandelbrot",
		CenterX:     -0.5,
		CenterY:     0.0,
		Zoom:        1.0,
		MaxIter:     50,
		Width:       80,
		Height:      40,
	}

	grid := make([][]int, 40)
	for i := range grid {
		grid[i] = make([]int, 80)
	}

	cache.Set(config, grid)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Get(config)
	}
}

func BenchmarkCacheKeyGeneration(b *testing.B) {
	cache := NewIterationCache(3)

	config := persistence.Config{
		FractalType: "mandelbrot",
		CenterX:     -0.5,
		CenterY:     0.0,
		Zoom:        1.0,
		MaxIter:     50,
		Width:       80,
		Height:      40,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cache.cacheKey(config)
	}
}
