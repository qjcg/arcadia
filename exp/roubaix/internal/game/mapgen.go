package game

import (
	"math/rand"
	"slices"
)

// Level info for each of the 3 floors.
var LevelInfos = []LevelInfo{
	{Name: "Roubaix — le centre-ville", NumRooms: 10, Monsters: 6, Items: 5},
	{Name: "La Vieille Filature", NumRooms: 8, Monsters: 8, Items: 4},
	{Name: "Les Catacombes", NumRooms: 7, Monsters: 8, Items: 4},
}

const (
	mapW = 62
	mapH = 18
)

// GameMap holds all map data for one level.
type GameMap struct {
	Tiles    [][]Tile
	Visible  [][]bool
	Explored [][]bool
	W, H     int
	Level    int
	Name     string
	Monsters []*Entity
	Items    []*Item
	Rooms    []Room
	UpPos    *Pos // stairs up location (nil for surface)
	DownPos  *Pos // stairs down location (nil for last level)
}

// Pos represents a map position.
type Pos struct {
	X, Y int
}

// NewMap generates a new map for the given level.
func NewMap(level int) *GameMap {
	info := LevelInfos[level]
	m := &GameMap{
		W:     mapW,
		H:     mapH,
		Level: level,
		Name:  info.Name,
	}

	m.Tiles = make([][]Tile, m.H)
	m.Visible = make([][]bool, m.H)
	m.Explored = make([][]bool, m.H)
	for y := range m.Tiles {
		m.Tiles[y] = make([]Tile, m.W)
		m.Visible[y] = make([]bool, m.W)
		m.Explored[y] = make([]bool, m.W)
		for x := range m.Tiles[y] {
			m.Tiles[y][x] = TileWall
		}
	}

	// Carve rooms
	m.generateRooms(info)

	// Place stairs
	m.placeStairs(level)

	// Populate monsters
	m.placeMonsters(info.Monsters)

	// Place items
	m.placeItems(info.Items)

	return m
}

func (m *GameMap) generateRooms(info LevelInfo) {
	for i := 0; i < 50 && len(m.Rooms) < info.NumRooms; i++ {
		w := 4 + rand.Intn(6)
		h := 3 + rand.Intn(4)
		x := 1 + rand.Intn(m.W-w-2)
		y := 1 + rand.Intn(m.H-h-2)
		r := Room{X: x, Y: y, W: w, H: h}

		overlap := slices.ContainsFunc(m.Rooms, r.Overlaps)
		if overlap {
			continue
		}

		// Different level types have different floor tiles
		tile := TileFloor
		if m.Level == 0 {
			// On the surface, most rooms have floor, some have grass patches
			tile = TileFloor
			// Streets between rooms on surface level
		}

		m.carveRoom(r, tile)
		m.Rooms = append(m.Rooms, r)
	}

	// Connect rooms with corridors
	for i := 1; i < len(m.Rooms); i++ {
		a := m.Rooms[i-1]
		b := m.Rooms[i]
		ax, ay := a.Center()
		bx, by := b.Center()

		carveTile := TileStreet
		if m.Level > 0 {
			carveTile = TileFloor
		}

		// L-shaped corridor
		if rand.Intn(2) == 0 {
			m.carveHCorridor(ax, bx, ay, carveTile)
			m.carveVCorridor(ay, by, bx, carveTile)
		} else {
			m.carveVCorridor(ay, by, ax, carveTile)
			m.carveHCorridor(ax, bx, by, carveTile)
		}
	}

	// Surface level: add some water/grass decorations
	if m.Level == 0 {
		m.decorateSurface()
	}
}

func (m *GameMap) carveRoom(r Room, tile Tile) {
	for y := r.Y; y < r.Y+r.H; y++ {
		for x := r.X; x < r.X+r.W; x++ {
			m.Tiles[y][x] = tile
		}
	}
	// Add door at room edges (random wall -> door)
	for x := r.X; x < r.X+r.W; x++ {
		doors := []struct{ x, y int }{}
		if r.Y > 0 {
			doors = append(doors, struct{ x, y int }{x, r.Y - 1})
		}
		if r.Y+r.H < m.H {
			doors = append(doors, struct{ x, y int }{x, r.Y + r.H})
		}
		for _, d := range doors {
			if d.x >= 0 && d.x < m.W && d.y >= 0 && d.y < m.H && m.Tiles[d.y][d.x] == TileWall {
				if rand.Intn(4) == 0 {
					m.Tiles[d.y][d.x] = TileDoor
				}
			}
		}
	}
	for y := r.Y; y < r.Y+r.H; y++ {
		doors := []struct{ x, y int }{}
		if r.X > 0 {
			doors = append(doors, struct{ x, y int }{r.X - 1, y})
		}
		if r.X+r.W < m.W {
			doors = append(doors, struct{ x, y int }{r.X + r.W, y})
		}
		for _, d := range doors {
			if d.x >= 0 && d.x < m.W && d.y >= 0 && d.y < m.H && m.Tiles[d.y][d.x] == TileWall {
				if rand.Intn(4) == 0 {
					m.Tiles[d.y][d.x] = TileDoor
				}
			}
		}
	}
}

func (m *GameMap) carveHCorridor(x1, x2, y int, tile Tile) {
	if x1 > x2 {
		x1, x2 = x2, x1
	}
	for x := x1; x <= x2; x++ {
		if y >= 0 && y < m.H && x >= 0 && x < m.W {
			if m.Tiles[y][x] == TileWall {
				m.Tiles[y][x] = tile
			}
		}
	}
}

func (m *GameMap) carveVCorridor(y1, y2, x int, tile Tile) {
	if y1 > y2 {
		y1, y2 = y2, y1
	}
	for y := y1; y <= y2; y++ {
		if y >= 0 && y < m.H && x >= 0 && x < m.W {
			if m.Tiles[y][x] == TileWall {
				m.Tiles[y][x] = tile
			}
		}
	}
}

func (m *GameMap) decorateSurface() {
	// Add patches of water (canal) and grass
	for range 6 {
		x := 10 + rand.Intn(m.W-20)
		y := 3 + rand.Intn(m.H-6)
		if m.Tiles[y][x] == TileWall {
			continue
		}
		// Small water feature
		for dy := -1; dy <= 1; dy++ {
			for dx := -1; dx <= 1; dx++ {
				nx, ny := x+dx, y+dy
				if nx >= 0 && nx < m.W && ny >= 0 && ny < m.H && m.Tiles[ny][nx] == TileStreet {
					if rand.Intn(3) == 0 {
						m.Tiles[ny][nx] = TileWater
					}
				}
			}
		}
	}
	// Add grass tiles randomly near walls
	for y := 1; y < m.H-1; y++ {
		for x := 1; x < m.W-1; x++ {
			if m.Tiles[y][x] == TileStreet && m.Tiles[y-1][x] == TileWall && rand.Intn(5) == 0 {
				m.Tiles[y][x] = TileGrass
			}
			if m.Tiles[y][x] == TileStreet && m.Tiles[y][x-1] == TileWall && rand.Intn(5) == 0 {
				m.Tiles[y][x] = TileGrass
			}
		}
	}
}

func (m *GameMap) placeStairs(level int) {
	// Place down stairs on all but last level
	if level < len(LevelInfos)-1 {
		r := m.Rooms[len(m.Rooms)-1]
		sx, sy := r.Center()
		m.Tiles[sy][sx] = TileStairsDown
		m.DownPos = &Pos{sx, sy}
	}

	// Place up stairs on all but first level
	if level > 0 {
		r := m.Rooms[0]
		sx, sy := r.Center()
		m.Tiles[sy][sx] = TileStairsUp
		m.UpPos = &Pos{sx, sy}
	}
}

func (m *GameMap) placeMonsters(count int) {
	for i := 0; i < count*2 && len(m.Monsters) < count; i++ {
		r := m.Rooms[1+rand.Intn(len(m.Rooms)-1)] // skip first room (player spawn)
		mx := r.X + 1 + rand.Intn(r.W-2)
		my := r.Y + 1 + rand.Intn(r.H-2)

		// Don't place on stairs or other monsters
		blocked := false
		for _, e := range m.Monsters {
			if e.X == mx && e.Y == my {
				blocked = true
				break
			}
		}
		if m.Tiles[my][mx].IsBlocking() || blocked {
			continue
		}

		// Pick a monster template valid for this level
		templates := []MonsterTemplate{}
		for _, t := range MonsterTemplates {
			if m.Level+1 >= t.MinLvl && m.Level+1 <= t.MaxLvl {
				templates = append(templates, t)
			}
		}
		if len(templates) == 0 {
			continue
		}
		t := templates[rand.Intn(len(templates))]
		m.Monsters = append(m.Monsters, &Entity{
			Type:    t.Type,
			X:       mx,
			Y:       my,
			HP:      t.HP,
			MaxHP:   t.HP,
			Attack:  t.Attack,
			Defense: t.Defense,
			Symbol:  t.Symbol,
			Name:    t.Name,
			NameFr:  t.NameFr,
			IsEnemy: true,
			Alive:   true,
		})
	}
}

func (m *GameMap) placeItems(count int) {
	for i := 0; i < count*3 && len(m.Items) < count; i++ {
		r := m.Rooms[rand.Intn(len(m.Rooms))]
		ix := r.X + 1 + rand.Intn(r.W-2)
		iy := r.Y + 1 + rand.Intn(r.H-2)

		if m.Tiles[iy][ix].IsBlocking() {
			continue
		}
		// Don't place on monsters
		blocked := false
		for _, e := range m.Monsters {
			if e.X == ix && e.Y == iy {
				blocked = true
				break
			}
		}
		if blocked {
			continue
		}

		templates := []ItemTemplate{}
		for _, t := range ItemTemplates {
			if m.Level+1 >= t.MinLvl && m.Level+1 <= t.MaxLvl {
				templates = append(templates, t)
			}
		}
		if len(templates) == 0 {
			continue
		}
		t := templates[rand.Intn(len(templates))]
		m.Items = append(m.Items, &Item{
			Type:     t.Type,
			X:        ix,
			Y:        iy,
			OnGround: true,
			Name:     t.Name,
			NameFr:   t.NameFr,
			Symbol:   t.Symbol,
			Heal:     t.Heal,
			DefBonus: t.DefBonus,
			Desc:     t.Desc,
		})
	}
}

// TileAt returns the tile at (x,y), or TileWall if out of bounds.
func (m *GameMap) TileAt(x, y int) Tile {
	if x < 0 || x >= m.W || y < 0 || y >= m.H {
		return TileWall
	}
	return m.Tiles[y][x]
}

// IsWalkable checks if a tile can be walked on.
func (m *GameMap) IsWalkable(x, y int) bool {
	if x < 0 || x >= m.W || y < 0 || y >= m.H {
		return false
	}
	if m.Tiles[y][x].IsBlocking() {
		return false
	}
	for _, e := range m.Monsters {
		if e.Alive && e.X == x && e.Y == y {
			return false
		}
	}
	return true
}
