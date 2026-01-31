package ebiten

import (
	_ "embed"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
)

//go:embed shader.kage
var shaderCode []byte

// initShader initializes the Kage shader
func (g *Game) initShader() {
	var err error
	g.shader, err = ebiten.NewShader(shaderCode)
	if err != nil {
		log.Fatalf("Failed to compile shader: %v", err)
	}

	g.initialized = true
}
