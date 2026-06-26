package game

import (
	"math"
)

const sightRange = 8

// computeFOV computes field-of-view for the player and marks visible tiles.
func computeFOV(m *GameMap, px, py int) {
	// Reset visibility
	for y := 0; y < m.H; y++ {
		for x := 0; x < m.W; x++ {
			m.Visible[y][x] = false
		}
	}

	// Mark player tile
	if px >= 0 && px < m.W && py >= 0 && py < m.H {
		m.Visible[py][px] = true
		m.Explored[py][px] = true
	}

	// Cast rays to every tile within sight range
	for y := py - sightRange; y <= py+sightRange; y++ {
		for x := px - sightRange; x <= px+sightRange; x++ {
			if x < 0 || x >= m.W || y < 0 || y >= m.H {
				continue
			}
			if !inSightRange(px, py, x, y) {
				continue
			}
			if hasLineOfSight(m, px, py, x, y) {
				m.Visible[y][x] = true
				m.Explored[y][x] = true
			}
		}
	}
}

func inSightRange(x0, y0, x1, y1 int) bool {
	dx := float64(x1 - x0)
	dy := float64(y1 - y0)
	return math.Sqrt(dx*dx+dy*dy) <= sightRange
}

func hasLineOfSight(m *GameMap, x0, y0, x1, y1 int) bool {
	dx := abs(x1 - x0)
	dy := abs(y1 - y0)
	sx := 1
	if x0 > x1 {
		sx = -1
	}
	sy := 1
	if y0 > y1 {
		sy = -1
	}
	err := dx - dy
	cx, cy := x0, y0

	for {
		if cx == x1 && cy == y1 {
			return true
		}
		if m.TileAt(cx, cy).IsOpaque() {
			return false
		}
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			cx += sx
		}
		if e2 < dx {
			err += dx
			cy += sy
		}
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
