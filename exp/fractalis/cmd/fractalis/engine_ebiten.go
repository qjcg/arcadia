//go:build ebiten

package main

import (
	"github.com/qjcg/arcadia/exp/fractalis/internal/core/persistence"
	"github.com/qjcg/arcadia/exp/fractalis/internal/ebiten"
)

// engineChoiceHelp documents the --engine flag choices for this build.
const engineChoiceHelp = "Rendering engine (bubbletea, ebiten)"

// runEbitenEngine starts the graphical ebiten engine.
func runEbitenEngine(config persistence.Config) error {
	return ebiten.NewGame(config).Run()
}
