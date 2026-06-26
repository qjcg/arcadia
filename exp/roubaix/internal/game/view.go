package game

import (
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"
)

// Tile display characters and colors.
type tileDisplay struct {
	Char  string
	Color color.Color
}

var tileStyles = map[Tile]tileDisplay{
	TileWall:       {"#", lipgloss.Color("94")},
	TileFloor:      {".", lipgloss.Color("248")},
	TileStreet:     {",", lipgloss.Color("242")},
	TileDoor:       {"+", lipgloss.Color("130")},
	TileWater:      {"~", lipgloss.Color("27")},
	TileGrass:      {"\"", lipgloss.Color("34")},
	TileStairsDown: {">", lipgloss.Color("51")},
	TileStairsUp:   {"<", lipgloss.Color("51")},
}

var (
	unexploredStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("236"))
	exploredStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("238"))
)

// Entity colors.
var entityColors = map[EntityType]color.Color{
	EntityChien:     lipgloss.Color("208"),
	EntityIvrogne:   lipgloss.Color("196"),
	EntityRatGeant:  lipgloss.Color("226"),
	EntityMachine:   lipgloss.Color("45"),
	EntitySquelette: lipgloss.Color("248"),
	EntitySpectre:   lipgloss.Color("201"),
}

var playerColor = lipgloss.Color("231")

// Item colors.
var itemColors = map[ItemType]color.Color{
	ItemBaguette:    lipgloss.Color("220"),
	ItemCafe:        lipgloss.Color("130"),
	ItemBeret:       lipgloss.Color("27"),
	ItemPoutine:     lipgloss.Color("34"),
	ItemCle:         lipgloss.Color("226"),
	ItemSiropErable: lipgloss.Color("196"),
	ItemPasseport:   lipgloss.Color("231"),
}

// render produces the full terminal view.
func (g *Game) render() string {
	var b strings.Builder

	// Determine map view dimensions
	mapW := g.CurrentMap().W
	mapH := g.CurrentMap().H

	// Draw map
	for y := range mapH {
		for x := range mapW {
			b.WriteString(g.renderTile(x, y))
		}
		b.WriteByte('\n')
	}

	// Status area
	b.WriteString(g.renderStatusBar())
	b.WriteString("\n")
	b.WriteString(g.renderInventoryLine())
	b.WriteString("\n")
	b.WriteString(g.renderMessage())
	b.WriteString("\n")
	b.WriteString(g.renderControls())

	// Overlays
	switch g.state {
	case StateHelp:
		return g.renderHelp()
	case StateDead:
		return g.renderDeadScreen()
	case StateWon:
		return g.renderWinScreen()
	case StateInventory:
		return g.renderInventoryScreen()
	}

	return b.String()
}

func (g *Game) renderTile(x, y int) string {
	m := g.CurrentMap()

	if !m.Visible[y][x] && !m.Explored[y][x] {
		return " "
	}

	// Determine the base tile
	var tileChar string
	var tileColor color.Color
	tile := m.Tiles[y][x]

	if info, ok := tileStyles[tile]; ok {
		tileChar = info.Char
		tileColor = info.Color
	} else {
		tileChar = "?"
		tileColor = lipgloss.Color("196")
	}

	// Dim explored-but-not-visible tiles
	if !m.Visible[y][x] {
		return exploredStyle.Render(tileChar)
	}

	// Player
	if x == g.player.X && y == g.player.Y {
		return lipgloss.NewStyle().Foreground(playerColor).Bold(true).Render("@")
	}

	// Check for monsters (visible only)
	for _, e := range m.Monsters {
		if e.Alive && e.X == x && e.Y == y {
			if c, ok := entityColors[e.Type]; ok {
				return lipgloss.NewStyle().Foreground(c).Bold(true).Render(string(e.Symbol))
			}
			return string(e.Symbol)
		}
	}

	// Check for items on ground (visible only)
	for _, item := range m.Items {
		if item.OnGround && item.X == x && item.Y == y {
			if c, ok := itemColors[item.Type]; ok {
				return lipgloss.NewStyle().Foreground(c).Render(string(item.Symbol))
			}
			return string(item.Symbol)
		}
	}

	// Normal tile rendering
	return lipgloss.NewStyle().Foreground(tileColor).Render(tileChar)
}

func (g *Game) renderStatusBar() string {
	hp := max(g.player.HP, 0)
	maxHP := g.player.MaxHP
	hpPct := float64(hp) / float64(maxHP)
	barW := 16
	filled := int(float64(barW) * hpPct)
	bar := strings.Repeat("█", filled) + strings.Repeat("░", barW-filled)

	hpColor := lipgloss.Color("46") // green
	if hpPct < 0.3 {
		hpColor = lipgloss.Color("196") // red
	} else if hpPct < 0.6 {
		hpColor = lipgloss.Color("226") // yellow
	}

	hpStyle := lipgloss.NewStyle().Foreground(hpColor).Bold(true)
	infoStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("248"))
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("244"))

	hpStr := hpStyle.Render(bar) + " " + infoStyle.Render(itoa(hp)+"/"+itoa(maxHP))
	atkStr := infoStyle.Render("ATK:") + labelStyle.Render(itoa(g.player.Attack))
	defStr := infoStyle.Render("DEF:") + labelStyle.Render(itoa(g.player.Defense))
	floorStr := infoStyle.Render("Étage:") + labelStyle.Render(g.CurrentMap().Name)
	turnStr := infoStyle.Render("T:") + labelStyle.Render(itoa(g.turns))

	return " PV: " + hpStr + " | " + atkStr + " " + defStr + " | " + floorStr + " | " + turnStr
}

func (g *Game) renderInventoryLine() string {
	if len(g.inventory) == 0 {
		return ""
	}
	items := []string{}
	for _, item := range g.inventory {
		items = append(items, item.NameFr)
	}
	style := lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Italic(true)
	return " Sac: " + style.Render(strings.Join(items, ", "))
}

func (g *Game) renderMessage() string {
	if len(g.messages) == 0 {
		return ""
	}
	msg := g.messages[len(g.messages)-1]
	style := lipgloss.NewStyle().Foreground(lipgloss.Color("255"))
	return " ▶ " + style.Render(msg)
}

func (g *Game) renderControls() string {
	keyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("45"))
	sepStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	keys := []string{
		keyStyle.Render("h") + sepStyle.Render("←") + keyStyle.Render("j") + sepStyle.Render("↓") + keyStyle.Render("k") + sepStyle.Render("↑") + keyStyle.Render("l") + sepStyle.Render("→"),
		keyStyle.Render("Maj+") + sepStyle.Render("courir"),
		keyStyle.Render("g") + sepStyle.Render("ramasser"),
		keyStyle.Render("i") + sepStyle.Render("inventaire"),
		keyStyle.Render(">") + sepStyle.Render("descendre"),
		keyStyle.Render("?") + sepStyle.Render("aide"),
		keyStyle.Render("q") + sepStyle.Render("quitter"),
	}
	return strings.Join(keys, "  ")
}

func (g *Game) renderHelp() string {
	var b strings.Builder
	titleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("51")).Bold(true).Underline(true)
	keyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("45")).Bold(true)
	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("248"))

	b.WriteString(titleStyle.Render("AIDE — ROULETTE ROUBAIXIENNE"))
	b.WriteString("\n\n")

	helpItems := []struct{ keys, desc string }{
		{"h j k l  /  ← ↑ ↓ →", "Déplacement (vi + flèches)"},
		{"y u b n", "Déplacement diagonal"},
		{".", "Attendre un tour"},
		{"Maj + ← ↑ ↓ →", "Courir tout droit (flèches)"},
		{"Maj + H J K L Y U B N", "Courir tout droit (vi)"},
		{"g", "Ramasser un objet"},
		{"i", "Inventaire (voir/utiliser)"},
		{"d", "Déposer un objet"},
		{">", "Descendre un escalier"},
		{"<", "Monter un escalier"},
		{"?", "Cette aide"},
		{"q", "Quitter"},
	}

	for _, h := range helpItems {
		b.WriteString("  " + keyStyle.Render(h.keys))
		b.WriteString("  ")
		b.WriteString(descStyle.Render(h.desc))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(descStyle.Render("Appuyez sur ? pour fermer cette aide."))
	b.WriteString("\n")

	pad := strings.Repeat("\n", g.termH-6-len(helpItems))
	b.WriteString(pad)
	return b.String()
}

func (g *Game) renderInventoryScreen() string {
	var b strings.Builder
	titleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("51")).Bold(true).Underline(true)
	keyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("45")).Bold(true)

	b.WriteString(titleStyle.Render("INVENTAIRE"))
	b.WriteString("\n\n")

	if len(g.inventory) == 0 {
		b.WriteString("  (vide)")
		b.WriteString("\n")
	} else {
		for i, item := range g.inventory {
			letter := string(rune('a' + i))
			b.WriteString("  " + keyStyle.Render(letter) + ") ")
			b.WriteString(item.NameFr)
			b.WriteString(" — ")
			b.WriteString(item.Desc)
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("248")).Render("Appuyez sur une lettre pour utiliser. [i] pour fermer."))
	b.WriteString("\n")

	return b.String()
}

func (g *Game) renderDeadScreen() string {
	var b strings.Builder
	titleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
	bodyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("248"))

	b.WriteString(titleStyle.Render("VOUS ÊTES MORT"))
	b.WriteString("\n\n")
	b.WriteString(bodyStyle.Render("Votre histoire s'arrête dans ce trou perdu de Roubaix."))
	b.WriteString("\n")
	b.WriteString(bodyStyle.Render("Les vieilles pierres de la filature garderont votre secret."))
	b.WriteString("\n\n")
	b.WriteString(bodyStyle.Render("Appuyez sur Entrée pour quitter."))
	b.WriteString("\n")

	return b.String()
}

func (g *Game) renderWinScreen() string {
	var b strings.Builder
	titleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("46")).Bold(true)
	bodyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("248"))
	flagStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)

	b.WriteString(titleStyle.Render("VOUS AVEZ GAGNÉ!"))
	b.WriteString("\n\n")
	b.WriteString(bodyStyle.Render("Vous avez trouvé votre passeport!"))
	b.WriteString("\n")
	b.WriteString(bodyStyle.Render("Après cette aventure à Roubaix, vous pouvez enfin"))
	b.WriteString("\n")
	b.WriteString(bodyStyle.Render("rentrer au Canada, le cœur plein de souvenirs"))
	b.WriteString("\n")
	b.WriteString(bodyStyle.Render("et l'estomac plein de poutine et de bon pain."))
	b.WriteString("\n\n")
	b.WriteString(flagStyle.Render("                🍁  VIVE LE CANADA!  🍁"))
	b.WriteString("\n\n")
	b.WriteString(bodyStyle.Render("Appuyez sur Entrée pour quitter."))
	b.WriteString("\n")

	return b.String()
}
