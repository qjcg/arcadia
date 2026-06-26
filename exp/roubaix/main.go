package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/qjcg/arcadia/exp/roubaix/internal/game"
)

func main() {
	g := game.New()
	p := tea.NewProgram(g)
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
