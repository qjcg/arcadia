package game

// Tile represents a tile type on the map.
type Tile int

const (
	TileWall Tile = iota
	TileFloor
	TileStreet
	TileDoor
	TileWater
	TileStairsUp
	TileStairsDown
	TileGrass
)

func (t Tile) IsBlocking() bool {
	return t == TileWall || t == TileWater
}

func (t Tile) IsOpaque() bool {
	return t == TileWall
}

// EntityType represents a type of entity.
type EntityType int

const (
	EntityPlayer EntityType = iota
	EntityChien
	EntityIvrogne
	EntityRatGeant
	EntityMachine
	EntitySquelette
	EntitySpectre
)

// Entity represents a creature in the game.
type Entity struct {
	Type    EntityType
	X, Y    int
	HP      int
	MaxHP   int
	Attack  int
	Defense int
	Symbol  rune
	Name    string
	NameFr  string
	IsEnemy bool
	Alive   bool
}

// ItemType represents a type of item.
type ItemType int

const (
	ItemBaguette ItemType = iota
	ItemCafe
	ItemBeret
	ItemPoutine
	ItemCle
	ItemSiropErable
	ItemPasseport
)

// Item represents an item in the game.
type Item struct {
	Type     ItemType
	X, Y     int
	OnGround bool
	Name     string
	NameFr   string
	Symbol   rune
	Heal     int
	DefBonus int
	Desc     string
}

// MonsterTemplate defines the stats for a monster type on a given level.
type MonsterTemplate struct {
	Type    EntityType
	Name    string
	NameFr  string
	Symbol  rune
	HP      int
	Attack  int
	Defense int
	MinLvl  int
	MaxLvl  int
}

// MonsterTemplates defines all monster types and when they appear.
var MonsterTemplates = []MonsterTemplate{
	{EntityChien, "stray dog", "Chien errant", 'd', 6, 2, 0, 1, 1},
	{EntityIvrogne, "drunkard", "Ivrogne", 'i', 8, 3, 1, 1, 1},
	{EntityRatGeant, "giant rat", "Rat géant", 'r', 6, 4, 0, 2, 2},
	{EntityMachine, "machine", "Machine", 'M', 14, 5, 2, 2, 2},
	{EntitySquelette, "skeleton", "Squelette", 's', 8, 4, 1, 3, 3},
	{EntitySpectre, "spectre", "Spectre", 'S', 10, 6, 0, 3, 3},
}

// ItemTemplate defines item placement on a given level.
type ItemTemplate struct {
	Type     ItemType
	Name     string
	NameFr   string
	Symbol   rune
	Heal     int
	DefBonus int
	Desc     string
	MinLvl   int
	MaxLvl   int
}

// ItemTemplates defines all item types and when they appear.
var ItemTemplates = []ItemTemplate{
	{ItemBaguette, "baguette", "Baguette", '%', 3, 0, "Un pain tout chaud. Soigne 3 PV.", 1, 3},
	{ItemCafe, "café", "Café", '[', 2, 0, "Un petit noir bien serré. Soigne 2 PV.", 1, 3},
	{ItemBeret, "beret", "Béret", '^', 0, 1, "Un vrai béret français. Défense +1.", 1, 1},
	{ItemPoutine, "poutine", "Poutine", '!', 8, 0, "Un goût de chez-soi. Soigne 8 PV.", 2, 2},
	{ItemCle, "key", "Clé", '$', 0, 0, "Une clé rouillée. Peut-être utile?", 2, 2},
	{ItemSiropErable, "maple syrup", "Sirop d'érable", '♥', 15, 0, "Authentique! Soigne 15 PV.", 3, 3},
	{ItemPasseport, "passport", "Passeport", '○', 0, 0, "Votre passeport pour rentrer au Canada!", 3, 3},
}

// GameState represents the current game mode.
type GameState int

const (
	StatePlaying GameState = iota
	StateDead
	StateWon
	StateHelp
	StateInventory
	StateQuit
)

// Room represents a rectangular room on the map.
type Room struct {
	X, Y, W, H int
}

// Center returns the center position of the room.
func (r Room) Center() (int, int) {
	return r.X + r.W/2, r.Y + r.H/2
}

// Overlaps checks if this room overlaps with another (with padding).
func (r Room) Overlaps(other Room) bool {
	const pad = 1
	return r.X-pad < other.X+other.W+pad &&
		r.X+r.W+pad > other.X-pad &&
		r.Y-pad < other.Y+other.H+pad &&
		r.Y+r.H+pad > other.Y-pad
}

// LevelInfo describes each game level.
type LevelInfo struct {
	Name     string
	NumRooms int
	Monsters int
	Items    int
}
