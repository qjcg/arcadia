package game

import (
	tea "charm.land/bubbletea/v2"
)

// Game is the main bubbletea model for the roguelike.
type Game struct {
	player       Player
	maps         []*GameMap
	currentLevel int
	state        GameState
	inventory    []Item
	messages     []string
	turns        int
	termW        int
	termH        int
}

// Player holds player-specific state.
type Player struct {
	X, Y    int
	HP      int
	MaxHP   int
	Attack  int
	Defense int
}

// CurrentMap returns the active map.
func (g *Game) CurrentMap() *GameMap {
	return g.maps[g.currentLevel]
}

// New creates and initializes a new Game.
func New() *Game {
	g := &Game{
		player: Player{
			HP:      20,
			MaxHP:   20,
			Attack:  3,
			Defense: 1,
		},
		currentLevel: 0,
		state:        StatePlaying,
		messages:     []string{},
		turns:        0,
		termW:        80,
		termH:        24,
	}

	// Generate all maps
	for level := 0; level < len(LevelInfos); level++ {
		m := NewMap(level)
		g.maps = append(g.maps, m)
	}

	// Place player in the first room of level 0
	m := g.maps[0]
	if len(m.Rooms) > 0 {
		r := m.Rooms[0]
		g.player.X, g.player.Y = r.Center()
	}

	computeFOV(m, g.player.X, g.player.Y)
	g.addMessage("Vous vous réveillez dans une chambre inconnue à Roubaix.")
	g.addMessage("Un froid humide traverse les murs. Il faut sortir d'ici.")
	g.addMessage("Comment allez-vous rentrer au Canada?")

	return g
}

// Init implements tea.Model.
func (g *Game) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (g *Game) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		return g.handleKey(msg)
	case tea.WindowSizeMsg:
		g.termW = msg.Width
		g.termH = msg.Height
	}
	return g, nil
}

// handleKey processes keyboard input.
func (g *Game) handleKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	// Global keys (work in all states)
	if key == "ctrl+c" {
		return g, tea.Quit
	}

	switch g.state {
	case StatePlaying:
		return g.handlePlayingKey(key)
	case StateHelp:
		if key == "?" || key == "escape" || key == "enter" || key == "q" {
			g.state = StatePlaying
		}
		return g, nil
	case StateInventory:
		if key == "i" || key == "escape" || key == "enter" {
			g.state = StatePlaying
			return g, nil
		}
		// Use item by key
		if len(key) == 1 && key[0] >= 'a' && key[0] <= 'z' {
			idx := int(key[0] - 'a')
			g.useItem(idx)
			g.state = StatePlaying
		}
		return g, nil
	case StateDead:
		if key == "q" || key == "enter" {
			g.state = StateQuit
			return g, tea.Quit
		}
		return g, nil
	case StateWon:
		if key == "q" || key == "enter" {
			g.state = StateQuit
			return g, tea.Quit
		}
		return g, nil
	}
	return g, nil
}

// handlePlayingKey processes keys during normal gameplay.
func (g *Game) handlePlayingKey(key string) (tea.Model, tea.Cmd) {
	dx, dy := 0, 0
	action := false

	switch key {
	case "h", "left":
		dx = -1
	case "j", "down":
		dy = 1
	case "k", "up":
		dy = -1
	case "l", "right":
		dx = 1
	case "y":
		dx = -1
		dy = -1
	case "u":
		dx = 1
		dy = -1
	case "b":
		dx = -1
		dy = 1
	case "n":
		dx = 1
		dy = 1
	case ".":
		action = true
		g.addMessage("Un tour passe...")
	case ">":
		return g, g.goDown()
	case "<":
		return g, g.goUp()
	case "g":
		g.tryPickup()
		action = true
	case "d":
		if len(g.inventory) > 0 {
			g.dropItem(0)
		} else {
			g.addMessage("Vous n'avez rien à déposer.")
		}
		return g, nil
	case "i":
		g.state = StateInventory
		return g, nil
	case "?":
		g.state = StateHelp
		return g, nil
	case "q":
		g.state = StateQuit
		return g, tea.Quit
	default:
		return g, nil
	}

	// Handle movement or waiting
	if dx != 0 || dy != 0 {
		g.tryMove(dx, dy)
		action = true
	}

	if action {
		g.turns++
		g.monsterTurn()
		computeFOV(g.CurrentMap(), g.player.X, g.player.Y)
	}

	return g, nil
}

// tryMove attempts to move the player by (dx, dy).
func (g *Game) tryMove(dx, dy int) {
	m := g.CurrentMap()
	nx, ny := g.player.X+dx, g.player.Y+dy

	if nx < 0 || nx >= m.W || ny < 0 || ny >= m.H {
		return
	}
	if m.Tiles[ny][nx].IsBlocking() {
		if m.Tiles[ny][nx] == TileDoor {
			g.addMessage("Cette porte est verrouillée.")
		} else {
			g.addMessage("Il y a un obstacle.")
		}
		return
	}

	// Check for enemies
	for _, e := range m.Monsters {
		if e.Alive && e.X == nx && e.Y == ny {
			g.meleeAttack(e)
			return
		}
	}

	// Move player
	g.player.X = nx
	g.player.Y = ny

	// Check for items on ground
	for _, item := range m.Items {
		if item.OnGround && item.X == nx && item.Y == ny {
			g.addMessage("Vous voyez " + item.NameFr + " (" + item.Desc + ")")
		}
	}

	// Check for stairs
	if m.Tiles[ny][nx] == TileStairsDown {
		g.addMessage("Un escalier descend dans les profondeurs. Appuyez sur > pour descendre.")
	}
	if m.Tiles[ny][nx] == TileStairsUp {
		g.addMessage("Un escalier monte vers la surface. Appuyez sur < pour monter.")
	}
}

// meleeAttack attacks an entity.
func (g *Game) meleeAttack(e *Entity) {
	damage := g.player.Attack - e.Defense
	if damage < 1 {
		damage = 1
	}
	e.HP -= damage
	g.addMessage("Vous frappez " + e.NameFr + " pour " + itoa(damage) + " dégâts!")

	if e.HP <= 0 {
		e.Alive = false
		g.addMessage("Vous avez tué " + e.NameFr + "!")
	}
}

// monsterTurn processes all monster actions.
func (g *Game) monsterTurn() {
	m := g.CurrentMap()
	for _, e := range m.Monsters {
		if !e.Alive {
			continue
		}

		// Skip if not visible to player (monster doesn't "see" player)
		if !m.Visible[e.Y][e.X] {
			continue
		}

		// Simple AI: move toward player if adjacent
		dx := g.player.X - e.X
		dy := g.player.Y - e.Y

		if abs(dx) <= 1 && abs(dy) <= 1 {
			// Adjacent: attack player
			damage := e.Attack - g.player.Defense
			if damage < 1 {
				damage = 1
			}
			g.player.HP -= damage
			g.addMessage(e.NameFr + " vous attaque pour " + itoa(damage) + " dégâts!")
			if g.player.HP <= 0 {
				g.player.HP = 0
				g.state = StateDead
				g.addMessage("Vous êtes mort...")
				g.addMessage("Personne ne vous retrouvera dans ce trou perdu.")
			}
		} else {
			// Move toward player
			mdx, mdy := 0, 0
			if abs(dx) > abs(dy) {
				mdx = sign(dx)
			} else {
				mdy = sign(dy)
			}
			nx, ny := e.X+mdx, e.Y+mdy
			if m.IsWalkable(nx, ny) && !(nx == g.player.X && ny == g.player.Y) {
				e.X = nx
				e.Y = ny
			} else {
				// Try other direction
				if mdx != 0 {
					mdy = sign(dy)
				} else {
					mdx = sign(dx)
				}
				nx, ny = e.X+mdx, e.Y+mdy
				if m.IsWalkable(nx, ny) && !(nx == g.player.X && ny == g.player.Y) {
					e.X = nx
					e.Y = ny
				}
			}
		}
	}
}

// goDown moves the player down stairs.
func (g *Game) goDown() tea.Cmd {
	m := g.CurrentMap()
	if m.Tiles[g.player.Y][g.player.X] != TileStairsDown {
		g.addMessage("Il n'y a pas d'escalier ici.")
		return nil
	}
	if g.currentLevel >= len(g.maps)-1 {
		g.addMessage("Vous ne pouvez pas descendre plus profond.")
		return nil
	}
	g.currentLevel++
	g.addMessage("Vous descendez dans " + g.CurrentMap().Name + ".")
	// Place player at up stairs of new level
	nm := g.CurrentMap()
	if nm.UpPos != nil {
		g.player.X = nm.UpPos.X
		g.player.Y = nm.UpPos.Y
	}
	computeFOV(nm, g.player.X, g.player.Y)
	return nil
}

// goUp moves the player up stairs.
func (g *Game) goUp() tea.Cmd {
	m := g.CurrentMap()
	if m.Tiles[g.player.Y][g.player.X] != TileStairsUp {
		g.addMessage("Il n'y a pas d'escalier ici.")
		return nil
	}
	if g.currentLevel <= 0 {
		g.addMessage("Vous êtes déjà à la surface.")
		return nil
	}
	g.currentLevel--
	g.addMessage("Vous montez vers " + g.CurrentMap().Name + ".")
	nm := g.CurrentMap()
	if nm.DownPos != nil {
		g.player.X = nm.DownPos.X
		g.player.Y = nm.DownPos.Y
	}
	computeFOV(nm, g.player.X, g.player.Y)
	return nil
}

// tryPickup picks up items at the player's position.
func (g *Game) tryPickup() {
	m := g.CurrentMap()
	px, py := g.player.X, g.player.Y
	found := false
	for i := len(m.Items) - 1; i >= 0; i-- {
		item := m.Items[i]
		if !item.OnGround {
			continue
		}
		if item.X != px || item.Y != py {
			continue
		}
		found = true
		// Auto-use consumables
		if item.Heal > 0 || item.DefBonus > 0 {
			if item.Heal > 0 {
				g.player.HP += item.Heal
				if g.player.HP > g.player.MaxHP {
					g.player.HP = g.player.MaxHP
				}
				g.addMessage("Vous ramassez et utilisez " + item.NameFr + "! " + item.Desc)
			}
			if item.DefBonus > 0 {
				g.player.Defense += item.DefBonus
				g.addMessage("Vous portez " + item.NameFr + ". " + item.Desc)
			}
			// Special: passport wins the game
			if item.Type == ItemPasseport {
				g.state = StateWon
				g.addMessage("VOUS AVEZ TROUVÉ VOTRE PASSEPORT!")
				g.addMessage("Vous pouvez enfin rentrer au Canada! Vous avez gagné!")
			}
			// Remove from ground
			m.Items = append(m.Items[:i], m.Items[i+1:]...)
		} else {
			// Add to inventory
			item.OnGround = false
			g.inventory = append(g.inventory, *item)
			m.Items = append(m.Items[:i], m.Items[i+1:]...)
			g.addMessage("Vous ramassez " + item.NameFr + ".")
		}
	}
	if !found {
		g.addMessage("Il n'y a rien à ramasser ici.")
	}
}

// useItem uses an item from inventory at the given index.
func (g *Game) useItem(idx int) {
	if idx < 0 || idx >= len(g.inventory) {
		g.addMessage("Rien à utiliser.")
		return
	}
	item := g.inventory[idx]
	if item.Heal > 0 {
		g.player.HP += item.Heal
		if g.player.HP > g.player.MaxHP {
			g.player.HP = g.player.MaxHP
		}
		g.addMessage("Vous utilisez " + item.NameFr + ". " + item.Desc)
		g.inventory = append(g.inventory[:idx], g.inventory[idx+1:]...)
	} else if item.DefBonus > 0 {
		g.player.Defense += item.DefBonus
		g.addMessage("Vous portez " + item.NameFr + ". " + item.Desc)
		g.inventory = append(g.inventory[:idx], g.inventory[idx+1:]...)
	} else if item.Type == ItemCle {
		g.addMessage(item.NameFr + " semble ne servir à rien ici...")
	} else {
		g.addMessage("Vous ne pouvez pas utiliser " + item.NameFr + " maintenant.")
	}
}

// addMessage adds a message to the log.
func (g *Game) addMessage(msg string) {
	g.messages = append(g.messages, msg)
	if len(g.messages) > 100 {
		g.messages = g.messages[len(g.messages)-100:]
	}
}

// Floor returns the current floor number (1-based).
func (g *Game) Floor() int {
	return g.currentLevel + 1
}

// TotalFloors returns total number of floors.
func (g *Game) TotalFloors() int {
	return len(g.maps)
}

// View implements tea.Model. Returns the rendered terminal view.
func (g *Game) View() tea.View {
	return tea.NewView(g.render())
}

// Helper: integer to string (no strconv import needed)
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	s := ""
	neg := false
	if n < 0 {
		neg = true
		n = -n
	}
	for n > 0 {
		s = string(rune('0'+n%10)) + s
		n /= 10
	}
	if neg {
		s = "-" + s
	}
	return s
}

// sign returns -1, 0, or 1.
func sign(x int) int {
	if x < 0 {
		return -1
	}
	if x > 0 {
		return 1
	}
	return 0
}

// dropItem removes an item from inventory and places it on the ground.
func (g *Game) dropItem(idx int) {
	if idx < 0 || idx >= len(g.inventory) {
		return
	}
	item := g.inventory[idx]
	item.OnGround = true
	item.X = g.player.X
	item.Y = g.player.Y
	g.CurrentMap().Items = append(g.CurrentMap().Items, &item)
	g.inventory = append(g.inventory[:idx], g.inventory[idx+1:]...)
	g.addMessage("Vous déposez " + item.NameFr + ".")
}
