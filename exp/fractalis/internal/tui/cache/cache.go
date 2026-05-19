package cache

import (
	"crypto/md5"
	"fmt"
	"sync"

	"github.com/qjcg/arcadia/exp/fractalis/internal/core/persistence"
)

// IterationCache caches fractal iteration results to avoid recomputation
type IterationCache struct {
	mu    sync.RWMutex
	data  map[string][][]int // Key: hash of config + viewport, Value: grid of iterations
	sizes map[string]struct{ w, h int }
	// Control cache growth
	maxEntries int
	hits       int64
	misses     int64
}

// NewIterationCache creates a new cache with a max number of stored viewports
func NewIterationCache(maxEntries int) *IterationCache {
	return &IterationCache{
		data:       make(map[string][][]int),
		sizes:      make(map[string]struct{ w, h int }),
		maxEntries: maxEntries,
	}
}

// cacheKey generates a unique key for a viewport configuration
// Uses MD5 hash of config parameters to create a compact key
func (ic *IterationCache) cacheKey(config persistence.Config) string {
	h := md5.New()
	// Include all relevant fractal parameters and viewport settings
	fmt.Fprintf(
		h, "%s:%.16f,%.16f,%.16f:%d:%d:%d:%d",
		config.FractalType,
		config.CenterX, config.CenterY, config.Zoom,
		config.MaxIter, config.Width, config.Height,
		int(config.JuliaCr*10000)<<16|int(config.JuliaCi*10000), // Simplified Julia params hash
	)
	return fmt.Sprintf("%x", h.Sum(nil))[:16] // Use first 16 chars of hash
}

// Get retrieves cached iteration grid if available
func (ic *IterationCache) Get(config persistence.Config) ([][]int, bool) {
	ic.mu.RLock()
	defer ic.mu.RUnlock()

	key := ic.cacheKey(config)
	data, exists := ic.data[key]
	if exists {
		ic.hits++
	} else {
		ic.misses++
	}
	return data, exists
}

// Set stores an iteration grid in cache
func (ic *IterationCache) Set(config persistence.Config, grid [][]int) {
	ic.mu.Lock()
	defer ic.mu.Unlock()

	key := ic.cacheKey(config)

	// If cache full, remove an old entry (simple FIFO eviction)
	if len(ic.data) >= ic.maxEntries && ic.data[key] == nil {
		// Find and remove first key (simple eviction)
		for k := range ic.data {
			delete(ic.data, k)
			delete(ic.sizes, k)
			break
		}
	}

	ic.data[key] = grid
	if len(grid) > 0 {
		ic.sizes[key] = struct{ w, h int }{len(grid[0]), len(grid)}
	}
}

// Clear empties the cache
func (ic *IterationCache) Clear() {
	ic.mu.Lock()
	defer ic.mu.Unlock()
	ic.data = make(map[string][][]int)
	ic.sizes = make(map[string]struct{ w, h int })
}

// Stats returns cache hit/miss statistics
func (ic *IterationCache) Stats() (hits, misses int64, hitRate float64) {
	ic.mu.RLock()
	defer ic.mu.RUnlock()
	total := ic.hits + ic.misses
	if total > 0 {
		hitRate = float64(ic.hits) / float64(total)
	}
	return ic.hits, ic.misses, hitRate
}

// ColorCache caches color string lookups to avoid repeated theme lookups
type ColorCache struct {
	mu    sync.RWMutex
	cache map[string]string
	max   int
}

// NewColorCache creates a new color cache
func NewColorCache(maxEntries int) *ColorCache {
	return &ColorCache{
		cache: make(map[string]string),
		max:   maxEntries,
	}
}

// Get retrieves a cached color string
func (cc *ColorCache) Get(key string) (string, bool) {
	cc.mu.RLock()
	defer cc.mu.RUnlock()
	val, exists := cc.cache[key]
	return val, exists
}

// Set stores a color string in cache
func (cc *ColorCache) Set(key, value string) {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	// Simple eviction if full
	if len(cc.cache) >= cc.max && cc.cache[key] == "" {
		for k := range cc.cache {
			delete(cc.cache, k)
			break
		}
	}
	cc.cache[key] = value
}

// Clear empties the cache
func (cc *ColorCache) Clear() {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	cc.cache = make(map[string]string)
}
