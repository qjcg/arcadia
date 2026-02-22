package main

import (
	"fmt"
	"os"

	"github.com/qjcg/arcadia/exp/fractalis/internal/core/persistence"
	"github.com/qjcg/arcadia/exp/fractalis/internal/ebiten"
)

func main() {
	// Create default config
	config := persistence.Config{
		Width:       800,
		Height:      600,
		MaxIter:     50,
		CenterX:     -0.5,
		CenterY:     0.0,
		Zoom:        1.0,
		ColorScheme: "rainbow",
		FractalType: persistence.FractalMandelbrot,
	}

	// Create and run the 3D game
	game := ebiten.NewGame(config)
	if err := game.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running 3D mode: %v\n", err)
		os.Exit(1)
	}
}
